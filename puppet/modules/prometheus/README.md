# prometheus

#### Table of Contents

1. [Description](#description)
2. [Classes](#classes)
3. [Defined Types](#defined-types)
## Description
This module is part of [Tarmak](http://docs.tarmak.io) and should currently be considered alpha.

[![Travis](https://img.shields.io/travis/jetstack/puppet-module-prometheus.svg)](https://travis-ci.org/jetstack/puppet-module-prometheus/)

## Classes

### `prometheus`



#### Parameters

##### `systemd_path`

* Type: `String`
* Default: `'/etc/systemd/system'`

##### `namespace`

* Type: `String`
* Default: `'monitoring'`

##### `role`

* Type: `Optional[Enum['etcd','master','worker']]`
* Default: `$::prometheus::params::role`

##### `etcd_cluster_exporters`

* Type: `Any`
* Default: `$::prometheus::params::etcd_cluster_exporters`

##### `etcd_cluster_node_exporters`

* Type: `Any`
* Default: `$::prometheus::params::etcd_cluster_node_exporters`

##### `etcd_k8s_main_port`

* Type: `Integer[1025,65535]`
* Default: `$::prometheus::params::etcd_k8s_main_port`

##### `etcd_k8s_events_port`

* Type: `Integer[1025,65535]`
* Default: `$::prometheus::params::etcd_k8s_events_port`

##### `etcd_overlay_port`

* Type: `Integer[1024,65535]`
* Default: `$::prometheus::params::etcd_overlay_port`

##### `mode`

* Type: `String`
* Default: `'Full'`


### `prometheus::blackbox_exporter`

Sets up a blackbox exporter to blackbox probe in-cluster services and pods

#### Parameters

##### `image`

* Type: `String`
* Default: `'prom/blackbox-exporter'`

##### `version`

* Type: `String`
* Default: `'0.12.0'`

##### `port`

* Type: `Integer`
* Default: `9115`

##### `replicas`

* Type: `Integer`
* Default: `2`


### `prometheus::blackbox_exporter_etcd`

exporter_# Sets up a blackbox exporter to forward etcd metrics from etcd nodes

#### Parameters

##### `download_url`

* Type: `String`
* Default: `'https://github.com/jetstack-experimental/blackbox_exporter/releases/download/v#VERSION#/blackbox_exporter_#VERSION#_linux_amd64'`

##### `version`

* Type: `String`
* Default: `'0.4.0-jetstack'`

##### `config_dir`

* Type: `String`
* Default: `'/etc/blackbox_exporter'`

##### `port`

* Type: `Integer`
* Default: `9115`


### `prometheus::kube_state_metrics`



#### Parameters

##### `image`

* Type: `String`
* Default: `'gcr.io/google_containers/kube-state-metrics'`

##### `version`

* Type: `String`
* Default: `'1.2.0'`

##### `resizer_image`

* Type: `String`
* Default: `'gcr.io/google_containers/addon-resizer'`

##### `resizer_version`

* Type: `String`
* Default: `'1.0'`


### `prometheus::node_exporter`



#### Parameters

##### `image`

* Type: `String`
* Default: `'prom/node-exporter'`

##### `version`

* Type: `String`
* Default: `'0.15.2'`

##### `download_url`

* Type: `String`
* Default: `'https://github.com/prometheus/node_exporter/releases/download/v#VERSION#/node_exporter-#VERSION#.linux-amd64.tar.gz'`

##### `sha256sums_url`

* Type: `String`
* Default: `'https://github.com/prometheus/node_exporter/releases/download/v#VERSION#/sha256sums.txt'`

##### `signature_url`

* Type: `String`
* Default: `'https://releases.tarmak.io/signatures/node_exporter/#VERSION#/sha256sums.txt.asc'`

##### `port`

* Type: `Any`
* Default: `9100`


### `prometheus::params`




### `prometheus::server`



#### Parameters

##### `image`

* Type: `String`
* Default: `'prom/prometheus'`

##### `version`

* Type: `String`
* Default: `'2.2.1'`

##### `reloader_image`

* Type: `String`
* Default: `'jimmidyson/configmap-reload'`

##### `reloader_version`

* Type: `String`
* Default: `'0.1'`

##### `retention`

* Type: `String`
* Default: `'720h'`

##### `port`

* Type: `Integer[1025,65535]`
* Default: `9090`

##### `external_url`

* Type: `String`
* Default: `''`

##### `persistent_volume`

* Type: `Boolean`
* Default: `false`

##### `persistent_volume_size`

* Type: `Integer`
* Default: `15`

##### `kubernetes_token_file`

* Type: `String`
* Default: `'/var/run/secrets/kubernetes.io/serviceaccount/token'`

##### `kubernetes_ca_file`

* Type: `String`
* Default: `'/var/run/secrets/kubernetes.io/serviceaccount/ca.crt'`

##### `external_labels`

* Type: `Hash[String,String]`
* Default: `{}`

## DefinedTypes

### `prometheus::rule`



#### Parameters

##### `expr`

* Type: `String`

##### `summary`

* Type: `String`

##### `description`

* Type: `String`

##### `for`

* Type: `String`
* Default: `'5m'`

##### `labels`

* Type: `Any`
* Default: `{'severity' => 'page'}`

##### `order`

* Type: `Integer`
* Default: `10`


### `prometheus::scrape_config`



#### Parameters

##### `order`

* Type: `Integer`

##### `config`

* Type: `Any`
* Default: `{}`

##### `job_name`

* Type: `Any`
* Default: `$title`


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
