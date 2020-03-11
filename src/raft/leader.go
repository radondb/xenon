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
	"mysql"
	"strings"
	"sync"
	"time"
	"xbase/common"
)

// Leader tuple.
type Leader struct {
	*Raft
	// the smallest binlog which slaves executed by SQL-Thread
	relayMasterLogFile string

	// leader degrade to follower
	isDegradeToFollower bool

	// Used to wait for the async job done.
	wg sync.WaitGroup

	// the binlog which we should purge to
	nextPuregeBinlog string

	purgeBinlogTick   *time.Ticker
	checkSemiSyncTick *time.Ticker

	// leader process heartbeat request handler
	processHeartbeatRequestHandler func(*model.RaftRPCRequest) *model.RaftRPCResponse

	// leader process voterequest request handler
	processRequestVoteRequestHandler func(*model.RaftRPCRequest) *model.RaftRPCResponse

	// leader send heartbeat request to other followers
	sendHeartbeatHandler func(*bool, chan *model.RaftRPCResponse)

	// leader process send heartbeat response
	processHeartbeatResponseHandler func(*int, *model.RaftRPCResponse)

	// leader process ping request handler
	processPingRequestHandler func(*model.RaftRPCRequest) *model.RaftRPCResponse
}

const (
	semisyncTimeoutFor2Nodes = 300000              // 5 minutes
	semisyncTimeout          = 1000000000000000000 // for 3 or more nodes
)

// NewLeader creates new Leader.
func NewLeader(r *Raft) *Leader {
	L := &Leader{
		Raft: r,
	}
	L.initHandlers()
	return L
}

// Loop used to start the loop of the state machine.
//--------------------------------------
// State Machine
//--------------------------------------
//                  higher viewid
// State1. LEADER ------------------> FOLLOWER
//
func (r *Leader) Loop() {
	r.stateInit()
	defer r.stateExit()

	incViewID := false
	mysqlDown := false
	ackGranted := 1

	lessHtAcks := 0
	maxLessHtAcks := r.Raft.conf.AdmitDefeatHtCnt

	// send heartbeat
	respChan := make(chan *model.RaftRPCResponse, r.getAllMembers())
	r.sendHeartbeatHandler(&mysqlDown, respChan)
	r.resetHeartbeatTimeout()

	for r.getState() == LEADER {
		if mysqlDown {
			r.WARNING("feel.mysql.down.degrade.to.follower")
			r.degradeToFollower()
			break
		}

		select {
		case <-r.fired:
			r.WARNING("state.machine.loop.got.fired")
		case <-r.heartbeatTick.C:
			if ackGranted < r.getQuorums() {
				if r.getMembers() > 2 {
					lessHtAcks++
				}
				r.IncLessHeartbeatAcks()
				r.WARNING("heartbeat.acks.granted[%v].less.than.quorums[%v].lessHtAcks[%v].maxLessHtAcks[%v]", ackGranted, r.getQuorums(), lessHtAcks, maxLessHtAcks)
				if lessHtAcks >= maxLessHtAcks {
					r.WARNING("degrade.to.follower.lessHtAcks[%v]>=maxLessHtAcks[%v]", lessHtAcks, maxLessHtAcks)
					r.degradeToFollower()
					break
				}
			} else {
				lessHtAcks = 0

				// for brain split
				if ackGranted == r.getMembers() {
					if incViewID {
						r.WARNING("heartbeat.acks.granted[%v].equals.members[%v].again", ackGranted, r.getMembers())
						incViewID = false
					}
				} else if !incViewID {
					r.WARNING("heartbeat.acks.granted[%v].less.than.members[%v].for.the.first.time", ackGranted, r.getMembers())
					r.updateView(r.getViewID()+2, r.GetLeader())
					incViewID = true
				}
			}

			ackGranted = 1
			respChan = make(chan *model.RaftRPCResponse, r.getAllMembers())
			r.sendHeartbeatHandler(&mysqlDown, respChan)
			r.resetHeartbeatTimeout()
		case rsp := <-respChan:
			r.processHeartbeatResponseHandler(&ackGranted, rsp)
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
				rsp := r.processRequestVoteRequestHandler(req)
				e.response <- rsp
			// 3) Ping
			case MsgRaftPing:
				req := e.request.(*model.RaftRPCRequest)
				rsp := r.processPingRequestHandler(req)
				e.response <- rsp
			default:
				r.ERROR("get.unknown.request[%+v]", e.Type)
			}
		}
	}
}

// leaderProcessHeartbeatRequestHandler
// EFFECT
// handles the heartbeat request from the leader
//
// MYSQL
// nop
//
// RETURN
// 1. ErrorInvalidRequest: the request.From is not a member of this cluster
// 2. ErrorInvalidViewID: request leader viewid is old, he is a stale leader
// 3. OK: new leader eggs, we have to downgrade to FOLLOWER
func (r *Leader) processHeartbeatRequest(req *model.RaftRPCRequest) *model.RaftRPCResponse {
	rsp := model.NewRaftRPCResponse(model.OK)
	rsp.Raft.From = r.getID()
	rsp.Raft.ViewID = r.getViewID()
	rsp.Raft.EpochID = r.getEpochID()
	rsp.Raft.State = r.state.String()

	r.IncLeaderGetHeartbeatRequests()
	if !r.checkRequest(req) {
		rsp.RetCode = model.ErrorInvalidRequest
		return rsp
	}
	r.WARNING("get.heartbeat.from[%+v]", *req)
	vidiff := (int)(r.getViewID() - req.GetViewID())
	switch {
	case vidiff > 0:
		r.ERROR("get.heartbeat.from[N:%v, V:%v, E:%v].return.reject", req.GetFrom(), req.GetViewID(), req.GetEpochID())

		rsp.Raft.Leader = r.getLeader()
		rsp.RetCode = model.ErrorInvalidViewID

		// this case happens when two nodes all win in the same viewid
		// in the same viewid because we wait the 'reject' VOTE in random time
	case vidiff == 0:
		r.ERROR("get.heartbeat.from[N:%v, V:%v, E:%v].in.same.viewid", req.GetFrom(), req.GetViewID(), req.GetEpochID())

		// degrade to FOLLOWER
		r.degradeToFollower()

	// new leader eggs
	case vidiff < 0:
		r.WARNING("get.heartbeat.from[N:%v, V:%v, E:%v].down.follower", req.GetFrom(), req.GetViewID(), req.GetEpochID())

		// degrade to FOLLOWER
		r.degradeToFollower()
	}
	return rsp
}

// leaderProcessRequestVoteRequestHandler
// EFFECT
// process the requestvote request from other peer of this cluster
// in this case, some FOLLOWER can't get the leader's heartbeat
//
// MYSQL
// nop
//
// RETURNS
// 1. ErrorInvalidRequest: the request.From is not a member of this cluster
// 2. ErrorInvalidViewID: request viewid is old
// 3. ErrorInvalidGTID: the CANDIDATE has the smaller Read_Master_Log_Pos
// 4. OK: give a vote
func (r *Leader) processRequestVoteRequest(req *model.RaftRPCRequest) *model.RaftRPCResponse {
	rsp := model.NewRaftRPCResponse(model.OK)
	rsp.Raft.From = r.getID()
	rsp.Raft.ViewID = r.getViewID()
	rsp.Raft.EpochID = r.getEpochID()
	rsp.Raft.State = r.state.String()

	r.IncLeaderGetVoteRequests()
	if !r.checkRequest(req) {
		rsp.RetCode = model.ErrorInvalidRequest
		return rsp
	}

	r.WARNING("get.voterequest.from[%+v]", *req)

	// 1. check viewid
	//    request viewid is from an old view or equal with me, reject
	{
		if req.GetViewID() <= r.getViewID() {
			r.WARNING("get.requestvote.from[N:%v, V:%v, E:%v].stale.viewid", req.GetFrom(), req.GetViewID(), req.GetEpochID())
			rsp.RetCode = model.ErrorInvalidViewID
			return rsp
		}
	}

	// 2. check master GTID
	{
		greater, thisGTID, err := r.mysql.GTIDGreaterThan(&req.GTID)
		if err != nil {
			r.ERROR("process.requestvote.get.gtid.error[%v].ret.ErrorMySQLDown", err)
			rsp.RetCode = model.ErrorMySQLDown
			return rsp
		}
		rsp.GTID = thisGTID

		// if leader get a VoteRequest, the most likely reason MySQL doesn't work
		// 'greater' means that master binlog more than you
		if greater {
			// reject cases:
			// 1. I am promotable: I am alive and GTID greater than you
			if r.mysql.Promotable() {
				r.WARNING("get.requestvote.from[%v].stale.GTID[%+v]", req.GetFrom(), req.GetGTID())
				rsp.RetCode = model.ErrorInvalidGTID
				return rsp
			}
		}
	}

	// 3. update viewid, if Candidate viewid equal with Leader viewid don't update viewid
	{
		if req.GetViewID() > r.getViewID() {
			r.WARNING("get.requestvote.from[N:%v, V:%v, E:%v].degrade.to.follower", req.GetFrom(), req.GetViewID(), req.GetEpochID())
			r.updateView(req.GetViewID(), noLeader)
			// downgrade to FOLLOWER
			r.degradeToFollower()
		}
	}

	// 4. voted for this candidate
	r.votedFor = req.GetFrom()
	return rsp
}

// leaderSendHeartbeatHandler
// broadcast hearbeat requests to other peers of the cluster
func (r *Leader) sendHeartbeat(mysqlDown *bool, c chan *model.RaftRPCResponse) {
	// check MySQL down
	if r.mysql.GetState() == mysql.MysqlDead {
		*mysqlDown = true
		return
	}

	// broadcast heartbeat
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	allPeers := r.peers
	for k, peer := range r.idlePeers {
		allPeers[k] = peer
	}

	for _, peer := range allPeers {
		r.wg.Add(1)
		go func(peer *Peer) {
			defer r.wg.Done()
			peer.sendHeartbeat(c)
		}(peer)
	}
}

// leaderProcessHeartbeatResponseHandler
// process the send heartbeat response comes from other peers of the cluster
func (r *Leader) processHeartbeatResponse(ackGranted *int, rsp *model.RaftRPCResponse) {
	if rsp.RetCode != model.OK {
		r.ERROR("send.heartbeat.get.rsp[N:%v, V:%v, E:%v].error[%v]", rsp.GetFrom(), rsp.GetViewID(), rsp.GetEpochID(), rsp.RetCode)

		if rsp.RetCode == model.ErrorInvalidViewID {
			r.WARNING("send.heartbeat.get.rsp[N:%v, V:%v, E:%v].error[%v].degrade.to.follower", rsp.GetFrom(), rsp.GetViewID(), rsp.GetEpochID(), rsp.RetCode)
			// downgrade to FOLLOWER
			r.degradeToFollower()
		}
	} else {
		if rsp.Raft.State != IDLE.String() {
			*ackGranted++
		}
		// find the smallest binlog
		if r.relayMasterLogFile == "" {
			r.relayMasterLogFile = rsp.Relay_Master_Log_File
		} else if strings.Compare(r.relayMasterLogFile, rsp.Relay_Master_Log_File) > 0 {
			r.relayMasterLogFile = rsp.Relay_Master_Log_File
		}

		// to reset nextPuregeBinlog:
		// we must get all responses from the follower(s) and idle(s)
		// imagine that:
		// Master is doing backup for Slave2 restore
		// Master purged to Slave1-Relay_Master_Log_File
		// Slave2 starts up and can't find the binlog which she is want
		if *ackGranted == r.getMembers() {
			r.nextPuregeBinlog = r.relayMasterLogFile
		}
	}
}

func (r *Leader) processPingRequest(req *model.RaftRPCRequest) *model.RaftRPCResponse {
	rsp := model.NewRaftRPCResponse(model.OK)
	rsp.Raft.State = r.state.String()
	return rsp
}

func (r *Leader) degradeToFollower() {
	r.WARNING("degrade.to.follower.stop.the.vip...")
	if err := r.leaderStopShellCommand(); err != nil {
		r.ERROR("stopshell.error[%v]", err)
	}

	r.purgeBinlogStop()
	r.checkSemiSyncStop()
	r.IncLeaderDegrades()
	r.setState(FOLLOWER)
	r.isDegradeToFollower = true
}

// prepareSettingsAsync
// wait mysql WAIT_UNTIL_SQL_THREAD_AFTER_GTIDS done and other mysql settings
// since leader must periodically send heartbeat to followers, so setMysqlAsync is asynchronous
// WAIT_UNTIL_SQL_THREAD_AFTER_GTIDS maybe a long operation here
func (r *Leader) prepareSettingsAsync() {
	r.WARNING("async.setting.prepare....")

	r.wg.Add(1)
	go func() {
		defer r.wg.Done()
		gtid, err := r.mysql.GetGTID()
		if err != nil {
			r.ERROR("mysql.get.gtid.error[%v]", err)
		} else {
			r.WARNING("my.gtid.is:%v", gtid)
		}

		// MySQL1. wait relay log replay done
		r.WARNING("1. mysql.WaitUntilAfterGTID.prepare")
		r.SetRaftMysqlStatus(model.RAFTMYSQL_WAITUNTILAFTERGTID)
		if err := r.mysql.WaitUntilAfterGTID(gtid.Retrieved_GTID_Set); err != nil {
			r.ERROR("mysql.WaitUntilAfterGTID.error[%v]", err)
			r.setState(FOLLOWER)
			r.isDegradeToFollower = true
			return
		}
		r.ResetRaftMysqlStatus()
		r.WARNING("mysql.WaitUntilAfterGTID.done")

		// MySQL2. change to master
		r.WARNING("2. mysql.ChangeToMaster.prepare")
		if err := r.mysql.ChangeToMaster(); err != nil {
			r.ERROR("mysql.ChangeToMaster.error[%v]", err)
			r.setState(FOLLOWER)
			r.isDegradeToFollower = true
			return
		}
		r.WARNING("mysql.ChangeToMaster.done")

		// MySQL3. enable semi-sync on master
		// wait slave ack
		r.WARNING("3. mysql.EnableSemiSyncMaster.prepare")
		if err := r.mysql.EnableSemiSyncMaster(); err != nil {
			// WTF, what can we do?
			r.ERROR("mysql.EnableSemiSyncMaster.error[%v]", err)
		}
		r.WARNING("mysql.EnableSemiSyncMaster.done")

		// MySQL4. set mysql master system variables
		r.WARNING("4.mysql.SetSysVars.prepare")
		r.mysql.SetMasterGlobalSysVar()
		r.WARNING("mysql.SetSysVars.done")

		// MySQL5. set mysql to read/write
		r.WARNING("5. mysql.SetReadWrite.prepare")
		if err := r.mysql.SetReadWrite(); err != nil {
			// WTF, what can we do?
			r.ERROR("mysql.SetReadWrite.error[%v]", err)
		}
		r.WARNING("mysql.SetReadWrite.done")
		r.WARNING("6. start.vip.prepare")
		if err := r.leaderStartShellCommand(); err != nil {
			// TODO(array): what todo?
			r.ERROR("leader.StartShellCommand.error[%v]", err)
		}
		r.WARNING("start.vip.done")
		r.WARNING("async.setting.all.done....")
	}()
}

func (r *Leader) purgeBinlogStart() {
	r.purgeBinlogTick = common.NormalTicker(r.conf.PurgeBinlogInterval)
	go func(leader *Leader) {
		for range leader.purgeBinlogTick.C {
			leader.purgeBinlog()
		}
	}(r)
	r.INFO("purge.bing.start[%vms]...", r.conf.PurgeBinlogInterval)
}

func (r *Leader) purgeBinlogStop() {
	r.relayMasterLogFile = ""
	r.nextPuregeBinlog = ""
	r.purgeBinlogTick.Stop()
}

func (r *Leader) purgeBinlog() {
	if r.skipPurgeBinlog {
		r.WARNING("purge.binlog.skipped[skipPurgeBinlog is true]")
		return
	}

	if r.conf.PurgeBinlogDisabled {
		r.WARNING("purge.binlog.skipped[conf.PurgeBinlogDisabled is true]")
		return
	}

	if r.nextPuregeBinlog != "" {
		if err := r.mysql.PurgeBinlogsTo(r.nextPuregeBinlog); err != nil {
			r.ERROR("purge.binlogs.to[%v].error[%v]", r.nextPuregeBinlog, err)
			r.IncLeaderPurgeBinlogFails()
		} else {
			r.WARNING("purged.binlogs.to[%v]...", r.nextPuregeBinlog)
			r.relayMasterLogFile = ""
			r.nextPuregeBinlog = ""
			r.IncLeaderPurgeBinlogs()
		}
	}
}

func (r *Leader) checkSemiSyncStart() {
	interval := 5000
	r.checkSemiSyncTick = common.NormalTicker(interval)
	go func(leader *Leader) {
		for range leader.checkSemiSyncTick.C {
			leader.checkSemiSync()
		}
	}(r)
	r.INFO("check.semi-sync.thread.start[%vms]...", interval)
}

func (r *Leader) checkSemiSyncStop() {
	r.checkSemiSyncTick.Stop()
	r.INFO("check.semi-sync.thread.stop...")
}

// Disable the semi-sync if the nodes number less than 3.
func (r *Leader) checkSemiSync() {
	if r.skipCheckSemiSync {
		r.WARNING("check.semi-sync.skipped[skipCheckSemiSync is true]")
		return
	}

	min := 3
	cur := r.getMembers()
	if cur < min {
		if err := r.mysql.SetSemiSyncMasterTimeout(semisyncTimeoutFor2Nodes); err != nil {
			r.ERROR("mysql.set.semi-sync.master.timeout.to.default.error[%v]", err)
		}
	} else {
		if err := r.mysql.EnableSemiSyncMaster(); err != nil {
			r.ERROR("mysql.enable.semi-sync.error[%v]", err)
		}
		if err := r.mysql.SetSemiWaitSlaveCount((cur - 1) / 2); err != nil {
			r.ERROR("mysql.set.semi.wait.slave.count.error[%v]", err)
		}
		if err := r.mysql.SetSemiSyncMasterTimeout(semisyncTimeout); err != nil {
			r.ERROR("mysql.set.semi.sync.master.timeout.to.infinite.error[%v]", err)
		}
	}
}

func (r *Leader) stateInit() {
	r.WARNING("state.init")
	r.updateStateBegin()
	r.purgeBinlogStart()
	r.checkSemiSyncStart()
	r.prepareSettingsAsync()
	r.isDegradeToFollower = false

	r.WARNING("state.machine.run")
}

func (r *Leader) stateExit() {
	if !r.isDegradeToFollower {
		r.WARNING("state.machine.exit.stop.the.vip...")
		if err := r.leaderStopShellCommand(); err != nil {
			r.ERROR("stopshell.error[%v]", err)
		}

		r.purgeBinlogStop()
		r.checkSemiSyncStop()
	}
	// Wait for the LEADER state-machine async work done.
	r.wg.Wait()
	r.WARNING("leader.state.machine.exit.done")
}

// leader handlers
func (r *Leader) initHandlers() {
	// heartbeat request
	r.setProcessHeartbeatRequestHandler(r.processHeartbeatRequest)

	// vote request
	r.setProcessRequestVoteRequestHandler(r.processRequestVoteRequest)

	// send heartbeat
	r.setSendHeartbeatHandler(r.sendHeartbeat)
	r.setProcessHeartbeatResponseHandler(r.processHeartbeatResponse)

	// ping request
	r.setProcessPingRequestHandler(r.processPingRequest)
}

// for tests
func (r *Leader) setProcessHeartbeatRequestHandler(f func(*model.RaftRPCRequest) *model.RaftRPCResponse) {
	r.processHeartbeatRequestHandler = f
}

func (r *Leader) setProcessRequestVoteRequestHandler(f func(*model.RaftRPCRequest) *model.RaftRPCResponse) {
	r.processRequestVoteRequestHandler = f
}

func (r *Leader) setSendHeartbeatHandler(f func(*bool, chan *model.RaftRPCResponse)) {
	r.sendHeartbeatHandler = f
}

func (r *Leader) setProcessHeartbeatResponseHandler(f func(*int, *model.RaftRPCResponse)) {
	r.processHeartbeatResponseHandler = f
}

func (r *Leader) setProcessPingRequestHandler(f func(*model.RaftRPCRequest) *model.RaftRPCResponse) {
	r.processPingRequestHandler = f
}
