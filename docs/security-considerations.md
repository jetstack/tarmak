# Security Considerations

Kubernetes Cluster Deployment using Jetstack Toolkit

## Authentication and Authorization

Kubernetes’ API server is the interface over all of its components communicate
with each other. Therefore its security is crucial for the whole cluster. In
Toolkit deployments the API server instances are run and accessed through an
internal ELB.

The Kubernets API supports different methods of authentication:

* X.509 Client Certificate Auth
* Service Account token auth
* Static User / Password auth
* Static Token auth

After a client is authenticated Kubernetes, by default, authorizes this client
for all operations in all namespaces. Obvisouly this comes with large risks:
Every Pod in Kubernetes gets a so call Service Account, which allows it to
connect back to the kubernetes API, so without a more bespoke Authorization
policy, applications can trigger any operation in the Master API. To limit the
level of access for service account tokens, kubernetes allows different
authorization plugins like Attribute-Based Access Control (ABAC) and Role Based
Access Control (RBAC). The Toolkit allows any API access for Service accounts
in the namespace `kube-system` and read-only access for services in
`monitoring`. The other namespaces are limit to read-only operations to their
namespace only.

### Role Based Access Control (RBAC)

The overcome the quite static ABAC approach Kubernetes developed RBAC. This
allows a fine grained authorization of API Clients based on API objects. This
Authroization mode was graduated to Beta in the 1.6 Kubernetes release.

Currently the Toolkit doesn't make use of this as it needs a proper locked down
definition of API access. The work is already scheduled to integrate the needed
RBAC policies and gradually improved them until RBAC is stable enough to be the
default Authorization mode.

### Vault as PKI

The Toolkit uses X.509 Certificates for authenticating various components.
Every cluster has three different CAs:

* **Kubernetes**: Is used for authentication system components and user against
  the Kubernetes API
* **Etcd-Kubernetes**: Is used within the `k8s` and `k8s-events` cluster and
  for Kubernetes API servers to communicate with it components to communicate
  with it
* **Etcd-Overlay**: Is used within the `etcd-overlay` cluster and for Calico
  components to communicate with it

Operation of an PKI for a dynamic environment like Kubernetes is not possible
without full automation. Hashicorp's [Vault](https://www.vaultproject.io/)
provides a well featured PKI backend:

* Handles leasing, key revocation, key rolling, Audits CA events
* init-token generates a short-lived EC2 instance unique token, that is bound
  to the instance's role (etcd, master or worker)
* Short-lived X.509 certificates and instance tokens make sure that credentials
  expire after an EC2 instance is no longer available

While Vault is currently only used for handling infrastructure related
credentials, its also capable of handling credentials of the Cluster's
workload.

## Containment & Isolation

### Network isolation

By default Kubernetes enforces no Network Isolation between different Pods in
a cluster. Furthermore it is not possible to use the Docker's flag `--icc=false` to
disable inter-container communication.

#### Network Policy Enforcement using Calico

For getting Network isolation within a multi-tenant Kubernetes cluster you have
to utilize the Kubernetes NetworkPolicy resource. It allows to specify (using
labels) selections of pods that are allowed to communicate with each other and
other network endpoints described using white list rules. It only allow to
defined Ingress rules.

While Kubernetes implements only the storage of these objects it doesn't
actually enforce the specified rules itself. The Toolkit uses Calico to
translate the NetworkPolicy resources into iptables rules, that get applied to
every host node on the cluster. A walk-through guide of how to use
NetworkPolicy can be found in the Kubernetes docs: [NetworkPolicy
walk-through](https://kubernetes.io/docs/tasks/administer-cluster/declare-network-policy/).

#### Calico Egress Policies

A Pod's network access should be as limited as possible. While NetworkPolicy
only restricts ingress traffic for Pods, the outgoing traffic should be limited
as well.

Using Calico's API directly Egress rules can be specified as well. These rules
can use Pod labels and Namespaces. During the evaluation phase we create an
example [Egress
policy](https://github.com/Skyscanner/kubernetes.platform.charts/tree/master/egress-policy)

### Process isolation

While Docker isolates the containers by preventing a specific list of System
calls, the Toolkit also enables SELinux by default. CentOS ships a default
policy that works with most container workloads (Similar to the AppArmor
default container profile). Every Kubernetes Pod is using that default policies
if not specified otherwise. When the Pod configuration flag
`securityContext.privileged` is enabled, then SELinux is disabled for that
specific containers. The toolkit currently uses this privileged containers for
Infrastructure components like `calico` & `kube2iam`

Kubernetes also supports configuring custom SELinux per container. That allows
to specify custom policies for different Pods. By default Kubernetes uses unique
Multi-Category Security (MCS) labels per Pod to separate the containers of
different. That helps to prevent further privilege escalation, once a container
managed to break out of it's chrooted environment.

Using Kubernetes' Pod Security Policies, certain more privileged Pod
configurations can be limited to specific Roles (cf. RBAC section). With this
you could limit the creation of Privileged Pods, Usage of `root` user in
container, mounting of HostVolumes and many more. See the official
[documentation](https://kubernetes.io/docs/concepts/policy/pod-security-policy/)
for a full list

### Resource allocation/reservation

To protect from noisy neighbours Kubernetes manages resource allocations of CPU
and Memory using CGroups. By specifying requests and limits, these resources
can be allocated using different QoS.  (cf.
https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container)

Right now Kubernetes provides no support for IO resources like Disk and Network
access.

## Exposure to the Internet

The setup of the AWS resources has been designed in a way to minimize the
attack surface the cluster has towards the public internet. The only components
with public interfaces are:

- Bastion instance, which provides access to the Cluster’s private networks
- ELB for Jenkins Foreman, Admin interface for Puppet master
- ELB for Jenkins, CI/CD system for Cluster Infrastructure

These resources live in the HUB of the deployment, so they exist only once per
cluster. Using a terraform variable access to these can be limited to trusted
CIDR ranges. (default is 0.0.0.0/0)

All compute instances with the exception of the Bastion instance use a NAT
gateway for accessing external networks.
