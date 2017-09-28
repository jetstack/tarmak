package cmd

import (
	"github.com/spf13/cobra"
)

var clusterInstancesCmd = &cobra.Command{
	Use:     "instances",
	Short:   "operations on instances",
	Aliases: []string{"instance"},
}

func init() {
	clusterCmd.AddCommand(clusterInstancesCmd)
}
