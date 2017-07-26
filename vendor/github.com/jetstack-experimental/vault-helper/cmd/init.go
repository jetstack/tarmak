package cmd

import (
	"github.com/Sirupsen/logrus"
	//"github.com/hashicorp/vault/api"
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var k8sDevServerCmd = &cobra.Command{
	Use:   "dev-server",
	Short: "Run a vault server in development mode with kubernetes PKI created",
	Run: func(cmd *cobra.Command, args []string) {
		logrus.Fatal("dev-server unimplemented")
	},
}

func init() {
	RootCmd.AddCommand(k8sDevServerCmd)
}
