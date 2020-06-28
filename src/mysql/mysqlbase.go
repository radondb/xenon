/*
 * Xenon
 *
 * Copyright 2018-2019 The Xenon Authors.
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

// http://dev.mysql.com/doc/refman/5.7/en/privileges-provided.html
var (
	mysqlAllPrivileges = []string{
		"ALL",
	}

	mysqlReplPrivileges = []string{
		"REPLICATION SLAVE",
		"REPLICATION CLIENT",
	}

	mysqlNormalPrivileges = []string{
		"ALTER", "ALTER ROUTINE", "CREATE", "CREATE ROUTINE",
		"CREATE TEMPORARY TABLES", "CREATE VIEW", "DELETE",
		"DROP", "EXECUTE", "EVENT", "INDEX", "INSERT",
		"LOCK TABLES", "PROCESS", "RELOAD", "SELECT",
		"SHOW DATABASES", "SHOW VIEW", "UPDATE", "TRIGGER", "REFERENCES",
		"REPLICATION SLAVE", "REPLICATION CLIENT",
	}

	mysqlSSLType = []string{
		"YES", "NO",
	}

	_ MysqlHandler = &MysqlBase{}
	// timeout is 10s
	reqTimeout = 10000
)

// MysqlBase tuple.
type MysqlBase struct {
	MysqlHandler
}

// Ping has 2 affects:
// one for heath check
// other for get master_binglog the slave is syncing
func (my *MysqlBase) Ping(db *sql.DB) (*PingEntry, error) {
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
func (my *MysqlBase) SetReadOnly(db *sql.DB, readonly bool) error {
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
func (my *MysqlBase) GetSlaveGTID(db *sql.DB) (*model.GTID, error) {
	gtid := &model.GTID{}

	query := "SHOW SLAVE STATUS"
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
func (my *MysqlBase) GetMasterGTID(db *sql.DB) (*model.GTID, error) {
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

// GetUUID used to get local uuid.
func (my *MysqlBase) GetUUID(db *sql.DB) (string, error) {
	uuid := ""
	query := "SELECT @@SERVER_UUID"
	rows, err := QueryWithTimeout(db, reqTimeout, query)
	if err != nil {
		return uuid, err
	}
	if len(rows) > 0 {
		row := rows[0]
		uuid = row["@@SERVER_UUID"]
	}

	return uuid, nil
}

// StartSlaveIOThread used to start the io thread.
func (my *MysqlBase) StartSlaveIOThread(db *sql.DB) error {
	cmd := "START SLAVE IO_THREAD"
	return ExecuteWithTimeout(db, reqTimeout, cmd)
}

// StopSlaveIOThread used to stop the op thread.
func (my *MysqlBase) StopSlaveIOThread(db *sql.DB) error {
	cmd := "STOP SLAVE IO_THREAD"
	return ExecuteWithTimeout(db, reqTimeout, cmd)
}

// StartSlave used to start slave.
func (my *MysqlBase) StartSlave(db *sql.DB) error {
	cmd := "START SLAVE"
	return ExecuteWithTimeout(db, reqTimeout, cmd)
}

// StopSlave used to stop the slave.
func (my *MysqlBase) StopSlave(db *sql.DB) error {
	cmd := "STOP SLAVE"
	return ExecuteWithTimeout(db, reqTimeout, cmd)
}

func (my *MysqlBase) changeMasterToCommands(master *model.Repl) []string {
	var args []string

	args = append(args, fmt.Sprintf("MASTER_HOST = '%s'", master.Master_Host))
	args = append(args, fmt.Sprintf("MASTER_PORT = %d", master.Master_Port))
	args = append(args, fmt.Sprintf("MASTER_USER = '%s'", master.Repl_User))
	args = append(args, fmt.Sprintf("MASTER_PASSWORD = '%s'", master.Repl_Password))
	args = append(args, "MASTER_AUTO_POSITION = 1")
	changeMasterTo := "CHANGE MASTER TO\n  " + strings.Join(args, ",\n  ")
	return []string{changeMasterTo}
}

// ChangeMasterTo stop for all channels and reset all replication filter to null.
// In Xenon, we never set replication filter.
func (my *MysqlBase) ChangeMasterTo(db *sql.DB, master *model.Repl) error {
	cmds := []string{}
	cmds = append(cmds, "STOP SLAVE")
	cmds = append(cmds, my.changeMasterToCommands(master)...)
	cmds = append(cmds, "START SLAVE")
	return ExecuteSuperQueryListWithTimeout(db, reqTimeout, cmds)
}

// ChangeToMaster changes a slave to be master.
func (my *MysqlBase) ChangeToMaster(db *sql.DB) error {
	cmds := []string{"STOP SLAVE",
		"RESET SLAVE ALL"} //"ALL" makes it forget the master host:port
	return ExecuteSuperQueryListWithTimeout(db, reqTimeout, cmds)
}

// WaitUntilAfterGTID used to do 'SELECT WAIT_UNTIL_SQL_THREAD_AFTER_GTIDS' command.
// https://dev.mysql.com/doc/refman/5.7/en/gtid-functions.html
func (my *MysqlBase) WaitUntilAfterGTID(db *sql.DB, targetGTID string) error {
	query := fmt.Sprintf("SELECT WAIT_UNTIL_SQL_THREAD_AFTER_GTIDS('%s')", targetGTID)
	return Execute(db, query)
}

// GetGtidSubtract used to do "SELECT GTID_SUBTRACT('subsetGTID','setGTID') as gtid_sub" command
func (my *MysqlBase) GetGtidSubtract(db *sql.DB, subsetGTID string, setGTID string) (string, error) {
	query := fmt.Sprintf("SELECT GTID_SUBTRACT('%s','%s') as gtid_sub", subsetGTID, setGTID)
	rows, err := QueryWithTimeout(db, reqTimeout, query)
	if err != nil {
		return "", err
	}

	if len(rows) > 0 {
		row := rows[0]
		gtid_sub := row["gtid_sub"]
		return gtid_sub, nil
	}
	return "", nil
}

// SetGlobalSysVar used to set global variables.
func (my *MysqlBase) SetGlobalSysVar(db *sql.DB, varsql string) error {
	prefix := "SET GLOBAL"
	if !strings.HasPrefix(varsql, prefix) {
		return errors.Errorf("[%v].must.be.startwith:%v", varsql, prefix)
	}
	return ExecuteWithTimeout(db, reqTimeout, varsql)
}

// ResetMaster used to reset master.
func (my *MysqlBase) ResetMaster(db *sql.DB) error {
	cmds := "RESET MASTER"
	return ExecuteWithTimeout(db, reqTimeout, cmds)
}

// ResetSlaveAll used to reset slave.
func (my *MysqlBase) ResetSlaveAll(db *sql.DB) error {
	cmds := []string{"STOP SLAVE",
		"RESET SLAVE ALL"} //"ALL" makes it forget the master host:port
	return ExecuteSuperQueryListWithTimeout(db, reqTimeout, cmds)
}

// PurgeBinlogsTo used to purge binlog.
func (my *MysqlBase) PurgeBinlogsTo(db *sql.DB, binlog string) error {
	cmds := fmt.Sprintf("PURGE BINARY LOGS TO '%s'", binlog)
	return ExecuteWithTimeout(db, reqTimeout, cmds)
}

// EnableSemiSyncMaster used to enable the semi-sync on master.
func (my *MysqlBase) EnableSemiSyncMaster(db *sql.DB) error {
	cmds := "SET GLOBAL rpl_semi_sync_master_enabled=ON"
	return ExecuteWithTimeout(db, reqTimeout, cmds)
}

//SetSemiWaitSlaveCount used set rpl_semi_sync_master_wait_for_slave_count
func (my *MysqlBase) SetSemiWaitSlaveCount(db *sql.DB, count int) error {
	cmds := fmt.Sprintf("SET GLOBAL rpl_semi_sync_master_wait_for_slave_count = %d", count)
	return ExecuteWithTimeout(db, reqTimeout, cmds)
}

// DisableSemiSyncMaster used to disable the semi-sync from master.
func (my *MysqlBase) DisableSemiSyncMaster(db *sql.DB) error {
	cmds := "SET GLOBAL rpl_semi_sync_master_enabled=OFF"
	return ExecuteWithTimeout(db, reqTimeout, cmds)
}

// SetSemiSyncMasterTimeout used to set semi-sync master timeout
func (my *MysqlBase) SetSemiSyncMasterTimeout(db *sql.DB, timeout uint64) error {
	cmds := fmt.Sprintf("SET GLOBAL rpl_semi_sync_master_timeout=%d", timeout)
	return ExecuteWithTimeout(db, reqTimeout, cmds)
}

// CheckUserExists used to check the user exists or not.
func (my *MysqlBase) CheckUserExists(db *sql.DB, user string, host string) (bool, error) {
	query := fmt.Sprintf("SELECT User FROM mysql.user WHERE User = '%s' and Host = '%s'", user, host)
	rows, err := Query(db, query)
	if err != nil {
		return false, err
	}
	if len(rows) > 0 {
		return true, nil
	}
	return false, nil
}

// GetUser used to get the mysql user list
func (my *MysqlBase) GetUser(db *sql.DB) ([]model.MysqlUser, error) {
	query := fmt.Sprintf("SELECT User, Host, Super_priv FROM mysql.user")
	rows, err := Query(db, query)
	if err != nil {
		return nil, err
	}

	var Users = make([]model.MysqlUser, len(rows))
	for i, v := range rows {
		Users[i].User = v["User"]
		Users[i].Host = v["Host"]
		Users[i].SuperPriv = v["Super_priv"]
	}
	return Users, nil
}

// CreateUser use to create new user.
// see http://dev.mysql.com/doc/refman/5.7/en/string-literals.html
func (my *MysqlBase) CreateUser(db *sql.DB, user string, host string, passwd string, ssltype string) error {
	query := fmt.Sprintf("CREATE USER `%s`@`%s` IDENTIFIED BY '%s'", user, host, passwd)
	if ssltype == "YES" {
		query = fmt.Sprintf("%s REQUIRE X509", query)
	}
	return Execute(db, query)
}

// DropUser used to drop the user.
func (my *MysqlBase) DropUser(db *sql.DB, user string, host string) error {
	query := fmt.Sprintf("DROP USER `%s`@`%s`", user, host)
	return Execute(db, query)
}

// CreateReplUserWithoutBinlog create replication accounts without writing binlog.
func (my *MysqlBase) CreateReplUserWithoutBinlog(db *sql.DB, user string, passwd string) error {
	queryList := []string{
		"SET sql_log_bin=0",
		fmt.Sprintf("CREATE USER `%s` IDENTIFIED BY '%s'", user, passwd),
		fmt.Sprintf("GRANT %s ON *.* TO `%s`", strings.Join(mysqlReplPrivileges, ","), user),
		"SET sql_log_bin=1",
	}
	return ExecuteSuperQueryList(db, queryList)
}

// ChangeUserPasswd used to change the user password.
func (my *MysqlBase) ChangeUserPasswd(db *sql.DB, user string, host string, passwd string) error {
	query := fmt.Sprintf("ALTER USER `%s`@`%s` IDENTIFIED BY '%s'", user, host, passwd)
	return Execute(db, query)
}

// GrantNormalPrivileges used to grants normal privileges.
func (my *MysqlBase) GrantNormalPrivileges(db *sql.DB, user string, host string) error {
	query := fmt.Sprintf("GRANT %s ON *.* TO `%s`@`%s`", strings.Join(mysqlNormalPrivileges, ","), user, host)
	return my.grantPrivileges(db, query)
}

// CreateUserWithPrivileges for create normal user.
func (my *MysqlBase) CreateUserWithPrivileges(db *sql.DB, user, passwd, database, table, host, privs string, ssl string) error {
	// build normal privs map
	var query string
	normal := make(map[string]string)
	for _, priv := range mysqlNormalPrivileges {
		normal[priv] = priv
	}

	// check privs
	privs = strings.TrimSuffix(privs, ",")
	privsList := strings.Split(privs, ",")
	for _, priv := range privsList {
		priv = strings.ToUpper(strings.TrimSpace(priv))
		if _, ok := normal[priv]; !ok {
			return errors.Errorf("can't create user[%v] with privileges[%v]", user, priv)
		}
	}

	// build standard ssl_type map
	standardSSL := make(map[string]string)
	for _, ssltype := range mysqlSSLType {
		standardSSL[ssltype] = ssltype
	}

	// check ssl_type
	ssltype := strings.TrimSpace(ssl)
	if _, ok := standardSSL[ssltype]; !ok {
		return errors.Errorf("can't create user[%v] require ssl_type[%v]", user, ssltype)
	}

	if err := my.CreateUser(db, user, host, passwd, ssltype); err != nil {
		return errors.Errorf("create user[%v] with privileges[%v] require ssl_type[%v] failed: [%v]", user, privs, ssltype, err)
	}

	query = fmt.Sprintf("GRANT %s ON %s.%s TO `%s`@`%s`", privs, database, table, user, host)
	return my.grantPrivileges(db, query)
}

// GrantReplicationPrivileges used to grant repli privis.
func (my *MysqlBase) GrantReplicationPrivileges(db *sql.DB, user string) error {
	query := fmt.Sprintf("GRANT %s ON *.* TO `%s`", strings.Join(mysqlReplPrivileges, ","), user)
	return my.grantPrivileges(db, query)
}

// GrantAllPrivileges used to grant all privis.
func (my *MysqlBase) GrantAllPrivileges(db *sql.DB, user string, host string, passwd string, ssl string) error {
	var query string
	// build standard ssl_type map
	standardSSL := make(map[string]string)
	for _, ssltype := range mysqlSSLType {
		standardSSL[ssltype] = ssltype
	}

	// check ssl_type
	ssltype := strings.TrimSpace(ssl)
	if _, ok := standardSSL[ssltype]; !ok {
		return errors.Errorf("can't create user[%v]@[%v] require ssl_type[%v]", user, host, ssltype)
	}

	if err := my.CreateUser(db, user, host, passwd, ssltype); err != nil {
		return errors.Errorf("create user[%v]@[%v] with all privileges require ssl_type[%v] failed: [%v]", user, host, ssltype, err)
	}

	query = fmt.Sprintf("GRANT %s ON *.* TO `%s`@`%s` WITH GRANT OPTION", strings.Join(mysqlAllPrivileges, ","), user, host)
	return my.grantPrivileges(db, query)
}

func (my *MysqlBase) grantPrivileges(db *sql.DB, query string) error {
	return Execute(db, query)
}
