/*
 * Xenon
 *
 * Copyright 2018 The Xenon Authors.
 * Code is licensed under the GPLv3.
 *
 */

package server

import (
	"model"
	"raft"
	"testing"
	"xbase/common"
	"xbase/xlog"

	"github.com/stretchr/testify/assert"
)

func TestServerRPCUser(t *testing.T) {
	var leader, follower string
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	port := common.RandomPort(8000, 9000)
	servers, cleanup := MockServers(log, port, 2)
	defer cleanup()

	// leader&&follower
	{
		MockWaitLeaderEggs(servers, 1)
		for i, server := range servers {
			if server.raft.GetState() == raft.LEADER {
				leader = servers[i].Address()
			} else {
				follower = servers[i].Address()
			}
		}
	}

	// send to follower
	{
		c, cleanup := MockGetClient(t, follower)
		defer cleanup()

		req := model.NewMysqlUserRPCRequest()
		rsp := model.NewMysqlUserRPCResponse(model.OK)
		method := model.RPCMysqlCreateNormalUser
		if err := c.Call(method, req, rsp); err != nil {
			assert.Nil(t, err)
		}

		want := "nonleader.can.not.createuser"
		got := rsp.RetCode
		assert.Equal(t, want, got)
	}

	// send to leader
	{
		c, cleanup := MockGetClient(t, leader)
		defer cleanup()

		req := model.NewMysqlUserRPCRequest()
		rsp := model.NewMysqlUserRPCResponse(model.OK)
		method := model.RPCMysqlCreateNormalUser
		if err := c.Call(method, req, rsp); err != nil {
			assert.Nil(t, err)
		}

		want := model.OK
		got := rsp.RetCode
		assert.Equal(t, want, got)
	}

	// Create super user.
	{
		c, cleanup := MockGetClient(t, leader)
		defer cleanup()

		req := model.NewMysqlUserRPCRequest()
		rsp := model.NewMysqlUserRPCResponse(model.OK)
		method := model.RPCMysqlCreateSuperUser
		if err := c.Call(method, req, rsp); err != nil {
			assert.Nil(t, err)
		}

		want := model.OK
		got := rsp.RetCode
		assert.Equal(t, want, got)
	}

	// Get user.
	{
		c, cleanup := MockGetClient(t, leader)
		defer cleanup()

		req := model.NewMysqlUserRPCRequest()
		rsp := model.NewMysqlUserRPCResponse(model.OK)
		method := model.RPCMysqlGetUser
		if err := c.Call(method, req, rsp); err != nil {
			assert.Nil(t, err)
		}

		want := model.OK
		got := rsp.RetCode
		assert.Equal(t, want, got)

		want1 := []model.MysqlUser{
			{User: "user1",
				Host:      "localhost",
				SuperPriv: "N"},
			{User: "root",
				Host:      "localhost",
				SuperPriv: "Y"},
		}
		got1 := rsp.Users
		assert.Equal(t, want1, got1)
	}

	// Drop user.
	{
		c, cleanup := MockGetClient(t, leader)
		defer cleanup()

		req := model.NewMysqlUserRPCRequest()
		rsp := model.NewMysqlUserRPCResponse(model.OK)
		method := model.RPCMysqlDropUser
		if err := c.Call(method, req, rsp); err != nil {
			assert.Nil(t, err)
		}

		want := model.OK
		got := rsp.RetCode
		assert.Equal(t, want, got)
	}

	// Create user with privileges.
	{
		c, cleanup := MockGetClient(t, leader)
		defer cleanup()

		req := model.NewMysqlUserRPCRequest()
		rsp := model.NewMysqlUserRPCResponse(model.OK)
		method := model.RPCMysqlCreateUserWithPrivileges
		if err := c.Call(method, req, rsp); err != nil {
			assert.Nil(t, err)
		}

		want := model.OK
		got := rsp.RetCode
		assert.Equal(t, want, got)
	}

	// Change password.
	{
		c, cleanup := MockGetClient(t, leader)
		defer cleanup()

		req := model.NewMysqlUserRPCRequest()
		rsp := model.NewMysqlUserRPCResponse(model.OK)
		method := model.RPCMysqlChangePassword
		if err := c.Call(method, req, rsp); err != nil {
			assert.Nil(t, err)
		}

		want := model.OK
		got := rsp.RetCode
		assert.Equal(t, want, got)
	}

	// send to leader
	{
		c, cleanup := MockGetClient(t, leader)
		defer cleanup()

		req := model.NewMysqlUserRPCRequest()
		rsp := model.NewMysqlUserRPCResponse(model.OK)
		method := model.RPCMysqlCreateNormalUser
		if err := c.Call(method, req, rsp); err != nil {
			assert.Nil(t, err)
		}

		want := model.OK
		got := rsp.RetCode
		assert.Equal(t, want, got)
	}
}
