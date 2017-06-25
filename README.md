# calico

#### Table of Contents

1. [Description](#description)
2. [Classes](#classes)
## Description
calico init.pp

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

##### `mtu`

* Type: `Integer[1000,65535]`
* Default: `1480`


### `calico::config`




### `calico::disable_source_destination_check`

This class disable the source/destination check on AWS instances


### `calico::node`



#### Parameters

##### `node_image`

* Type: `String`
* Default: `'quay.io/calico/node'`

##### `node_version`

* Type: `String`
* Default: `'1.3.0'`

##### `cni_image`

* Type: `String`
* Default: `'quay.io/calico/cni'`

##### `cni_version`

* Type: `String`
* Default: `'1.9.1'`

##### `ipv4_pool_cidr`

* Type: `String`
* Default: `'10.231.0.0/16'`

##### `ipv4_pool_ipip_mode`

* Type: `Enum['always', 'cross-subnet', 'off']`
* Default: `'always'`


### `calico::params`

calico params.pp


### `calico::policy_controller`



#### Parameters

##### `image`

* Type: `String`
* Default: `'quay.io/calico/kube-policy-controller'`

##### `version`

* Type: `String`
* Default: `'0.6.0'`
