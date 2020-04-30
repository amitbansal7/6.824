package raft

// import "../labrpc"
import "time"
import crand "crypto/rand"
import "math/big"

// import "fmt"

func (rf *Raft) sendRequestVote(server int, args *RequestVoteArgs, reply *RequestVoteReply, replyMap map[int]*RequestVoteReply) bool {
	// DPrintf("Sending RequestVote to %d", server)
	rf.mu.Lock()
	peer := rf.peers[server]
	rf.mu.Unlock()
	ok := peer.Call("Raft.RequestVoteRpc", args, reply)
	rf.mu.Lock()
	replyMap[server] = reply
	rf.mu.Unlock()
	return ok
}

func (rf *Raft) MakeFollower() {
	rf.currentState = FOLLOWER
}

//5.4.1 Election restriction check.
func (rf *Raft) CandidateIsUptoDate(args *RequestVoteArgs) bool {
	lastLog := rf.LastLog()

	// DPrintln("[", rf.me, "]", args.LastLogTerm, lastLog.Term, args.LastLogIndex, len(rf.log)-1)

	if args.LastLogTerm != lastLog.Term {
		return args.LastLogTerm > lastLog.Term
	} else {
		//TODO make sure this is correct.
		return args.LastLogIndex+1 >= len(rf.log)
	}
}

func (rf *Raft) RequestVoteRpc(args *RequestVoteArgs, reply *RequestVoteReply) {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	// DPrintln("[", rf.me, "]", "*********RequestVote received from", args.CandidateId, "by ", rf.me)

	// DPrintln("[", rf.me, "]", "args.Term => ", args.Term, "term =>", rf.currentTerm, "rf.votedFor =>", rf.votedFor, "args.LastLogIndex => ", args.LastLogIndex, "rf.commitIndex =>", rf.commitIndex)

	if args.Term < rf.currentTerm {
		reply.VoteGranted = false
		reply.Term = rf.currentTerm
		rf.votedFor = -1
	} else if (rf.votedFor == -1 || rf.votedFor == args.CandidateId) && rf.CandidateIsUptoDate(args) {
		reply.VoteGranted = true
		rf.ResetRpcTimer()
		rf.MakeFollower()
		rf.votedFor = args.CandidateId
		reply.Term = rf.currentTerm
		// DPrintln("[", rf.me, "]", rf.me, " Voting for ", args.CandidateId)
	} else {
		if args.Term > rf.currentTerm {
			rf.ResetRpcTimer()
			rf.MakeFollower()
			rf.currentTerm = args.Term
			rf.votedFor = -1
		}
	}
	// DPrintln("[", rf.me, "]", "*********RequestVote Sent for", args.CandidateId, "by ", rf.me, "=>", reply)
}

func (rf *Raft) MakeLeader() {
	rf.currentState = LEADER
	rf.nextIndex = make([]int, len(rf.peers), len(rf.peers))
	rf.matchIndex = make([]int, len(rf.peers), len(rf.peers))

	nextLogIdx := len(rf.log)

	for i, _ := range rf.peers {
		rf.nextIndex[i] = nextLogIdx
		rf.matchIndex[i] = 0
	}
}

func (rf *Raft) ConductElection() {
	rf.mu.Lock()
	if rf.killed() {
		rf.mu.Unlock()
		return
	}
	rf.ResetRpcTimer()
	rf.IncTerm()
	// DPrintln("[", rf.me, "]", "Election started by %d", rf.me, "for term =>", rf.currentTerm)
	rf.currentState = CANDIDATE
	rf.votedFor = rf.me

	requestVoteReplyMp := make(map[int]*RequestVoteReply)

	requestVoteReplyMp[rf.me] = &RequestVoteReply{
		Term:        rf.currentTerm,
		VoteGranted: true,
	}
	//Voted for itself.
	repliesCount := 1

	for i, _ := range rf.peers {
		if i != rf.me {
			reply := &RequestVoteReply{}
			// requestVoteReplyMp[i] = reply
			requestVoteArgs := &RequestVoteArgs{
				Term:         rf.currentTerm,
				CandidateId:  rf.me,
				LastLogIndex: len(rf.log) - 1,
				LastLogTerm:  rf.LastLog().Term,
			}
			go func(i int, args *RequestVoteArgs, reply *RequestVoteReply, repliesCount *int, requestVoteReplyMp map[int]*RequestVoteReply) {
				ok := rf.sendRequestVote(i, requestVoteArgs, reply, requestVoteReplyMp)
				rf.mu.Lock()
				if ok {
					*repliesCount = *repliesCount + 1
				}
				rf.mu.Unlock()
			}(i, requestVoteArgs, reply, &repliesCount, requestVoteReplyMp)
		}
	}

	rf.mu.Unlock()
	// DPrintln("[", rf.me, "]", "Counting votes for ", rf.me, "Replies count", repliesCount, "peers => ", len(rf.peers)/2)
	sleeps := 0
	for {
		if sleeps > 20 {
			return
		}
		rf.mu.Lock()
		if rf.currentState != CANDIDATE {
			rf.mu.Unlock()
			return
		}
		if repliesCount < rf.Majority() {
			rf.mu.Unlock()
			time.Sleep(time.Millisecond * 10)
			sleeps += 1
			continue
		}

		if repliesCount >= rf.Majority() {
			votes := 0
			for _, reply := range requestVoteReplyMp {
				// DPrintln("[", rf.me, "]", "Reply from ", i, reply)

				if reply.Term > rf.currentTerm {
					rf.MakeFollower()
					rf.currentTerm = reply.Term
					rf.votedFor = -1
					rf.rfCond.Broadcast()
					rf.mu.Unlock()
					return
				} else {
					if reply.VoteGranted {
						votes += 1
					}
				}
			}
			if votes > rf.Majority() {
				rf.MakeLeader()
				// fmt.Println("[", rf.me, ",", rf.currentTerm, "]", "LEADER is here....", rf.me)
				rf.rfCond.Broadcast()
			}
		}
		rf.mu.Unlock()
	}
}

func (rf *Raft) StartElectionProcess() {
	for {
		rf.mu.Lock()

		if rf.killed() {
			rf.mu.Unlock()
			return
		}

		if rf.currentState == CANDIDATE {
			rf.mu.Unlock()
			rf.ConductElection()
		} else {
			rf.votedFor = -1
			rf.mu.Unlock()
			return
		}
		time.Sleep(rf.ElectionTimeOut())
	}
}

func (rf *Raft) ElectionTimeOut() time.Duration {
	tt, _ := crand.Int(crand.Reader, big.NewInt(180))
	r := 280 + int(tt.Int64())
	tm := time.Duration(int(r)) * time.Millisecond
	return tm
}

//if communication is not received for LastRpcTimeOut calls for an election
func (rf *Raft) CheckAndKickOfLeaderElection() {
	for {

		rf.mu.Lock()
		if rf.killed() {
			rf.mu.Unlock()
			return
		}

		timeOutAt := rf.lastRpcReceived.Add(rf.ElectionTimeOut())
		if time.Now().After(timeOutAt) {
			rf.currentState = CANDIDATE
			go rf.StartElectionProcess()

			rf.mu.Unlock()
			// let the election happen
			time.Sleep(rf.ElectionTimeOut())

		} else {
			rf.mu.Unlock()
			time.Sleep(timeOutAt.Sub(time.Now()))
		}
	}
}
