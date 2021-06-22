/*
 * Xenon
 *
 * Copyright 2018 The Xenon Authors.
 * Code is licensed under the GPLv3.
 *
 */

package cmd

import (
	"encoding/json"
	"fmt"

	"xbase/common"
	"xbase/xlog"

	"github.com/spf13/cobra"
)

func quickStack(log *xlog.Log) (string, error) {
	timeout := 10 * 1000 // 10s
	cmds := "bash"
	args := []string{
		"-c",
		"sudo quickstack -s -k 10 -p `pidof mysqld`",
	}

	cmd := common.NewLinuxCommand(log)
	return cmd.RunCommandWithTimeout(timeout, cmds, args)
}

func NewPerfCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "perf <subcommand>",
		Short: "perf related commands",
	}

	cmd.AddCommand(NewQuickStackCommand())

	return cmd
}

func NewQuickStackCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "quickstack",
		Short: "capture the stack of mysqld using quickstack",
		Run:   quickStackCommandFn,
	}
	cmd.AddCommand(NewQuickStackJsonCommand())

	return cmd
}

func quickStackCommandFn(cmd *cobra.Command, args []string) {
	outs, err := quickStack(log)
	ErrorOK(err)
	fmt.Printf("%v", outs)
}

func NewQuickStackJsonCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "json",
		Short: "json format",
		Run:   quickStackJsonCommandFn,
	}

	return cmd
}

func quickStackJsonCommandFn(cmd *cobra.Command, args []string) {
	type Status struct {
		Results string `json:"status"`
	}
	status := &Status{}
	outs, err := quickStack(log)
	ErrorOK(err)
	status.Results = outs

	statusB, _ := json.Marshal(status)
	fmt.Printf("%s", string(statusB))
}
