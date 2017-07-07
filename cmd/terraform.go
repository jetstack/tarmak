package cmd

import (
	flag "github.com/spf13/pflag"
)

func terraformPFlags(fs *flag.FlagSet) {
	fs.StringSlice("terraform-stacks", []string{}, "terraform stacks to execute")
}
