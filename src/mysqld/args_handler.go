/*
 * Xenon
 *
 * Copyright 2018 The Xenon Authors.
 * Code is licensed under the GPLv3.
 *
 */

package mysqld

// ArgsHandler for mysqld operation.
type ArgsHandler interface {
	// mysqld start args
	Start() []string

	// mysqld stop args
	Stop() []string

	// mysqld isrunning args
	IsRunning() []string

	// mysqld kill args
	Kill() []string
}
