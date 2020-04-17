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
import crand "crypto/rand"
import "math/big"


// import "bytes"
// import "../labgob"

const (
	LastRpcTimeOut = time.Millisecond * 350 // if no communication is received, calls for an election
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

//
// A Go object implementing a single Raft peer.
//
type Raft struct {
	mu              sync.Mutex          // Lock to protect shared access to this peer's state
	peers           []*labrpc.ClientEnd // RPC end points of all peers
	persister       *Persister          // Object to hold this peer's persisted state
	me              int                 // this peer's index into peers[]
	dead            int32               // set by Kill()
	currentTerm     int
	votedFor        int
	commitIndex     int
	lastApplied     int
	nextIndex       []int
	matchIndex      []int
	lastRpcReceived time.Time
	currentState    int
	rfCond          *sync.Cond
	// Your data here (2A, 2B, 2C).
	// Look at the paper's Figure 2 for a description of what
	// state a Raft server must maintain.

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
	Entries      []int
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

func (rf *Raft) RequestVoteRpc(args *RequestVoteArgs, reply *RequestVoteReply) {
	// Your code here (2A, 2B).
	rf.mu.Lock()
	defer rf.mu.Unlock()
	term := rf.currentTerm
	DPrintln("[", rf.me, "]", "*********RequestVote received from", args.CandidateId, "by ", rf.me)

	DPrintln("[", rf.me, "]", "args.Term => ", args.Term, "term =>", term, "rf.votedFor =>", rf.votedFor, "args.LastLogIndex => ", args.LastLogIndex, "rf.commitIndex =>", rf.commitIndex)

	if args.Term < term {
		reply.VoteGranted = false
		reply.Term = rf.currentTerm
		rf.votedFor = -1
	} else if rf.votedFor == -1 || args.LastLogIndex >= rf.commitIndex {
		reply.VoteGranted = true
		DPrintln("[", rf.me, "]", rf.me, " Voting for ", args.CandidateId)
		// rf.currentTerm = args.Term
		rf.votedFor = args.CandidateId
		reply.Term = args.Term
	}
	DPrintln("[", rf.me, "]", "*********RequestVote Sent for", args.CandidateId, "by ", rf.me, "=>", reply)
}

func (rf *Raft) AppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply) {
	rf.mu.Lock()
	defer rf.mu.Unlock()

	DPrintln("[", rf.me, "]", "AppendEntries received from ", args.LeaderId, "by ", rf.me, "args.term => ", args.Term, "rf.currentTerm => ", rf.currentTerm)
	rf.lastRpcReceived = time.Now()

	if args.LeaderId != rf.me && args.Term >= rf.currentTerm {
		rf.currentState = FOLLOWER
		rf.currentTerm = args.Term
		rf.rfCond.Broadcast()
	}

	DPrintln("[", rf.me, "]", "Marking ", rf.me, "as a FOLLOWER")

	if args.LeaderId != rf.me && args.Term >= rf.currentTerm {
		reply.Success = true
	}
	reply.Term = rf.currentTerm

	DPrintln("[", rf.me, "]", "Receving HeartBeats from ")
}

//
// example code to send a RequestVote RPC to a server.
// server is the index of the target server in rf.peers[].
// expects RPC arguments in args.
// fills in *reply with RPC reply, so caller should
// pass &reply.
// the types of the args and reply passed to Call() must be
// the same as the types of the arguments declared in the
// handler function (including whether they are pointers).
//
// The labrpc package simulates a lossy network, in which servers
// may be unreachable, and in which requests and replies may be lost.
// Call() sends a request and waits for a reply. If a reply arrives
// within a timeout interval, Call() returns true; otherwise
// Call() returns false. Thus Call() may not return for a while.
// A false return can be caused by a dead server, a live server that
// can't be reached, a lost request, or a lost reply.
//
// Call() is guaranteed to return (perhaps after a delay) *except* if the
// handler function on the server side does not return.  Thus there
// is no need to implement your own timeouts around Call().
//
// look at the comments in ../labrpc/labrpc.go for more details.
//
// if you're having trouble getting RPC to work, check that you've
// capitalized all field names in structs passed over RPC, and
// that the caller passes the address of the reply struct with &, not
// the struct itself.
//
func (rf *Raft) sendRequestVote(server int, args *RequestVoteArgs, reply *RequestVoteReply) bool {
	// DPrintf("Sending RequestVote to %d", server)
	ok := rf.peers[server].Call("Raft.RequestVoteRpc", args, reply)
	return ok
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

//
// the tester doesn't halt goroutines created by Raft after each test,
// but it does call the Kill() method. your code can use killed() to
// check whether Kill() has been called. the use of atomic avoids the
// need for a lock.
//
// the issue is that long-running goroutines use memory and may chew
// up CPU time, perhaps causing later tests to fail and generating
// confusing debug output. any goroutine with a long-running loop
// should call killed() to check whether it should stop.
//
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
	time.Sleep(time.Millisecond * 60)
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

		max := big.NewInt(1000)
		rr, _ := crand.Int(crand.Reader, max)
		DPrintln("[", rf.me, "]", "Election paused for ", rr.Int64())
		time.Sleep(time.Duration(rr.Int64()) * time.Millisecond)

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
			time.Sleep(2000 * time.Millisecond)

		} else {
			rf.mu.Unlock()
			time.Sleep(timeOutAt.Sub(time.Now()))
		}
	}
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
	rf.lastRpcReceived = time.Now()
	rf.currentState = FOLLOWER
	rf.rfCond = sync.NewCond(&rf.mu)

	// initialize from state persisted before a crash
	rf.readPersist(persister.ReadRaftState())

	go rf.CheckAndKickOfLeaderElection()
	go rf.SendHeartBeats()

	return rf
}
