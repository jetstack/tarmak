// Copyright Jetstack Ltd. See LICENSE for details.
package cmd

import (
	"fmt"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/spf13/cobra"

	"github.com/jetstack/tarmak/pkg/tagging_control"
)

var handleCmd = &cobra.Command{
	Use:   "handle",
	Short: "Launch lambda request handler",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("running tagging_control lambda function request handler...\n")

		lambda.Start(tagging_control.HandleRequests)
	},
}

func init() {
	RootCmd.AddCommand(handleCmd)
}
