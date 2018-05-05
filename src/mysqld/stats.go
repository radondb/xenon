/*
 * Xenon
 *
 * Copyright 2018 The Xenon Authors.
 * Code is licensed under the GPLv3.
 *
 */

package mysqld

import (
	"model"
	"sync/atomic"
)

// IncBackups used to increase the backup counter.
func (s *Backup) IncBackups() {
	atomic.AddUint64(&s.stats.Backups, 1)
}

// IncBackupErrs used to increase the backup error counter.
func (s *Backup) IncBackupErrs() {
	atomic.AddUint64(&s.stats.BackupErrs, 1)
}

// IncCancels used to increase the backup cancel counter.
func (s *Backup) IncCancels() {
	atomic.AddUint64(&s.stats.Cancels, 1)
}

// IncApplyLogs used to increase the apply counter.
func (s *Backup) IncApplyLogs() {
	atomic.AddUint64(&s.stats.AppLogs, 1)
}

// IncApplyLogErrs used to increase the apply error counter.
func (s *Backup) IncApplyLogErrs() {
	atomic.AddUint64(&s.stats.AppLogErrs, 1)
}

func (s *Backup) getStats() *model.BackupStats {
	return &model.BackupStats{
		Backups:    atomic.LoadUint64(&s.stats.Backups),
		BackupErrs: atomic.LoadUint64(&s.stats.BackupErrs),
		AppLogs:    atomic.LoadUint64(&s.stats.AppLogs),
		AppLogErrs: atomic.LoadUint64(&s.stats.AppLogErrs),
		Cancels:    atomic.LoadUint64(&s.stats.Cancels),
	}
}

// IncMysqldStarts used to increase the mysql start counter.
func (s *Mysqld) IncMysqldStarts() {
	atomic.AddUint64(&s.stats.MysqldStarts, 1)
}

// IncMysqldStops used to increase the mysql stop counter.
func (s *Mysqld) IncMysqldStops() {
	atomic.AddUint64(&s.stats.MysqldStops, 1)
}

// IncMonitorStarts used to increase the monitor start counter.
func (s *Mysqld) IncMonitorStarts() {
	atomic.AddUint64(&s.stats.MonitorStarts, 1)
}

// IncMonitorStops used to increase the monitor stop counter.
func (s *Mysqld) IncMonitorStops() {
	atomic.AddUint64(&s.stats.MonitorStops, 1)
}

func (s *Mysqld) getStats() *model.MysqldStats {
	return &model.MysqldStats{
		MysqldStarts:  atomic.LoadUint64(&s.stats.MysqldStarts),
		MysqldStops:   atomic.LoadUint64(&s.stats.MysqldStops),
		MonitorStarts: atomic.LoadUint64(&s.stats.MonitorStarts),
		MonitorStops:  atomic.LoadUint64(&s.stats.MonitorStops),
	}
}
