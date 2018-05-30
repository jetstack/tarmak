// Copyright Jetstack Ltd. See LICENSE for details.
package main

import (
	"os"

	"github.com/jetstack/tarmak/cmd/tarmak/cmd"
)

var (
	commit   = "unknown"
	date     = ""
	version  = "dev"
	wingHash = "unknown"
)

func main() {
	cmd.Version.Version = version
	cmd.Version.Commit = commit
	cmd.Version.BuildDate = date
	cmd.Version.WingHash = wingHash
	cmd.Execute(os.Args[1:])
}
