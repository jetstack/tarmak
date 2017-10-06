// Copyright Jetstack Ltd. See LICENSE for details.
package cmd

import (
	"github.com/spf13/cobra"
)

var environmentCmd = &cobra.Command{
	Use:     "environments",
	Short:   "operations on environments",
	Aliases: []string{"environment"},
}

func init() {
	RootCmd.AddCommand(environmentCmd)
}
