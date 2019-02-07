// Copyright Jetstack Ltd. See LICENSE for details.
package cmd

import (
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
)

var environmentCmd = &cobra.Command{
	Use:     "environments",
	Short:   "Operations on environments",
	Aliases: []string{"environment"},
}

func environmentDestroyFlags(fs *flag.FlagSet) {
	store := &globalFlags.Environment.Destroy

	fs.BoolVar(
		&store.AutoApprove,
		"auto-approve",
		false,
		"auto-approve destroy of a complete environment",
	)
}

func init() {
	RootCmd.AddCommand(environmentCmd)
}
