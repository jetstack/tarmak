// Copyright Jetstack Ltd. See LICENSE for details.
package main

import (
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/lambda"

	"github.com/jetstack/tarmak/cmd/tagging_control/cmd"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "zip" {

		err := cmd.Zip(os.Args[2:])
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}

		os.Exit(0)
	}

	lambda.Start(cmd.HandleRequest)
}
