// Copyright Jetstack Ltd. See LICENSE for details.
package cmd

import (
	"github.com/jetstack/tarmak/pkg/tarmak"
	rpcServer "github.com/jetstack/tarmak/pkg/tarmak/rpc"
	"github.com/spf13/cobra"
)

var rpc = &cobra.Command{
	Use:   "rpc",
	Short: "Run standalone RPC server",
	Run: func(cmd *cobra.Command, args []string) {
		t := tarmak.New(globalFlags)
		t.Must(rpcServer.Start())
	},
}

func init() {
	RootCmd.AddCommand(rpc)
}
