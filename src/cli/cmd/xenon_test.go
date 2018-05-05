/*
 * Xenon
 *
 * Copyright 2018 The Xenon Authors.
 * Code is licensed under the GPLv3.
 *
 */

package cmd

import (
	"server"
	"testing"
	"xbase/common"
	"xbase/xlog"

	"github.com/stretchr/testify/assert"
)

func TestCLIXenonCommand(t *testing.T) {

	err := createConfig()
	ErrorOK(err)
	defer removeConfig()

	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	port := common.RandomPort(8000, 9000)
	servers, cleanup := server.MockServers(log, port, 1)
	defer cleanup()

	// setting xenon is leader
	{
		conf, err := GetConfig()
		ErrorOK(err)
		conf.Server.Endpoint = servers[0].Address()
		err = SaveConfig(conf)
		ErrorOK(err)
	}

	// ping
	{
		cmd := NewXenonCommand()
		_, err := executeCommand(cmd, "ping")
		assert.Nil(t, err)
	}
}
