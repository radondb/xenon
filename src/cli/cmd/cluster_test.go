/*
 * Xenon
 *
 * Copyright 2018 The Xenon Authors.
 * Code is licensed under the GPLv3.
 *
 */

package cmd

import (
	"fmt"
	"raft"
	"server"
	"testing"
	"xbase/common"
	"xbase/xlog"

	"github.com/stretchr/testify/assert"
)

func TestCLIClusterCommand(t *testing.T) {
	var leader string
	var follower string

	err := createConfig()
	ErrorOK(err)
	defer removeConfig()

	// get leader
	{
		log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
		port := common.RandomPort(6000, 10000)
		servers, cleanup := server.MockServers(log, port, 2)
		defer cleanup()

		server.MockWaitLeaderEggs(servers, 1)
		for _, server := range servers {
			if server.GetState() == raft.LEADER {
				leader = server.Address()
			} else {
				follower = server.Address()
			}
		}
	}

	// 1. test add node direct to leader
	{
		// setting xenon is leader
		{
			conf, err := GetConfig()
			ErrorOK(err)
			conf.Server.Endpoint = leader
			err = SaveConfig(conf)
			ErrorOK(err)
		}

		// status.
		{
			cmd := NewClusterCommand()
			_, err := executeCommand(cmd, "status")
			assert.Nil(t, err)
		}

		// add node.
		{
			ip, err := common.GetLocalIP()
			cmd := NewClusterCommand()
			_, err = executeCommand(cmd, "add", fmt.Sprintf("%s:6001,%s:6002", ip, ip))
			assert.Nil(t, err)
		}

		// status.
		{
			cmd := NewClusterCommand()
			_, err := executeCommand(cmd, "status")
			assert.Nil(t, err)
		}
	}

	// 2. test add node forward to leader
	{
		// setting xenon is follower
		{
			conf, err := GetConfig()
			ErrorOK(err)
			conf.Server.Endpoint = follower
			err = SaveConfig(conf)
			ErrorOK(err)
		}

		// add nodes.
		{
			ip, err := common.GetLocalIP()
			cmd := NewClusterCommand()
			_, err = executeCommand(cmd, "add", fmt.Sprintf("%s:7001,%s:7002", ip, ip))
			assert.Nil(t, err)
		}

		// status.
		{
			cmd := NewClusterCommand()
			_, err := executeCommand(cmd, "status")
			assert.Nil(t, err)
		}
	}

	// 3. test remove node forward to leader
	{
		// setting xenon is follower
		{
			conf, err := GetConfig()
			ErrorOK(err)
			conf.Server.Endpoint = follower
			err = SaveConfig(conf)
			ErrorOK(err)
		}

		// remove nodes.
		{
			ip, err := common.GetLocalIP()
			cmd := NewClusterCommand()
			_, err = executeCommand(cmd, "remove", fmt.Sprintf("%s:7001,%s:7002", ip, ip))
			assert.Nil(t, err)
		}

		// status.
		{
			cmd := NewClusterCommand()
			_, err := executeCommand(cmd, "status")
			assert.Nil(t, err)
		}

		// json status
		{
			cmd := NewClusterCommand()
			_, err := executeCommand(cmd, "status", "json")
			assert.Nil(t, err)
		}

		// msyql status.
		{
			cmd := NewClusterCommand()
			_, err := executeCommand(cmd, "mysql")
			assert.Nil(t, err)
		}

		// msyql GTID.
		{
			cmd := NewClusterCommand()
			_, err := executeCommand(cmd, "gtid")
			assert.Nil(t, err)
		}

		// raft
		{
			cmd := NewClusterCommand()
			_, err := executeCommand(cmd, "raft")
			assert.Nil(t, err)
		}

		// xenon
		{
			cmd := NewClusterCommand()
			_, err := executeCommand(cmd, "xenon")
			assert.Nil(t, err)
		}

	}
}
