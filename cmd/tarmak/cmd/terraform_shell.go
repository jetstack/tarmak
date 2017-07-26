package cmd

import (
	"github.com/spf13/cobra"

	"github.com/jetstack/tarmak/pkg/tarmak"
)

var terraformShellCmd = &cobra.Command{
	Use:     "terraform-shell",
	Aliases: []string{"t-s"},
	Short:   "This prepare a terraform container and executes a shell in this context",
	Run: func(cmd *cobra.Command, args []string) {
		t := tarmak.New(cmd)
		t.Must(t.CmdTerraformShell(args))
	},
}

func init() {
	RootCmd.AddCommand(terraformShellCmd)
}
