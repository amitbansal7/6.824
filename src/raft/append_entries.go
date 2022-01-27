package raft

import "time"

type AppendEntriesArgs struct {
	Term     int
	LeaderId int
}

type AppendEntriesReply struct {
	Term    int
	Success bool
}

func (rf *Raft) AppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply) {
	rf.mu.Lock()

	if args.Term >= rf.currentTerm {
		rf.ResetRpcTimer()
		rf.currentTerm = args.Term
		rf.currentState = FOLLOWER
	}
	reply.Term = rf.currentTerm
	reply.Success = true
	rf.mu.Unlock()
}

func (rf *Raft) AppendEntriesTicker() {
	for !rf.killed() {
		rf.mu.Lock()
		if rf.currentState == LEADER {
			rf.SendAppendEntries()
		}
		rf.mu.Unlock()
		time.Sleep(time.Millisecond * 100)
	}
}

func (rf *Raft) SendAppendEntries() {
	rf.ResetRpcTimer()
	for id := range rf.peers {
		if id != rf.me {
			go func(peer int) {
				args := &AppendEntriesArgs{}
				reply := &AppendEntriesReply{}
				rf.peers[peer].Call("Raft.AppendEntries", args, reply)

				// DPrintf("[%d] Sending append to [%d] response", rf.me, id)
			}(id)
		}
	}
}
