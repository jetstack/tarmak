.. vim:set ft=rst spell:

Terraform Provider
==================

This proposal suggests how to approach the implementation of a Terraform
provider for *Tarmak*, to make Tarmak <-> Terraform interactions within a
Terraform run more straightforward.

Background
----------

Right now the Terraform code for the AWS provider (the only one implemented)
consists of multiple separate stacks (state, network, tools, vault,
kubernetes). The Main reason for having these stacks is to enable Tarmak to do
operations in between parts of the resource spin up. Examples for such
operations are (*list might not be complete*):

- Bastion node needs to be up for other instances to check into Wing.
- Vault needs to be up and initialised, before PKI resources are created.
- Vault needs to contain a cluster's PKI resources, before Kubernetes instances
  can be created (``init-tokens``).

The separation of stacks comes with some overhead for preparing Terraform apply
(pull state, lock stack, plan run, apply run). Terraform can't make use of
parallel creation of resources that are independent from each other.


Objective
---------

An integration of these stacks into a single stack could lead to a substantial
reduction of execution time.

As Terraform is running in a container is quite isolated from the Tarmak
process:

* Requires some Terraform refactoring
* Should be done before implementing multiple providers

Changes
-------

Terraform code base
*******************

Terraform resources
*******************

The proposal is to implement a Tarmak provider_ for Terraform, with at least these
three resources.

.. _provider: https://www.terraform.io/docs/plugins/provider.html

``tarmak_bastion_instance``
~~~~~~~~~~~~~~~~~~~~~~~~~~~

A bastion instance

::

  Input:
  - Bastion IP address or hostname
  - Username for SSH

  Blocks until Wing API server is running.

``tarmak_vault_cluster``
~~~~~~~~~~~~~~~~~~~~~~~~

A vault cluster

::

  Input:
  - List of Vault internal FQDNs or IP addresses

  Blocks until Vault is initialised & unsealed.


``tarmak_vault_instance_role``
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

This creates (once per role) an init token for such instances in Vault. 

::

  Input:
  - Name of Vault cluster
  - Role name

  Output:
  - init_token per role


  Blocks until init token is setup.



Notable items
-------------


Communication with the process in the container
***********************************************

The main difficulty is communication with the Tarmak process, as
Terraform is run within a Docker container with no communication available to
the main Tarmak process (stdin/out is used by the Terraform main process).

The proposal suggests that all ``terraform-provider-tarmak`` resources block
until the point when the main Tarmak process connects using another exec to a
so-called ``tarmak-connector`` executable that speaks via a local Unix socket
to the ``terraform-provider-tarmak``.

This provides a secure and platform-independent channel between Tarmak and
``terraform-provider-tarmak``.

::

   <<Tarmak>>  -- launches -- <<terraform|container>> 

       stdIN/OUT -- exec into  ---- <<exec terraform apply>>
                                    <<subprocess terraform-provider-tarmak
                                        |
                                     connects
                                        |
                                    unix socket
                                        |
                                     listens
                                        |
       stdIN/OUT -- exec into  ---- <<exec tarmak-connector>>


The protocol on that channel should be using Golang's `net/rpc
<https://golang.org/pkg/net/rpc/>`_.

Initial proof-of-concept
************************

An initial `proof-of-concept
<https://gitlab.jetstack.net/christian.simon/terraform-provider-tarmak/tree/master>`_
has been explored to test what the Terraform model looks like. Although it's
not really having anything implemented at this point, it might serve as a
starting point.

Out of scope
------------

This proposal is not suggesting that we migrate features that are currently
done by the Tarmak main process. The reason for that is that we don't want
Terraform to become involved in the configuration provisioning of e.g. the
Vault cluster. This proposal should only improve the control we have from
Tarmak over things that happen in Terraform.
