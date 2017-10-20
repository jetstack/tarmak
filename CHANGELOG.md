# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## [Unreleased]:

### Added

### Changed

## 0.1.2 - 2017-10-19

### Initial release
- First public release
- Support for AWS provider only
- Prepare and drive infrastructure updates using Terraform
- Prepare configuration updates using Puppet and drive them using Wing on the
  instances
- Provides wrappers for basic administrative task: kubectl, ssh

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

[Unreleased]: https://github.com/jetstack/tarmak/compare/0.1.2...HEAD
