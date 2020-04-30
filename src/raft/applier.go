package raft

// import "../labrpc"
// import "time"
// import "fmt"

func (rf *Raft) Applier() {

	for {

		rf.mu.Lock()
		for rf.commitIndex <= rf.lastApplied {
			rf.rfCond.Wait()
		}

		for rf.lastApplied < rf.commitIndex {
			rf.lastApplied += 1
			log := rf.log[rf.lastApplied]
			// fmt.Println("[", rf.me, "]", "Applier ", "index =>", rf.lastApplied, "Logs =>", rf.log)

			rf.applyCh <- ApplyMsg{
				CommandValid: true,
				Command:      log.Command,
				CommandIndex: log.Index,
			}
		}
		rf.mu.Unlock()
	}

}
