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
	"strings"

	"github.com/pkg/errors"
)

var (
	_ MysqlHandler = &Mysql56{}
)

// Mysql56 tuple.
type Mysql56 struct {
	MysqlBase
}

//SetSemiWaitSlaveCount used set rpl_semi_sync_master_wait_for_slave_count
func (my *Mysql56) SetSemiWaitSlaveCount(db *sql.DB, count int) error {
	return nil
}

// ChangeUserPasswd used to change the user password.
func (my *Mysql56) ChangeUserPasswd(db *sql.DB, user string, host string, passwd string) error {
	query := fmt.Sprintf("SET PASSWORD FOR `%s`@`%s` = PASSWORD('%s')", user, host, passwd)
	return Execute(db, query)
}

// CreateUser use to create new user.
func (my *Mysql56) CreateUser(db *sql.DB, user string, host string, passwd string, ssltype string) error {
	query := fmt.Sprintf("GRANT USAGE ON *.* TO `%s`@`%s` IDENTIFIED BY '%s'", user, host, passwd)
	if ssltype == "YES" {
		query = fmt.Sprintf("%s REQUIRE SSL", query)
	}
	return Execute(db, query)
}

// CreateUserWithPrivileges for create normal user.
func (my *Mysql56) CreateUserWithPrivileges(db *sql.DB, user, passwd, database, table, host, privs string, ssl string) error {
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
			return errors.Errorf("can't create user[%v]@[%v] with privileges[%v]", user, host, priv)
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
		return errors.Errorf("can't create user[%v]@[%v] require ssl_type[%v]", user, host, ssltype)
	}

	if err := my.CreateUser(db, user, host, passwd, ssltype); err != nil {
		return errors.Errorf("create user[%v]@[%v] with privileges[%v] require ssl_type[%v] failed: [%v]", user, host, privs, ssltype, err)
	}

	query = fmt.Sprintf("GRANT %s ON %s.%s TO `%s`@`%s`", privs, database, table, user, host)
	return my.grantPrivileges(db, query)
}

// GrantAllPrivileges used to grant all privis.
func (my *Mysql56) GrantAllPrivileges(db *sql.DB, user string, host string, passwd string, ssl string) error {
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

func (my *Mysql56) grantPrivileges(db *sql.DB, query string) error {
	return Execute(db, query)
}
