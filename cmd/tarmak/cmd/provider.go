// Copyright Jetstack Ltd. See LICENSE for details.
package cmd

import (
	"github.com/spf13/cobra"
)

var providerCmd = &cobra.Command{
	Use:     "providers",
	Short:   "Operations on providers",
	Aliases: []string{"provider"},
}

func init() {
	RootCmd.AddCommand(providerCmd)
}
