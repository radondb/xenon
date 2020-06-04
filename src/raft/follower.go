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
	"strings"
	"sync"
)

// Follower tuple.
type Follower struct {
	*Raft
	// Whether the 'change to master' is success or not when the new leader eggs.
	ChangeToMasterError bool

	// Used to wait for the async job done.
	wg sync.WaitGroup

	// follower process heartbeat request handler
	processHeartbeatRequestHandler func(*model.RaftRPCRequest) *model.RaftRPCResponse

	// follower process voterequest request handler
	processRequestVoteRequestHandler func(*model.RaftRPCRequest) *model.RaftRPCResponse

	// follower process raft ping request handler
	processPingRequestHandler func(*model.RaftRPCRequest) *model.RaftRPCResponse
}

// NewFollower creates new Follower.
func NewFollower(r *Raft) *Follower {
	F := &Follower{Raft: r}
	F.initHandlers()

	return F
}

// Loop used to start the loop of the state machine.
//--------------------------------------
// State Machine
//--------------------------------------
//   timeout and ping ack greater or equal to n/2+1
// State1. FOLLOWER ------------------> CANDIDATE
//
func (r *Follower) Loop() {
	r.stateInit()
	defer r.stateExit()

	r.resetElectionTimeout()
	for r.getState() == FOLLOWER {
		select {
		case <-r.fired:
			r.WARNING("state.machine.loop.got.fired")
		case <-r.electionTick.C:
			r.WARNING("timeout.to.do.new.election")
			// promotable cases:
			// 1. MySQL is MYSQL_ALIVE
			// 2. Slave_SQL_RNNNING is OK
			if !r.isBrainSplit && r.mysql.Promotable() {
				r.WARNING("timeout.and.ping.almost.node.successed.promote.to.candidate")
				r.upgradeToCandidate()
			}

			// reset timeout
			r.resetElectionTimeout()
		case e := <-r.c:
			switch e.Type {
			case MsgRaftHeartbeat:
				req := e.request.(*model.RaftRPCRequest)
				rsp := r.processHeartbeatRequestHandler(req)
				e.response <- rsp

				if rsp.RetCode != model.OK {
					r.WARNING("process.heartbeat.request.RetCode.not.OK:%+v", rsp.RetCode)
				}
				// reset timeout
				r.resetElectionTimeout()
			case MsgRaftRequestVote:
				req := e.request.(*model.RaftRPCRequest)
				rsp := r.processRequestVoteRequestHandler(req)
				e.response <- rsp

				// reset timeout
				if rsp.RetCode == model.OK {
					r.resetElectionTimeout()
				}
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

// followerProcessHeartbeatRequest
// EFFECT
// handles the heartbeat request from the leader
//
// MYSQL
// we should check mysql slave_io_thread is stopped(by requestvote) or not
// if stopped we start it
//
// RETURN
// 1. ErrorInvalidRequest: the request.From is not a member of this cluster
// 2. ErrorInvalidViewID: request leader viewid is old, he is a stale leader
// 3. OK: new leader eggs, we downgrade to FOLLOWER and do mysql change master
func (r *Follower) processHeartbeatRequest(req *model.RaftRPCRequest) *model.RaftRPCResponse {
	rsp := model.NewRaftRPCResponse(model.OK)
	rsp.Raft.From = r.getID()
	rsp.Raft.ViewID = r.getViewID()
	rsp.Raft.EpochID = r.getEpochID()
	rsp.Raft.State = r.state.String()
	rsp.Relay_Master_Log_File = r.mysql.RelayMasterLogFile()

	r.DEBUG("get.heartbeat.from[N:%v, V:%v, E:%v]...", req.GetFrom(), req.GetViewID(), req.GetEpochID())
	if !r.checkRequest(req) {
		rsp.RetCode = model.ErrorInvalidRequest
		return rsp
	}

	viewdiff := (int)(r.getViewID() - req.GetViewID())
	epochdiff := (int)(r.getEpochID() - req.GetEpochID())
	switch {
	case viewdiff > 0:
		r.ERROR("get.heartbeat.from[N:%v, V:%v, E:%v].stale.viewid.ret.ErrorInvalidViewID", req.GetFrom(), req.GetViewID(), req.GetEpochID())
		rsp.Raft.Leader = r.getLeader()
		rsp.RetCode = model.ErrorInvalidViewID

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
			gtid, err := r.mysql.GetGTID()
			if err == nil {
				r.WARNING("get.heartbeat.my.gtid.is:%v", gtid)
			}

			// before change master need check gtid, if local gtid bigger than remote gtid degrade to INVALID
			if r.getMembers() > 2 {
				r.degradeToInvalid(&gtid, &req.GTID)
			}

			r.WARNING("get.heartbeat.from[N:%v, V:%v, E:%v].change.mysql.master", req.GetFrom(), req.GetViewID(), req.GetEpochID())

			if err := r.mysql.ChangeMasterTo(&req.Repl); err != nil {
				r.ERROR("change.master.to[FROM:%v, GTID:%v].error[%v]", req.GetFrom(), req.GetRepl(), err)
				// ChangeToMasterError is true, means we can't promotable to CANDIDATE.
				r.ChangeToMasterError = true
				rsp.RetCode = model.ErrorChangeMaster

				// return
				return rsp
			}

			r.ChangeToMasterError = false
			r.leader = req.GetFrom()
			r.WARNING("get.heartbeat.change.to.the.new.master[%v].successed", req.GetFrom())
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

// followerProcessRequestVoteRequest
// EFFECT
// handles the requestvote request from other CANDIDATEs
//
// MYSQL
// stop mysql slave_io_thread to get a GTID coordinate of this view
//
// RETURN
// 1. ErrorInvalidRequest: the request.From is not a member of this cluster
// 2. ErrorInvalidViewID: request viewid is old
// 3. ErrorInvalidGTID: the CANDIDATE has the smaller Read_Master_Log_Pos
// 4. OK: give a vote
func (r *Follower) processRequestVoteRequest(req *model.RaftRPCRequest) *model.RaftRPCResponse {
	rsp := model.NewRaftRPCResponse(model.OK)
	rsp.Raft.From = r.getID()
	rsp.Raft.ViewID = r.getViewID()
	rsp.Raft.EpochID = r.getEpochID()
	rsp.Raft.State = r.state.String()

	if !r.checkRequest(req) {
		rsp.RetCode = model.ErrorInvalidRequest
		return rsp
	}

	r.WARNING("get.voterequest.from[%+v].request[%v]", req.GetFrom(), req.GetGTID())
	// 1. check viewid(req.viewid < thisnode.viewid)
	{
		if req.GetViewID() < r.getViewID() {
			r.WARNING("get.requestvote.from[N:%v, V:%v, E:%v].stale.viewid.ret.reject", req.GetFrom(), req.GetViewID(), req.GetEpochID())
			rsp.RetCode = model.ErrorInvalidViewID
			return rsp
		}
	}

	// 2. check GTID
	{
		// stop io thread
		// it will re-start again when heartbeat received
		if err := r.mysql.StopSlaveIOThread(); err != nil {
			r.ERROR("mysql.StopSlaveIOThread.error[%+v]", err)
		}

		greater, thisGTID, err := r.mysql.GTIDGreaterThan(&req.GTID)
		if err != nil {
			r.ERROR("process.requestvote.get.gtid.error[%v].ret.ErrorMySQLDown", err)
			rsp.RetCode = model.ErrorMySQLDown
			return rsp
		}
		rsp.GTID = thisGTID

		if greater {
			// reject cases:
			// 1. I am promotable: I am alive and GTID greater than you
			if r.mysql.Promotable() {
				r.WARNING("get.requestvote.from[N:%v, V:%v, E:%v].stale.ret.ErrorInvalidGTID", req.GetFrom(), req.GetViewID(), req.GetEpochID())
				rsp.RetCode = model.ErrorInvalidGTID
				return rsp
			}
		}
	}

	// 3. check viewid(req.viewid >= thisnode.viewid)
	// if the req.viewid is larger than or equal with this node, update the viewid
	// if the req.viewid is less than this node, we don't voted for other one then
	// voted for this candidate
	{
		if req.GetViewID() >= r.getViewID() {
			r.updateView(req.GetViewID(), noLeader)
		} else {
			if (r.votedFor != noVote) && (r.votedFor != req.GetFrom()) {
				r.WARNING("get.requestvote.from[N:%v, V:%v, E:%v].already.vote.ret.reject", req.GetFrom(), req.GetViewID(), req.GetEpochID())
				rsp.RetCode = model.ErrorVoteNotGranted
				return rsp
			}
		}
	}

	// 4. voted for this candidate
	r.votedFor = req.GetFrom()
	return rsp
}

func (r *Follower) processPingRequest(req *model.RaftRPCRequest) *model.RaftRPCResponse {
	rsp := model.NewRaftRPCResponse(model.OK)
	rsp.Raft.State = r.state.String()
	return rsp
}

// startCheckBrainSplit check for split brain
func (r *Follower) startCheckBrainSplit() {
	r.isBrainSplit = true
	r.INFO("start.CheckBrainSplit")

	cnt := 1
	respChan := make(chan *model.RaftRPCResponse, r.getMembers())
	r.resetCheckBrainSplitTimeout()
	go func() {
		for r.getState() == FOLLOWER {
			select {
			case <-r.fired:
				r.WARNING("check.brain.split.loop.got.fired")
			case <-r.checkBrainSplitTick.C:
				r.DEBUG("timeout.to.check.brain.split")

				if r.isBrainSplit {
					r.WARNING("ping.responses[%v].is.less.than.half.maybe.brain.split", cnt)
				}

				cnt = 1
				respChan = make(chan *model.RaftRPCResponse, r.getMembers())
				r.sendClusterPing(respChan)
				r.resetCheckBrainSplitTimeout()

			case rsp := <-respChan:
				if rsp.RetCode == model.OK {
					if rsp.Raft.State == "LEADER" {
						r.DEBUG("receive.responses.of.leader.skip.check.brain.split")
						continue
					}
					if strings.Contains("FOLLOWER CANDIDATE LEARNER", rsp.Raft.State) {
						r.DEBUG("receive.responses.of.%v.skip.check.brain.split", rsp.Raft.State)
						cnt++
					}
				}

				if cnt < r.GetQuorums() {
					r.isBrainSplit = true
					continue
				}

				if r.isBrainSplit {
					r.isBrainSplit = false
					r.WARNING("ping.responses[%v].is.greater.than.half.again", cnt)
				}
			}
		}
	}()
}

func (r *Follower) sendClusterPing(respChan chan *model.RaftRPCResponse) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	for _, peer := range r.peers {
		go func(peer *Peer) {
			peer.SendPing(respChan)
		}(peer)
	}
}

func (r *Follower) upgradeToCandidate() {
	// only you
	if len(r.peers) == 0 {
		r.WARNING("peers.is.null.can.not.upgrade.to.candidate")
		return
	}

	if r.ChangeToMasterError {
		r.WARNING("change.to.master.error.can.not.upgrade.to.candidate")
		return
	}

	// stop io thread
	// it will re-start again when heartbeat received
	if err := r.mysql.StopSlaveIOThread(); err != nil {
		r.ERROR("mysql.StopSlaveIOThread.error[%v]", err)
	}
	r.setState(CANDIDATE)
	r.IncCandidatePromotes()
}

func (r *Follower) degradeToInvalid(followerGTID *model.GTID, candidateGTID *model.GTID) {
	// only you
	if len(r.peers) == 0 {
		r.WARNING("peers.is.null.can.not.upgrade.to.candidate")
		return
	}

	// stop io thread
	// it will re-start again when heartbeat received
	if err := r.mysql.StopSlaveIOThread(); err != nil {
		r.ERROR("mysql.StopSlaveIOThread.error[%v]", err)
	}

	// if error can not vote candidate
	greater := r.mysql.CheckGTID(followerGTID, candidateGTID)
	if greater {
		// degrade to INVALID
		r.setState(INVALID)
		return
	}
}

// setMySQLAsync used to setting mysql in async
func (r *Follower) setMySQLAsync() {
	r.WARNING("mysql.waitMysqlDoneAsync.prepare")

	r.wg.Add(1)
	go func() {
		defer r.wg.Done()
		// MySQL1: set readonly
		if err := r.mysql.SetReadOnly(); err != nil {
			r.ERROR("mysql.SetReadOnly.error[%v]", err)
		}
		r.WARNING("mysql.SetReadOnly.done")

		// MySQL2. set mysql slave system variables
		if err := r.mysql.SetSlaveGlobalSysVar(); err != nil {
			r.ERROR("mysql.SetSlaveGlobalSysVar.error[%v]", err)
		}
		r.WARNING("mysql.SetSlaveGlobalSysVar.done")
		r.WARNING("prepareAsync.done")

		// Log the gtid info.
		if gtid, err := r.mysql.GetGTID(); err != nil {
			r.ERROR("init.get.mysql.gtid.error:%v", err)
		} else {
			r.WARNING("init.my.gtid.is:%v", gtid)
		}
	}()
}

func (r *Follower) stateInit() {
	r.WARNING("state.init")
	r.updateStateBegin()
	// 1. stop vip
	if err := r.leaderStopShellCommand(); err != nil {
		// TODO(array): what todo?
		r.ERROR("stopShellCommand.error[%v]", err)
	}
	r.setMySQLAsync()
	r.WARNING("state.machine.run")
}

func (r *Follower) stateExit() {
	// Wait for the FOLLOWER state-machine async work done.
	r.wg.Wait()
	r.WARNING("follower.state.machine.exit")
}

// follower handlers
func (r *Follower) initHandlers() {
	r.setProcessHeartbeatRequestHandler(r.processHeartbeatRequest)
	r.setProcessRequestVoteRequestHandler(r.processRequestVoteRequest)
	r.setProcessPingRequestHandler(r.processPingRequest)
}

// for tests
func (r *Follower) setProcessHeartbeatRequestHandler(f func(*model.RaftRPCRequest) *model.RaftRPCResponse) {
	r.processHeartbeatRequestHandler = f
}

func (r *Follower) setProcessRequestVoteRequestHandler(f func(*model.RaftRPCRequest) *model.RaftRPCResponse) {
	r.processRequestVoteRequestHandler = f
}

func (r *Follower) setProcessPingRequestHandler(f func(*model.RaftRPCRequest) *model.RaftRPCResponse) {
	r.processPingRequestHandler = f
}
