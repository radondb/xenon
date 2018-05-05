/*
 * Xenon
 *
 * Copyright 2018 The Xenon Authors.
 * Code is licensed under the GPLv3.
 *
 */

package raft

import (
	"fmt"
)

// log wrapper for raft
func (r *Raft) logMsg(format string, v ...interface{}) string {
	return fmt.Sprintf("%v[ID:%v, V:%v, E:%v].%v", r.state.String(), r.getID(), r.getViewID(), r.getEpochID(), fmt.Sprintf(format, v...))
}

// DEBUG level log.
func (r *Raft) DEBUG(format string, v ...interface{}) {
	r.log.Debug("%v", r.logMsg(format, v...))
}

// INFO level log.
func (r *Raft) INFO(format string, v ...interface{}) {
	r.log.Info("%v", r.logMsg(format, v...))
}

// WARNING level log.
func (r *Raft) WARNING(format string, v ...interface{}) {
	r.log.Warning("%v", r.logMsg(format, v...))
}

// ERROR level log.
func (r *Raft) ERROR(format string, v ...interface{}) {
	r.log.Error("%v", r.logMsg(format, v...))
}

// PANIC level log.
func (r *Raft) PANIC(format string, v ...interface{}) {
	r.log.Panic("%v", r.logMsg(format, v...))
}
