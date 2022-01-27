package raft

type RequestVoteArgs struct {
	Term        int
	CandidateId int
}

type RequestVoteReply struct {
	Term        int
	VoteGranted bool
}

func (rf *Raft) RequestVote(args *RequestVoteArgs, reply *RequestVoteReply) {
	rf.mu.Lock()
	defer rf.mu.Unlock()

	if rf.lastVotedTerm < args.Term && args.Term > rf.currentTerm {
		rf.ResetRpcTimer()
		reply.VoteGranted = true
		reply.Term = args.Term
		rf.lastVotedTerm = args.Term
		rf.currentTerm = args.Term
		rf.currentState = FOLLOWER
		DPrintf("[%d] %d voting for %d", rf.currentTerm, rf.me, args.CandidateId)
	} else {
		reply.VoteGranted = false
		reply.Term = rf.currentTerm
	}
}

func (rf *Raft) sendRequestVote(server int, args *RequestVoteArgs, reply *RequestVoteReply) bool {
	ok := rf.peers[server].Call("Raft.RequestVote", args, reply)
	return ok
}

func (rf *Raft) ConductElection() {
	rf.mu.Lock()

	rf.ResetRpcTimer()
	rf.currentState = CANDIDATE
	rf.currentTerm += 1
	rf.lastVotedTerm = rf.currentTerm
	electionTerm := rf.currentTerm

	replies := make(chan RequestVoteReply, len(rf.peers))
	replies <- RequestVoteReply{VoteGranted: true, Term: rf.currentTerm}

	for peer := range rf.peers {
		if peer != rf.me {
			go func(peer int, currentTerm int, candidateId int, repliesChan chan RequestVoteReply) {
				reply := RequestVoteReply{VoteGranted: false}
				rf.sendRequestVote(
					peer,
					&RequestVoteArgs{Term: currentTerm, CandidateId: candidateId},
					&reply,
				)
				repliesChan <- reply
			}(peer, electionTerm, rf.me, replies)
		}
	}

	rf.mu.Unlock()
	votes := 0
	for i := 0; i < len(rf.peers); i++ {
		reply := <-replies
		if reply.VoteGranted && reply.Term == electionTerm {
			votes += 1
			if votes >= ((len(rf.peers) / 2) + 1) {
				rf.MakeLeader(electionTerm)
				return
			}
		}
	}
}

func (rf *Raft) MakeLeader(electionTerm int) {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	if !rf.killed() && rf.currentTerm == electionTerm && rf.currentState == CANDIDATE {
		rf.currentState = LEADER
		go rf.AppendEntriesTicker()
		DPrintf("[%d][%d] Leader is here", rf.me, rf.currentTerm)
	}
}
