// Copyright Jetstack Ltd. See LICENSE for details.
package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/jetstack/tarmak/pkg/terraform"
)

var internalPluginCmd = &cobra.Command{
	Use: "internal-plugin",
	Run: func(cmd *cobra.Command, args []string) {
		os.Exit(terraform.InternalPlugin(args))
	},
	Hidden:             true,
	DisableFlagParsing: true,
}

var terraformCmd = &cobra.Command{
	Use:                "terraform",
	Hidden:             true,
	DisableFlagParsing: true,
}

var terraformPlanCmd = &cobra.Command{
	Use: "plan",
	Run: func(cmd *cobra.Command, args []string) {
		os.Exit(terraform.Plan(args))
	},
	Hidden:             true,
	DisableFlagParsing: true,
}

var terraformApplyCmd = &cobra.Command{
	Use: "apply",
	Run: func(cmd *cobra.Command, args []string) {
		os.Exit(terraform.Apply(args))
	},
	Hidden:             true,
	DisableFlagParsing: true,
}

var terraformDestroyCmd = &cobra.Command{
	Use: "destroy",
	Run: func(cmd *cobra.Command, args []string) {
		os.Exit(terraform.Destroy(args))
	},
	Hidden:             true,
	DisableFlagParsing: true,
}

var terraformOutputCmd = &cobra.Command{
	Use: "output",
	Run: func(cmd *cobra.Command, args []string) {
		os.Exit(terraform.Output(args))
	},
	Hidden:             true,
	DisableFlagParsing: true,
}

var terraformInitCmd = &cobra.Command{
	Use: "init",
	Run: func(cmd *cobra.Command, args []string) {
		os.Exit(terraform.Init(args))
	},
	Hidden:             true,
	DisableFlagParsing: true,
}

var terraformForceUnlockCmd = &cobra.Command{
	Use: "force-unlock",
	Run: func(cmd *cobra.Command, args []string) {
		os.Exit(terraform.Unlock(args))
	},
	Hidden:             true,
	DisableFlagParsing: true,
}

func init() {
	RootCmd.AddCommand(internalPluginCmd)
	terraformCmd.AddCommand(terraformInitCmd)
	terraformCmd.AddCommand(terraformPlanCmd)
	terraformCmd.AddCommand(terraformApplyCmd)
	terraformCmd.AddCommand(terraformDestroyCmd)
	terraformCmd.AddCommand(terraformForceUnlockCmd)
	terraformCmd.AddCommand(terraformOutputCmd)
	RootCmd.AddCommand(terraformCmd)
}
