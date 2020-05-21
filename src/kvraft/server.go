package kvraft

import (
	"../labgob"
	"../labrpc"
	"../raft"
	"fmt"
	"log"
	// "reflect"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

const Debug = 0

func DPrintf(format string, a ...interface{}) (n int, err error) {
	if Debug > 0 {
		log.Printf(format, a...)
	}
	return
}

func DPrintln(a ...interface{}) (n int, err error) {
	if Debug > 0 {
		log.Println(a...)
	}
	return
}

const LockUnlockDebug = false

func (kv *KVServer) lock(by string) {
	kv.mu.Lock()
	if LockUnlockDebug {
		fmt.Println("Lock by", by)
	}
}

func (kv *KVServer) unlock(by string) {
	kv.mu.Unlock()
	if LockUnlockDebug {
		fmt.Println("UnLock by", by)
	}
}

type Op struct {
	// Your definitions here.
	// Field names must start with capital letters,
	// otherwise RPC will break.
	Key      string
	Value    string
	Opr      string // "Put" or "Append" or "Get"
	ReqId    int64
	ClientId int64
}

type CallStartReply struct {
	Err   Err
	Value string
}

type KVServer struct {
	mu      sync.Mutex
	me      int
	rf      *raft.Raft
	applyCh chan raft.ApplyMsg
	dead    int32 // set by Kill()

	maxraftstate int // snapshot if log grows this big

	replyMessages map[string](chan raft.ApplyMsg)
	data          map[string]string
	appliedToData map[string]bool

	// Your definitions here.
}

func (kv *KVServer) ReqKey(ClientId, ReqId int64) string {
	return strconv.FormatInt(ClientId, 10) + strconv.FormatInt(ReqId, 10)
}

func (kv *KVServer) Get(args *GetArgs, reply *GetReply) {
	DPrintln(kv.me, "Get received", args)
	op := Op{
		Opr:      GET,
		Key:      args.Key,
		ReqId:    args.ReqId,
		ClientId: args.ClientId,
	}

	callStartReply := kv.CallStart("Get", op, args.Key, args.ClientId, args.ReqId)

	reply.Err = callStartReply.Err
	reply.Value = callStartReply.Value
}

func (kv *KVServer) PutAppend(args *PutAppendArgs, reply *PutAppendReply) {
	DPrintln(kv.me, "PutAppend received", args)
	op := Op{
		Opr:      args.Op,
		Key:      args.Key,
		Value:    args.Value,
		ReqId:    args.ReqId,
		ClientId: args.ClientId,
	}

	callStartReply := kv.CallStart("PutAppend", op, args.Key, args.ClientId, args.ReqId)

	reply.Err = callStartReply.Err
}

func (kv *KVServer) CallStart(method string, op Op, key string, clientId int64, reqId int64) CallStartReply {
	reply := CallStartReply{}

	_, _, isLeader := kv.rf.Start(op)

	if !isLeader {
		reply.Err = ErrWrongLeader
		return reply
	}

	replyChan := make(chan raft.ApplyMsg, 2)
	reqKey := kv.ReqKey(clientId, reqId)
	kv.mu.Lock()
	kv.replyMessages[reqKey] = replyChan
	kv.mu.Unlock()

	ticker := time.NewTicker(500 * time.Millisecond)
	for {
		select {
		case applyMsg := <-replyChan:
			kv.lock(method)
			reply.Err = OK

			if method == GET {
				reply.Value = kv.data[key]
			}
			DPrintln(kv.me, method, applyMsg)
			delete(kv.replyMessages, reqKey)
			kv.unlock(method)
			return reply
		case <-ticker.C:
			// fmt.Println(method, "ApplyMsg wait ErrTimeout")
			reply.Err = ErrTimeout
			kv.lock(method)
			delete(kv.replyMessages, reqKey)
			kv.unlock(method)
			return reply
		}
	}
}

func (kv *KVServer) ReceiveApplyMsg() {
	for applyMsg := range kv.applyCh {
		kv.lock("ReceiveApplyMsg")
		switch op := applyMsg.Command.(type) {
		case Op:
			switch op.Opr {
			case PUT, APPEND:
				reqKey := kv.ReqKey(op.ClientId, op.ReqId)
				kv.ApplyReqToData(applyMsg)
				if ch, ok := kv.replyMessages[reqKey]; ok {
					ch <- applyMsg
				} else {
					// fmt.Println("Channel not found in replyMessages for ", applyMsg)
				}
			case GET:
				reqKey := kv.ReqKey(op.ClientId, op.ReqId)
				if ch, ok := kv.replyMessages[reqKey]; ok {
					ch <- applyMsg
				} else {
					// fmt.Println("Channel not found in replyMessages for ", applyMsg)
				}
			default:
				DPrintln("Invalid op.Opr", op.Opr)
			}
		}
		kv.unlock("ReceiveApplyMsg")
	}
}

func (kv *KVServer) ApplyReqToData(applyMsg raft.ApplyMsg) {
	switch op := applyMsg.Command.(type) {
	case Op:
		reqKey := kv.ReqKey(op.ClientId, op.ReqId)
		switch op.Opr {
		case PUT:
			// fmt.Println(kv.me, "Putting... key => ", op.Key, "Value =>", op.Value)
			if _, ok := kv.appliedToData[reqKey]; !ok {
				kv.data[op.Key] = op.Value
				kv.appliedToData[reqKey] = true
			} else {
				DPrintln("Request already applied", applyMsg)
				return
			}
		case APPEND:
			// fmt.Println(kv.me, "Appending... key => ", op.Key, "Value =>", op.Value)
			if _, ok := kv.appliedToData[reqKey]; !ok {
				prev := ""
				if val, found := kv.data[op.Key]; found {
					prev = val
				}
				kv.data[op.Key] = prev + op.Value
				kv.appliedToData[reqKey] = true
			} else {
				// fmt.Println("Request already applied", applyMsg)
				return
			}
		default:
			DPrintln("Invalid op.Op", op.Opr)
		}
	}

	// fmt.Println(kv.me, "Data =>", kv.data)
}

//
// the tester calls Kill() when a KVServer instance won't
// be needed again. for your convenience, we supply
// code to set rf.dead (without needing a lock),
// and a killed() method to test rf.dead in
// long-running loops. you can also add your own
// code to Kill(). you're not required to do anything
// about this, but it may be convenient (for example)
// to suppress debug output from a Kill()ed instance.
//
func (kv *KVServer) Kill() {
	atomic.StoreInt32(&kv.dead, 1)
	kv.rf.Kill()
	// Your code here, if desired.
}

func (kv *KVServer) killed() bool {
	z := atomic.LoadInt32(&kv.dead)
	return z == 1
}

//
// servers[] contains the ports of the set of
// servers that will cooperate via Raft to
// form the fault-tolerant key/value service.
// me is the index of the current server in servers[].
// the k/v server should store snapshots through the underlying Raft
// implementation, which should call persister.SaveStateAndSnapshot() to
// atomically save the Raft state along with the snapshot.
// the k/v server should snapshot when Raft's saved state exceeds maxraftstate bytes,
// in order to allow Raft to garbage-collect its log. if maxraftstate is -1,
// you don't need to snapshot.
// StartKVServer() must return quickly, so it should start goroutines
// for any long-running work.
//
func StartKVServer(servers []*labrpc.ClientEnd, me int, persister *raft.Persister, maxraftstate int) *KVServer {
	// call labgob.Register on structures you want
	// Go's RPC library to marshall/unmarshall.
	labgob.Register(Op{})

	kv := new(KVServer)
	kv.me = me
	kv.maxraftstate = maxraftstate

	// You may need initialization code here.

	kv.applyCh = make(chan raft.ApplyMsg)
	kv.rf = raft.Make(servers, me, persister, kv.applyCh)
	kv.replyMessages = make(map[string](chan raft.ApplyMsg))
	kv.data = make(map[string]string)
	kv.appliedToData = make(map[string]bool)

	go kv.ReceiveApplyMsg()

	// You may need initialization code here.

	return kv
}
