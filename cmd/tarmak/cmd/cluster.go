package cmd

import (
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
)

var ClusterFlagInfrastructureStacks = "infrastructure-stacks"

var clusterCmd = &cobra.Command{
	Use:     "clusters",
	Short:   "Operations on clusters",
	Aliases: []string{"cluster"},
}

func clusterApplyFlags(fs *flag.FlagSet) {
	store := &globalFlags.Cluster.Apply
	clusterFlagDryRun(fs, &store.DryRun)
	clusterFlagInfrastructureStacks(fs, &store.InfrastructureStacks)

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
}

func clusterDestroyFlags(fs *flag.FlagSet) {
	store := &globalFlags.Cluster.Destroy
	clusterFlagDryRun(fs, &store.DryRun)
	clusterFlagInfrastructureStacks(fs, &store.InfrastructureStacks)

	fs.BoolVar(
		&store.ForceDestroyStateStack,
		"force-destroy-state-stack",
		false,
		"force destroy the state stack, this is unreversible (!!!)",
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

func clusterFlagInfrastructureStacks(fs *flag.FlagSet, store *[]string) {
	fs.StringArrayVarP(
		store,
		"infrastructure-stacks",
		"S",
		[]string{},
		// TODO: add validation based on cluster type
		"run operation on these stacks only, valid stacks are: state, network, tools, bastion, vault, kubernetes",
	)
}

func init() {
	RootCmd.AddCommand(clusterCmd)
}
