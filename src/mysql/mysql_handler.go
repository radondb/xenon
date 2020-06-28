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

// MysqlHandler interface.
type MysqlHandler interface {
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

	// get local uuid
	GetUUID(db *sql.DB) (string, error)

	// get gtid subtract with slavegtid and master gtid
	GetGtidSubtract(*sql.DB, string, string) (string, error)

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

	// set semi-sync master-timeout
	SetSemiSyncMasterTimeout(db *sql.DB, timeout uint64) error

	//set rpl_semi_master_wait_for_slave_count
	SetSemiWaitSlaveCount(db *sql.DB, count int) error

	// User handlers.
	GetUser(*sql.DB) ([]model.MysqlUser, error)
	CheckUserExists(*sql.DB, string, string) (bool, error)
	CreateUser(*sql.DB, string, string, string, string) error
	DropUser(*sql.DB, string, string) error
	ChangeUserPasswd(*sql.DB, string, string, string) error
	CreateReplUserWithoutBinlog(*sql.DB, string, string) error
	GrantAllPrivileges(*sql.DB, string, string, string, string) error
	GrantNormalPrivileges(*sql.DB, string, string) error
	CreateUserWithPrivileges(db *sql.DB, user, passwd, database, table, host, privs string, ssl string) error
	GrantReplicationPrivileges(*sql.DB, string) error
}

var (
	handlers = make(map[string]MysqlHandler)
)

func init() {
	handlers["mysql56"] = new(Mysql56)
	handlers["mysql57"] = new(Mysql57)
}

func getHandler(name string) MysqlHandler {
	handler, ok := handlers[name]
	if !ok {
		return new(Mysql57)
	}
	return handler
}
