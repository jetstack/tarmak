# Security Considerations

Kubernetes Cluster Deployment using Jetstack Toolkit

## Authentication

Kubernetes' API server is the interface over which all of it's components
communicate with each other.  Therefore, it's proper securing is crucial for
the whole cluster. In Toolkit deployments the API server instances are run and
accessed through an
internal ELB.

The Kubernetes API supports various different methods of authentication:

* X.509 Client Certificate Auth
* OIDC Tokens
* Service Account token auth
* Static User / Password auth
* Static Token auth

### Vault as PKI

The Toolkit uses X.509 Certificates for authenticating various components.
Every cluster has three distinct CAs:

* **Kubernetes**: Is used for authenticating system components and users against
  the Kubernetes API
* **Etcd-Kubernetes**: Is used by the `k8s` and `k8s-events` etcd clusters and
  by the Kubernetes API server to communicate with these clusters.
* **Etcd-Overlay**: Is used by the `etcd-overlay` etcd cluster and by the
  Calico components to communicate with the cluster.

Operation of a PKI for a dynamic environment like Kubernetes is not possible
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

## Authorization

After a client is authenticated Kubernetes, by default, authorizes the client
for all operations in all namespaces. This is espacially dangerous as every pod
in Kubernetes is assigned a [Service
Account](https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/)
that is used to identify it with the API server. By default, this authorizes
the pod to perform any action on any resource in any namespace. This behaviour
can be changed with the Kubernetes [authorization
plugins](https://kubernetes.io/docs/admin/authorization/). These include:

* Attribute-based access control (ABAC)
* Role-based access control (RBAC)

By default, the toolkit will configure two privileged namespaces:

* `kube-system` - pods in this namespace can perform any operation on any resource in any namespace
* `monitoring` - pods in this namespace can perform read-only actions on any resource in any namespace

All other pods are not allowed to do any API operations.

### Role Based Access Control (RBAC)

To overcome the rather static ABAC plugin, Kubernetes developed RBAC.  This
allows fine grained authorization of API Clients based on [API
objects](https://kubernetes.io/docs/admin/authorization/rbac/#role-and-clusterrole).
This Authorization mode was graduated to Beta in the 1.6 Kubernetes release.

Currently the Toolkit doesn't make use of this as it needs a proper locked down
definition of API access. The work is already scheduled to integrate the needed
RBAC policies and gradually improve them until RBAC and its policies are stable
enough to be the default Authorization mode.


## Containment & Isolation

### Network isolation

By default Kubernetes enforces no Network Isolation between different Pods in
a cluster. Furthermore it is not possible to use the Docker's flag `--icc=false` to
disable inter-container communication. Kubernetes instead provides the
NetworkPolicy API for managing communication between pods.

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
default container profile). Every Kubernetes pod uses the default policy if not
otherwise specified. When the Pod configuration flag
`securityContext.privileged` is enabled, SELinux is disabled for that specific
containers. The Toolkit currently uses privileged containers for Infrastructure
components like `calico` & `kube2iam`

Kubernetes also supports configuring a custom SELinux policy per container.
That allows to specify custom policies for different Containers. By default
Kubernetes uses unique [Multi-Category Security
(MCS)](http://james-morris.livejournal.com/5583.html) labels per Pod. This
helps to prevent further privilege escalation if a container manages to break
out of its environment.

Using Kubernetes' Pod Security Policies, certain more privileged Pod
configurations can be limited to specific Roles (cf. RBAC section). With this
you could limit the creation of Privileged Pods, Usage of `root` user in
container, mounting of HostVolumes and many more. See the official
[documentation](https://kubernetes.io/docs/concepts/policy/pod-security-policy/)
for a full list.

### Resource allocation/reservation

To protect from noisy neighbours Kubernetes manages resource allocations of CPU
and Memory using CGroups. By specifying requests and limits, these resources
can be placed into different QoS tiers. (cf.
https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container)

Right now Kubernetes provides no support for managing IO resources like Disk
and Network access.

## Exposure to the Internet

The setup of the AWS resources has been designed in a way to minimize the
attack surface the cluster has on the public internet. The only components with
public ingress interfaces are:

- Bastion instance, which provides access to the Cluster’s private networks
- ELB for Jenkins Foreman, Admin interface for Puppet master
- ELB for Jenkins, CI/CD system for Cluster Infrastructure

These resources live in the HUB of the deployment, so they exist only once per
environment. Using a terraform variable access to these can be limited to trusted
CIDR ranges. (default is 0.0.0.0/0)

All compute instances with the exception of the Bastion instance use a NAT
gateway for accessing external networks.
