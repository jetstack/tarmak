package main

import (
	"flag"

	"github.com/jetstack-experimental/vault-unsealer/cmd"
)

var (
	version string = "dev"
	commit  string = "unknown"
	date    string = "unknown"
)

func main() {
	flag.Parse()
	cmd.Version.Version = version
	cmd.Version.Commit = commit
	cmd.Version.BuildDate = date
	cmd.Execute()
}
