# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.3.0] - 2025-01-13

### Added
- Build attestations for all release artifacts (binaries, Docker images, Helm charts)
- Helm repository published to GitHub Pages with automated index.yaml generation
- Comprehensive release process documentation (`.github/RELEASE_PROCESS.md`)
- SLSA provenance compliance for supply chain security
- Automated semantic versioning from git tags

### Changed
- Chart.yaml now uses `0.0.0-dev` as a placeholder version in the repository
- Actual chart version is automatically set during release based on git tags
- Release workflow enhanced with `chart-releaser-action` for GitHub Pages publishing
- Updated CONTRIBUTING.md with detailed release process and versioning strategy
- Improved workflow documentation with attestation and GitHub Pages details

### Fixed
- Fixed version mismatch between released charts and Chart.yaml in repository
- Helm charts now properly published to a Helm repository (not just GitHub Releases)

### Security
- All release artifacts now include cryptographically signed build attestations
- Verifiable provenance for binaries, Docker images, and Helm charts
- Enhanced supply chain security with OIDC authentication

## [0.1.0] - 2024-12-23

### Added
- Initial release of Helm Repository Exporter
- HTTP client for fetching Helm repository index.yaml files
- Chart analyzer for processing Helm chart metadata
- Prometheus metrics exporter with comprehensive metrics:
  - Total charts and versions
  - Per-chart version counts
  - Age tracking (oldest, newest, median)
  - Scrape duration and error metrics
- Optional HTML dashboard for visualizing charts
- Kubernetes deployment via Helm chart
- Support for multiple authentication methods (Basic Auth, Bearer tokens, custom headers)
- Configurable scan intervals and timeouts
- Health and readiness probes
- ServiceMonitor for Prometheus Operator integration
- Comprehensive documentation and examples
- Security best practices with non-root containers
- Example configurations for various deployment scenarios
- GitHub Actions CI/CD pipeline
- Makefile for build automation

### Security
- Runs as non-root user (UID 65532)
- Read-only root filesystem
- Dropped all Linux capabilities
- Distroless base image for minimal attack surface

[Unreleased]: https://github.com/obezpalko/helm-repo-exporter/compare/v0.3.0...HEAD
[0.3.0]: https://github.com/obezpalko/helm-repo-exporter/compare/v0.1.0...v0.3.0
[0.1.0]: https://github.com/obezpalko/helm-repo-exporter/releases/tag/v0.1.0

