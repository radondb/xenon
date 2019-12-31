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
	"sync/atomic"

	"github.com/pkg/errors"
)

const (
	noVote      = ""
	noLeader    = ""
	maxSendTime = 1000 // max send timeout 1sec
)

// State enum.
type State int

const (
	// FOLLOWER state.
	FOLLOWER State = 1 << iota

	// CANDIDATE state.
	CANDIDATE

	// LEADER state.
	LEADER

	// IDLE state.
	// neither process heartbeat nor voterequest(return ErrorInvalidRequest)
	IDLE

	// INVALID state.
	// neither process heartbeat nor voterequest(return ErrorInvalidRequest)
	INVALID

	// LEARNER state.
	LEARNER

	// STOPPED state.
	STOPPED
)

func (s State) String() string {
	switch s {
	case 1 << 0:
		return "FOLLOWER"
	case 1 << 1:
		return "CANDIDATE"
	case 1 << 2:
		return "LEADER"
	case 1 << 3:
		return "IDLE"
	case 1 << 4:
		return "INVALID"
	case 1 << 5:
		return "LEARNER"
	case 1 << 6:
		return "STOPPED"
	}
	return "UNKNOW"
}

const (
	// MsgNone type.
	MsgNone = iota + 1

	// MsgRaftHeartbeat type.
	MsgRaftHeartbeat

	// MsgRaftRequestVote type.
	MsgRaftRequestVote

	// MsgRaftPing type.
	MsgRaftPing
)

var (
	errStop = errors.New("raft.has.been.stopped")
	errSend = errors.New("raft.send.timeout")
)

// raft attributes
func (r *Raft) getState() State {
	return r.state
}

func (r *Raft) setState(state State) {
	r.setLeader(noLeader)
	r.state = state
}

func (r *Raft) getID() string {
	return r.id
}

func (r *Raft) getQuorums() int {
	return (len(r.meta.Peers) / 2) + 1
}

// all members include me and exclude idle nodes
func (r *Raft) getMembers() int {
	return len(r.meta.Peers)
}

// all members include me and idle nodes
func (r *Raft) getAllMembers() int {
	return len(r.meta.Peers) + len(r.meta.IdlePeers)
}

func (r *Raft) getPeers() []string {
	return r.meta.Peers
}

func (r *Raft) getIdlePeers() []string {
	return r.meta.IdlePeers
}

func (r *Raft) getAllPeers() []string {
	allPeers := r.meta.Peers
	allPeers = append(allPeers, r.meta.IdlePeers...)
	return allPeers
}

func (r *Raft) getElectionTimeout() int {
	return r.conf.ElectionTimeout
}

func (r *Raft) getHeartbeatTimeout() int {
	return r.conf.HeartbeatTimeout
}

func (r *Raft) incViewID() {
	atomic.AddUint64(&r.meta.ViewID, 1)
}

func (r *Raft) getViewID() uint64 {
	return atomic.LoadUint64(&r.meta.ViewID)
}

func (r *Raft) incEpochID() {
	atomic.AddUint64(&r.meta.EpochID, 1)
}

func (r *Raft) getEpochID() uint64 {
	return atomic.LoadUint64(&r.meta.EpochID)
}

func (r *Raft) getGTID() (model.GTID, error) {
	return r.mysql.GetGTID()
}

func (r *Raft) getLeader() string {
	return r.leader
}

func (r *Raft) setLeader(leader string) {
	r.leader = leader
}
