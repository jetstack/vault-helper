# Changelog
All notable changes to this project will be documented in this file.

### Changed
- Use vault 0.9.6

## [0.9.13] - 2018-07-10

### Fixed
- Prevent panic during ensure dry runs PR #29, issue #28

## [0.9.12] - 2018-06-14

### Changed
- Ensure both secrets and configmaps are stored with encryption at rest [PR
  26](https://github.com/jetstack/vault-helper/pull/26)

## [0.9.11] - 2018-06-12

### Fixed
- Migrate `docker_*` build targets to use docker cp [PR 25](https://github.com/jetstack/vault-helper/pull/25)
- Improve vault backend names [PR 24](https://github.com/jetstack/vault-helper/pull/24)

### Changed
- Write the kubernetes encryption config file to vault [PR 23](https://github.com/jetstack/vault-helper/pull/23)
- Add dryrun, delete and versioning for kubernetes vault provider [PR 20](https://github.com/jetstack/vault-helper/pull/20)

## [0.9.10] - 2018-04-19

### Fixed

- Certificates are now getting renewed on every command run #22

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
