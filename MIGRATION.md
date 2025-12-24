# Migration Guide: Multi-Repository Support

This guide helps you migrate from the single-URL configuration to the new multi-repository configuration with authentication support.

## What Changed?

### New Features

✅ **Multiple Repositories** - Monitor multiple chart repositories simultaneously  
✅ **Flexible Authentication** - Basic Auth, Bearer tokens, custom headers  
✅ **Config File Support** - YAML configuration via Secret or ConfigMap  
✅ **Backward Compatible** - Old single-URL method still works  

### Breaking Changes

⚠️ **Helm Values Structure Changed**

**Old (deprecated but still works):**
```yaml
repository:
  indexURL: "https://charts.example.com/index.yaml"
```

**New (recommended):**
```yaml
config:
  inline:
    enabled: true
    url: "https://charts.example.com/index.yaml"
```

## Migration Paths

### Path 1: Keep Using Single Repository (No Changes Needed)

If you're using environment variables directly, **no changes required**:

```bash
# Still works!
export INDEX_URL=https://charts.bitnami.com/bitnami/index.yaml
./exporter
```

### Path 2: Migrate to Inline Config (Minimal Changes)

**Before:**
```bash
helm install my-exporter ./charts/helm-s3-exporter \
  --set repository.indexURL=https://charts.bitnami.com/bitnami/index.yaml
```

**After:**
```bash
helm install my-exporter ./charts/helm-s3-exporter \
  --set config.inline.enabled=true \
  --set config.inline.url=https://charts.bitnami.com/bitnami/index.yaml
```

**Or update your values.yaml:**

```yaml
# Before
repository:
  indexURL: "https://charts.bitnami.com/bitnami/index.yaml"

# After
config:
  inline:
    enabled: true
    url: "https://charts.bitnami.com/bitnami/index.yaml"
```

### Path 3: Upgrade to Multi-Repository (Recommended)

**Step 1: Create config file**

```yaml
# config.yaml
repositories:
  - name: bitnami
    url: https://charts.bitnami.com/bitnami/index.yaml
  
  # Add more repositories
  - name: prometheus
    url: https://prometheus-community.github.io/helm-charts/index.yaml

scanInterval: 10m
scanTimeout: 1m
```

**Step 2: Create Secret**

```bash
kubectl create secret generic helm-repo-config \
  --from-file=config.yaml=config.yaml \
  --namespace monitoring
```

**Step 3: Upgrade Helm release**

```bash
helm upgrade my-exporter ./charts/helm-s3-exporter \
  --namespace monitoring \
  --set config.existingSecret.enabled=true \
  --set config.existingSecret.name=helm-repo-config \
  --reuse-values
```

## Migration Examples

### Example 1: Single Public Repository

**Before:**
```yaml
# values.yaml
repository:
  indexURL: "https://charts.bitnami.com/bitnami/index.yaml"

scan:
  interval: "5m"

serviceMonitor:
  enabled: true
```

**After (Option A - Inline):**
```yaml
# values.yaml
config:
  inline:
    enabled: true
    url: "https://charts.bitnami.com/bitnami/index.yaml"

scan:
  interval: "5m"

serviceMonitor:
  enabled: true
```

**After (Option B - Config File):**
```yaml
# config.yaml
repositories:
  - name: bitnami
    url: https://charts.bitnami.com/bitnami/index.yaml

scanInterval: 5m
```

```yaml
# values.yaml
config:
  existingConfigMap:
    enabled: true
    name: helm-repo-config

serviceMonitor:
  enabled: true
```

### Example 2: Adding Authentication

If your repository now requires authentication:

**Create config with auth:**
```yaml
# config.yaml
repositories:
  - name: private-charts
    url: https://charts.company.com/index.yaml
    auth:
      basic:
        username: helm-user
        password: secret-password

scanInterval: 10m
```

**Create Secret and upgrade:**
```bash
kubectl create secret generic helm-repo-config \
  --from-file=config.yaml=config.yaml \
  --namespace monitoring

helm upgrade my-exporter ./charts/helm-s3-exporter \
  --namespace monitoring \
  --set config.existingSecret.enabled=true \
  --set config.existingSecret.name=helm-repo-config
```

### Example 3: Multiple Repositories

**Expand from one to many:**

```yaml
# config.yaml
repositories:
  # Your existing repository
  - name: production
    url: https://charts.company.com/prod/index.yaml
    auth:
      basic:
        username: helm-user
        password: prod-password
  
  # Add staging
  - name: staging
    url: https://charts.company.com/staging/index.yaml
    auth:
      basic:
        username: helm-user
        password: staging-password
  
  # Add public repos for monitoring
  - name: bitnami
    url: https://charts.bitnami.com/bitnami/index.yaml

scanInterval: 10m
```

## Rollback Plan

If you need to rollback:

### Option 1: Use Previous Helm Release

```bash
helm rollback my-exporter -n monitoring
```

### Option 2: Redeploy with Old Configuration

```bash
helm upgrade my-exporter ./charts/helm-s3-exporter \
  --namespace monitoring \
  --set config.inline.enabled=true \
  --set config.inline.url=https://your-repo.com/index.yaml
```

## Verification After Migration

### 1. Check Pod Status

```bash
kubectl get pods -n monitoring -l app.kubernetes.io/name=helm-s3-exporter
```

### 2. Check Logs

```bash
kubectl logs -n monitoring -l app.kubernetes.io/name=helm-s3-exporter

# You should see:
# Configuration loaded:
#   Repositories: X
#     - name1: url1
#     - name2: url2
# Scrape completed in XXXms
```

### 3. Verify Metrics

```bash
kubectl port-forward -n monitoring svc/my-exporter 9571:9571
curl http://localhost:9571/metrics | grep helm_s3_charts_total
```

### 4. Check for Errors

```bash
kubectl logs -n monitoring -l app.kubernetes.io/name=helm-s3-exporter | grep -i error
```

## Troubleshooting Migration Issues

### Issue: "either CONFIG_FILE or INDEX_URL environment variable is required"

**Cause:** No configuration method is enabled

**Solution:**
```bash
# Enable one configuration method
helm upgrade my-exporter ./charts/helm-s3-exporter \
  --set config.inline.enabled=true \
  --set config.inline.url=https://charts.example.com/index.yaml
```

### Issue: "failed to read config file"

**Cause:** Secret/ConfigMap doesn't exist or isn't mounted

**Solution:**
```bash
# Verify secret exists
kubectl get secret helm-repo-config -n monitoring

# If missing, create it
kubectl create secret generic helm-repo-config \
  --from-file=config.yaml=config.yaml \
  --namespace monitoring

# Restart pods
kubectl rollout restart deployment -n monitoring my-exporter
```

### Issue: Authentication failures (401/403)

**Cause:** Invalid credentials

**Solution:**
```bash
# Test credentials manually
curl -u username:password https://charts.example.com/index.yaml

# Update secret
kubectl create secret generic helm-repo-config \
  --from-file=config.yaml=config.yaml \
  --namespace monitoring \
  --dry-run=client -o yaml | kubectl apply -f -

# Restart
kubectl rollout restart deployment -n monitoring my-exporter
```

### Issue: Metrics show zero charts

**Cause:** URL is incorrect or inaccessible

**Solution:**
```bash
# Check logs for specific error
kubectl logs -n monitoring -l app.kubernetes.io/name=helm-s3-exporter

# Test URL manually
curl -I https://your-repo-url/index.yaml

# Verify config
kubectl get secret helm-repo-config -n monitoring -o jsonpath='{.data.config\.yaml}' | base64 -d
```

## FAQ

### Q: Can I use both inline and config file?

**A:** No, only one method can be enabled at a time. The precedence is:
1. `config.existingSecret` (highest priority)
2. `config.existingConfigMap`
3. `config.inline`

### Q: Do I need to migrate immediately?

**A:** No, the old `INDEX_URL` environment variable still works for backward compatibility. However, the new config file method is recommended for better flexibility.

### Q: Can I mix public and private repositories?

**A:** Yes! Just omit the `auth` section for public repositories:

```yaml
repositories:
  - name: public
    url: https://charts.bitnami.com/bitnami/index.yaml
  
  - name: private
    url: https://charts.company.com/index.yaml
    auth:
      basic:
        username: user
        password: pass
```

### Q: How do I rotate credentials?

**A:** Update the secret and restart:

```bash
# Update config.yaml with new credentials
kubectl create secret generic helm-repo-config \
  --from-file=config.yaml=config.yaml \
  --namespace monitoring \
  --dry-run=client -o yaml | kubectl apply -f -

# Restart to pick up changes
kubectl rollout restart deployment -n monitoring my-exporter
```

### Q: Can I use environment variables in config file?

**A:** Yes, you can use `${VAR_NAME}` syntax in the config file and pass environment variables via Helm values:

```yaml
# config.yaml
repositories:
  - name: private
    url: ${REPO_URL}
    auth:
      basic:
        username: ${REPO_USER}
        password: ${REPO_PASS}
```

```yaml
# values.yaml
extraEnv:
  - name: REPO_URL
    value: "https://charts.example.com/index.yaml"
  - name: REPO_USER
    valueFrom:
      secretKeyRef:
        name: repo-creds
        key: username
  - name: REPO_PASS
    valueFrom:
      secretKeyRef:
        name: repo-creds
        key: password
```

## Support

If you encounter issues during migration:

1. Check the [Troubleshooting](#troubleshooting-migration-issues) section above
2. Review [QUICKSTART.md](QUICKSTART.md) for working examples
3. See [examples/CONFIGURATION_GUIDE.md](examples/CONFIGURATION_GUIDE.md) for detailed config options
4. Open an issue on GitHub with logs and configuration

## Summary

- ✅ Backward compatible - old method still works
- ✅ New config file method is more flexible
- ✅ Migration is straightforward
- ✅ Can be done gradually
- ✅ Easy rollback if needed

Happy monitoring! 🎉

