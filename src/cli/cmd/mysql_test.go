/*
 * Xenon
 *
 * Copyright 2018 The Xenon Authors.
 * Code is licensed under the GPLv3.
 *
 */

package cmd

import (
	"raft"
	"server"
	"testing"
	"xbase/common"
	"xbase/xlog"

	"github.com/stretchr/testify/assert"
)

func TestCLIMysqlCommand(t *testing.T) {
	var leader string

	err := createConfig()
	ErrorOK(err)
	defer removeConfig()

	// get leader
	{
		log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
		port := common.RandomPort(6000, 10000)
		servers, scleanup := server.MockServers(log, port, 2)
		defer scleanup()

		server.MockWaitLeaderEggs(servers, 1)
		for _, server := range servers {
			if server.GetState() == raft.LEADER {
				leader = server.Address()
				break
			}
		}
	}

	// setting xenon is leader
	{
		conf, err := GetConfig()
		ErrorOK(err)
		conf.Server.Endpoint = leader
		err = SaveConfig(conf)
		ErrorOK(err)
	}

	// create normal user.
	{
		cmd := NewMysqlCommand()
		_, err := executeCommand(cmd, "createuser", "userxx", "192.168.0.%", "passwdxx", "NO")
		assert.Nil(t, err)
	}

	// create super user.
	{
		cmd := NewMysqlCommand()
		_, err := executeCommand(cmd, "createsuperuser", "192.168.0.%", "userxx", "passwdxx", "NO")
		assert.Nil(t, err)
	}

	// change password
	{
		cmd := NewMysqlCommand()
		_, err := executeCommand(cmd, "changepassword", "userxx", "192.168.0.%", "passwdxx")
		assert.Nil(t, err)
	}

	// get mysql user list
	{
		cmd := NewMysqlCommand()
		_, err := executeCommand(cmd, "getuser")
		assert.Nil(t, err)
	}

	// drop normal user.
	{
		cmd := NewMysqlCommand()
		_, err := executeCommand(cmd, "dropuser", "userxx", "192.168.0.%")
		assert.Nil(t, err)
	}

	// set global sysvar
	{
		cmd := NewMysqlCommand()
		_, err := executeCommand(cmd, "sysvar", "SET GLOBAL GTID_MODE='ON'")
		assert.Nil(t, err)
	}

	// kill mysql
	{
		cmd := NewMysqlCommand()
		_, err := executeCommand(cmd, "kill")
		assert.Nil(t, err)
	}

	// status
	{
		cmd := NewMysqlCommand()
		_, err := executeCommand(cmd, "status")
		assert.Nil(t, err)
	}

	// create user with privileges
	{
		cmd := NewMysqlCommand()
		_, err := executeCommand(cmd, "createuserwithgrants", "--user", "xx", "--passwd", "xx", "--database", "db1", "--host", "192.168.0.%", "--privs", "SELECT,DROP")
		assert.Nil(t, err)
	}
}
