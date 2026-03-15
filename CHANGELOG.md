# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- New resource `pangolin_site`: create and manage Pangolin sites (Newt-type network endpoints).
  - Supports `name`, `newt_id`, `secret`, `address`, and `type` fields.
  - `newt_id`, `secret`, `type`, and `org_id` are immutable and force recreation on change.
  - `name` and `address` are updatable in-place.
  - Supports import via `org_id/site_id`.
- New client methods: `GetSite`, `UpdateSite`, `DeleteSite`.

### Fixed

- `pangolin_site_resource` create was calling the non-existent `/org/{orgId}/private-resource` endpoint; corrected to `/org/{orgId}/site-resource`.

### Changed

- `client.CreateSite` now accepts a `*Site` struct instead of a plain name string, enabling full field control.
- `Site` struct extended with `NewtID`, `Secret`, `Address`, and `Type` fields.

## 1.0.0 (2026-02-08)


### Bug Fixes

* add missing os import in target acceptance test ([3f34515](https://github.com/groteck/pangolin-tf/commit/3f345151cf6b814fb7d74c13c96efc47faecf517))
* lint errors and update provider version handling ([35fb230](https://github.com/groteck/pangolin-tf/commit/35fb2300495f89aa81e77af1556750264b5beea3))
* **pipeline:** Fix linting and tests ([92a69de](https://github.com/groteck/pangolin-tf/commit/92a69defc03c40793fef3697e5556d5ac76747cc))

## [0.1.0] - 2026-02-08

### Added
- Initial release of the Pangolin Terraform provider.
- Supported Resources:
    - `pangolin_role`: Manage organization roles.
    - `pangolin_site_resource`: Manage Host/CIDR resources.
    - `pangolin_resource`: Manage App resources (HTTP/TCP/UDP).
    - `pangolin_target`: Manage backend targets for resources.
- Supported Data Sources:
    - `pangolin_role`: Fetch role ID by name.
    - `pangolin_site`: Fetch site ID by name.
- Provider authentication via token and environment variables.
- Automated documentation generation via `tfplugindocs`.
- Multi-architecture release automation via GoReleaser.
- Docker-based acceptance testing suite.
