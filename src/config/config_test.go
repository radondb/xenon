/*
 * Xenon
 *
 * Copyright 2018 The Xenon Authors.
 * Code is licensed under the GPLv3.
 *
 */

package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseConfig(t *testing.T) {
	data := []byte(
		`{
	"server":
	{
		"endpoint":"127.0.0.1:8080",
		"enable-apis":false,
		"peer-address":":6060"
	},

	"raft":
	{
		"meta-datadir":".",
		"heartbeat-timeout":1000,
		"election-timeout":3000,
		"leader-start-command":"nop",
		"leader-stop-command":"nop"
	},

	"mysql":
	{
		"admin":"root",
		"passwd":"",
		"host":"localhost",
		"port":3306,
		"basedir":"/u01/mysql_20160606/",
		"defaults-file":"/etc/my3306.cnf",
		"ping-timeout":1000
	},

	"replication":
	{
		"user":"repl",
		"passwd":"repl"
	},

	"backup":
	{
		"ssh-port":22,
		"backupdir":"/u01/backup",
		"xtrabackup-bindir":".",
		"backup-iops-limits":100000,
		"backup-use-memory":"2GB",
		"backup-parallel": 2,
		"mysqld-monitor-interval": 1000
	},

	"rpc":
	{
		"request-timeout":1000
	},

	"log":
	{
		"level":"INFO"
	}
}
`)

	got, err := parseConfig(data)
	want := DefaultConfig()
	assert.Nil(t, err)
	assert.Equal(t, want, got)
}

func TestWriteConfig(t *testing.T) {
	path := "/tmp/test.config.json"
	os.Remove(path)
	conf := DefaultConfig()
	err := WriteConfig(path, conf)
	assert.Nil(t, err)
}

func TestLoadConfig(t *testing.T) {
	path := "/tmp/test.config.json"
	want := DefaultConfig()
	got, err := LoadConfig(path)
	assert.Nil(t, err)
	assert.Equal(t, want, got)
	os.Remove(path)
}
