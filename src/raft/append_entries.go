package raft

import "../labrpc"
import "time"
import "fmt"

// import crand "crypto/rand"
// import "math/big"

func (rf *Raft) SendAppendEntries() {
	rf.mu.Lock()
	defer rf.mu.Unlock()

	// commitTillIndex := rf.log[len(rf.log)-1].Index
	commitTillIndex := len(rf.log) - 1

	for i, clientEnd := range rf.peers {
		if i != rf.me {
			go rf.AppendEntriesSyncForClientEnd(i, clientEnd, commitTillIndex)
		}
	}

}

func (rf *Raft) AppendEntriesSyncForClientEnd(i int, clientEnd *labrpc.ClientEnd, commitTillIndex int) {

	rf.mu.Lock()
	if rf.currentState != LEADER {
		rf.mu.Unlock()
		return
	}

	lastLog := &rf.log[rf.nextIndex[i]-1]
	prevLogIndex := (*lastLog).Index
	prevLogTerm := (*lastLog).Term

	entries := make([]Log, len(rf.log[rf.nextIndex[i]:]))
	copy(entries, rf.log[rf.nextIndex[i]:])

	reply := &AppendEntriesReply{}
	args := &AppendEntriesArgs{
		Term:         rf.currentTerm,
		LeaderId:     rf.me,
		PrevLogIndex: prevLogIndex,
		PrevLogTerm:  prevLogTerm,
		Entries:      entries,
		LeaderCommit: rf.commitIndex,
	}

	lastLog = nil
	rf.mu.Unlock()
	ok := clientEnd.Call("Raft.AppendEntries", args, reply)
	if !ok {
		// time.Sleep(5 * time.Millisecond)
		return
	}
	rf.mu.Lock()
	if rf.currentState != LEADER {
		rf.mu.Unlock()
		return
	}
	if reply.Success {
		if len(args.Entries) > 0 {
			rf.nextIndex[i] = args.Entries[len(args.Entries)-1].Index + 1
		}
		rf.matchIndex[i] = commitTillIndex

		if rf.nextIndex[i] > len(rf.log) {
			rf.nextIndex[i] = len(rf.log)
		}

		for N := len(rf.log) - 1; N > rf.commitIndex; N-- {
			count := 0
			for _, index := range rf.matchIndex {
				if index >= N {
					count += 1
				}
			}
			if count > rf.Majority() {
				rf.commitIndex = N
				rf.rfCond.Broadcast()
				break
			}
		}

		rf.mu.Unlock()
		return
	} else {
		if reply.Term > rf.currentTerm {
			rf.currentTerm = reply.Term
			rf.MakeFollower()
			rf.persist()
			rf.mu.Unlock()
			return
		} else {

			for rf.log[reply.NextIndex].Term == args.PrevLogTerm {
				reply.NextIndex -= 1
			}

			rf.nextIndex[i] = reply.NextIndex
			if rf.nextIndex[i] < 1 {
				rf.nextIndex[i] = 1
				// fmt.Println("Next Index is < 1 for =>", i)
			}
		}
	}
	rf.mu.Unlock()
}

func (rf *Raft) AppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply) {
	rf.mu.Lock()
	defer rf.mu.Unlock()

	if rf.killed() {
		return
	}
	reply.Term = rf.currentTerm
	reply.Success = true

	if args.Term >= rf.currentTerm {
		rf.MakeFollower()
		rf.ResetRpcTimer()
		rf.currentTerm = args.Term
		rf.rfCond.Broadcast()
	}

	//1
	if args.Term < rf.currentTerm {
		reply.Success = false
		return
	}

	//2
	if args.PrevLogIndex >= len(rf.log) || rf.log[args.PrevLogIndex].Term != args.PrevLogTerm || rf.log[args.PrevLogIndex].Index != args.PrevLogIndex {
		reply.Success = false
		last := args.PrevLogIndex
		if last >= len(rf.log) {
			last = len(rf.log) - 1
		}
		for i := last; i > 0; i-- {
			if rf.log[i].Term != args.PrevLogTerm {
				reply.NextIndex = i
				break
			}
		}
		return
	}

	//3
	for _, entry := range args.Entries {
		for i := len(rf.log) - 1; i >= 0; i-- {
			if entry.Index == rf.log[i].Index && entry.Term != rf.log[i].Term {
				rf.log = append(rf.log[:i])
				break
			}
		}
	}

	//4
	for _, entry := range args.Entries {
		found := false
		for i := len(rf.log) - 1; i >= 0; i-- {
			if entry.Index == rf.log[i].Index && entry.Term == rf.log[i].Term {
				found = true
			}
		}
		if !found {
			rf.log = append(rf.log, entry)
		}
	}

	//5
	lastEntry := rf.LastLog()

	if lastEntry.Index != len(rf.log)-1 {
		fmt.Println("[", rf.me, ", ", rf.currentTerm, "]", "index and .Index mismatch....!!!!!!!!")
		fmt.Println(lastEntry, len(rf.log)-1)
	}

	if args.LeaderCommit > rf.commitIndex {
		if args.LeaderCommit < lastEntry.Index {
			rf.commitIndex = args.LeaderCommit
		} else {
			rf.commitIndex = lastEntry.Index
		}
	}
	rf.rfCond.Broadcast()
	rf.persist()
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

		rf.ResetRpcTimer()

		go rf.SendAppendEntries()

		rf.mu.Unlock()
		time.Sleep(110 * time.Millisecond)
	}
}
