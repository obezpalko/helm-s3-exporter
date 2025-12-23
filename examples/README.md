# Examples

This directory contains example configurations for common deployment scenarios.

## Quick Reference

| File | Description | Use Case |
|------|-------------|----------|
| `values-simple.yaml` | Minimal setup | Quick start with IRSA |
| `values-eks-irsa.yaml` | Production EKS | AWS EKS with IRSA and monitoring |
| `values-static-credentials.yaml` | Development | Non-production with static credentials |
| `prometheus-rules.yaml` | Alerting | Prometheus rules and alerts |
| `aws-iam-policy.json` | IAM Policy | Required S3 permissions |
| `aws-trust-policy.json` | Trust Policy | IRSA trust relationship |
| `setup-irsa.sh` | Automation | Script to setup IRSA |

## Using Examples

### Simple Setup (Recommended)

The easiest way to get started with IRSA:

```bash
# 1. Edit values-simple.yaml with your bucket and role ARN
# 2. Install
helm install helm-s3-exporter ../charts/helm-s3-exporter \
  -f values-simple.yaml \
  --namespace monitoring \
  --create-namespace
```

### Production EKS Setup

For a full-featured production deployment:

```bash
helm install helm-s3-exporter ../charts/helm-s3-exporter \
  -f values-eks-irsa.yaml \
  --namespace monitoring \
  --create-namespace
```

### Development Setup

For testing with static credentials:

```bash
# Create secret first
kubectl create secret generic aws-credentials \
  --from-literal=AWS_ACCESS_KEY_ID=xxx \
  --from-literal=AWS_SECRET_ACCESS_KEY=yyy \
  --namespace monitoring

# Install
helm install helm-s3-exporter ../charts/helm-s3-exporter \
  -f values-static-credentials.yaml \
  --namespace monitoring
```

### Prometheus Integration

Apply the example alerting rules:

```bash
kubectl apply -f prometheus-rules.yaml -n monitoring
```

## IAM Role Setup

### Option 1: Using the setup script

```bash
./setup-irsa.sh my-cluster us-west-2 my-helm-bucket monitoring
```

### Option 2: Manual setup

1. Create IAM policy using `aws-iam-policy.json`:
```bash
aws iam create-policy \
  --policy-name helm-s3-exporter-policy \
  --policy-document file://aws-iam-policy.json
```

2. Create service account with IRSA:
```bash
eksctl create iamserviceaccount \
  --cluster=my-cluster \
  --namespace=monitoring \
  --name=helm-s3-exporter \
  --attach-policy-arn=arn:aws:iam::ACCOUNT:policy/helm-s3-exporter-policy \
  --approve
```

3. Install with the created service account:
```bash
helm install helm-s3-exporter ../charts/helm-s3-exporter \
  --set s3.bucket=my-bucket \
  --set serviceAccount.create=false \
  --set serviceAccount.name=helm-s3-exporter
```

### Option 3: Let Helm create the service account

Simply specify the role ARN in values:

```yaml
serviceAccount:
  create: true
  roleArn: "arn:aws:iam::123456789012:role/helm-s3-exporter-role"
```

The `eks.amazonaws.com/role-arn` annotation will be added automatically!

## Customizing Examples

All examples can be customized by:

1. **Copying the file**:
```bash
cp values-eks-irsa.yaml my-values.yaml
```

2. **Editing your copy**:
```bash
vim my-values.yaml
```

3. **Installing with your values**:
```bash
helm install helm-s3-exporter ../charts/helm-s3-exporter -f my-values.yaml
```

## Common Customizations

### Different Scan Interval

```yaml
scan:
  interval: "10m"  # Scan every 10 minutes
  timeout: "1m"    # 1 minute timeout
```

### Enable HTML Dashboard

```yaml
html:
  enabled: true
  path: "/charts"
```

### Resource Limits

```yaml
resources:
  limits:
    cpu: 200m
    memory: 256Mi
  requests:
    cpu: 100m
    memory: 128Mi
```

### Multiple Annotations

```yaml
serviceAccount:
  roleArn: "arn:aws:iam::123456789012:role/my-role"
  annotations:
    custom-annotation: "custom-value"
    another-annotation: "another-value"
```

The role ARN annotation is added automatically, and your custom annotations are merged in!

## Validation

Test your values file before installing:

```bash
# Dry run
helm install test ../charts/helm-s3-exporter -f my-values.yaml --dry-run --debug

# Template rendering
helm template test ../charts/helm-s3-exporter -f my-values.yaml

# Lint
helm lint ../charts/helm-s3-exporter -f my-values.yaml
```

## Troubleshooting

If you encounter issues:

1. **Check the rendered template**:
```bash
helm template test ../charts/helm-s3-exporter -f my-values.yaml | less
```

2. **Validate ServiceAccount annotation**:
```bash
kubectl get serviceaccount helm-s3-exporter -n monitoring -o yaml
```

3. **Check logs**:
```bash
kubectl logs -n monitoring -l app.kubernetes.io/name=helm-s3-exporter
```

## More Information

- See [QUICKSTART.md](../QUICKSTART.md) for step-by-step installation
- See [README.md](../README.md) for complete documentation
- See [SECURITY.md](../SECURITY.md) for security best practices

