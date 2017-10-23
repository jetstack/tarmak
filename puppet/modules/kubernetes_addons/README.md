# kubernetes_addons

#### Table of Contents

1. [Description](#description)
2. [Classes](#classes)
## Description
This module is part of [Tarmak](http://docs.tarmak.io) and should currently be considered alpha.

[![Travis](https://img.shields.io/travis/jetstack/puppet-module-kubernetes_addons.svg)](https://travis-ci.org/jetstack/puppet-module-kubernetes_addons/)

## Classes

### `kubernetes_addons`

Class: kubernetes_addons


### `kubernetes_addons::cluster_autoscaler`



#### Parameters

##### `image`

* Type: `String`
* Default: `'gcr.io/google_containers/cluster-autoscaler'`

##### `version`

* Type: `String`
* Default: `''`

##### `limit_cpu`

* Type: `String`
* Default: `'200m'`

##### `limit_mem`

* Type: `String`
* Default: `'500Mi'`

##### `request_cpu`

* Type: `String`
* Default: `'100m'`

##### `request_mem`

* Type: `String`
* Default: `'300Mi'`

##### `min_instances`

* Type: `Integer`
* Default: `3`

##### `max_instances`

* Type: `Integer`
* Default: `6`

##### `ca_mounts`

* Type: `Any`
* Default: `$::kubernetes_addons::params::ca_mounts`

##### `cloud_provider`

* Type: `Any`
* Default: `$::kubernetes_addons::params::cloud_provider`

##### `aws_region`

* Type: `Any`
* Default: `$::kubernetes_addons::params::aws_region`


### `kubernetes_addons::dashboard`



#### Parameters

##### `image`

* Type: `String`
* Default: `'gcr.io/google_containers/kubernetes-dashboard-amd64'`

##### `version`

* Type: `String`
* Default: `'v1.5.1'`

##### `limit_cpu`

* Type: `String`
* Default: `'100m'`

##### `limit_mem`

* Type: `String`
* Default: `'128Mi'`

##### `request_cpu`

* Type: `String`
* Default: `'10m'`

##### `request_mem`

* Type: `String`
* Default: `'64Mi'`

##### `replicas`

* Type: `Any`
* Default: `undef`


### `kubernetes_addons::default_backend`



#### Parameters

##### `image`

* Type: `Any`
* Default: `$::kubernetes_addons::params::default_backend_image`

##### `version`

* Type: `Any`
* Default: `$::kubernetes_addons::params::default_backend_version`

##### `request_cpu`

* Type: `Any`
* Default: `$::kubernetes_addons::params::default_backend_request_cpu`

##### `request_mem`

* Type: `Any`
* Default: `$::kubernetes_addons::params::default_backend_request_mem`

##### `limit_cpu`

* Type: `Any`
* Default: `$::kubernetes_addons::params::default_backend_limit_cpu`

##### `limit_mem`

* Type: `Any`
* Default: `$::kubernetes_addons::params::default_backend_limit_mem`

##### `namespace`

* Type: `Any`
* Default: `$::kubernetes_addons::params::namespace`

##### `replicas`

* Type: `Any`
* Default: `undef`


### `kubernetes_addons::elasticsearch`



#### Parameters

##### `namespace`

* Type: `String`
* Default: `$::kubernetes_addons::params::namespace`

##### `image`

* Type: `String`
* Default: `'gcr.io/google_containers/elasticsearch'`

##### `version`

* Type: `String`
* Default: `'v2.4.1-1'`

##### `persistent_storage`

* Type: `Boolean`
* Default: `false`

##### `persistent_storage_request`

* Type: `String`
* Default: `'20Gi'`

##### `persistent_storage_class`

* Type: `String`
* Default: `'fast'`

##### `request_cpu`

* Type: `String`
* Default: `'100m'`

##### `request_mem`

* Type: `String`
* Default: `'512Mi'`

##### `limit_cpu`

* Type: `String`
* Default: `'1000m'`

##### `limit_mem`

* Type: `String`
* Default: `'2048Mi'`

##### `node_port`

* Type: `Integer[0,65535]`
* Default: `0`

##### `replicas`

* Type: `Integer`
* Default: `2`


### `kubernetes_addons::fluentd_elasticsearch`



#### Parameters

##### `namespace`

* Type: `String`
* Default: `$::kubernetes_addons::params::namespace`

##### `image`

* Type: `String`
* Default: `'gcr.io/google_containers/fluentd-elasticsearch'`

##### `version`

* Type: `String`
* Default: `'1.22'`

##### `request_cpu`

* Type: `String`
* Default: `'200m'`

##### `request_mem`

* Type: `String`
* Default: `'384Mi'`

##### `limit_cpu`

* Type: `String`
* Default: `'100m'`

##### `limit_mem`

* Type: `String`
* Default: `'256Mi'`


### `kubernetes_addons::grafana`



#### Parameters

##### `image`

* Type: `Any`
* Default: `$::kubernetes_addons::params::grafana_image`

##### `version`

* Type: `Any`
* Default: `$::kubernetes_addons::params::grafana_version`


### `kubernetes_addons::heapster`



#### Parameters

##### `image`

* Type: `Any`
* Default: `$::kubernetes_addons::params::heapster_image`

##### `version`

* Type: `Any`
* Default: `$::kubernetes_addons::params::heapster_version`

##### `cpu`

* Type: `Any`
* Default: `$::kubernetes_addons::params::heapster_cpu`

##### `mem`

* Type: `Any`
* Default: `$::kubernetes_addons::params::heapster_mem`

##### `extra_cpu`

* Type: `Any`
* Default: `$::kubernetes_addons::params::heapster_extra_cpu`

##### `extra_mem`

* Type: `Any`
* Default: `$::kubernetes_addons::params::heapster_extra_mem`

##### `nanny_image`

* Type: `Any`
* Default: `$::kubernetes_addons::params::heapster_nanny_image`

##### `nanny_version`

* Type: `Any`
* Default: `$::kubernetes_addons::params::heapster_nanny_version`

##### `nanny_request_cpu`

* Type: `Any`
* Default: `$::kubernetes_addons::params::heapster_nanny_request_cpu`

##### `nanny_request_mem`

* Type: `Any`
* Default: `$::kubernetes_addons::params::heapster_nanny_request_mem`

##### `nanny_limit_cpu`

* Type: `Any`
* Default: `$::kubernetes_addons::params::heapster_nanny_limit_cpu`

##### `nanny_limit_mem`

* Type: `Any`
* Default: `$::kubernetes_addons::params::heapster_nanny_limit_mem`

##### `sink`

* Type: `Any`
* Default: `undef`


### `kubernetes_addons::influxdb`



#### Parameters

##### `image`

* Type: `Any`
* Default: `$::kubernetes_addons::params::influxdb_image`

##### `version`

* Type: `Any`
* Default: `$::kubernetes_addons::params::influxdb_version`


### `kubernetes_addons::kibana`



#### Parameters

##### `namespace`

* Type: `String`
* Default: `$::kubernetes_addons::params::namespace`

##### `image`

* Type: `String`
* Default: `'gcr.io/google_containers/kibana'`

##### `version`

* Type: `String`
* Default: `'v4.6.1-1'`

##### `request_cpu`

* Type: `String`
* Default: `'50m'`

##### `request_mem`

* Type: `String`
* Default: `'768Mi'`

##### `limit_cpu`

* Type: `String`
* Default: `'1'`

##### `limit_mem`

* Type: `String`
* Default: `'2Gi'`

##### `replicas`

* Type: `Integer`
* Default: `2`


### `kubernetes_addons::kube2iam`



#### Parameters

##### `base_role_arn`

* Type: `String`
* Default: `''`

##### `namespace`

* Type: `String`
* Default: `'kube-system'`

##### `image`

* Type: `String`
* Default: `'jtblin/kube2iam'`

##### `version`

* Type: `String`
* Default: `'0.6.5'`

##### `request_cpu`

* Type: `String`
* Default: `'0.1'`

##### `request_mem`

* Type: `String`
* Default: `'64Mi'`

##### `limit_cpu`

* Type: `String`
* Default: `''`

##### `limit_mem`

* Type: `String`
* Default: `'256Mi'`


### `kubernetes_addons::nginx_ingress`



#### Parameters

##### `image`

* Type: `Any`
* Default: `$::kubernetes_addons::params::nginx_ingress_image`

##### `version`

* Type: `Any`
* Default: `$::kubernetes_addons::params::nginx_ingress_version`

##### `request_cpu`

* Type: `Any`
* Default: `$::kubernetes_addons::params::nginx_ingress_request_cpu`

##### `request_mem`

* Type: `Any`
* Default: `$::kubernetes_addons::params::nginx_ingress_request_mem`

##### `limit_cpu`

* Type: `Any`
* Default: `$::kubernetes_addons::params::nginx_ingress_limit_cpu`

##### `limit_mem`

* Type: `Any`
* Default: `$::kubernetes_addons::params::nginx_ingress_limit_mem`

##### `namespace`

* Type: `Any`
* Default: `$::kubernetes_addons::params::namespace`

##### `replicas`

* Type: `Any`
* Default: `undef`

##### `host_port`

* Type: `Any`
* Default: `false`


### `kubernetes_addons::params`




### `kubernetes_addons::tiller`



#### Parameters

##### `image`

* Type: `String`
* Default: `'gcr.io/kubernetes-helm/tiller'`

##### `version`

* Type: `String`
* Default: `'v2.6.1'`
