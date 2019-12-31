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
	"xbase/xrpc"
)

// Peer tuple.
type Peer struct {
	raft             *Raft
	requestTimeout   int // peer client request timneout
	heartbeatTimeout int
	connectionStr    string // peer connection string
}

// NewPeer creates new Peer.
func NewPeer(raft *Raft, connectionStr string, requestTimeout int, heartbeatTimeout int) *Peer {
	return &Peer{
		raft:             raft,
		connectionStr:    connectionStr,
		requestTimeout:   requestTimeout,
		heartbeatTimeout: heartbeatTimeout,
	}
}

// sendHeartbeat
// send heartbeat rpc request
func (p *Peer) sendHeartbeat(c chan *model.RaftRPCResponse) {
	// response
	rsp := model.NewRaftRPCResponse(model.OK)

	// request body
	req := model.NewRaftRPCRequest()
	req.Raft.EpochID = p.raft.getEpochID()
	req.Raft.ViewID = p.raft.getViewID()
	req.Raft.From = p.raft.getID()
	req.Raft.To = p.getID()
	req.Raft.Leader = p.raft.getLeader()
	req.Peers = p.raft.getPeers()
	req.IdlePeers = p.raft.getIdlePeers()
	req.Repl = p.raft.mysql.GetRepl()
	req.GTID, _ = p.raft.mysql.GetGTID()

	client, cleanup, err := p.NewClient()
	if err != nil {
		p.raft.ERROR("send.heartbeat.to.peer[%v].new.client.error[%v]", p.getID(), err)
		rsp.RetCode = model.ErrorRPCCall
		c <- rsp
		return
	}
	defer cleanup()

	method := model.RPCRaftHeartbeat
	err = client.CallTimeout(p.requestTimeout, method, req, rsp)
	if err != nil {
		p.raft.ERROR("send.heartbeat.to[%v].client.call.error[%v]", p.getID(), err)
		rsp.RetCode = model.ErrorRPCCall
		c <- rsp
		return
	}
	c <- rsp
}

// sendRequestVote
// send vote rpc request
func (p *Peer) sendRequestVote(c chan *model.RaftRPCResponse) {
	var err error

	// response
	rsp := model.NewRaftRPCResponse(model.OK)

	// request body
	req := model.NewRaftRPCRequest()
	req.Raft.EpochID = p.raft.meta.EpochID
	req.Raft.ViewID = p.raft.meta.ViewID
	req.Raft.From = p.raft.getID()
	req.Raft.To = p.connectionStr
	req.Raft.Leader = p.raft.getLeader()
	req.GTID, err = p.raft.getGTID()
	if err != nil {
		p.raft.ERROR("send.requestvote.to.peer[%v].get.gtid.error[%v]", p.getID(), err)
		rsp.RetCode = model.ErrorMySQLDown
		c <- rsp
		return
	}
	p.raft.WARNING("send.requestvote.to.peer[%v].request.gtid[%v]", p.getID(), req.GTID)

	client, cleanup, err := p.NewClient()
	if err != nil {
		p.raft.ERROR("send.requestvote.to.peer[%v].new.client.error[%v]", p.getID(), err)
		rsp.RetCode = model.ErrorRPCCall
		c <- rsp
		return
	}
	defer cleanup()

	method := model.RPCRaftRequestVote
	err = client.CallTimeout(p.requestTimeout, method, req, rsp)
	if err != nil {
		p.raft.ERROR("send.requestvote.to.peer[%v].client.call.error[%v]", p.getID(), err)
		rsp.RetCode = model.ErrorRPCCall
		c <- rsp
		return
	}
	c <- rsp
}

// follower SendPing
func (p *Peer) SendPing(c chan *model.RaftRPCResponse) {
	// response
	rsp := model.NewRaftRPCResponse(model.OK)

	// request body
	req := model.NewRaftRPCRequest()

	client, cleanup, err := p.NewClient()
	if err != nil {
		p.raft.ERROR("send.ping.to.peer[%v].new.client.error[%v]", p.getID(), err)
		rsp.RetCode = model.ErrorRPCCall
		c <- rsp
		return
	}
	defer cleanup()

	method := model.RPCRaftPing
	err = client.CallTimeout(p.requestTimeout, method, req, rsp)
	if err != nil {
		p.raft.ERROR("send.ping.to.peer[%v].client.call.error[%v]", p.getID(), err)
		rsp.RetCode = model.ErrorRPCCall
		c <- rsp
		return
	}
	p.raft.WARNING("send.ping.to.peer[%v].client.call.ok.rsp[%v]", p.getID(), rsp)
	c <- rsp
}

// NewClient creates new client.
func (p *Peer) NewClient() (*xrpc.Client, func(), error) {
	client, err := xrpc.NewClient(p.connectionStr, p.requestTimeout)
	if err != nil {
		return nil, nil, err
	}
	return client, func() {
		client.Close()
	}, nil
}

// attributes
func (p *Peer) freePeer() {
	// nop
}

func (p *Peer) getID() string {
	return p.connectionStr
}
