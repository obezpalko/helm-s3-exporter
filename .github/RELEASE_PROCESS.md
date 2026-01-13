# Release Process

This document explains how releases work for the Helm Repository Exporter project.

## Overview

Releases are fully automated using GitHub Actions. When you push a version tag, the workflow automatically:

1. Builds binaries for multiple platforms
2. Creates Docker images with multi-arch support
3. Packages and publishes the Helm chart
4. Generates build attestations for supply chain security
5. Creates a GitHub Release with release notes
6. Updates the Helm repository on GitHub Pages

## Versioning Strategy

### Chart.yaml Versioning

The `Chart.yaml` file in the repository uses **`0.0.0-dev`** as a placeholder version. This is intentional:

- **In the repository**: `version: 0.0.0-dev` indicates this is development code
- **During release**: The workflow automatically updates the version to match the git tag
- **Published chart**: Contains the actual release version (e.g., `1.2.3`)

**Why this approach?**
- Prevents version mismatches between the repository and released artifacts
- Makes it clear which code is released vs. in-development
- Eliminates the need to commit version bumps back to the repository
- Follows best practices for automated releases

### Semantic Versioning

This project follows [Semantic Versioning](https://semver.org/):

- **MAJOR** version (v2.0.0): Incompatible API changes or breaking changes
- **MINOR** version (v1.1.0): New functionality in a backwards compatible manner
- **PATCH** version (v1.0.1): Backwards compatible bug fixes

## Creating a Release

### Prerequisites

- Write access to the repository
- All changes merged to the `main` branch
- CHANGELOG.md updated with the new version

### Steps

1. **Update CHANGELOG.md**

   Add a new section for the release version:

   ```markdown
   ## [1.2.3] - 2024-01-15

   ### Added
   - New feature description

   ### Changed
   - Changed behavior description

   ### Fixed
   - Bug fix description
   ```

2. **Commit and push changes**

   ```bash
   git add CHANGELOG.md
   git commit -m "docs: update changelog for v1.2.3"
   git push origin main
   ```

3. **Create and push the version tag**

   ```bash
   # Create an annotated tag
   git tag -a v1.2.3 -m "Release v1.2.3"
   
   # Push the tag to trigger the release workflow
   git push origin v1.2.3
   ```

4. **Monitor the workflow**

   - Go to: https://github.com/obezpalko/helm-repo-exporter/actions
   - Watch the "Release" workflow run
   - The workflow has three jobs that run in parallel:
     - `release`: Builds binaries and creates GitHub Release
     - `docker`: Builds and pushes Docker images
     - `helm`: Packages and publishes Helm chart

5. **Verify the release**

   After the workflow completes, verify:

   - **GitHub Release**: https://github.com/obezpalko/helm-repo-exporter/releases
     - Binaries are attached
     - Release notes are generated
     - Attestations are created
   
   - **Docker Images**: https://github.com/obezpalko/helm-repo-exporter/pkgs/container/helm-repo-exporter
     - Images are tagged with version, major.minor, major, and latest
     - Attestations are visible
   
   - **Helm Chart**: https://obezpalko.github.io/helm-repo-exporter/
     - `index.yaml` is updated
     - Chart is available via `helm repo add`

## Release Workflow Details

### Job: `release` (Binaries)

Builds Go binaries for multiple platforms:
- linux/amd64
- linux/arm64
- darwin/amd64 (macOS Intel)
- darwin/arm64 (macOS Apple Silicon)

Creates:
- Compiled binaries
- SHA256 checksums
- Build attestations
- GitHub Release with auto-generated notes

### Job: `docker` (Container Images)

Builds multi-architecture Docker images:
- linux/amd64
- linux/arm64

Publishes to GitHub Container Registry (ghcr.io) with tags:
- `v1.2.3` (exact version)
- `v1.2` (major.minor)
- `v1` (major)
- `latest` (latest release)

Creates build attestations pushed to the registry.

### Job: `helm` (Helm Chart)

Packages the Helm chart and publishes to GitHub Pages:
1. Updates `Chart.yaml` with the release version
2. Packages the chart into a `.tgz` file
3. Creates build attestation for the chart
4. Uploads chart to GitHub Release
5. Updates `index.yaml` on GitHub Pages
6. Makes chart available via Helm repository

## Using Released Artifacts

### Binaries

Download from GitHub Releases:

```bash
# Download for Linux AMD64
wget https://github.com/obezpalko/helm-repo-exporter/releases/download/v1.2.3/exporter-linux-amd64

# Verify with checksums
wget https://github.com/obezpalko/helm-repo-exporter/releases/download/v1.2.3/checksums.txt
sha256sum -c checksums.txt

# Verify attestation
gh attestation verify exporter-linux-amd64 --owner obezpalko
```

### Docker Images

Pull from GitHub Container Registry:

```bash
# Pull specific version
docker pull ghcr.io/obezpalko/helm-repo-exporter:v1.2.3

# Pull latest
docker pull ghcr.io/obezpalko/helm-repo-exporter:latest

# Verify attestation
gh attestation verify oci://ghcr.io/obezpalko/helm-repo-exporter:v1.2.3 --owner obezpalko
```

### Helm Chart

Add the Helm repository and install:

```bash
# Add the repository
helm repo add helm-repo-exporter https://obezpalko.github.io/helm-repo-exporter

# Update repositories
helm repo update

# Install specific version
helm install my-exporter helm-repo-exporter/helm-repo-exporter --version 1.2.3

# Install latest version
helm install my-exporter helm-repo-exporter/helm-repo-exporter

# Download chart for verification
helm pull helm-repo-exporter/helm-repo-exporter --version 1.2.3
gh attestation verify helm-repo-exporter-1.2.3.tgz --owner obezpalko
```

## Build Attestations

All release artifacts include cryptographically signed build attestations that provide:

- **Provenance**: Verifiable proof of how the artifact was built
- **Integrity**: Ensures artifacts haven't been tampered with
- **Transparency**: Public record of the build process
- **Supply Chain Security**: Meets SLSA (Supply-chain Levels for Software Artifacts) standards

Attestations are created using GitHub's `attest-build-provenance` action and can be verified using the `gh` CLI.

## Troubleshooting

### Release workflow fails

1. Check the workflow logs in GitHub Actions
2. Common issues:
   - Tag format must be `v*` (e.g., `v1.2.3`)
   - Ensure GitHub Pages is enabled in repository settings
   - Verify permissions are correct in the workflow file

### Chart not appearing in Helm repository

1. Verify GitHub Pages is enabled:
   - Go to Settings → Pages
   - Source should be "Deploy from a branch"
   - Branch should be `gh-pages`
2. Check that the workflow completed successfully
3. Wait a few minutes for GitHub Pages to update
4. Verify `index.yaml` exists: https://obezpalko.github.io/helm-repo-exporter/index.yaml

### Attestation verification fails

1. Ensure you have the latest `gh` CLI: `gh version`
2. Authenticate with GitHub: `gh auth login`
3. Use the correct owner name: `--owner obezpalko`
4. Download the artifact from the official release

## First-Time Setup

### Enable GitHub Pages

Before the first release, enable GitHub Pages:

1. Go to repository Settings → Pages
2. Under "Source", select "Deploy from a branch"
3. Select the `gh-pages` branch (will be created on first release)
4. Click "Save"

The `gh-pages` branch will be automatically created by the `chart-releaser-action` on the first release.

### Verify Permissions

The workflow requires these permissions (already configured):
- `contents: write` - Create releases and push to gh-pages
- `packages: write` - Push Docker images
- `id-token: write` - OIDC authentication for attestations
- `attestations: write` - Create build attestations

## Questions?

- Open a [GitHub Discussion](https://github.com/obezpalko/helm-repo-exporter/discussions)
- Check existing [Issues](https://github.com/obezpalko/helm-repo-exporter/issues)
- Review the [CONTRIBUTING.md](../CONTRIBUTING.md) guide
