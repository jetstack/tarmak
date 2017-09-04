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

* Provide a tool independet CLI experience, that simplifies common tasks and
  allows to investigate the status of the cluster quick and easy

* Allow customisation of parts of the code, to follow internal standards

* Use battle tested concepts to run components

Non-Goals
---------

* Reinvent the wheel

Architecture
---------------------

Namespaces & Clusters
*********************

Server Pools
************

Tools used under the bonnet
---------------------------

Docker
******

Terraform
*********

Puppet
******

Systemd
*******
