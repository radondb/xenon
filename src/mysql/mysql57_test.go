/*
 * Xenon
 *
 * Copyright 2018 The Xenon Authors.
 * Code is licensed under the GPLv3.
 *
 */

package mysql

import (
	"model"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

var mysql57 = new(Mysql57)

func TestMysql57Ping(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer db.Close()

	query := "SHOW SLAVE STATUS"
	columns := []string{"Master_Log_File",
		"Read_Master_Log_Pos",
		"Relay_Master_Log_File",
	}
	mockRows := sqlmock.NewRows(columns).AddRow("mysql-bin.000001",
		"147",
		"mysql-bin.000001",
	)

	mock.ExpectQuery(query).WillReturnRows(mockRows)
	pe, err := mysql57.Ping(db)
	assert.Nil(t, err)

	want := "mysql-bin.000001"
	got := pe.Relay_Master_Log_File
	assert.Equal(t, want, got)
}

func TestMysql57GetSlaveGTIDGotZeroRow(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer db.Close()

	query := "SHOW SLAVE STATUS"
	columns := []string{"Master_Log_File",
		"Read_Master_Log_Pos",
		"Retrieved_Gtid_Set",
		"Executed_Gtid_Set",
		"Slave_IO_Running",
		"Slave_SQL_Running",
		"Seconds_Behind_Master",
		"Last_Error",
		"Slave_SQL_Running_State",
	}

	want := model.GTID{}
	mockRows := sqlmock.NewRows(columns).AddRow("", "", "", "", "", "", "", "", "")
	mock.ExpectQuery(query).WillReturnRows(mockRows)
	got, err := mysql57.GetSlaveGTID(db)
	assert.Nil(t, err)
	assert.Equal(t, want, *got)
}

func TestMysql57GetSlaveGTID(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer db.Close()

	query := "SHOW SLAVE STATUS"
	columns := []string{"Master_Log_File",
		"Read_Master_Log_Pos",
		"Retrieved_Gtid_Set",
		"Executed_Gtid_Set",
		"Slave_IO_Running",
		"Slave_SQL_Running",
		"Seconds_Behind_Master",
		"Last_Error",
		"Slave_SQL_Running_State",
	}

	want := model.GTID{Master_Log_File: "mysql-bin.000001",
		Read_Master_Log_Pos:     147,
		Retrieved_GTID_Set:      "84030605-66aa-11e6-9465-52540e7fd51c:154-160",
		Executed_GTID_Set:       "84030605-66aa-11e6-9465-52540e7fd51c:1-159,ebd03dad-69ad-11e6-aa22-52540e7fd51c:1",
		Slave_IO_Running:        true,
		Slave_IO_Running_Str:    "Yes",
		Slave_SQL_Running:       true,
		Slave_SQL_Running_Str:   "Yes",
		Seconds_Behind_Master:   "11",
		Last_Error:              "",
		Slave_SQL_Running_State: "Slave has read all relay log; waiting for the slave I/O thread to update it",
	}

	mockRows := sqlmock.NewRows(columns).AddRow("mysql-bin.000001",
		"147",
		"84030605-66aa-11e6-9465-52540e7fd51c:154-160",
		"84030605-66aa-11e6-9465-52540e7fd51c:1-159,ebd03dad-69ad-11e6-aa22-52540e7fd51c:1",
		"Yes",
		"Yes",
		"11",
		"",
		"Slave has read all relay log; waiting for the slave I/O thread to update it",
	)

	mock.ExpectQuery(query).WillReturnRows(mockRows)
	got, err := mysql57.GetSlaveGTID(db)
	assert.Nil(t, err)
	assert.Equal(t, want, *got)
}

func TestMysql57GetMasterGTID(t *testing.T) {
	// 1. mysql up
	{
		db, mock, err := sqlmock.New()
		assert.Nil(t, err)
		defer db.Close()

		query := "SHOW MASTER STATUS"
		columns := []string{"File",
			"Position",
			"Binlog_Do_DB",
			"Binlog_Ignore_DB",
			"Executed_Gtid_Set",
		}

		want := model.GTID{Master_Log_File: "mysql-bin.000001",
			Read_Master_Log_Pos:     147,
			Retrieved_GTID_Set:      "",
			Executed_GTID_Set:       "84030605-66aa-11e6-9465-52540e7fd51c:154-160",
			Slave_IO_Running:        true,
			Slave_SQL_Running:       true,
			Seconds_Behind_Master:   "0",
			Last_Error:              "",
			Slave_SQL_Running_State: "",
		}

		mockRows := sqlmock.NewRows(columns).AddRow("mysql-bin.000001",
			"147",
			"",
			"",
			"84030605-66aa-11e6-9465-52540e7fd51c:154-160",
		)

		mock.ExpectQuery(query).WillReturnRows(mockRows)
		got, err := mysql57.GetMasterGTID(db)
		assert.Nil(t, err)
		assert.Equal(t, want, *got)
	}
}

func TestMysql57ChangeMasterToCommand(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.Nil(t, err)
	defer db.Close()

	want := []string{
		`CHANGE MASTER TO
  MASTER_HOST = 'localhost',
  MASTER_PORT = 123,
  MASTER_USER = 'username',
  MASTER_PASSWORD = 'password',
  MASTER_AUTO_POSITION = 1 FOR CHANNEL ''`}

	master := model.Repl{Master_Host: "localhost",
		Master_Port:   123,
		Repl_User:     "username",
		Repl_Password: "password"}

	got := mysql57.changeMasterToCommands(&master)
	assert.Equal(t, want, got)
}

func TestMysql57ChangeMasterTo(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer db.Close()

	queryList := []string{"STOP SLAVE",
		"CHANGE REPLICATION FILTER REPLICATE_DO_DB=(),REPLICATE_IGNORE_DB=(),REPLICATE_DO_TABLE=(),REPLICATE_IGNORE_TABLE=(),REPLICATE_WILD_DO_TABLE=(),REPLICATE_WILD_IGNORE_TABLE=(),REPLICATE_REWRITE_DB=()",
		`CHANGE MASTER TO MASTER_HOST = 'localhost', MASTER_PORT = 123, MASTER_USER = 'username', MASTER_PASSWORD = 'password', MASTER_AUTO_POSITION = 1 FOR CHANNEL ''`,
		"START SLAVE FOR CHANNEL ''",
	}

	mock.ExpectExec(queryList[0]).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(queryList[1]).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(queryList[2]).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(queryList[3]).WillReturnResult(sqlmock.NewResult(1, 1))

	master := model.Repl{Master_Host: "localhost",
		Master_Port:   123,
		Repl_User:     "username",
		Repl_Password: "password"}
	err = mysql57.ChangeMasterTo(db, &master)
	assert.Nil(t, err)
}

func TestMysql57ChangeToMaster(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer db.Close()

	queryList := []string{"STOP SLAVE",
		"RESET SLAVE ALL",
	}

	mock.ExpectExec(queryList[0]).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(queryList[1]).WillReturnResult(sqlmock.NewResult(1, 1))
	err = mysql57.ChangeToMaster(db)
	assert.Nil(t, err)
}

func TestMysql57SlaveIOThread(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer db.Close()

	queryList := []string{
		"START SLAVE IO_THREAD FOR CHANNEL ''",
		"STOP SLAVE IO_THREAD FOR CHANNEL ''",
	}

	mock.ExpectExec(queryList[0]).WillReturnResult(sqlmock.NewResult(1, 1))
	err = mysql57.StartSlaveIOThread(db)
	assert.Nil(t, err)

	mock.ExpectExec(queryList[1]).WillReturnResult(sqlmock.NewResult(1, 1))
	err = mysql57.StopSlaveIOThread(db)
	assert.Nil(t, err)
}

func TestMysql57ReadOnly(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer db.Close()

	queryList := []string{
		"SET GLOBAL read_only = 1",
		"SET GLOBAL super_read_only = 1",
		"SET GLOBAL read_only = 0",
		"SET GLOBAL super_read_only = 0",
	}

	mock.ExpectExec(queryList[0]).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(queryList[1]).WillReturnResult(sqlmock.NewResult(1, 1))
	err = mysql57.SetReadOnly(db, true)
	assert.Nil(t, err)

	mock.ExpectExec(queryList[2]).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(queryList[3]).WillReturnResult(sqlmock.NewResult(1, 1))
	err = mysql57.SetReadOnly(db, false)
	assert.Nil(t, err)
}

func TestMysql57SetGlobalVar(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer db.Close()

	queryList := []string{
		"SET GLOBAL read_only = 1",
		"SET GLOBAL gtid_mode = 'ON'",
		"XET GLOBAL gtid_mode = 'ON'",
	}

	mock.ExpectExec(queryList[0]).WillReturnResult(sqlmock.NewResult(1, 1))
	err = mysql57.SetGlobalSysVar(db, queryList[0])
	assert.Nil(t, err)

	mock.ExpectExec(queryList[1]).WillReturnResult(sqlmock.NewResult(1, 1))
	err = mysql57.SetGlobalSysVar(db, queryList[1])
	assert.Nil(t, err)

	err = mysql57.SetGlobalSysVar(db, queryList[2])
	want := "[XET GLOBAL gtid_mode = 'ON'].must.be.startwith:SET GLOBAL"
	got := err.Error()
	assert.Equal(t, want, got)
}

func TestMysql57ResetMaster(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer db.Close()

	queryList := []string{
		"RESET MASTER",
	}

	mock.ExpectExec(queryList[0]).WillReturnResult(sqlmock.NewResult(1, 1))
	err = mysql57.ResetMaster(db)
	assert.Nil(t, err)
}

func TestMysql57PurgeBinlogsTo(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer db.Close()

	queryList := []string{
		"PURGE BINARY LOGS TO 'mysql-bin.000032'",
	}

	mock.ExpectExec(queryList[0]).WillReturnResult(sqlmock.NewResult(1, 1))
	err = mysql57.PurgeBinlogsTo(db, "mysql-bin.000032")
	assert.Nil(t, err)
}

func TestMysql57SemiMaster(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer db.Close()

	queryList := []string{
		"SET GLOBAL rpl_semi_sync_master_enabled=ON",
		"SET GLOBAL rpl_semi_sync_master_enabled=OFF",
	}

	mock.ExpectExec(queryList[0]).WillReturnResult(sqlmock.NewResult(1, 1))
	err = mysql57.EnableSemiSyncMaster(db)
	assert.Nil(t, err)

	mock.ExpectExec(queryList[1]).WillReturnResult(sqlmock.NewResult(1, 1))
	err = mysql57.DisableSemiSyncMaster(db)
	assert.Nil(t, err)
}
