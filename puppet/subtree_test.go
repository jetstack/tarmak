// Copyright Jetstack Ltd. See LICENSE for details.
package main

import (
	"testing"

	"github.com/jetstack/tarmak/pkg/tarmak/utils/subtree"
)

func TestSubtreeUpstreamedPuppetModuleAWSEBS(t *testing.T) {
	subtree.New(
		"puppet/modules/aws_ebs",
		"https://github.com/jetstack/puppet-module-aws_ebs.git",
	).TestSubtreeUpstream(t)
}

func TestSubtreeUpstreamedPuppetModuleCalico(t *testing.T) {
	subtree.New(
		"puppet/modules/calico",
		"https://github.com/jetstack/puppet-module-calico.git",
	).TestSubtreeUpstream(t)
}

func TestSubtreeUpstreamedPuppetModuleEtcd(t *testing.T) {
	subtree.New(
		"puppet/modules/etcd",
		"https://github.com/jetstack/puppet-module-etcd.git",
	).TestSubtreeUpstream(t)
}

func TestSubtreeUpstreamedPuppetModuleKubernetes(t *testing.T) {
	t.Skip("skip module kubernetes check, as it's tested in tarmak")
	subtree.New(
		"puppet/modules/kubernetes",
		"https://github.com/jetstack/puppet-module-kubernetes.git",
	).TestSubtreeUpstream(t)
}

func TestSubtreeUpstreamedPuppetModuleKubernetesAddons(t *testing.T) {
	subtree.New(
		"puppet/modules/kubernetes_addons",
		"https://github.com/jetstack/puppet-module-kubernetes_addons.git",
	).TestSubtreeUpstream(t)
}

func TestSubtreeUpstreamedPuppetModulePrometheus(t *testing.T) {
	subtree.New(
		"puppet/modules/prometheus",
		"https://github.com/jetstack/puppet-module-prometheus.git",
	).TestSubtreeUpstream(t)
}

func TestSubtreeUpstreamedPuppetModuleTarmak(t *testing.T) {
	t.Skip("skip module tarmak check, as it's tested in tarmak")
	subtree.New(
		"puppet/modules/tarmak",
		"https://github.com/jetstack/puppet-module-tarmak.git",
	).TestSubtreeUpstream(t)
}

func TestSubtreeUpstreamedPuppetModuleVaultClient(t *testing.T) {
	subtree.New(
		"puppet/modules/vault_client",
		"https://github.com/jetstack/puppet-module-vault_client.git",
	).TestSubtreeUpstream(t)
}
