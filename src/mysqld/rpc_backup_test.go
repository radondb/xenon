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

func TestBackupRPCBackupDo(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	port := common.RandomPort(8000, 9000)
	endpoint, mysqld, cleanup := MockMysqld(log, port)
	defer cleanup()

	mysqld.backup.cmd = common.NewMockCommand()

	// do backup
	{
		go func() {
			c, _ := MockGetClient(t, endpoint)
			method := model.RPCBackupDo
			req := model.NewBackupRPCRequest()
			req.SSHHost = "127.0.0.1"
			req.SSHUser = "backup"
			req.SSHPasswd = "backup"
			req.SSHPort = 22

			rsp := model.NewBackupRPCResponse(model.OK)
			err := c.Call(method, req, rsp)
			assert.Nil(t, err)
			want := model.OK
			got := rsp.RetCode
			assert.Equal(t, want, got)
		}()
	}

	// cancel
	{
		req := model.NewBackupRPCRequest()
		rsp := model.NewBackupRPCResponse(model.OK)
		c, _ := MockGetClient(t, endpoint)
		{
			method := model.RPCBackupCancel
			err := c.Call(method, req, rsp)
			assert.Nil(t, err)
			want := model.OK
			got := rsp.RetCode
			assert.Equal(t, want, got)
		}
	}
}

func TestBackupRPCApplyLog(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	port := common.RandomPort(8000, 9000)
	endpoint, mysqld, cleanup := MockMysqld(log, port)
	defer cleanup()
	mysqld.backup.cmd = common.NewMockCommand()
	mysqld.MonitorStop()

	{
		req := model.NewBackupRPCRequest()
		rsp := model.NewBackupRPCResponse(model.OK)
		c, _ := MockGetClient(t, endpoint)

		go func() {
			method := model.RPCBackupApplyLog
			err := c.Call(method, req, rsp)
			assert.Nil(t, err)
			want := model.OK
			got := rsp.RetCode
			assert.Equal(t, want, got)
		}()
	}

	{
		req := model.NewBackupRPCRequest()
		rsp := model.NewBackupRPCResponse(model.OK)
		c, _ := MockGetClient(t, endpoint)

		{
			method := model.RPCBackupCancel
			err := c.Call(method, req, rsp)
			assert.Nil(t, err)
			want := model.OK
			got := rsp.RetCode
			assert.Equal(t, want, got)
		}
	}
}
