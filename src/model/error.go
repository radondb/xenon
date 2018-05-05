/*
 * Xenon
 *
 * Copyright 2018 The Xenon Authors.
 * Code is licensed under the GPLv3.
 *
 */

package model

const (
	OK                    = "OK"
	ErrorRPCCall          = "ErrorRpcCall"
	ErrorMySQLDown        = "ErrorMySQLDown"
	ErrorServerDown       = "ErrorServerDown"
	ErrorInvalidGTID      = "ErrorInvalidGTID"
	ErrorInvalidViewID    = "ErrorInvalidViewID"
	ErrorVoteNotGranted   = "ErrorVoteNotGranted"
	ErrorInvalidRequest   = "ErrorInvalidRequest"
	ErrorChangeMaster     = "ErrorChangeMaster"
	ErrorBackupNotFound   = "ErrorBackupNotFound"
	ErrorMysqldNotRunning = "ErrorMysqldNotRunning"
)

const (
	RPCError_MySQLUnpromotable = "RPCError_MySQLUnpromotable"
)
