package cmd

import (
	"github.com/spf13/cobra"
)

var clusterCmd = &cobra.Command{
	Use:     "clusters",
	Short:   "operations on clusters",
	Aliases: []string{"cluster"},
}

func init() {
	RootCmd.AddCommand(clusterCmd)
}
