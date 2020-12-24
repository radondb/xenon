/*
 * Xenon
 *
 * Copyright 2018 The Xenon Authors.
 * Code is licensed under the GPLv3.
 *
 */

package mysql

import (
	"model"
	"testing"
	"xbase/common"
	"xbase/xlog"

	"github.com/stretchr/testify/assert"
)

func TestMysqlRPCStatus(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	port := common.RandomPort(8000, 9000)
	id, _, cleanup := MockMysql(log, port, NewMockGTIDB())
	defer cleanup()

	// client
	{
		method := model.RPCMysqlStatus
		req := model.NewMysqlStatusRPCRequest()
		rsp := model.NewMysqlStatusRPCResponse(model.OK)
		c, cleanup := MockGetClient(t, id)
		defer cleanup()
		err := c.Call(method, req, rsp)
		assert.Nil(t, err)

		GTID := model.GTID{
			Master_Log_File:     "mysql-bin.000001",
			Read_Master_Log_Pos: 123,
			Executed_GTID_Set:   "c78e798a-cccc-cccc-cccc-525433e8e796:1-2",
			Slave_IO_Running:    true,
			Slave_SQL_Running:   true,
		}
		want := model.NewMysqlStatusRPCResponse(model.OK)
		want.GTID = GTID
		want.Status = string(model.MysqlDead)
		want.Stats = &model.MysqlStats{}

		got := rsp
		assert.Equal(t, want, got)
	}
}

func TestMysqlRPCSetState(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	port := common.RandomPort(8000, 9000)
	id, _, cleanup := MockMysql(log, port, NewMockGTIDB())
	defer cleanup()

	// MysqlDead
	{
		method := model.RPCMysqlSetState
		req := model.NewMysqlSetStateRPCRequest()
		req.State = model.MysqlDead
		rsp := model.NewMysqlSetStateRPCResponse(model.OK)
		c, cleanup := MockGetClient(t, id)
		defer cleanup()
		err := c.Call(method, req, rsp)
		assert.Nil(t, err)

		want := model.NewMysqlSetStateRPCResponse(model.OK)
		got := rsp
		assert.Equal(t, want, got)
	}

	// MysqlAlive
	{
		method := model.RPCMysqlSetState
		req := model.NewMysqlSetStateRPCRequest()
		req.State = model.MysqlAlive
		rsp := model.NewMysqlSetStateRPCResponse(model.OK)
		c, cleanup := MockGetClient(t, id)
		defer cleanup()
		err := c.Call(method, req, rsp)
		assert.Nil(t, err)

		want := model.NewMysqlSetStateRPCResponse(model.OK)
		got := rsp
		assert.Equal(t, want, got)
	}
}

func TestMysqlRPCSetSysVar(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	port := common.RandomPort(8000, 9000)
	id, _, cleanup := MockMysql(log, port, new(Mysql57))
	defer cleanup()

	// client
	{
		method := model.RPCMysqlSetGlobalSysVar
		req := model.NewMysqlVarRPCRequest()
		rsp := model.NewMysqlVarRPCResponse(model.OK)
		c, cleanup := MockGetClient(t, id)
		defer cleanup()
		err := c.Call(method, req, rsp)
		assert.Nil(t, err)
		want := "[].must.be.startwith:SET GLOBAL"
		got := rsp.RetCode
		assert.Equal(t, want, got)
	}
}

func TestMysqlRPCResetMaster(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	port := common.RandomPort(8000, 9000)
	id, _, cleanup := MockMysql(log, port, new(Mysql57))
	defer cleanup()

	// client
	{
		method := model.RPCMysqlResetMaster
		req := model.NewMysqlRPCRequest()
		rsp := model.NewMysqlRPCResponse(model.OK)
		c, cleanup := MockGetClient(t, id)
		defer cleanup()
		err := c.Call(method, req, rsp)
		assert.Nil(t, err)
	}
}

func TestMysqlRPCResetSlaveAll(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	port := common.RandomPort(8000, 9000)
	id, _, cleanup := MockMysql(log, port, new(Mysql57))
	defer cleanup()

	// client
	{
		method := model.RPCMysqlResetSlaveAll
		req := model.NewMysqlRPCRequest()
		rsp := model.NewMysqlRPCResponse(model.OK)
		c, cleanup := MockGetClient(t, id)
		defer cleanup()
		err := c.Call(method, req, rsp)
		assert.Nil(t, err)
	}
}

func TestMysqlRPCSlaves(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	port := common.RandomPort(8000, 9000)
	id, _, cleanup := MockMysql(log, port, new(Mysql57))
	defer cleanup()

	// start
	{
		method := model.RPCMysqlStartSlave
		req := model.NewMysqlRPCRequest()
		rsp := model.NewMysqlRPCResponse(model.OK)
		c, cleanup := MockGetClient(t, id)
		defer cleanup()
		err := c.Call(method, req, rsp)
		assert.Nil(t, err)
	}

	// stop
	{
		method := model.RPCMysqlStopSlave
		req := model.NewMysqlRPCRequest()
		rsp := model.NewMysqlRPCResponse(model.OK)
		c, cleanup := MockGetClient(t, id)
		defer cleanup()
		err := c.Call(method, req, rsp)
		assert.Nil(t, err)
	}
}

func TestMysqlRPCIsWorking(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	port := common.RandomPort(8000, 9000)
	id, _, cleanup := MockMysql(log, port, new(Mysql57))
	defer cleanup()

	{
		method := model.RPCMysqlIsWorking
		req := model.NewMysqlRPCRequest()
		rsp := model.NewMysqlRPCResponse(model.OK)
		c, cleanup := MockGetClient(t, id)
		defer cleanup()
		err := c.Call(method, req, rsp)
		assert.Nil(t, err)
		want := model.ErrorMySQLDown
		got := rsp.RetCode
		assert.Equal(t, want, got)
	}
}
