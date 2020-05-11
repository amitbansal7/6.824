package raft


import "sync"
import "sync/atomic"
import "../labrpc"
import "time"
import "fmt"

// import crand "crypto/rand"
// import "math/big"

import "bytes"
import "../labgob"

const (
	FOLLOWER       = 0
	LEADER         = 1
	CANDIDATE      = 2
)

type ApplyMsg struct {
	CommandValid bool
	Command      interface{}
	CommandIndex int
}

type Log struct {
	Command interface{}
	Term    int
	Index   int
}

type Raft struct {
	mu        sync.Mutex          // Lock to protect shared access to this peer's state
	peers     []*labrpc.ClientEnd // RPC end points of all peers
	persister *Persister          // Object to hold this peer's persisted state
	me        int                 // this peer's index into peers[]
	dead      int32               // set by Kill()

	currentTerm int
	votedFor    int
	log         []Log
	commitIndex int
	lastApplied int
	applyCh     chan ApplyMsg

	//Volatile state
	nextIndex  []int
	matchIndex []int

	lastRpcReceived time.Time
	currentState    int
	rfCond          *sync.Cond
}

func (rf *Raft) GetState() (int, bool) {

	rf.mu.Lock()
	defer rf.mu.Unlock()
	return rf.currentTerm, rf.currentState == LEADER
}

func (rf *Raft) IncTerm() {
	rf.currentTerm += 1
	// voterFor nill in new term
	rf.votedFor = -1
}


type PersistedData struct {
	CurrentTerm int
	VotedFor    int
	Log         *[]Log
}

func (rf *Raft) persist() {

	persisted := &PersistedData{
		CurrentTerm: rf.currentTerm,
		VotedFor: rf.votedFor,
		Log: &rf.log,
	}

	w := new(bytes.Buffer)
	e := labgob.NewEncoder(w)
	e.Encode(persisted)
	data := w.Bytes()
	rf.persister.SaveRaftState(data)
}

func (rf *Raft) readPersist(data []byte) {
	if data == nil || len(data) < 1 {
		return
	}

	r := bytes.NewBuffer(data)
	d := labgob.NewDecoder(r)

	var persisted PersistedData
	if d.Decode(&persisted) != nil {
	  fmt.Println("readPersist error......")
	} else {
	  rf.currentTerm = persisted.CurrentTerm
	  rf.votedFor = persisted.VotedFor
	  rf.log = *persisted.Log
	}
}

//RPC types
type AppendEntriesArgs struct {
	Term         int
	LeaderId     int
	PrevLogIndex int
	PrevLogTerm  int
	Entries      []Log
	LeaderCommit int
}

type AppendEntriesReply struct {
	Term    int
	Success bool
	NextIndex int
}

type RequestVoteArgs struct {
	Term         int
	CandidateId  int
	LastLogIndex int
	LastLogTerm  int
}

type RequestVoteReply struct {
	Term        int
	VoteGranted bool
}

///*************

func (rf *Raft) ResetRpcTimer() {
	rf.lastRpcReceived = time.Now()
}

func (rf *Raft) Majority() int {
	return len(rf.peers) / 2
}

func (rf *Raft) LastLog() Log {
	return rf.log[len(rf.log)-1]
}

func (rf *Raft) Start(command interface{}) (int, int, bool) {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	index := -1
	term := -1
	isLeader := rf.currentState == LEADER

	if !isLeader {
		return index, term, isLeader
	}

	log := Log{
		Command: command,
		Term:    rf.currentTerm,
		Index:   rf.LastLog().Index + 1,
	}

	entries := []Log{log}

	rf.log = append(rf.log, entries...)
	rf.persist()

	rf.matchIndex[rf.me] = log.Index
	rf.nextIndex[rf.me] = log.Index + 1

	go rf.SendAppendEntries()

	return log.Index, log.Term, isLeader
}

func (rf *Raft) Kill() {
	atomic.StoreInt32(&rf.dead, 1)
	// DPrintln("[", rf.me, "]", "Kill called....")
	// Your code here, if desired.
}

func (rf *Raft) killed() bool {
	z := atomic.LoadInt32(&rf.dead)
	return z == 1
}

func Make(peers []*labrpc.ClientEnd, me int,
	persister *Persister, applyCh chan ApplyMsg) *Raft {
	rf := &Raft{}
	rf.peers = peers
	rf.persister = persister
	rf.me = me
	rf.currentTerm = 0
	rf.votedFor = -1
	rf.log = []Log{Log{Term: 0, Index: 0}}
	rf.commitIndex = 0
	rf.lastApplied = 0
	rf.ResetRpcTimer()
	rf.MakeFollower()
	rf.rfCond = sync.NewCond(&rf.mu)
	rf.applyCh = applyCh

	// initialize from state persisted before a crash
	rf.readPersist(persister.ReadRaftState())

	go rf.CheckAndKickOfLeaderElection()
	go rf.SendHeartBeats()
	go rf.Applier()

	return rf
}
