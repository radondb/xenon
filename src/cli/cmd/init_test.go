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
	"testing"
	"xbase/common"

	"github.com/stretchr/testify/assert"
)

func TestCLIInitCommand(t *testing.T) {
	err := createConfig()
	ErrorOK(err)
	defer removeConfig()

	ip, err := common.GetLocalIP()
	assert.Nil(t, err)

	cmd := NewInitCommand()
	_, err = executeCommand(cmd, "init", "--address", ip, "--port", "8080", "--repluser", "repl", "--replpwd", "repl", "--vip", ip)
	assert.Nil(t, err)

	conf, err := GetConfig()
	assert.Nil(t, err)

	want := fmt.Sprintf("%s:8080", ip)
	got := conf.Server.Endpoint
	assert.Equal(t, want, got)
}
