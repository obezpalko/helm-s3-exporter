# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] - 2025-01-13

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
- Example configurations for various deployment scenarios
- GitHub Actions CI/CD pipeline with:
  - Automated builds for multiple platforms (Linux, macOS, amd64, arm64)
  - Docker multi-arch images published to GitHub Container Registry
  - Build attestations for supply chain security
  - Helm chart packaging and publishing
- Helm repository hosted via GitHub raw content
- Automated release process with semantic versioning
- Makefile for build automation

### Security
- Runs as non-root user (UID 65532)
- Read-only root filesystem
- Dropped all Linux capabilities
- Distroless base image for minimal attack surface
- Cryptographically signed build attestations for all artifacts
- SLSA provenance compliance

### Documentation
- Complete release process guide (`.github/RELEASE_PROCESS.md`)
- Contributing guidelines with versioning strategy
- Configuration examples for various scenarios
- Grafana dashboard and Prometheus queries

[Unreleased]: https://github.com/obezpalko/helm-repo-exporter/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/obezpalko/helm-repo-exporter/releases/tag/v0.1.0

