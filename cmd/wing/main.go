package main

import (
	"github.com/jetstack/tarmak/cmd/wing/cmd"
)

var (
	version string = "dev"
	commit  string = "unknown"
	date    string = ""
)

func main() {
	cmd.Version.Version = version
	cmd.Version.Commit = commit
	cmd.Version.BuildDate = date
	cmd.Execute()
}
