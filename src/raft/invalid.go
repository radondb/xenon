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

// Invalid is a special STATE with other FOLLOWER/CANDICATE/LEADER states.
// But it seems like IDLE state.
// It is usually used as READ-ONLY but does not have RAFT features, such as
// LEADER election
// FOLLOWER promotion
//
// IDLE is one member of a RAFT cluster but without the rights to vote and return ErrorInvalidRequest to CANDICATEs

// Invalid tuple.
type Invalid struct {
	*Raft

	// Invalid process heartbeat request handler
	processHeartbeatRequestHandler func(*model.RaftRPCRequest) *model.RaftRPCResponse

	// Invalid process voterequest request handler
	processRequestVoteRequestHandler func(*model.RaftRPCRequest) *model.RaftRPCResponse

	// Invalid process ping request handler
	processPingRequestHandler func(*model.RaftRPCRequest) *model.RaftRPCResponse
}

// NewInvalid creates new Invalid.
func NewInvalid(r *Raft) *Invalid {
	IV := &Invalid{Raft: r}
	IV.initHandlers()
	return IV
}

// Loop used to start the loop of the state machine.
//--------------------------------------
// State Machine
//--------------------------------------
// in INVALID state, we never do leader election
//
func (r *Invalid) Loop() {
	// update begin
	r.updateStateBegin()
	r.stateInit()

	for r.getState() == INVALID {
		select {
		case <-r.fired:
			r.WARNING("state.machine.loop.got.fired")
		case e := <-r.c:
			switch e.Type {
			// 1) Heartbeat
			case MsgRaftHeartbeat:
				req := e.request.(*model.RaftRPCRequest)
				rsp := r.processHeartbeatRequestHandler(req)
				e.response <- rsp

			// 2) RequestVote
			case MsgRaftRequestVote:
				req := e.request.(*model.RaftRPCRequest)
				rsp := r.processRequestVoteRequest(req)
				e.response <- rsp

			// 3) Ping
			case MsgRaftPing:
				req := e.request.(*model.RaftRPCRequest)
				rsp := r.processPingRequestHandler(req)
				e.response <- rsp
			default:
				r.ERROR("get.unknow.request[%v].[%v]", r.getID(), e.Type)
			}
		}
	}
}

// processHeartbeatRequest
// EFFECT
// handles the heartbeat request from the leader
// In Invalid state, we only handle the master changed
//
func (r *Invalid) processHeartbeatRequest(req *model.RaftRPCRequest) *model.RaftRPCResponse {
	rsp := model.NewRaftRPCResponse(model.OK)
	rsp.Raft.From = r.getID()
	rsp.Raft.ViewID = r.getViewID()
	rsp.Raft.EpochID = r.getEpochID()
	rsp.Raft.State = r.state.String()
	rsp.Relay_Master_Log_File = r.mysql.RelayMasterLogFile()

	if !r.checkRequest(req) {
		rsp.RetCode = model.ErrorInvalidRequest
		return rsp
	}

	viewdiff := (int)(r.getViewID() - req.GetViewID())
	epochdiff := (int)(r.getEpochID() - req.GetEpochID())
	switch {
	case viewdiff <= 0:
		// MySQL1: disable master semi-sync because I am a slave
		if err := r.mysql.DisableSemiSyncMaster(); err != nil {
			r.ERROR("mysql.DisableSemiSyncMaster.error[%v]", err)
		}

		// MySQL2: set mysql readonly(mysql maybe down and up then the LEADER changes)
		if err := r.mysql.SetReadOnly(); err != nil {
			r.ERROR("mysql.SetReadOnly.error[%v]", err)
		}

		// MySQL3: start slave
		if err := r.mysql.StartSlave(); err != nil {
			r.ERROR("mysql.StartSlave.error[%v]", err)
		}

		// MySQL4: change master
		if r.getLeader() != req.GetFrom() {
			r.WARNING("get.heartbeat.from[N:%v, V:%v, E:%v].change.mysql.master[%+v]", req.GetFrom(), req.GetViewID(), req.GetEpochID(), req.GetGTID())

			if err := r.mysql.ChangeMasterTo(&req.Repl); err != nil {
				r.ERROR("change.master.to[FROM:%v, GTID:%v].error[%v]", req.GetFrom(), req.GetRepl(), err)
				rsp.RetCode = model.ErrorChangeMaster
				return rsp
			}
			r.leader = req.GetFrom()
		}

		// view change
		if viewdiff < 0 {
			r.WARNING("get.heartbeat.from[N:%v, V:%v, E:%v].update.view", req.GetFrom(), req.GetViewID(), req.GetEpochID())
			r.updateView(req.GetViewID(), req.GetFrom())
		}

		// epoch change
		if epochdiff != 0 {
			r.WARNING("get.heartbeat.from[N:%v, V:%v, E:%v].update.epoch", req.GetFrom(), req.GetViewID(), req.GetEpochID())
			r.updateEpoch(req.GetEpochID(), req.GetPeers(), req.GetIdlePeers())
		}
	}
	return rsp
}

// processRequestVoteRequest
// EFFECT
// handles the requestvote request from other CANDIDATEs
// Invalid is special, it returns ErrorInvalidRequest
//
// RETURN
// 1. ErrorInvalidRequest: do not give a vote
func (r *Invalid) processRequestVoteRequest(req *model.RaftRPCRequest) *model.RaftRPCResponse {
	rsp := model.NewRaftRPCResponse(model.ErrorInvalidRequest)
	rsp.Raft.From = r.getID()
	rsp.Raft.ViewID = r.getViewID()
	rsp.Raft.EpochID = r.getEpochID()
	rsp.Raft.State = r.state.String()

	return rsp
}

func (r *Invalid) stateInit() {
	// 1. stop vip
	if err := r.leaderStopShellCommand(); err != nil {
		// TODO(array): what todo?
		r.ERROR("stopshell.error[%v]", err)
	}

	// MySQL1: set readonly
	if err := r.mysql.SetReadOnly(); err != nil {
		r.ERROR("mysql.SetReadOnly.error[%v]", err)
	}

	// MySQL2. set mysql slave system variables
	if err := r.mysql.SetSlaveGlobalSysVar(); err != nil {
		r.ERROR("mysql.SetSlaveGlobalSysVar.error[%v]", err)
	}
}

func (r *Invalid) processPingRequest(req *model.RaftRPCRequest) *model.RaftRPCResponse {
	rsp := model.NewRaftRPCResponse(model.OK)
	rsp.Raft.State = r.state.String()

	return rsp
}

// handlers
func (r *Invalid) initHandlers() {
	r.setProcessHeartbeatRequestHandler(r.processHeartbeatRequest)
	r.setProcessRequestVoteRequestHandler(r.processRequestVoteRequest)
	r.setProcessPingRequestHandler(r.processPingRequest)
}

// for tests
func (r *Invalid) setProcessHeartbeatRequestHandler(f func(*model.RaftRPCRequest) *model.RaftRPCResponse) {
	r.processHeartbeatRequestHandler = f
}

func (r *Invalid) setProcessRequestVoteRequestHandler(f func(*model.RaftRPCRequest) *model.RaftRPCResponse) {
	r.processRequestVoteRequestHandler = f
}

func (r *Invalid) setProcessPingRequestHandler(f func(*model.RaftRPCRequest) *model.RaftRPCResponse) {
	r.processPingRequestHandler = f
}
