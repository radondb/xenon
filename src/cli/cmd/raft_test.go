/*
 * Xenon
 *
 * Copyright 2018 The Xenon Authors.
 * Code is licensed under the GPLv3.
 *
 */

package cmd

import (
	"errors"
	"fmt"
	"raft"
	"server"
	"testing"
	"time"
	"xbase/common"
	"xbase/xlog"

	"github.com/stretchr/testify/assert"
)

func TestCLIRaftCommand(t *testing.T) {
	var leader, newleader string

	err := createConfig()
	ErrorOK(err)
	defer removeConfig()

	log := xlog.NewStdLog(xlog.Level(xlog.DEBUG))
	port := common.RandomPort(8000, 9000)
	servers, scleanup := server.MockServers(log, port, 3)
	defer scleanup()

	// get leader
	{
		server.MockWaitLeaderEggs(servers, 1)
		for _, server := range servers {
			if server.GetState() == raft.LEADER {
				leader = server.Address()
				break
			}
		}
	}

	conf, err := GetConfig()

	// 1. test disable raft to leader
	{
		// setting xenon is leader
		{
			ErrorOK(err)
			conf.Server.Endpoint = leader
			err = SaveConfig(conf)
			ErrorOK(err)
		}

		// disable raft
		{
			cmd := NewRaftCommand()
			_, err := executeCommand(cmd, "disable")
			assert.Nil(t, err)
		}

		// check the new leader
		{
			server.MockWaitLeaderEggs(servers, 1)
			for _, server := range servers {
				if server.GetState() == raft.LEADER {
					newleader = server.Address()
					break
				}
			}

			if leader == newleader {
				ErrorOK(errors.New("leader==newleader.error"))
			}
		}

		// enable raft
		{
			cmd := NewRaftCommand()
			_, err := executeCommand(cmd, "enable")
			assert.Nil(t, err)
		}

		// trytoleader
		{
			cmd := NewRaftCommand()
			_, err := executeCommand(cmd, "trytoleader")
			assert.Nil(t, err)
		}

		// wait cli done, avoid enable cmd changes STOPPED to FOLLOWER
		time.Sleep(time.Duration(conf.Raft.ElectionTimeout))
	}

	// 2. test add/remove ndoes to local
	{

		// add
		{
			ip, err := common.GetLocalIP()
			assert.Nil(t, err)
			arg := fmt.Sprintf("%s:7001,%s:7002", ip, ip)
			cmd := NewRaftCommand()
			_, err = executeCommand(cmd, "add", arg)
			assert.Nil(t, err)
		}

		// ls
		{
			cmd := NewRaftCommand()
			_, err := executeCommand(cmd, "nodes")
			assert.Nil(t, err)
		}

		// remove
		{
			ip, err := common.GetLocalIP()
			assert.Nil(t, err)
			arg := fmt.Sprintf("%s:7001,%s:7002", ip, ip)
			cmd := NewRaftCommand()
			_, err = executeCommand(cmd, "remove", arg)
			assert.Nil(t, err)
		}

		// ls
		{
			cmd := NewRaftCommand()
			_, err := executeCommand(cmd, "nodes")
			assert.Nil(t, err)
		}

		// status
		{
			cmd := NewRaftCommand()
			_, err := executeCommand(cmd, "status")
			assert.Nil(t, err)
		}

		// purge binlog
		{
			cmd := NewRaftCommand()
			_, err := executeCommand(cmd, "disablepurgebinlog")
			assert.Nil(t, err)
		}

		// purge binlog enable
		{
			cmd := NewRaftCommand()
			_, err := executeCommand(cmd, "enablepurgebinlog")
			assert.Nil(t, err)
		}

		// disable check semi-sync
		{
			cmd := NewRaftCommand()
			_, err := executeCommand(cmd, "disablechecksemisync")
			assert.Nil(t, err)
		}

		// enable check semi-sync
		{
			cmd := NewRaftCommand()
			_, err := executeCommand(cmd, "enablechecksemisync")
			assert.Nil(t, err)
		}
	}
}
