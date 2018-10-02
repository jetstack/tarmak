// Copyright Jetstack Ltd. See LICENSE for details.
package cmd

import (
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
)

var clusterCmd = &cobra.Command{
	Use:     "clusters",
	Short:   "Operations on clusters",
	Aliases: []string{"cluster"},
}

func clusterApplyFlags(fs *flag.FlagSet) {
	store := &globalFlags.Cluster.Apply
	clusterFlagDryRun(fs, &store.DryRun)

	fs.BoolVarP(
		&store.ConfigurationOnly,
		"configuration-only",
		"C",
		false,
		"apply changes to configuration only, by running only puppet",
	)

	fs.BoolVarP(
		&store.InfrastructureOnly,
		"infrastructure-only",
		"I",
		false,
		"apply changes to infrastructure only, by running only terraform",
	)

	fs.BoolVar(
		&store.AutoApprove,
		"auto-approve",
		false,
		"auto approve to responses when applying cluster",
	)

	fs.BoolVar(
		&store.AutoApprove,
		"auto-approve-deleting-data",
		false,
		"auto approve deletion of any data as a cause from applying cluster",
	)
}

func clusterDestroyFlags(fs *flag.FlagSet) {
	store := &globalFlags.Cluster.Destroy
	clusterFlagDryRun(fs, &store.DryRun)
}

func clusterImagesBuildFlags(fs *flag.FlagSet) {
	store := &globalFlags.Cluster.Images.Build

	fs.BoolVarP(
		&store.RebuildExisting,
		"rebuild-existing",
		"R",
		false,
		"build all images regardless whether they already exist",
	)
}

func clusterFlagDryRun(fs *flag.FlagSet, store *bool) {
	fs.BoolVar(
		store,
		"dry-run",
		false,
		"don't actually change anything, just show changes that would occur",
	)
}

func init() {
	RootCmd.AddCommand(clusterCmd)
}
