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
	"mysql"
	"testing"
	"time"
	"xbase/common"
	"xbase/xlog"

	"github.com/stretchr/testify/assert"
)

func TestRaftStateString(t *testing.T) {
	// FOLLOWER
	{
		state := FOLLOWER
		want := "FOLLOWER"
		got := state.String()
		assert.Equal(t, want, got)
	}

	// LEADER
	{
		state := LEADER
		want := "LEADER"
		got := state.String()
		assert.Equal(t, want, got)
	}

	// CANDIDATE
	{
		state := CANDIDATE
		want := "CANDIDATE"
		got := state.String()
		assert.Equal(t, want, got)
	}

	// IDLE
	{
		state := IDLE
		want := "IDLE"
		got := state.String()
		assert.Equal(t, want, got)
	}

	// STOPPED
	{
		state := STOPPED
		want := "STOPPED"
		got := state.String()
		assert.Equal(t, want, got)
	}
}

// TEST EFFECTS:
// test the quorums of cluster
//
// TEST PROCESSES:
// 1. Start 3 rafts
// 2. check quorums
func TestRaftQuorums(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	port := common.RandomPort(8000, 9000)
	_, rafts, cleanup := MockRafts(log, port, 3, -1)
	defer cleanup()
	raft := rafts[0]

	// Start rafts
	for _, raft := range rafts {
		raft.Start()
	}

	want := 3
	got := raft.GetMembers()
	assert.Equal(t, want, got)

	want = 2
	got = raft.GetQuorums()
	assert.Equal(t, want, got)
}

// TEST EFFECTS:
// test the election of a cluster have only one node
//
// TEST PROCESSES:
// 1. Start 1 raft state as FOLLOWER
// 2. wait new leader election
// 3. the FOLLOWER can't upgrade to CANDIDATE state
// 4. keep the FOLLOWER state machine
func TestRaftClusterOnly1Node(t *testing.T) {
	var want, got State

	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	port := common.RandomPort(8000, 9000)
	_, rafts, cleanup := MockRafts(log, port, 1, -1)
	defer cleanup()

	// 1. Start 1 rafts state as FOLLOWER
	{
		for _, raft := range rafts {
			raft.Start()
		}

		got = 0
		want = (FOLLOWER)
		for _, raft := range rafts {
			got += raft.getState()
		}

		// [FOLLOWER, FOLLOWER, FOLLOWER]
		assert.Equal(t, want, got)
	}

	// 2. wait leader election, but the cluster has one node(only me)
	// we can't win the election
	{
		MockWaitLeaderEggs(rafts, 0)
		got = 0
		want = (FOLLOWER)
		for _, raft := range rafts {
			got += raft.getState()
		}
		// [FOLLOWER]
		assert.Equal(t, want, got)
	}
}

// TEST EFFECTS:
// test the leader down case
//
// TEST PROCESSES:
// 1. Start 3 rafts state as FOLLOWER
// 2. wait leader election from 3 FOLLOWER
// 3. Stop leader
// 4. wait new leader election from 2 FOLLOWER
// 5. Stop the new leader
// 6. wait new leader election, but never get out
func TestRaftLeaderDown(t *testing.T) {
	var testName = "TestRaftLeaderDown"
	var want, got State
	var whoisleader int
	var leader *Raft

	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	port := common.RandomPort(8000, 9000)
	_, rafts, cleanup := MockRafts(log, port, 3, -1)
	defer cleanup()

	// 1. Start 3 rafts state as FOLLOWER
	{
		for _, raft := range rafts {
			raft.Start()
		}

		got = 0
		want = (FOLLOWER + FOLLOWER + FOLLOWER)
		for _, raft := range rafts {
			got += raft.getState()
		}

		// [FOLLOWER, FOLLOWER, FOLLOWER]
		assert.Equal(t, want, got)
	}

	// 2. wait leader election from 3 FOLLOWER
	{
		MockWaitLeaderEggs(rafts, 1)

		whoisleader = 0
		got = 0
		want = (LEADER + FOLLOWER + FOLLOWER)
		for i, raft := range rafts {
			got += raft.getState()
			if raft.getState() == LEADER {
				whoisleader = i
			}
		}
		// [LEADER, FOLLOWER, FOLLOWER]
		assert.Equal(t, want, got)
	}

	// 3. Stop leader
	{
		leader = rafts[whoisleader]
		log.Warning("%v.leader[%v].prepare.down", testName, leader.getID())
		leader.Stop()
	}

	// 4. wait new leader election from 2 FOLLOWER
	{
		MockWaitLeaderEggs(rafts, 1)
		whoisleader = 0
		got = 0
		want = (LEADER + FOLLOWER + STOPPED)
		for i, raft := range rafts {
			got += raft.getState()
			if raft.getState() == LEADER {
				whoisleader = i
			}
		}

		// [LEADER, FOLLOWER, STOPPED]
		assert.Equal(t, want, got)
	}

	// 5. Stop new leader
	{
		leader = rafts[whoisleader]
		log.Warning("%v.newleader[%v].prepare.down", testName, leader.getID())
		rafts[whoisleader].Stop()
	}

	// 6. wait new leader election from 1 FOLLOWER, but never eggs
	{
		MockWaitLeaderEggs(rafts, 0)

		got = 0
		for _, raft := range rafts {
			got += raft.getState()
		}
		// [FOLLOWER, STOPPED, STOPPED]
		want = (FOLLOWER + STOPPED + STOPPED)
		assert.Equal(t, want, got)
	}
}

// TEST EFFECTS:
// test the leader localcommit case
//
// TEST PROCESSES:
// 1. Start 3 rafts state as FOLLOWER
// 2. wait leader election from 3 FOLLOWER
// 3. Stop leader wait new leader election from 2 FOLLOWER
//    remock old leader to localcommit then wait a heartbeat timeout
func TestRaftLeaderLocalCommit(t *testing.T) {
	var want, got State
	var whoisleader int
	var leader *Raft

	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	port := common.RandomPort(8000, 9000)
	_, rafts, cleanup := MockRafts(log, port, 3, -1)
	defer cleanup()

	// 1. Start 3 rafts state as FOLLOWER
	{
		for _, raft := range rafts {
			raft.Start()
		}

		got = 0
		want = (FOLLOWER + FOLLOWER + FOLLOWER)
		for _, raft := range rafts {
			got += raft.getState()
		}

		// [FOLLOWER, FOLLOWER, FOLLOWER]
		assert.Equal(t, want, got)
	}

	// 2. wait leader election from 3 FOLLOWER
	{
		MockWaitLeaderEggs(rafts, 1)

		whoisleader = 0
		got = 0
		want = (LEADER + FOLLOWER + FOLLOWER)
		for i, raft := range rafts {
			got += raft.getState()
			if raft.getState() == LEADER {
				whoisleader = i
			}
		}
		// [LEADER, FOLLOWER, FOLLOWER]
		assert.Equal(t, want, got)
	}

	leader = rafts[whoisleader]
	leader.mysql.SetMysqlHandler(mysql.NewMockGTIDLC())

	// 3. stop leader wait new leader election from 2 FOLLOWER
	//    remock leader to localcommit

	{
		leader.Stop()

		MockWaitLeaderEggs(rafts, 1)
		got = 0
		want = (LEADER + FOLLOWER + STOPPED)
		for _, raft := range rafts {
			got += raft.getState()
		}

		// [LEADER, FOLLOWER, STOPPED]
		assert.Equal(t, want, got)

		got = 0
		leader.mysql.SetMysqlHandler(mysql.NewMockGTIDLC())
		leader.Start()
		MockWaitHeartBeatTimeout()
		want = (LEADER + FOLLOWER + INVALID)
		for _, raft := range rafts {
			got += raft.getState()
		}

		// [LEADER, FOLLOWER, INVALID]
		assert.Equal(t, want, got)
	}
}

// TEST EFFECTS:
// Double-Raft-Diffraction experiment
// how evil this testcase is!!!
// we make these two raft clusters under diffraction patterns
// to prevent this happenning again, we sit
//   if !r.checkRequest(req) {
//        return ErrorInvalidRequest
//    }
// at the door
//
// TEST PROCESSES:
// 1.  Start cluster1 which has 2 rafts state as FOLLOWER
// 2.  Start cluster2 which has 2 rafts state as FOLLOWER
// 3.  wait cluster1.leader election from 2 FOLLOWERs
// 4.  wait cluster2.leader election from 2 FOLLOWERs
// 5.  Stop cluster1.leader
// 6.  Stop cluster2.follower
// 6.  wait cluster1.candidate eggs
// 8.  cluster2.leader add cluster1.candidate as his peers and broadcast heartbeat to him
// 9.  cluster1.candidate give an ErrorInvalidRequest to cluster2.leader, since you are not a member of cluster1
// 10. cluster1.candidate add cluster2.leader as his peers
// 11. cluster2.leader get a requestvote from cluster1.candidate
// 12. cluster1.candidate granted enough votes and updgrde to leader
// 13. cluster2.leader degrade to FOLLOWER
func TestRaftDoubleClusterDiffraction(t *testing.T) {
	var testName = "TestRaftDoubleClusterDiffraction"
	var want, got State
	var cluster1WhoIsLeader int
	var cluster1Leader *Raft
	var cluster1WhoIsFollower int
	var cluster1Follower *Raft

	var cluster2WhoIsLeader int
	var cluster2Leader *Raft
	var cluster2WhoIsFollower int
	var cluster2Follower *Raft

	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))

	// cluster1
	port := common.RandomPort(8000, 9000)
	_, rafts1, cleanup1 := MockRafts(log, port, 2, -1)
	defer cleanup1()

	// cluster2
	port = common.RandomPort(8000, 9000)
	_, rafts2, cleanup2 := MockRafts(log, port, 2, -1)
	defer cleanup2()

	// Start cluster1
	{
		for _, raft := range rafts1 {
			raft.Start()
		}
	}

	// Start cluster2
	{
		for _, raft := range rafts2 {
			raft.Start()
		}
	}

	// 2. check cluster1
	{
		MockWaitLeaderEggs(rafts1, 1)
		got = 0
		want = (LEADER + FOLLOWER)
		for i, raft := range rafts1 {
			got += raft.getState()
			if raft.getState() == LEADER {
				cluster1WhoIsLeader = i
			}
		}
		// [LEADER, FOLLOWER]
		assert.Equal(t, want, got)
	}

	// 2. check cluster2
	{
		MockWaitLeaderEggs(rafts2, 1)
		got = 0
		want = (LEADER + FOLLOWER)
		for i, raft := range rafts2 {
			got += raft.getState()
			if raft.getState() == LEADER {
				cluster2WhoIsLeader = i
			}

			if raft.getState() == FOLLOWER {
				cluster2WhoIsFollower = i
			}
		}
		// [LEADER, FOLLOWER]
		assert.Equal(t, want, got)
	}

	// 3. Stop cluster1.leader and cluster2WhoIsFollower
	{
		cluster1Leader = rafts1[cluster1WhoIsLeader]
		log.Warning("%v.cluster1.leader[%v].prepare.down", testName, cluster1Leader.getID())
		cluster1Leader.Stop()

		cluster2Follower = rafts2[cluster2WhoIsFollower]
		log.Warning("%v.cluster2WhoIsFollower[%v].prepare.down", testName, cluster2Follower.getID())
		cluster2Follower.Stop()
	}

	// 4. check cluster1 state
	{
		MockWaitLeaderEggs(rafts1, 0)

		got = 0
		want = (FOLLOWER + STOPPED)
		for i, raft := range rafts1 {
			got += raft.getState()
			if raft.getState() == FOLLOWER {
				cluster1WhoIsFollower = i
			}
		}

		// [FOLLOWER, STOPPED]
		assert.Equal(t, want, got)
	}

	// 5. add candidate to cluster2
	{
		cluster2Leader = rafts2[cluster2WhoIsLeader]

		// cluster2Leader never degrade with this hook
		cluster2Leader.L.setProcessHeartbeatResponseHandler(cluster2Leader.mockLeaderProcessSendHeartbeatResponse)

		cluster1Follower = rafts1[cluster1WhoIsFollower]
		cluster2Leader.AddPeer(cluster1Follower.getID())

		// wait a hearbeat broadcast
		// cluster1Follower will give ErrorInvalidRequest because cluster2Leader not one of our cluster member
		MockWaitLeaderEggs(rafts1, 0)
	}

	// 6. add cluster2Leader to cluster1Follower
	{
		cluster2Leader = rafts2[cluster2WhoIsLeader]
		cluster1Follower = rafts1[cluster1WhoIsFollower]
		cluster1Follower.AddPeer(cluster2Leader.getID())

		// wait a hearbeat broadcast
		MockWaitLeaderEggs(rafts1, 0)
	}

	// 6. check cluster1
	{
		MockWaitLeaderEggs(rafts1, 1)
		got = 0
		want = (LEADER + STOPPED)
		for _, raft := range rafts1 {
			got += raft.getState()
		}
		// [LEADER, STOPPED]
		assert.Equal(t, want, got)
	}

	// 7. check cluster2
	{
		MockWaitLeaderEggs(rafts2, 0)
		got = 0
		want = (FOLLOWER + STOPPED)
		for _, raft := range rafts2 {
			got += raft.getState()
		}
		// want = (FOLLOWER + STOPPED)
		assert.Equal(t, want, got)
	}
}

// TEST EFFECTS:
// test the leader down and up case, it mocks the leader rejoin after a network partition
//
// TEST PROCESSES:
// 1.  Start 3 rafts and all state as FOLLOWER
// 2.  wait leader election from 3 FOLLOWER
// 3.  set leader handlers to mock
// 4.  Stop leader hearbeat
// 5.  wait a pingtimeout, wait new-leader election from 2 FOLLOWER
// 6.  old-leader get requestvote and return ErrorInvalidRequest
// 7.  new-leader eggs
// 8.  old-leader reset handlers to work
// 9.  old-leader Start heartbeat
// 10. new-leader get old-leader's hearbeat and reject
// 11. old-leader get new-leader's hearbeat(with larger viewid)
// 12. old-leader down to FOLLOWER
func TestRaftLeaderDownAndUp(t *testing.T) {
	var testName = "TestRaftLeaderDownAndUp"
	var want, got State
	var whoisleader int
	var leader *Raft
	var newleader *Raft

	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	port := common.RandomPort(8000, 9000)
	_, rafts, cleanup := MockRafts(log, port, 3, -1)
	defer cleanup()

	// 1.  Start 3 rafts and all state as FOLLOWER
	{
		for _, raft := range rafts {
			raft.Start()
		}

		got = 0
		want = (FOLLOWER + FOLLOWER + FOLLOWER)
		for _, raft := range rafts {
			got += raft.getState()
		}

		// [FOLLOWER, FOLLOWER, FOLLOWER]
		assert.Equal(t, want, got)
	}

	// 2.  wait leader election from 3 FOLLOWER
	{
		MockWaitLeaderEggs(rafts, 1)
		whoisleader = 0
		got = 0
		want = (LEADER + FOLLOWER + FOLLOWER)
		for i, raft := range rafts {
			got += raft.getState()
			if raft.getState() == LEADER {
				whoisleader = i
			}
		}
		// [LEADER, FOLLOWER, FOLLOWER]
		assert.Equal(t, want, got)
	}

	// 3.  set leader handlers to mock
	{
		leader = rafts[whoisleader]
		log.Warning("%v.leader[%v].set.mock.functions", testName, rafts[whoisleader].getID())
		leader.L.setProcessHeartbeatRequestHandler(leader.mockLeaderProcessHeartbeatRequest)
		leader.L.setProcessRequestVoteRequestHandler(leader.mockLeaderProcessRequestVoteRequest)
	}

	// 4.  Stop leader hearbeat
	{
		log.Warning("%v.leader[%v].Stop.heartbeat", testName, leader.getID())
		leader.L.setSendHeartbeatHandler(leader.mockLeaderSendHeartbeat)
	}

	// 5. wait a pingtimeout, wait new-leader election from 2 FOLLOWER
	{
		MockWaitMySQLPingTimeout()
		MockWaitLeaderEggs(rafts, 1)

		got = 0
		imoldleader := whoisleader
		want = (LEADER + FOLLOWER + FOLLOWER)
		for i, raft := range rafts {
			got += raft.getState()
			if raft.getState() == LEADER {
				// skip the old-leader
				if imoldleader != i {
					whoisleader = i
				}
			}
		}

		newleader = rafts[whoisleader]

		// [LEADER, LEADER, FOLLOWER]
		assert.Equal(t, want, got)
	}

	// 6.  old-leader reset handlers to work
	{
		log.Warning("%v.old.leader[%v].Start.heartbeat", testName, leader.getID())
		leader.L.setSendHeartbeatHandler(leader.L.sendHeartbeat)
		leader.L.setProcessRequestVoteRequestHandler(leader.L.processRequestVoteRequest)
		MockWaitLeaderEggs(rafts, 0)
	}

	// 7. old-leader down to FOLLOWER
	{
		got = 0
		want = (LEADER + FOLLOWER + FOLLOWER)
		for _, raft := range rafts {
			got += raft.getState()
		}

		// [LEADER, FOLLOWER, FOLLOWER]
		assert.Equal(t, want, got)

		// check the new-leader != old-leader
		assert.NotEqual(t, leader.getID(), newleader.getID())
	}
}

// TEST EFFECTS:
// test epoch change
//
// TEST PROCESSES:
// 1. Start 5 rafts state as FOLLOWER
// 2. wait leader election from 5 FOLLOWERs
// 3. change leader epoch
// 4. check leader and follower epoch
func TestRaftEpochChange(t *testing.T) {
	var testName = "TestRaftEpochChange"
	var want, got, raftnums State
	var newpeer = "127.0.0.1:18888"
	var exists int
	var whoisleader int
	var leader *Raft

	raftnums = 5
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	port := common.RandomPort(8000, 9000)
	cluster, rafts, cleanup := MockRafts(log, port, int(raftnums), -1)
	defer cleanup()

	// 1. Start rafts state as FOLLOWER
	{
		for _, raft := range rafts {
			raft.Start()
		}

		got = 0
		want = (FOLLOWER * raftnums)
		for _, raft := range rafts {
			got += raft.getState()
		}

		// (FOLLOWER * raftnums)
		assert.Equal(t, want, got)
	}

	// 2. wait leader election from FOLLOWERs
	{
		MockWaitLeaderEggs(rafts, 1)

		whoisleader = 0
		got = 0
		exists = 0
		want = (LEADER + FOLLOWER*(raftnums-1))
		for i, raft := range rafts {
			got += raft.getState()
			if raft.getState() == LEADER {
				whoisleader = i
			}

			if raft.peers[newpeer] != nil {
				exists++
			}
		}

		//(LEADER + FOLLOWER * int(raftnums - 1))
		assert.Equal(t, want, got)

		// check newpeer non exists
		assert.Equal(t, exists, 0)

		leader = rafts[whoisleader]
		log.Warning("%v.leader[%v].prepare.change.epoch.by.add.peer[%v]", testName, leader.getID(), newpeer)
		leader.AddPeer(newpeer)
	}

	// 3. wait epoch change broadcast
	{
		MockWaitLeaderEggs(rafts, 0)
	}

	// 4. check newpeer exists
	{
		for _, raft := range rafts {
			peers := raft.GetPeers()
			for _, peer := range peers {
				if peer == newpeer {
					exists++
				}
			}
		}

		assert.Equal(t, exists, len(cluster))

		// 5. remove a alive peer
		if whoisleader+1 == len(cluster) {
			newpeer = cluster[whoisleader-1]
		} else {
			newpeer = cluster[whoisleader+1]
		}
		log.Warning("%v.leader[%v].prepare.change.epoch.by.remove.peer[%v]", testName, leader.getID(), newpeer)
		leader.RemovePeer(newpeer)
	}

	// 5. wait epoch change broadcast
	MockWaitLeaderEggs(rafts, 0)

	// 6. check newpeer exists
	{
		exists = 0
		for _, raft := range rafts {

			peers := raft.GetPeers()
			for _, peer := range peers {
				if peer == newpeer {
					exists++
				}
			}
		}
		assert.Equal(t, exists, 1)
	}
}

// TEST EFFECTS:
// test epoch change under IDLE state
//
// TEST PROCESSES:
// 1. Start 5 rafts state as FOLLOWER
// 2. wait leader election from 5 FOLLOWERs
// 3. set 2 FOLLOWERs to IDLE state
// 4. change leader epoch
// 5. check leader and follower epoch
func TestRaftEpochChangeUnderIDLE(t *testing.T) {
	var testName = "TestRaftEpochChange"
	var want, got, raftnums State
	var newpeer = "127.0.0.1:18888"
	var exists int
	var whoisleader int
	var leader *Raft

	raftnums = 5
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	port := common.RandomPort(8100, 8200)
	cluster, rafts, cleanup := MockRafts(log, port, int(raftnums), -1)
	defer cleanup()

	// 1. Start rafts state as FOLLOWER
	{
		for _, raft := range rafts {
			raft.Start()
		}

		got = 0
		want = (FOLLOWER * raftnums)
		for _, raft := range rafts {
			got += raft.getState()
		}

		// (FOLLOWER * raftnums)
		assert.Equal(t, want, got)
	}

	// 2. wait leader election from FOLLOWERs
	{
		MockWaitLeaderEggs(rafts, 1)

		whoisleader = 0
		got = 0
		exists = 0
		want = (LEADER + FOLLOWER*(raftnums-1))
		for i, raft := range rafts {
			got += raft.getState()
			if raft.getState() == LEADER {
				whoisleader = i
			}

			if raft.peers[newpeer] != nil {
				exists++
			}
		}

		//(LEADER + FOLLOWER * int(raftnums - 1))
		assert.Equal(t, want, got)

		// check newpeer non exists
		assert.Equal(t, exists, 0)

		leader = rafts[whoisleader]
		log.Warning("%v.leader[%v].set.to.IDLE.state", testName, leader.getID())
		MockStateTransition(leader, IDLE)
	}

	// 3. wait new leader eggs
	{
		MockWaitLeaderEggs(rafts, 0)
		MockWaitLeaderEggs(rafts, 1)
		whoisleader = 0
		got = 0
		exists = 0
		want = (LEADER + IDLE + FOLLOWER*(raftnums-2))
		for i, raft := range rafts {
			got += raft.getState()
			if raft.getState() == LEADER {
				whoisleader = i
			}
		}

		// want = (LEADER + IDLE + FOLLOWER*(raftnums-2))
		assert.Equal(t, want, got)
		leader = rafts[whoisleader]
		log.Warning("%v.leader[%v].prepare.change.epoch.by.add.peer[%v]", testName, leader.getID(), newpeer)
		leader.AddPeer(newpeer)
	}

	// 4. check newpeer exists
	{
		// wait epoch change broadcast
		MockWaitLeaderEggs(rafts, 0)

		for _, raft := range rafts {
			peers := raft.GetPeers()
			for _, peer := range peers {
				if peer == newpeer {
					exists++
				}
			}
		}

		// except the IDLE
		want := len(cluster)
		got := exists
		assert.Equal(t, want, got)
	}
}

// TEST EFFECTS:
// test election under IDLE in the majority
//
// TEST PROCESSES:
// 1. Start 7 rafts state as FOLLOWER
// 2. set 4 FOLLOWERs to IDLE state
// 3. wait leader election from 3 FOLLOWERs
// 4. make leader to IDLE and wait the new leader eggs
// 5. check the IDLE get the new leader
func TestRaftElectionUnderIDLEInMajority(t *testing.T) {
	var want, got, raftnums State
	var whoisleader int
	var leader string

	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))

	raftnums = 7
	idles := int(raftnums - (raftnums / 2))
	actives := int(raftnums) - idles
	port := common.RandomPort(8100, 8200)
	_, rafts, cleanup := MockRafts(log, port, int(raftnums), int(raftnums/2))
	defer cleanup()

	// 1. Start rafts state as FOLLOWER
	{
		for _, raft := range rafts {
			raft.Start()
		}

		// set more than half of the nodes to IDLE
		for i := actives; i < int(raftnums); i++ {
			MockStateTransition(rafts[i], IDLE)
		}

		// set rafts[0] mysql as GTID_E(mock mysql error)
		MockSetMysqlHandler(rafts[0], mysql.NewMockGTIDError())
	}

	// 2. wait leader election from actives FOLLOWERs
	{
		MockWaitLeaderEggs(rafts, 1)

		got = 0
		want = (LEADER + FOLLOWER*State(actives-1) + IDLE*State(idles))
		for i, raft := range rafts {
			got += raft.getState()
			if raft.getState() == LEADER {
				whoisleader = i
			}
		}

		//(LEADER + FOLLOWER*(actives - 1) + IDLE*idles))
		assert.Equal(t, want, got)

		// check the leader idx < idles - 1
		assert.True(t, whoisleader < (idles-1), "")

		// get [idles-1]'s leader
		leader = rafts[idles-1].getLeader()
	}

	// 3. make leader to IDLE and check the IDLEs got the newone
	{
		MockSetMysqlHandler(rafts[0], mysql.NewMockGTIDA())

		// make leader to IDLE
		MockStateTransition(rafts[whoisleader], IDLE)
		MockWaitLeaderEggs(rafts, 1)

		// get [idles-1]'s leader again
		for i := 0; i < idles; i++ {
			leader1 := rafts[i].getLeader()
			assert.NotEqual(t, leader, leader1)
		}
	}
}

// TEST EFFECTS:
// test run as IDLE with config.StartAsIDLE=true configuration
//
// TEST PROCESSES:
// 1. Start 1 raft with SuperIDLE=true
// 2. check the IDLE
func TestRaftStartAsIDLE(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	conf := config.DefaultRaftConfig()
	conf.SuperIDLE = true
	port := common.RandomPort(8100, 8200)
	_, rafts, cleanup := MockRaftsWithConfig(log, conf, port, 1, 0)
	defer cleanup()

	// 1. Start rafts
	{
		for _, raft := range rafts {
			raft.Start()
		}
	}

	// 2. check state
	{
		want := IDLE
		got := rafts[0].getState()
		assert.Equal(t, want, got)
	}
}

// TEST EFFECTS:
// test run as FOLLOWER
//
// TEST PROCESSES:
// 1. Start 1 raft as FOLLOWER
// 2. check the STATE still FOLLOWER
func TestRaftStartAsFOLLOWER(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	conf := config.DefaultRaftConfig()
	port := common.RandomPort(8100, 8200)
	_, rafts, cleanup := MockRaftsWithConfig(log, conf, port, 1, -1)
	defer cleanup()

	// 1. Start rafts
	{
		for _, raft := range rafts {
			raft.Start()
		}
	}

	// 2. check state
	{
		want := FOLLOWER
		got := rafts[0].getState()
		assert.Equal(t, want, got)
	}

	// 3. wait a election timeout check state
	{
		MockWaitLeaderEggs(rafts, 0)
		want := FOLLOWER
		got := rafts[0].getState()
		assert.Equal(t, want, got)
	}
}

// TEST EFFECTS:
// test a cluster with 11 rafts
//
// TEST PROCESSES:
// 1. Start 11 rafts state as FOLLOWER
// 2. wait leader election from 11 FOLLOWERs
// 3. Stop the leader
// 4. wait the new leader eggs
// 5. Byzantine Failures Attack
func TestRaft11Rafts1Cluster(t *testing.T) {
	var raftnums State
	var whoisleader int
	var leader *Raft

	raftnums = 11
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	port := common.RandomPort(8600, 9000)
	_, rafts, cleanup := MockRafts(log, port, int(raftnums), -1)
	defer cleanup()

	// 1. Start rafts state as FOLLOWER
	{
		for _, raft := range rafts {
			raft.Start()
		}

		var got State
		want := (FOLLOWER * raftnums)
		for _, raft := range rafts {
			got += raft.getState()
		}

		// (FOLLOWER * raftnums)
		assert.Equal(t, want, got)
	}

	// 2. wait the leader eggs
	{
		// wait leader eggs
		MockWaitLeaderEggs(rafts, 1)

		whoisleader = 0
		for i, raft := range rafts {
			if raft.getState() == LEADER {
				whoisleader = i
				break
			}
		}

		var got State
		leader = rafts[whoisleader]
		want := (LEADER + FOLLOWER*(raftnums-1))
		for _, raft := range rafts {
			got += raft.getState()
		}

		// (LEADER + FOLLOWER*(raftnums-1))
		assert.Equal(t, want, got)
	}

	// 3. Stop the leader(mock to INVALID)
	MockStateTransition(leader, INVALID)

	// 4. wait the new leader eggs
	{
		// wait leader eggs
		MockWaitLeaderEggs(rafts, 1)

		for _, raft := range rafts {
			if raft.getState() == LEADER {
				break
			}
		}

		var got State
		want := (LEADER + INVALID + FOLLOWER*(raftnums-2))
		for _, raft := range rafts {
			got += raft.getState()
		}

		//want = (LEADER + INVALID + FOLLOWER*(raftnums-2))
		assert.Equal(t, want, got)
	}

	// 5. Byzantine Failures Attack
	{
		for _, raft := range rafts {
			MockStateTransition(raft, CANDIDATE)
		}
		MockWaitLeaderEggs(rafts, 1)

		for _, raft := range rafts {
			MockStateTransition(raft, LEADER)
		}
		MockWaitLeaderEggs(rafts, 1)
	}
}

// TEST EFFECTS:
// test the leader election with gtid
//
// TEST PROCESSES:
// 1. set rafts GTID
//    1.0 rafts[0]  with MockGTIDB{Master_Log_File = "", Read_Master_Log_Pos = 0}
//    1.1 rafts[1]  with MockGTIDB{Master_Log_File = "mysql-bin.000001", Read_Master_Log_Pos = 123
//            gtid.Executed_GTID_Set = "c78e798a-cccc-cccc-cccc-525433e8e796:1"}
//    1.2 rafts[2]  with MockGTIDC{Master_Log_File = "mysql-bin.000001", Read_Master_Log_Pos = 124
//            gtid.Executed_GTID_Set = "c78e798a-cccc-cccc-cccc-525433e8e796:2"}
// 2. Start 3 rafts state as FOLLOWER
// 3. wait rafts[2] elected as leader
// 4. Stop rafts[2]
// 5. wait rafts[1] elected as new leader
func TestRaftLeaderWithGTID(t *testing.T) {
	var whoisleader int

	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	port := common.RandomPort(8000, 9000)
	_, rafts, cleanup := MockRafts(log, port, 3, -1)
	defer cleanup()

	GTIDAIDX := 0
	GTIDBIDX := 1
	GTIDCIDX := 2

	// 1. set rafts GTID
	//    1.0 rafts[0]  with MockGTIDAA{Master_Log_File = "mysql-bin.000001", Read_Master_Log_Pos = 122,
	//    							   gtid.Retrieved_GTID_Set = "c78e798a-cccc-cccc-cccc-525433e8e796:1"}
	//    1.1 rafts[1]  with MockGTIDB{Master_Log_File = "mysql-bin.000001", Read_Master_Log_Pos = 123,
	//                                 gtid.Retrieved_GTID_Set = "c78e798a-cccc-cccc-cccc-525433e8e796:1-2"}
	//    1.2 rafts[2]  with MockGTIDC{Master_Log_File = "mysql-bin.000001", Read_Master_Log_Pos = 124,
	//								   gtid.Retrieved_GTID_Set = "c78e798a-cccc-cccc-cccc-525433e8e796:1-3"}
	{
		rafts[GTIDAIDX].mysql.SetMysqlHandler(mysql.NewMockGTIDAA())
		rafts[GTIDBIDX].mysql.SetMysqlHandler(mysql.NewMockGTIDB())
		rafts[GTIDCIDX].mysql.SetMysqlHandler(mysql.NewMockGTIDC())
	}

	// 2. Start 3 rafts state as FOLLOWER
	{
		for _, raft := range rafts {
			raft.Start()
		}

		var got State
		want := (FOLLOWER + FOLLOWER + FOLLOWER)
		for _, raft := range rafts {
			got += raft.getState()
		}

		// [FOLLOWER, FOLLOWER, FOLLOWER]
		assert.Equal(t, want, got)
	}

	// 3. wait rafts[2] elected as leader
	{
		MockWaitLeaderEggs(rafts, 1)

		var got State
		whoisleader = 0
		want := (LEADER + FOLLOWER + FOLLOWER)
		for i, raft := range rafts {
			got += raft.getState()
			if raft.getState() == LEADER {
				whoisleader = i
			}
		}
		MockWaitLeaderEggs(rafts, 1)
		// [LEADER, FOLLOWER, FOLLOWER]
		assert.Equal(t, want, got)

		// leader should be rafts[GTIDCIDX]
		assert.Equal(t, whoisleader, GTIDCIDX)
	}

	// 4. Stop rafts[2]
	{
		leader := rafts[whoisleader]
		log.Warning("leader[%v].prepare.down", leader.getID())
		leader.Stop()
	}

	// 5. wait rafts[1] elected as new leader
	{
		MockWaitLeaderEggs(rafts, 1)

		var got State
		whoisleader = 0
		want := (LEADER + FOLLOWER + STOPPED)
		for i, raft := range rafts {
			got += raft.getState()
			if raft.getState() == LEADER {
				whoisleader = i
			}
		}

		// [LEADER, FOLLOWER, STOPPED]
		assert.Equal(t, want, got)

		// leader should be rafts[GTIDBIDX]
		assert.Equal(t, whoisleader, GTIDBIDX)
	}
}

// TEST EFFECTS:
// GTIDE with GetGTID error can't be elected as a leader
//
// TEST PROCESSES:
// 1. set rafts GTID
//    1.0 rafts[0]  with MockGTIDE{nil, err}
//    1.1 rafts[1]  with MockGTIDB{Master_Log_File = "mysql-bin.000001", Read_Master_Log_Pos = 123}
//    1.2 rafts[2]  with MockGTIDC{Master_Log_File = "mysql-bin.000001", Read_Master_Log_Pos = 124}
// 2. Start 3 rafts state as FOLLOWER
// 3. wait rafts[2] elected as leader
// 4. Stop all rafts
func TestRaftWithFollowerGetSlaveGTIDError(t *testing.T) {
	var whoisleader int

	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	port := common.RandomPort(8000, 9000)
	_, rafts, cleanup := MockRafts(log, port, 3, -1)
	defer cleanup()

	GTIDAIDX := 0
	GTIDBIDX := 1
	GTIDCIDX := 2

	// 1. set rafts GTID
	//    1.0 rafts[0]  with MockGTIDE{nil, err}
	//    1.1 rafts[1]  with MockGTIDB{Master_Log_File = "mysql-bin.000001", Read_Master_Log_Pos = 123}
	//    1.2 rafts[2]  with MockGTIDC{Master_Log_File = "mysql-bin.000001", Read_Master_Log_Pos = 124}
	{
		rafts[GTIDAIDX].mysql.SetMysqlHandler(mysql.NewMockGTIDError())
		rafts[GTIDBIDX].mysql.SetMysqlHandler(mysql.NewMockGTIDB())
		rafts[GTIDCIDX].mysql.SetMysqlHandler(mysql.NewMockGTIDC())
	}

	// 2. Start 3 rafts state as FOLLOWER
	{
		for _, raft := range rafts {
			raft.Start()
		}

		var got State
		want := (FOLLOWER + FOLLOWER + FOLLOWER)
		for _, raft := range rafts {
			got += raft.getState()
		}

		// [FOLLOWER, FOLLOWER, FOLLOWER]
		assert.Equal(t, want, got)
	}

	// 3. wait rafts[GTIDCIDX] elected as leader
	{
		MockWaitLeaderEggs(rafts, 1)
		var got State
		want := (LEADER + FOLLOWER + FOLLOWER)
		for i, raft := range rafts {
			got += raft.getState()
			if raft.getState() == LEADER {
				whoisleader = i
			}
		}
		// [LEADER, FOLLOWER, FOLLOWER]
		assert.Equal(t, want, got)

		// leader should be rafts[GTIDCIDX]
		assert.Equal(t, whoisleader, GTIDCIDX)
	}
}

// TEST EFFECTS:
// test the leader purge binlog
//
// TEST PROCESSES:
// 1. set rafts GTID
//    1.0 rafts[0]  with MockGTID_X1{Master_Log_File = "mysql-bin.000001", Read_Master_Log_Pos = 123}
//    1.1 rafts[1]  with MockGTID_X3{Master_Log_File = "mysql-bin.000003", Read_Master_Log_Pos = 123}
//    1.2 rafts[2]  with MockGTID_X5{Master_Log_File = "mysql-bin.000005", Read_Master_Log_Pos = 123}
// 2. Start 3 rafts state as FOLLOWER
// 3. wait rafts[2] elected as leader
// 4. check rafts[2] stats
// 5. Stop all rafts
func TestRaftLeaderPurgeBinlog(t *testing.T) {
	conf := config.DefaultRaftConfig()
	conf.PurgeBinlogInterval = 1
	conf.MetaDatadir = "/tmp/"

	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	port := common.RandomPort(8000, 9000)
	_, rafts, cleanup := MockRaftsWithConfig(log, conf, port, 3, -1)
	defer cleanup()

	// 1. set rafts GTID
	//    1.0 rafts[0]  with MockGTIDB{Master_Log_File = "mysql-bin.000001", Read_Master_Log_Pos = 123}
	//    1.1 rafts[1]  with MockGTIDB{Master_Log_File = "mysql-bin.000003", Read_Master_Log_Pos = 123}
	//    1.2 rafts[2]  with MockGTIDC{Master_Log_File = "mysql-bin.000005", Read_Master_Log_Pos = 123}
	{
		rafts[0].mysql.SetMysqlHandler(mysql.NewMockGTIDX1())
		rafts[1].mysql.SetMysqlHandler(mysql.NewMockGTIDX3())
		rafts[2].mysql.SetMysqlHandler(mysql.NewMockGTIDX5())
	}

	// 2. Start 3 rafts state as FOLLOWER
	for _, raft := range rafts {
		raft.Start()
	}

	// leader
	{
		var got State
		var whoisleader int

		MockWaitLeaderEggs(rafts, 1)
		MockWaitLeaderEggs(rafts, 0)
		want := (LEADER + FOLLOWER + FOLLOWER)
		for i, raft := range rafts {
			got += raft.getState()
			if raft.getState() == LEADER {
				whoisleader = i
			}
		}

		assert.Equal(t, want, got)
		assert.Equal(t, 2, whoisleader)
		MockWaitLeaderEggs(rafts, 0)
		MockWaitLeaderEggs(rafts, 0)
		assert.NotZero(t, rafts[2].stats.LeaderPurgeBinlogs)
	}

	// disable purge binlog
	{
		rafts[2].SetSkipPurgeBinlog(true)
		purged := rafts[2].stats.LeaderPurgeBinlogs

		MockWaitLeaderEggs(rafts, 0)
		MockWaitLeaderEggs(rafts, 0)

		want := purged
		got := rafts[2].stats.LeaderPurgeBinlogs
		assert.Equal(t, want, got)
	}

	// enable purge binlog
	{
		rafts[2].SetSkipPurgeBinlog(false)
		purged := rafts[2].stats.LeaderPurgeBinlogs

		MockWaitLeaderEggs(rafts, 0)
		MockWaitLeaderEggs(rafts, 0)

		want := purged
		got := rafts[2].stats.LeaderPurgeBinlogs
		assert.NotEqual(t, want, got)
	}

	// disable purge by setting conf.PurgeBinlogDisabled=true
	{
		conf.PurgeBinlogDisabled = true
		purged := rafts[2].stats.LeaderPurgeBinlogs

		MockWaitLeaderEggs(rafts, 0)
		MockWaitLeaderEggs(rafts, 0)

		want := purged
		got := rafts[2].stats.LeaderPurgeBinlogs
		assert.Equal(t, want, got)
	}
}

// TEST EFFECTS:
// test the leader check semi-sync
//
// TEST PROCESSES:
// 1. set rafts GTID
//    1.0 rafts[0]  with MockGTID_X1{Master_Log_File = "mysql-bin.000001", Read_Master_Log_Pos = 123}
//    1.1 rafts[1]  with MockGTID_X3{Master_Log_File = "mysql-bin.000003", Read_Master_Log_Pos = 123}
//    1.2 rafts[2]  with MockGTID_X5{Master_Log_File = "mysql-bin.000005", Read_Master_Log_Pos = 123}
// 2. Start 3 rafts state as FOLLOWER
// 3. wait rafts[2] elected as leader
// 4. check rafts[2] skipCheckSemiSync
// 5. Stop all rafts
func TestRaftLeaderCheckSemiSync(t *testing.T) {
	conf := config.DefaultRaftConfig()
	conf.MetaDatadir = "/tmp/"

	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	port := common.RandomPort(8000, 9000)
	_, rafts, cleanup := MockRaftsWithConfig(log, conf, port, 3, -1)
	defer cleanup()

	// 1. set rafts GTID
	//    1.0 rafts[0]  with MockGTIDB{Master_Log_File = "mysql-bin.000001", Read_Master_Log_Pos = 123}
	//    1.1 rafts[1]  with MockGTIDB{Master_Log_File = "mysql-bin.000003", Read_Master_Log_Pos = 123}
	//    1.2 rafts[2]  with MockGTIDC{Master_Log_File = "mysql-bin.000005", Read_Master_Log_Pos = 123}
	{
		rafts[0].mysql.SetMysqlHandler(mysql.NewMockGTIDX1())
		rafts[1].mysql.SetMysqlHandler(mysql.NewMockGTIDX3())
		rafts[2].mysql.SetMysqlHandler(mysql.NewMockGTIDX5())
	}

	// 2. Start 3 rafts state as FOLLOWER
	for _, raft := range rafts {
		raft.Start()
	}

	// leader
	{
		var got State
		var whoisleader int

		MockWaitLeaderEggs(rafts, 1)
		MockWaitLeaderEggs(rafts, 0)
		want := (LEADER + FOLLOWER + FOLLOWER)
		for i, raft := range rafts {
			got += raft.getState()
			if raft.getState() == LEADER {
				whoisleader = i
			}
		}

		assert.Equal(t, want, got)
		assert.Equal(t, 2, whoisleader)

		// wait for check semi-sync to be invoked and skipCheckSemiSync changed
		time.Sleep(time.Millisecond * time.Duration(rafts[0].getElectionTimeout()*16))
		assert.Equal(t, false, rafts[2].skipCheckSemiSync)
	}

	// disable check semi-sync
	{
		rafts[2].SetSkipCheckSemiSync(true)
		time.Sleep(time.Millisecond * time.Duration(rafts[0].getElectionTimeout()*16))
		assert.Equal(t, true, rafts[2].skipCheckSemiSync)
	}

	// enable check semi-sync
	{
		rafts[2].SetSkipCheckSemiSync(false)

		MockWaitLeaderEggs(rafts, 0)
		MockWaitLeaderEggs(rafts, 0)

		assert.Equal(t, false, rafts[2].skipCheckSemiSync)
	}
}

// TEST EFFECTS:
// test the follower change master to failed
//
// TEST PROCESSES:
// 1.  set rafts GTID
//     1.0 rafts[0]  with MockGTID_X1{Master_Log_File = "mysql-bin.000001", Read_Master_Log_Pos = 123}
//     1.1 rafts[1]  with MockGTID_X3{Master_Log_File = "mysql-bin.000003", Read_Master_Log_Pos = 123}
//     1.2 rafts[2]  with MockGTID_X5{Master_Log_File = "mysql-bin.000005", Read_Master_Log_Pos = 123}
// 2.  Start 3 rafts state as FOLLOWER
// 3.  wait rafts[2] elected as leader
// 4.  set rafts[2].MySQL to MockGTIDE(mock MySQL down)
// 5.  wait rafts[1] elected as new-leader
// 6.  wait rafts[1] send heartbeats to rafts[2]
// 7.  rafts[2].MySQL change-master-to rafts[1].MySQL failed
// 8.  check rafts[2].Leader is NULL, because change-master-to failed
// 9.  set rafts[2].MySQL to OK
// 10. rafts[2].MySQL change-master-to rafts[1].MySQL OK
// 11. check rafts[2].Leader is rafts[1]
func TestRaftChangeMasterToFail(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	port := common.RandomPort(8000, 9000)
	_, rafts, cleanup := MockRafts(log, port, 3, -1)
	defer cleanup()

	// 1. set rafts GTID
	//    1.0 rafts[0]  with MockGTIDX1{Master_Log_File = "mysql-bin.000001", Read_Master_Log_Pos = 123}
	//    1.1 rafts[1]  with MockGTIDX3{Master_Log_File = "mysql-bin.000003", Read_Master_Log_Pos = 123}
	//    1.2 rafts[2]  with MockGTIDX5{Master_Log_File = "mysql-bin.000005", Read_Master_Log_Pos = 123}
	{
		rafts[0].mysql.SetMysqlHandler(mysql.NewMockGTIDX1())
		rafts[1].mysql.SetMysqlHandler(mysql.NewMockGTIDX3())
		rafts[2].mysql.SetMysqlHandler(mysql.NewMockGTIDX5())
	}

	// 2. Start 3 rafts state as FOLLOWER
	for _, raft := range rafts {
		raft.Start()
	}

	// check new leader is rafts[2]
	{
		var got State
		var whoisleader int

		MockWaitLeaderEggs(rafts, 1)
		MockWaitLeaderEggs(rafts, 0)
		want := (LEADER + FOLLOWER + FOLLOWER)
		for i, raft := range rafts {
			got += raft.getState()
			if raft.getState() == LEADER {
				whoisleader = i
			}
		}

		assert.Equal(t, want, got)
		assert.Equal(t, 2, whoisleader)
		// check leader is rafts[2]
		assert.Equal(t, rafts[2].id, rafts[2].leader)
		MockWaitLeaderEggs(rafts, 0)
	}

	// set leader MySQL to Error
	{
		rafts[2].mysql.SetMysqlHandler(mysql.NewMockGTIDError())
		MockWaitMySQLPingTimeout()

		var got State
		var whoisleader int

		MockWaitLeaderEggs(rafts, 0)
		MockWaitLeaderEggs(rafts, 1)
		want := (LEADER + FOLLOWER + FOLLOWER)
		for i, raft := range rafts {
			got += raft.getState()
			if raft.getState() == LEADER {
				whoisleader = i
			}
		}

		assert.Equal(t, want, got)
		assert.Equal(t, 1, whoisleader)
	}

	// check new leader is rafts[1]
	{
		var got State
		var whoisleader int

		MockWaitLeaderEggs(rafts, 0)
		want := (LEADER + FOLLOWER + FOLLOWER)
		for i, raft := range rafts {
			got += raft.getState()
			if raft.getState() == LEADER {
				whoisleader = i
			}
		}

		assert.Equal(t, want, got)
		assert.Equal(t, 1, whoisleader)

		// check rafts[2].leader is NULL, because change-master-to failed
		assert.Equal(t, "", rafts[2].leader)
	}

	// set rafts[2].MySQL to OK
	{
		rafts[2].mysql.SetMysqlHandler(mysql.NewMockGTIDX5())
		MockWaitMySQLPingTimeout()

		var got State
		var whoisleader int

		MockWaitLeaderEggs(rafts, 0)
		MockWaitLeaderEggs(rafts, 1)
		want := (LEADER + FOLLOWER + FOLLOWER)
		for i, raft := range rafts {
			got += raft.getState()
			if raft.getState() == LEADER {
				whoisleader = i
			}
		}

		assert.Equal(t, want, got)
		assert.Equal(t, 1, whoisleader)

		// check rafts[2].leader is rafts[1], because change-master-to is OK
		assert.Equal(t, rafts[1].id, rafts[2].leader)
	}
}

// TEST EFFECTS:
// test the 1nodes of cluster
//
// TEST PROCESSES:
// 1. Start 1 rafts
// 2. check leader
func TestRaft1Nodes(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	port := common.RandomPort(8000, 9000)
	_, rafts, cleanup := MockRafts(log, port, 1, -1)
	defer cleanup()
	raft := rafts[0]

	// Start rafts
	for _, raft := range rafts {
		raft.Start()
	}

	{
		want := 1
		got := raft.GetMembers()
		assert.Equal(t, want, got)

		want = 1
		got = raft.GetQuorums()
		assert.Equal(t, want, got)
	}

	// 2. wait leader election from 3 FOLLOWER
	{
		var want, got State
		MockWaitLeaderEggs(rafts, 0)
		MockWaitLeaderEggs(rafts, 0)
		MockWaitLeaderEggs(rafts, 0)
		want = (FOLLOWER)
		for _, raft := range rafts {
			got += raft.getState()
		}
		// [FOLLOWER]
		assert.Equal(t, want, got)
	}
}

/*
// TEST EFFECTS:
// test the 2nodes of cluster
//
// TEST PROCESSES:
// 1. Start 2 rafts
// 2. check quorums and leader
func TestRaft2Nodes(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	port := common.RandomPort(8000, 9000)
	_, rafts, cleanup := MockRafts(log, port, 2, -1)
	defer cleanup()
	raft := rafts[0]

	// Start rafts
	for _, raft := range rafts {
		raft.Start()
	}

	{
		want := 2
		got := raft.GetMembers()
		assert.Equal(t, want, got)

		want = 2
		got = raft.GetQuorums()
		assert.Equal(t, want, got)
	}

	// 2. wait leader election from 3 FOLLOWER
	{
		var want, got State
		MockWaitLeaderEggs(rafts, 1)
		want = (LEADER + FOLLOWER)
		for _, raft := range rafts {
			got += raft.getState()
		}
		// [LEADER, FOLLOWER]
		assert.Equal(t, want, got)
	}

	time.Sleep(time.Millisecond * 5000)
}

// TEST EFFECTS:
// test the 2 nodes leader election
//
// TEST PROCESSES:
// 1.  set rafts GTID
//     1.0 rafts[0]  with MockGTID_X1{Master_Log_File = "mysql-bin.000001", Read_Master_Log_Pos = 123}
//     1.1 rafts[1]  with MockGTID_X3{Master_Log_File = "mysql-bin.000003", Read_Master_Log_Pos = 123}
// 2.  Start 2 rafts state as FOLLOWER
// 3.  wait rafts[1] elected as leader
// 4.  set rafts[1].MySQL to MockGTIDE(mock MySQL down)
// 5.  wait rafts[0] elected as new-leader
func TestRaft2NodesWithGTID(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	port := common.RandomPort(8000, 9000)
	_, rafts, cleanup := MockRafts(log, port, 2, -1)
	defer cleanup()

	// 1. set rafts GTID
	//    1.0 rafts[0]  with MockGTIDB{Master_Log_File = "mysql-bin.000001", Read_Master_Log_Pos = 123}
	//    1.1 rafts[1]  with MockGTIDB{Master_Log_File = "mysql-bin.000003", Read_Master_Log_Pos = 123}
	{
		rafts[0].mysql.SetMysqlHandler(mysql.NewMockGTIDX1())
		rafts[1].mysql.SetMysqlHandler(mysql.NewMockGTIDX3())
	}

	// 2. Start 2 rafts state as FOLLOWER
	for _, raft := range rafts {
		raft.Start()
	}

	// check new leader is rafts[1]
	{
		var got State
		var whoisleader int

		MockWaitLeaderEggs(rafts, 1)
		want := (LEADER + FOLLOWER)
		for i, raft := range rafts {
			got += raft.getState()
			if raft.getState() == LEADER {
				whoisleader = i
			}
		}

		assert.Equal(t, want, got)
		assert.Equal(t, 1, whoisleader)
	}

	// set leader MySQL to Error
	{
		rafts[1].mysql.SetMysqlHandler(mysql.NewMockGTIDError())
		MockWaitMySQLPingTimeout()

		var got State
		var whoisleader int

		MockWaitLeaderEggs(rafts, 0)
		MockWaitLeaderEggs(rafts, 1)
		want := (LEADER + FOLLOWER)
		time.Sleep(time.Second + 1)
		for i, raft := range rafts {
			got += raft.getState()
			if raft.getState() == LEADER {
				whoisleader = i
			}
		}

		assert.Equal(t, want, got)
		assert.Equal(t, 0, whoisleader)
	}
}
*/

// TEST EFFECTS:
// test the leader heartbeat acks less than the quorum.
func TestRaftLeaderAckLessThanQuorum(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	port := common.RandomPort(8000, 9000)
	_, rafts, cleanup := MockRafts(log, port, 3, -1)
	defer cleanup()

	// 1. set rafts GTID
	//    1.0 rafts[0]  with MockGTIDB{Master_Log_File = "mysql-bin.000001", Read_Master_Log_Pos = 123}
	{
		rafts[0].mysql.SetMysqlHandler(mysql.NewMockGTIDX1())
		rafts[1].mysql.SetMysqlHandler(mysql.NewMockGTIDX1())
		rafts[2].mysql.SetMysqlHandler(mysql.NewMockGTIDX1())
	}

	// 2. Start 3 rafts state as FOLLOWER
	for _, raft := range rafts {
		raft.Start()
	}

	var got State
	var whoisleader int
	// check new leader
	{

		MockWaitLeaderEggs(rafts, 1)
		want := (LEADER + FOLLOWER + FOLLOWER)
		for i, raft := range rafts {
			got += raft.getState()
			if raft.getState() == LEADER {
				whoisleader = i
			}
		}

		assert.Equal(t, want, got)
		MockWaitLeaderEggs(rafts, 0)
	}

	{
		leader := rafts[whoisleader]
		leader.L.setProcessHeartbeatResponseHandler(leader.mockLeaderProcessSendHeartbeatResponse)
		for i := 0; i < 11; i++ {
			MockWaitLeaderEggs(rafts, 0)
		}
	}
}

// TEST EFFECTS:
// test the leader state init WaitUntilAfterGTID failed.
//
// TEST PROCESSES:
// 1.  set rafts GTID
//     1.0 rafts[0]  with MockGTID_X1{Master_Log_File = "mysql-bin.000001", Read_Master_Log_Pos = 123}
//     1.1 rafts[1]  with MockGTID_X3{Master_Log_File = "mysql-bin.000003", Read_Master_Log_Pos = 123}
//     1.2 rafts[2]  with MockGTID_X5{Master_Log_File = "mysql-bin.000005", Read_Master_Log_Pos = 123} with WaitUntilAfterGTID error
// 2.  Start 3 rafts state as FOLLOWER
// 3.  wait rafts[2] elected as leader
// 4.  rafts[2] WaitUntilAfterGTID error degrade to FOLLOWER

func TestRaftLeaderWaitUntilAfterGTIDError(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	port := common.RandomPort(8000, 9000)
	_, rafts, cleanup := MockRafts(log, port, 3, -1)
	defer cleanup()

	// 1. set rafts GTID
	//    1.0 rafts[0]  with MockGTIDB{Master_Log_File = "mysql-bin.000001", Read_Master_Log_Pos = 123}
	//    1.1 rafts[1]  with MockGTIDB{Master_Log_File = "mysql-bin.000003", Read_Master_Log_Pos = 123}
	//    1.2 rafts[2]  with MockGTIDC{Master_Log_File = "mysql-bin.000005", Read_Master_Log_Pos = 123} with WaitUntilAfterGTID error
	{
		rafts[0].mysql.SetMysqlHandler(mysql.NewMockGTIDX1())
		rafts[1].mysql.SetMysqlHandler(mysql.NewMockGTIDX3())
		rafts[2].mysql.SetMysqlHandler(mysql.NewMockGTIDX5WaitUntilAfterGTIDError())
	}

	// 2. Start 3 rafts state as FOLLOWER
	for _, raft := range rafts {
		raft.Start()
	}

	// check new leader is rafts[2]
	{
		var got State
		time.Sleep(time.Millisecond * 3000)
		want := (LEADER + FOLLOWER + FOLLOWER)
		for _, raft := range rafts {
			got += raft.getState()
		}
		assert.True(t, got < want)
	}
}

// TEST EFFECTS:
// test the leader state init ChangeMasterToFn failed.
//
// TEST PROCESSES:
// 1.  set rafts GTID
//     1.0 rafts[0]  with MockGTID_X1{Master_Log_File = "mysql-bin.000001", Read_Master_Log_Pos = 123}
//     1.1 rafts[1]  with MockGTID_X3{Master_Log_File = "mysql-bin.000003", Read_Master_Log_Pos = 123}
//     1.2 rafts[2]  with MockGTID_X5{Master_Log_File = "mysql-bin.000005", Read_Master_Log_Pos = 123} with ChangeToMaster error
// 2.  Start 3 rafts state as FOLLOWER
// 3.  wait rafts[2] elected as leader
// 4.  rafts[2] ChangeToMaster error, degrade to FOLLOWER

func TestRaftLeaderChangeToMasterError(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	port := common.RandomPort(8000, 9000)
	_, rafts, cleanup := MockRafts(log, port, 3, -1)
	defer cleanup()

	// 1. set rafts GTID
	//    1.0 rafts[0]  with MockGTIDB{Master_Log_File = "mysql-bin.000001", Read_Master_Log_Pos = 123}
	//    1.1 rafts[1]  with MockGTIDB{Master_Log_File = "mysql-bin.000003", Read_Master_Log_Pos = 123}
	//    1.2 rafts[2]  with MockGTIDC{Master_Log_File = "mysql-bin.000005", Read_Master_Log_Pos = 123} with ChangeToMaster error
	{
		rafts[0].mysql.SetMysqlHandler(mysql.NewMockGTIDX1())
		rafts[1].mysql.SetMysqlHandler(mysql.NewMockGTIDX3())
		rafts[2].mysql.SetMysqlHandler(mysql.NewMockGTIDX5ChangeToMasterError())
	}

	// 2. Start 3 rafts state as FOLLOWER
	for _, raft := range rafts {
		raft.Start()
	}

	// check new leader is rafts[2]
	{
		var got State
		time.Sleep(time.Millisecond * 3000)
		want := (LEADER + FOLLOWER + FOLLOWER)
		for _, raft := range rafts {
			got += raft.getState()
		}
		assert.True(t, got < want)
	}
}

// TEST EFFECTS:
// test election under LEARNER in the minority
//
// TEST PROCESSES:
// 1.  set rafts GTID
//     1.0 rafts[0]  with MockGTID_X1{Master_Log_File = "mysql-bin.000001", Read_Master_Log_Pos = 123}
//     1.1 rafts[1]  with MockGTID_X3{Master_Log_File = "mysql-bin.000003", Read_Master_Log_Pos = 123}
//     1.2 rafts[2]  with MockGTID_X5{Master_Log_File = "mysql-bin.000005", Read_Master_Log_Pos = 123}
// 2.  Start 3 rafts state as FOLLOWER
// 3.  wait rafts[2] elected as leader
// 4.  set rafts[0] to LEARNER
// 5.  wait a few election cycles, the leader remains the same
func TestRaftElectionUnderLearnerInMinority(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	port := common.RandomPort(8000, 9000)
	_, rafts, cleanup := MockRafts(log, port, 3, -1)
	defer cleanup()

	// 1. set rafts GTID
	//    1.0 rafts[0]  with MockGTIDB{Master_Log_File = "mysql-bin.000001", Read_Master_Log_Pos = 123}
	//    1.1 rafts[1]  with MockGTIDB{Master_Log_File = "mysql-bin.000003", Read_Master_Log_Pos = 123}
	//    1.2 rafts[2]  with MockGTIDC{Master_Log_File = "mysql-bin.000005", Read_Master_Log_Pos = 123}
	{
		rafts[0].mysql.SetMysqlHandler(mysql.NewMockGTIDX1())
		rafts[1].mysql.SetMysqlHandler(mysql.NewMockGTIDX3())
		rafts[2].mysql.SetMysqlHandler(mysql.NewMockGTIDX5())
	}

	// 2. start 3 rafts state as FOLLOWER
	for _, raft := range rafts {
		raft.Start()
	}

	// 3. check new leader is rafts[2]
	var whoisleader int
	{
		var got State
		MockWaitLeaderEggs(rafts, 1)
		want := (LEADER + FOLLOWER + FOLLOWER)
		for i, raft := range rafts {
			got += raft.getState()
			if raft.getState() == LEADER {
				whoisleader = i
			}
		}
		assert.Equal(t, want, got)
		assert.Equal(t, whoisleader, 2)
	}

	// 4. set rafts[0] to LEARNER and set rafts[2] to FOLLOWER
	MockStateTransition(rafts[whoisleader], FOLLOWER)
	MockStateTransition(rafts[0], LEARNER)

	// 5. wait a few election cycles, the leader remains the same
	{
		var got State
		time.Sleep(time.Millisecond * 3000)
		want := (LEADER + FOLLOWER + LEARNER)
		for i, raft := range rafts {
			got += raft.getState()
			if raft.getState() == LEADER {
				whoisleader = i
			}
		}
		assert.Equal(t, want, got)
		assert.Equal(t, whoisleader, 2)
	}
}

// TEST EFFECTS:
// test election under Follower and Candidate alternate
//
// TEST PROCESSES:
// 1.  set rafts GTID
//     1.0 rafts[0]  with MockGTID_X1{Master_Log_File = "mysql-bin.000001", Read_Master_Log_Pos = 123}
//     1.1 rafts[1]  with MockGTID_X3{Master_Log_File = "mysql-bin.000003", Read_Master_Log_Pos = 123}
//     1.2 rafts[2]  with MockGTID_X3{Master_Log_File = "mysql-bin.000003", Read_Master_Log_Pos = 123}
// 2.  start rafts[0] state as CANDIDATE
// 3.  wait 30 times the election timeout
// 4.  start rafts[1] as FOLLOWER and rafts[2] as IDLE
//                   InvalidGITD
//     rafts[0]: C --------------> F --------------> C -> ... -> F
//                                    InvalidViewID
//     rafts[1]: F --------------> C --------------> F -> ... -> C

// 5.  wait 8 times the election timeout
// 6.  check if rafts[1] is the leader
func TestRaftElectionUnderFollowerAndCandidateAlternate(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	port := common.RandomPort(8000, 9000)
	_, rafts, cleanup := MockRafts(log, port, 3, -1)
	defer cleanup()

	// 1. set rafts GTID
	{
		rafts[0].mysql.SetMysqlHandler(mysql.NewMockGTIDX1())
		rafts[1].mysql.SetMysqlHandler(mysql.NewMockGTIDX3())
		rafts[2].mysql.SetMysqlHandler(mysql.NewMockGTIDX3())
	}

	// 2. start rafts[0] state as CANDIDATE
	{
		rafts[0].Start()
		MockStateTransition(rafts[0], CANDIDATE)
	}

	// 3. wait 30 times the election timeout
	time.Sleep(time.Millisecond * time.Duration(rafts[0].getElectionTimeout()*30))

	// 4. start rafts[1] as FOLLOWER and rafts[2] as IDLE
	{
		rafts[1].Start()
		MockStateTransition(rafts[1], FOLLOWER)
		MockStateTransition(rafts[0], CANDIDATE)
		rafts[2].Start()
		MockStateTransition(rafts[2], IDLE)
	}

	// 5. wait 8 times the election timeout
	time.Sleep(time.Millisecond * time.Duration(rafts[0].getElectionTimeout()*8))

	//6. check if rafts[1] is the leader
	{
		var got State
		var whoisleader int
		want := (FOLLOWER + LEADER + IDLE)
		for i, raft := range rafts {
			got += raft.getState()
			if raft.getState() == LEADER {
				whoisleader = i
			}
		}
		assert.Equal(t, want, got)
		assert.Equal(t, whoisleader, 1)
	}
}
