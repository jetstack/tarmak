.. vault-setup-config:

******************************
Vault Setup and Configurations
******************************

`Vault <https://www.vaultproject.io>`_ is a tool developed by `Hashicorp
<https://www.hashicorp.com>`_ that securely manages secrets - handling leasing,
key revocation, key rolling, and auditing.

Vault is used in Tarmak to provide a PKI (private key infrastructure). Vault is
run on several high-availability instances that serve Tarmak clusters in the
same environment (single or multi-cluster).

Certificate Authorities (CAs)
-----------------------------
Vault is used as a Certificate Authority for Tarmak's Kubernetes Clusters.
Three CA's are needed to provide multiple roles for each cluster as follows:

* Etcd cluster that serves API server (etcd-k8s)
    * server role for etcd daemons
    * client role for kube-apiserver

* Etcd cluster that serves as overlay backend (etcd-overlay)
    * server role for etcd daemons
    * client role for overlay daemons

* Kubernetes API (k8s)
    * master role for Kubernetes master components (kube-apiserver,
      kube-controller-manager, kube-scheduler)
    * worker role for Kubernetes worker components (kube-proxy, kubelet)
    * admin role for users requesting access to Kubernetes API as admin
    * more specific bespoke roles that limit access e.g. read-only,
      namespace-specific roles

* Verifying aggregated API calls (k8s-api-proxy)
    * single role for Kubernetes components (kube-apiserver-proxy)
    * verified through a custom API server
    * CA stored in the configmap extension-apiserver-authentication in the
      kube-system namespace


Init Tokens
-----------
Tokens are used as the main authentication method in Vault and provide a
mapping to one or more policies. On first boot, each instance generates their
own unique token via a given token - the init token. These init-tokens are role
dependant meaning the same init-token is shared with instances only with the
same role. Once generated, the init token is erased by all instances in favour
of their own new unique token making the init token no longer accessible on any
instance. Unlike the init-tokens, generated tokens are short lived and so need
renewal regularly.

Purpose of Node Unique Tokens
-----------------------------
Every instance type has a unique set of policies which need to be realised in
order to be able to execute its specific operations. With unique tokens, each
instance is able to uniquely authenticate themselves against Vault with its
required policies that the tokens map to. This means renewing and revocation
can be controlled on a per instance basis. With this, each instance generates a
private key and sends a Certificate Singing Request (CSR) containing the
policies needed. Vault then verifies the CSR by ensuring the CSR matches the
requirements of the policy - if successful, returns a signed certificate.
Instances can only obtain certificates from CSRs because of the permissions
that its unique token provides. Upon receiving, the instance will store the
signed certificate locally to be shared with its relevant services and start or
restart all services which are dependant.

Expiration of Tokens and Certificates
-------------------------------------
Both signed certificates and tokens issued to each instance are short lived
meaning they need to be renewed regularly. Two Systemd timers `cert.timer` and
`token-renewal.timer` are run on each instance that will renew its
certificate and token at a default value of 24 hours. This ensures all
instances always have valid certificates. If an instance were to become offline
or the Vault server became unreachable for a sufficient amount of time,
certificates and tokens will no longer be renewable. If a certificate expires
it will become invalid and will cause the relevant operation to be halted until
its certificates are renewed. If an instance's unique token is not renewed, it
will no longer be able to ever authenticate itself against Vault and so will
need to be replaced.

Certificate Roles on Kubernetes CA
----------------------------------
**etcd-client**: Certificates with client flag - short ttl.

**etcd-server**: Certificates with client and server flag - short ttl.

**admin**: Allowed to get admin domain, certified for server certificates -
long ttl.

**kube-apiserver**: Allowed to get any domain name certified for server
certificates - short ttl.

**worker**: Allowed to get "kubelet" and "system:node" domains certified for
server and client certificates - short ttl.

**admin (kube-scheduler, kube-controller-manager, kube-proxy)**: Allowed to get
`system:<rolename>` domains (i.e. `system:kube-scheduler`) certified for client
certificates - short ttl.

**kube-apiserver-proxy**: Allowed to get "kube-apiserver-proxy" domain,
certified for client certificates - short ttl.
