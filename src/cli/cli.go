/*
 * Xenon
 *
 * Copyright 2018 The Xenon Authors.
 * Code is licensed under the GPLv3.
 *
 */

package main

import (
	"cli/cmd"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const (
	cliName        = "xenoncli"
	cliDescription = "A simple command line client for xenon"
)

var (
	rootCmd = &cobra.Command{
		Use:        cliName,
		Short:      cliDescription,
		SuggestFor: []string{"xenonctl"},
	}
)

func init() {
	rootCmd.AddCommand(cmd.NewVersionCommand())
	rootCmd.AddCommand(cmd.NewInitCommand())
	rootCmd.AddCommand(cmd.NewClusterCommand())
	rootCmd.AddCommand(cmd.NewMysqlCommand())
	rootCmd.AddCommand(cmd.NewRaftCommand())
	rootCmd.AddCommand(cmd.NewXenonCommand())
	rootCmd.AddCommand(cmd.NewPerfCommand())
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
