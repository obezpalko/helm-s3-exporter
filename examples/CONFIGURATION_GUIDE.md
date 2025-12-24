# Configuration Guide

The Helm S3 Exporter supports multiple repositories with flexible authentication. This guide shows all configuration options.

## Table of Contents

1. [Simple Single Repository](#simple-single-repository)
2. [Multiple Repositories](#multiple-repositories)
3. [Authentication Methods](#authentication-methods)
4. [Kubernetes Deployment](#kubernetes-deployment)
5. [Config File Examples](#config-file-examples)

---

## Simple Single Repository

For a single public repository, use inline configuration:

```bash
helm install my-exporter ./charts/helm-s3-exporter \
  --set config.inline.enabled=true \
  --set config.inline.url=https://charts.bitnami.com/bitnami/index.yaml
```

Or with a values file (`values-simple-url.yaml`):

```yaml
config:
  inline:
    enabled: true
    url: "https://charts.bitnami.com/bitnami/index.yaml"
```

---

## Multiple Repositories

For multiple repositories or when authentication is needed, use a config file.

### Step 1: Create Config File

Create `config.yaml`:

```yaml
repositories:
  - name: bitnami
    url: https://charts.bitnami.com/bitnami/index.yaml
  
  - name: jetstack
    url: https://charts.jetstack.io/index.yaml

scanInterval: 10m
scanTimeout: 1m
```

### Step 2: Create Secret

```bash
kubectl create secret generic helm-repo-config \
  --from-file=config.yaml=config.yaml \
  --namespace monitoring
```

### Step 3: Deploy

```bash
helm install my-exporter ./charts/helm-s3-exporter \
  --namespace monitoring \
  --set config.existingSecret.enabled=true \
  --set config.existingSecret.name=helm-repo-config
```

---

## Authentication Methods

### Basic Authentication

```yaml
repositories:
  - name: private-repo
    url: https://charts.company.com/index.yaml
    auth:
      basic:
        username: myuser
        password: mypassword
```

### Bearer Token

```yaml
repositories:
  - name: github-private
    url: https://raw.githubusercontent.com/company/charts/main/index.yaml
    auth:
      bearerToken: ghp_xxxxxxxxxxxxxxxxxxxx
```

### Custom Headers

```yaml
repositories:
  - name: api-key-auth
    url: https://charts.example.com/index.yaml
    auth:
      headers:
        X-API-Key: your-api-key
        X-Custom-Header: custom-value
```

### Combined Authentication

You can combine multiple auth methods:

```yaml
repositories:
  - name: multi-auth
    url: https://charts.example.com/index.yaml
    auth:
      bearerToken: token123
      headers:
        X-Request-ID: unique-id
```

---

## Kubernetes Deployment

### Option 1: ConfigMap (Public Repos Only)

For public repositories without authentication:

```bash
# Create ConfigMap
kubectl create configmap helm-repo-config \
  --from-file=config.yaml=config.yaml \
  --namespace monitoring

# Deploy
helm install my-exporter ./charts/helm-s3-exporter \
  --namespace monitoring \
  --set config.existingConfigMap.enabled=true \
  --set config.existingConfigMap.name=helm-repo-config
```

### Option 2: Secret (Recommended for Private Repos)

For repositories with authentication:

```bash
# Create Secret
kubectl create secret generic helm-repo-config \
  --from-file=config.yaml=config.yaml \
  --namespace monitoring

# Deploy
helm install my-exporter ./charts/helm-s3-exporter \
  --namespace monitoring \
  --set config.existingSecret.enabled=true \
  --set config.existingSecret.name=helm-repo-config
```

### Option 3: Sealed Secrets

For GitOps workflows:

```bash
# Create sealed secret
kubectl create secret generic helm-repo-config \
  --from-file=config.yaml=config.yaml \
  --dry-run=client -o yaml | \
  kubeseal -o yaml > sealed-secret.yaml

# Apply sealed secret
kubectl apply -f sealed-secret.yaml

# Deploy
helm install my-exporter ./charts/helm-s3-exporter \
  --namespace monitoring \
  --set config.existingSecret.enabled=true \
  --set config.existingSecret.name=helm-repo-config
```

---

## Config File Examples

### Example 1: Public Repositories

```yaml
# config-public.yaml
repositories:
  - name: bitnami
    url: https://charts.bitnami.com/bitnami/index.yaml
  
  - name: prometheus-community
    url: https://prometheus-community.github.io/helm-charts/index.yaml
  
  - name: ingress-nginx
    url: https://kubernetes.github.io/ingress-nginx/index.yaml

scanInterval: 5m
scanTimeout: 30s
metricsPort: "9571"
enableHTML: true
```

### Example 1b: Per-Repository Scan Intervals

```yaml
# config-per-repo-intervals.yaml
repositories:
  # Critical production repo - check frequently
  - name: production-charts
    url: https://charts.company.com/prod/index.yaml
    scanInterval: 5m  # Override: scan every 5 minutes
  
  # Public bitnami - less frequent checks
  - name: bitnami
    url: https://charts.bitnami.com/bitnami/index.yaml
    scanInterval: 30m  # Override: scan every 30 minutes
  
  # This repo uses the default interval (10m)
  - name: staging
    url: https://charts.company.com/staging/index.yaml

# Default interval for repos without specific scanInterval
scanInterval: 10m
scanTimeout: 30s
metricsPort: "9571"
enableHTML: true
```

### Example 2: Mixed Public and Private

```yaml
# config-mixed.yaml
repositories:
  # Public
  - name: bitnami
    url: https://charts.bitnami.com/bitnami/index.yaml
  
  # Private with basic auth
  - name: company-internal
    url: https://charts.company.com/internal/index.yaml
    auth:
      basic:
        username: helm-user
        password: ${HELM_PASSWORD}  # Can use env vars
  
  # Private with token
  - name: github-enterprise
    url: https://github.company.com/api/v3/repos/helm/charts/contents/index.yaml
    auth:
      bearerToken: ${GITHUB_TOKEN}

scanInterval: 10m
```

### Example 3: S3 Buckets

```yaml
# config-s3.yaml
repositories:
  # Public S3 bucket
  - name: s3-public
    url: https://my-charts.s3.us-west-2.amazonaws.com/index.yaml
  
  # S3 with presigned URL (generate URL with auth, then use it)
  - name: s3-private
    url: https://my-private-charts.s3.us-west-2.amazonaws.com/index.yaml?X-Amz-Security-Token=...

scanInterval: 15m
```

### Example 4: Enterprise Setup

```yaml
# config-enterprise.yaml
repositories:
  # Production
  - name: prod-charts
    url: https://artifactory.company.com/helm-prod/index.yaml
    auth:
      basic:
        username: ${ARTIFACTORY_USER}
        password: ${ARTIFACTORY_PASSWORD}
  
  # Staging  
  - name: staging-charts
    url: https://artifactory.company.com/helm-staging/index.yaml
    auth:
      basic:
        username: ${ARTIFACTORY_USER}
        password: ${ARTIFACTORY_PASSWORD}
  
  # Development
  - name: dev-charts
    url: https://harbor.dev.company.com/chartrepo/library/index.yaml
    auth:
      headers:
        Authorization: "Basic ${HARBOR_AUTH}"

scanInterval: 5m
scanTimeout: 2m
metricsPort: "9571"
enableHTML: true
```

---

## Environment Variable Substitution

The config file supports environment variable substitution:

```yaml
repositories:
  - name: private
    url: ${REPO_URL}
    auth:
      basic:
        username: ${REPO_USERNAME}
        password: ${REPO_PASSWORD}
```

Pass environment variables via Helm:

```yaml
# values.yaml
extraEnv:
  - name: REPO_URL
    value: "https://charts.example.com/index.yaml"
  - name: REPO_USERNAME
    valueFrom:
      secretKeyRef:
        name: repo-creds
        key: username
  - name: REPO_PASSWORD
    valueFrom:
      secretKeyRef:
        name: repo-creds
        key: password
```

---

## Verification

After deployment, verify the configuration:

```bash
# Check logs
kubectl logs -n monitoring -l app.kubernetes.io/name=helm-s3-exporter

# You should see:
# Configuration loaded:
#   Repositories: 3
#     - bitnami: https://charts.bitnami.com/bitnami/index.yaml
#     - company: https://charts.company.com/index.yaml
#       Auth: Basic (username: helm-user)
#     - github: https://raw.githubusercontent.com/.../index.yaml
#       Auth: Bearer Token
```

Check metrics:

```bash
kubectl port-forward -n monitoring svc/my-exporter 9571:9571
curl http://localhost:9571/metrics | grep helm_s3_charts_total
```

---

## Troubleshooting

### Issue: Authentication Fails

**Symptom**: Logs show `403 Forbidden` or `401 Unauthorized`

**Solution**: 
- Verify credentials are correct
- Check if token has expired
- Ensure headers are properly formatted

### Issue: Repository Not Found

**Symptom**: Logs show `404 Not Found`

**Solution**:
- Verify URL is correct and accessible
- Test URL manually: `curl -I <url>`
- Check if repository requires authentication

### Issue: Config Not Loaded

**Symptom**: Logs show `failed to read config file`

**Solution**:
- Verify Secret/ConfigMap exists: `kubectl get secret helm-repo-config`
- Check volume mount: `kubectl describe pod <pod-name>`
- Verify config file syntax: `yamllint config.yaml`

---

## Best Practices

1. **Use Secrets for credentials** - Never put passwords in ConfigMaps
2. **Use environment variables** - For sensitive values, use env var substitution
3. **Separate configs per environment** - Different configs for dev/staging/prod
4. **Monitor scrape errors** - Set up alerts on `helm_s3_scrape_errors_total`
5. **Use appropriate scan intervals** - Longer intervals for stable repos
6. **Test URLs manually first** - Verify access before deploying

---

## References

- [Main README](../README.md)
- [Quick Start Guide](../QUICKSTART.md)
- [Config File Examples](.)

