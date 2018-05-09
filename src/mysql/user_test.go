/*
 * Xenon
 *
 * Copyright 2018 The Xenon Authors.
 * Code is licensed under the GPLv3.
 *
 */

package mysql

import (
	"config"
	"model"
	"testing"
	"xbase/xlog"

	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestCheckUserExists(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer db.Close()

	// log
	log := xlog.NewStdLog(xlog.Level(xlog.DEBUG))
	conf := config.DefaultMysqlConfig()
	mysql := NewMysql(conf, log)
	mysql.db = db

	query := "SELECT User FROM mysql.user WHERE User = 'xx'"
	columns := []string{"User"}
	want := true
	mockRows := sqlmock.NewRows(columns).AddRow(want)
	mock.ExpectQuery(query).WillReturnRows(mockRows)

	exists, err := mysql.CheckUserExists("xx")
	assert.Nil(t, err)

	got := exists
	assert.Equal(t, want, got)
}

func TestCreateUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer db.Close()

	// log
	log := xlog.NewStdLog(xlog.Level(xlog.DEBUG))
	conf := config.DefaultMysqlConfig()
	mysql := NewMysql(conf, log)
	mysql.db = db

	query := "CREATE USER `xx` IDENTIFIED BY 'xxx'"
	mock.ExpectExec(query).WillReturnResult(sqlmock.NewResult(1, 1))
	err = mysql.CreateUser("xx", "xxx")
	assert.Nil(t, err)
}

func TestCreateUserWithPrivileges(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer db.Close()

	// log
	log := xlog.NewStdLog(xlog.Level(xlog.DEBUG))
	conf := config.DefaultMysqlConfig()
	mysql := NewMysql(conf, log)
	mysql.db = db

	query := "GRANT ALTER , ALTER ROUTINE ON test.* TO `xx`@'%' IDENTIFIED BY 'pwd'"
	mock.ExpectExec(query).WillReturnResult(sqlmock.NewResult(1, 1))
	err = mysql.CreateUserWithPrivileges("xx", "pwd", "test", "*", "%", "ALTER , ALTER ROUTINE")
	assert.Nil(t, err)

}

func TestGetUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer db.Close()

	// log
	log := xlog.NewStdLog(xlog.Level(xlog.DEBUG))
	conf := config.DefaultMysqlConfig()
	mysql := NewMysql(conf, log)
	mysql.db = db

	query := "SELECT User, Host FROM mysql.user"
	columns := []string{"User", "Host"}
	want := []model.MysqlUser{
		{User: "user1",
			Host: "localhost"},
		{User: "root",
			Host: "localhost"},
	}
	mockRows := sqlmock.NewRows(columns).AddRow("user1", "localhost").AddRow("root", "localhost")
	mock.ExpectQuery(query).WillReturnRows(mockRows)

	got, err := mysql.GetUser()
	assert.Nil(t, err)

	assert.Equal(t, want, got)
}

func TestDropUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer db.Close()

	// log
	log := xlog.NewStdLog(xlog.Level(xlog.DEBUG))
	conf := config.DefaultMysqlConfig()
	mysql := NewMysql(conf, log)
	mysql.db = db

	query := "DROP USER `xx`"
	mock.ExpectExec(query).WillReturnResult(sqlmock.NewResult(1, 1))
	err = mysql.DropUser("xx")
	assert.Nil(t, err)
}

func TestCreateReplUserWithoutBinlog(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer db.Close()

	// log
	log := xlog.NewStdLog(xlog.Level(xlog.DEBUG))
	conf := config.DefaultMysqlConfig()
	mysql := NewMysql(conf, log)
	mysql.db = db

	queryList := []string{
		"SET sql_log_bin=0",
		"CREATE USER `repl` IDENTIFIED BY 'replpwd'",
		"GRANT REPLICATION SLAVE,REPLICATION CLIENT ON *.* TO `repl`",
		"SET sql_log_bin=1",
	}

	mock.ExpectExec(queryList[0]).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(queryList[1]).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(queryList[2]).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(queryList[3]).WillReturnResult(sqlmock.NewResult(1, 1))
	err = mysql.CreateReplUserWithoutBinlog("repl", "replpwd")
	assert.Nil(t, err)
}

func TestChangeUserPasswd(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer db.Close()

	// log
	log := xlog.NewStdLog(xlog.Level(xlog.DEBUG))
	conf := config.DefaultMysqlConfig()
	mysql := NewMysql(conf, log)
	mysql.db = db

	query := "ALTER USER `xx` IDENTIFIED BY 'xxx'"
	mock.ExpectExec(query).WillReturnResult(sqlmock.NewResult(1, 1))
	err = mysql.ChangeUserPasswd("xx", "xxx")
	assert.Nil(t, err)
}

func TestGrantAllPrivileges(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer db.Close()

	// log
	log := xlog.NewStdLog(xlog.Level(xlog.DEBUG))
	conf := config.DefaultMysqlConfig()
	mysql := NewMysql(conf, log)
	mysql.db = db

	query := "GRANT ALL ON *.* TO `xx` WITH GRANT OPTION"
	mock.ExpectExec(query).WillReturnResult(sqlmock.NewResult(1, 1))
	err = mysql.GrantAllPrivileges("xx")
	assert.Nil(t, err)
}

func TestGrantNormalPrivileges(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer db.Close()

	// log
	log := xlog.NewStdLog(xlog.Level(xlog.DEBUG))
	conf := config.DefaultMysqlConfig()
	mysql := NewMysql(conf, log)
	mysql.db = db

	query := "GRANT ALTER,ALTER ROUTINE,CREATE,CREATE ROUTINE,CREATE TEMPORARY TABLES,CREATE VIEW,DELETE,DROP,EXECUTE,EVENT,INDEX,INSERT,LOCK TABLES,PROCESS,RELOAD,SELECT,SHOW DATABASES,SHOW VIEW,UPDATE,TRIGGER,REFERENCES,REPLICATION SLAVE,REPLICATION CLIENT ON *.* TO `xx`"
	mock.ExpectExec(query).WillReturnResult(sqlmock.NewResult(1, 1))
	err = mysql.GrantNormalPrivileges("xx")
	assert.Nil(t, err)
}

func TestGrantReplicationPrivileges(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer db.Close()

	// log
	log := xlog.NewStdLog(xlog.Level(xlog.DEBUG))
	conf := config.DefaultMysqlConfig()
	mysql := NewMysql(conf, log)
	mysql.db = db

	query := "GRANT REPLICATION SLAVE,REPLICATION CLIENT ON *.* TO `xx`"
	mock.ExpectExec(query).WillReturnResult(sqlmock.NewResult(1, 1))
	err = mysql.GrantReplicationPrivileges("xx")
	assert.Nil(t, err)
}

func TestGrantUserPrivileges(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer db.Close()

	// log
	log := xlog.NewStdLog(xlog.Level(xlog.DEBUG))
	conf := config.DefaultMysqlConfig()
	mysql := NewMysql(conf, log)
	mysql.db = db

	// ok
	{
		user := "xx"
		passwd := "pwd"
		database := "db"
		table := "table1"
		host := "192.168.0.1"
		privs := "SELECT"
		query := "GRANT SELECT ON db.table1 TO `xx`@'192.168.0.1' IDENTIFIED BY 'pwd'"
		mock.ExpectExec(query).WillReturnResult(sqlmock.NewResult(1, 1))
		err = mysql.CreateUserWithPrivileges(user, passwd, database, table, host, privs)
		assert.Nil(t, err)
	}

	// ok
	{
		user := "xx"
		passwd := "pwd"
		database := "db"
		table := "table1"
		host := "192.168.0.1"
		privs := "SELECT,"
		query := "GRANT SELECT ON db.table1 TO `xx`@'192.168.0.1' IDENTIFIED BY 'pwd'"
		mock.ExpectExec(query).WillReturnResult(sqlmock.NewResult(1, 1))
		err = mysql.CreateUserWithPrivileges(user, passwd, database, table, host, privs)
		assert.Nil(t, err)
	}

	// error
	{
		user := "xx"
		passwd := "pwd"
		database := "db"
		table := "table1"
		host := "192.168.0.1"
		privs := "SELECT,GRANT OPTION"
		query := "GRANT SELECT ON db.table1 TO `xx`@'192.168.0.1' IDENTIFIED BY 'pwd'"
		mock.ExpectExec(query).WillReturnResult(sqlmock.NewResult(1, 1))
		err = mysql.CreateUserWithPrivileges(user, passwd, database, table, host, privs)
		want := "cant' create user[xx] with privileges[GRANT OPTION]"
		got := err.Error()
		assert.Equal(t, want, got)
	}
}
