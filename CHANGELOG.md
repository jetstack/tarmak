# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## [Unreleased]:

### Added

### Changed

### Fixed

### Versions

| Application | Supported versions | Default  |
|-------------|-------------------:|---------:|
| Packer      |                    | `1.0.2`  |
| Terraform   |                    | `0.11.8` |
| Consul      |                    | `1.2.3`  |
| Vault       |                    | `0.9.5`  |
| Kubernetes  | `>= 1.7 && < 1.11` | `1.10.6` |
| Calico      |                    | `3.1.3`  |
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

[0.4.1]: https://github.com/jetstack/tarmak/compare/0.4.0...0.4.1
[0.4.0]: https://github.com/jetstack/tarmak/compare/0.3.0...0.4.0
[0.3.0]: https://github.com/jetstack/tarmak/compare/0.2.0...0.3.0
[0.2.1]: https://github.com/jetstack/tarmak/compare/0.2.0...0.2.1
[0.2.0]: https://github.com/jetstack/tarmak/compare/0.1.2...0.2.0
[Unreleased]: https://github.com/jetstack/tarmak/compare/0.4.0...HEAD
