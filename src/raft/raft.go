package raft


import "sync"
import "sync/atomic"
import "../labrpc"
import "time"
// import "fmt"

// import crand "crypto/rand"
// import "math/big"

// import "bytes"
// import "../labgob"

const (
	FOLLOWER       = 0
	LEADER         = 1
	CANDIDATE      = 2
)

//
// as each Raft peer becomes aware that successive log entries are
// committed, the peer should send an ApplyMsg to the service (or
// tester) on the same server, via the applyCh passed to Make(). set
// CommandValid to true to indicate that the ApplyMsg contains a newly
// committed log entry.
//
// in Lab 3 you'll want to send other kinds of messages (e.g.,
// snapshots) on the applyCh; at that point you can add fields to
// ApplyMsg, but set CommandValid to false for these other uses.
//
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

//
// save Raft's persistent state to stable storage,
// where it can later be retrieved after a crash and restart.
// see paper's Figure 2 for a description of what should be persistent.
//
func (rf *Raft) persist() {
	// Your code here (2C).
	// Example:
	// w := new(bytes.Buffer)
	// e := labgob.NewEncoder(w)
	// e.Encode(rf.xxx)
	// e.Encode(rf.yyy)
	// data := w.Bytes()
	// rf.persister.SaveRaftState(data)
}

//
// restore previously persisted state.
//
func (rf *Raft) readPersist(data []byte) {
	if data == nil || len(data) < 1 { // bootstrap without any state?
		return
	}
	// Your code here (2C).
	// Example:
	// r := bytes.NewBuffer(data)
	// d := labgob.NewDecoder(r)
	// var xxx
	// var yyy
	// if d.Decode(&xxx) != nil ||
	//    d.Decode(&yyy) != nil {
	//   error...
	// } else {
	//   rf.xxx = xxx
	//   rf.yyy = yyy
	// }
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

	lastLog := rf.LastLog()
	rf.log = append(rf.log, entries...)

	go rf.SendAppendEntries(entries, lastLog)

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
