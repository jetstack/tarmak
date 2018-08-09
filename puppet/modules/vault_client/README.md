# vault_client

#### Table of Contents

1. [Description](#description)
2. [Classes](#classes)
3. [Defined Types](#defined-types)
## Description
This module is part of [Tarmak](http://docs.tarmak.io) and should currently be considered alpha.

[![Travis](https://img.shields.io/travis/jetstack/puppet-module-vault_client.svg)](https://travis-ci.org/jetstack/puppet-module-vault_client/)

## Classes

### `vault_client`

Class: vault_client
===========================

Puppet module to install and manage a vault client install

=== Parameters

[*version*]
  The package version to install

[*token*]
  Static token for the vault client
  Either token or init_token needs to be specified

[*init_token*]
  Initial token for the vault client to generate node unique token
  Either token or init_token needs to be specified

[*init_policies*]
  TODO

[*init_role*]
  TODO

[*download_url*]
  Download url for the `vault-helper` binary. It supports the placeholder `#VERSION#`
  that gets replaced by `$version` variable.

#### Parameters

##### `version`

* Type: `Any`
* Default: `$::vault_client::params::version`

##### `bin_dir`

* Type: `Any`
* Default: `$::vault_client::params::bin_dir`

##### `download_dir`

* Type: `Any`
* Default: `$::vault_client::params::download_dir`

##### `download_url`

* Type: `Any`
* Default: `$::vault_client::params::download_url`

##### `dest_dir`

* Type: `Any`
* Default: `$::vault_client::params::dest_dir`

##### `server_url`

* Type: `Any`
* Default: `$::vault_client::params::server_url`

##### `systemd_dir`

* Type: `Any`
* Default: `$::vault_client::params::systemd_dir`

##### `init_token`

* Type: `Any`
* Default: `undef`

##### `init_role`

* Type: `Any`
* Default: `undef`

##### `token`

* Type: `Any`
* Default: `undef`

##### `ca_cert_path`

* Type: `Any`
* Default: `undef`


### `vault_client::config`

== Class vault_client::config

This class is called from vault_client for service config.


### `vault_client::install`




### `vault_client::params`

== Class vault_client::params

This class is meant to be called from vault_client.
It sets variables according to platform.


### `vault_client::service`



## DefinedTypes

### `vault_client::cert_service`



#### Parameters

##### `base_path`

* Type: `String`

##### `common_name`

* Type: `String`

##### `role`

* Type: `String`

##### `alt_names`

* Type: `Array[String]`
* Default: `[]`

##### `ip_sans`

* Type: `Array[String]`
* Default: `[]`

##### `uid`

* Type: `Integer`
* Default: `0`

##### `gid`

* Type: `Integer`
* Default: `0`

##### `key_type`

* Type: `String`
* Default: `'rsa'`

##### `key_bits`

* Type: `Integer`
* Default: `2048`

##### `frequency`

* Type: `Integer`
* Default: `86400`

##### `exec_post`

* Type: `Array`
* Default: `[]`


### `vault_client::secret_service`



#### Parameters

##### `secret_path`

* Type: `String`

##### `field`

* Type: `String`

##### `dest_path`

* Type: `String`

##### `uid`

* Type: `Integer`
* Default: `0`

##### `gid`

* Type: `Integer`
* Default: `0`

##### `user`

* Type: `String`
* Default: `'root'`

##### `group`

* Type: `String`
* Default: `'root'`

##### `exec_post`

* Type: `Array`
* Default: `[]`
