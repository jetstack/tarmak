// Copyright Jetstack Ltd. See LICENSE for details.
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	tarmakv1alpha1 "github.com/jetstack/tarmak/pkg/apis/tarmak/v1alpha1"
	"github.com/jetstack/tarmak/pkg/tarmak/utils"
	"github.com/jetstack/tarmak/pkg/tarmak/utils/consts"
	"github.com/jetstack/tarmak/pkg/terraform"
)

var globalFlags = &tarmakv1alpha1.Flags{}

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:               "tarmak",
	Short:             "Tarmak is a toolkit for provisioning and managing Kubernetes clusters.",
	DisableAutoGenTag: true,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(args []string) {
	// take build version and publish it to tarmak
	globalFlags.Version = Version.Version

	// become a terraform provider if required
	prefix := "terraform-provider-"
	basename := filepath.Base(os.Args[0])
	if strings.HasPrefix(basename, prefix) && len(prefix) < len(basename) {
		providerName := strings.TrimPrefix(basename, prefix)
		if _, ok := terraform.InternalProviders[providerName]; ok {
			os.Exit(terraform.InternalPlugin([]string{"provider", providerName}))
		}
	}

	cmd, _, err := RootCmd.Traverse(args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for _, c := range []struct {
		use  string
		name string
	}{
		{use: clusterKubectlCmd.Use, name: "kubectl"},
		{use: clusterSshCmd.Use, name: "ssh"},
	} {
		if cmd.Use != c.use {
			continue
		}

		i := utils.IndexOfString(args, c.name)
		if i == -1 {
			break
		}

		RootCmd.SetArgs(
			append(args[:i+1], append([]string{"--"}, args[i+1:]...)...))
		break
	}

	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// set default tarmak config folder
	tarmakConfigPath := "~/.tarmak"
	if envConfigPath := os.Getenv("TARMAK_CONFIG"); envConfigPath != "" {
		tarmakConfigPath = envConfigPath
	}

	RootCmd.PersistentFlags().StringVarP(
		&globalFlags.ConfigDirectory,
		"config-directory",
		"c",
		tarmakConfigPath,
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

	RootCmd.PersistentFlags().StringVar(
		&globalFlags.CurrentCluster,
		"current-cluster",
		"",
		"override the current cluster set in the config",
	)

	if version == "dev" {
		RootCmd.PersistentFlags().BoolVar(
			&globalFlags.WingDevMode,
			"wing-dev-mode",
			false,
			"use a bundled wing version rather than a tagged release from GitHub",
		)
	}

	RootCmd.PersistentFlags().BoolVar(
		&globalFlags.PublicAPIEndpoint,
		consts.KubeconfigFlagName,
		false,
		"Override kubeconfig to point to cluster's public API endpoint",
	)
}
