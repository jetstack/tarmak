// Copyright Jetstack Ltd. See LICENSE for details.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	tarmakv1alpha1 "github.com/jetstack/tarmak/pkg/apis/tarmak/v1alpha1"
)

var globalFlags = &tarmakv1alpha1.Flags{}

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "tarmak",
	Short: "Tarmak is a toolkit for provisioning and managing Kubernetes clusters.",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	RootCmd.PersistentFlags().StringVarP(
		&globalFlags.ConfigDirectory,
		"config-directory",
		"c",
		"~/.tarmak",
		"config directory for tarmak's configuration",
	)

	RootCmd.PersistentFlags().BoolVarP(
		&globalFlags.Verbose,
		"verbose",
		"v",
		false,
		"enable verbose logging",
	)

	RootCmd.PersistentFlags().BoolVar(
		&globalFlags.KeepContainers,
		"keep-containers",
		false,
		"do not clean-up terraform/packer containers after running them",
	)

	/*RootCmd.PersistenFlags().BoolVar(
		&globalFlags.RPC,
		"rpc",
		flase,
		"start RPC server"
	)*/
}
