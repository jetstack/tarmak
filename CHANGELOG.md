# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## [Unreleased]:

### Added

### Changed

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
| Terraform   |                    | `0.10.6` |
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

[0.2.0]: https://github.com/jetstack/tarmak/compare/0.1.2...0.2.0
[Unreleased]: https://github.com/jetstack/tarmak/compare/0.2.0...HEAD
