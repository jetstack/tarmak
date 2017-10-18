# prometheus

#### Table of Contents

1. [Description](#description)
2. [Classes](#classes)
3. [Defined Types](#defined-types)
## Description
This module is part of [Tarmak](http://docs.tarmak.io) and should currently be considered alpha.

[![Travis](https://img.shields.io/travis/jetstack/puppet-module-calico.svg)](https://travis-ci.org/jetstack/puppet-module-calico/)

## Classes

### `prometheus`



#### Parameters

##### `role`

* Type: `Any`
* Default: `''`

##### `etcd_cluster`

* Type: `Any`
* Default: `$::prometheus::params::etcd_cluster`

##### `etcd_k8s_port`

* Type: `Any`
* Default: `$::prometheus::params::etcd_k8s_port`

##### `etcd_events_port`

* Type: `Any`
* Default: `$::prometheus::params::etcd_events_port`

##### `etcd_overlay_port`

* Type: `Any`
* Default: `$::prometheus::params::etcd_overlay_port`

##### `blackbox_download_url`

* Type: `Any`
* Default: `$::prometheus::params::blackbox_download_url`

##### `blackbox_dest_dir`

* Type: `Any`
* Default: `$::prometheus::params::blackbox_dest_dir`

##### `blackbox_config_dir`

* Type: `Any`
* Default: `$::prometheus::params::blackbox_config_dir`

##### `systemd_path`

* Type: `Any`
* Default: `$::prometheus::params::systemd_path`

##### `node_exporter_image`

* Type: `Any`
* Default: `$::prometheus::params::node_exporter_image`

##### `node_exporter_version`

* Type: `Any`
* Default: `$::prometheus::params::node_exporter_version`

##### `node_exporter_port`

* Type: `Any`
* Default: `$::prometheus::params::node_exporter_port`

##### `addon_dir`

* Type: `Any`
* Default: `$::prometheus::params::addon_dir`

##### `helper_dir`

* Type: `Any`
* Default: `$::prometheus::params::helper_dir`

##### `prometheus_namespace`

* Type: `Any`
* Default: `$::prometheus::params::prometheus_namespace`

##### `prometheus_image`

* Type: `Any`
* Default: `$::prometheus::params::prometheus_image`

##### `prometheus_version`

* Type: `Any`
* Default: `$::prometheus::params::prometheus_version`

##### `prometheus_storage_local_retention`

* Type: `Any`
* Default: `$::prometheus::params::prometheus_storage_local_retention`

##### `prometheus_storage_local_memchunks`

* Type: `Any`
* Default: `$::prometheus::params::prometheus_storage_local_memchunks`

##### `prometheus_port`

* Type: `Any`
* Default: `$::prometheus::params::prometheus_port`

##### `prometheus_use_module_config`

* Type: `Any`
* Default: `$::prometheus::params::prometheus_use_module_config`

##### `prometheus_use_module_rules`

* Type: `Any`
* Default: `$::prometheus::params::prometheus_use_module_rules`

##### `prometheus_install_state_metrics`

* Type: `Any`
* Default: `$::prometheus::params::prometheus_install_state_metrics`

##### `prometheus_install_node_exporter`

* Type: `Any`
* Default: `$::prometheus::params::prometheus_install_node_exporter`


### `prometheus::blackbox_etcd`

Get blackbox_exporter

#### Parameters

##### `download_url`

* Type: `Any`
* Default: `$::prometheus::blackbox_download_url`

##### `config_dir`

* Type: `Any`
* Default: `$::prometheus::blackbox_config_dir`

##### `dest_dir`

* Type: `Any`
* Default: `$::prometheus::blackbox_dest_dir`

##### `systemd_path`

* Type: `Any`
* Default: `$::prometheus::systemd_path`


### `prometheus::node_exporter_service`



#### Parameters

##### `systemd_path`

* Type: `Any`
* Default: `$::prometheus::systemd_path`

##### `node_exporter_image`

* Type: `Any`
* Default: `$::prometheus::node_exporter_image`

##### `node_exporter_version`

* Type: `Any`
* Default: `$::prometheus::node_exporter_version`

##### `node_exporter_port`

* Type: `Any`
* Default: `$::prometheus::node_exporter_port`


### `prometheus::params`




### `prometheus::prometheus_deployment`



#### Parameters

##### `addon_dir`

* Type: `Any`
* Default: `$::prometheus::addon_dir`

##### `helper_dir`

* Type: `Any`
* Default: `$::prometheus::helper_dir`

##### `prometheus_namespace`

* Type: `Any`
* Default: `$::prometheus::prometheus_namespace`

##### `prometheus_image`

* Type: `Any`
* Default: `$::prometheus::prometheus_image`

##### `prometheus_version`

* Type: `Any`
* Default: `$::prometheus::prometheus_version`

##### `prometheus_storage_local_retention`

* Type: `Any`
* Default: `$::prometheus::prometheus_storage_local_retention`

##### `prometheus_storage_local_memchunks`

* Type: `Any`
* Default: `$::prometheus::prometheus_storage_local_memchunks`

##### `prometheus_port`

* Type: `Any`
* Default: `$::prometheus::prometheus_port`

##### `prometheus_use_module_config`

* Type: `Any`
* Default: `$::prometheus::prometheus_use_module_config`

##### `etcd_cluster`

* Type: `Any`
* Default: `$::prometheus::etcd_cluster`

##### `etcd_k8s_port`

* Type: `Any`
* Default: `$::prometheus::etcd_k8s_port`

##### `etcd_events_port`

* Type: `Any`
* Default: `$::prometheus::etcd_events_port`

##### `etcd_overlay_port`

* Type: `Any`
* Default: `$::prometheus::etcd_overlay_port`

##### `prometheus_use_module_rules`

* Type: `Any`
* Default: `$::prometheus::prometheus_use_module_rules`

##### `prometheus_install_state_metrics`

* Type: `Any`
* Default: `$::prometheus::prometheus_install_state_metrics`

##### `prometheus_install_node_exporter`

* Type: `Any`
* Default: `$::prometheus::prometheus_install_node_exporter`

##### `node_exporter_image`

* Type: `Any`
* Default: `$::prometheus::node_exporter_image`

##### `node_exporter_port`

* Type: `Any`
* Default: `$::prometheus::node_exporter_port`

##### `node_exporter_version`

* Type: `Any`
* Default: `$::prometheus::node_exporter_version`

## DefinedTypes

### `prometheus::wget_file`



#### Parameters

##### `url`

* Type: `String`

##### `destination_dir`

* Type: `String`

##### `destination_file`

* Type: `String`
* Default: `''`

##### `user`

* Type: `String`
* Default: `'root'`

##### `umask`

* Type: `String`
* Default: `'022'`
