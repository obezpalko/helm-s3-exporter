# Feature: Per-Repository Scan Intervals

## Overview

Each repository can now have its own dedicated scan interval. If not defined, the repository uses the global default `scanInterval`.

## Use Cases

This feature is useful when:

1. **Different Update Frequencies** - Some repositories update more frequently than others
2. **Load Management** - Reduce scraping frequency for large or slow repositories
3. **SLA Compliance** - Critical repos need frequent checks, legacy repos can be checked less often
4. **Cost Optimization** - Minimize API calls to rate-limited or metered endpoints

## Configuration

### Basic Example

```yaml
repositories:
  # Production repo - check every 5 minutes
  - name: production
    url: https://charts.company.com/prod/index.yaml
    scanInterval: 5m
  
  # Staging - check every 15 minutes
  - name: staging
    url: https://charts.company.com/staging/index.yaml
    scanInterval: 15m
  
  # This repo uses the default interval (10m)
  - name: bitnami
    url: https://charts.bitnami.com/bitnami/index.yaml

# Default interval for repos without specific scanInterval
scanInterval: 10m
```

### Advanced Example

```yaml
repositories:
  # Critical production - very frequent checks
  - name: critical-prod
    url: https://charts.company.com/critical/index.yaml
    scanInterval: 2m
    auth:
      basic:
        username: user
        password: pass
  
  # Standard production - frequent checks
  - name: production
    url: https://charts.company.com/prod/index.yaml
    scanInterval: 5m
    auth:
      basic:
        username: user
        password: pass
  
  # Staging - moderate checks
  - name: staging
    url: https://charts.company.com/staging/index.yaml
    scanInterval: 15m
    auth:
      basic:
        username: user
        password: pass
  
  # Public repos - less frequent
  - name: bitnami
    url: https://charts.bitnami.com/bitnami/index.yaml
    scanInterval: 30m
  
  - name: prometheus
    url: https://prometheus-community.github.io/helm-charts/index.yaml
    scanInterval: 30m
  
  # Legacy repo - infrequent checks
  - name: legacy
    url: https://charts.company.com/legacy/index.yaml
    scanInterval: 2h
    auth:
      basic:
        username: user
        password: pass

# Default interval (not used in this example as all repos specify their own)
scanInterval: 10m
scanTimeout: 1m
```

## How It Works

### Implementation Details

1. **Independent Goroutines**: Each repository runs in its own goroutine with its own ticker
2. **Concurrent Scraping**: Different repositories can be scraped at different times
3. **Fallback Logic**: If a repository doesn't specify `scanInterval`, it uses the global default
4. **Configuration Loading**: Intervals are set during config parsing and applied per-repo

### Execution Flow

```
Time    | Repo 1 (5m) | Repo 2 (10m) | Repo 3 (30m)
--------|-------------|--------------|-------------
00:00   | Scrape      | Scrape       | Scrape
00:05   | Scrape      |              |
00:10   | Scrape      | Scrape       |
00:15   | Scrape      |              |
00:20   | Scrape      | Scrape       |
00:25   | Scrape      |              |
00:30   | Scrape      | Scrape       | Scrape
```

## Example Use Cases

### Use Case 1: Environment-Based Intervals

```yaml
repositories:
  - name: production
    url: https://charts.company.com/prod/index.yaml
    scanInterval: 5m   # Check production frequently
  
  - name: staging
    url: https://charts.company.com/staging/index.yaml
    scanInterval: 10m  # Staging less frequently
  
  - name: development
    url: https://charts.company.com/dev/index.yaml
    scanInterval: 30m  # Dev even less frequently

scanInterval: 15m  # Default
```

### Use Case 2: Repository Size Optimization

```yaml
repositories:
  # Small repo - check frequently
  - name: microservices
    url: https://charts.company.com/microservices/index.yaml
    scanInterval: 5m
  
  # Large repo - check less frequently
  - name: platform
    url: https://charts.company.com/platform/index.yaml
    scanInterval: 30m
  
  # Huge public repo - check infrequently
  - name: bitnami
    url: https://charts.bitnami.com/bitnami/index.yaml
    scanInterval: 1h

scanInterval: 15m
```

### Use Case 3: Rate Limit Compliance

```yaml
repositories:
  # API with strict rate limits - space out checks
  - name: github-org
    url: https://raw.githubusercontent.com/company/charts/main/index.yaml
    scanInterval: 15m
    auth:
      bearerToken: token
  
  # Self-hosted - can check frequently
  - name: internal-artifactory
    url: https://artifactory.company.com/helm/index.yaml
    scanInterval: 5m
    auth:
      basic:
        username: user
        password: pass

scanInterval: 10m
```

## Benefits

### Performance

- **Reduced Load**: Don't over-scrape stable repositories
- **Better Resource Usage**: Spread scraping operations over time
- **Faster Critical Updates**: Priority repos get checked more often

### Operational

- **Flexibility**: Each repo can have its own schedule
- **Cost Management**: Reduce API calls to metered services
- **SLA Compliance**: Meet different monitoring requirements per repo

### Monitoring

- **Granular Control**: Fine-tune monitoring per repository
- **Efficient Alerting**: Critical repos trigger faster alerts
- **Load Distribution**: Avoid scraping everything at once

## Logging

The exporter logs show per-repository intervals:

```
Configuration loaded:
  Repositories: 3
    - production: https://charts.company.com/prod/index.yaml (interval: 5m0s)
    - staging: https://charts.company.com/staging/index.yaml (interval: 15m0s)
    - bitnami: https://charts.bitnami.com/bitnami/index.yaml (interval: 30m0s)
  Default Scan Interval: 10m0s

Started scraper for production with interval 5m0s
Started scraper for staging with interval 15m0s
Started scraper for bitnami with interval 30m0s
Exporter started successfully

Scraping repository: production
Repository production scraped in 234ms: 45 charts, 892 versions

Scraping repository: staging
Repository staging scraped in 156ms: 23 charts, 445 versions
```

## Configuration Options

### Duration Format

Intervals support standard Go duration format:

- `30s` - 30 seconds
- `5m` - 5 minutes
- `1h` - 1 hour
- `2h30m` - 2 hours 30 minutes

### Minimum Recommended Intervals

- **Critical Repos**: 2m - 5m
- **Standard Repos**: 5m - 15m
- **Low-Priority Repos**: 30m - 2h
- **Archive Repos**: 6h - 24h

### Performance Considerations

- Very short intervals (< 1m) may impact system resources
- Consider repository size and network latency
- Monitor scrape duration metrics
- Adjust based on actual update frequencies

## Testing

### Test Configuration

Create a test config with varying intervals:

```yaml
repositories:
  - name: fast
    url: https://charts.bitnami.com/bitnami/index.yaml
    scanInterval: 1m
  
  - name: medium
    url: https://charts.jetstack.io/index.yaml
    scanInterval: 5m
  
  - name: slow
    url: https://prometheus-community.github.io/helm-charts/index.yaml
    scanInterval: 10m

scanInterval: 5m
scanTimeout: 30s
metricsPort: "9571"
```

### Run Test

```bash
CONFIG_FILE=test-intervals.yaml ./bin/exporter
```

### Expected Output

```
Configuration loaded:
  Repositories: 3
    - fast: ... (interval: 1m0s)
    - medium: ... (interval: 5m0s)
    - slow: ... (interval: 10m0s)
  Default Scan Interval: 5m0s

Started scraper for fast with interval 1m0s
Started scraper for medium with interval 5m0s
Started scraper for slow with interval 10m0s
Exporter started successfully

# After 1 minute
Scraping repository: fast

# After 5 minutes
Scraping repository: medium

# After 10 minutes
Scraping repository: slow
```

## Migration from Global Interval

### Before (Global Only)

```yaml
repositories:
  - name: prod
    url: https://charts.company.com/prod/index.yaml
  - name: staging
    url: https://charts.company.com/staging/index.yaml
  - name: dev
    url: https://charts.company.com/dev/index.yaml

scanInterval: 10m  # All repos use this
```

### After (Per-Repo Intervals)

```yaml
repositories:
  - name: prod
    url: https://charts.company.com/prod/index.yaml
    scanInterval: 5m   # Production more frequent
  
  - name: staging
    url: https://charts.company.com/staging/index.yaml
    # Uses default 10m
  
  - name: dev
    url: https://charts.company.com/dev/index.yaml
    scanInterval: 30m  # Development less frequent

scanInterval: 10m  # Default for repos without specific interval
```

## Troubleshooting

### Issue: Repository Not Being Scraped

**Check**: Verify the repository has a valid interval
```bash
# Look for "Started scraper for X with interval Y" in logs
kubectl logs -n monitoring deployment/helm-s3-exporter | grep "Started scraper"
```

### Issue: Too Frequent Scraping

**Solution**: Increase the repository's `scanInterval`
```yaml
- name: repo
  url: https://...
  scanInterval: 30m  # Increase from previous value
```

### Issue: Interval Not Applied

**Check**: Ensure YAML syntax is correct
```bash
# Validate YAML
yamllint config.yaml

# Check loaded config in logs
kubectl logs -n monitoring deployment/helm-s3-exporter | head -20
```

## Future Enhancements

Potential future improvements:

1. **Dynamic Intervals**: Adjust interval based on repository activity
2. **Scheduled Windows**: Only scrape during specific time windows
3. **Conditional Intervals**: Different intervals based on time of day
4. **Rate Limiting**: Built-in rate limit awareness per repository
5. **Retry Logic**: Exponential backoff for failed scrapes

## Summary

Per-repository scan intervals provide:
- ✅ **Flexibility**: Each repo has its own schedule
- ✅ **Efficiency**: Don't over-scrape stable repos
- ✅ **Control**: Fine-tune monitoring per repository
- ✅ **Backward Compatible**: Repos without intervals use default
- ✅ **Simple Configuration**: Just add `scanInterval` to repository

This feature enables more efficient and targeted monitoring of Helm chart repositories!

