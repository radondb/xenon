/*
 * Xenon
 *
 * Copyright 2018 The Xenon Authors.
 * Code is licensed under the GPLv3.
 *
 */

package cmd

import (
	"build"
	"fmt"

	"github.com/spf13/cobra"
)

func NewVersionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version number of xenon client",
		Run:   versionCommandFn,
	}

	return cmd
}

func versionCommandFn(cmd *cobra.Command, args []string) {
	build := build.GetInfo()
	fmt.Printf("xenoncli:[%+v]\n", build)
}
