.. vim:set ft=rst spell:

Managing SSH Known Hosts
========================

This proposal suggests a method to more securely manage and ensure
authentication of remote hosts through the SSH protocol within Tarmak
environments on AWS.

Background
----------

Currently we solely use the external OpenSHH SSH command line tool to connect to
remote instances on EC2 for both interactive shell sessions as well as
tunnelling and proxy commands for other services, including, connecting to the
Vault cluster and the private Kubernetes API server endpoint. Currently in
development is the replacement of our programmatic use cases of SSH in favour of
the in package Go solution, a choice stemming from pain points developing more
sophisticated utility functions for Tarmak and the desire for improvements in
control of connections to remote hosts.

During development of this replacement it became clear that proper care must be
taken during authentication of host public keys during connection and manual
management of our ``ssh_known_hosts`` cluster file. Our current implementation
allows OpenSSH to maintain this file however, does not exit with an error if
public keys do not match due to the flag ``StrictHostKeyChecking`` set to
``no``. Not only does a miss-match in public keys not cause an error, the
population of known public keys on different authenticated machines to the same
EC2 hosts will always use the hosts presented public key, meaning the set of
public keys could potentially be different for users accessing the same cluster.

Objective
---------

By implementing stricter enforcement of the ``ssh_known_hosts`` file and passing
it's management to Tarmak, we can improve the security of SSH connections to
remote hosts. The key high level points to achieving this is as follows:

 - Disable writes from the OpenSSH command to the ``ssh_known_hosts`` file and
   enforce strict checking.
 - Enforce that our in package implementation of SSH connections adheres to this
   file also.
 - Collect public keys during instance start up that are then stored, tightly
   coupled with that host. These keys are able to be used as a source of truth
   for other authenticated users attempting to connect to remote hosts on the
   cluster that have empty or an incomplete ``ssh_known_hosts`` file.

Changes
-------

Firstly, we must restrict the OpenSSH command line tool from editing the
``ssh_known_hosts`` file and strictly enforce it by updating the generator for the
``ssh_config`` file. This enables Tarmak to take control of the ``ssh_known_hosts``
file management.

In order to create a source of truth for each host's public key, each instance
will have it's public key attached as a tag, shortly after boot time like the
following:

+------------+------------------------------------------------------------------------------------------------------------------------------------------------------------------+
| public_key | ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBPF+xkIGMUNVI0gElRaTLjfA4QMN/XGJhHswDyv59DNSOtG3KwZvDF3YkAb0PkTQAYo8N5fxoKqimGugOAaefPc= |
+------------+------------------------------------------------------------------------------------------------------------------------------------------------------------------+

The population of this tag will occur during a Terraform apply in which the
instance was created. A new custom Terraform resource should be created -
``tarmak_ssh_public_key_tag`` - that is dependant on the AWS instance Terraform
resource. This new resource will be handled by the current Tarmak Terraform
provider which is to be extended to consume it. Once called for creation, it
shall attempt to create the initial SSH connection to this host which will
provide Tarmak with it's public key. Once acquired, it will attach this public
key to the AWS instance and add it to the ``ssh_known_hosts`` file.

All other SSH connections will rely on the contents of the ``ssh_known_hosts``
file however, in the case the host is not present in the file, will attempt to
use the AWS instance's ``public_key`` tag to populate it's entry.

Notable items
-------------

Care should be taken when waiting for the instance to become ready to create the
initial SSH connection for each host. It is important not to make this a
bottleneck during the Terraform apply.

Out of scope
------------

We should not disrupt the current flow of key generation on the host instances
such as using key injection. At no point should private keys be in flight.

We should not store or rely on the public key being stored in the Terraform
state as this would require all commands that rely on SSH, to also rely on
fetching and updating the Terraform state - significantly increasing completion
time for even trivial tasks.
