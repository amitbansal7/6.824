package raft

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

func (rf *Raft) SendAppendEntries() {
	rf.mu.Lock()
	defer rf.mu.Unlock()

	rf.ResetRpcTimer()
	for id := range rf.peers {
		if id != rf.me {
			go rf.SendAppendEntriesToPeer(id)
		}
	}
}

func (rf *Raft) SendAppendEntriesToPeer(peer int) {
	rf.mu.Lock()
	defer rf.mu.Unlock()

	args := &AppendEntriesArgs{Term: rf.currentTerm, LeaderId: rf.me}
	reply := &AppendEntriesReply{}
	go rf.peers[peer].Call("Raft.AppendEntries", args, reply)
}
