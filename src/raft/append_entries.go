package raft

import "../labrpc"

import "time"

// import "fmt"

// import "sort"

// import crand "crypto/rand"
// import "math/big"

func (rf *Raft) SendAppendEntries(entries []Log, lastLog Log) {
	rf.mu.Lock()
	defer rf.mu.Unlock()

	successCount := 1
	isCommited := false
	commitTillIndex := len(rf.log) - 1

	for i, clientEnd := range rf.peers {
		if i != rf.me {
			go rf.AppendEntriesSyncForClientEnd(i, clientEnd, entries, &successCount, &isCommited, commitTillIndex, &lastLog)
		}
	}

}

func (rf *Raft) AppendEntriesSyncForClientEnd(i int, clientEnd *labrpc.ClientEnd, startEntries []Log, successCount *int, isCommited *bool, commitTillIndex int, lastLog *Log) {

	for {

		rf.mu.Lock()
		if rf.currentState != LEADER {
			rf.mu.Unlock()
			return
		}

		// if lastLog == nil {
			lastLog = &rf.log[rf.nextIndex[i]-1]
		// }

		prevLogIndex := (*lastLog).Index
		prevLogTerm := (*lastLog).Term

		reply := &AppendEntriesReply{}
		args := &AppendEntriesArgs{
			Term:         rf.currentTerm,
			LeaderId:     rf.me,
			PrevLogIndex: prevLogIndex,
			PrevLogTerm:  prevLogTerm,
			Entries:      append(rf.log[rf.nextIndex[i]:]),
			LeaderCommit: rf.commitIndex,
		}

		lastLog = nil
		rf.mu.Unlock()
		ok := clientEnd.Call("Raft.AppendEntries", args, reply)
		if !ok {
			// time.Sleep(5 * time.Millisecond)
			continue
		}
		rf.mu.Lock()
		if rf.currentState != LEADER {
			rf.mu.Unlock()
			return
		}
		if reply.Success {
			rf.nextIndex[i] += len(args.Entries)

			rf.matchIndex[i] = commitTillIndex

			if rf.nextIndex[i] > len(rf.log) {
				rf.nextIndex[i] = len(rf.log)
			}
			*successCount = *successCount + 1

			if !*isCommited && *successCount > rf.Majority() && commitTillIndex < len(rf.log) && rf.log[commitTillIndex].Term == rf.currentTerm {
				rf.commitIndex = commitTillIndex
				*isCommited = true
				rf.rfCond.Broadcast()
			}
			rf.mu.Unlock()
			return
		} else {
			if reply.Term > rf.currentTerm {
				rf.currentTerm = reply.Term
				rf.MakeFollower()
				// rf.persist()
				rf.mu.Unlock()
				return
			} else {
				rf.nextIndex[i] = reply.NextIndex
				if rf.nextIndex[i] < 1 {
					rf.nextIndex[i] = 1
					// fmt.Println("Next Index is < 1 for =>", i)
				}
			}
		}
		rf.mu.Unlock()
	}
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
		if last >= len(rf.log){
			last = len(rf.log)-1
		}
		for i := last; i > 0; i-- {
			if rf.log[i].Term != args.PrevLogTerm{
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
	if len(args.Entries) > 0 {
		lastEntry = args.Entries[len(args.Entries)-1]
	}

	// if lastEntry.Index != len(rf.log)-1 {
		// fmt.Println("[", rf.me, ", ", rf.currentTerm, "]", "index and .Index mismatch....!!!!!!!!")
		// fmt.Println(lastEntry, rf.log, args.Entries)
	// }

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

		go rf.SendAppendEntries([]Log{}, rf.LastLog())

		lastMajority := 0

		for {
			lastMajority += 1
			count := 0
			for _, val := range rf.matchIndex{
				if val >= lastMajority {
					count += 1
				}
			}
			if count <= rf.Majority() {
				lastMajority -= 1
				break
			}
		}

		if lastMajority > rf.commitIndex {
			rf.commitIndex = lastMajority
			rf.rfCond.Broadcast()
		}

		rf.mu.Unlock()
		time.Sleep(280 * time.Millisecond)
	}
}
