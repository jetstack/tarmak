// Copyright Jetstack Ltd. See LICENSE for details.
package v1alpha1

// constants for kubernetes role names
const (
	KubernetesMasterRoleName = "master"
	KubernetesWorkerRoleName = "worker"
	KubernetesEtcdRoleName   = "etcd"

	ImageBaseDefault       = "centos-puppet-agent"
	ImageBaseDefaultWorker = "centos-puppet-agent-k8s-worker"
)
