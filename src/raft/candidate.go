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
	"sync"
	"time"
)

// Candidate tuple.
type Candidate struct {
	*Raft

	// Used to wait for the async job done.
	wg sync.WaitGroup

	// candidate process heartbeat request handler
	processHeartbeatRequestHandler func(*model.RaftRPCRequest) *model.RaftRPCResponse

	// candidate process voterequest request handler
	processRequestVoteRequestHandler func(*model.RaftRPCRequest) *model.RaftRPCResponse

	// candiadte send requestvote request to other followers
	sendRequestVoteHandler func(chan *model.RaftRPCResponse)

	// candiadte process requestvote response
	processRequestVoteResponseHandler func(*int, *model.RaftRPCResponse, *bool)

	// candidate process ping request handler
	processPingRequestHandler func(*model.RaftRPCRequest) *model.RaftRPCResponse
}

// NewCandidate creates the new Candidate.
func NewCandidate(r *Raft) *Candidate {
	C := &Candidate{Raft: r}
	C.initHandlers()
	return C
}

// Loop used to start the loop of the state machine.
//--------------------------------------
// State Machine
//--------------------------------------
//                   get majority votes
// State1. CANDIDATE ------------------------> LEADER
//
//                   higher viewid/new leader
// State2. CANDIDATE ------------------------> FOLLOWER
//
//                   timeout
// State3. CANDIDATE ------------------------> CANDIDATE
//
func (r *Candidate) Loop() {
	r.stateInit()
	defer r.stateExit()

	// reset timeout
	r.resetElectionTimeout()
	r.resetCheckVotesTimeout()

	// broadcast voterequest
	voteGranted := 1
	respChan := make(chan *model.RaftRPCResponse, r.getAllMembers())
	r.sendRequestVoteHandler(respChan)

	switchMaster := false

	for r.getState() == CANDIDATE {
		select {
		case <-r.fired:
			r.WARNING("state.machine.loop.got.fired")
		case <-r.checkVotesTick.C:
			// in one checkvotes timeout,
			// if we granted majority votes and no **DENY**, we are the winner
			if voteGranted >= r.getQuorums() {
				r.WARNING("get.enough.votes[%v]/members[%v].become.leader", voteGranted, r.getMembers())

				// upgrade to LEADER
				r.upgradeToLeader()
			}
			r.resetCheckVotesTimeout()
		case <-r.electionTick.C:
			voteGranted = 1
			// broadcast voterequest
			respChan = make(chan *model.RaftRPCResponse, r.getAllMembers())
			r.sendRequestVoteHandler(respChan)

			// reset timeout
			r.resetCheckVotesTimeout()
			r.resetElectionTimeout()
		case rsp := <-respChan:
			r.processRequestVoteResponseHandler(&voteGranted, rsp, &switchMaster)
			members := r.getMembers()
			if voteGranted == members {
				r.WARNING("grants.unanimous.votes[%v]/members[%v].become.leader", voteGranted, members)

				// upgrade to LEADER
				r.upgradeToLeader()
			}
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
				r.ERROR("get.unknown.request[%v]", e.Type)
			}
		}
	}
}

// candidateProcessHeartbeatRequest
// EFFECT
// handles the heartbeat request from the leader
//
// MYSQL
// nop
//
// RETURN
// 1. ErrorInvalidRequest: the request.From is not a member of this cluster
// 2. ErrorInvalidViewID: request leader viewid is old, he is a stale leader
// 3. OK: new leader eggs, we downgrade to FOLLOWER
func (r *Candidate) processHeartbeatRequest(req *model.RaftRPCRequest) *model.RaftRPCResponse {
	rsp := model.NewRaftRPCResponse(model.OK)
	rsp.Raft.From = r.getID()
	rsp.Raft.ViewID = r.getViewID()
	rsp.Raft.EpochID = r.getEpochID()
	rsp.Raft.State = r.state.String()

	if !r.checkRequest(req) {
		rsp.RetCode = model.ErrorInvalidRequest
		return rsp
	}

	vidiff := (int)(r.getViewID() - req.GetViewID())
	switch {
	case vidiff > 0:
		r.ERROR("get.heartbeat.from[N:%v, V:%v, E:%v].stale.viewid.ret.invalidviewid", req.GetFrom(), req.GetViewID(), req.GetEpochID())
		rsp.Raft.Leader = r.getLeader()
		rsp.RetCode = model.ErrorInvalidViewID
	case vidiff <= 0:
		r.WARNING("get.heartbeat.from[N:%v, V:%v, E:%v].down.to.follower", req.GetFrom(), req.GetViewID(), req.GetEpochID())

		// just down to FOLLOWER
		r.degradeToFollower()
	}
	return rsp
}

// candidateProcessRequestVoteRequest
// EFFECT
// handles the requestvote request from other CANDIDATEs
//
// MYSQL
// nop
//
// RETURN
// 1. ErrorInvalidRequest: the request.From is not a member of this cluster
// 2. ErrorInvalidViewID: request viewid is old
// 3. ErrorInvalidGTID: the CANDIDATE has the smaller Read_Master_Log_Pos
// 4. OK: give a vote
func (r *Candidate) processRequestVoteRequest(req *model.RaftRPCRequest) *model.RaftRPCResponse {
	rsp := model.NewRaftRPCResponse(model.OK)
	rsp.Raft.From = r.getID()
	rsp.Raft.ViewID = r.getViewID()
	rsp.Raft.EpochID = r.getEpochID()
	rsp.Raft.State = r.state.String()

	if !r.checkRequest(req) {
		rsp.RetCode = model.ErrorInvalidRequest
		return rsp
	}

	r.WARNING("get.voterequest.from[%+v]", *req)
	// 1. check viewid(req.viewid < thisnode.viewid)
	{
		if req.GetViewID() < r.getViewID() {
			r.WARNING("get.requestvote.from[N:%v, V:%v, E:%v].stale.viewid", req.GetFrom(), req.GetViewID(), req.GetEpochID())
			rsp.RetCode = model.ErrorInvalidViewID
			return rsp
		}
	}

	// 2. check GTID
	{
		greater, thisGTID, err := r.mysql.GTIDGreaterThan(&req.GTID)
		if err != nil {
			r.ERROR("process.requestvote.get.gtid.error[%v].ret.ErrorMySQLDown", err)
			rsp.RetCode = model.ErrorMySQLDown
			return rsp
		}
		rsp.GTID = thisGTID

		if greater {
			// keep up with the latest viewid to get the latest nodes of data elected faster
			if req.GetViewID() > r.getViewID() {
				r.updateView(req.GetViewID(), noLeader)
			}

			// reject cases:
			// 1. I am promotable: I am alive and GTID greater than you
			if r.mysql.Promotable() {
				r.WARNING("get.requestvote.from[N:%v, V:%v, E:%v].stale.GTID", req.GetFrom(), req.GetViewID(), req.GetEpochID())
				rsp.RetCode = model.ErrorInvalidGTID
				return rsp
			}
		}
	}

	// 3. check viewid(req.viewid >= thisnode.viewid)
	// if the req.viewid is larger than this node, update the viewid
	// if the req.viewid is equal and we have voted for other one then
	// don't voted for this candidate
	{
		if req.GetViewID() > r.getViewID() {
			r.updateView(req.GetViewID(), noLeader)
			r.degradeToFollower()
		} else {
			if (r.votedFor != noVote) && (r.votedFor != req.GetFrom()) {
				r.WARNING("get.requestvote.from[N:%v, V:%v, E:%v].already.vote", req.GetFrom(), req.GetViewID(), req.GetEpochID())
				rsp.RetCode = model.ErrorVoteNotGranted
				return rsp
			}
		}
	}

	// 4. voted for this candidate
	r.votedFor = req.GetFrom()
	// 5. a loser
	r.degradeToFollower()
	return rsp
}

func (r *Candidate) sendRequestVote(respChan chan *model.RaftRPCResponse) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	r.incViewID()
	for _, peer := range r.peers {
		r.wg.Add(1)
		go func(peer *Peer) {
			defer r.wg.Done()
			r.WARNING("prepare.send.requestvote.to[%v]", peer.getID())
			peer.sendRequestVote(respChan)
			r.WARNING("send.requestvote.done.to[%v]", peer.getID())
		}(peer)
	}

	// Wait for the send requestvote work done.
	r.wg.Wait()
}

// Votes who comes from IDLE machine will be filitered out.
func (r *Candidate) processRequestVoteResponse(voteGranted *int, rsp *model.RaftRPCResponse, switchMaster *bool) {
	r.WARNING("get.vote.response.from[N:%+v, R:%v].rsp.gtid[%v].retcode[%v]", rsp.GetFrom(), rsp.Raft.State, rsp.GetGTID(), rsp.RetCode)
	switch rsp.RetCode {
	case model.OK:
		if rsp.Raft.State == IDLE.String() {
			return
		}
		*voteGranted++
		r.INFO("get.vote.response.from[N:%v, V:%v].ok.votegranted[%v].majoyrity[%v]", rsp.GetFrom(), rsp.GetViewID(), *voteGranted, r.getQuorums())
	case model.ErrorInvalidViewID:
		r.WARNING("get.vote.response.from[N:%v, V:%v].fail[ErrorInvalidViewID].downgrade.to.follower", rsp.GetFrom(), rsp.GetViewID())
		r.updateView(rsp.GetViewID(), noLeader)
		r.degradeToFollower()
		return
	case model.ErrorInvalidGTID:
		r.WARNING("get.vote.response.from[N:%v, V:%v].deny[ErrorInvalidGTID].downgrade.to.follower", rsp.GetFrom(), rsp.GetViewID())
		r.degradeToFollower()
		return
	case model.ErrorMySQLDown:
		peers := r.getMembers()
		r.WARNING("get.vote.response.from[N:%v, V:%v].error[ErrorMySQLDown].peers.number[%v]", rsp.GetFrom(), rsp.GetViewID(), peers)
		// If the pees less than 3 and the Seconds_Behind_master is 0, we grant the vote though the mysql is down.
		if peers < 3 {
			if *switchMaster {
				*voteGranted++
			} else {
				time.Sleep(time.Duration(r.conf.CandidateWaitFor2Nodes) * time.Millisecond)
				*switchMaster = true
			}
		}
		return
	default:
		// this error is not enough to make us downgrade, just catch it
		r.WARNING("get.vote.response.from[N:%v, V:%v].error[%v].but.not.downgrade.to.follower", rsp.GetFrom(), rsp.GetViewID(), rsp.RetCode)
		return
	}
}

func (r *Candidate) processPingRequest(req *model.RaftRPCRequest) *model.RaftRPCResponse {
	rsp := model.NewRaftRPCResponse(model.OK)
	rsp.Raft.State = r.state.String()
	return rsp
}

// candidateUpgradeToLeader
// 1. goto the LEADER state
// 2. start the vip for public rafts
func (r *Candidate) upgradeToLeader() {
	r.setState(LEADER)
	r.setLeader(r.getID())
	r.IncLeaderPromotes()
}

func (r *Candidate) degradeToFollower() {
	r.setState(FOLLOWER)
}

func (r *Candidate) stateInit() {
	// update begin
	r.updateStateBegin()
	r.WARNING("state.machine.run")
}

func (r *Candidate) stateExit() {
	r.WARNING("candidate.state.machine.exit")
}

// candidate handlers
func (r *Candidate) initHandlers() {
	r.setProcessHeartbeatRequestHandler(r.processHeartbeatRequest)
	r.setProcessRequestVoteRequestHandler(r.processRequestVoteRequest)

	// send vote requet
	r.setSendRequestVoteHandler(r.sendRequestVote)
	r.setProcessRequestVoteResponseHandler(r.processRequestVoteResponse)

	// ping request
	r.setProcessPingRequestHandler(r.processPingRequest)
}

// for tests
func (r *Candidate) setProcessHeartbeatRequestHandler(f func(*model.RaftRPCRequest) *model.RaftRPCResponse) {
	r.processHeartbeatRequestHandler = f
}

func (r *Candidate) setProcessRequestVoteRequestHandler(f func(*model.RaftRPCRequest) *model.RaftRPCResponse) {
	r.processRequestVoteRequestHandler = f
}

func (r *Candidate) setSendRequestVoteHandler(f func(chan *model.RaftRPCResponse)) {
	r.sendRequestVoteHandler = f
}

func (r *Candidate) setProcessRequestVoteResponseHandler(f func(*int, *model.RaftRPCResponse, *bool)) {
	r.processRequestVoteResponseHandler = f
}

func (r *Candidate) setProcessPingRequestHandler(f func(*model.RaftRPCRequest) *model.RaftRPCResponse) {
	r.processPingRequestHandler = f
}
