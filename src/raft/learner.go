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

// LEARNER is a special STATE with other FOLLOWER/CANDICATE/LEADER states.
// It is usually used as READ-ONLY but does not have RAFT features, such as
// LEADER election
// FOLLOWER promotion
//
// Because of we bring LEARNER state in RaftRPCResponse as vote-request response,
// the LEARNER vote will be filtered out by other CANDIDATEs.
// LEARNER is one member of a RAFT cluster but without the rights to vote.

// Learner tuple.
type Learner struct {
	*Raft

	// learner process heartbeat request handler
	processHeartbeatRequestHandler func(*model.RaftRPCRequest) *model.RaftRPCResponse

	// learner process voterequest request handler
	processRequestVoteRequestHandler func(*model.RaftRPCRequest) *model.RaftRPCResponse

	// learner process raft ping request handler
	processPingRequestHandler func(*model.RaftRPCRequest) *model.RaftRPCResponse
}

// NewLearner creates new learner.
func NewLearner(r *Raft) *Learner {
	B := &Learner{Raft: r}
	B.initHandlers()
	return B
}

// Loop used to start the loop of the state machine.
//--------------------------------------
// State Machine
//--------------------------------------
// in LEARNER state, we never do leader election
//
func (r *Learner) Loop() {
	// update begin
	r.updateStateBegin()
	r.stateInit()

	for r.getState() == LEARNER {
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
				r.ERROR("get.unknown.request[%v]", e.Type)
			}
		}
	}
}

// processHeartbeatRequest
// EFFECT
// handles the heartbeat request from the leader
// In LEARNER state, we only handle the master changed
//
func (r *Learner) processHeartbeatRequest(req *model.RaftRPCRequest) *model.RaftRPCResponse {
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
// LEARNER is special, it returns ErrorVoteNotGranted expect Request Denied
//
// RETURN
// 1. ErrorVoteNotGranted: dont give any vote
func (r *Learner) processRequestVoteRequest(req *model.RaftRPCRequest) *model.RaftRPCResponse {
	rsp := model.NewRaftRPCResponse(model.ErrorVoteNotGranted)
	rsp.Raft.From = r.getID()
	rsp.Raft.ViewID = r.getViewID()
	rsp.Raft.EpochID = r.getEpochID()
	rsp.Raft.State = r.state.String()

	if !r.checkRequest(req) {
		rsp.RetCode = model.ErrorInvalidRequest
		return rsp
	}
	return rsp
}

func (r *Learner) processPingRequest(req *model.RaftRPCRequest) *model.RaftRPCResponse {
	rsp := model.NewRaftRPCResponse(model.OK)
	rsp.Raft.State = r.state.String()
	return rsp
}

func (r *Learner) stateInit() {
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

// handlers
func (r *Learner) initHandlers() {
	r.setProcessHeartbeatRequestHandler(r.processHeartbeatRequest)
	r.setProcessRequestVoteRequestHandler(r.processRequestVoteRequest)
	r.setProcessPingRequestHandler(r.processPingRequest)
}

// for tests
func (r *Learner) setProcessHeartbeatRequestHandler(f func(*model.RaftRPCRequest) *model.RaftRPCResponse) {
	r.processHeartbeatRequestHandler = f
}

func (r *Learner) setProcessRequestVoteRequestHandler(f func(*model.RaftRPCRequest) *model.RaftRPCResponse) {
	r.processRequestVoteRequestHandler = f
}

func (r *Learner) setProcessPingRequestHandler(f func(*model.RaftRPCRequest) *model.RaftRPCResponse) {
	r.processPingRequestHandler = f
}
