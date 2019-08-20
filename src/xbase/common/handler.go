/*
 * Xenon
 *
 * Copyright 2018 The Xenon Authors.
 * Code is licensed under the GPLv3.
 *
 */

package common

type Command interface {
	Run(string, []string) error
	Scan(string, int) error
	Kill() error
	Await() error
	RunCommand(string, []string) (string, error)
	RunCommandWithTimeout(int, string, []string) (string, error)
}
