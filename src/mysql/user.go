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
)

// User tuple.
type User struct {
	UserHandler
}

// CheckUserExists used to check the user exists or not.
func (u *User) CheckUserExists(db *sql.DB, user string) (bool, error) {
	query := fmt.Sprintf("SELECT User FROM mysql.user WHERE User = '%s'", user)
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
func (u *User) GetUser(db *sql.DB) ([]model.MysqlUser, error) {
	query := fmt.Sprintf("SELECT User, Host FROM mysql.user")
	rows, err := Query(db, query)
	if err != nil {
		return nil, err
	}

	var Users = make([]model.MysqlUser, len(rows))

	for i, v := range rows {
		Users[i].User = v["User"]
		Users[i].Host = v["Host"]
	}
	return Users, nil
}

// CreateUser use to create new user.
// see http://dev.mysql.com/doc/refman/5.7/en/string-literals.html
func (u *User) CreateUser(db *sql.DB, user string, passwd string) error {
	query := fmt.Sprintf("CREATE USER `%s` IDENTIFIED BY '%s'", user, passwd)
	return Execute(db, query)
}

// DropUser used to drop the user.
func (u *User) DropUser(db *sql.DB, user string, host string) error {
	query := fmt.Sprintf("DROP USER `%s`@'%s'", user, host)
	return Execute(db, query)
}

// CreateReplUserWithoutBinlog create replication accounts without writing binlog.
func (u *User) CreateReplUserWithoutBinlog(db *sql.DB, user string, passwd string) error {
	queryList := []string{
		"SET sql_log_bin=0",
		fmt.Sprintf("CREATE USER `%s` IDENTIFIED BY '%s'", user, passwd),
		fmt.Sprintf("GRANT %s ON *.* TO `%s`", strings.Join(mysqlReplPrivileges, ","), user),
		"SET sql_log_bin=1",
	}
	return ExecuteSuperQueryList(db, queryList)
}

// ChangeUserPasswd used to change the user password.
func (u *User) ChangeUserPasswd(db *sql.DB, user string, passwd string) error {
	query := fmt.Sprintf("ALTER USER `%s` IDENTIFIED BY '%s'", user, passwd)
	return Execute(db, query)
}

// GrantNormalPrivileges used to grants normal privileges.
func (u *User) GrantNormalPrivileges(db *sql.DB, user string) error {
	query := fmt.Sprintf("GRANT %s ON *.* TO `%s`", strings.Join(mysqlNormalPrivileges, ","), user)
	return u.grantPrivileges(db, query)
}

// CreateUserWithPrivileges for create normal user.
func (u *User) CreateUserWithPrivileges(db *sql.DB, user, passwd, database, table, host, privs string, ssl string) error {
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
			return errors.Errorf("cant' create user[%v] with privileges[%v]", user, priv)
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
		return errors.Errorf("cant' create user[%v] require ssl_type[%v]", user, ssltype)
	}
	if ssltype == "YES" {
		query = fmt.Sprintf("GRANT %s ON %s.%s TO `%s`@'%s' IDENTIFIED BY '%s' REQUIRE X509", privs, database, table, user, host, passwd)

	} else {
		query = fmt.Sprintf("GRANT %s ON %s.%s TO `%s`@'%s' IDENTIFIED BY '%s'", privs, database, table, user, host, passwd)
	}

	return u.grantPrivileges(db, query)
}

// GrantReplicationPrivileges used to grant repli privis.
func (u *User) GrantReplicationPrivileges(db *sql.DB, user string) error {
	query := fmt.Sprintf("GRANT %s ON *.* TO `%s`", strings.Join(mysqlReplPrivileges, ","), user)
	return u.grantPrivileges(db, query)
}

// GrantAllPrivileges used to grant all privis.
func (u *User) GrantAllPrivileges(db *sql.DB, user string) error {
	query := fmt.Sprintf("GRANT %s ON *.* TO `%s` WITH GRANT OPTION", strings.Join(mysqlAllPrivileges, ","), user)
	return u.grantPrivileges(db, query)
}

func (u *User) grantPrivileges(db *sql.DB, query string) error {
	return Execute(db, query)
}
