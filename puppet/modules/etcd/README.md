# etcd

#### Table of Contents

1. [Description](#description)
2. [Classes](#classes)
3. [Defined Types](#defined-types)
## Description
This module is part of [Tarmak](http://docs.tarmak.io) and should currently be considered alpha.

[![Travis](https://img.shields.io/travis/jetstack/puppet-module-etcd.svg)](https://travis-ci.org/jetstack/puppet-module-etcd/)

## Classes

### `etcd`

Install/configure an etcd node.

#### Parameters

##### `data_dir`

* The directory to store etcd data
* Type: `Any`
* Default: `$::etcd::params::data_dir`

##### `config_dir`

* The directory to store etcd config
* Type: `Any`
* Default: `$::etcd::params::config_dir`

##### `user`

* The username to run etcd
* Type: `Any`
* Default: `$::etcd::params::user`

##### `uid`

* The user ID to run etcd
* Type: `Any`
* Default: `$::etcd::params::uid`

##### `group`

* The group to run etcd
* Type: `Any`
* Default: `$::etcd::params::group`

##### `gid`

* The etcd group ID
* Type: `Any`
* Default: `$::etcd::params::gid`


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
