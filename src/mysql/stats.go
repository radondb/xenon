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
	"sync/atomic"
)

// IncMysqlDowns used to increase the mysql down counter.
func (s *Mysql) IncMysqlDowns() {
	atomic.AddUint64(&s.stats.MysqlDowns, 1)
}

func (s *Mysql) getStats() *model.MysqlStats {
	return &model.MysqlStats{
		MysqlDowns: atomic.LoadUint64(&s.stats.MysqlDowns),
	}
}
