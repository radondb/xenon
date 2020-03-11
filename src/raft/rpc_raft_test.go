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

func TestRaftRPCStatus(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	port := common.RandomPort(8000, 9000)
	names, rafts, scleanup := MockRafts(log, port, 3, -1)
	defer scleanup()
	var whoisleader int

	{
		for _, raft := range rafts {
			raft.Start()
		}

		MockWaitLeaderEggs(rafts, 1)
		for i, raft := range rafts {
			if raft.getState() == LEADER {
				whoisleader = i
				break
			}
		}
	}

	{
		MockWaitLeaderEggs(rafts, 1)
		c, cleanup := MockGetClient(t, names[whoisleader])
		defer cleanup()

		method := model.RPCRaftStatus
		req := model.NewRaftStatusRPCRequest()
		rsp := model.NewRaftStatusRPCResponse(model.OK)
		err := c.Call(method, req, rsp)
		assert.Nil(t, err)

		want := 1
		got := int(rsp.Stats.LeaderPromotes)
		assert.Equal(t, want, got)
	}
}

func TestRaftRPCs(t *testing.T) {
	mockHost := ":6666"
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	port := common.RandomPort(8000, 9000)
	names, rafts, scleanup := MockRafts(log, port, 1, -1)
	defer scleanup()

	// start
	{
		for _, raft := range rafts {
			raft.Start()
		}
		rafts[0].AddPeer(mockHost)
	}

	// test: heartbeat with large ViewID to CANDIDATE
	{
		MockStateTransition(rafts[0], CANDIDATE)
		c, cleanup := MockGetClient(t, names[0])
		defer cleanup()

		method := model.RPCRaftHeartbeat
		req := model.NewRaftRPCRequest()
		req.Raft.From = mockHost
		req.Raft.ViewID = 1000

		rsp := model.NewRaftRPCResponse(model.OK)
		err := c.Call(method, req, rsp)
		assert.Nil(t, err)

		want := model.OK
		got := rsp.RetCode
		assert.Equal(t, want, got)
	}

	// test: requestvote with small ViewID to LEADER
	{
		MockStateTransition(rafts[0], LEADER)
		c, cleanup := MockGetClient(t, names[0])
		defer cleanup()

		method := model.RPCRaftRequestVote
		req := model.NewRaftRPCRequest()
		req.Raft.From = mockHost
		req.Raft.ViewID = 0

		rsp := model.NewRaftRPCResponse(model.OK)
		err := c.Call(method, req, rsp)
		assert.Nil(t, err)

		want := model.ErrorInvalidViewID
		got := rsp.RetCode
		assert.Equal(t, want, got)
	}
}

func TestRaftRPCPurgeBinlog(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	port := common.RandomPort(8000, 9000)
	names, rafts, scleanup := MockRafts(log, port, 3, -1)
	defer scleanup()
	whoisleader := 2

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

	// wait leader eggs
	{
		MockWaitLeaderEggs(rafts, 1)
	}
	// check(default is enable)
	{
		MockWaitLeaderEggs(rafts, 0)
		MockWaitLeaderEggs(rafts, 0)
		purged := rafts[whoisleader].stats.LeaderPurgeBinlogs
		assert.NotZero(t, purged)
	}

	// disable purge binlog
	{
		c, cleanup := MockGetClient(t, names[whoisleader])
		defer cleanup()

		method := model.RPCRaftDisablePurgeBinlog
		req := model.NewRaftStatusRPCRequest()
		rsp := model.NewRaftStatusRPCResponse(model.OK)
		err := c.Call(method, req, rsp)
		assert.Nil(t, err)

		// check
		want := rafts[whoisleader].stats.LeaderPurgeBinlogs
		MockWaitLeaderEggs(rafts, 0)
		MockWaitLeaderEggs(rafts, 0)
		got := rafts[whoisleader].stats.LeaderPurgeBinlogs
		assert.Equal(t, want, got)
	}

	// enable purge binlog
	{
		MockWaitLeaderEggs(rafts, 1)
		c, cleanup := MockGetClient(t, names[whoisleader])
		defer cleanup()

		method := model.RPCRaftEnablePurgeBinlog
		req := model.NewRaftStatusRPCRequest()
		rsp := model.NewRaftStatusRPCResponse(model.OK)
		err := c.Call(method, req, rsp)
		assert.Nil(t, err)

		// check
		want := rafts[whoisleader].stats.LeaderPurgeBinlogs
		MockWaitLeaderEggs(rafts, 0)
		MockWaitLeaderEggs(rafts, 0)
		got := rafts[whoisleader].stats.LeaderPurgeBinlogs
		assert.NotEqual(t, want, got)
	}
}

func TestRaftRPCCheckSemiSync(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	port := common.RandomPort(8000, 9000)
	names, rafts, scleanup := MockRafts(log, port, 3, -1)
	defer scleanup()
	whoisleader := 2

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

	// wait leader eggs
	{
		MockWaitLeaderEggs(rafts, 1)
	}
	// check(default is enable)
	{
		MockWaitLeaderEggs(rafts, 0)
		MockWaitLeaderEggs(rafts, 0)
		check := rafts[whoisleader].skipCheckSemiSync
		assert.Equal(t, false, check)
	}

	// disable check semi-sync
	{
		c, cleanup := MockGetClient(t, names[whoisleader])
		defer cleanup()

		method := model.RPCRaftDisableCheckSemiSync
		req := model.NewRaftStatusRPCRequest()
		rsp := model.NewRaftStatusRPCResponse(model.OK)
		err := c.Call(method, req, rsp)
		assert.Nil(t, err)

		// check
		MockWaitLeaderEggs(rafts, 0)
		MockWaitLeaderEggs(rafts, 0)
		got := rafts[whoisleader].skipCheckSemiSync
		assert.Equal(t, true, got)
	}

	// enable check semi-sync
	{
		MockWaitLeaderEggs(rafts, 1)
		c, cleanup := MockGetClient(t, names[whoisleader])
		defer cleanup()

		method := model.RPCRaftEnableCheckSemiSync
		req := model.NewRaftStatusRPCRequest()
		rsp := model.NewRaftStatusRPCResponse(model.OK)
		err := c.Call(method, req, rsp)
		assert.Nil(t, err)

		// check
		MockWaitLeaderEggs(rafts, 0)
		MockWaitLeaderEggs(rafts, 0)
		got := rafts[whoisleader].skipCheckSemiSync
		assert.Equal(t, false, got)
	}
}
