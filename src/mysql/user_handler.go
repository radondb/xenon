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

// UserHandler interface.
type UserHandler interface {
	GetUser(*sql.DB) ([]model.MysqlUser, error)
	CheckUserExists(*sql.DB, string) (bool, error)
	CreateUser(*sql.DB, string, string) error
	DropUser(*sql.DB, string, string) error
	ChangeUserPasswd(*sql.DB, string, string) error
	CreateReplUserWithoutBinlog(*sql.DB, string, string) error
	GrantAllPrivileges(*sql.DB, string) error
	GrantNormalPrivileges(*sql.DB, string) error
	CreateUserWithPrivileges(db *sql.DB, user, passwd, database, table, host, privs string) error
	GrantReplicationPrivileges(*sql.DB, string) error
}
