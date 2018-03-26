.. _introduction:

Introduction
============

What is Tarmak?
---------------

Tarmak is a toolkit for Kubernetes cluster lifecycle management. It focuses on
best practice cluster security and cluster management/operation. It has been
built from the ground-up to be cloud provider-agnostic and hence provides a
means for consistent and reliable cluster deployment and management, across
clouds and on-premises environments.

Tarmak and its underlying components are the product of Jetstack_'s work with
its customers to build and deploy Kubernetes in production at scale.

.. _Jetstack: https://www.jetstack.io/

Design goals
------------

Goals
*****

* Build and manage as similar as possible cluster deployments across different
  cloud and on-premises environments.

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

Non-goals
*********

* Reinventing the wheel

.. _architecture_overview:

Architecture overview
---------------------

.. todo::
   A high-level architecture diagram is coming soon!

Tarmak configuration resources
******************************

The Tarmak configuration uses Kubernetes' API tooling and consists of various
different resources. While the Tarmak specific resources (Providers_ and
Environments_) are defined by the Tarmak project, Clusters_ are derived from a
draft version of the `Cluster API
<https://github.com/kubernetes/community/tree/master/wg-cluster-api>`_. This is
a community effort to have a standardised way of defining Kubernetes clusters.
By default, Tarmak configuration is located in ``~/.tarmak/tarmak.yaml``.

.. note::
   Although we do not anticipate breaking changes in our configuration, at this
   stage this cannot be absolutely guaranteed. Through the use of the
   Kubernetes API tooling, we have the option of migrating between different
   versions of the configuration in a controlled way.

.. _providers_resource:

Providers
^^^^^^^^^

A Provider contains credentials for and information about cloud provider
accounts. A single Provider can be used for many Environments, while every
Environment has to be associated with exactly one Provider.

Currently, the only supported Provider is **Amazon**. An Amazon Provider object
references credentials to make use of an AWS account to provision resources.

.. _environments_resource:

Environments
^^^^^^^^^^^^

An Environment consists of one or more Clusters. If an Environment has exactly
one cluster, it is called a *Single Cluster Environment*. A Cluster in such an
environment also contains the Environment-wide tooling.

For *Multi-Cluster Environments*, these components are placed in a special
``hub`` Cluster resource. This enables reuse of bastion and Vault nodes
throughout all Clusters.

.. _clusters_resource:

Clusters
^^^^^^^^

A Cluster resource represents exactly one Kubernetes cluster. The only
exception being the ``hub`` in a Multi Cluster Environment. Hubs do not contain
a Kubernetes cluster, as they are just where the Environment-wide tooling is
placed.

All instances in a Cluster are defined by an InstancePool_.

.. _Stacks:

Stacks
^^^^^^

The Cluster-specific Terraform code is broken down into separate,
self-contained Stacks. Stacks share Terraform outputs via the remote Terraform
state. Some Stacks depend on others, so the order in which they are provisioned
is important. Tarmak currently uses the following Stacks to build environments:


* ``state``: contains the stateful resources of the Cluster (data stores,
  persistent disk volumes)
* ``network``: sets-up the necessary network objects to allow communication
* ``tools``: contains the Environment-wide tooling, like bastion and
  CI/CD instances
* ``vault``: spins up a Vault cluster, backed by a Consul key-value store
* ``kubernetes``: contains Kubernetes' master, worker and etcd instances

.. figure:: static/providers-environments-clusters.png
   :alt: Config resources architecture: Providers, Environments and Clusters

   This is what a single cluster, production setup might look like. While the
   dev environment allows for multiple clusters (e.g. each with different
   features and/or team members), the staging and production environments
   consist of a single cluster each. The same AWS account is used for the dev
   and staging environment, while production runs in separate account.

.. _InstancePool:

InstancePools
^^^^^^^^^^^^^

Every Cluster contains InstancePools that group instances of a similar type
together. Every InstancePool has a name and role attached to it. Other
parameters allow us to customise the instances regarding size, count and
location.

These roles are defined:

* ``bastion``: Bastion instance within the ``tools`` stack. Has a public IP
  address and allows Tarmak to connect to other instances that only have
  private IP addresses.
* ``vault``: Vault instance within the ``vault`` stack. Has persistent disks,
  that back a Consul cluster, which backs Vault itself.
* ``etcd``: Stateful instances within ``kubernetes`` stack. etcd is the
  key-value store backing Kubernetes and potentially other components, overlay
  networks such as Calico for example.
* ``master``: Stateless Kubernetes master instances.
* ``worker``: Stateless Kubernetes worker instances.


Tools used under the hood
-------------------------

Tarmak is backed by tried-and-tested tools, which act as the glue and
automation behind the Tarmak CLI interface. These tools are plugable, but at
this stage we use the following:

Docker
******

Docker is used to package the tools necessary and run them in a uniform
environment across different operating systems. This allows Tarmak to run on
Linux and macOS (as well as potentially Windows in the future).

Packer
******

Packer helps build reproducible VM images for various environments. Using
Packer, we build custom VM images containing the latest kernel upgrades and
supported puppet version.

Terraform
*********

Terraform is a well-known tool for infrastructure provisioning in public and
private clouds. We use Terraform to manage the lifecycle of resources and store
cluster state.

Puppet
******

As soon as instances are spun up, Tarmak uses Puppet to configure them. Puppet
is used in a 'masterless' architecture, so as to avoid the complexity of a full
Puppet master setup. All the services are configured in such a way that, once
converged, the instance can run independently of Puppet.

Why Puppet over other means of configuration (i.e. bash scripts, Ansible,
Chef)? The main reason is its testability (at various levels) as well as the
concept of explicit dependency definition (allowing a tree of dependencies to
be built helping predict the changes with a dry-run).

Systemd
*******

Tarmak uses Systemd units and timers. Units are used to maintain the
dependencies between services while timers enable periodic application
execution - e.g. for certificate renewal.
