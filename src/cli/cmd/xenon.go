/*
 * Xenon
 *
 * Copyright 2018 The Xenon Authors.
 * Code is licensed under the GPLv3.
 *
 */

package cmd

import (
	"fmt"

	"cli/callx"

	"github.com/spf13/cobra"
)

func NewXenonCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "xenon <subcommand>",
		Short: "xenon related commands",
	}

	cmd.AddCommand(NewXenonStatusCommand())

	return cmd
}

func NewXenonStatusCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ping",
		Short: "check node work or not",
		Run:   xenonPingCommandFn,
	}

	return cmd
}

func xenonPingCommandFn(cmd *cobra.Command, args []string) {
	if len(args) != 0 {
		ErrorOK(fmt.Errorf("too.many.args"))
	}

	// send ping to self
	{
		conf, err := GetConfig()
		ErrorOK(err)
		self := conf.Server.Endpoint
		rsp, err := callx.ServerPingRPC(self)
		ErrorOK(err)
		RspOK(rsp.RetCode)
	}
}
