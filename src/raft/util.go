package raft

import "log"

// Debugging
const Debug = 1

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

func min(a, b int) int{
	if a < b {
		return a
	}else{
		return b
	}
}

func max(a, b int) int{
	if a > b {
		return a
	}else{
		return b
	}
}
