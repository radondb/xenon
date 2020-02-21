/*
 * Xenon
 *
 * Copyright 2018 The Xenon Authors.
 * Code is licensed under the GPLv3.
 *
 */

package mysqld

import (
	"config"
	"fmt"
	"model"
	"testing"
	"time"
	"xbase/common"
	"xbase/xlog"

	"github.com/stretchr/testify/assert"
)

func TestStartMysqld(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	mysqld := NewMysqld(config.DefaultBackupConfig(), log)
	err := mysqld.StartMysqld()
	assert.Nil(t, err)
}

func TestStopMysqld(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	conf := config.DefaultBackupConfig()
	mysqld := NewMysqld(conf, log)
	err := mysqld.StopMysqld()
	assert.Nil(t, err)
}

func TestKillMysqld(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	conf := config.DefaultBackupConfig()
	mysqld := NewMysqld(conf, log)

	// mock a mysqld running
	go func() {
		args := []string{
			"-c",
			fmt.Sprintf("watch -d 1 '--defaults-file=%v'", conf.DefaultsFile)}
		common.RunCommand("bash", args...)
	}()

	err := mysqld.KillMysqld()
	assert.Nil(t, err)
}

func TestIsMysqldRunning(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	conf := config.DefaultBackupConfig()
	mysqld := NewMysqld(conf, log)

	// mock a mysqld running
	go func() {
		args := []string{
			"-c",
			fmt.Sprintf("watch -d 1 'mysqld_safe --defaults-file=%v'", conf.DefaultsFile)}
		common.RunCommand("bash", args...)
	}()

	want := true
	got := mysqld.isMysqldRunning()
	assert.Equal(t, want, got)
}

func TestMonitor(t *testing.T) {
	conf := config.DefaultBackupConfig()
	// 100ms
	conf.MysqldMonitorInterval = 100
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	conf.DefaultsFile = "/etc/my.cnf"
	mysqld := NewMysqld(conf, log)

	{
		want := model.MYSQLD_NOTRUNNING
		got := mysqld.getStatus()
		assert.Equal(t, want, got)
		mysqld.MonitorStart()
		time.Sleep(500 * time.Millisecond)

		want = model.MYSQLD_NOTRUNNING
		got = mysqld.getStatus()
		assert.Equal(t, want, got)
	}

	{
		mysqld.MonitorStop()

		wantstatus := model.MYSQLD_UNKNOW
		gotstatus := mysqld.getStatus()
		assert.Equal(t, wantstatus, gotstatus)

		want := false
		got := mysqld.monitorRunning
		assert.Equal(t, want, got)
	}

	{
		want := false
		got := mysqld.monitorRunning
		assert.Equal(t, want, got)
		mysqld.MonitorStart()
		time.Sleep(500 * time.Millisecond)

		wantstatus := model.MYSQLD_NOTRUNNING
		gotstatus := mysqld.getStatus()
		assert.Equal(t, wantstatus, gotstatus)
	}
}
