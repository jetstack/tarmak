# aws_es_proxy

#### Table of Contents

1. [Description](#description)
2. [Classes](#classes)
3. [Defined Types](#defined-types)
## Description
This module is part of [Tarmak](http://docs.tarmak.io) and should currently be considered alpha.

[![Travis](https://img.shields.io/travis/jetstack/puppet-module-kubernetes.svg)](https://travis-ci.org/jetstack/puppet-module-kubernetes/)

## Classes

### `aws_es_proxy`



#### Parameters

##### `version`

* Type: `Any`
* Default: `$::aws_es_proxy::params::version`

##### `download_url`

* Type: `Any`
* Default: `$::aws_es_proxy::params::download_url`

##### `dest_dir`

* Type: `Any`
* Default: `$::aws_es_proxy::params::dest_dir`


### `aws_es_proxy::install`




### `aws_es_proxy::params`



## DefinedTypes

### `aws_es_proxy::instance`



#### Parameters

##### `dest_address`

* Type: `String`

##### `tls`

* Type: `Boolean`
* Default: `true`

##### `dest_port`

* Type: `Integer`
* Default: `9200`

##### `listen_port`

* Type: `Integer`
* Default: `9200`
