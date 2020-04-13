package mr

//
// RPC definitions.
//
// remember to capitalize all names.
//

import (
	"os"
	"strconv"
	"time"
)

//
// example to show how to declare the arguments
// and reply for an RPC.
//

type ExampleArgs struct {
	X int
}

type ExampleReply struct {
	Y int
}

// Add your RPC definitions here.

type Work struct {
	Id      string
	Status  string // "idle", "done", "working"
	Timeout time.Time
	Task    *Task
}

type Task struct {
	Action         string
	File           string
	TempToResFiles map[string]string
}

type SyncResponse struct {
	NewWork *Work
	NReduce int
	MapDone bool
	AllDone bool
}

// Cook up a unique-ish UNIX-domain socket name
// in /var/tmp, for the master.
// Can't use the current directory since
// Athena AFS doesn't support UNIX-domain sockets.
func masterSock() string {
	s := "/var/tmp/824-mr-"
	s += strconv.Itoa(os.Getuid())
	return s
}
