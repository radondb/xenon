/*
 * Xenon
 *
 * Copyright 2018 The Xenon Authors.
 * Code is licensed under the GPLv3.
 *
 */

package raft

const (
	bash = "bash"
)

// leaderStartShellCommand execute the shell commands
// when leader start, such as START-VIP command
func (r *Raft) leaderStartShellCommand() error {
	args := []string{
		"-c",
		r.conf.LeaderStartCommand,
	}

	if out, err := r.cmd.RunCommand(bash, args); err != nil {
		r.ERROR("leaderStartShellCommand[%v].out[%v].error[%+v]", args, out, err)
		return err
	}
	r.WARNING("leaderStartShellCommand[%v].done", args)
	return nil
}

// leaderStopShellCommand executes the shell commands
// when leader stop, such as STOP-VIP command
func (r *Raft) leaderStopShellCommand() error {
	args := []string{
		"-c",
		r.conf.LeaderStopCommand,
	}

	if out, err := r.cmd.RunCommand(bash, args); err != nil {
		r.ERROR("leaderStopShellCommand[%v].out[%v].error[%+v]", args, out, err)
		return err
	}
	r.WARNING("leaderStopShellCommand[%v].done", args)
	return nil
}
