// Copyright Jetstack Ltd. See LICENSE for details.
package cmd

import (
	"github.com/spf13/cobra"

	"github.com/jetstack/tarmak/pkg/tarmak"
	"github.com/jetstack/tarmak/pkg/tarmak/rpc"
)

var rpcCmd = &cobra.Command{
	Use:   "rpc",
	Short: "Run standalone RPC server",
	Run: func(cmd *cobra.Command, args []string) {
		t := tarmak.New(globalFlags)
		/*func(ctx context.Context) error {
			return t.CmdTerraformApply(args, ctx)
		},*/
		//t.Must(rpcServer.Start())
		rpc.Start(t)
	},
}

func init() {
	RootCmd.AddCommand(rpcCmd)
}
