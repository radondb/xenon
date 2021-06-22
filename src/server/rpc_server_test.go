/*
 * Xenon
 *
 * Copyright 2018 The Xenon Authors.
 * Code is licensed under the GPLv3.
 *
 */

package server

import (
	"testing"

	"model"
	"xbase/common"
	"xbase/xlog"

	"github.com/stretchr/testify/assert"
)

// TEST EFFECTS:
// test a ping command from the client
//
// TEST PROCESSES:
// 1. Start rpc server
// 2. send command to rpc server
// 3. check the response
func TestServerRPCPing(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	port := common.RandomPort(8000, 9000)
	servers, cleanup := MockServers(log, port, 1)
	defer cleanup()
	name := servers[0].Address()

	// rpc call
	{
		req := model.NewServerRPCRequest()
		rsp := model.NewServerRPCResponse(model.OK)
		c, cleanup := MockGetClient(t, name)
		defer cleanup()

		method := model.RRCServerPing
		if err := c.Call(method, req, rsp); err != nil {
			assert.Nil(t, err)
		}

		want := model.OK
		got := rsp.RetCode
		assert.Equal(t, want, got)
	}
}

func TestServerRPCStatus(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	port := common.RandomPort(8000, 9000)
	servers, cleanup := MockServers(log, port, 1)
	defer cleanup()
	name := servers[0].Address()

	// rpc call
	{
		req := model.NewServerRPCRequest()
		rsp := model.NewServerRPCResponse(model.OK)
		c, cleanup := MockGetClient(t, name)
		defer cleanup()

		method := model.RPCServerStatus
		if err := c.Call(method, req, rsp); err != nil {
			assert.Nil(t, err)
		}

		config := &model.ConfigStatus{
			LogLevel:              "INFO",
			BackupDir:             "/u01/backup",
			BackupIOPSLimits:      100000,
			XtrabackupBinDir:      ".",
			MysqldBaseDir:         "/u01/mysql_20160606/",
			MysqldDefaultsFile:    "/etc/my3306.cnf",
			MysqlAdmin:            "root",
			MysqlHost:             "localhost",
			MysqlPort:             3306,
			MysqlReplUser:         "repl",
			MysqlPingTimeout:      1000,
			RaftDataDir:           ".",
			RaftHeartbeatTimeout:  100,
			RaftElectionTimeout:   300,
			RaftRPCRequestTimeout: 1000,
			RaftStartVipCommand:   "nop",
			RaftStopVipCommand:    "nop",
		}
		want := config
		got := rsp.Config
		assert.Equal(t, want, got)
	}
}
