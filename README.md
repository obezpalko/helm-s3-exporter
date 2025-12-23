# Helm S3 Exporter

<p align="center">
  <img src="icon.svg" alt="Helm S3 Exporter Logo" width="150" height="150">
</p>

<p align="center">
  <a href="LICENSE"><img src="https://img.shields.io/badge/License-MIT-blue.svg" alt="License"></a>
  <a href="https://go.dev/"><img src="https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go" alt="Go Version"></a>
  <a href="https://github.com/obezpalko/helm-s3-exporter/actions"><img src="https://github.com/obezpalko/helm-s3-exporter/workflows/Build%20and%20Test/badge.svg" alt="Build Status"></a>
</p>

A Prometheus exporter for Helm charts stored in AWS S3 buckets. This exporter helps administrators analyze and monitor the state of Helm chart repositories by collecting data from S3 and exposing comprehensive metrics.

## Features

✅ **Kubernetes Native** - Designed to run in Kubernetes with full cloud-native support  
✅ **Prometheus Metrics** - Exposes detailed metrics about charts, versions, and ages  
✅ **Flexible Authentication** - Supports both IAM roles (IRSA) and static credentials  
✅ **Configurable Scanning** - Adjustable scan intervals and timeouts  
✅ **Optional HTML Dashboard** - Beautiful web interface to visualize chart information  
✅ **ServiceMonitor Support** - First-class support for Prometheus Operator  
✅ **Secure by Default** - Runs as non-root with minimal privileges  

## Architecture

The exporter periodically scans an S3 bucket for Helm repository `index.yaml` files, analyzes the chart metadata, and exposes the following information:

- Total number of distinct charts
- Number of versions per chart
- Age information (oldest, newest, median) for each chart
- Overall repository statistics

## Metrics

The exporter provides the following Prometheus metrics:

| Metric Name | Type | Description | Labels |
|-------------|------|-------------|--------|
| `helm_s3_charts_total` | Gauge | Total number of distinct Helm charts | - |
| `helm_s3_versions_total` | Gauge | Total number of chart versions across all charts | - |
| `helm_s3_chart_versions` | Gauge | Number of versions for each chart | `chart` |
| `helm_s3_chart_age_oldest_seconds` | Gauge | Unix timestamp of the oldest version | `chart` |
| `helm_s3_chart_age_newest_seconds` | Gauge | Unix timestamp of the newest version | `chart` |
| `helm_s3_chart_age_median_seconds` | Gauge | Unix timestamp of the median version | `chart` |
| `helm_s3_overall_age_oldest_seconds` | Gauge | Unix timestamp of the oldest chart across all | - |
| `helm_s3_overall_age_newest_seconds` | Gauge | Unix timestamp of the newest chart across all | - |
| `helm_s3_overall_age_median_seconds` | Gauge | Unix timestamp of the median chart across all | - |
| `helm_s3_scrape_duration_seconds` | Histogram | Duration of the S3 scrape operation | - |
| `helm_s3_scrape_errors_total` | Counter | Total number of scrape errors | - |
| `helm_s3_last_scrape_success` | Gauge | Unix timestamp of the last successful scrape | - |

## Installation

### Prerequisites

- Kubernetes cluster (1.19+)
- Helm 3.x
- AWS S3 bucket with Helm repository
- AWS credentials or IAM role for S3 access

### Quick Start

1. **Add the Helm repository** (when published):

```bash
helm repo add helm-s3-exporter https://obezpalko.github.io/helm-s3-exporter
helm repo update
```

2. **Install the chart**:

```bash
helm install my-exporter helm-s3-exporter/helm-s3-exporter \
  --set s3.bucket=my-helm-repo-bucket \
  --set s3.region=us-east-1
```

### Local Installation

For local development or testing:

```bash
helm install my-exporter ./charts/helm-s3-exporter \
  --set s3.bucket=my-helm-repo-bucket \
  --set s3.region=us-east-1
```

## Configuration

### Basic Configuration

The minimum required configuration:

```yaml
s3:
  bucket: "my-helm-repo-bucket"
  region: "us-east-1"
```

### Authentication Options

#### Option 1: IAM Role (Recommended for EKS)

Using [IAM Roles for Service Accounts (IRSA)](https://docs.aws.amazon.com/eks/latest/userguide/iam-roles-for-service-accounts.html):

```yaml
auth:
  useIAMRole: true

serviceAccount:
  create: true
  roleArn: arn:aws:iam::123456789012:role/helm-s3-exporter-role
```

**Required IAM Policy**:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "s3:GetObject",
        "s3:ListBucket"
      ],
      "Resource": [
        "arn:aws:s3:::my-helm-repo-bucket",
        "arn:aws:s3:::my-helm-repo-bucket/*"
      ]
    }
  ]
}
```

#### Option 2: Existing Secret

```yaml
auth:
  useIAMRole: false
  existingSecret: "aws-credentials"
```

The secret should contain:
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: aws-credentials
type: Opaque
data:
  AWS_ACCESS_KEY_ID: <base64-encoded>
  AWS_SECRET_ACCESS_KEY: <base64-encoded>
```

#### Option 3: Static Credentials (NOT RECOMMENDED)

⚠️ **WARNING**: Only use for development/testing!

```yaml
auth:
  useIAMRole: false
  credentials:
    accessKeyId: "AKIAIOSFODNN7EXAMPLE"
    secretAccessKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
```

### Advanced Configuration

```yaml
# Scan configuration
scan:
  interval: "5m"     # How often to scan S3
  timeout: "30s"     # Timeout for S3 operations

# Enable HTML dashboard
html:
  enabled: true
  path: "/charts"

# Enable ServiceMonitor for Prometheus Operator
serviceMonitor:
  enabled: true
  interval: 30s
  additionalLabels:
    prometheus: kube-prometheus

# Resource limits
resources:
  limits:
    cpu: 200m
    memory: 256Mi
  requests:
    cpu: 100m
    memory: 128Mi
```

See [`charts/helm-s3-exporter/values.yaml`](charts/helm-s3-exporter/values.yaml) for all available options.

## Usage Examples

### Example 1: EKS with IRSA

```yaml
# values-eks.yaml
s3:
  bucket: "production-helm-charts"
  region: "us-west-2"

auth:
  useIAMRole: true

serviceAccount:
  create: true
  roleArn: arn:aws:iam::123456789012:role/helm-s3-exporter

serviceMonitor:
  enabled: true
  additionalLabels:
    prometheus: kube-prometheus

html:
  enabled: true
```

Install:
```bash
helm install helm-exporter ./charts/helm-s3-exporter -f values-eks.yaml
```

### Example 2: Development with LocalStack

```yaml
# values-localstack.yaml
s3:
  bucket: "test-bucket"
  region: "us-east-1"
  key: "index.yaml"

auth:
  useIAMRole: false
  credentials:
    accessKeyId: "test"
    secretAccessKey: "test"

scan:
  interval: "1m"

html:
  enabled: true

extraEnv:
  - name: AWS_ENDPOINT_URL
    value: "http://localstack:4566"
```

### Example 3: Multiple Buckets

Deploy multiple instances for different buckets:

```bash
helm install charts-prod ./charts/helm-s3-exporter \
  --set s3.bucket=prod-charts \
  --set nameOverride=prod

helm install charts-staging ./charts/helm-s3-exporter \
  --set s3.bucket=staging-charts \
  --set nameOverride=staging
```

## Development

### Building Locally

```bash
# Build the binary
make build

# Run locally (requires AWS credentials)
export S3_BUCKET=my-bucket
export S3_REGION=us-east-1
make run

# Run tests
make test
```

### Building Docker Image

```bash
# Build image
make docker-build

# Push image (use GitHub Container Registry)
make docker-push DOCKER_REGISTRY=ghcr.io/obezpalko
```

The official images are published to GitHub Container Registry:
```bash
docker pull ghcr.io/obezpalko/helm-s3-exporter:latest
```

### Project Structure

```
helm-s3-exporter/
├── cmd/
│   └── exporter/          # Main application entry point
├── internal/
│   ├── s3/                # S3 client implementation
│   ├── analyzer/          # Chart analysis logic
│   ├── metrics/           # Prometheus metrics
│   └── web/               # HTML dashboard
├── pkg/
│   └── config/            # Configuration management
├── charts/
│   └── helm-s3-exporter/  # Helm chart
├── Dockerfile             # Container image definition
└── Makefile              # Build automation
```

## Monitoring and Alerting

### Example Prometheus Rules

```yaml
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: helm-s3-exporter-alerts
spec:
  groups:
  - name: helm-s3-exporter
    interval: 30s
    rules:
    - alert: HelmS3ExporterDown
      expr: up{job="helm-s3-exporter"} == 0
      for: 5m
      labels:
        severity: critical
      annotations:
        summary: "Helm S3 Exporter is down"
    
    - alert: HelmS3ScrapeErrors
      expr: rate(helm_s3_scrape_errors_total[5m]) > 0
      for: 10m
      labels:
        severity: warning
      annotations:
        summary: "Helm S3 Exporter is experiencing scrape errors"
    
    - alert: StaleHelmCharts
      expr: (time() - helm_s3_chart_age_newest_seconds) > 86400 * 90
      for: 1h
      labels:
        severity: info
      annotations:
        summary: "Chart {{ $labels.chart }} hasn't been updated in 90 days"
```

### Example Grafana Dashboard

Query examples:

```promql
# Total charts
helm_s3_charts_total

# Chart versions
helm_s3_chart_versions

# Scrape duration
rate(helm_s3_scrape_duration_seconds_sum[5m]) / rate(helm_s3_scrape_duration_seconds_count[5m])

# Age of newest chart (in days)
(time() - helm_s3_chart_age_newest_seconds{chart="your-chart"}) / 86400
```

## Troubleshooting

### Exporter Can't Access S3

1. **Check IAM permissions**:
```bash
kubectl logs -l app.kubernetes.io/name=helm-s3-exporter
```

2. **Verify service account annotations** (for IRSA):
```bash
kubectl describe sa helm-s3-exporter
```

3. **Test S3 access** from pod:
```bash
kubectl exec -it deploy/helm-s3-exporter -- sh
# (won't work with distroless, use debug container)
```

### No Metrics Appearing

1. **Check if exporter is scraping**:
```bash
kubectl logs -l app.kubernetes.io/name=helm-s3-exporter
```

2. **Verify ServiceMonitor** (if using Prometheus Operator):
```bash
kubectl get servicemonitor
kubectl describe servicemonitor helm-s3-exporter
```

3. **Check Prometheus targets**:
Visit Prometheus UI → Status → Targets

### HTML Dashboard Not Working

Ensure HTML is enabled:
```bash
kubectl port-forward svc/helm-s3-exporter 9571:9571
curl http://localhost:9571/charts
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## References

- [Helm S3 Plugin](https://github.com/hypnoglow/helm-s3)
- [Prometheus Exporters](https://prometheus.io/docs/instrumenting/exporters/)
- [AWS IAM Roles for Service Accounts](https://docs.aws.amazon.com/eks/latest/userguide/iam-roles-for-service-accounts.html)

## Support

For issues, questions, or contributions, please visit:
- GitHub Issues: https://github.com/obezpalko/helm-s3-exporter/issues
- GitHub Discussions: https://github.com/obezpalko/helm-s3-exporter/discussions
