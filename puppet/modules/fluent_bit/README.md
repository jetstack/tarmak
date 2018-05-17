# fluent_bit

#### Table of Contents

1. [Description](#description)
2. [Classes](#classes)
3. [Defined Types](#defined-types)
## Description
This module is part of [Tarmak](http://docs.tarmak.io) and should currently be considered alpha.

[![Travis](https://img.shields.io/travis/jetstack/puppet-module-kubernetes.svg)](https://travis-ci.org/jetstack/puppet-module-kubernetes/)

## Classes

### `fluent_bit`



#### Parameters

##### `package_name`

* Type: `Any`
* Default: `$::fluent_bit::params::package_name`

##### `service_name`

* Type: `Any`
* Default: `$::fluent_bit::params::service_name`


### `fluent_bit::config`




### `fluent_bit::daemonset`



#### Parameters

##### `fluent_bit_image`

* Type: `String`
* Default: `'fluent/fluent-bit'`

##### `fluent_bit_version`

* Type: `String`
* Default: `'0.13.1'`

##### `platform_namespaces`

* Type: `Array[String]`
* Default: `['kube-system','service-broker','monitoring']`


### `fluent_bit::install`




### `fluent_bit::params`




### `fluent_bit::service`



## DefinedTypes

### `fluent_bit::output`



#### Parameters

##### `config`

* Type: `Hash`
* Default: `{}`
