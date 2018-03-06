.. _vault-helper-guide:

Vault Helper In Tarmak
======================

`Vault <https://www.vaultproject.io>`_ is used in Tarmak to `provide a PKI
(private key infrastructure) <vault-setup-config.html>`_. `vault-helper
<https://github.com/jetstack/vault-helper>`_ is a tool designed to facilitate
and automate the PKI tasks required. Each Kubernetes instance type in a Tarmak
cluster is in need of signed certificates from Vault to operate. These
certificates need to be stored locally and regularly renewed.

It is essential for the Vault stack to be executed and completed before the
Kubernetes stacks as they rely on communication from Vault. With the Vault
stack completed and the connection to the Vault server established, the
vault-helper package is used to mount all backends (Etcd, Etcd overlay, K8s,
K8s proxy and secrets generic) to Vault if they have not already. Mounts with
incorrect default or max lease TTLs will be re-tuned accordingly.

These backends serve as the CAs for the Kubernetes components. Roles and then
polices to these roles are written to each CA as described `here
<vault-setup-config.html#certificate-roles-on-kubernetes-ca>`_.  The init-token
polices and roles are then written to Vault also. This whole process is
idempotent.

Vault is now set up correctly for each CA, tokens, roles and policies.

The vault-helper binary is stored on all cluster instances (etcd, worker and
master). Two Systemd timers are run on every cluster instance in order to
renew both tokens and certificates every day. These will trigger oneshot
services (token-renewal.service, cert.service) to be fired, executing the
locally stored vault-helper binary to renew certificates and tokens. When
executing either renew-token or cert, vault-helper will recognise if an
init-token is present in local file, generating a new unique token to be
stored, deleting the init-token. The cert subcommand will ensure a private key
is generated, if one does not exist, before sending a CSR to the Vault server.
The returned signed certificate it then stored locally, replacing any previous
certificates.
