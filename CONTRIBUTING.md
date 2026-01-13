# Contributing to Helm Repository Exporter

Thank you for your interest in contributing to Helm Repository Exporter! This document provides guidelines and instructions for contributing.

## Code of Conduct

Please be respectful and constructive in all interactions with the project maintainers and community members.

## How to Contribute

### Reporting Bugs

Before creating a bug report, please check existing issues to avoid duplicates.

When filing a bug report, include:
- A clear, descriptive title
- Steps to reproduce the issue
- Expected vs actual behavior
- Your environment (Kubernetes version, cloud provider, etc.)
- Relevant logs and configuration
- Screenshots if applicable

### Suggesting Enhancements

Enhancement suggestions are tracked as GitHub issues. When creating an enhancement suggestion, include:
- A clear, descriptive title
- Detailed description of the proposed functionality
- Use cases and benefits
- Potential implementation approach (optional)

### Pull Requests

1. **Fork the repository** and create your branch from `main`
2. **Make your changes**
   - Follow the existing code style
   - Add tests for new functionality
   - Update documentation as needed
3. **Test your changes**
   - Run `make test` to ensure tests pass
   - Test manually if applicable
4. **Commit your changes**
   - Use clear, descriptive commit messages
   - Follow [Conventional Commits](https://www.conventionalcommits.org/)
5. **Push to your fork** and submit a pull request

## Development Setup

### Prerequisites

- Go 1.21 or later
- Docker (for building images)
- Kubernetes cluster (for testing)
- Helm 3.x
- Make

### Setting Up Development Environment

```bash
# Clone the repository
git clone https://github.com/obezpalko/helm-repo-exporter.git
cd helm-repo-exporter

# Install dependencies
make deps

# Build the project
make build

# Run tests
make test

# Run locally
export CONFIG_FILE=examples/config-single.yaml
make run
```

### Running Tests

```bash
# Run all tests
make test

# Run with coverage
go test -v -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Building Docker Image

```bash
# Build locally
make docker-build

# Build with custom tag
make docker-build VERSION=dev
```

### Testing Helm Chart

```bash
# Lint the chart
make helm-lint

# Test installation (dry-run)
helm install test-release ./charts/helm-repo-exporter \
  --set config.inline.enabled=true \
  --set config.inline.url=https://charts.bitnami.com/bitnami/index.yaml \
  --dry-run --debug

# Install to test cluster
helm install test-release ./charts/helm-repo-exporter \
  --set config.inline.enabled=true \
  --set config.inline.url=https://charts.bitnami.com/bitnami/index.yaml
```

## Code Style

### Go Code

- Follow standard Go conventions
- Use `gofmt` for formatting
- Run `golangci-lint` before submitting
- Keep functions small and focused
- Add comments for exported functions and types

```bash
# Format code
make fmt

# Run linter
make lint
```

### Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
feat: add support for custom repository endpoints
fix: correct metric label for chart names
docs: update README with new examples
chore: update dependencies
test: add tests for analyzer package
```

### Branch Naming

- `feature/description` - New features
- `fix/description` - Bug fixes
- `docs/description` - Documentation updates
- `chore/description` - Maintenance tasks

## Testing Guidelines

### Unit Tests

- Write tests for all new functionality
- Maintain or improve code coverage
- Use table-driven tests where appropriate
- Mock external dependencies

Example:
```go
func TestAnalyzeCharts(t *testing.T) {
    tests := []struct {
        name     string
        index    *HelmIndex
        expected *ChartAnalysis
    }{
        // test cases
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := AnalyzeCharts(tt.index)
            // assertions
        })
    }
}
```

### Integration Tests

For integration tests with repositories:
- Use LocalStack or MinIO for local testing
- Don't hardcode credentials in tests
- Clean up resources after tests

## Documentation

Update documentation when:
- Adding new features
- Changing existing behavior
- Adding new configuration options
- Fixing bugs that affect usage

Documentation to update:
- README.md - Main documentation
- README.md - Main documentation
- charts/helm-repo-exporter/values.yaml - Configuration comments
- examples/ - Example configurations

## Release Process

Releases are automated via GitHub Actions:

### Versioning Strategy

- **Chart.yaml versions**: The `Chart.yaml` file in the repository uses `0.0.0-dev` as a placeholder. The actual version is automatically set during the release process based on the git tag.
- **Git tags**: Create a new release by pushing a semantic version tag (e.g., `v1.2.3`)
- **Automated workflow**: The release workflow automatically:
  - Updates `Chart.yaml` with the tag version
  - Builds and publishes Docker images
  - Packages and publishes the Helm chart
  - Creates GitHub release with attestations
  - Updates the Helm repository index on GitHub Pages

### Creating a Release

1. **Update CHANGELOG.md** with the new version and changes
2. **Commit changes** to the main branch
3. **Create and push a tag**:
   ```bash
   git tag -a v1.2.3 -m "Release v1.2.3"
   git push origin v1.2.3
   ```
4. **Monitor the workflow** at `.github/workflows/release.yaml`
5. The release will be automatically created with:
   - Compiled binaries for multiple platforms
   - Docker images with attestations
   - Helm chart published to GitHub Pages
   - Release notes generated from commits

### Version Numbers

Follow [Semantic Versioning](https://semver.org/):
- **MAJOR** (v2.0.0): Breaking changes
- **MINOR** (v1.1.0): New features, backwards compatible
- **PATCH** (v1.0.1): Bug fixes, backwards compatible

## Questions?

- Open a [GitHub Discussion](https://github.com/obezpalko/helm-repo-exporter/discussions)
- Check existing [Issues](https://github.com/obezpalko/helm-repo-exporter/issues)

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

