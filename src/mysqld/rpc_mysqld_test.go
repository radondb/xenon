/*
 * Xenon
 *
 * Copyright 2018 The Xenon Authors.
 * Code is licensed under the GPLv3.
 *
 */

package mysqld

import (
	"testing"

	"model"
	"xbase/common"
	"xbase/xlog"

	"github.com/stretchr/testify/assert"
)

func TestMysqldRPCMonitor(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	port := common.RandomPort(8000, 9000)
	endpoint, _, cleanup := MockMysqld(log, port)
	defer cleanup()

	req := model.NewMysqldRPCRequest()
	rsp := model.NewMysqldRPCResponse(model.OK)
	c, ccleanup := MockGetClient(t, endpoint)
	defer ccleanup()

	{
		method := model.RPCMysqldStartMonitor
		err := c.Call(method, req, rsp)
		assert.Nil(t, err)
		want := model.OK
		got := rsp.RetCode
		assert.Equal(t, want, got)
	}

	{
		method := model.RPCMysqldStopMonitor
		err := c.Call(method, req, rsp)
		assert.Nil(t, err)
		want := model.OK
		got := rsp.RetCode
		assert.Equal(t, want, got)
	}
}

func TestMysqldRPCShutDownStartAndStatus(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	port := common.RandomPort(8000, 9000)
	endpoint, mysqld, cleanup := MockMysqld(log, port)
	defer cleanup()
	mysqld.backup.cmd = common.NewMockCommand()
	mysqld.MonitorStop()

	// shutdown
	{
		c, ccleanup := MockGetClient(t, endpoint)
		defer ccleanup()

		method := model.RPCMysqldShutDown
		req := model.NewMysqldRPCRequest()
		rsp := model.NewMysqldRPCResponse(model.OK)
		err := c.Call(method, req, rsp)
		assert.Nil(t, err)

		want := model.OK
		got := rsp.RetCode
		assert.Equal(t, want, got)
	}

	// isrunning
	{
		c, ccleanup := MockGetClient(t, endpoint)
		defer ccleanup()

		method := model.RPCMysqldIsRuning
		req := model.NewMysqldRPCRequest()
		rsp := model.NewMysqldRPCResponse(model.OK)
		err := c.Call(method, req, rsp)
		assert.Nil(t, err)

		want := model.OK
		got := rsp.RetCode
		assert.Equal(t, want, got)
	}

	// status
	{
		c, ccleanup := MockGetClient(t, endpoint)
		defer ccleanup()

		method := model.RPCMysqldStatus
		req := model.NewMysqldStatusRPCRequest()
		rsp := model.NewMysqldStatusRPCResponse(model.OK)
		err := c.Call(method, req, rsp)
		assert.Nil(t, err)

		want := &model.MysqldStatusRPCResponse{
			MonitorInfo:  "OFF",
			MysqldInfo:   "UNKNOW",
			BackupInfo:   "NONE",
			BackupStatus: "NONE",
			RetCode:      "OK",
		}
		rsp.BackupStats = nil
		rsp.MysqldStats = nil
		got := rsp
		assert.Equal(t, want, got)
	}

	// start
	{
		req := model.NewMysqldRPCRequest()
		rsp := model.NewMysqldRPCResponse(model.OK)
		c, ccleanup := MockGetClient(t, endpoint)
		defer ccleanup()

		method := model.RPCMysqldStart
		err := c.Call(method, req, rsp)
		assert.Nil(t, err)

		want := model.OK
		got := rsp.RetCode
		assert.Equal(t, want, got)
	}

	// isrunning
	{
		c, ccleanup := MockGetClient(t, endpoint)
		defer ccleanup()

		method := model.RPCMysqldIsRuning
		req := model.NewMysqldRPCRequest()
		rsp := model.NewMysqldRPCResponse(model.OK)
		err := c.Call(method, req, rsp)
		assert.Nil(t, err)

		want := model.OK
		got := rsp.RetCode
		assert.Equal(t, want, got)
	}

	// kill
	{
		c, ccleanup := MockGetClient(t, endpoint)
		defer ccleanup()

		method := model.RPCMysqldKill
		req := model.NewMysqldRPCRequest()
		rsp := model.NewMysqldRPCResponse(model.OK)
		err := c.Call(method, req, rsp)
		assert.Nil(t, err)

		want := model.OK
		got := rsp.RetCode
		assert.Equal(t, want, got)
	}
}
