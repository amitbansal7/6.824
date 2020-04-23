package raft

import "../labrpc"
import "time"
import crand "crypto/rand"
import "math/big"

func (rf *Raft) sendRequestVote(server int, args *RequestVoteArgs, reply *RequestVoteReply) bool {
	// DPrintf("Sending RequestVote to %d", server)
	ok := rf.peers[server].Call("Raft.RequestVoteRpc", args, reply)
	return ok
}

func (rf *Raft) SendHeartBeats() {
	for {
		rf.mu.Lock()
		if rf.killed() {
			rf.mu.Unlock()
			return
		}
		for rf.currentState != LEADER {
			rf.rfCond.Wait()
		}
		DPrintln("[", rf.me, "]", "Sending HeartBeats for term =>", rf.currentTerm)
		appendEntriesArgs := AppendEntriesArgs{
			Term:     rf.currentTerm,
			LeaderId: rf.me,
			// PreLogIndex  int
			// PrevLogTerm  int
			// Entries      []int
			// LeaderCommit int
		}
		for _, clientEnd := range rf.peers {
			reply := AppendEntriesReply{}
			go func(clientEnd *labrpc.ClientEnd, args *AppendEntriesArgs, reply *AppendEntriesReply) {
				clientEnd.Call("Raft.AppendEntries", args, reply)
				rf.mu.Lock()
				if reply.Term > rf.currentTerm {
					rf.currentTerm = reply.Term
					rf.currentState = FOLLOWER
				}
				rf.mu.Unlock()

			}(clientEnd, &appendEntriesArgs, &reply)
		}

		rf.mu.Unlock()
		time.Sleep(250 * time.Millisecond)
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
	DPrintln("[", rf.me, "]", "Election started by %d", rf.me, "for term =>", rf.currentTerm)
	rf.currentState = CANDIDATE
	rf.votedFor = rf.me

	requestVoteArgs := &RequestVoteArgs{
		Term:        rf.currentTerm,
		CandidateId: rf.me,
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
	DPrintln("[", rf.me, "]", "Counting votes for ", rf.me, "Replies count", repliesCount, "peers => ", len(rf.peers)/2)
	if repliesCount >= (len(rf.peers) / 2) {
		votes := 0
		for i, reply := range requestVoteReplyMp {
			DPrintln("[", rf.me, "]", "Reply from ", i, reply)

			if reply.Term > rf.currentTerm {
				rf.currentState = FOLLOWER
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
		if votes > (len(rf.peers) / 2) {
			rf.currentState = LEADER
			DPrintln("[", rf.me, "]", "LEADER is here....", rf.me)
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
		DPrintln("[", rf.me, "]", "Election paused for ", rr.Int64())
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
