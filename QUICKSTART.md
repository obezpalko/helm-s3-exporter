# Quick Start Guide

This guide will help you get Helm S3 Exporter up and running in under 5 minutes.

## Prerequisites

- Kubernetes cluster (1.19+)
- Helm 3.x installed
- S3 bucket with Helm charts
- AWS credentials or IAM role

## Installation Steps

### Step 1: Choose Your Authentication Method

#### Option A: AWS EKS with IRSA (Recommended)

If you're on AWS EKS, set up IRSA first:

```bash
# Run the setup script
./examples/setup-irsa.sh my-cluster us-west-2 my-helm-bucket monitoring
```

Or create manually:
```bash
# Create IAM policy
aws iam create-policy \
  --policy-name helm-s3-exporter-policy \
  --policy-document file://examples/aws-iam-policy.json

# Create service account with IRSA
eksctl create iamserviceaccount \
  --cluster=my-cluster \
  --namespace=monitoring \
  --name=helm-s3-exporter \
  --attach-policy-arn=arn:aws:iam::ACCOUNT:policy/helm-s3-exporter-policy \
  --approve
```

#### Option B: Static Credentials (Development Only)

Create a secret with your AWS credentials:

```bash
kubectl create secret generic aws-credentials \
  --from-literal=AWS_ACCESS_KEY_ID=your-key \
  --from-literal=AWS_SECRET_ACCESS_KEY=your-secret \
  --namespace=monitoring
```

### Step 2: Install the Helm Chart

#### Using IRSA:

```bash
helm install helm-s3-exporter ./charts/helm-s3-exporter \
  --namespace monitoring \
  --create-namespace \
  --set s3.bucket=my-helm-charts \
  --set s3.region=us-west-2 \
  --set serviceAccount.roleArn=arn:aws:iam::ACCOUNT_ID:role/helm-s3-exporter-role
```

#### Using Static Credentials:

```bash
helm install helm-s3-exporter ./charts/helm-s3-exporter \
  --namespace monitoring \
  --create-namespace \
  --set s3.bucket=my-helm-charts \
  --set s3.region=us-west-2 \
  --set auth.useIAMRole=false \
  --set auth.existingSecret=aws-credentials
```

### Step 3: Verify Installation

```bash
# Check if the pod is running
kubectl get pods -n monitoring

# View logs
kubectl logs -n monitoring -l app.kubernetes.io/name=helm-s3-exporter

# Port-forward to access metrics
kubectl port-forward -n monitoring svc/helm-s3-exporter 9571:9571
```

### Step 4: Access Metrics

Open your browser or use curl:

```bash
# View metrics
curl http://localhost:9571/metrics

# View HTML dashboard (if enabled)
curl http://localhost:9571/charts
```

## Example Metrics Output

You should see metrics like:

```prometheus
# HELP helm_s3_charts_total Total number of distinct Helm charts
# TYPE helm_s3_charts_total gauge
helm_s3_charts_total 15

# HELP helm_s3_chart_versions Number of versions for each chart
# TYPE helm_s3_chart_versions gauge
helm_s3_chart_versions{chart="nginx"} 5
helm_s3_chart_versions{chart="postgresql"} 3

# HELP helm_s3_scrape_duration_seconds Duration of the S3 scrape operation
# TYPE helm_s3_scrape_duration_seconds histogram
helm_s3_scrape_duration_seconds_bucket{le="0.005"} 0
helm_s3_scrape_duration_seconds_bucket{le="0.01"} 0
helm_s3_scrape_duration_seconds_bucket{le="0.025"} 1
```

## Optional: Enable HTML Dashboard

Update your installation with:

```bash
helm upgrade helm-s3-exporter ./charts/helm-s3-exporter \
  --namespace monitoring \
  --reuse-values \
  --set html.enabled=true
```

Then access via port-forward:

```bash
kubectl port-forward -n monitoring svc/helm-s3-exporter 9571:9571
# Open http://localhost:9571/charts in your browser
```

## Optional: Enable Prometheus Integration

If you're using Prometheus Operator:

```bash
helm upgrade helm-s3-exporter ./charts/helm-s3-exporter \
  --namespace monitoring \
  --reuse-values \
  --set serviceMonitor.enabled=true \
  --set serviceMonitor.additionalLabels.prometheus=kube-prometheus
```

Or apply the example Prometheus rules:

```bash
kubectl apply -f examples/prometheus-rules.yaml
```

## Troubleshooting

### Pod Not Starting

Check logs for errors:
```bash
kubectl logs -n monitoring -l app.kubernetes.io/name=helm-s3-exporter
```

Common issues:
- **AWS credentials error**: Check service account annotations or secret
- **S3 bucket not accessible**: Verify IAM permissions and bucket name
- **Region mismatch**: Ensure correct AWS region is specified

### No Metrics Appearing

1. Check if the exporter is scraping:
```bash
kubectl logs -n monitoring -l app.kubernetes.io/name=helm-s3-exporter | grep "Scrape completed"
```

2. Verify the endpoint is accessible:
```bash
kubectl port-forward -n monitoring svc/helm-s3-exporter 9571:9571
curl http://localhost:9571/metrics
```

### ServiceMonitor Not Working

Verify ServiceMonitor is created:
```bash
kubectl get servicemonitor -n monitoring
kubectl describe servicemonitor helm-s3-exporter -n monitoring
```

Check if your Prometheus is configured to discover ServiceMonitors in this namespace.

## Next Steps

- Configure alerting rules (see `examples/prometheus-rules.yaml`)
- Adjust scan interval for your needs
- Set up resource limits for production
- Review security best practices in `SECURITY.md`
- Create Grafana dashboards for visualization

## Uninstallation

To remove the exporter:

```bash
helm uninstall helm-s3-exporter -n monitoring
```

To also remove the namespace:

```bash
kubectl delete namespace monitoring
```

## Getting Help

- Check the full [README.md](README.md) for detailed documentation
- Review [examples/](examples/) for configuration templates
- Open an issue on GitHub for bugs or feature requests

