# Prometheus

#### Table of Contents

1. [Description](#description)
1. [Setup - The basics of getting started with prometheus](#setup)
    * [What prometheus does](#what-prometheus-affects)
    * [Setup requirements](#setup-requirements)
    * [Beginning with prometheus](#beginning-with-prometheus)
1. [Usage - Configuration options and additional functionality](#usage)
1. [Reference - An under-the-hood peek at what the module is doing and how](#reference)
1. [Limitations - OS compatibility, etc.](#limitations)
1. [Development - Guide for contributing to the module](#development)

## Description

This module installs [Prometheus](https://www.prometheus.io/) on a
Kubernetes cluster and was initially developed to do this on CentOS 7 on AWS EC2.

## Setup

### What prometheus does

On a master node:
* Installs Prometheus with local storage via a deployment and exposes it on a nodePort
* **Optionally** Installs Prometheus node-exporter as a daemonset
* **Optionally** Installs kube-state-metrics as a deployment
* **Optionally** adds a default configuration that will discover endpoints and integrate with puppernetes
* **Optionally** adds some default rules to monitor the cluster

On an etcd node:
* **Can** be used to deploy node-exporter as a systemd service / docker container (requiring docker on the node)
* **Can** be used to deploy a customised blackbox exporter as a systemd service, providing an authenticated metrics proxy to etcd

### Setup Requirements

A working Kubernetes cluster, and a list of etcd endpoints if wishing to monitor this

### Beginning with prometheus

```puppet

#master
class { 'prometheus':
  role         => 'master',
  etcd_cluster => [ 'etcd1.mydomain', 'etcd3.mydomain', 'etcd3.mydomain' ]
}

#etcd
class { 'prometheus':
  role => 'etcd'
}

```

## Usage

This section is where you describe how to customize, configure, and do the
fancy stuff with your module here. It's especially helpful if you include usage
examples and code samples for doing things with your module.

## Reference

Here, include a complete list of your module's classes, types, providers,
facts, along with the parameters for each. Users refer to this section (thus
the name "Reference") to find specific details; most users don't read it per
se.

## Limitations

This is where you list OS compatibility, version compatibility, etc. If there
are Known Issues, you might want to include them under their own heading here.

## Development

Since your module is awesome, other users will want to play with it. Let them
know what the ground rules for contributing are.

## Release Notes/Contributors/Etc. **Optional**

If you aren't using changelog, put your release notes here (though you should
consider using changelog). You can also add any additional sections you feel
are necessary or important to include here. Please use the `## ` header.
