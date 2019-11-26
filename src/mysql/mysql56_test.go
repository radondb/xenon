/*
 * Xenon
 *
 * Copyright 2018 The Xenon Authors.
 * Code is licensed under the GPLv3.
 *
 */

package mysql

import (
	"testing"

	"config"
	"xbase/xlog"

	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestMysql56Handler(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	conf := config.DefaultMysqlConfig()
	conf.Version = "mysql56"

	mysql := NewMysql(conf, log)
	want := new(Mysql56)
	got := mysql.mysqlHandler
	assert.Equal(t, want, got)
}

func TestMysql56SetSemiWaitSlaveCount(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer db.Close()

	queryList := []string{
		"SET GLOBAL rpl_semi_sync_master_wait_for_slave_count = 2",
	}

	mysql56 := new(Mysql56)
	mock.ExpectExec(queryList[0]).WillReturnResult(sqlmock.NewResult(1, 1))
	err = mysql56.SetSemiWaitSlaveCount(db, 2)
	assert.Nil(t, err)
}

func TestMysql56ChangeUserPassword(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer db.Close()

	queryList := []string{
		"SET PASSWORD FOR `usr`@'127.0.0.1' = PASSWORD('pwd')",
	}

	mysql56 := new(Mysql56)
	mock.ExpectExec(queryList[0]).WillReturnResult(sqlmock.NewResult(1, 1))
	err = mysql56.ChangeUserPasswd(db, "usr", "127.0.0.1", "pwd")
	assert.Nil(t, err)
}
