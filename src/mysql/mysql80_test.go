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
)

func TestMysql80Handler(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	conf := config.DefaultMysqlConfig()
	conf.Version = "mysql80"

	mysql := NewMysql(conf, 10000, log)
	want := new(Mysql80)
	want.SetQueryTimeout(10000)
	got := mysql.mysqlHandler
	assert.Equal(t, want, got)
}
