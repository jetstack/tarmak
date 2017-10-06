package cmd

import (
	"github.com/spf13/cobra"
)

var environmentCmd = &cobra.Command{
	Use:     "environments",
	Short:   "Operations on environments",
	Aliases: []string{"environment"},
}

func init() {
	RootCmd.AddCommand(environmentCmd)
}
