/*
 * Xenon
 *
 * Copyright 2018 The Xenon Authors.
 * Code is licensed under the GPLv3.
 *
 */

package mysql

import (
	"database/sql"
	"model"
)

// ReplHandler interface.
type ReplHandler interface {
	// check health and return log_bin_basename
	Ping(*sql.DB) (*PingEntry, error)

	// set mysql readonly variable
	SetReadOnly(*sql.DB, bool) error

	// get GTID from traversal binlog folder and find the newest one
	GetMasterGTID(*sql.DB) (*model.GTID, error)

	// get GTID from SHOW SLAVE STATUS
	GetSlaveGTID(*sql.DB) (*model.GTID, error)

	// start slave io_thread
	StartSlaveIOThread(*sql.DB) error

	// stop slave io_thread
	StopSlaveIOThread(*sql.DB) error

	// start slave
	StartSlave(*sql.DB) error

	// stop slave
	StopSlave(*sql.DB) error

	// use the provided master as the new master
	ChangeMasterTo(*sql.DB, *model.Repl) error

	// change a slave to master
	ChangeToMaster(*sql.DB) error

	// waits until slave replication reaches at least targetGTID
	WaitUntilAfterGTID(*sql.DB, string) error

	// set global variables
	SetGlobalSysVar(db *sql.DB, varsql string) error

	// reset master
	ResetMaster(db *sql.DB) error

	// reset slave all
	ResetSlaveAll(db *sql.DB) error

	// purge binglog to
	PurgeBinlogsTo(*sql.DB, string) error

	// enable master semi sync: wait slave ack
	EnableSemiSyncMaster(db *sql.DB) error

	// disable master semi sync: don't wait slave ack
	DisableSemiSyncMaster(db *sql.DB) error

	// set semi-sync master-timeout = default
	SetSemiSyncMasterDefault(db *sql.DB) error

	//set rpl_semi_master_wait_for_slave_count
	SetSemiWaitSlaveCount(db *sql.DB, count int) error
}
