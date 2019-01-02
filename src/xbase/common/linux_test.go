/*
 * Xenon
 *
 * Copyright 2018 The Xenon Authors.
 * Code is licensed under the GPLv3.
 *
 */

package common

import (
	"testing"
	"xbase/xlog"

	"github.com/stretchr/testify/assert"
)

// TEST EFFECTS:
// test linux api
//
// TEST PROCESSES:
// 1. run ls -l
// 2. run sleep 2
func TestRunCommandWithTimeout(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	timeout := 5 * 1000
	cmds := "ls"
	args := []string{
		"-l",
	}
	_, err := runCommandWithTimeout(log, timeout, cmds, args...)
	assert.Nil(t, err)

	timeout = 1 * 1000
	cmds = "sleep"
	args = []string{
		"2",
	}
	_, err = runCommandWithTimeout(log, timeout, cmds, args...)
	assert.NotNil(t, err)
}

func TestRunCommand(t *testing.T) {
	cmds := "ls"
	args := []string{
		"-l",
	}
	_, err := RunCommand(cmds, args...)
	assert.Nil(t, err)
}

func TestCommand(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	cmds := "bash"
	args := []string{"-c", "sleep"}
	cmd := NewLinuxCommand(log)
	err := cmd.Run(cmds, args)
	assert.Nil(t, err)
	err = cmd.Scan("sleep: missing operand", 1)
	assert.Nil(t, err)

	args = []string{"-c", "sleep 30"}
	err = cmd.Run(cmds, args)
	assert.Nil(t, err)
	cmd.Kill()
	err = cmd.Scan("sleep: missing operand", 0)
	assert.Nil(t, err)

	args = []string{"-c", "ls -l"}
	err = cmd.Run(cmds, args)
	assert.Nil(t, err)
	err = cmd.Scan("sleep: missing operand", 0)
	assert.Nil(t, err)

	args = []string{"-c", "ls -l"}
	_, err = cmd.RunCommand(cmds, args)
	assert.Nil(t, err)

	args = []string{"-c", "ls -l"}
	_, err = cmd.RunCommandWithTimeout(100, cmds, args)
	assert.Nil(t, err)

	args = []string{"-c", "sleep 1000"}
	_, err = cmd.RunCommandWithTimeout(1, cmds, args)
	assert.NotNil(t, err)
}
