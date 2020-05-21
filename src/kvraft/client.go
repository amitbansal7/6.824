package kvraft

import "../labrpc"
import "crypto/rand"
import "math/big"
import "fmt"
import "time"

type Clerk struct {
	servers  []*labrpc.ClientEnd
	leaderId int
	id       int64
	// You will have to modify this struct.
}

func nrand() int64 {
	max := big.NewInt(int64(1) << 62)
	bigx, _ := rand.Int(rand.Reader, max)
	x := bigx.Int64()
	return x
}

func MakeClerk(servers []*labrpc.ClientEnd) *Clerk {
	ck := new(Clerk)
	ck.servers = servers
	ck.id = nrand()
	// You'll have to add code here.
	return ck
}

//
// fetch the current value for a key.
// returns "" if the key does not exist.
// keeps trying forever in the face of all other errors.
//
// you can send an RPC with code like this:
// ok := ck.servers[i].Call("KVServer.Get", &args, &reply)
//
// the types of args and reply (including whether they are pointers)
// must match the declared types of the RPC handler function's
// arguments. and reply must be passed as a pointer.
//
func (ck *Clerk) Get(key string) string {

	args := GetArgs{
		Key:      key,
		ReqId:    nrand(),
		ClientId: ck.id,
	}

	leaderId := ck.leaderId
	timeoutTtries := 0
	wrongLeader := 0
	
	for {
		reply := GetReply{}
		if timeoutTtries > 50 {
			// fmt.Println("KVServer.Get try limit exceeded", args)
			// return ""
		}
		if wrongLeader % len(ck.servers) == 0 &&  wrongLeader > 100 {
			// fmt.Println("KVServer.Get try limit exceeded, couldn't connect to a leader", args)
			// return ""
		}

		if ok := ck.servers[leaderId].Call("KVServer.Get", &args, &reply); ok {
			switch reply.Err {
			case OK:
				ck.leaderId = leaderId
				return reply.Value
			case ErrNoKey:
				return ""
			case ErrWrongLeader:
				// fmt.Println("KVServer.Get ErrWrongLeader trying again", args)
				wrongLeader += 1
				if wrongLeader % len(ck.servers) == 0 {
					time.Sleep(100 * time.Millisecond)
				}
			case ErrTimeout:
				// fmt.Println("KVServer.Get ErrTimeout trying again", args)
				timeoutTtries += 1
				time.Sleep(100 * time.Millisecond)
			default:
				fmt.Println("Get Unknown", reply.Err)
			}
		} else {
			time.Sleep(100 * time.Millisecond)
			timeoutTtries += 1
			// fmt.Println("KVServer.Get req not ok", args)
		}
		leaderId = (leaderId + 1) % len(ck.servers)
	}

	// return ""
}

//
// shared by Put and Append.
//
// you can send an RPC with code like this:
// ok := ck.servers[i].Call("KVServer.PutAppend", &args, &reply)
//
// the types of args and reply (including whether they are pointers)
// must match the declared types of the RPC handler function's
// arguments. and reply must be passed as a pointer.
//
func (ck *Clerk) PutAppend(key string, value string, op string) {
	args := PutAppendArgs{
		Key:      key,
		Value:    value,
		Op:       op,
		ReqId:    nrand(),
		ClientId: ck.id,
	}

	leaderId := ck.leaderId
	timeoutTries := 0
	wrongLeader := 0

	for {
		reply := GetReply{}
		if timeoutTries > 50 {
			// fmt.Println("KVServer.PutAppend try limit exceeded", args)
			// return
		}
		if wrongLeader % len(ck.servers) == 0 &&  wrongLeader > 100 {
			// fmt.Println("KVServer.PutAppend try limit exceeded, couldn't connect to a leader", args)
			// return
		}
		// fmt.Println("KVServer.PutAppend Request to raft =>", leaderId)
		if ok := ck.servers[leaderId].Call("KVServer.PutAppend", &args, &reply); ok {
			switch reply.Err {
			case OK:
				ck.leaderId = leaderId
				return
			case ErrWrongLeader:
				// fmt.Println("KVServer.PutAppend ErrWrongLeader trying again", args)
				wrongLeader += 1
				if wrongLeader % len(ck.servers) == 0 {
					time.Sleep(100 * time.Millisecond)
				}
			case ErrTimeout:
				// fmt.Println("KVServer.PutAppend ErrTimeout trying again", args)
				timeoutTries += 1
				time.Sleep(100 * time.Millisecond)
			default:
				fmt.Println("PutAppend Unknown", reply.Err)
			}
		} else {
			// fmt.Println("KVServer.PutAppend req not ok", args)
			time.Sleep(100 * time.Millisecond)
			timeoutTries += 1
		}
		leaderId = (leaderId + 1) % len(ck.servers)
	}
}

func (ck *Clerk) Put(key string, value string) {
	ck.PutAppend(key, value, "Put")
}
func (ck *Clerk) Append(key string, value string) {
	ck.PutAppend(key, value, "Append")
}
