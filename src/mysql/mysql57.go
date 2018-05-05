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
	"fmt"
	"model"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

var (
	_ ReplHandler = &Mysql57{}
)

var (
	// timeout is 5s
	reqTimeout = 1000 * 5
)

// Mysql57 tuple.
type Mysql57 struct {
	ReplHandler
}

// Ping has 2 affects:
// one for heath check
// other for get master_binglog the slave is syncing
func (my *Mysql57) Ping(db *sql.DB) (*PingEntry, error) {
	pe := &PingEntry{}
	query := "SHOW SLAVE STATUS"
	rows, err := QueryWithTimeout(db, reqTimeout, query)
	if err != nil {
		return nil, err
	}
	if len(rows) > 0 {
		pe.Relay_Master_Log_File = rows[0]["Relay_Master_Log_File"]
	}
	return pe, nil
}

// SetReadOnly used to set mysql to readonly.
func (my *Mysql57) SetReadOnly(db *sql.DB, readonly bool) error {
	enabled := 0
	if readonly {
		enabled = 1
	}

	cmds := []string{}
	cmds = append(cmds, fmt.Sprintf("SET GLOBAL read_only = %d", enabled))
	// Set super_read_only on the slave.
	// https://dev.mysql.com/doc/refman/5.7/en/server-system-variables.html#sysvar_super_read_only
	cmds = append(cmds, fmt.Sprintf("SET GLOBAL super_read_only = %d", enabled))
	return ExecuteSuperQueryListWithTimeout(db, reqTimeout, cmds)
}

// GetSlaveGTID gets the gtid from the default channel.
// Here, We just show the default slave channel.
func (my *Mysql57) GetSlaveGTID(db *sql.DB) (*model.GTID, error) {
	gtid := &model.GTID{}

	query := "SHOW SLAVE STATUS FOR CHANNEL ''"
	rows, err := QueryWithTimeout(db, reqTimeout, query)
	if err != nil {
		return gtid, err
	}
	if len(rows) > 0 {
		row := rows[0]
		gtid.Master_Log_File = row["Master_Log_File"]
		gtid.Read_Master_Log_Pos, _ = strconv.ParseUint(row["Read_Master_Log_Pos"], 10, 64)
		gtid.Retrieved_GTID_Set = row["Retrieved_Gtid_Set"]
		gtid.Executed_GTID_Set = row["Executed_Gtid_Set"]
		gtid.Slave_IO_Running = (row["Slave_IO_Running"] == "Yes")
		gtid.Slave_IO_Running_Str = row["Slave_IO_Running"]
		gtid.Slave_SQL_Running = (row["Slave_SQL_Running"] == "Yes")
		gtid.Slave_SQL_Running_Str = row["Slave_SQL_Running"]
		gtid.Seconds_Behind_Master = row["Seconds_Behind_Master"]
		gtid.Last_Error = row["Last_Error"]
		gtid.Last_IO_Error = row["Last_IO_Error"]
		gtid.Last_SQL_Error = row["Last_SQL_Error"]
		gtid.Slave_SQL_Running_State = row["Slave_SQL_Running_State"]
	}
	return gtid, nil
}

// GetMasterGTID used to get binlog info from master.
func (my *Mysql57) GetMasterGTID(db *sql.DB) (*model.GTID, error) {
	gtid := &model.GTID{}

	query := "SHOW MASTER STATUS"
	rows, err := QueryWithTimeout(db, reqTimeout, query)
	if err != nil {
		return nil, err
	}
	if len(rows) > 0 {
		row := rows[0]
		gtid.Master_Log_File = row["File"]
		gtid.Read_Master_Log_Pos, _ = strconv.ParseUint(row["Position"], 10, 64)
		gtid.Executed_GTID_Set = row["Executed_Gtid_Set"]
		gtid.Seconds_Behind_Master = "0"
		gtid.Slave_IO_Running = true
		gtid.Slave_SQL_Running = true
	}
	return gtid, nil
}

// StartSlaveIOThread used to start the io thread.
func (my *Mysql57) StartSlaveIOThread(db *sql.DB) error {
	cmd := "START SLAVE IO_THREAD FOR CHANNEL ''"
	return ExecuteWithTimeout(db, reqTimeout, cmd)
}

// StopSlaveIOThread used to stop the op thread.
func (my *Mysql57) StopSlaveIOThread(db *sql.DB) error {
	cmd := "STOP SLAVE IO_THREAD FOR CHANNEL ''"
	return ExecuteWithTimeout(db, reqTimeout, cmd)
}

// StartSlave used to start slave.
func (my *Mysql57) StartSlave(db *sql.DB) error {
	cmd := "START SLAVE FOR CHANNEL ''"
	return ExecuteWithTimeout(db, reqTimeout, cmd)
}

// StopSlave used to stop the slave.
func (my *Mysql57) StopSlave(db *sql.DB) error {
	cmd := "STOP SLAVE FOR CHANNEL ''"
	return ExecuteWithTimeout(db, reqTimeout, cmd)
}

func (my *Mysql57) changeMasterToCommands(master *model.Repl) []string {
	var args []string

	args = append(args, fmt.Sprintf("MASTER_HOST = '%s'", master.Master_Host))
	args = append(args, fmt.Sprintf("MASTER_PORT = %d", master.Master_Port))
	args = append(args, fmt.Sprintf("MASTER_USER = '%s'", master.Repl_User))
	args = append(args, fmt.Sprintf("MASTER_PASSWORD = '%s'", master.Repl_Password))
	args = append(args, "MASTER_AUTO_POSITION = 1")
	changeMasterTo := "CHANGE MASTER TO\n  " + strings.Join(args, ",\n  ") + " FOR CHANNEL ''"
	return []string{changeMasterTo}
}

// ChangeMasterTo stop for all channels and reset all replication filter to null.
// In Xenon, we never set replication filter.
func (my *Mysql57) ChangeMasterTo(db *sql.DB, master *model.Repl) error {
	cmds := []string{}
	cmds = append(cmds, "STOP SLAVE")
	cmds = append(cmds, "CHANGE REPLICATION FILTER REPLICATE_DO_DB=(),REPLICATE_IGNORE_DB=(),REPLICATE_DO_TABLE=(),REPLICATE_IGNORE_TABLE=(),REPLICATE_WILD_DO_TABLE=(),REPLICATE_WILD_IGNORE_TABLE=(),REPLICATE_REWRITE_DB=()")
	cmds = append(cmds, my.changeMasterToCommands(master)...)
	cmds = append(cmds, "START SLAVE FOR CHANNEL ''")
	return ExecuteSuperQueryListWithTimeout(db, reqTimeout, cmds)
}

// ChangeToMaster changes a slave to be master.
func (my *Mysql57) ChangeToMaster(db *sql.DB) error {
	cmds := []string{"STOP SLAVE",
		"RESET SLAVE ALL"} //"ALL" makes it forget the master host:port
	return ExecuteSuperQueryListWithTimeout(db, reqTimeout, cmds)
}

// WaitUntilAfterGTID used to do 'SELECT WAIT_UNTIL_SQL_THREAD_AFTER_GTIDS' command.
// https://dev.mysql.com/doc/refman/5.7/en/gtid-functions.html
func (my *Mysql57) WaitUntilAfterGTID(db *sql.DB, targetGTID string) error {
	query := fmt.Sprintf("SELECT WAIT_UNTIL_SQL_THREAD_AFTER_GTIDS('%s')", targetGTID)
	return Execute(db, query)
}

// SetGlobalSysVar used to set global variables.
func (my *Mysql57) SetGlobalSysVar(db *sql.DB, varsql string) error {
	prefix := "SET GLOBAL"
	if !strings.HasPrefix(varsql, prefix) {
		return errors.Errorf("[%v].must.be.startwith:%v", varsql, prefix)
	}
	return ExecuteWithTimeout(db, reqTimeout, varsql)
}

// ResetMaster used to reset master.
func (my *Mysql57) ResetMaster(db *sql.DB) error {
	cmds := "RESET MASTER"
	return ExecuteWithTimeout(db, reqTimeout, cmds)
}

// ResetSlaveAll used to reset slave.
func (my *Mysql57) ResetSlaveAll(db *sql.DB) error {
	cmds := []string{"STOP SLAVE",
		"RESET SLAVE ALL"} //"ALL" makes it forget the master host:port
	return ExecuteSuperQueryListWithTimeout(db, reqTimeout, cmds)
}

// PurgeBinlogsTo used to purge binlog.
func (my *Mysql57) PurgeBinlogsTo(db *sql.DB, binlog string) error {
	cmds := fmt.Sprintf("PURGE BINARY LOGS TO '%s'", binlog)
	return ExecuteWithTimeout(db, reqTimeout, cmds)
}

// EnableSemiSyncMaster used to enable the semi-sync on master.
func (my *Mysql57) EnableSemiSyncMaster(db *sql.DB) error {
	cmds := "SET GLOBAL rpl_semi_sync_master_enabled=ON"
	return ExecuteWithTimeout(db, reqTimeout, cmds)
}

// DisableSemiSyncMaster used to disable the semi-sync from master.
func (my *Mysql57) DisableSemiSyncMaster(db *sql.DB) error {
	cmds := "SET GLOBAL rpl_semi_sync_master_enabled=OFF"
	return ExecuteWithTimeout(db, reqTimeout, cmds)
}
