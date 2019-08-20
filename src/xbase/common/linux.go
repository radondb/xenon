/*
 * Xenon
 *
 * Copyright 2018 The Xenon Authors.
 * Code is licensed under the GPLv3.
 *
 */

package common

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"os/exec"
	"strings"
	"sync/atomic"
	"time"
	"xbase/xlog"
)

func RunCommand(cmds string, args ...string) (string, error) {
	cmd := exec.Command(cmds, args...)
	outs, err := cmd.CombinedOutput()
	if err != nil {
		return strings.Join(cmd.Args, " ") + string(outs), err
	}
	return string(outs), nil
}

func runCommandWithTimeout(log *xlog.Log, timeout int, cmds string, args ...string) (out string, err error) {
	const tmpl = `Stdout: %v, Stderr: %v, Error: %v`

	cmdStr := cmds + " " + strings.Join(args, " ")
	log.Warning(fmt.Sprintf("==> Executing: %s", cmdStr))

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	cmd := exec.Command(cmds, args...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	if err = cmd.Start(); err != nil {
		err = fmt.Errorf(tmpl, stdout.String(), stderr.String(), err)
		return
	}

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()
	select {
	case err = <-done:
	case <-time.After(time.Duration(timeout) * time.Millisecond):
		cmd.Process.Kill()
		err = errors.Errorf("cmd[%v].exec.timeout[%v]", cmdStr, timeout)
	}

	if err != nil {
		return "", fmt.Errorf(tmpl, stdout.String(), stderr.String(), err)
	}
	return stdout.String(), nil
}

// a warapper of command
type LinuxCommand struct {
	log    *xlog.Log
	cmd    *exec.Cmd
	stderr io.ReadCloser
	stdout io.ReadCloser
}

func NewLinuxCommand(log *xlog.Log) Command {
	return &LinuxCommand{
		log: log,
	}
}

func (c *LinuxCommand) Run(cmds string, args []string) error {
	cmd := exec.Command(cmds, args...)

	c.log.Warning("LinuxCommand.prepare.to.run.cmds[%v]", strings.Join(args, " "))
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return errors.WithStack(err)
	}
	c.stderr = stderr

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return errors.WithStack(err)
	}
	c.stdout = stdout

	if err = cmd.Start(); err != nil {
		return errors.WithStack(err)
	}
	c.cmd = cmd
	return nil
}

// Wait cmd done

func (c *LinuxCommand) Await() error {
	if err := c.cmd.Wait(); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// scan the substr until times reached or io.EOF got
func (c *LinuxCommand) Scan(substr string, times int) error {
	var founds int32
	log := c.log
	go func() {
		scanner := bufio.NewScanner(c.stdout)
		for scanner.Scan() {
			text := scanner.Text()
			log.Warning("LinuxCommand.STDOUT==>%v", text)
			if strings.Contains(text, substr) {
				atomic.AddInt32(&founds, 1)
			}
		}

		if err := scanner.Err(); err != nil {
			log.Error("LinuxCommand.stdout.scanner.error:%+v", err)
		}
	}()

	go func() {
		scanner := bufio.NewScanner(c.stderr)
		for scanner.Scan() {
			text := scanner.Text()
			c.log.Warning("LinuxCommand.STDERR==>%v", text)
			if strings.Contains(text, substr) {
				atomic.AddInt32(&founds, 1)
			}
		}

		if err := scanner.Err(); err != nil {
			log.Error("LinuxCommand.stderr.scanner.error:%+v", err)
		}
	}()

	// wait the cmd to finish.
	c.cmd.Wait()
	if int(founds) != times {
		return errors.Errorf("cmd.outs.[%v].found[%v]!=expects[%v]", substr, founds, times)
	}
	return nil
}

func (c *LinuxCommand) Kill() error {
	if c.cmd != nil {
		return c.cmd.Process.Kill()
	}
	return nil
}

func (c *LinuxCommand) RunCommand(cmds string, args []string) (string, error) {
	return RunCommand(cmds, args...)
}

func (c *LinuxCommand) RunCommandWithTimeout(timeout int, cmds string, args []string) (string, error) {
	return runCommandWithTimeout(c.log, timeout, cmds, args...)
}
