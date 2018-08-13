.. vim:set ft=rst spell:

SPIFFE/SPIRE Integration
========================

This proposal investigates and evaluates the feasibility of integrating
SPIFFE/SPIRE into Tarmak. The integration of SPIFFE/SPIRE would be for the
intention of making the Tarmak cluster, as a whole, more secure.

Background
----------

Currently the Tarmak cluster relies on Vault to create and maintain a PKI
between components within the cluster. Although functional, this system relies
on ``init-tokens`` to bootstrap the authentication procedure. This provides
only a single form of authentication for acceptance of a component into the
trusted PKI.

An alternative to Vault along with the other peripheral utility tools
(vault-helper, vault-unsealer) currently in use is the adoption of spire_. Spire
is a runtime environment that implements the SPIFFE specification. SPIFFE aims
to provide authenticated identity to individual work loads through a unique,
standardised identifier rather than architectural identifiers such as an IP
address; no longer reliable or suitable for distributed infrastructures
like Kubernetes.

SPIFFE provides URIs comprised of a 'trust domain' followed by an unique
arbitrary path that identifies some resource, user or object that some trust is
to be established, e.g:

.. _spire: https://github.com/spiffe/spire

::

    spiffe://trust-domain/some/path/to/workload

Trust domains represent the CA and root trust of that domain. With trust
established with the trust domain, the identity receives a SPIFFE Verifiable
Identity Document (SVID). This document includes the SPIFFE ID, the identity's
private key and a bundle of certificates to establish trust between other
identities in the domain, including the trust domain.

Spires strengths come in the form of alternative bootstrapping processes that
differ to that of Vault's init-tokens. Instead of these, workloads are able to
prove identity through using other verifiable identifiers such as use of kernel
primitives, cloud provider labels and even Kubernetes primitives. Used in
conjunction, multiple identifiers provide a greater amount of security within
verifiable, authentication of identity, not just a single leak-able token.

Objective
---------

Use Spire within the Tarmak cluster to bolster the authentication of workload
identities, resulting in a more secure cluster. This would ultimately mean a
complete replacement of Vault and it's periphery tools, in favour of Spire and
it's stack.

Changes
-------

Spire works to a server/agent architecture. As such, the Vault HA cluster will
be replaced with a Spire HA server cluster. This means the Terraform Vault
provider will also be replaced with a Spire equivalent.

Spire agents will then sit on each Kubernetes instance that are able to connect
to the Spire server using AWS identifiers. These agents, once authenticated with
the Spire server cluster are then able to set up PKI within the instances
through use of Kubernetes primitives to ensure identity. A spire-helper type
tool will need to be developed to setup PKI, just like how vault-helper manages
this.

Steps as I currently see them (Subject to Change):

- Set up Spire HA server cluster.
- Create Spire Terraform provider.
- Authenticate Spire agents on AWS instances against the server cluster.
- Create spire-helper type tool.
- Set up PKI with Kubernetes components using Spire agents.

Useful Plugins
--------------
- AWS Node Attestor: https://github.com/spiffe/aws-iid-attestor.
  This plugin provides functionality for AWS instances to authenticates
  themselves using AWS meta data. This is required for Spire agents to
  authenticate against the spire server cluster.

- K8s Workload Attestor: https://github.com/spiffe/spire/blob/master/doc/plugin_agent_workloadattestor_k8s.md.
  Plugin that uses Kubernetes primitives to authenticate workloads against
  SPIFFE agents.

- Unix Workload Attestor: https://github.com/spiffe/spire/blob/master/doc/plugin_agent_workloadattestor_unix.md.
  Plugin that uses Unix primitives to authenticate workloads against SPIFFE
  agents.

- SQL Data Store: https://github.com/spiffe/spire/blob/master/doc/plugin_server_datastore_sql.md.
  Plugin used as a data backend for Spire using an SQL database solution.

Notable items
-------------

- This represents a significant amount of implementation time.
- Spire is still in an early stage meaning it hasn't been as industry tested in
  production like Vault has been.
- Spire is still in very much heavy development so breaking changes are highly
  probable.
- It may just end up not working and be incompatible with Tarmak clusters.

Out of scope
------------

Avoiding writing our own plugins. This is better left to experts and if is
required, Spire should be dropped due to not being ready yet for our
requirements.
