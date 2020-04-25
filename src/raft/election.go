package raft

// import "../labrpc"
import "time"
import crand "crypto/rand"
import "math/big"

func (rf *Raft) sendRequestVote(server int, args *RequestVoteArgs, reply *RequestVoteReply) bool {
	// DPrintf("Sending RequestVote to %d", server)
	ok := rf.peers[server].Call("Raft.RequestVoteRpc", args, reply)
	return ok
}

func (rf *Raft) MakeFollower() {
	rf.currentState = FOLLOWER
}

func (rf *Raft) RequestVoteRpc(args *RequestVoteArgs, reply *RequestVoteReply) {
	// Your code here (2A, 2B).
	rf.mu.Lock()
	defer rf.mu.Unlock()
	// DPrintln("[", rf.me, "]", "*********RequestVote received from", args.CandidateId, "by ", rf.me)

	// DPrintln("[", rf.me, "]", "args.Term => ", args.Term, "term =>", rf.currentTerm, "rf.votedFor =>", rf.votedFor, "args.LastLogIndex => ", args.LastLogIndex, "rf.commitIndex =>", rf.commitIndex)

	if args.Term < rf.currentTerm {
		reply.VoteGranted = false
		reply.Term = rf.currentTerm
		rf.votedFor = -1
	} else if (rf.votedFor == -1 || rf.votedFor == args.CandidateId) && args.LastLogIndex >= rf.commitIndex {
		reply.VoteGranted = true
		rf.ResetRpcTimer()
		rf.votedFor = args.CandidateId
		reply.Term = args.Term
		// DPrintln("[", rf.me, "]", rf.me, " Voting for ", args.CandidateId)
	} else {
		if args.Term > rf.currentTerm {
			rf.ResetRpcTimer()
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

	nextLogIdx := len(rf.log) + 1

	for i, _ := range rf.peers {
		rf.nextIndex[i] = nextLogIdx
		//TODO matchIndex
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

	lastLogTerm := rf.log[len(rf.log)-1].Term
	lastLogIndex := len(rf.log) - 1

	requestVoteArgs := &RequestVoteArgs{
		Term:         rf.currentTerm,
		CandidateId:  rf.me,
		LastLogIndex: lastLogIndex,
		LastLogTerm:  lastLogTerm,
	}

	requestVoteReplyMp := make(map[int]*RequestVoteReply)

	requestVoteReplyMp[rf.me] = &RequestVoteReply{
		Term:        rf.currentTerm,
		VoteGranted: true,
	}
	var repliesCount int
	repliesCount = 0
	for i, _ := range rf.peers {
		if i != rf.me {
			reply := &RequestVoteReply{}
			requestVoteReplyMp[i] = reply
			go func(i int, args *RequestVoteArgs, reply *RequestVoteReply, repliesCount *int) {
				ok := rf.sendRequestVote(i, requestVoteArgs, reply)
				rf.mu.Lock()
				if ok {
					*repliesCount = *repliesCount + 1
				}
				rf.mu.Unlock()
			}(i, requestVoteArgs, reply, &repliesCount)
		}
	}

	rf.mu.Unlock()
	//Wait for votes
	time.Sleep(time.Millisecond * 20)
	rf.mu.Lock()
	// DPrintln("[", rf.me, "]", "Counting votes for ", rf.me, "Replies count", repliesCount, "peers => ", len(rf.peers)/2)
	if repliesCount >= rf.Majority() {
		votes := 0
		for _, reply := range requestVoteReplyMp {
			// DPrintln("[", rf.me, "]", "Reply from ", i, reply)

			if reply.Term > rf.currentTerm {
				rf.MakeLeader()
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
			// DPrintln("[", rf.me, "]", "LEADER is here....", rf.me)
			rf.rfCond.Broadcast()
		}
	}
	rf.mu.Unlock()
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

		max := big.NewInt(400)
		rr, _ := crand.Int(crand.Reader, max)
		// DPrintln("[", rf.me, "]", "Election paused for ", rr.Int64())
		time.Sleep((200 + time.Duration(rr.Int64())) * time.Millisecond)

	}
}

//if communication is not received for LastRpcTimeOut calls for an election
func (rf *Raft) CheckAndKickOfLeaderElection() {
	for {
		max := big.NewInt(200)
		rr, _ := crand.Int(crand.Reader, max)
		time.Sleep(time.Duration(rr.Int64()) * time.Millisecond)

		rf.mu.Lock()
		if rf.killed() {
			rf.mu.Unlock()
			return
		}

		timeOutAt := rf.lastRpcReceived.Add(LastRpcTimeOut)
		if time.Now().After(timeOutAt) {
			rf.currentState = CANDIDATE
			go rf.StartElectionProcess()

			rf.mu.Unlock()
			// let the election happen
			time.Sleep(1000 * time.Millisecond)

		} else {
			rf.mu.Unlock()
			time.Sleep(timeOutAt.Sub(time.Now()))
		}
	}
}
