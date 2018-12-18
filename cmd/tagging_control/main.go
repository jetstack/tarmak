// Copyright Jetstack Ltd. See LICENSE for details.
package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/jetstack/tarmak/cmd/tagging_control/cmd"
)

func main() {
	lambda.Start(cmd.HandleRequest)
}
