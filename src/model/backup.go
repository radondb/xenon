/*
 * Xenon
 *
 * Copyright 2018 The Xenon Authors.
 * Code is licensed under the GPLv3.
 *
 */

package model

const (
	RPCBackupStatus   = "BackupRPC.GetBackupStatus"
	RPCBackupDo       = "BackupRPC.DoBackup"
	RPCBackupCancel   = "BackupRPC.CancelBackup"
	RPCBackupApplyLog = "BackupRPC.DoApplyLog"
)

type BackupStats struct {
	// How many times backup have been called
	Backups uint64

	// How many times backup have failed
	BackupErrs uint64

	// How many times apply-log have been called
	AppLogs uint64

	// How many times apply-log have failed
	AppLogErrs uint64

	// How many times cannel have been taken
	Cancels uint64

	// The last error message of backup/applylog
	LastError string

	// The last backup command info  we call
	LastCMD string
}

type BackupRPCRequest struct {
	// The IP of this request
	From string

	// The Backup dir of this request
	BackupDir string

	// The SSH IP of this request
	SSHHost string

	// The SSH user of this request
	SSHUser string

	// The SSH password of this request
	SSHPasswd string

	// The SSH port(default is 22) of this request
	SSHPort int

	// The Backup IOPS throttle of this request
	IOPSLimits int

	// The xtrabackup/xbstream binary dir
	XtrabackupBinDir string
}

type BackupRPCResponse struct {
	// Return code to rpc client:
	// OK or other errors
	RetCode string
}

func NewBackupRPCRequest() *BackupRPCRequest {
	return &BackupRPCRequest{}
}

func NewBackupRPCResponse(code string) *BackupRPCResponse {
	return &BackupRPCResponse{RetCode: code}
}
