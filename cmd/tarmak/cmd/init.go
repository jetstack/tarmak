package cmd

import (
	"github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/jetstack/tarmak/pkg/tarmak"
)

var initCmd = &cobra.Command{
	Use:     "init",
	Aliases: []string{"initialise", "initialize"},
	Short:   "init a cluster configuration",
	Run: func(cmd *cobra.Command, args []string) {
		t := tarmak.New(cmd)
		err := t.Init()
		if err != nil {
			logrus.Fatal(err)
		}
	},
}

func init() {
	RootCmd.AddCommand(initCmd)
}
