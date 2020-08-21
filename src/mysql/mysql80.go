/*
 * Xenon
 *
 * Copyright 2018-2019 The Xenon Authors.
 * Code is licensed under the GPLv3.
 *
 */

package mysql

var (
	_ MysqlHandler = &Mysql80{}
)

// Mysql80 tuple.
type Mysql80 struct {
	MysqlBase
}
