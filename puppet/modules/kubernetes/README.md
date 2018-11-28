# kubernetes

#### Table of Contents

1. [Description](#description)
2. [Classes](#classes)
3. [Defined Types](#defined-types)
## Description
This module is part of [Tarmak](http://docs.tarmak.io) and should currently be considered alpha.

[![Travis](https://img.shields.io/travis/jetstack/puppet-module-kubernetes.svg)](https://travis-ci.org/jetstack/puppet-module-kubernetes/)

## Classes

### `kubernetes`

Class: kubernetes

#### Parameters

##### `version`

* Type: `Any`
* Default: `$::kubernetes::params::version`

##### `bin_dir`

* Type: `Any`
* Default: `$::kubernetes::params::bin_dir`

##### `download_dir`

* Type: `Any`
* Default: `$::kubernetes::params::download_dir`

##### `download_url`

* Type: `Any`
* Default: `$::kubernetes::params::download_url`

##### `dest_dir`

* Type: `Any`
* Default: `$::kubernetes::params::dest_dir`

##### `config_dir`

* Type: `Any`
* Default: `$::kubernetes::params::config_dir`

##### `systemd_dir`

* Type: `Any`
* Default: `$::kubernetes::params::systemd_dir`

##### `run_dir`

* Type: `Any`
* Default: `$::kubernetes::params::run_dir`

##### `apply_dir`

* Type: `Any`
* Default: `$::kubernetes::params::apply_dir`

##### `uid`

* Type: `Any`
* Default: `$::kubernetes::params::uid`

##### `gid`

* Type: `Any`
* Default: `$::kubernetes::params::gid`

##### `user`

* Type: `Any`
* Default: `$::kubernetes::params::user`

##### `group`

* Type: `Any`
* Default: `$::kubernetes::params::group`

##### `master_url`

* Type: `String`
* Default: `''`

##### `curl_path`

* Type: `Any`
* Default: `$::kubernetes::params::curl_path`

##### `ssl_dir`

* Type: `Any`
* Default: `undef`

##### `source`

* Type: `Any`
* Default: `undef`

##### `cloud_provider`

* Type: `Enum['aws', '']`
* Default: `''`

##### `cluster_name`

* Type: `Any`
* Default: `undef`

##### `dns_root`

* Type: `Any`
* Default: `undef`

##### `cluster_dns`

* Type: `Any`
* Default: `undef`

##### `cluster_domain`

* Type: `Any`
* Default: `'cluster.local'`

##### `service_ip_range_network`

* Type: `Any`
* Default: `'10.254.0.0'`

##### `service_ip_range_mask`

* Type: `Any`
* Default: `'16'`

##### `leader_elect`

* Type: `Any`
* Default: `true`

##### `allow_privileged`

* Type: `Any`
* Default: `true`

##### `pod_security_policy`

* Type: `Any`
* Default: `undef`

##### `enable_pod_priority`

* Type: `Optional[Boolean]`
* Default: `undef`

##### `service_account_key_file`

* Type: `Any`
* Default: `undef`

##### `service_account_key_generate`

* Type: `Any`
* Default: `false`

##### `pod_network`

* Type: `Optional[String]`
* Default: `undef`

##### `apiserver_insecure_port`

* Type: `Integer[-1,65535]`
* Default: `-`

##### `apiserver_secure_port`

* Type: `Integer[0,65535]`
* Default: `6443`

##### `authorization_mode`

* Type: `Array[Enum['AlwaysAllow', 'ABAC', 'RBAC']]`
* Default: `[]`


### `kubernetes::apiserver`

class kubernetes::master

#### Parameters

##### `allow_privileged`

* Type: `Any`
* Default: `true`

##### `audit_enabled`

* Type: `Optional[Boolean]`
* Default: `undef`

##### `audit_log_directory`

* Type: `String`
* Default: `'/var/log/kubernetes'`

##### `audit_log_maxbackup`

* Type: `Integer`
* Default: `1`

##### `audit_log_maxsize`

* Type: `Integer`
* Default: `100`

##### `admission_control`

* Type: `Any`
* Default: `undef`

##### `feature_gates`

* Type: `Any`
* Default: `[]`

##### `count`

* Type: `Any`
* Default: `1`

##### `storage_backend`

* Type: `Any`
* Default: `undef`

##### `encryption_config_file`

* Type: `Optional[String]`
* Default: `undef`

##### `etcd_nodes`

* Type: `Any`
* Default: `['localhost']`

##### `etcd_port`

* Type: `Any`
* Default: `2379`

##### `etcd_events_port`

* Type: `Any`
* Default: `undef`

##### `etcd_ca_file`

* Type: `Any`
* Default: `undef`

##### `etcd_cert_file`

* Type: `Any`
* Default: `undef`

##### `etcd_key_file`

* Type: `Any`
* Default: `undef`

##### `kubelet_client_cert_file`

* Type: `Any`
* Default: `undef`

##### `kubelet_client_key_file`

* Type: `Any`
* Default: `undef`

##### `requestheader_allowed_names`

* Type: `String`
* Default: `'kube-apiserver-proxy'`

##### `requestheader_extra_headers_prefix`

* Type: `String`
* Default: `'X-Remote-Extra-'`

##### `requestheader_group_headers`

* Type: `String`
* Default: `'X-Remote-Group'`

##### `requestheader_username_headers`

* Type: `String`
* Default: `'X-Remote-User'`

##### `requestheader_client_ca_file`

* Type: `Any`
* Default: `undef`

##### `proxy_client_cert_file`

* Type: `Any`
* Default: `undef`

##### `proxy_client_key_file`

* Type: `Any`
* Default: `undef`

##### `ca_file`

* Type: `Any`
* Default: `undef`

##### `cert_file`

* Type: `Any`
* Default: `undef`

##### `key_file`

* Type: `Any`
* Default: `undef`

##### `oidc_client_id`

* Type: `Optional[String]`
* Default: `undef`

##### `oidc_groups_claim`

* Type: `Optional[String]`
* Default: `undef`

##### `oidc_groups_prefix`

* Type: `Optional[String]`
* Default: `undef`

##### `oidc_issuer_url`

* Type: `Optional[String]`
* Default: `undef`

##### `oidc_signing_algs`

* Type: `Array[String]`
* Default: `[]`

##### `oidc_username_claim`

* Type: `Optional[String]`
* Default: `undef`

##### `oidc_username_prefix`

* Type: `Optional[String]`
* Default: `undef`

##### `systemd_wants`

* Type: `Any`
* Default: `[]`

##### `systemd_requires`

* Type: `Any`
* Default: `[]`

##### `systemd_after`

* Type: `Any`
* Default: `[]`

##### `systemd_before`

* Type: `Any`
* Default: `[]`

##### `runtime_config`

* Type: `Any`
* Default: `[]`

##### `insecure_bind_address`

* Type: `Any`
* Default: `undef`

##### `abac_full_access_users`

* Type: `Array[String]`
* Default: `[]`

##### `abac_read_only_access_users`

* Type: `Array[String]`
* Default: `[]`


### `kubernetes::controller_manager`

class kubernetes::master

#### Parameters

##### `ca_file`

* Type: `Any`
* Default: `undef`

##### `cert_file`

* Type: `Any`
* Default: `undef`

##### `key_file`

* Type: `Any`
* Default: `undef`

##### `systemd_wants`

* Type: `Any`
* Default: `[]`

##### `systemd_requires`

* Type: `Any`
* Default: `[]`

##### `systemd_after`

* Type: `Any`
* Default: `[]`

##### `systemd_before`

* Type: `Any`
* Default: `[]`

##### `allocate_node_cidrs`

* Type: `Boolean`
* Default: `false`


### `kubernetes::dns`

== Class kubernetes::dns

#### Parameters

##### `image`

* Type: `Any`
* Default: `'gcr.io/google_containers/k8s-dns-kube-dns-amd64'`

##### `version`

* Type: `Any`
* Default: `'1.14.5'`

##### `dnsmasq_image`

* Type: `Any`
* Default: `'gcr.io/google_containers/k8s-dns-dnsmasq-nanny-amd64'`

##### `dnsmasq_version`

* Type: `Any`
* Default: `'1.14.5'`

##### `sidecar_image`

* Type: `Any`
* Default: `'gcr.io/google_containers/k8s-dns-sidecar-amd64'`

##### `sidecar_version`

* Type: `Any`
* Default: `'1.14.5'`

##### `autoscaler_image`

* Type: `Any`
* Default: `'gcr.io/google_containers/cluster-proportional-autoscaler-amd64'`

##### `autoscaler_version`

* Type: `Any`
* Default: `'1.1.1-r2'`

##### `min_replicas`

* Type: `Any`
* Default: `3`


### `kubernetes::install`

download and install hyperkube


### `kubernetes::kubectl`

class kubernetes::kubectl

#### Parameters

##### `ca_file`

* Type: `Any`
* Default: `undef`

##### `cert_file`

* Type: `Any`
* Default: `undef`

##### `key_file`

* Type: `Any`
* Default: `undef`


### `kubernetes::kubelet`



#### Parameters

##### `cgroup_kubernetes_name`

* name of cgroup slice for kubernetes related processes

##### `cgroup_kubernetes_reserved_memory`

* memory reserved for kubernetes related processes

##### `cgroup_kubernetes_reserved_cpu`

* CPU reserved for kubernetes related processes

##### `cgroup_system_name`

* name of cgroup slice for system processes
* Type: `Optional[String]`
* Default: `'/system.slice'`

##### `cgroup_system_reserved_memory`

* memory reserved for system processes
* Type: `Optional[String]`
* Default: `'128Mi'`

##### `cgroup_system_reserved_cpu`

* CPU reserved for system processes
* Type: `Optional[String]`
* Default: `'100m'`

##### `role`

* Type: `String`
* Default: `'worker'`

##### `container_runtime`

* Type: `String`
* Default: `'docker'`

##### `kubelet_dir`

* Type: `String`
* Default: `'/var/lib/kubelet'`

##### `eviction_hard_memory_available_threshold`

* Type: `Optional[String]`
* Default: `'5%'`

##### `eviction_hard_nodefs_available_threshold`

* Type: `Optional[String]`
* Default: `'10%'`

##### `eviction_hard_nodefs_inodes_free_threshold`

* Type: `Optional[String]`
* Default: `'5%'`

##### `eviction_soft_enabled`

* Type: `Boolean`
* Default: `true`

##### `eviction_soft_memory_available_threshold`

* Type: `Optional[String]`
* Default: `'10%'`

##### `eviction_soft_nodefs_available_threshold`

* Type: `Optional[String]`
* Default: `'15%'`

##### `eviction_soft_nodefs_inodes_free_threshold`

* Type: `Optional[String]`
* Default: `'10%'`

##### `eviction_soft_memory_available_grace_period`

* Type: `Optional[String]`
* Default: `'0m'`

##### `eviction_soft_nodefs_available_grace_period`

* Type: `Optional[String]`
* Default: `'0m'`

##### `eviction_soft_nodefs_inodes_free_grace_period`

* Type: `Optional[String]`
* Default: `'0m'`

##### `eviction_max_pod_grace_period`

* Type: `String`
* Default: `'-1'`

##### `eviction_pressure_transition_period`

* Type: `String`
* Default: `'2m'`

##### `eviction_minimum_reclaim_memory_available`

* Type: `Optional[String]`
* Default: `'100Mi'`

##### `eviction_minimum_reclaim_nodefs_available`

* Type: `Optional[String]`
* Default: `'1Gi'`

##### `eviction_minimum_reclaim_nodefs_inodes_free`

* Type: `Optional[String]`
* Default: `undef`

##### `network_plugin`

* Type: `Optional[String]`
* Default: `undef`

##### `network_plugin_mtu`

* Type: `Integer`
* Default: `1460`

##### `allow_privileged`

* Type: `Boolean`
* Default: `true`

##### `register_node`

* Type: `Boolean`
* Default: `true`

##### `register_schedulable`

* Type: `Optional[Boolean]`
* Default: `undef`

##### `ca_file`

* Type: `Optional[String]`
* Default: `undef`

##### `cert_file`

* Type: `Optional[String]`
* Default: `undef`

##### `key_file`

* Type: `Optional[String]`
* Default: `undef`

##### `client_ca_file`

* Type: `Optional[String]`
* Default: `undef`

##### `feature_gates`

* Type: `Any`
* Default: `[]`

##### `node_labels`

* Type: `Any`
* Default: `undef`

##### `node_taints`

* Type: `Any`
* Default: `undef`

##### `pod_cidr`

* Type: `Any`
* Default: `undef`

##### `hostname_override`

* Type: `Any`
* Default: `undef`

##### `cgroup_driver`

* Type: `Enum['systemd', 'cgroupfs']`
* Default: `$::osfamily`

##### `cgroup_root`

* Type: `String`
* Default: `'/'`

##### `cgroup_kube_name`

* Type: `Optional[String]`
* Default: `'/podruntime.slice'`

##### `cgroup_kube_reserved_memory`

* Type: `Optional[String]`
* Default: `undef`

##### `cgroup_kube_reserved_cpu`

* Type: `Optional[String]`
* Default: `'100m'`

##### `systemd_wants`

* Type: `Array[String]`
* Default: `[]`

##### `systemd_requires`

* Type: `Array[String]`
* Default: `[]`

##### `systemd_after`

* Type: `Array[String]`
* Default: `[]`

##### `systemd_before`

* Type: `Array[String]`
* Default: `[]`


### `kubernetes::master`

class kubernetes::master

#### Parameters

##### `disable_kubelet`

* Type: `Any`
* Default: `false`

##### `disable_proxy`

* Type: `Any`
* Default: `false`


### `kubernetes::master_params`

== Class kubernetes::params


### `kubernetes::params`

== Class kubernetes::params


### `kubernetes::pod_security_policy`

This class manages RBAC manifests


### `kubernetes::proxy`

class kubernetes::kubelet

#### Parameters

##### `ca_file`

* Type: `Optional[String]`
* Default: `undef`

##### `cert_file`

* Type: `Optional[String]`
* Default: `undef`

##### `key_file`

* Type: `Optional[String]`
* Default: `undef`

##### `systemd_wants`

* Type: `Array[String]`
* Default: `[]`

##### `systemd_requires`

* Type: `Array[String]`
* Default: `[]`

##### `systemd_after`

* Type: `Array[String]`
* Default: `[]`

##### `systemd_before`

* Type: `Array[String]`
* Default: `[]`


### `kubernetes::rbac`

This class manages RBAC manifests


### `kubernetes::scheduler`

class kubernetes::master

#### Parameters

##### `ca_file`

* Type: `Any`
* Default: `undef`

##### `cert_file`

* Type: `Any`
* Default: `undef`

##### `key_file`

* Type: `Any`
* Default: `undef`

##### `systemd_wants`

* Type: `Any`
* Default: `[]`

##### `systemd_requires`

* Type: `Any`
* Default: `[]`

##### `systemd_after`

* Type: `Any`
* Default: `[]`

##### `systemd_before`

* Type: `Any`
* Default: `[]`

##### `feature_gates`

* Type: `Any`
* Default: `[]`


### `kubernetes::storage_classes`

This class sets up the default storage classes for cloud providers


### `kubernetes::worker`

class kubernetes::worker

## DefinedTypes

### `kubernetes::apply`

adds resources to a kubernetes master

#### Parameters

##### `manifests`

* Type: `Any`
* Default: `[]`

##### `force`

* Type: `Any`
* Default: `false`

##### `format`

* Type: `Any`
* Default: `'yaml'`

##### `systemd_wants`

* Type: `Any`
* Default: `[]`

##### `systemd_requires`

* Type: `Any`
* Default: `[]`

##### `systemd_after`

* Type: `Any`
* Default: `[]`

##### `systemd_before`

* Type: `Any`
* Default: `[]`

##### `type`

* Type: `Enum['manifests','concat']`
* Default: `'manifests'`


### `kubernetes::apply_fragment`

Concat fragment for apply

#### Parameters

##### `content`

* Type: `Any`

##### `order`

* Type: `Any`

##### `target`

* Type: `Any`

##### `format`

* Type: `Any`
* Default: `'yaml'`


### `kubernetes::symlink`

adds a symlink to hyperkube
