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
	"model"
	"testing"
	"xbase/common"
	"xbase/xlog"

	"github.com/stretchr/testify/assert"
)

func TestBackupCommand(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.DEBUG))
	conf := config.DefaultBackupConfig()
	backup := NewBackup(conf, log)
	backup.SetCMDHandler(common.NewMockACommand())

	req := model.NewBackupRPCRequest()
	req.From = "127.0.0.2"
	req.BackupDir = "/u01/backup"
	req.XtrabackupBinDir = "/u01/xtrabackup_20161216"
	req.SSHPasswd = "sshpasswd"
	req.SSHUser = "user"
	req.SSHHost = "127.0.0.1"
	req.SSHPort = 22
	req.IOPSLimits = 100

	// test xtrabackup commands
	{
		got := backup.backupCommands(true, req)
		want := []string{
			"-c",
			"./xtrabackup --defaults-file=/etc/my3306.cnf --host=localhost --port=3306 --user=root --backup --throttle=100 --parallel=2 --stream=xbstream --target-dir=./ | ssh -o 'StrictHostKeyChecking=no' user@127.0.0.1 -p 22 \"/u01/xtrabackup_20161216/xbstream -x -C /u01/backup\"",
		}
		assert.Equal(t, want, got)

		// ssh with password
		got = backup.backupCommands(false, req)
		want = []string{
			"-c",
			"./xtrabackup --defaults-file=/etc/my3306.cnf --host=localhost --port=3306 --user=root --backup --throttle=100 --parallel=2 --stream=xbstream --target-dir=./ | sshpass -p sshpasswd ssh -o 'StrictHostKeyChecking=no' user@127.0.0.1 -p 22 \"/u01/xtrabackup_20161216/xbstream -x -C /u01/backup\"",
		}
		assert.Equal(t, want, got)
	}

	// test with innodbbackup password
	{
		conf.Passwd = "123"
		got := backup.backupCommands(true, req)
		want := []string{
			"-c",
			"./xtrabackup --defaults-file=/etc/my3306.cnf --host=localhost --port=3306 --user=root --password=123 --backup --throttle=100 --parallel=2 --stream=xbstream --target-dir=./ | ssh -o 'StrictHostKeyChecking=no' user@127.0.0.1 -p 22 \"/u01/xtrabackup_20161216/xbstream -x -C /u01/backup\"",
		}
		assert.Equal(t, want, got)

		got = backup.backupCommands(false, req)
		want = []string{
			"-c",
			"./xtrabackup --defaults-file=/etc/my3306.cnf --host=localhost --port=3306 --user=root --password=123 --backup --throttle=100 --parallel=2 --stream=xbstream --target-dir=./ | sshpass -p sshpasswd ssh -o 'StrictHostKeyChecking=no' user@127.0.0.1 -p 22 \"/u01/xtrabackup_20161216/xbstream -x -C /u01/backup\"",
		}
		assert.Equal(t, want, got)

	}

	// test backup and cancel
	{
		err := backup.Backup(req)
		assert.Nil(t, err)

		err = backup.Cancel()
		assert.Nil(t, err)
	}

	// test backup last error
	{
		got := backup.getLastError()
		want := ""
		assert.Equal(t, want, got)
	}
}

func TestApplyLog(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	conf := config.DefaultBackupConfig()
	backup := NewBackup(conf, log)

	req := model.NewBackupRPCRequest()
	req.BackupDir = "/tmp/xtrabackup_test"
	// test commands
	{
		got := backup.applylogCommands(req)
		want := []string{
			"-c",
			"./xtrabackup --defaults-file=/etc/my3306.cnf --use-memory=2GB --prepare --target-dir=/tmp/xtrabackup_test",
		}
		assert.Equal(t, want, got)
	}

	// test apply-log and cancel
	{
		err := backup.ApplyLog(req)
		assert.NotNil(t, err)
		backup.Cancel()
	}

	// test applylog
	{
		got := backup.getLastError()
		want := "cmd.outs.[completed OK!].found[0]!=expects[1]"
		assert.Equal(t, want, got)
	}
}

func TestCheckSSH(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	conf := config.DefaultBackupConfig()
	backup := NewBackup(conf, log)
	req := model.NewBackupRPCRequest()

	// run command OK
	{
		backup.SetCMDHandler(common.NewMockCommand())
		// test pass
		{
			got := backup.checkSSHTunnelWithPass(req)
			want := true
			assert.Equal(t, want, got)
		}

		// test key
		{
			req := model.NewBackupRPCRequest()
			got := backup.checkSSHTunnelWithKey(req)
			want := true
			assert.Equal(t, want, got)
		}
	}

	// run command error
	{
		backup.SetCMDHandler(common.NewMockBCommand())
		// test pass
		{
			got := backup.checkSSHTunnelWithPass(req)
			want := false
			assert.Equal(t, want, got)
		}

		// test key
		{
			got := backup.checkSSHTunnelWithKey(req)
			want := false
			assert.Equal(t, want, got)
		}

		// test backup under run cmd error
		{
			err := backup.Backup(req)
			want := "backup.ssh.tunnel.to[@ port:0 passwd:].can.not.connect"
			got := err.Error()
			assert.Equal(t, want, got)
		}
	}
}
