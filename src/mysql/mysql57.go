/*
 * Xenon
 *
 * Copyright 2018-2019 The Xenon Authors.
 * Code is licensed under the GPLv3.
 *
 */

package mysql

var (
	_ MysqlHandler = &Mysql57{}
)

// Mysql57 tuple.
type Mysql57 struct {
	MysqlBase
}
