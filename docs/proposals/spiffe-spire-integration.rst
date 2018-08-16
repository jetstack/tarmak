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
to provide authenticated identity to individual workloads through a unique,
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
primitives, cloud provider meta data and even Kubernetes primitives. Used in
conjunction, multiple identifiers provide a greater amount of security within
verifiable, authentication of identity, not just a single leak-able token.

Objective
---------

Use Spire within the Tarmak cluster to bolster the authentication of workload
identities, resulting in a more secure cluster. This could ultimately mean a
large or complete replacement of Vault and it's periphery tools, in favour of
Spire and it's stack.

Changes
-------

Spire works to a server/slave architecture. As such, the Vault HA cluster will
be replaced with a Spire HA server cluster. This means the Terraform Vault
provider will also be replaced with a Spire equivalent.

Spire agents will then sit on each Kubernetes instance that are able to connect
to the Spire server using AWS identifiers. These agents, once authenticated with
the Spire server cluster are then able to set up PKI within the instances
through use of attestor plugins with suitable selectors to authenticate
identity. A spire-helper type tool will need to be developed to setup PKI, just
like how vault-helper manages this.

Steps as I currently see them (Subject to Change):

- Set up Spire HA server cluster.
- Create Spire Terraform provider.
- Authenticate Spire agents on AWS instances against the server cluster.
- Create spire-helper type tool.
- Set up PKI with Kubernetes components using Spire agents.

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

Testing Review
--------------
I was able to create Spire servers on each Vault instance. These all connect to
a single PostgreSQL elected master as a backend. PostgreSQL replicas reside on
each Vault instance that promote a leader through Consul.

Spire agents reside on each Kubernetes instance that are able to attest through
their AWS identity meta data. No token is required, they can simply connect to a
spire sever and it's done!

SVIDs can be generated for each Kubernetes component that needs them through
first registering components with kernel primitive selectors (uid, gid) against the
server(s). The components can then request their SVID against the agent and receive
their key and certificate material.

Problems During Testing
-----------------------

Spire is not currently well built for HA and really only expects a single server
in cluster. PostgreSQL support for HA is quite poor from what I have found,
whilst also being the only supported "HA" backend. I was able to use a tool Patroni_ that
facilitates a PostgreSQL cluster with Consul however only the master is in write
mode. All replicas on other instances are in read mode which causes spire severs
on those instances to panic. Spire servers have to connect to the PostgreSQL
cluster master elected through Consul.

Spire uses Elliptic Curve keys which is odd and can be annoying.

.. _Patroni: https://github.com/zalando/patroni

Road Map
--------

High availability support is targeted to be included at the end of October 2018.
Perhaps when this has been included a re-review of the inclusion of Spire should
be considered.
