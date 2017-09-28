package cmd

import (
	"github.com/spf13/cobra"
)

var clusterDebugCmd = &cobra.Command{
	Use:   "debug",
	Short: "used to debug clusters",
}

func init() {
	clusterCmd.AddCommand(clusterDebugCmd)
}
