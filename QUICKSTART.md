# Quick Start Guide

Get up and running with Helm Repository Exporter in minutes!

## Installation Options

### Option 1: Single Public Repository (Simplest)

Perfect for monitoring a single public Helm repository.

```bash
helm install my-exporter ./charts/helm-s3-exporter \
  --set config.inline.enabled=true \
  --set config.inline.url=https://charts.bitnami.com/bitnami/index.yaml \
  --namespace monitoring \
  --create-namespace
```

**Verify it's working:**

```bash
# Check pod status
kubectl get pods -n monitoring

# View logs
kubectl logs -n monitoring -l app.kubernetes.io/name=helm-s3-exporter

# Access metrics
kubectl port-forward -n monitoring svc/my-exporter 9571:9571
curl http://localhost:9571/metrics | grep helm_s3
```

---

### Option 2: Multiple Public Repositories

Monitor multiple public repositories.

**Step 1: Create config file**

```bash
cat > config.yaml <<EOF
repositories:
  - name: bitnami
    url: https://charts.bitnami.com/bitnami/index.yaml
  
  - name: prometheus-community
    url: https://prometheus-community.github.io/helm-charts/index.yaml
  
  - name: jetstack
    url: https://charts.jetstack.io/index.yaml

scanInterval: 10m  # Default interval for all repos
scanTimeout: 1m
enableHTML: true
EOF
```

**Pro Tip:** You can set different scan intervals per repository:
```yaml
repositories:
  - name: critical-prod
    url: https://charts.company.com/prod/index.yaml
    scanInterval: 5m  # Check every 5 minutes
  
  - name: bitnami
    url: https://charts.bitnami.com/bitnami/index.yaml
    scanInterval: 30m  # Check every 30 minutes
```

**Step 2: Create ConfigMap**

```bash
kubectl create configmap helm-repo-config \
  --from-file=config.yaml=config.yaml \
  --namespace monitoring \
  --dry-run=client -o yaml | kubectl apply -f -
```

**Step 3: Deploy**

```bash
helm install my-exporter ./charts/helm-s3-exporter \
  --namespace monitoring \
  --create-namespace \
  --set config.existingConfigMap.enabled=true \
  --set config.existingConfigMap.name=helm-repo-config \
  --set html.enabled=true
```

**Access the dashboard:**

```bash
kubectl port-forward -n monitoring svc/my-exporter 9571:9571
# Open http://localhost:9571/charts in your browser
```

---

### Option 3: Private Repositories with Authentication

Monitor private repositories that require authentication.

**Step 1: Create config file with credentials**

```bash
cat > config.yaml <<EOF
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
        password: your-password-here
  
  # Private with Bearer Token (e.g., GitHub)
  - name: github-private
    url: https://raw.githubusercontent.com/yourorg/charts/main/index.yaml
    auth:
      bearerToken: ghp_your_github_token_here

scanInterval: 10m
enableHTML: true
EOF
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
  --create-namespace \
  --set config.existingSecret.enabled=true \
  --set config.existingSecret.name=helm-repo-config \
  --set html.enabled=true
```

---

### Option 4: With Prometheus Operator

Enable ServiceMonitor for automatic Prometheus scraping.

```bash
helm install my-exporter ./charts/helm-s3-exporter \
  --namespace monitoring \
  --create-namespace \
  --set config.inline.enabled=true \
  --set config.inline.url=https://charts.bitnami.com/bitnami/index.yaml \
  --set serviceMonitor.enabled=true \
  --set serviceMonitor.additionalLabels.prometheus=kube-prometheus
```

**Verify in Prometheus:**

```bash
# Port forward to Prometheus
kubectl port-forward -n monitoring svc/prometheus-operated 9090:9090

# Open http://localhost:9090 and search for: helm_repo_charts_total
```

---

## Common Use Cases

### S3 Bucket (Public)

```bash
helm install s3-exporter ./charts/helm-s3-exporter \
  --namespace monitoring \
  --set config.inline.enabled=true \
  --set config.inline.url=https://my-bucket.s3.us-west-2.amazonaws.com/index.yaml
```

### GitHub Repository (Private)

```yaml
# config.yaml
repositories:
  - name: my-charts
    url: https://raw.githubusercontent.com/myorg/helm-charts/main/index.yaml
    auth:
      bearerToken: ghp_xxxxxxxxxxxx
```

### Artifactory

```yaml
# config.yaml
repositories:
  - name: artifactory
    url: https://artifactory.company.com/artifactory/helm-local/index.yaml
    auth:
      basic:
        username: admin
        password: password
```

### Harbor

```yaml
# config.yaml
repositories:
  - name: harbor
    url: https://harbor.company.com/chartrepo/library/index.yaml
    auth:
      basic:
        username: admin
        password: Harbor12345
```

---

## Verification Checklist

After installation, verify everything is working:

- [ ] **Pod is running**
  ```bash
  kubectl get pods -n monitoring -l app.kubernetes.io/name=helm-s3-exporter
  ```

- [ ] **Logs show successful scraping**
  ```bash
  kubectl logs -n monitoring -l app.kubernetes.io/name=helm-s3-exporter
  # Should see: "Scrape completed" messages
  ```

- [ ] **Metrics are exposed**
  ```bash
  kubectl port-forward -n monitoring svc/my-exporter 9571:9571
  curl http://localhost:9571/metrics | grep helm_repo_charts_total
  ```

- [ ] **ServiceMonitor is created** (if enabled)
  ```bash
  kubectl get servicemonitor -n monitoring
  ```

- [ ] **Prometheus is scraping** (if using Prometheus Operator)
  - Check Prometheus UI → Status → Targets
  - Look for `helm-s3-exporter` target

---

## Troubleshooting

### Pod is CrashLooping

```bash
# Check logs
kubectl logs -n monitoring -l app.kubernetes.io/name=helm-s3-exporter

# Common issues:
# - Missing config: Enable one of config.inline, config.existingSecret, or config.existingConfigMap
# - Invalid URL: Test URL manually with curl
# - Auth failure: Verify credentials
```

### No Metrics Showing

```bash
# Verify service is accessible
kubectl get svc -n monitoring my-exporter

# Port forward and test
kubectl port-forward -n monitoring svc/my-exporter 9571:9571
curl http://localhost:9571/metrics

# Check for errors in logs
kubectl logs -n monitoring -l app.kubernetes.io/name=helm-s3-exporter | grep -i error
```

### Authentication Failures

```bash
# Test URL manually
curl -I https://charts.example.com/index.yaml

# Test with auth
curl -u username:password https://charts.example.com/index.yaml
curl -H "Authorization: Bearer token" https://charts.example.com/index.yaml

# Verify secret exists
kubectl get secret helm-repo-config -n monitoring -o yaml
```

---

## Next Steps

- 📖 Read the [Configuration Guide](examples/CONFIGURATION_GUIDE.md) for advanced options
- 🔍 Set up [Prometheus alerts](examples/prometheus-rules.yaml)
- 📊 Explore the [HTML dashboard](http://localhost:9571/charts) (if enabled)
- 🎯 Check out more [examples](examples/)

---

## Quick Reference

### Useful Commands

```bash
# View logs
kubectl logs -n monitoring -l app.kubernetes.io/name=helm-s3-exporter -f

# Restart deployment
kubectl rollout restart deployment -n monitoring my-exporter

# Update config
kubectl create secret generic helm-repo-config \
  --from-file=config.yaml=config.yaml \
  --namespace monitoring \
  --dry-run=client -o yaml | kubectl apply -f -
kubectl rollout restart deployment -n monitoring my-exporter

# Uninstall
helm uninstall my-exporter -n monitoring
```

### Example Metrics Queries

```promql
# Total charts
helm_repo_charts_total

# Charts with many versions
helm_repo_chart_version_count > 10

# Scrape errors
rate(helm_repo_scrape_errors_total[5m])

# Scrape duration
helm_repo_scrape_duration_seconds
```

---

For more details, see the [full README](README.md).
