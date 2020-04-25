package raft

import "../labrpc"

import "time"

// import crand "crypto/rand"
// import "math/big"

func (rf *Raft) SendAppendEntries(entries []Log) {
	rf.mu.Lock()
	defer rf.mu.Unlock()

	prevLogTerm := rf.log[len(rf.log)-1].Term
	prevLogIndex := len(rf.log) - 1

	appendEntriesArgs := AppendEntriesArgs{
		Term:         rf.currentTerm,
		LeaderId:     rf.me,
		PrevLogIndex: prevLogIndex,
		PrevLogTerm:  prevLogTerm,
		Entries:      entries,
		LeaderCommit: rf.commitIndex,
	}

	for _, clientEnd := range rf.peers {
		reply := AppendEntriesReply{}
		go rf.AppendEntriesSyncForClientEnd(clientEnd, &appendEntriesArgs, &reply)
	}

	//TODO handle reply.

	// successCount := 1

	// for {
	// 	successCount = 1

	// 	for _, clientEnd := range rf.peers {
	// 		reply := AppendEntriesReply{}
	// 		go rf.HandleAppendEntriesReply(clientEnd, &appendEntriesArgs, &reply, &successCount)
	// 	}

	// 	rf.mu.Unlock()
	// 	time.Sleep(time.Millisecond * 20)

	// 	if successCount >= rf.Majority() {

	// 	}
	// }

}

func (rf *Raft) AppendEntriesSyncForClientEnd(clientEnd *labrpc.ClientEnd, args *AppendEntriesArgs, reply *AppendEntriesReply) {
	for {
		clientEnd.Call("Raft.AppendEntries", args, reply)
		rf.mu.Lock()
		if reply.Term > rf.currentTerm {
			rf.currentTerm = reply.Term
			rf.MakeLeader()
		}
		// if reply.Success {
		// 	rf.mu.Unlock()
		// 	return
		// }

		//TODO...
		rf.mu.Unlock()
		return
	}
}

func (rf *Raft) HandleAppendEntriesReply(clientEnd *labrpc.ClientEnd, args *AppendEntriesArgs, reply *AppendEntriesReply, successCount *int) {
	clientEnd.Call("Raft.AppendEntries", args, reply)
	rf.mu.Lock()
	if reply.Term > rf.currentTerm {
		rf.currentTerm = reply.Term
		rf.MakeLeader()
		if reply.Success {
			*successCount = *successCount + 1
		}
	}
	rf.mu.Unlock()
}

func (rf *Raft) AppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply) {
	rf.mu.Lock()
	defer rf.mu.Unlock()

	// DPrintln("[", rf.me, "]", "AppendEntries received from ", args.LeaderId, "by ", rf.me, "args.term => ", args.Term, "rf.currentTerm => ", rf.currentTerm)

	if args.LeaderId != rf.me && args.Term >= rf.currentTerm {
		if rf.currentState != FOLLOWER {
			rf.MakeFollower()
		}
		rf.currentTerm = args.Term
		rf.rfCond.Broadcast()
	}

	if args.Term >= rf.currentTerm {
		rf.ResetRpcTimer()
	}

	// DPrintln("[", rf.me, "]", "Marking ", rf.me, "as a FOLLOWER")

	if (args.LeaderId != rf.me && args.Term >= rf.currentTerm) || args.LeaderId == rf.me {
		reply.Success = true
	}

	reply.Term = rf.currentTerm

	// DPrintln("[", rf.me, "]", "Receving HeartBeats from ")
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

		go rf.SendAppendEntries([]Log{})

		rf.mu.Unlock()
		time.Sleep(300 * time.Millisecond)
	}
}
