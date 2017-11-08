# Changelog
All notable changes to this project will be documented in this file.

## [0.8.0] - 2017-08-15
### Added
- vault-helper binary
- Docker image containing vault-helper binary saved to vault-helper-image.tar
- Tests for vault-helper
- Flags for subcommands on vault-helper

### Changed
- Entry point command in Docker image now displays help
- Updated README.md
- Upgraded vault in docker image to 0.7.3
- Docker ignores all except vault-helper binaries

### Removed
- vault-helper bash script
- vault-setup bash script
- No longer testing on the docker image through release
- Removed Gemfiles and Rakefile
