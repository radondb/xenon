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
	"testing"
	"xbase/common"
	"xbase/xlog"

	"github.com/stretchr/testify/assert"
)

// TEST EFFECTS:
// test a hadisable command from the client
//
// TEST PROCESSES:
// 1. Start rpc server
// 2. send command to rpc server
// 3. check the response
func TestRaftRPCHA(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	port := common.RandomPort(8000, 9000)
	names, rafts, scleanup := MockRafts(log, port, 3, -1)
	defer scleanup()

	// 1. Start 3 rafts state as FOLLOWER
	{
		for _, raft := range rafts {
			raft.Start()
		}

		var want, got State
		got = 0
		want = (FOLLOWER + FOLLOWER + FOLLOWER)
		for _, raft := range rafts {
			got += raft.getState()
		}

		// [FOLLOWER, FOLLOWER, FOLLOWER]
		assert.Equal(t, want, got)
	}

	// 2. all rafts to ha enable(invalid reqeust)
	{
		for i := range names {
			c, cleanup := MockGetClient(t, names[i])
			defer cleanup()

			method := model.RPCHAEnable
			req := model.NewHARPCRequest()
			rsp := model.NewHARPCResponse(model.OK)
			err := c.Call(method, req, rsp)
			assert.Nil(t, err)

			want := model.OK
			got := rsp.RetCode
			assert.Equal(t, want, got)
		}
	}

	// 3. all rafts to ha disable
	{
		for i := range rafts {
			c, cleanup := MockGetClient(t, names[i])
			defer cleanup()

			method := model.RPCHADisable
			req := model.NewHARPCRequest()
			rsp := model.NewHARPCResponse(model.OK)
			err := c.Call(method, req, rsp)
			assert.Nil(t, err)

			want := model.OK
			got := rsp.RetCode
			assert.Equal(t, want, got)
		}
	}

	// 4. check
	{
		MockWaitLeaderEggs(rafts, 0)

		var want, got State
		got = 0
		want = (IDLE + IDLE + IDLE)
		for _, raft := range rafts {
			got += raft.getState()
		}
		// [IDLE, IDLE, IDLE]
		assert.Equal(t, want, got)
	}

	// 5. all rafts to HaEnable
	{
		for i := range names {
			c, cleanup := MockGetClient(t, names[i])
			defer cleanup()

			method := model.RPCHAEnable
			req := model.NewHARPCRequest()
			rsp := model.NewHARPCResponse(model.OK)
			err := c.Call(method, req, rsp)
			assert.Nil(t, err)

			want := model.OK
			got := rsp.RetCode
			assert.Equal(t, want, got)
		}
	}

	// 6. check
	{
		MockWaitLeaderEggs(rafts, 1)

		var want, got State
		got = 0
		want = (LEADER + FOLLOWER + FOLLOWER)
		for _, raft := range rafts {
			got += raft.getState()
		}
		// [LEADER, FOLLOWER, FOLLOWER]
		assert.Equal(t, want, got)
	}

	// 7. enable ha all
	{
		for i := range names {
			c, cleanup := MockGetClient(t, names[i])
			defer cleanup()

			method := model.RPCHAEnable
			req := model.NewHARPCRequest()
			rsp := model.NewHARPCResponse(model.OK)
			err := c.Call(method, req, rsp)
			assert.Nil(t, err)

			want := model.OK
			got := rsp.RetCode
			assert.Equal(t, want, got)
		}
	}

	// 8. all raftsto ha disable
	{
		for i := range names {
			c, cleanup := MockGetClient(t, names[i])
			defer cleanup()

			method := model.RPCHADisable
			req := model.NewHARPCRequest()
			rsp := model.NewHARPCResponse(model.OK)
			err := c.Call(method, req, rsp)
			assert.Nil(t, err)
			want := model.OK
			got := rsp.RetCode
			assert.Equal(t, want, got)
		}
	}
}

// TEST EFFECTS:
// test a hasetlearner command from follower by the client
//
// TEST PROCESSES:
// 1. Start rpc server
// 2. send command to rpc server
// 3. check the response
func TestRaftRPCHASetLearnerFromFollower(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	port := common.RandomPort(8000, 9000)
	names, rafts, scleanup := MockRafts(log, port, 3, -1)
	learner := 2
	defer scleanup()

	// 1. Start 3 rafts state as FOLLOWER
	{
		for _, raft := range rafts {
			raft.Start()
		}

		var want, got State
		got = 0
		want = (FOLLOWER + FOLLOWER + FOLLOWER)
		for _, raft := range rafts {
			got += raft.getState()
		}

		// [FOLLOWER, FOLLOWER, FOLLOWER]
		assert.Equal(t, want, got)
	}

	// 2. set rafts[2] to LEARNER
	{
		c, cleanup := MockGetClient(t, names[learner])
		defer cleanup()

		method := model.RPCHASetLearner
		req := model.NewHARPCRequest()
		rsp := model.NewHARPCResponse(model.OK)
		err := c.Call(method, req, rsp)
		assert.Nil(t, err)

		want := model.OK
		got := rsp.RetCode
		assert.Equal(t, want, got)
	}

	// 3. check
	{
		MockWaitLeaderEggs(rafts, 1)

		var want, got State
		got = 0
		want = (LEADER + FOLLOWER + LEARNER)
		for _, raft := range rafts {
			got += raft.getState()
		}
		// [LEADER, FOLLOWER, LEARNER]
		assert.Equal(t, want, got)
	}

	// 4. enable ha for rafts[2]
	{
		c, cleanup := MockGetClient(t, names[learner])
		defer cleanup()

		method := model.RPCHAEnable
		req := model.NewHARPCRequest()
		rsp := model.NewHARPCResponse(model.OK)
		err := c.Call(method, req, rsp)
		assert.Nil(t, err)

		want := model.OK
		got := rsp.RetCode
		assert.Equal(t, want, got)
	}

	// 5. check
	{
		MockWaitLeaderEggs(rafts, 1)

		var want, got State
		got = 0
		want = (LEADER + FOLLOWER + FOLLOWER)
		for _, raft := range rafts {
			got += raft.getState()
		}
		// [LEADER, FOLLOWER, FOLLOWER]
		assert.Equal(t, want, got)
	}
}

// TEST EFFECTS:
// test a hasetlearner command from invalid by the client
//
// TEST PROCESSES:
// 1. Start rpc server
// 2. send command to rpc server
// 3. check the response
func TestRaftRPCHASetLearnerFromInvalid(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	port := common.RandomPort(8000, 9000)
	names, rafts, scleanup := MockRafts(log, port, 3, -1)
	learner := 2
	defer scleanup()

	// 1. Start 3 rafts state as FOLLOWER and set rafts[2] to INVALID
	{
		for _, raft := range rafts {
			raft.Start()
		}

		MockStateTransition(rafts[learner], INVALID)

		var want, got State
		got = 0
		want = (FOLLOWER + FOLLOWER + INVALID)
		for _, raft := range rafts {
			got += raft.getState()
		}

		// [FOLLOWER, FOLLOWER, INVALID]
		assert.Equal(t, want, got)
	}

	// 2. set rafts[2] to LEARNER
	{
		c, cleanup := MockGetClient(t, names[learner])
		defer cleanup()

		method := model.RPCHASetLearner
		req := model.NewHARPCRequest()
		rsp := model.NewHARPCResponse(model.OK)
		err := c.Call(method, req, rsp)
		assert.Nil(t, err)

		want := model.OK
		got := rsp.RetCode
		assert.Equal(t, want, got)
	}

	// 3. check
	{
		MockWaitLeaderEggs(rafts, 1)

		var want, got State
		got = 0
		want = (LEADER + FOLLOWER + LEARNER)
		for _, raft := range rafts {
			got += raft.getState()
		}
		// [LEADER, FOLLOWER, LEARNER]
		assert.Equal(t, want, got)
	}

	// 4. enable ha for rafts[2]
	{
		c, cleanup := MockGetClient(t, names[learner])
		defer cleanup()

		method := model.RPCHAEnable
		req := model.NewHARPCRequest()
		rsp := model.NewHARPCResponse(model.OK)
		err := c.Call(method, req, rsp)
		assert.Nil(t, err)

		want := model.OK
		got := rsp.RetCode
		assert.Equal(t, want, got)
	}

	// 5. check
	{
		MockWaitLeaderEggs(rafts, 1)

		var want, got State
		got = 0
		want = (LEADER + FOLLOWER + FOLLOWER)
		for _, raft := range rafts {
			got += raft.getState()
		}
		// [LEADER, FOLLOWER, FOLLOWER]
		assert.Equal(t, want, got)
	}
}

// TEST EFFECTS:
// test a hasetlearner command from idle by the client
//
// TEST PROCESSES:
// 1. Start rpc server
// 2. send command to rpc server
// 3. check the response
func TestRaftRPCHASetLearnerFromIdle(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	port := common.RandomPort(8000, 9000)
	names, rafts, scleanup := MockRafts(log, port, 3, -1)
	learner := 2
	defer scleanup()

	// 1. Start 3 rafts state as FOLLOWER and set rafts[2] to IDLE
	{
		for _, raft := range rafts {
			raft.Start()
		}

		MockStateTransition(rafts[learner], IDLE)

		var want, got State
		got = 0
		want = (FOLLOWER + FOLLOWER + IDLE)
		for _, raft := range rafts {
			got += raft.getState()
		}

		// [FOLLOWER, FOLLOWER, IDLE]
		assert.Equal(t, want, got)
	}

	// 2. set rafts[2] to LEARNER
	{
		c, cleanup := MockGetClient(t, names[learner])
		defer cleanup()

		method := model.RPCHASetLearner
		req := model.NewHARPCRequest()
		rsp := model.NewHARPCResponse(model.OK)
		err := c.Call(method, req, rsp)
		assert.Nil(t, err)

		want := model.OK
		got := rsp.RetCode
		assert.Equal(t, want, got)
	}

	// 3. check
	{
		MockWaitLeaderEggs(rafts, 1)

		var want, got State
		got = 0
		want = (LEADER + FOLLOWER + LEARNER)
		for _, raft := range rafts {
			got += raft.getState()
		}
		// [LEADER, FOLLOWER, LEARNER]
		assert.Equal(t, want, got)
	}

	// 4. enable ha for rafts[2]
	{
		c, cleanup := MockGetClient(t, names[learner])
		defer cleanup()

		method := model.RPCHAEnable
		req := model.NewHARPCRequest()
		rsp := model.NewHARPCResponse(model.OK)
		err := c.Call(method, req, rsp)
		assert.Nil(t, err)

		want := model.OK
		got := rsp.RetCode
		assert.Equal(t, want, got)
	}

	// 5. check
	{
		MockWaitLeaderEggs(rafts, 1)

		var want, got State
		got = 0
		want = (LEADER + FOLLOWER + FOLLOWER)
		for _, raft := range rafts {
			got += raft.getState()
		}
		// [LEADER, FOLLOWER, FOLLOWER]
		assert.Equal(t, want, got)
	}
}

// TEST EFFECTS:
// test TryToLeader command from the client
//
// TEST PROCESSES:
// 1. Start rpc server
// 2. send command to rpc server
// 3. check the response
func TestRaftRPCHATryToLeader(t *testing.T) {
	var whoisleader, whoisleadernow int
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	port := common.RandomPort(8000, 9000)
	names, rafts, scleanup := MockRafts(log, port, 3, -1)
	defer scleanup()

	// 1. Start 3 rafts state as FOLLOWER
	{
		for _, raft := range rafts {
			raft.Start()
		}

		var want, got State
		got = 0
		want = (FOLLOWER + FOLLOWER + FOLLOWER)
		for _, raft := range rafts {
			got += raft.getState()
		}

		// [FOLLOWER, FOLLOWER, FOLLOWER]
		assert.Equal(t, want, got)
	}

	// 2. check
	{
		MockWaitLeaderEggs(rafts, 1)

		var want, got State
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

	// 3. try to leader
	{
		for i := range names {
			if i == whoisleader {
				continue
			}
			c, cleanup := MockGetClient(t, names[i])
			defer cleanup()

			method := model.RPCHATryToLeader
			req := model.NewHARPCRequest()
			rsp := model.NewHARPCResponse(model.OK)
			err := c.Call(method, req, rsp)
			assert.Nil(t, err)

			want := model.OK
			got := rsp.RetCode
			assert.Equal(t, want, got)
			whoisleadernow = i
			break
		}
	}

	// 4. check
	{
		MockWaitLeaderEggs(rafts, 1)
		assert.NotEqual(t, whoisleader, whoisleadernow)

		var want, got State
		got = 0
		want = (LEADER + FOLLOWER + FOLLOWER)
		for i, raft := range rafts {
			got += raft.getState()
			if raft.getState() == LEADER {
				assert.Equal(t, i, whoisleadernow)
			}
		}
		// [LEADER, FOLLOWER, FOLLOWER]
		assert.Equal(t, want, got)
	}
}

// TEST EFFECTS:
// test HATryToLeader RPC failed call from the client
//
// TEST PROCESSES:
// 1. Start rpc server
// 2. send command to rpc server
// 3. check the response
func TestRaftRPCHATryToLeaderFail_GTID(t *testing.T) {
	var whoisleader int

	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	port := common.RandomPort(8000, 9000)
	names, rafts, cleanup := MockRafts(log, port, 3, -1)
	defer cleanup()

	GTIDBIDX := 1
	GTIDCIDX := 2

	// 1. set rafts GTID
	//    1.0 rafts[0]  with MockGTIDB{Master_Log_File = "", Read_Master_Log_Pos = 0}
	//    1.1 rafts[1]  with MockGTIDB{Master_Log_File = "mysql-bin.000001", Read_Master_Log_Pos = 123}
	//    1.2 rafts[2]  with MockGTIDC{Master_Log_File = "mysql-bin.000001", Read_Master_Log_Pos = 124}
	{
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
		MockWaitLeaderEggs(rafts, 0)
		MockWaitLeaderEggs(rafts, 0)
		MockWaitLeaderEggs(rafts, 0)

		var got State
		whoisleader = 0
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

	// 4. try rafts[1] to leader
	{
		c, cleanup := MockGetClient(t, names[1])
		defer cleanup()

		method := model.RPCHATryToLeader
		req := model.NewHARPCRequest()
		rsp := model.NewHARPCResponse(model.OK)
		err := c.Call(method, req, rsp)
		assert.Nil(t, err)

		want := model.OK
		got := rsp.RetCode
		assert.Equal(t, want, got)
	}

	// 4.1 wait rafts[2] elected as leader again
	{
		var got, g State
		var foundLeader bool
		whoisleader = 0
		want := (LEADER + FOLLOWER + FOLLOWER)
		for !foundLeader {
			g = 0
			MockWaitLeaderEggs(rafts, 0)
			for i, raft := range rafts {
				g += raft.getState()
				if raft.getState() == LEADER {
					if i == GTIDCIDX {
						whoisleader = i
						foundLeader = true
					}
				}
			}
			got = g
		}
		// [LEADER, FOLLOWER, FOLLOWER]
		assert.Equal(t, want, got)

		// leader should be rafts[GTIDCIDX]
		assert.Equal(t, whoisleader, GTIDCIDX)
	}
}

// TEST EFFECTS:
// test TryToLeader command from the client
//
// TEST PROCESSES:
// 1. Start rpc server
// 2. send command to rpc server
// 3. check the response
func TestRaftRPCHATryToLeaderFail_MySQLUnpromotble(t *testing.T) {
	var whoisleader int

	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	port := common.RandomPort(8000, 9000)
	names, rafts, cleanup := MockRafts(log, port, 3, -1)
	defer cleanup()

	GTIDERRIDX := 0
	GTIDBIDX := 1
	GTIDCIDX := 2

	// 1. set rafts GTID
	//    1.0 rafts[0]  with Ping error(mocks MySQL down)
	//    1.1 rafts[1]  with MockGTIDB{Master_Log_File = "mysql-bin.000001", Read_Master_Log_Pos = 123}
	//    1.2 rafts[2]  with MockGTIDC{Master_Log_File = "mysql-bin.000001", Read_Master_Log_Pos = 124}
	{
		rafts[GTIDERRIDX].mysql.SetMysqlHandler(mysql.NewMockGTIDPingError())
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
		// [LEADER, FOLLOWER, FOLLOWER]
		assert.Equal(t, want, got)

		// leader should be rafts[GTIDCIDX]
		assert.Equal(t, whoisleader, GTIDCIDX)
	}

	// 4. try rafts[2](already leader) to leader
	{
		c, cleanup := MockGetClient(t, names[2])
		defer cleanup()

		method := model.RPCHATryToLeader
		req := model.NewHARPCRequest()
		rsp := model.NewHARPCResponse(model.OK)
		err := c.Call(method, req, rsp)
		assert.Nil(t, err)

		want := model.OK
		got := rsp.RetCode
		assert.Equal(t, want, got)
	}

	// 4. try rafts[0] to leader
	{
		c, cleanup := MockGetClient(t, names[0])
		defer cleanup()

		method := model.RPCHATryToLeader
		req := model.NewHARPCRequest()
		rsp := model.NewHARPCResponse(model.OK)
		err := c.Call(method, req, rsp)
		assert.Nil(t, err)

		want := model.RPCError_MySQLUnpromotable
		got := rsp.RetCode
		assert.Equal(t, want, got)
	}

	// 4.1 wait rafts[2] elected as leader again
	{
		MockWaitLeaderEggs(rafts, 0)
		MockWaitLeaderEggs(rafts, 0)
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
		// [LEADER, FOLLOWER, FOLLOWER]
		assert.Equal(t, want, got)

		// leader should be rafts[GTIDCIDX]
		assert.Equal(t, whoisleader, GTIDCIDX)
	}
}

func TestRaftSuperIDLEEnableHA(t *testing.T) {
	var testName = "TestRaftSuperIDLEEnableHA"
	var want, got State
	var whoisleader int
	var leader *Raft
	var idler *Raft

	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	port := common.RandomPort(8000, 9000)
	names, rafts, cleanup := MockRafts(log, port, 3, -1)
	defer cleanup()

	// 1.  Start 3 rafts.
	{
		for i, raft := range rafts {
			if i == 2 {
				raft.conf.SuperIDLE = true
				idler = raft
			}
			raft.Start()
		}
	}

	// 2.  wait leader election.
	{
		MockWaitLeaderEggs(rafts, 1)
		whoisleader = 0
		got = 0
		want = (LEADER + FOLLOWER + IDLE)
		for i, raft := range rafts {
			got += raft.getState()
			if raft.getState() == LEADER {
				whoisleader = i
			}
		}
		// [LEADER, FOLLOWER, IDLE]
		assert.Equal(t, want, got)
	}
	idlerLeader1 := idler.getLeader()

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

	// 5.  Wait new leader.
	{
		MockWaitLeaderEggs(rafts, 1)
	}

	{
		c, cleanup := MockGetClient(t, names[2])
		defer cleanup()

		method := model.RPCHAEnable
		req := model.NewHARPCRequest()
		rsp := model.NewHARPCResponse(model.OK)
		err := c.Call(method, req, rsp)
		assert.Nil(t, err)

		want := model.OK
		got := rsp.RetCode
		assert.Equal(t, want, got)
	}

	idlerLeader2 := idler.getLeader()
	assert.NotEqual(t, idlerLeader1, idlerLeader2)
}
