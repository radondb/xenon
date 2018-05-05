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
	"time"
)

// IncLeaderPromotes counter.
func (s *Raft) IncLeaderPromotes() {
	atomic.AddUint64(&s.stats.LeaderPromotes, 1)
}

// IncLeaderDegrades counter.
func (s *Raft) IncLeaderDegrades() {
	atomic.AddUint64(&s.stats.LeaderDegrades, 1)
}

// IncLeaderGetHeartbeatRequests counter.
func (s *Raft) IncLeaderGetHeartbeatRequests() {
	atomic.AddUint64(&s.stats.LeaderGetHeartbeatRequests, 1)
}

// IncLeaderPurgeBinlogs counter.
func (s *Raft) IncLeaderPurgeBinlogs() {
	atomic.AddUint64(&s.stats.LeaderPurgeBinlogs, 1)
}

// IncLeaderPurgeBinlogFails counter.
func (s *Raft) IncLeaderPurgeBinlogFails() {
	atomic.AddUint64(&s.stats.LeaderPurgeBinlogFails, 1)
}

// IncLeaderGetVoteRequests counter.
func (s *Raft) IncLeaderGetVoteRequests() {
	atomic.AddUint64(&s.stats.LeaderGetVoteRequests, 1)
}

// IncLessHeartbeatAcks counter.
func (s *Raft) IncLessHeartbeatAcks() {
	atomic.AddUint64(&s.stats.LessHearbeatAcks, 1)
}

// IncCandidatePromotes counter.
func (s *Raft) IncCandidatePromotes() {
	atomic.AddUint64(&s.stats.CandidatePromotes, 1)
}

// IncCandidateDegrades counter.
func (s *Raft) IncCandidateDegrades() {
	atomic.AddUint64(&s.stats.CandidateDegrades, 1)
}

// SetRaftMysqlStatus used to set mysql status.
func (s *Raft) SetRaftMysqlStatus(rms model.RAFTMYSQL_STATUS) {
	s.stats.RaftMysqlStatus = rms
}

// ResetRaftMysqlStatus used to reset mysql status.
func (s *Raft) ResetRaftMysqlStatus() {
	s.stats.RaftMysqlStatus = model.RAFTMYSQL_NONE
}

func (s *Raft) getStats() *model.RaftStats {
	return &model.RaftStats{
		HaEnables:                  atomic.LoadUint64(&s.stats.HaEnables),
		LeaderPromotes:             atomic.LoadUint64(&s.stats.LeaderPromotes),
		LeaderDegrades:             atomic.LoadUint64(&s.stats.LeaderDegrades),
		LeaderGetHeartbeatRequests: atomic.LoadUint64(&s.stats.LeaderGetHeartbeatRequests),
		LeaderGetVoteRequests:      atomic.LoadUint64(&s.stats.LeaderGetVoteRequests),
		LeaderPurgeBinlogs:         atomic.LoadUint64(&s.stats.LeaderPurgeBinlogs),
		LeaderPurgeBinlogFails:     atomic.LoadUint64(&s.stats.LeaderPurgeBinlogFails),
		LessHearbeatAcks:           atomic.LoadUint64(&s.stats.LessHearbeatAcks),
		CandidatePromotes:          atomic.LoadUint64(&s.stats.CandidatePromotes),
		CandidateDegrades:          atomic.LoadUint64(&s.stats.CandidateDegrades),
		StateUptimes:               uint64(time.Since(s.stateBegin).Seconds()),
		RaftMysqlStatus:            s.stats.RaftMysqlStatus,
	}
}
