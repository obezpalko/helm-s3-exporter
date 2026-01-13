# GitHub Actions Workflows

This directory contains CI/CD workflows for the Helm Repository Exporter project.

> **📖 For detailed release instructions, see [RELEASE_PROCESS.md](../RELEASE_PROCESS.md)**

## Workflows

### 1. Build and Test (`build.yaml`)

**Triggers**: Push to `main`/`develop` branches, Pull Requests

**Jobs**:

- **Lint** - Code quality checks
  - Format checking with `gofmt`
  - Linting with `golangci-lint`
  
- **Test** - Run test suite
  - Unit tests with race detection
  - Coverage reporting to Codecov
  
- **Build** - Compile the application
  - Build Go binary
  - Upload artifact for download
  
- **Docker** - Build container image
  - Multi-stage Docker build
  - Cache optimization
  
- **Helm Lint** - Validate Helm chart
  - Chart linting
  - Template validation

**Features**:
- ✅ Go module caching for faster builds
- ✅ Parallel job execution
- ✅ Artifact upload (v4)
- ✅ Docker layer caching

### 2. Release (`release.yaml`)

**Triggers**: Push tags matching `v*` (e.g., `v1.2.3`)

**Jobs**:

- **Release** - Create GitHub release
  - Build binaries for multiple platforms:
    - Linux (amd64, arm64)
    - macOS (amd64, arm64)
  - Generate SHA256 checksums
  - Generate build attestations
  - Create release with auto-generated notes
  
- **Docker** - Build and push images
  - Multi-platform builds (amd64, arm64)
  - Push to GitHub Container Registry (ghcr.io)
  - Tag with semantic versions (v1.2.3, v1.2, v1, latest)
  - Generate and push image attestations
  
- **Helm** - Package and publish chart
  - Update chart version from git tag
  - Package Helm chart
  - Generate chart attestations
  - Publish to GitHub Pages
  - Update Helm repository index.yaml

**Required Secrets**:
- `GITHUB_TOKEN` - Automatically provided (used for GHCR, releases, and attestations)

**Features**:
- ✅ Multi-platform binary builds
- ✅ Automated semantic versioning
- ✅ Release notes generation
- ✅ Docker multi-arch support
- ✅ Build attestations for supply chain security
- ✅ Helm repository on GitHub Pages
- ✅ SLSA provenance compliance

**Note**: Docker images are published to GitHub Container Registry (ghcr.io). No additional secrets required!

## Usage

### Running Locally

Before pushing, ensure your code passes all checks:

```bash
# Format code
make fmt

# Run tests
make test

# Build
make build

# Lint (requires golangci-lint)
make lint

# Lint Helm chart
helm lint charts/helm-repo-exporter
```

### Creating a Release

> **📖 See [RELEASE_PROCESS.md](../RELEASE_PROCESS.md) for complete release instructions**

Quick steps:

1. **Update CHANGELOG.md** with the new version

2. **Commit changes**:
   ```bash
   git add CHANGELOG.md
   git commit -m "docs: update changelog for v1.2.3"
   git push
   ```

3. **Create and push tag**:
   ```bash
   git tag -a v1.2.3 -m "Release v1.2.3"
   git push origin v1.2.3
   ```

4. **Monitor workflow**:
   - Go to Actions tab in GitHub
   - Watch the release workflow
   - Verify release artifacts and attestations

**Note**: `Chart.yaml` uses `0.0.0-dev` as a placeholder. The actual version is automatically set during the release process.

### Setting Up Secrets

For the release workflow to work, configure these secrets in your repository:

1. Go to **Settings** → **Secrets and variables** → **Actions**

2. Add the following secrets:
   - `DOCKER_USERNAME`: Your Docker Hub username
   - `DOCKER_PASSWORD`: Docker Hub access token (not password!)

3. To create a Docker Hub token:
   - Go to Docker Hub → Account Settings → Security
   - Click "New Access Token"
   - Give it a name and copy the token
   - Use this as `DOCKER_PASSWORD`

## Workflow Status

Check the status of workflows:

- **Build Status**: Shows on README badges
- **Actions Tab**: Detailed logs and history
- **Pull Requests**: Status checks before merge

## Troubleshooting

### Build Failures

**Problem**: `directory not found` error

**Solution**: Ensure `go mod download` runs before build steps

**Problem**: Format check fails

**Solution**: Run `make fmt` locally before pushing

### Release Failures

**Problem**: Docker push fails

**Solution**: Check Docker Hub credentials in secrets

**Problem**: Multi-platform build fails

**Solution**: Ensure QEMU is set up (handled by `docker/setup-buildx-action`)

### Cache Issues

**Problem**: Builds are slow

**Solution**: Workflows use caching automatically. If issues persist:
- Check cache hit/miss in logs
- Clear cache: Settings → Actions → Caches

## Best Practices

1. **Always run locally first**:
   ```bash
   make fmt && make test && make build
   ```

2. **Use descriptive commit messages**:
   - Follow [Conventional Commits](https://www.conventionalcommits.org/)
   - Examples: `feat:`, `fix:`, `docs:`, `chore:`

3. **Test before tagging**:
   - Ensure main branch builds successfully
   - Run full test suite locally
   - Update CHANGELOG.md

4. **Semantic Versioning**:
   - MAJOR: Breaking changes
   - MINOR: New features
   - PATCH: Bug fixes

## Workflow Optimization

Current optimizations:
- ✅ Go module caching (`cache: true`)
- ✅ Docker layer caching (`cache-from: type=gha`)
- ✅ Parallel job execution
- ✅ Conditional job dependencies

Future improvements:
- [ ] Add integration tests
- [ ] Add security scanning (Trivy, Snyk)
- [ ] Add SBOM generation
- [ ] Add chart signing

## References

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Docker Build Push Action](https://github.com/docker/build-push-action)
- [Helm Chart Releaser](https://github.com/helm/chart-releaser-action)

