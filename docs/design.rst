.. _design:

Design of Tarmak
================

Tarmak is a toolkit for Kubernetes cluster provisioning and management. It
focuses on best practice cluster security and cluster management/operation. It
has been built to be cloud provider-agnostic and hence provides a means for
consistent and reliable cluster deployment and management, across clouds and
on-premise environments (metal/private cloud).

Goals
-----

* Build and manage as similar as possible cluster deployments across different
  Cloud and on-premise environments

* Combine tried-and-tested tools throughout the stack to provide
  production-ready and ready-to-use clusters

* Follow security best practices

* Support for a fully automated CI/CD operation

* Provide minimal invasive upgrades, which can be predicted using dry runs

* Have a testable code base, that follows KISS and DRY. Don't use convoluted
  bash scripts per environment and operating system

* Provide a tool independent CLI experience, that simplifies common tasks and
  allows to investigate the status of the cluster quick and easy

* Allow customisation of parts of the code, to follow internal standards

* Use battle tested concepts to run components

Non-Goals
---------

* Reinvent the wheel

Architecture
---------------------

The architecture how clusters are build follows a couple of concepts that
turned out to be beneficial to operate Kubernetes clusters

Namespaces & Clusters & Hub
***************************

A namespace can contain one or more kubernetes clusters. Every namespace
contains exactly one bastion node and one vault cluster. In a multi-cluster
namespace these bastion and vault instances are run in an separate cluster
called  `hub`.

For a single node namespace, there's always exactly one cluster. This cluster contains all the server pool for the types:

* etcd: Stateful instances with etcd key-value store backing Kubernetes and possible overlay networks
* master: Stateless Kubernetes master instances
* node: Stateless Kubernetes node (aka worker) instances
* vault: Stateful vault instances, that contain the clusters PKI
* bastion: Bastion instance with public IP address to reach all other nodes, that don't have public IP space assigned

Server Pools
************

Server pools abstract instances of the same type. Every server pool has a type
attached that defines it's role. The ServerPool abstraction allows to create
multiple groups of instances of the same type. This is especially useful for
node/worker type instances

Tools used under the bonnet
---------------------------

Tarmak uses different tools for different problem areas of concern. It acts as
glue between the various tools.

Docker
******

Docker is used to package the tools necessary and run them in an uniform
environment across different operating system. This allows use to support Linux
and Mac OS X right and potentially Windows in the future

Terraform
*********

Terraform is a well known tool for infrastructure provisioning in public and
private clouds. We use terrraform to manage the lifecycle of resources in them
and store the state of clusters in Terraform remote state

Puppet
******

As soon as instances are spun up, Tarmak uses Puppet to configure them once.
Puppet is ran masterless, to not have to deal with the complexity of a Puppet
master setup. All the services are configured in a way that the instance from
now can run without any involvement of Puppet.

The reason for choosing Puppet over other means of configuration (Like bash
scripts, ansible, chef), was its testability on various levels and also the
concept of defining dependencies explicit, that allows to build a tree of
dependencies which helps to predict the changes within a dry-run.

Systemd
*******

Systemd units are used to maintain the dependencies between services. It also
ensures the certificate renewal is happening regularly using systemd timers
