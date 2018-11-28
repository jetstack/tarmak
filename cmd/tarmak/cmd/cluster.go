// Copyright Jetstack Ltd. See LICENSE for details.
package cmd

import (
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"

	"github.com/jetstack/tarmak/pkg/tarmak/utils/consts"
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
		true,
		"auto approve to responses when applying cluster",
	)

	fs.BoolVar(
		&store.AutoApproveDeletingData,
		"auto-approve-deleting-data",
		false,
		"auto approve deletion of any data as a cause from applying cluster",
	)

	fs.StringVarP(
		&store.PlanFileLocation,
		"plan-file-location",
		"P",
		consts.DefaultPlanLocationPlaceholder,
		"location of stored terraform plan executable file to be used",
	)

	fs.BoolVarP(
		&store.WaitForConvergence,
		"wait-for-convergence",
		"W",
		true,
		"wait for wing convergence on applied instances",
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

func clusterPlanFlags(fs *flag.FlagSet) {
	store := &globalFlags.Cluster.Plan

	fs.StringVarP(
		&store.PlanFileStore,
		"plan-file-store",
		"P",
		consts.DefaultPlanLocationPlaceholder,
		"location to store terraform plan executable file",
	)
}

func clusterImagesDestroyFlags(fs *flag.FlagSet) {
	store := &globalFlags.Cluster.Images.Destroy

	fs.BoolVarP(
		&store.All,
		"all",
		"A",
		false,
		"destroy all tarmak images for this cluster",
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

func clusterKubeconfigFlags(fs *flag.FlagSet) {
	store := &globalFlags.Cluster.Kubeconfig

	fs.StringVarP(
		&store.Path,
		"path",
		"p",
		consts.DefaultKubeconfigPath,
		"Path to store kubeconfig file",
	)
}

func init() {
	RootCmd.AddCommand(clusterCmd)
}
