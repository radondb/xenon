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
func (my *Mysql56) ChangeUserPasswd(db *sql.DB, user string, passwd string) error {
	query := fmt.Sprintf("SET PASSWORD FOR `%s` = PASSWORD('%s')", user, passwd)
	return Execute(db, query)
}
