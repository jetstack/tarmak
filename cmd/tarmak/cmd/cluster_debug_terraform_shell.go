package cmd

import (
	"github.com/spf13/cobra"

	"github.com/jetstack/tarmak/pkg/tarmak"
)

var clusterDebugTerraformShellCmd = &cobra.Command{
	Use:   "shell",
	Short: "Prepares a terraform container and executes a shell in this cluster",
	Run: func(cmd *cobra.Command, args []string) {
		t := tarmak.New(cmd)
		t.Must(t.CmdTerraformShell(args))
	},
}

func init() {
	clusterDebugTerraformCmd.AddCommand(clusterDebugTerraformShellCmd)
}
