/*
 * Xenon
 *
 * Copyright 2018 The Xenon Authors.
 * Code is licensed under the GPLv3.
 *
 */

package mysql

import (
	"fmt"
)

func (m *Mysql) setState(state State) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.state = state
}

func (m *Mysql) getState() State {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.state
}

func (m *Mysql) setOption(o Option) {
	m.option = o
}

func (m *Mysql) getOption() Option {
	return m.option
}

func (m *Mysql) getConnStr() string {
	return fmt.Sprintf("%s:%d", m.conf.Host, m.conf.Port)
}
