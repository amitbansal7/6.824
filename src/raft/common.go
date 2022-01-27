package raft

import (
	"sync"
	"time"

	"6.824/labrpc"
)

// A Go object implementing a single Raft peer.
//
type Raft struct {
	mu        sync.Mutex          // Lock to protect shared access to this peer's state
	peers     []*labrpc.ClientEnd // RPC end points of all peers
	persister *Persister          // Object to hold this peer's persisted state
	me        int                 // this peer's index into peers[]
	dead      int32               // set by Kill()

	currentTerm    int
	lastReceivedAt time.Time
	currentState   int
	// votedFor       int
	lastVotedTerm int
}

const (
	FOLLOWER  = 0
	LEADER    = 1
	CANDIDATE = 2
)

func (rf *Raft) ResetRpcTimer() {
	rf.lastReceivedAt = time.Now()
}
