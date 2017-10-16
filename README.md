# etcd

Install/configure an etcd node.

This module is part of [Tarmak](http://docs.tarmak.io) and should currently be
considered alpha.

#### Table of Contents

1. [Description](#description)
2. [Classes](#classes)
3. [Defined Types](#defined-types)
## Description
etcd

## Classes

### `etcd`

etcd

#### Parameters

##### `data_dir`

* Type: `Any`
* Default: `$::etcd::params::data_dir`

##### `config_dir`

* Type: `Any`
* Default: `$::etcd::params::config_dir`

##### `uid`

* Type: `Any`
* Default: `$::etcd::params::uid`

##### `gid`

* Type: `Any`
* Default: `$::etcd::params::gid`

##### `user`

* Type: `Any`
* Default: `$::etcd::params::user`

##### `group`

* Type: `Any`
* Default: `$::etcd::params::group`


### `etcd::params`

etcd variable defaults

## DefinedTypes

### `etcd::install`



#### Parameters

##### `ensure`

* Type: `String`
* Default: `'present'`


### `etcd::instance`



#### Parameters

##### `version`

* Type: `String`

##### `client_port`

* Type: `Integer`
* Default: `2379`

##### `peer_port`

* Type: `Integer`
* Default: `2380`

##### `members`

* Type: `Integer`
* Default: `1`

##### `nodename`

* Type: `String`
* Default: `$::fqdn`

##### `tls`

* Type: `Boolean`
* Default: `false`

##### `tls_cert_path`

* Type: `String`
* Default: `nil`

##### `tls_key_path`

* Type: `String`
* Default: `nil`

##### `tls_ca_path`

* Type: `String`
* Default: `nil`

##### `advertise_client_network`

* Type: `String`
* Default: `nil`

##### `initial_cluster`

* Type: `Array`
* Default: `[]`
