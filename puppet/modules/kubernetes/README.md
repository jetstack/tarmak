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

* Type: `Any`
* Default: `$::kubernetes::params::master_url`

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

##### `service_account_key_file`

* Type: `Any`
* Default: `undef`

##### `service_account_key_generate`

* Type: `Any`
* Default: `false`

##### `authorization_mode`

* Type: `Array[Enum['AlwaysAllow', 'ABAC', 'RBAC']]`
* Default: `[]`


### `kubernetes::apiserver`

class kubernetes::master

#### Parameters

##### `allow_privileged`

* Type: `Any`
* Default: `true`

##### `admission_control`

* Type: `Any`
* Default: `undef`

##### `count`

* Type: `Any`
* Default: `1`

##### `storage_backend`

* Type: `Any`
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

class kubernetes::kubelet

#### Parameters

##### `role`

* Type: `Any`
* Default: `'worker'`

##### `container_runtime`

* Type: `Any`
* Default: `'docker'`

##### `kubelet_dir`

* Type: `Any`
* Default: `'/var/lib/kubelet'`

##### `network_plugin`

* Type: `Any`
* Default: `undef`

##### `network_plugin_mtu`

* Type: `Any`
* Default: `1460`

##### `allow_privileged`

* Type: `Any`
* Default: `true`

##### `register_node`

* Type: `Any`
* Default: `true`

##### `register_schedulable`

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

##### `node_labels`

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
* Default: `'systemd'`


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


### `kubernetes::proxy`

class kubernetes::kubelet

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
