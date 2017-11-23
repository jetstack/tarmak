.. vault-setup-config:

******************************
Vault Setup and Configurations
******************************

`Vault <https://www.vaultproject.io>`_ is a tool developed by `Hashicorp <https://www.hashicorp.com>`_ that securely manages secrets - handling leasing, key revocation, key rolling, and auditing.

Certificate Authorities (CAs)
-----------------------------
Vault is used as a Certificate Authority for Tarmak's Kubernetes Clusters.
Three CA's are needed to provide multiple roles for each cluster as follows:

* Etcd cluster that serves API server (etcd-k8s)
    * server role for etcd daemons
    * client role for kube-apiserver

* CA for Etcd cluster that serves as overlay backend (flannel or calico) (etcd-overlay)
    * server role for etcd daemons
    * client role for flannel/calico daemons

* Kubernetes API (k8s)
    * master role for Kubernetes master components (kube-apiserver, kube-controller-manager, kube-scheduler)
    * worker role for Kubernetes worker components (kube-proxy, kubelet)
    * admin role for users requesting access to Kubernetes API as admin
    * more specific bespoke roles that limit access e.g. read-only, namespace-specific roles

Init Tokens
-----------
Tokens are used as the main authentication method in Vault and provide a mapping to one or more policies.
On first boot, each instance generates their own unique token via a given shared token - the init token.
Once generated, this init token is erased by all instances in favour of their own new unique token meaning the init token is no longer accessible on any instance.

Purpose of Node Unique Tokens
-----------------------------
Every instance type has a unique set of policies which need to be realised in order to be able to execute its specific operations.
With unique tokens, each instance is able to uniquely authenticate themselves against Vault with its required policies that the tokens map to.
This means renewing and revocation can be controlled on a per instance basis.
With this, each instance generates a private key and sends a Certificate Singing Request (CSR) containing the policies needed.
Vault then verifies the CSR by ensuring the CSR matches the requirements of the policy - if successful, returns a signed certificate.
Upon receiving, the instance will store the signed certificate locally to be shared with its relevant services and start or restart all services which are dependant.

Expiration of Tokens and Certificates
-------------------------------------
Both signed certificates and tokens issued to each instance are short lived meaning they need to be renewed regularly.
A Cron Job is run on each instance that will periodically renew its certificate and token in time before expiry, ensuring all instances always have valid certificates.
If an instance became offline or the Vault server became unreachable for a sufficient amount of time, certificates and tokens will no longer be able to be renewed.
If a certificate expires it will become invalid and will cause the relevant operation to be halted until it's certificates are renewed.
If an instance's unique token is not renewed, it will no longer be able to ever authenticate itself against Vault and so will need to be replaced.

Certificate Roles on Kubernetes CA
----------------------------------
etcd-client role:

::

   use_csr_common_name: false
   use_csr_sans:        false
   allow_any_name:      true
   allow_ip_sans:       true
   server_flag:         false
   client_flag:         true
   max_ttl:             30 days
   ttl:                 30 days

etcd-server role:

::

  use_csr_common_name: false
  use_csr_sans:        false
  allow_any_name:      true
  allow_ip_sans:       true
  server_flag:         true
  client_flag:         true
  max_ttl:             30 days
  ttl:                 30 days

admin role:

::

  use_csr_common_name: false
  enforce_hostnames:   false
  organization:        "system:masters"
  allowed_domains:     admin
  allow_bare_domains:  true
  allow_localhost:     false
  allow_subdomains:    false
  allow_ip_sans:       false
  server_flag:         false
  client_flag:         true
  max_ttl:             365 days
  ttl:                 365 days

master role (Kube API Server):

::

  use_csr_common_name: false
  use_csr_sans:        false
  enforce_hostnames:   false
  allow_localhost:     true
  allow_any_name:      true
  allow_bare_domains:  true
  allow_ip_sans:       true
  server_flag:         true
  client_flag:         false
  max_ttl:             30 days
  ttl:                 30 days

master role (kube-apiserver Proxy):

::

  use_csr_common_name: false
  use_csr_sans:        false
  enforce_hostnames:   false
  server_flag:         false
  client_flag:         true
  allowed_domains:     ["kube-apiserver-proxy"]
  max_ttl:             30 days
  ttl:                 30 days

worker role (kubelet):

::

  use_csr_common_name: false
  use_csr_sans:        false
  enforce_hostnames:   false
  organization:        "system:nodes"
  allowed_domains:     ["kubelet", "system:node", "system:node:*"]
  allow_bare_domains:  true
  allow_glob_domains:  true
  allow_localhost:     false
  allow_subdomains:    false
  server_flag:         true
  client_flag:         true
  max_ttl:             30 days
  ttl:                 30 days

admin role (kube-scheduler, kube-controller-manager, kube-proxy):

::

  use_csr_common_name: false
  enforce_hostnames:   false
  allowed_domains:     ["system:<rolename>"]
  allow_bare_domains:  true
  allow_localhost:     false
  allow_subdomains:    false
  allow_ip_sans:       true
  server_flag:         false
  client_flag:         true
  max_ttl:             30 days
  ttl:                 30 days
