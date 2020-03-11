/*
 * Xenon
 *
 * Copyright 2018 The Xenon Authors.
 * Code is licensed under the GPLv3.
 *
 */

package raft

import (
	"config"
	"model"
	"mysql"
	"os"
	"path/filepath"
	"sync"
	"time"
	"xbase/common"
	"xbase/xlog"
)

const (
	// metaFile is the file for storing raft metadata
	metaFile = "peers.json"
)

type ev struct {
	Type     int
	request  interface{}
	response chan interface{}
}

// RaftMeta tuple.
type RaftMeta struct {
	ViewID  uint64
	EpochID uint64

	// The Peers(endpoint) expect SuperIDLE
	Peers []string

	// The SuperIDLE Peers(endpoint)
	IdlePeers []string
}

// Raft tuple.
type Raft struct {
	log                 *xlog.Log
	mysql               *mysql.Mysql
	cmd                 common.Command
	conf                *config.RaftConfig
	leader              string
	votedFor            string
	id                  string
	fired               chan bool
	state               State
	meta                *RaftMeta
	mutex               sync.RWMutex
	lock                sync.WaitGroup
	heartbeatTick       *time.Timer
	electionTick        *time.Timer
	checkBrainSplitTick *time.Timer
	checkVotesTick      *time.Timer
	stateBegin          time.Time
	c                   chan *ev
	L                   *Leader
	C                   *Candidate
	F                   *Follower
	I                   *Idle
	IV                  *Invalid
	LN                  *Learner
	peers               map[string]*Peer // all peers expect SuperIDLE
	idlePeers           map[string]*Peer // all SuperIDLE peers
	stats               model.RaftStats
	skipPurgeBinlog     bool // if true, purge binlog will skipped
	skipCheckSemiSync   bool // if true, check semi-sync will skipped
	isBrainSplit        bool // if true, follower can upgrade to candidate
}

// NewRaft creates the new raft.
func NewRaft(id string, conf *config.RaftConfig, log *xlog.Log, mysql *mysql.Mysql) *Raft {
	r := &Raft{
		id:                id,
		conf:              conf,
		log:               log,
		cmd:               common.NewLinuxCommand(log),
		mysql:             mysql,
		leader:            noLeader,
		state:             FOLLOWER,
		meta:              &RaftMeta{},
		peers:             make(map[string]*Peer),
		idlePeers:         make(map[string]*Peer),
		skipCheckSemiSync: false,
	}

	// state handler
	r.L = NewLeader(r)
	r.C = NewCandidate(r)
	r.F = NewFollower(r)
	r.I = NewIdle(r)
	r.IV = NewInvalid(r)
	r.LN = NewLearner(r)

	// setup raft timeout
	r.resetHeartbeatTimeout()
	r.resetElectionTimeout()
	r.resetCheckVotesTimeout()

	// setup peers
	r.initPeers()

	// setup meta datadir
	if err := os.MkdirAll(r.conf.MetaDatadir, 0777); err != nil {
		log.Panic("create.meta.dir[%v].error[%v]", r.conf.MetaDatadir, err)
	}
	return r
}

// Start used to start the raft.
func (r *Raft) Start() error {
	// channels
	r.fired = make(chan bool)
	r.c = make(chan *ev)

	// state
	if r.conf.SuperIDLE {
		r.setState(IDLE)
		r.WARNING("start.as.super.IDLE")
	} else {
		r.setState(FOLLOWER)
	}

	// state loops
	r.lock.Add(1)
	go func() {
		defer r.lock.Done()
		r.stateLoop()
	}()
	r.INFO("raft.start...")
	return nil
}

func (r *Raft) running() bool {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	return r.state != STOPPED
}

// Stop used to stop the raft.
func (r *Raft) Stop() error {
	if r.getState() == STOPPED {
		return nil
	}

	close(r.fired)
	r.setState(STOPPED)

	// wait all goroutine stopped
	r.lock.Wait()
	r.freePeers()
	r.WARNING("raft.stopped...")
	return nil
}

// init all peers for raft.Peers(from RaftConfig.Peers) and raft.IdlePeers(from RaftConfig.IdlePeers)
func (r *Raft) initPeers() {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	metaPath := filepath.Join(r.conf.MetaDatadir, metaFile)
	if _, err := os.Stat(metaPath); os.IsNotExist(err) {
		r.WARNING("peers.json.file[%v].does.not.exist", metaPath)
	} else {
		peers, idlePeers, _ := readPeersJSON(filepath.Join(r.conf.MetaDatadir, metaFile))
		r.meta.Peers = append(r.meta.Peers, peers...)
		r.meta.IdlePeers = append(r.meta.IdlePeers, idlePeers...)
		r.WARNING("prepare.to.recovery.peers.from.[%v].peers[%v].idlePeers[%v]", r.conf.MetaDatadir, r.meta.Peers, r.meta.IdlePeers)
	}

	// create peers
	for _, connStr := range r.meta.Peers {
		if connStr != r.getID() {
			p := NewPeer(r, connStr, r.conf.RequestTimeout, r.conf.HeartbeatTimeout)
			r.peers[connStr] = p
		}
	}

	// if peers is empty, append this peer
	if len(r.meta.Peers) == 0 && !r.conf.SuperIDLE {
		r.meta.Peers = append(r.meta.Peers, r.getID())
	}

	// create idle peers
	for _, connStr := range r.meta.IdlePeers {
		if connStr != r.getID() {
			p := NewPeer(r, connStr, r.conf.RequestTimeout, r.conf.HeartbeatTimeout)
			r.idlePeers[connStr] = p
		}
	}

	// if peers is empty, append this peer
	if len(r.meta.IdlePeers) == 0 && r.conf.SuperIDLE {
		r.meta.IdlePeers = append(r.meta.IdlePeers, r.getID())
	}
}

// free all peers
func (r *Raft) freePeers() {
	for _, peer := range r.peers {
		peer.freePeer()
	}
}

// send command to state machine(F/C/L/I/S) loop with maxSendTime tiemout
// (F/C/L/I/S)-loop should handle it and return
func (r *Raft) send(t int, request interface{}) (interface{}, error) {
	if !r.running() {
		return nil, errStop
	}

	tm1 := common.NormalTimeout(maxSendTime)
	defer common.NormalTimerRelaese(tm1)

	event := &ev{Type: t, request: request, response: make(chan interface{}, 1)}
	select {
	case r.c <- event:
	case <-tm1.C:
	}

	tm2 := common.NormalTimeout(maxSendTime)
	defer common.NormalTimerRelaese(tm2)
	select {
	case <-tm2.C:
		return nil, errSend
	case rsp := <-event.response:
		return rsp, nil
	}
}

// loopFired is used to fire the state loop and do state transition
func (r *Raft) loopFired() {
	r.fired <- true
}

// wait for state machine changing
func (r *Raft) stateLoop() {
	state := r.getState()

	for state != STOPPED {
		switch state {
		case FOLLOWER:
			r.F.startCheckBrainSplit()
			r.F.Loop()
		case CANDIDATE:
			r.C.Loop()
		case LEADER:
			r.L.Loop()
		case IDLE:
			r.I.Loop()
		case INVALID:
			r.IV.Loop()
		case LEARNER:
			r.LN.Loop()
		}
		state = r.getState()
	}
	r.WARNING("raft.stateLoop.end")
}

// check the request comes from this cluster
func (r *Raft) checkRequest(req *model.RaftRPCRequest) bool {
	return r.peers[req.GetFrom()] != nil
}

// updateView
func (r *Raft) updateView(viewid uint64, leader string) {
	r.WARNING("do.updateViewID[FROM:%v TO:%v]", r.meta.ViewID, viewid)

	// update leader and viewid
	r.leader = leader
	r.votedFor = noVote
	r.meta.ViewID = viewid
}

func (r *Raft) updateEpoch(epochid uint64, peers []string, idlePeers []string) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	mark := make(map[string]bool)

	// update peers
	for _, name := range peers {
		if r.peers[name] == nil {
			if name != r.getID() {
				p := NewPeer(r, name, r.conf.RequestTimeout, r.conf.HeartbeatTimeout)
				r.peers[name] = p
			}
		}
		mark[name] = true
	}

	for name, peer := range r.peers {
		if _, ok := mark[name]; !ok {
			peer.freePeer()
			delete(r.peers, name)
		}
	}
	r.meta.Peers = peers

	// update idle peers
	for _, name := range idlePeers {
		if r.idlePeers[name] == nil {
			if name != r.getID() {
				p := NewPeer(r, name, r.conf.RequestTimeout, r.conf.HeartbeatTimeout)
				r.idlePeers[name] = p
			}
		}
		mark[name] = true
	}

	for name, peer := range r.idlePeers {
		if _, ok := mark[name]; !ok {
			peer.freePeer()
			delete(r.idlePeers, name)
		}
	}
	r.meta.IdlePeers = idlePeers

	r.meta.EpochID = epochid
	r.writePeersJSON()
}

func (r *Raft) writePeersJSON() {
	metaPath := filepath.Join(r.conf.MetaDatadir, metaFile)
	if err := writePeersJSON(metaPath, r.meta.Peers, r.meta.IdlePeers); err != nil {
		r.PANIC("writePeers[%v].to[%v].error[%+v]", metaPath, r.meta.Peers, err)
	}

	// Check the meta path.
	fileInfo, err := os.Lstat(metaPath)
	if err != nil {
		r.ERROR("check.peers.json[%s].error[%+v]", metaPath, err)
	}
	r.INFO("check.peers.json.file[%s].stat[name:%v, mode:%v, size:%v, lastmodified:%v]", metaPath, fileInfo.Name(), fileInfo.Mode(), fileInfo.Size(), fileInfo.ModTime())
}

func (r *Raft) updateStateBegin() {
	r.stateBegin = time.Now()
}

func (r *Raft) resetHeartbeatTimeout() {
	common.NormalTimerRelaese(r.heartbeatTick)
	r.heartbeatTick = common.NormalTimeout(r.getHeartbeatTimeout())
}

func (r *Raft) resetElectionTimeout() {
	common.NormalTimerRelaese(r.electionTick)
	r.electionTick = common.RandomTimeout(r.getElectionTimeout())
}

func (r *Raft) resetCheckBrainSplitTimeout() {
	common.NormalTimerRelaese(r.checkBrainSplitTick)
	r.checkBrainSplitTick = common.NormalTimeout(r.getElectionTimeout() / 2)
}

func (r *Raft) resetCheckVotesTimeout() {
	// timeout is 1/2 of electiontimout
	common.NormalTimerRelaese(r.checkVotesTick)
	r.checkVotesTick = common.NormalTimeout(r.getElectionTimeout() / 2)
}

// SetSkipPurgeBinlog used to set purge binlog or not.
func (r *Raft) SetSkipPurgeBinlog(v bool) {
	r.skipPurgeBinlog = v
}

// SetSkipCheckSemiSync used to set check semi-sync or not.
func (r *Raft) SetSkipCheckSemiSync(v bool) {
	r.skipCheckSemiSync = v
}
