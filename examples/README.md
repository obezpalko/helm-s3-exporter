# Examples

This directory contains example configurations for common deployment scenarios.

## Quick Reference

| File | Description | Use Case |
|------|-------------|----------|
| `values-simple.yaml` | Minimal setup | Quick start with single public repository |
| `values-simple-url.yaml` | Single URL | Simple inline configuration |
| `values-multi-repos.yaml` | Multiple repos | Multiple repositories with config file |
| `values-static-credentials.yaml` | Development | Non-production with static credentials |
| `prometheus-rules.yaml` | Alerting | Prometheus rules and alerts |
| `config-single.yaml` | Single repo config | Example config file for one repository |
| `config-multi-auth.yaml` | Multi-auth config | Multiple repos with different auth methods |
| `config-multi-interval.yaml` | Per-repo intervals | Different scan intervals per repository |

## Using Examples

### Simple Setup (Recommended)

The easiest way to get started:

```bash
# Install with a single public repository
helm install helm-repo-exporter ../charts/helm-repo-exporter \
  -f values-simple.yaml \
  --namespace monitoring \
  --create-namespace
```

### Multiple Repositories

For multiple repositories or authentication:

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
EOF

# 2. Create secret
kubectl create secret generic helm-repo-config \
  --from-file=config.yaml=config.yaml \
  --namespace monitoring

# 3. Install
helm install helm-repo-exporter ../charts/helm-repo-exporter \
  -f values-multi-repos.yaml \
  --namespace monitoring
```

### Development Setup

For testing with static credentials:

```bash
# 1. Create config file with credentials
cat > config.yaml <<EOF
repositories:
  - name: private-repo
    url: https://charts.example.com/index.yaml
    auth:
      basic:
        username: dev-user
        password: dev-password
EOF

# 2. Create secret
kubectl create secret generic helm-repo-credentials \
  --from-file=config.yaml=config.yaml \
  --namespace monitoring

# 3. Install
helm install helm-repo-exporter ../charts/helm-repo-exporter \
  -f values-static-credentials.yaml \
  --namespace monitoring
```

### Prometheus Integration

Apply the example alerting rules:

```bash
kubectl apply -f prometheus-rules.yaml -n monitoring
```

## Customizing Examples

All examples can be customized by:

1. **Copying the file**:
```bash
cp values-simple.yaml my-values.yaml
```

2. **Editing your copy**:
```bash
vim my-values.yaml
```

3. **Installing with your values**:
```bash
helm install helm-repo-exporter ../charts/helm-repo-exporter -f my-values.yaml
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

## Validation

Test your values file before installing:

```bash
# Dry run
helm install test ../charts/helm-repo-exporter -f my-values.yaml --dry-run --debug

# Template rendering
helm template test ../charts/helm-repo-exporter -f my-values.yaml

# Lint
helm lint ../charts/helm-repo-exporter -f my-values.yaml
```

## Troubleshooting

If you encounter issues:

1. **Check the rendered template**:
```bash
helm template test ../charts/helm-repo-exporter -f my-values.yaml | less
```

2. **Check logs**:
```bash
kubectl logs -n monitoring -l app.kubernetes.io/name=helm-repo-exporter
```

3. **Verify configuration**:
```bash
kubectl get secret helm-repo-config -n monitoring -o yaml
```

## More Information

- See [QUICKSTART.md](../QUICKSTART.md) for step-by-step installation
- See [README.md](../README.md) for complete documentation
- See [CONFIGURATION_GUIDE.md](./CONFIGURATION_GUIDE.md) for detailed configuration options
- See [SECURITY.md](../SECURITY.md) for security best practices
