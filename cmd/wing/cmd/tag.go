// Copyright Jetstack Ltd. See LICENSE for details.
package cmd

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/jetstack/tarmak/pkg/wing/tags"
)

var tagCmd = &cobra.Command{
	Use:   "tag",
	Short: "Ensure public key tagging of instance",
	Run: func(cmd *cobra.Command, args []string) {
		log := logrus.New()
		log.SetLevel(logrus.DebugLevel)

		t, err := tags.New(logrus.NewEntry(log))
		if err != nil {
			log.Fatal(err)
		}

		if err := t.EnsureMachineTags(); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	RootCmd.AddCommand(tagCmd)
}
