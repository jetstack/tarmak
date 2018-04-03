.. vim:set ft=rst spell:

Custom Vault Auth Provider for EC2 Instances - A Proposal
=========================================================

This proposal suggests writing a custom authentication provider for Vault.
This provider would allow access privileges, especially to the PKI secrets
provider, to be locked down per EC2 Instance, which is not currently possible.

Background
==========

Glossary
--------

-  IAM Role (iam-role) - an IAM principal. Our instances have an IAM
   role attached which they use as a service account.
   https://docs.aws.amazon.com/IAM/latest/UserGuide/id\_roles.html
-  IAM Policy (iam-policy) - the authz rules (and rules documents) in
   AWS.
-  Vault Token (vault-token) - common identity document given by all
   vault auth providers. Has a list of policy (documents).
-  Vault Policy (vault-policy) - both the policy language itself (giving
   authorization for vault operations), document of those policy
   statement,s and the list of such documents named on a vault-token.
   vault-policy has its own type, which we'll call *policy*
-  Vault PKI Role (pki-role) - vault-policy isn't very powerful wrt to
   PKI operations, e.g. it doesn't let you lock PKI certificate issuance
   down to specific CNs etc. The PKI secrets provider has its own
   mechanism for this; the PKI role. pki-roles live at paths and access
   is controlled by vault-policy. pki-roles are implemented as vault
   generic secrets.
   https://www.vaultproject.io/api/secret/pki/index.html#create-update-role

Problem
-------

We would like to have individual Vault policies per EC2 instance, so
that they can't issue themselves certificates for anyone else's FQDN.

See:

- https://github.com/jetstack/tarmak/issues/120
- https://github.com/jetstack/tarmak/issues/34

Posibilities
------------

Due to several Vault limitations, especially in the free version, it's
not possible to:

- Add a vault-policy (name) to a token after it's been created (by vault ec2 login)
- Specify a vault-policy expressive enough to re-use one policy for all EC2 Instances (i.e. no variable back-references)
- Vault Premium has a different policy language called Sentinel, which looks like it could do this
- Use the k8s auth provider, as we also want to use this mechanism with etcd Instances, etc, which do not run Kubelet

Objective
=========

Solution
--------

Our proposed solution is Vault auth provider. This will auth EC2 instances using their IID like the current AWS auth provider does, but it will also generate new vault-policies and pki-roles on the fly, which will e.g. limit their cert creation power to just that instance's DNS name.

This vault auth plugin will serve exactly one environment (i.e. all kubernetes clusters in the same (provider, region))

On login, our provider will:

- Auth the Instance like the AWS provider's EC2 mode does (can we simply defer to that code?)
- Match the iam-role attached to the Instance against our Provider's config, using its ARN
- Make a pki-role from the configured template for that iam-role
- Make a vault-policy (document) from the configured template for that iam-role, including templating in the name of the pki-role

  - This should have a unique name based on the AWS Instance ID / boot ID / etc
- Like any other auth provider, ultimately make and return a vault-token (from a template?), including the unique vault-policy just made

Changes
=======

Configuration
-------------

Config of the provider will be though Vault paths as normal.

``config/client``
~~~~~~~~~~~~~~~~~

Global configuration of the plugin.

Fields:

::

    cloud_provider=[aws|gce]
    aws_access_id= (use instance role if empty)
    aws_secret_id= (use instance role if emtpy)
    aws_region=    (detect from metadata service if empty)
    gce_*          (equivalents...)

``roles/<rolename>``
~~~~~~~~~~~~~~~~~~~~

Tell Vault about the IAM roles in use by the instances.

Recall that each instance we bring up has an IAM role attached depending on its type, e.g. etcd, master, or worker. We can easily use this to tell the different instance types apart, as they need different policy templates.

Note: we have to store this information at a vault path, which means coming up with yet another set of symbolic names, and a name for this type of thing. I'll loosely call them "roles", as there's a 1:1 mapping with the IAM Roles they're modelling. Recall that the vault server is shared between clusters. We won't do any namespacing in the path, so we should strongly encourage (or enforce?) that names are e.g. alice\_cluster-etcd\_role

Fields:

::

    iam_role_arn="arn:aws:iam::$account:role/$role"
    base_path="..."  # base path to cluster's vault secrets, such that the kubernetes PKI  lives at {{base_path}}/pki/k8s. E.g. "dev-cluster" 

``templates/<rolename>/<templatename>``
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

Recall that we need to dynamically create two types of thing from templates:

- vault-policies; to attach to tokens, which amoungst other things point at pki-roles
- pki-roles; to actually limit EC2 Instances' PKI power. These are stored as generic secrets

Semantics:

- These templates are to be golang templates, and substitute at least: ``{{.InstanceHash}}``, ``{{.FQDN}}``, ``{{.InternalIPv4}}``, ``{{.BasePath}}``
- Secrets are to be specified in JSON \* Policies are to be specified in vault free-edition policy-language 
- ``path`` is where the rendered template should be written during the log-in process, relative to the ``base_path``

  - e.g. /pki/k8s/roles/:name would result in a pki-role at /alice\_cluster/pki/k8s/roles/kubelet
  - e.g. such a role for the kubelet would be crated by a "worker" role.

Fields:

::

    type="policy|generic"
    path="relative/path/of/template/output"
    template="<golang template of either policy document or JSON-encoded generic secret>"


Notable items
=============

Concerns
--------

-  Huge part of security critical code in our hands
-  Clean up of roles and templates once they are no longer used

Out of scope
============

- AWS auth provider's IAM mode
