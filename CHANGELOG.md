# Changelog
All notable changes to this project will be documented in this file.

## [0.9.9] - 2018-04-19

### Fixed

- Add internal EC2 domain for us-east-1 #18

### Changed

- Added CLI tests #15

## [0.9.8] - 2018-04-10

### Fixed
- Parsing of flags for `kubeconfig` #11

### Changed
- Mark kube-apiserver's certificate as client cert #15

## [0.9.7] - 2018-03-21
### Fixed
- Allow usage of AWS SAN DNS names in kubelet's role

## [0.9.6] - 2018-03-07
### Fixed
- Do not use CA names in SAN DNS names

### Changed
- Use vault 0.9.5
- Upgrade golang dependencies
- Upgrade golang to 1.10

## [0.9.5] - 2018-03-06
### Fixed
- Support ca_chain properly for reading certificates (#6)

## [0.9.4] - 2018-03-06
### Fixed
- Ensure vault-helper setup is not using instance tokens

## [0.9.3] - 2018-02-12
### Changed
- Use Update to use lowercase logrus import

## [0.9.2] - 2017-11-23
### Fixed
- Fix role for kube-apiserver-proxy, allow only bare domains

## [0.9.1] - 2017-11-23
### Added
- Sign binaries using GPG key

## [0.9.0] - 2017-11-22
### Added
- Add additonal CA for Kubernetes API server's proxy clients. This enables
  running API aggregation on a kubernetes cluster

### Changed
- Move the repository from jetstack-experimental to jetstack
- Updated to Golang 1.9.2

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
