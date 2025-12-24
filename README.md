# Helm S3 Exporter

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Report Card](https://goreportcard.com/badge/github.com/obezpalko/helm-s3-exporter)](https://goreportcard.com/report/github.com/obezpalko/helm-s3-exporter)

A Prometheus exporter for Helm chart repositories. Monitor multiple chart repositories with flexible authentication support.

## Overview

The Helm S3 Exporter monitors Helm chart repositories by periodically fetching and analyzing their `index.yaml` files. It exposes Prometheus metrics about charts, versions, and repository health.

## Features

✅ **Multiple Repositories** - Monitor multiple chart repositories simultaneously  
✅ **Flexible Authentication** - Support for Basic Auth, Bearer tokens, and custom headers  
✅ **HTTP/HTTPS Support** - Fetch from any accessible URL (including S3, GitHub, Artifactory, Harbor, etc.)  
✅ **Kubernetes Native** - Designed to run in Kubernetes with full cloud-native support  
✅ **Prometheus Metrics** - Exposes detailed metrics about charts, versions, and ages  
✅ **ServiceMonitor Ready** - First-class Prometheus Operator support  
✅ **Optional HTML Dashboard** - Human-friendly web interface  
✅ **Lightweight** - Minimal resource footprint  
✅ **Secure** - Runs as non-root with read-only filesystem

## Quick Start

### Simple Single Repository

```bash
# Install with Helm (single public repository)
helm install helm-s3-exporter ./charts/helm-s3-exporter \
  --set config.inline.enabled=true \
  --set config.inline.url=https://charts.bitnami.com/bitnami/index.yaml \
  --namespace monitoring --create-namespace
```

### Multiple Repositories with Authentication

```bash
# 1. Create config file
cat > config.yaml <<EOF
repositories:
  - name: bitnami
    url: https://charts.bitnami.com/bitnami/index.yaml
  
  - name: private-repo
    url: https://charts.company.com/index.yaml
    auth:
      basic:
        username: myuser
        password: mypassword

scanInterval: 10m
EOF

# 2. Create secret
kubectl create secret generic helm-repo-config \
  --from-file=config.yaml=config.yaml \
  --namespace monitoring

# 3. Install with Helm
helm install helm-s3-exporter ./charts/helm-s3-exporter \
  --set config.existingSecret.enabled=true \
  --set config.existingSecret.name=helm-repo-config \
  --namespace monitoring --create-namespace
```

## Configuration

### Configuration Methods

The exporter supports three configuration methods:

1. **Inline Config** - Simple single repository
2. **Secret** - Multiple repositories with authentication (recommended)
3. **ConfigMap** - Multiple public repositories

### Method 1: Inline Configuration (Simple)

Best for: Single public repository

```yaml
# values.yaml
config:
  inline:
    enabled: true
    url: "https://charts.bitnami.com/bitnami/index.yaml"
```

Deploy:
```bash
helm install my-exporter ./charts/helm-s3-exporter -f values.yaml
```

### Method 2: Secret (Recommended for Private Repos)

Best for: Multiple repositories with authentication

**Step 1: Create config.yaml**

```yaml
repositories:
  # Public repository
  - name: bitnami
    url: https://charts.bitnami.com/bitnami/index.yaml
  
  # Private with Basic Auth
  - name: company-charts
    url: https://charts.company.com/index.yaml
    auth:
      basic:
        username: helm-user
        password: secret-password
  
  # Private with Bearer Token
  - name: github-private
    url: https://raw.githubusercontent.com/company/charts/main/index.yaml
    auth:
      bearerToken: ghp_xxxxxxxxxxxx
  
  # Custom headers (API keys, etc.)
  - name: custom-auth
    url: https://charts.example.com/index.yaml
    auth:
      headers:
        X-API-Key: your-api-key
        X-Custom-Header: custom-value

scanInterval: 10m
scanTimeout: 1m
metricsPort: "9571"
enableHTML: true
```

**Step 2: Create Secret**

```bash
kubectl create secret generic helm-repo-config \
  --from-file=config.yaml=config.yaml \
  --namespace monitoring
```

**Step 3: Deploy**

```bash
helm install my-exporter ./charts/helm-s3-exporter \
  --namespace monitoring \
  --set config.existingSecret.enabled=true \
  --set config.existingSecret.name=helm-repo-config
```

### Method 3: ConfigMap (Public Repos Only)

Best for: Multiple public repositories without sensitive data

```bash
kubectl create configmap helm-repo-config \
  --from-file=config.yaml=config.yaml \
  --namespace monitoring

helm install my-exporter ./charts/helm-s3-exporter \
  --namespace monitoring \
  --set config.existingConfigMap.enabled=true \
  --set config.existingConfigMap.name=helm-repo-config
```

### Authentication Methods

#### Basic Authentication

```yaml
repositories:
  - name: private-repo
    url: https://charts.company.com/index.yaml
    auth:
      basic:
        username: myuser
        password: mypassword
```

#### Bearer Token

```yaml
repositories:
  - name: github-private
    url: https://raw.githubusercontent.com/company/charts/main/index.yaml
    auth:
      bearerToken: ghp_xxxxxxxxxxxxxxxxxxxx
```

#### Custom Headers (API Keys, etc.)

```yaml
repositories:
  - name: api-protected
    url: https://charts.example.com/index.yaml
    auth:
      headers:
        X-API-Key: your-api-key
        X-Request-ID: unique-id
```

#### Combined Authentication

```yaml
repositories:
  - name: multi-auth
    url: https://charts.example.com/index.yaml
    auth:
      bearerToken: token123
      headers:
        X-Custom-Header: value
```

### Configuration File Reference

Full configuration file options:

```yaml
# List of repositories to monitor
repositories:
  - name: string              # Friendly name for the repository
    url: string               # URL to index.yaml
    scanInterval: duration    # Optional: Override default interval (e.g., 5m, 1h, 30s)
    auth:                     # Optional authentication
      basic:                  # Basic authentication
        username: string
        password: string
      bearerToken: string     # Bearer token
      headers:                # Custom headers
        Header-Name: value

# Scan configuration
scanInterval: duration        # Default scan interval (e.g., 5m, 1h) [default: 5m]
                              # Repos can override with their own scanInterval
scanTimeout: duration         # Timeout for each scan [default: 30s]

# Server configuration
metricsPort: string           # Port for metrics [default: "9571"]
metricsPath: string           # Path for metrics [default: "/metrics"]

# Optional HTML dashboard
enableHTML: bool              # Enable HTML dashboard [default: false]
htmlPath: string              # Path for HTML [default: "/charts"]
```

See [examples/CONFIGURATION_GUIDE.md](examples/CONFIGURATION_GUIDE.md) for comprehensive examples.

## Prometheus Metrics

### Available Metrics

All metrics include a `repository` label to filter by specific repositories.

| Metric | Type | Description | Labels |
|--------|------|-------------|--------|
| `helm_s3_charts_total` | Gauge | Total number of unique charts | `repository` |
| `helm_s3_versions_total` | Gauge | Total number of chart versions | `repository` |
| `helm_s3_chart_versions` | Gauge | Number of versions for each chart | `repository`, `chart` |
| `helm_s3_chart_age_oldest_seconds` | Gauge | Unix timestamp of the oldest version | `repository`, `chart` |
| `helm_s3_chart_age_newest_seconds` | Gauge | Unix timestamp of the newest version | `repository`, `chart` |
| `helm_s3_chart_age_median_seconds` | Gauge | Unix timestamp of the median version | `repository`, `chart` |
| `helm_s3_overall_age_oldest_seconds` | Gauge | Unix timestamp of the oldest chart in the repository | `repository` |
| `helm_s3_overall_age_newest_seconds` | Gauge | Unix timestamp of the newest chart in the repository | `repository` |
| `helm_s3_overall_age_median_seconds` | Gauge | Unix timestamp of the median chart in the repository | `repository` |
| `helm_s3_scrape_duration_seconds` | Histogram | Duration of the scrape operation | `repository` |
| `helm_s3_scrape_errors_total` | Counter | Total number of scrape errors | `repository` |
| `helm_s3_last_scrape_success` | Gauge | Unix timestamp of the last successful scrape | `repository` |

### Example Queries

```promql
# Total number of charts across all repositories
sum(helm_s3_charts_total)

# Total charts per repository
helm_s3_charts_total

# Total charts in production repository
helm_s3_charts_total{repository="production"}

# Charts with more than 10 versions in bitnami repository
helm_s3_chart_versions{repository="bitnami"} > 10

# Total charts across specific repositories
sum(helm_s3_charts_total{repository=~"production|staging"})

# Age of the newest chart in hours for a specific repository
(time() - helm_s3_overall_age_newest_seconds{repository="production"}) / 3600

# Scrape duration per repository
helm_s3_scrape_duration_seconds_sum{repository="bitnami"} / helm_s3_scrape_duration_seconds_count{repository="bitnami"}

# Scrape error rate per repository
rate(helm_s3_scrape_errors_total{repository="production"}[5m])

# Charts that haven't been updated in 90 days
(time() - helm_s3_chart_age_newest_seconds) / 86400 > 90

# Compare chart counts across repositories
sum by (repository) (helm_s3_chart_versions)

# Find repositories with scrape errors
sum by (repository) (increase(helm_s3_scrape_errors_total[1h])) > 0
```

## Helm Chart Values

### Key Configuration Options

| Parameter | Description | Default |
|-----------|-------------|---------|
| `config.inline.enabled` | Enable inline config | `false` |
| `config.inline.url` | Repository URL (if inline enabled) | `""` |
| `config.existingSecret.enabled` | Use existing secret | `false` |
| `config.existingSecret.name` | Secret name | `""` |
| `config.existingConfigMap.enabled` | Use existing configmap | `false` |
| `config.existingConfigMap.name` | ConfigMap name | `""` |
| `scan.interval` | Scan interval | `"5m"` |
| `scan.timeout` | Scan timeout | `"30s"` |
| `serviceMonitor.enabled` | Enable ServiceMonitor | `false` |
| `serviceMonitor.additionalLabels` | Labels for ServiceMonitor | `{}` |
| `html.enabled` | Enable HTML dashboard | `false` |
| `resources.limits.memory` | Memory limit | `128Mi` |
| `resources.requests.memory` | Memory request | `64Mi` |

See [charts/helm-s3-exporter/values.yaml](charts/helm-s3-exporter/values.yaml) for all options.

## Examples

### Public Repositories

```bash
helm install bitnami-exporter ./charts/helm-s3-exporter \
  -f examples/values-simple-url.yaml
```

### Multiple Repositories with Auth

```bash
# Create secret
kubectl create secret generic helm-repos \
  --from-file=config.yaml=examples/config-multi-auth.yaml

# Deploy
helm install multi-exporter ./charts/helm-s3-exporter \
  -f examples/values-multi-repos.yaml
```

### S3 Bucket (Public)

```yaml
# values.yaml
config:
  inline:
    enabled: true
    url: "https://my-charts.s3.us-west-2.amazonaws.com/index.yaml"
```

### GitHub Repository

```yaml
repositories:
  - name: my-charts
    url: https://raw.githubusercontent.com/myorg/helm-charts/main/index.yaml
    auth:
      bearerToken: ${GITHUB_TOKEN}
```

See [examples/](examples/) directory for more examples.

## Prometheus Integration

### ServiceMonitor (Prometheus Operator)

```yaml
serviceMonitor:
  enabled: true
  additionalLabels:
    prometheus: kube-prometheus
  interval: 30s
```

Or set from command line:

```bash
helm install my-exporter ./charts/helm-s3-exporter \
  --set serviceMonitor.enabled=true \
  --set serviceMonitor.additionalLabels.prometheus=kube-prometheus
```

### Manual Scrape Configuration

```yaml
# prometheus.yml
scrape_configs:
  - job_name: 'helm-s3-exporter'
    static_configs:
      - targets: ['helm-s3-exporter.monitoring.svc.cluster.local:9571']
    scrape_interval: 30s
```

### Alerting Rules

Example alerts in [examples/prometheus-rules.yaml](examples/prometheus-rules.yaml):

```yaml
groups:
  - name: helm-repository
    rules:
      - alert: HelmRepositoryScrapeError
        expr: increase(helm_s3_scrape_errors_total[5m]) > 0
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Failed to scrape Helm repository"
```

## Development

### Prerequisites

- Go 1.21+
- Docker
- kubectl
- Helm 3.x
- Make (optional)

### Building

```bash
# Build binary
make build

# Run tests
make test

# Run linter
make lint

# Build Docker image
make docker-build
```

### Local Testing

```bash
# Run locally
export CONFIG_FILE=examples/config-single.yaml
go run cmd/exporter/main.go

# Or with environment variables
export INDEX_URL=https://charts.bitnami.com/bitnami/index.yaml
export SCAN_INTERVAL=1m
go run cmd/exporter/main.go
```

Access metrics at http://localhost:9571/metrics

## Troubleshooting

### Common Issues

#### "failed to read config file"

**Solution**: Verify the secret/configmap exists and is mounted correctly:

```bash
kubectl get secret helm-repo-config -n monitoring
kubectl describe pod <pod-name> -n monitoring
```

#### "403 Forbidden" or "401 Unauthorized"

**Solution**: Check authentication credentials:

```bash
# Test manually
curl -u username:password https://charts.example.com/index.yaml
curl -H "Authorization: Bearer token" https://charts.example.com/index.yaml
```

#### High memory usage

**Solution**: Increase scan interval or add resource limits:

```yaml
scan:
  interval: 15m

resources:
  limits:
    memory: 256Mi
```

### Viewing Logs

```bash
# Check logs
kubectl logs -n monitoring -l app.kubernetes.io/name=helm-s3-exporter

# Follow logs
kubectl logs -n monitoring -l app.kubernetes.io/name=helm-s3-exporter -f
```

## Architecture

```
┌─────────────────────────────────────────────┐
│           Helm S3 Exporter                  │
│                                             │
│  ┌────────────────────────────────────┐   │
│  │      Config Loader                 │   │
│  │  (Secret/ConfigMap/Inline)         │   │
│  └────────────────────────────────────┘   │
│                  │                         │
│  ┌───────────────┴────────────────────┐   │
│  │     HTTP Fetcher (Multi-Repo)      │   │
│  │  - Basic Auth                      │   │
│  │  - Bearer Token                    │   │
│  │  - Custom Headers                  │   │
│  └───────────────┬────────────────────┘   │
│                  │                         │
│  ┌───────────────┴────────────────────┐   │
│  │      Index.yaml Parser             │   │
│  └───────────────┬────────────────────┘   │
│                  │                         │
│  ┌───────────────┴────────────────────┐   │
│  │     Chart Analyzer                 │   │
│  └───────────────┬────────────────────┘   │
│                  │                         │
│  ┌───────────────┴────────────────────┐   │
│  │  Prometheus Metrics Exporter       │   │
│  │  Optional HTML Dashboard           │   │
│  └────────────────────────────────────┘   │
└─────────────────────────────────────────────┘
               │
               ▼
        Prometheus Server
```

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Credits

Developed by [@obezpalko](https://github.com/obezpalko)

## Support

- 📖 [Documentation](examples/CONFIGURATION_GUIDE.md)
- 🐛 [Issue Tracker](https://github.com/obezpalko/helm-s3-exporter/issues)
- 💬 [Discussions](https://github.com/obezpalko/helm-s3-exporter/discussions)
