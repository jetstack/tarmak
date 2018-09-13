# calico

#### Table of Contents

1. [Description](#description)
2. [Classes](#classes)
## Description
This module is part of [Tarmak](http://docs.tarmak.io) and should currently be considered alpha.

[![Travis](https://img.shields.io/travis/jetstack/puppet-module-calico.svg)](https://travis-ci.org/jetstack/puppet-module-calico/)

## Classes

### `calico`

calico init.pp

#### Parameters

##### `etcd_cluster`

* Type: `Array[String]`
* Default: `$::calico::params::etcd_cluster`

##### `etcd_overlay_port`

* Type: `Integer[1,65535]`
* Default: `$::calico::params::etcd_overlay_port`

##### `backend`

* Type: `Enum['etcd', 'kubernetes']`
* Default: `'etcd'`

##### `etcd_ca_file`

* Type: `String`
* Default: `''`

##### `etcd_cert_file`

* Type: `String`
* Default: `''`

##### `etcd_key_file`

* Type: `String`
* Default: `''`

##### `cloud_provider`

* Type: `String`
* Default: `$::calico::params::cloud_provider`

##### `namespace`

* Type: `String`
* Default: `'kube-system'`

##### `pod_network`

* Type: `Optional[String]`
* Default: `undef`

##### `mtu`

* Type: `Integer[1000,65535]`
* Default: `1480`


### `calico::config`




### `calico::disable_source_destination_check`

This class disable the source/destination check on AWS instances

#### Parameters

##### `image`

* Type: `String`
* Default: `'ottoyiu/k8s-ec2-srcdst'`

##### `version`

* Type: `String`
* Default: `'0.1.0'`


### `calico::node`

Calico Node

Calico Node contains a Daemon Set that spinsup the overlay network on every
workern node.

#### Parameters

##### `metrics_port`

* Port for felix metrics endpoint, 0 disables metrics collection
* Type: `Integer[0,65535]`
* Default: `9091`

##### `node_image`

* Type: `String`
* Default: `'quay.io/calico/node'`

##### `node_version`

* Type: `String`
* Default: `'3.1.1'`

##### `cni_image`

* Type: `String`
* Default: `'quay.io/calico/cni'`

##### `cni_version`

* Type: `String`
* Default: `'3.1.1'`

##### `ipv4_pool_ipip_mode`

* Type: `Enum['always', 'cross-subnet', 'off']`
* Default: `'always'`


### `calico::params`

calico params.pp


### `calico::policy_controller`



#### Parameters

##### `image`

* Type: `String`
* Default: `'quay.io/calico/kube-controllers'`

##### `version`

* Type: `String`
* Default: `'3.1.1'`
