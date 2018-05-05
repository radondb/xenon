/*
 * Xenon
 *
 * Copyright 2018 The Xenon Authors.
 * Code is licensed under the GPLv3.
 *
 */

package mysqld

import (
	"model"
)

// BackupRPC tuple.
type BackupRPC struct {
	mysqld *Mysqld
}

// GetBackupRPC returns BackupRPC tuple.
func (m *Mysqld) GetBackupRPC() *BackupRPC {
	return &BackupRPC{m}
}

// DoBackup used to execute the xtrabackup command.
func (b *BackupRPC) DoBackup(req *model.BackupRPCRequest, rsp *model.BackupRPCResponse) error {
	rsp.RetCode = model.OK
	err := b.mysqld.backup.Backup(req)
	if err != nil {
		rsp.RetCode = err.Error()
		return nil
	}
	return nil
}

// DoApplyLog used to execute the apply log command.
func (b *BackupRPC) DoApplyLog(req *model.BackupRPCRequest, rsp *model.BackupRPCResponse) error {
	rsp.RetCode = model.OK
	err := b.mysqld.backup.ApplyLog(req)
	if err != nil {
		rsp.RetCode = err.Error()
		return nil
	}
	return nil
}

// CancelBackup used to cancel the job of backup.
func (b *BackupRPC) CancelBackup(req *model.BackupRPCRequest, rsp *model.BackupRPCResponse) error {
	rsp.RetCode = model.OK
	err := b.mysqld.backup.Cancel()
	if err != nil {
		rsp.RetCode = err.Error()
		return nil
	}
	return nil
}
