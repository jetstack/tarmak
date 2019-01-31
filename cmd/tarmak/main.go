// Copyright Jetstack Ltd. See LICENSE for details.
package main

import (
	"os"

	"github.com/jetstack/tarmak/cmd/tarmak/cmd"
)

func main() {
	cmd.Execute(os.Args[1:])
}
