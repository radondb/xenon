/*
 * Xenon
 *
 * Copyright 2018 The Xenon Authors.
 * Code is licensed under the GPLv3.
 *
 */

package raft

import (
	"model"
)

// HARPC tuple.
type HARPC struct {
	raft *Raft
}

// HADisable rpc.
func (h *HARPC) HADisable(req *model.HARPCRequest, rsp *model.HARPCResponse) error {
	h.raft.WARNING("RPC.HADisable.call.from[%v]", req.GetFrom())

	// except state IDLE/STOPPED
	state := h.raft.getState()
	switch state {
	case IDLE:
		rsp.RetCode = model.OK
		return nil
	case STOPPED:
		rsp.RetCode = model.ErrorInvalidRequest
		return nil
	}
	h.raft.setState(IDLE)
	h.raft.loopFired()
	rsp.RetCode = model.OK
	return nil
}

// HASetLearner rpc.
func (h *HARPC) HASetLearner(req *model.HARPCRequest, rsp *model.HARPCResponse) error {
	h.raft.WARNING("RPC.HASetLearner.call.from[%v]", req.GetFrom())

	// except state STOPPED
	state := h.raft.getState()
	switch state {
	case STOPPED:
		rsp.RetCode = model.ErrorInvalidRequest
		return nil
	}
	h.raft.setState(LEARNER)
	h.raft.loopFired()
	rsp.RetCode = model.OK
	return nil
}

// HAEnable rpc.
func (h *HARPC) HAEnable(req *model.HARPCRequest, rsp *model.HARPCResponse) error {
	h.raft.WARNING("RPC.HAEnable.call.from[%v]", req.GetFrom())

	// expect state IDLE
	state := h.raft.getState()
	switch state {
	case IDLE:
		if h.raft.conf.SuperIDLE {
			// Set SuperIDLE to noLeader to fire the 'change master to'.
			h.raft.setLeader(noLeader)
		} else {
			h.raft.setState(FOLLOWER)
			h.raft.loopFired()
		}
		rsp.RetCode = model.OK
		return nil
	case LEARNER:
		h.raft.setState(FOLLOWER)
		h.raft.loopFired()
		rsp.RetCode = model.OK
		return nil
	case STOPPED:
		rsp.RetCode = model.ErrorInvalidRequest
		return nil
	}
	rsp.RetCode = model.OK
	return nil
}

// HATryToLeader rpc.
func (h *HARPC) HATryToLeader(req *model.HARPCRequest, rsp *model.HARPCResponse) error {
	h.raft.WARNING("RPC.HATryToLeader.call.from[%v]", req.GetFrom())

	// expect state FOLLOWER
	state := h.raft.getState()
	switch state {
	case LEADER:
		rsp.RetCode = model.OK
		return nil
	case CANDIDATE:
		rsp.RetCode = model.ErrorInvalidRequest
		return nil
	case IDLE:
		rsp.RetCode = model.ErrorInvalidRequest
		return nil
	case INVALID:
		rsp.RetCode = model.ErrorInvalidRequest
		return nil
	}
	// promotable cases:
	// 1. MySQL is MYSQL_ALIVE
	// 2. Slave_SQL_RNNNING is OK
	if h.raft.mysql.Promotable() {
		h.raft.WARNING("RPC.TryToLeader.promote.to.candidate")
		// stop io thread
		// it will re-start again when heartbeat received
		if err := h.raft.mysql.StopSlaveIOThread(); err != nil {
			h.raft.ERROR("RPC.TryToLeader.mysql.StopSlaveIOThread.error[%+v]", err)
			rsp.RetCode = err.Error()
			return nil
		}
		h.raft.setState(CANDIDATE)
		h.raft.loopFired()
		h.raft.IncCandidatePromotes()
	} else {
		rsp.RetCode = model.RPCError_MySQLUnpromotable
		return nil
	}
	rsp.RetCode = model.OK
	return nil
}

// GetHARPC returns HARPC.
func (s *Raft) GetHARPC() *HARPC {
	return &HARPC{s}
}
