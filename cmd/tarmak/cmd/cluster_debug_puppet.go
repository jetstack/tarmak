package cmd

import (
	"github.com/spf13/cobra"
)

var clusterDebugPuppetCmd = &cobra.Command{
	Use:   "puppet",
	Short: "puppet debug operations on cluster",
}

func init() {
	clusterDebugCmd.AddCommand(clusterDebugPuppetCmd)
}
