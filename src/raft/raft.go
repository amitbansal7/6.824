package raft

//
// this is an outline of the API that raft must expose to
// the service (or tester). see comments below for
// each of these functions for more details.
//
// rf = Make(...)
//   create a new Raft server.
// rf.Start(command interface{}) (index, term, isleader)
//   start agreement on a new log entry
// rf.GetState() (term, isLeader)
//   ask a Raft for its current term, and whether it thinks it is leader
// ApplyMsg
//   each time a new entry is committed to the log, each Raft peer
//   should send an ApplyMsg to the service (or tester)
//   in the same server.
//

import "sync"
import "sync/atomic"
import "../labrpc"
import "time"

// import crand "crypto/rand"
// import "math/big"

// import "bytes"
// import "../labgob"

const (
	LastRpcTimeOut = time.Millisecond * 300 // if no communication is received, calls for an election
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

	//Volatile state
	nextIndex  []int
	matchIndex []int

	lastRpcReceived time.Time
	currentState    int
	rfCond          *sync.Cond
}

// return currentTerm and whether this server
// believes it is the leader.
func (rf *Raft) GetState() (int, bool) {

	rf.mu.Lock()
	defer rf.mu.Unlock()
	DPrintln("[", rf.me, "]", "GetState from", rf.me, "=> ", rf.currentTerm, rf.currentState == LEADER)
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
	PreLogIndex  int
	PrevLogTerm  int
	Entries      []Log
	LeaderCommit int
}

type AppendEntriesReply struct {
	Term    int
	Success bool
}

type RequestVoteArgs struct {
	// Your data here (2A, 2B).
	Term         int //Candidates term
	CandidateId  int
	LastLogIndex int
	LastLogTerm  int
}

type RequestVoteReply struct {
	// Your data here (2A).
	Term        int //current Term for candidate to update itself
	VoteGranted bool
}

///*************

func (rf *Raft) ResetRpcTimer() {
	rf.lastRpcReceived = time.Now()
}

func (rf *Raft) RequestVoteRpc(args *RequestVoteArgs, reply *RequestVoteReply) {
	// Your code here (2A, 2B).
	rf.mu.Lock()
	defer rf.mu.Unlock()
	DPrintln("[", rf.me, "]", "*********RequestVote received from", args.CandidateId, "by ", rf.me)

	DPrintln("[", rf.me, "]", "args.Term => ", args.Term, "term =>", rf.currentTerm, "rf.votedFor =>", rf.votedFor, "args.LastLogIndex => ", args.LastLogIndex, "rf.commitIndex =>", rf.commitIndex)

	if args.Term < rf.currentTerm {
		reply.VoteGranted = false
		reply.Term = rf.currentTerm
		rf.votedFor = -1
	} else if (rf.votedFor == -1 || rf.votedFor == args.CandidateId) && args.LastLogIndex >= rf.commitIndex {
		reply.VoteGranted = true
		rf.ResetRpcTimer()
		rf.votedFor = args.CandidateId
		reply.Term = args.Term
		DPrintln("[", rf.me, "]", rf.me, " Voting for ", args.CandidateId)
	}else{
		if args.Term > rf.currentTerm {
			rf.ResetRpcTimer()
			rf.currentTerm = args.Term
			rf.votedFor = -1
		}
	}
	DPrintln("[", rf.me, "]", "*********RequestVote Sent for", args.CandidateId, "by ", rf.me, "=>", reply)
}

func (rf *Raft) AppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply) {
	rf.mu.Lock()
	defer rf.mu.Unlock()

	DPrintln("[", rf.me, "]", "AppendEntries received from ", args.LeaderId, "by ", rf.me, "args.term => ", args.Term, "rf.currentTerm => ", rf.currentTerm)

	if args.LeaderId != rf.me && args.Term >= rf.currentTerm {
		rf.currentState = FOLLOWER
		rf.currentTerm = args.Term
		rf.rfCond.Broadcast()
	}

	if args.Term >= rf.currentTerm {
		rf.ResetRpcTimer()
	}

	DPrintln("[", rf.me, "]", "Marking ", rf.me, "as a FOLLOWER")

	if args.LeaderId != rf.me && args.Term >= rf.currentTerm {
		reply.Success = true
	}
	reply.Term = rf.currentTerm

	DPrintln("[", rf.me, "]", "Receving HeartBeats from ")
}

//
// the service using Raft (e.g. a k/v server) wants to start
// agreement on the next command to be appended to Raft's log. if this
// server isn't the leader, returns false. otherwise start the
// agreement and return immediately. there is no guarantee that this
// command will ever be committed to the Raft log, since the leader
// may fail or lose an election. even if the Raft instance has been killed,
// this function should return gracefully.
//
// the first return value is the index that the command will appear at
// if it's ever committed. the second return value is the current
// term. the third return value is true if this server believes it is
// the leader.
//
func (rf *Raft) Start(command interface{}) (int, int, bool) {
	index := -1
	term := -1
	isLeader := true

	// Your code here (2B).

	return index, term, isLeader
}

func (rf *Raft) Kill() {
	atomic.StoreInt32(&rf.dead, 1)
	DPrintln("[", rf.me, "]", "Kill called....")
	// Your code here, if desired.
}

func (rf *Raft) killed() bool {
	DPrintln("[", rf.me, "]", "Killed called....")
	z := atomic.LoadInt32(&rf.dead)
	return z == 1
}

//
// the service or tester wants to create a Raft server. the ports
// of all the Raft servers (including this one) are in peers[]. this
// server's port is peers[me]. all the servers' peers[] arrays
// have the same order. persister is a place for this server to
// save its persistent state, and also initially holds the most
// recent saved state, if any. applyCh is a channel on which the
// tester or service expects Raft to send ApplyMsg messages.
// Make() must return quickly, so it should start goroutines
// for any long-running work.
//
func Make(peers []*labrpc.ClientEnd, me int,
	persister *Persister, applyCh chan ApplyMsg) *Raft {
	rf := &Raft{}
	rf.peers = peers
	rf.persister = persister
	rf.me = me
	rf.currentTerm = 0
	rf.votedFor = -1
	rf.commitIndex = 0
	rf.lastApplied = 0
	rf.ResetRpcTimer()
	rf.currentState = FOLLOWER
	rf.rfCond = sync.NewCond(&rf.mu)

	// initialize from state persisted before a crash
	rf.readPersist(persister.ReadRaftState())

	go rf.CheckAndKickOfLeaderElection()
	go rf.SendHeartBeats()

	return rf
}
