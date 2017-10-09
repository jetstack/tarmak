.. _design:

Tarmak design
=============

Tarmak is a toolkit for Kubernetes cluster lifecycle management. It focuses on
best practice cluster security and cluster management/operation. It has been
built from the ground-up to be cloud provider-agnostic and hence provides a
means for consistent and reliable cluster deployment and management, across
clouds and on-premise environments.

Goals
-----

* Build and manage as similar as possible cluster deployments across different
  cloud and on-premise environments.

* Combine tried-and-tested and well-understood system tools throughout the
  stack to provide production-ready and ready-to-use clusters.

* Follow security best practices.

* Support for a fully automated CI/CD operation.

* Provide minimally invasive upgrades, which can be predicted using dry-runs.

* Have a testable code base, that follows KISS and DRY. For example, avoidance
  of convoluted bash scripts that are environment and operating
  system-specific.

* Provide a tool-independent CLI experience, that simplifies common tasks and
  allows to investigate cluster status and health quickly and easily.

* Allow customisation to parts of the code, to follow internal standards.

Non-Goals
---------

* Reinventing the wheel

Architecture
------------

The architecture how clusters are build follows a couple of concepts that
turned out to be beneficial to operate Kubernetes clusters

Tarmak configuration resources
******************************

.. figure:: providers-environments-clusters.png
   :alt: Config resources architecture: Providers, Environments & Clusters

* Providers:

* Environments: contains one or more Kubernetes clusters. Every Environment
  contains exactly one bastion node and one Vault cluster.

* Clusters: A

  Hub: In a multi-cluster Environment, the bastion and Vault instances are run
  in an separate cluster called the  `Hub`.

* Stacks

For a single-node Environment, there's always exactly one Cluster. This cluster
contains all the InstnacePools for the following system components:


InstancePools, Roles and Stacks
*******************************

* etcd: Stateful instances with etcd key-value store backing Kubernetes and
  possible overlay networks * master: Stateless Kubernetes master instances
* node: Stateless Kubernetes node (aka worker) instances
* vault: Stateful Vault instances, that back the cluster's PKI
* bastion: Bastion instance with public IP address to reach all other nodes
  (private IPs by default)

InstancePools group instances of the same type together. Every InstancePool has
a type attached that defines it's role. The InstancePool abstraction allows to
create multiple groups of instances of the same type. This is especially useful
for worker type instances, that should run with slightly modified parameters.

Tools used under the hood
-------------------------

Tarmak is backed by tried-and-tested tools, effectively acting as glue and
automation, managed by a CLI UX. These tools are plugable, but at this stage we
use the following:

Docker
******

Docker is used to package the tools necessary and run them in a uniform
environment across different operating systems. This allows Tarmak to be
supported on Linux and Mac OS X (and potentially Windows in the future).

Terraform
*********

Terraform is a well-known tool for infrastructure provisioning in public and
private clouds. We use Terraform to manage the lifecycle of resources and store
the state of clusters in Terraform remote state.

Puppet
******

As soon as instances are spun up, Tarmak uses Puppet to configure them.  Puppet
is used in a 'masterless' architecture, to not require the complexity of a full
Puppet master setup. All the services are configured in such a way that once
converged, the instance can run without any further involvement of Puppet.

Why Puppet over other means of configuration (i.e. bash scripts, Ansible,
Chef)? The main reason is its testability at various levels and also the
concept of explicit dependency definition (allowing a tree of dependencies to
be built which helps to predict the changes with a dry-run).

Systemd
*******

Systemd units are used to maintain the dependencies between services.

Systemd timers enable periodic application execution, such as for certificate renewal.
