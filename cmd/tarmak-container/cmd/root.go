// Copyright Jetstack Ltd. See LICENSE for details.
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var flagVerbose bool

var subCommands []*cobra.Command

// RootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "tarmak-conntainer",
	Short: "Tarmak is a toolkit for provisioning and managing Kubernetes clusters.",
}

// execute a subcommand directly for matching basenames, similar to hyperkube
func Execute() {
	cmd := rootCmd

	basename := filepath.Base(os.Args[0])

	if basename == "tarmak-connector" {
		cmd = connectorCmd
	} else if basename == "terraform-provider-awstag" {
		cmd = providerAWSTagCmd
	} else if basename == "terraform-provider-tarmak" {
		cmd = providerTarmakCmd
	} else {
		cmd.AddCommand(subCommands...)
	}

	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func newLogger() *logrus.Entry {
	log := logrus.New()
	if flagVerbose {
		log.SetLevel(logrus.DebugLevel)
	} else {
		log.SetLevel(logrus.InfoLevel)
	}
	return log.WithField("app", "tarmak-container")
}
