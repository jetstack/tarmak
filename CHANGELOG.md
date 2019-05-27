# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## [0.6.5]: 0.6.5 - 2019-05-27

### Changed

* Upgrade golang to 1.11.10 (#814, [@jetstack-bot](https://github.com/jetstack-bot))
* Upgrade default Kubernetes version to 1.12.8 (#812, [@jetstack-bot](https://github.com/jetstack-bot))

### Fixed

* Fix consul-backinator installation (#799, [@JoshVanL](https://github.com/JoshVanL))
* Fix validity of vault init tokens by upgrading vault-helper to 0.9.15 (#803, [@JoshVanL](https://github.com/JoshVanL))
* Remove Initializers apiserver admission plugin for versions >= 1.14  (#795, [@JoshVanL](https://github.com/JoshVanL))
* Adds deployment update permissions to metrics-server-nanny container (#791, [@JoshVanL](https://github.com/JoshVanL))

### Versions

| Application | Supported versions  | Default   |
|-------------|--------------------:|----------:|
| Packer      |                     | `1.2.5`   |
| Terraform   |                     | `0.11.11` |
| Consul      |                     | `1.2.4`   |
| Vault       |                     | `0.9.6`   |
| Kubernetes  | `>= 1.10 && < 1.14` | `1.12.8`  |
| Calico      |                     | `3.1.4`   |
| Vault Helper|                     | `0.9.15`  |
| Etcd        |                     | `3.2.25`  |

## [0.6.4]: 0.6.4 - 2019-03-29

### Changed

* Upgrade default Kubernetes version to 1.12.7  (#788, [@simonswine](https://github.com/simonswine))

### Versions

| Application | Supported versions  | Default   |
|-------------|--------------------:|----------:|
| Packer      |                     | `1.2.5`   |
| Terraform   |                     | `0.11.11` |
| Consul      |                     | `1.2.4`   |
| Vault       |                     | `0.9.6`   |
| Kubernetes  | `>= 1.10 && < 1.14` | `1.12.7`  |
| Calico      |                     | `3.1.4`   |
| Vault Helper|                     | `0.9.13`  |
| Etcd        |                     | `3.2.25`  |

## [0.6.3]: 0.6.3 - 2019-03-22

### Fixed

* Disable SCTP for CVE-2019-3874 (#784, [@MattiasGees](https://github.com/MattiasGees))

### Changed

* Upgrade default Kubernetes version to 1.12.6  (#786, [@simonswine](https://github.com/simonswine))

### Versions

| Application | Supported versions  | Default   |
|-------------|--------------------:|----------:|
| Packer      |                     | `1.2.5`   |
| Terraform   |                     | `0.11.11` |
| Consul      |                     | `1.2.4`   |
| Vault       |                     | `0.9.6`   |
| Kubernetes  | `>= 1.10 && < 1.14` | `1.12.6`  |
| Calico      |                     | `3.1.4`   |
| Vault Helper|                     | `0.9.13`  |
| Etcd        |                     | `3.2.25`  |

## [0.6.2]: 0.6.2 - 2019-03-15

### Fixed

* Ensure ssh_known_host file shares the same ssh_config directory. (#778, [@JoshVanL](https://github.com/JoshVanL))
* Replace EnsureDirectoryExists by os.MkdirAll (#773, [@simonswine](https://github.com/simonswine))
* Fixes route 53 domain name reporting while using network-existing-vpc (#775, [@simonswine](https://github.com/simonswine))
* Fix bug with Terraform that causes problems on Tarmak upgrades. (#763, [@MattiasGees](https://github.com/MattiasGees))
* Set correct datacenter in consul and respect existing datacenters (#766, [@simonswine](https://github.com/simonswine))
* Support ext4/xfs for new filesystem. Detect fstype of existing ones automatically (#767, [@simonswine](https://github.com/simonswine))
* Use depends_on on the resource rather than the data object (#768, [@simonswine](https://github.com/simonswine))

### Versions

| Application | Supported versions  | Default   |
|-------------|--------------------:|----------:|
| Packer      |                     | `1.2.5`   |
| Terraform   |                     | `0.11.11` |
| Consul      |                     | `1.2.4`   |
| Vault       |                     | `0.9.6`   |
| Kubernetes  | `>= 1.10 && < 1.14` | `1.12.5`  |
| Calico      |                     | `3.1.4`   |
| Vault Helper|                     | `0.9.13`  |
| Etcd        |                     | `3.2.25`  |

## [0.6.1]: 0.6.1 - 2019-03-04

A problem during our build process has built a broken Darwin binary. Small
bug fix release.

### Fixed

* Down grade upx to 3.94 to build working Darwin binaries during releases (#758, [@simonswine](https://github.com/simonswine))

### Versions

| Application | Supported versions  | Default   |
|-------------|--------------------:|----------:|
| Packer      |                     | `1.2.5`   |
| Terraform   |                     | `0.11.11` |
| Consul      |                     | `1.2.4`   |
| Vault       |                     | `0.9.6`   |
| Kubernetes  | `>= 1.10 && < 1.14` | `1.12.5`  |
| Calico      |                     | `3.1.4`   |
| Vault Helper|                     | `0.9.13`  |
| Etcd        |                     | `3.2.25`  |

## [0.6.0]: 0.6.0 - 2019-02-27

The 0.6 release of Tarmak comes with many more features and improvements to
internals. Notable new additions include pre-built AMI images that are used when
one has not yet been built, making getting a cluster running for new users much
faster. A new worker AMI image type that will pre-install and configure Kubernetes
worker nodes so nodes become ready much faster during auto scaling. Finally, we
have also included an option to deploy Calico using Kubernetes as a backend,
rather than using Etcd directly.

A large focus of this release has been on improving the use of SSH by now
utilising the in package standard Go libraries. This has meant we now have
better control of SSH connections whilst running. We have also developed a
significant change to how SSH host keys are handled, whereby instances will now
tag themselves with their public keys securely, via an Amazon Lambda function.
These tags are then used to populate, verify and update our local host key file
during SSH connections.

We do not report any specific action required for upgrading to 0.6.0 from 0.5.3
besides our normal upgrade method.

More detailed and other changes not mentioned are as follows:

### Added

* Add Packer image that pre-installs Kubernetes dependencies drastically improving node ready time (#390 [@MattiasGees](github.com/MattiasGees))
* Expose feature flags for Kubernetes components in Tarmak configuration (#431 [@joshvanl](github.com/JoshVanL))
* Use puppet to install and manage configuration and Systemd Units on Vault instances (#494 [@joshvanl](github.com/JoshVanL))
* New command `tarmak environment destroy` to destroy all clusters in an environment (#527 [@MattiasGees](github.com/MattiasGees))
* New command `tarmak cluster logs` to gather systemd logs from target instances (#575 [@JoshVanL](github.com/JoshVanL))
* Allow custom Vault-Helper URLs to be used to download (#619 [@joshvanl](github.com/JoshVanL))
* Proposal on how to manage the SSH known hosts file and securely propagate instance public keys (#643 [@joshvanl](github.com/JoshVanL))
* Create OWNER files in sub paths of the Tarmak project (#656 [@simonswine](github.com/simonswine))
* Documentation on how to install and use Ark in Tarmak (#657 [@alljames](github.com/alljames))
* Wing tags its instance through an Amazon Lambda function securely to advertise it's public key with trust. Tarmak relies on these keys for SSH connection. (#664 [@joshvanl](github.com/JoshVanL))
* Wing dev mode now also enabled for the bastion instance (#678 [@joshvanl](github.com/JoshVanL))
* Release pre-built packer images with every release (#682 [@simonswine](github.com/simonswine))
* Give optional Kubernetes backend to calico add-on (#683 [@joshvanl](github.com/JoshVanL))
* Tarmak created Kubernetes resources have their life cycle managed by Kube-Addon-Manager (#688 [@joshvanl](github.com/JoshVanL))
* Documentation on how to add Pod Security Policies to arbitrary Namespaces (#694 [@MattiasGees](github.com/MattiasGees))
* Use Core-DNS DNS and Service Discovery project instead of Kube-DNS for clusters >= 0.10 (#715 [@joshvanl](github.com/JoshVanL))
* programmatic end to end testing with Sonobuoy (#743 [@joshvanl](github.com/JoshVanL))
* Disable Overlay ETCD servers when calico in Kubernetes backend mode (#724 [@joshvanl](github.com/JoshVanL))
* More rigorous fluent-bit acceptance tests (#747 [@simonswine](github.com/simonswine))
* Adds AddListener and RemoveListenerCertificates permissions to ELB nodes (#749 [@joshvanl](github.com/JoshVanL))
* Adds de-register permissions to ELB nodes (#750 [@joshvanl](github.com/JoshVanL))

### Changed

* Enable dry mode for vault-helper ensure to ensure to write during plan and when in a converged state (#572 [@joshvanl](github.com/JoshVanL))
* Use in package SSH over a forked exec of OpenSSH. This gives greater control and efficiency of SSH connections in Tarmak (#635 [@joshvanl](github.com/JoshVanL))
* Hard code Centos version to mitigate errors during minor releases (#649 [@simonswine](github.com/simonswine))
* Upgrade Vault to 0.9.6 and Consul to 1.2.4 (#674 [@joshvanl](github.com/JoshVanL))
* Upgrade Terraform to 0.11.11 (#675 [@joshvanl](github.com/JoshVanL))
* Upgrade wing API server internals to upstream Kubernetes (1.13) (#677 [@joshvanl](github.com/JoshVanL))
* Upgrade Golang to 1.11.4 (#680 [@simonswine](github.com/simonswine))
* Change gobindata dependency to maintained project (#699 [@simonswine](github.com/simonswine))
* Use upstream Kubernetes for binary versioning (#704 [@simonswine](github.com/simonswine))
* Separate Tarmak binaries and assets (#705 [@simonswine](github.com/simonswine))
* Makefile improvements (#709 [@simonswine](github.com/simonswine))
* Use Jetstack's patch metrics-server to scrape Kubelet summary via the Kubernetes API server proxy. Enabled Scraping Kubelets on Master nodes. (#712 [@joshvanl](github.com/JoshVanL))
* Remove gorelaser from Makefile(#714 [@simonswine](github.com/simonswine))
* Known hosts keys managed by Tarmak and will update if the instance public key tags have updated (#721 [@joshvanl](github.com/JoshVanL))
* If no private images have been built for non EBS encrypted clusters, fallback
  to using Jetstack's pre-built images (#724 [@joshvanl](github.com/JoshVanL))
* Upgrade Fluentbit to 1.0.4 (#725 [@simonswine](github.com/simonswine))
* Upgrade Centos to 7.6.1810 (#726 [@simonswine](github.com/simonswine))
* Improve Elastic Search settings (#732 [@simonswine](github.com/simonswine))
* SSH tunnels have a timeout after 10 minutes of inactivity (#730 [@joshvanl](github.com/JoshVanL))
* Heapster, InfluxDB and Grafana have toggles in the Tarmak configuration. They
  are enabled for current clusters but disable by default for all newly created
  clusters via init (#740 [@joshvanl](github.com/JoshVanL))
* Upgrade default Kubernetes version to 1.12.5 (#753 [@simonswine](github.com/simonswine))

### Fixed

* Correctly parse Kubectl arguments (#477 [@joshvanl](github.com/JoshVanL))
* Ensure the latest kernel version is being used (#658 [@simonswine](github.com/simonswine))
* Use correct Kubconfig certificate when using Kubernees API server with public ELB (#660 [@MattiasGees](github.com/MattiasGees))
* Correctly mount NVME volumes for vault instances (#697 [@joshvanl](github.com/JoshVanL))
* Spelling correction (#701 [@simonswine](github.com/simonswine))
* Create fresh cluster directory in configuration if none existing (#702 [@joshvanl](github.com/JoshVanL))
* Don't create VPC S3 endpoint when using an existing VPC (#707 [@joshvanl](github.com/JoshVanL))
* Tarmak no longer falsely reports an instance type as unavailable in an available zone (#732 [@MattiasGees](github.com/MattiasGees))
* Correct Tiller documentation (#693 [@MattiasGees](github.com/MattiasGees))
* Tarmak no longer falsely over reports a bad connection to the Bastion instance (#710 [@joshvanl](github.com/JoshVanL))
* Fix Hashsum of wing (#711 [@simonswine](github.com/simonswine))
* Fix sources in Makefile (#713 [@simonswine](github.com/simonswine))
* Fix heapster vertical pod autoscaler race condition (#720 [@simonswine](github.com/simonswine))
* Update Heptio's Ark to the newly named Velero (#722 [@joshvanl](github.com/JoshVanL))
* Fix the attachment of additional policies to non Kubernetes instances (#727 [@simonswine](github.com/simonswine))
* Input query during Terraform running fixed from a breaking change (#729 [@joshvanl](github.com/JoshVanL))
* Tunnels to the Kubernetes API server are re-used is available (#736 [@joshvanl](github.com/JoshVanL))
* Fix Kube-state-metrics RBAC (#754 [@MattiasGees](github.com/MattiasGees))
* Increase inotify watch limits of instances with Kubelet (#756 [@JoshVanL](github.com/joshvanl))

### Versions

| Application | Supported versions  | Default   |
|-------------|--------------------:|----------:|
| Packer      |                     | `1.2.5`   |
| Terraform   |                     | `0.11.11` |
| Consul      |                     | `1.2.4`   |
| Vault       |                     | `0.9.6`   |
| Kubernetes  | `>= 1.10 && < 1.14` | `1.12.5`  |
| Calico      |                     | `3.1.4`   |
| Vault Helper|                     | `0.9.13`  |
| Etcd        |                     | `3.2.25`  |

## [0.5.3]: 0.5.3 - 2018-12-21

More bugfixes...

### Fixed

* Fix bug with kubectl/kubeconfig and public apiserver (#660, [@MattiasGees](https://github.com/MattiasGees))
* Make sure centos-puppet-agent-latest-kernel is booting into the right kernel (#658, [@simonswine](https://github.com/simonswine))

### Versions

| Application | Supported versions | Default  |
|-------------|-------------------:|---------:|
| Packer      |                    | `1.0.2`  |
| Terraform   |                    | `0.11.8` |
| Consul      |                    | `1.2.3`  |
| Vault       |                    | `0.9.5`  |
| Kubernetes  | `>= 1.9 && < 1.13` | `1.11.5` |
| Calico      |                    | `3.1.4`  |
| Vault Helper|                    | `0.9.13` |
| Etcd        |                    | `3.2.25` |

## [0.5.2]: 0.5.2 - 2018-12-07

Bugfix release to fix regression that come up in the 0.5 release branch.
Notably now hard coding the Centos release to 7.5. To avoid instability from a
new Centos minor version.

### Changed

* Hardcode centos image release to 7.5.1804 (#649, [@simonswine](https://github.com/simonswine))

### Fixed

* Override local kubeconfig if errors (#652, [@JoshVanL](https://github.com/JoshVanL))
* Correctly mount nvme drives on etcd instances (#538, [@JoshVanL](https://github.com/JoshVanL))
* Fix centos 7.6 aws cli, download it through pip if it's not working (#646, [@simonswine](https://github.com/simonswine))

### Versions

| Application | Supported versions | Default  |
|-------------|-------------------:|---------:|
| Packer      |                    | `1.0.2`  |
| Terraform   |                    | `0.11.8` |
| Consul      |                    | `1.2.3`  |
| Vault       |                    | `0.9.5`  |
| Kubernetes  | `>= 1.9 && < 1.13` | `1.11.5` |
| Calico      |                    | `3.1.4`  |
| Vault Helper|                    | `0.9.13` |
| Etcd        |                    | `3.2.25` |

## [0.5.1]: 0.5.1 - 2018-12-04

Release to update default Kubernetes version to 1.11.5: CVE-2018-1002105: proxy
request handling in kube-apiserver can leave vulnerable TCP connections
([details](https://github.com/kubernetes/kubernetes/issues/71411)).

### Changed

* Update default kubernetes version for new clusters to 1.11.5 (#645, [@JoshVanL](https://github.com/JoshVanL))

### Versions

| Application | Supported versions | Default  |
|-------------|-------------------:|---------:|
| Packer      |                    | `1.0.2`  |
| Terraform   |                    | `0.11.8` |
| Consul      |                    | `1.2.3`  |
| Vault       |                    | `0.9.5`  |
| Kubernetes  | `>= 1.9 && < 1.13` | `1.11.5` |
| Calico      |                    | `3.1.4`  |
| Vault Helper|                    | `0.9.13` |
| Etcd        |                    | `3.2.25` |


## [0.5.0]: 0.5.0 - 2018-11-26

The 0.5 release of Tarmak adds support for Kubernetes up to minor version 1.12.
A focus of the release was to ensure all data stores are encrypted at rest.
Another focus was on the stability of tarmak. Various components had version
and/or configuration upgrades to ensure resiliency in the operation.

This detailed changes have happend since the last minor version of Tarmak:

### Added

* Update default kubernetes version for new clusters to 1.11.4 (#638, [@simonswine](https://github.com/simonswine))
* Istio example in documentation (#551, [@charlieegan3](https://github.com/charlieegan3))
* Option to enable EBS encryption (#496, [@alljames](https://github.com/alljames))
* Toogle EBS encryption and protect EBS data from being deleted (#531, [@simonswine](https://github.com/simonswine))
* Kube bench proposed security fixes (#639, [@simonswine](https://github.com/simonswine))
* Point Tarmak CLI to new multicluster environment's 'hub' cluster by default (#566, [@alljames](https://github.com/alljames))
* Jetstack Navigator example in documentation (#539, [@charlieegan3](https://github.com/charlieegan3))
* SPIFFE/SPIRE proposal/feasibility document. (#445, [@JoshVanL](https://github.com/JoshVanL))
* Documentation regarding using AWS instance storage (#545, [@MattiasGees](https://github.com/MattiasGees))
* Prometheus collection of systemd unit status (#612, [@simonswine](https://github.com/simonswine))
* Bastion and Vault instance pools now support additional policies declared in the config (#579, [@JoshVanL](https://github.com/JoshVanL))
* Etcd backup strategy (daily push of KMS encrypted snapshots of every instance) (#558, [@simonswine](https://github.com/simonswine))
* Auto-generated CLI documentation (#589, [@JoshVanL](https://github.com/JoshVanL))
* Flag --auto-approve and --auto-approve-deleting-data for `cluster apply` command (#560, [@JoshVanL](https://github.com/JoshVanL))
* KMS Server Side Encryption to Consul S3 backups (#614, [@JoshVanL](https://github.com/JoshVanL))
* KMS encrypt terraform remote S3 state data. (#505, [@JoshVanL](https://github.com/JoshVanL))
* `plan --plan-file-store` and `apply --plan-file-location` (#563, [@JoshVanL](https://github.com/JoshVanL))
* `cluster apply --auto-approve` and `cluster apply --auto-approve-deleting-data` (#560, [@JoshVanL](https://github.com/JoshVanL))
* Format terraform code for CI (#580, [@JoshVanL](https://github.com/JoshVanL))
* Tests for auto-generated terraform code (#535, [@JoshVanL](https://github.com/JoshVanL))
* Restart Consul on failure (#502, [@dippynark](https://github.com/dippynark))
* Restart etcd and wing-server on the bastion automatically on failure (#510, [@dippynark](https://github.com/dippynark))
* Metrics-server add-on from Kubernetes version 1.7 onwards (#487, [@dippynark](https://github.com/dippynark))
* Vault_server puppet module to initiate vault servers (#476, [@JoshVanL](https://github.com/JoshVanL))
* Support to enable API Server ELB access logs (#492, [@JoshVanL](https://github.com/JoshVanL))
* Set root volume attribute variables, previously only default was used. (#447, [@charlieegan3](https://github.com/charlieegan3))
* Cluster force-unlock subcommand for to release terraform state lock. (#522, [@JoshVanL](https://github.com/JoshVanL))
* Expose auto-cluster's `--scale-down-utilization-threshold` in .tarmak.yaml (#456, [@JoshVanL](https://github.com/JoshVanL))
* Validate configuration, so that hubs in multi cluster environments contain all zones of their clusters  (#471, [@JoshVanL](https://github.com/JoshVanL))
* `cluster kubeconfig` (#632, [@JoshVanL](https://github.com/JoshVanL))
* Configuration file for Kubelet and Kube-Proxy for Kubrnetes clusters >= 1.11  (#442, [@JoshVanL](https://github.com/JoshVanL))

### Changed

* Unset API Server depreciated flags for Kubernetes version >= 1.11 (#440, [@JoshVanL](https://github.com/JoshVanL))
* Only wait for wing conversion when infrastructure-only mode specified (#493, [@JoshVanL](https://github.com/JoshVanL))
* Encrypt S3 puppet tar ball and consul backup buckets. (#504, [@JoshVanL](https://github.com/JoshVanL))
* Generate API documentation in site (#533, [@JoshVanL](https://github.com/JoshVanL))
* Use SSL protocol for API server health checks (#524, [@lostick](https://github.com/lostick))
* Ensure connected vault tunnel is healthy (#512, [@JoshVanL](https://github.com/JoshVanL))
* Move tarmak terraform provider socket to /tmp (#587, [@JoshVanL](https://github.com/JoshVanL))
* Better advice when remote state has been destroyed (#576, [@JoshVanL](https://github.com/JoshVanL))
* Make Jenkins a valid instancepool for hub (#478, [@MattiasGees](https://github.com/MattiasGees))
* Bump default Kubernetes version for new clusters to 1.11.4 (#638, [@simonswine](https://github.com/simonswine))
* Bump fluentbit to 0.14.6 (#585, [@MattiasGees](https://github.com/MattiasGees))
* Bump node-exporter to 0.16.0 (#537, [@lostick](https://github.com/lostick))
* Bump etcd to 3.2.25 (#623, [@JoshVanL](https://github.com/JoshVanL))
* Bump Terraform to v0.11.8 (#516, [@simonswine](https://github.com/simonswine))
* Bump Calico to 3.1.4 (#622, [@JoshVanL](https://github.com/JoshVanL))
* Bump Heapster to 1.5.4 (#491, [@dippynark](https://github.com/dippynark))
* Bump Prometheus to 2.3.2 and related components to latest version (#624, [@JoshVanL](https://github.com/JoshVanL))

### Fixed

* Terraform debug shell error when binary version is incompatible (#495, [@dippynark](https://github.com/dippynark))
* Bug with conversion of yaml loggingsink to puppetcode (#581, [@MattiasGees](https://github.com/MattiasGees))
* Bug with Grafana in cluster service (#460, [@MattiasGees](https://github.com/MattiasGees))
* Better `cluster images build` behavior (#604, [@JoshVanL](https://github.com/JoshVanL))
* Node exporter port on etcd nodes (#553, [@simonswine](https://github.com/simonswine))
* Consul and update behaviour (#570, [@simonswine](https://github.com/simonswine))
* Packer image updates to fix failing services (#562, [@simonswine](https://github.com/simonswine))
* Clean up ssh run time assets (#597, [@JoshVanL](https://github.com/JoshVanL))
* Correctly mount docker storage on NVMe driver AWS instances. (#461, [@JoshVanL](https://github.com/JoshVanL))
* Ensure code generation is verified correctly (#462, [@simonswine](https://github.com/simonswine))
* Propagate interrupt signals to sub-processes and tasks (#356, [@JoshVanL](https://github.com/JoshVanL))

### Versions

| Application | Supported versions | Default  |
|-------------|-------------------:|---------:|
| Packer      |                    | `1.0.2`  |
| Terraform   |                    | `0.11.8` |
| Consul      |                    | `1.2.3`  |
| Vault       |                    | `0.9.5`  |
| Kubernetes  | `>= 1.9 && < 1.13` | `1.11.4` |
| Calico      |                    | `3.1.4`  |
| Vault Helper|                    | `0.9.13` |
| Etcd        |                    | `3.2.25` |


## [0.4.1]: 0.4.1 - 2018-08-24

### Fixed

* Correctly mount docker storage on NVMe driver AWS instances. (#461, [@JoshVanL](https://github.com/JoshVanL))
* Fix grafana in cluster service (#460, [@MattiasGees](https://github.com/MattiasGees))
* Ensure code generation is verified correctly  (#462, [@simonswine](https://github.com/simonswine))
* Set root volume attribute variables, previously only default was used. (#447, [@charlieegan3](https://github.com/charlieegan3))

### Versions

| Application | Supported versions | Default  |
|-------------|-------------------:|---------:|
| Packer      |                    | `1.0.2`  |
| Terraform   |                    | `0.11.7` |
| Consul      |                    | `1.0.6`  |
| Vault       |                    | `0.9.5`  |
| Kubernetes  | `>= 1.7 && < 1.11` | `1.9.10` |
| Calico      |                    | `3.1.1`  |
| Vault Helper|                    | `0.9.13` |
| Etcd        |                    | `3.2.17` |

## [0.4.0]: 0.4.0 - 2018-08-07

### Added
- Add Tarmak Terraform provider for ordering infrastructure creation (#12, @simonswine)
- Add support for automatically adding taints and labels to instance pools (#369, @charlieegan3)
- Support log forwarding (#197, @dippynark)
- Add Jenkins module to Terraform stack (#240, @MattiasGees)
- Support autoscaling arbitrary worker instance pools (#325, @dippynark)

### Changed
- Merged Terraform stacks (state, bastion, vault, network, kubernetes) into a single stack. This allows a plan to be run against all infrastructure at the same time and also benefit from Terraform's parallelisation  capabilities (#148, @dippynark)
- Vendor Terraform instead of shelling out to binary inside the Tarmak Docker container. This gives us more control over how terraform is run and the version used. Care must be take when running terraform commands within the Tarmak debug shell as using a version of Tarmak higher than the version vendored by Tarmak will prevent Tarmak from running further Terraform commands
- Change cgroup driver from systemd to cgroupfs as cgroupfs has better support in the kubelet for enforcing node allocatable (#300, @dippynark)

# Fixed
- Add security group to allow cluster autoscaler scaping (#338, @dippynark)
- Remove unneeded infrastructure (#329 #336 #321 @dippynark @MattiasGees)
- Pass through etcd instance pool min count to puppet (#322, @dippynark)
- Fix etcd mount race condition (#313, @dippynark)
- Add RBAC support to Dashboard (#343, @dippynark)
- Use correct versions for cluster autoscaler (#346, @dippynark)
- Return informative error when failing to parse tarmak configuration (#326, @dippynark)
- Use ClusterFirstWithHostNet for fluent-bit ds (#319, @charlieegan3)
- Prepare Terraform when running kubectl (#185, @dippynark)

### Versions

| Application | Supported versions | Default  |
|-------------|-------------------:|---------:|
| Packer      |                    | `1.0.2`  |
| Terraform   |                    | `0.11.7` |
| Consul      |                    | `1.0.6`  |
| Vault       |                    | `0.9.5`  |
| Kubernetes  | `>= 1.7 && < 1.11` | `1.10.6` |
| Calico      |                    | `3.1.1`  |
| Vault Helper|                    | `0.9.13` |
| Etcd        |                    | `3.2.17` |

## [0.3.0]: 0.3.0 - 2018-02-20

### Added

* Add `--keep-containers` flag to preserve container environment launched by tarmak (#108, [@simonswine](https://github.com/simonswine))
* Adds vault setup and config to docs (#51, [@JoshVanL](https://github.com/JoshVanL))
* Upgrade prometheus monitoring to 2.0 (support for RBAC, customizable scraping + alerting configs) (#68, [@simonswine](https://github.com/simonswine))

### Changed

* Use upstream goreleaser, GPG signing merged upstream (#116, [@simonswine](https://github.com/simonswine))
* Update calico to 2.6.6 (#91, [@simonswine](https://github.com/simonswine))
* Enhance kube state metrics (#95, [@simonswine](https://github.com/simonswine))
* Update terraform to 0.11 (#87, [@simonswine](https://github.com/simonswine))
* Update vault to 0.9.1 and consul to 1.0.2 (#88, [@simonswine](https://github.com/simonswine))
* Tarmak is now compiled against k8s.io release-1.8 branches. (#14, [@wallrj](https://github.com/wallrj))

### Fixed

* Fix multi cluster environments by supporting multiple clusters in a single VPC (#100, [@dippynark](https://github.com/dippynark))
* Retry SSH connection to bastion during tools stack (#81, [@JoshVanL](https://github.com/JoshVanL))
* Ensure systemd unit order for kubelet and kube-proxy (#69, [@simonswine](https://github.com/simonswine))

### Versions

| Application | Supported versions | Default  |
|-------------|-------------------:|---------:|
| Packer      |                    | `1.0.2`  |
| Terraform   |                    | `0.11.2` |
| Consul      |                    | `1.0.2`  |
| Vault       |                    | `0.9.1`  |
| Kubernetes  | `>= 1.6 && < 1.10` | `1.8.8`  |
| Calico      |                    | `2.6.6`  |


## [0.2.1]: 0.2.1 - 2017-12-05

### Fixed

* Fix concurrency issues with Wing, ensure only a single puppet run happens at a time  (#61, [@simonswine](https://github.com/simonswine))

### Versions

| Application | Supported versions | Default  |
|-------------|-------------------:|---------:|
| Packer      |                    | `1.0.2`  |
| Terraform   |                    | `0.10.8` |
| Consul      |                    | `0.8.5`  |
| Vault       |                    | `0.7.3`  |
| Kubernetes  | `>= 1.6 && < 1.9`  | `1.7.10` |

## [0.2.0]: 0.2.0 - 2017-12-01

### Added

* Adds signal handling to Wing to handle TERM and HUP, SIGHUP: Cause a node to be reconverged, SIGTERM: Forward sigterm to puppet subprocess (if exists) (#32, [@JoshVanL](https://github.com/JoshVanL))
* Sign released binaries using GPG (#58, [@simonswine](https://github.com/simonswine))
* Update default kubernetes version to 1.7.10 (#54, [@simonswine](https://github.com/simonswine))
* Add support for API server aggregation, enabled by default for kubernetes 1.7+ (#53, [@simonswine](https://github.com/simonswine))
* Validate minCount and maxCount of Instance Pool (#52, [@JoshVanL](https://github.com/JoshVanL))
* Enable authorization and authentication for kubelet (#46, [@simonswine](https://github.com/simonswine))
* Enable Node authorizer and related admission controller for 1.8 compatibility  (#41, [@simonswine](https://github.com/simonswine))
* Add experimental support for deploying clusters into existing AWS VPCs (#31, [@kragniz](https://github.com/kragniz))


### Changed

* Allow master to communicate with workers on any port (#50, [@simonswine](https://github.com/simonswine))
* Raise the master LoadBalancer time out to 3600 seconds (#49, [@simonswine](https://github.com/simonswine))
* Verify at least one image exists before running terraform apply (#36, [@JoshVanL](https://github.com/JoshVanL))
* Disable apiserver binding insecure-port on the master (#48, [@simonswine](https://github.com/simonswine))
* Update vendored k8s.io packages to target release-1.8/release-5.0 branches (#15, [@simonswine](https://github.com/simonswine))
* Disable source/destination check on cloud-provider AWS using a controller run on kubernetes masters. No need to authorize worker instances for ec2:ModifyInstanceAttribute anymore. (#28, [@mattbates](https://github.com/mattbates))
* Update vendored vault-helper and vault-unsealer to latest releases (#20, [@JoshVanL](https://github.com/JoshVanL))
* Update kubernetes master taints and cgroup fixes (#38, [@simonswine](https://github.com/simonswine))
* Upgrade terraform to 0.10.8 (#40, [@simonswine](https://github.com/simonswine))

### Versions

| Application | Supported versions | Default  |
|-------------|-------------------:|---------:|
| Packer      |                    | `1.0.2`  |
| Terraform   |                    | `0.10.8` |
| Consul      |                    | `0.8.5`  |
| Vault       |                    | `0.7.3`  |
| Kubernetes  | `>= 1.6 && < 1.9`  | `1.7.10` |

## 0.1.2 - 2017-10-19

### Initial release (*alpha*)
- First public release
- Support for AWS provider only
- Prepare and drive infrastructure updates using Terraform
- Prepare configuration updates using Puppet and drive them using Wing on the
  instances
- Provides wrappers for basic administrative task: kubectl, ssh
- Experimental vendoring of Kubicorn's Cluster API (https://github.com/kris-nova/kubicorn) for cluster configuration

> Disclaimer - please note that current releases of Tarmak are alpha (unless
> explicitly marked). Although we do not anticipate breaking changes, at this
> stage this cannot be absolutely guaranteed.

### Versions used

| Application | Supported versions | Default  |
|-------------|-------------------:|---------:|
| Packer      |                    | `1.0.2`  |
| Terraform   |                    | `0.10.6` |
| Consul      |                    | `0.8.5`  |
| Vault       |                    | `0.7.3`  |
| Kubernetes  | `>= 1.5 && < 1.8`  | `1.7.7`  |

[0.6.4]: https://github.com/jetstack/tarmak/compare/0.6.3...0.6.4
[0.6.3]: https://github.com/jetstack/tarmak/compare/0.6.2...0.6.3
[0.6.2]: https://github.com/jetstack/tarmak/compare/0.6.1...0.6.2
[0.6.1]: https://github.com/jetstack/tarmak/compare/0.6.0...0.6.1
[0.6.0]: https://github.com/jetstack/tarmak/compare/0.5.0...0.6.0
[0.5.3]: https://github.com/jetstack/tarmak/compare/0.5.2...0.5.3
[0.5.2]: https://github.com/jetstack/tarmak/compare/0.5.1...0.5.2
[0.5.1]: https://github.com/jetstack/tarmak/compare/0.5.0...0.5.1
[0.5.0]: https://github.com/jetstack/tarmak/compare/0.4.1...0.5.0
[0.4.1]: https://github.com/jetstack/tarmak/compare/0.4.0...0.4.1
[0.4.0]: https://github.com/jetstack/tarmak/compare/0.3.0...0.4.0
[0.3.0]: https://github.com/jetstack/tarmak/compare/0.2.0...0.3.0
[0.2.1]: https://github.com/jetstack/tarmak/compare/0.2.0...0.2.1
[0.2.0]: https://github.com/jetstack/tarmak/compare/0.1.2...0.2.0
[Unreleased]: https://github.com/jetstack/tarmak/compare/0.5.0...HEAD
