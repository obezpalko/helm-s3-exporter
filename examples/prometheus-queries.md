# Prometheus Query Examples

This document provides example PromQL queries for the Helm S3 Exporter metrics.

## Basic Queries

### Repository Statistics

```promql
# Total number of charts per repository
helm_repo_charts_total

# Total chart versions per repository
helm_repo_versions_total

# Specific repository stats
helm_repo_charts_total{repository="production"}
helm_repo_versions_total{repository="bitnami"}
```

### Aggregate Across Repositories

```promql
# Total charts across all repositories
sum(helm_repo_charts_total)

# Total versions across all repositories
sum(helm_repo_versions_total)

# Average charts per repository
avg(helm_repo_charts_total)
```

## Chart Analysis

### Version Counts

```promql
# Charts with many versions (>10) in production
helm_repo_chart_versions{repository="production"} > 10

# Chart with most versions in bitnami repo
topk(1, helm_repo_chart_versions{repository="bitnami"})

# Charts with few versions (<5) across all repos
helm_repo_chart_versions < 5

# Top 10 charts by version count in a specific repo
topk(10, helm_repo_chart_versions{repository="production"})
```

### Chart Age

```promql
# Oldest chart in repository (in days)
(time() - helm_repo_overall_age_oldest_seconds{repository="production"}) / 86400

# Newest chart age in hours
(time() - helm_repo_overall_age_newest_seconds) / 3600

# Charts not updated in 90 days
(time() - helm_repo_chart_age_newest_seconds) / 86400 > 90

# Charts updated in last 7 days
(time() - helm_repo_chart_age_newest_seconds) / 86400 < 7
```

## Performance Monitoring

### Scrape Duration

```promql
# Average scrape duration per repository
rate(helm_repo_scrape_duration_seconds_sum[5m]) / rate(helm_repo_scrape_duration_seconds_count[5m])

# Slowest repository scrapes
topk(3, rate(helm_repo_scrape_duration_seconds_sum[5m]) / rate(helm_repo_scrape_duration_seconds_count[5m]))

# Scrape duration > 5 seconds
rate(helm_repo_scrape_duration_seconds_sum[5m]) / rate(helm_repo_scrape_duration_seconds_count[5m]) > 5
```

### Scrape Errors

```promql
# Error rate per repository
rate(helm_repo_scrape_errors_total[5m])

# Total errors in last hour per repository
increase(helm_repo_scrape_errors_total[1h])

# Repositories with errors
sum by (repository) (increase(helm_repo_scrape_errors_total[1h])) > 0

# Error rate across all repositories
sum(rate(helm_repo_scrape_errors_total[5m]))
```

### Success Monitoring

```promql
# Last successful scrape time per repository
helm_repo_last_scrape_success

# Repositories not scraped in last 10 minutes
(time() - helm_repo_last_scrape_success) > 600

# Time since last successful scrape (in minutes)
(time() - helm_repo_last_scrape_success) / 60
```

## Repository Comparisons

### Compare Repositories

```promql
# Chart count comparison
sum by (repository) (helm_repo_chart_versions)

# Repository sizes (total versions)
helm_repo_versions_total

# Largest repositories
topk(5, helm_repo_versions_total)

# Smallest repositories
bottomk(5, helm_repo_versions_total)
```

### Multi-Repository Queries

```promql
# Total charts in production and staging
sum(helm_repo_charts_total{repository=~"production|staging"})

# Compare scrape durations between environments
rate(helm_repo_scrape_duration_seconds_sum{repository=~"prod.*"}[5m]) / 
rate(helm_repo_scrape_duration_seconds_count{repository=~"prod.*"}[5m])

# Charts only in production (not in other repos)
helm_repo_chart_versions{repository="production"} unless helm_repo_chart_versions{repository!="production"}
```

## Chart-Specific Queries

### Specific Chart Analysis

```promql
# Versions of a specific chart across repositories
helm_repo_chart_versions{chart="nginx"}

# Nginx chart in production
helm_repo_chart_versions{repository="production",chart="nginx"}

# When was nginx last updated (in days ago)
(time() - helm_repo_chart_age_newest_seconds{chart="nginx"}) / 86400

# Age of oldest nginx version
(time() - helm_repo_chart_age_oldest_seconds{chart="nginx"}) / 86400
```

### Chart Comparison

```promql
# Compare versions of the same chart across repos
helm_repo_chart_versions{chart="cert-manager"}

# Charts with more versions in staging than production
helm_repo_chart_versions{repository="staging"} > 
ignoring(repository) helm_repo_chart_versions{repository="production"}
```

## Alerting Queries

### Alert Examples

```promql
# Alert: Repository hasn't been scraped in 15 minutes
(time() - helm_repo_last_scrape_success) > 900

# Alert: High error rate
rate(helm_repo_scrape_errors_total[5m]) > 0.1

# Alert: Slow scrapes (>10 seconds average)
rate(helm_repo_scrape_duration_seconds_sum[5m]) / 
rate(helm_repo_scrape_duration_seconds_count[5m]) > 10

# Alert: No charts found (potential configuration issue)
helm_repo_charts_total == 0

# Alert: Chart not updated in 180 days (6 months)
(time() - helm_repo_chart_age_newest_seconds) / 86400 > 180
```

## Advanced Queries

### Trends and Rates

```promql
# Chart growth rate (if tracking over time)
rate(helm_repo_charts_total[1h])

# Version addition rate
rate(helm_repo_versions_total[1h])

# Scrape frequency
rate(helm_repo_scrape_duration_seconds_count[5m])
```

### Statistical Analysis

```promql
# Standard deviation of scrape duration
stddev_over_time(helm_repo_scrape_duration_seconds_sum[1h])

# Quantile scrape duration (95th percentile)
histogram_quantile(0.95, rate(helm_repo_scrape_duration_seconds_bucket[5m]))

# Median chart age across all charts
helm_repo_overall_age_median_seconds
```

### Filtering and Grouping

```promql
# Group charts by repository and count
count by (repository) (helm_repo_chart_versions)

# Sum versions by repository
sum by (repository) (helm_repo_chart_versions)

# Filter production repositories only
{__name__=~"helm_repo_.*", repository=~"prod.*"}

# Exclude specific repositories
helm_repo_charts_total{repository!~"test|dev"}
```

## Dashboard Queries

### Overview Panel

```promql
# Total Repositories
count(count by (repository) (helm_repo_charts_total))

# Total Charts Across All Repos
sum(helm_repo_charts_total)

# Total Versions
sum(helm_repo_versions_total)

# Overall Error Rate
sum(rate(helm_repo_scrape_errors_total[5m]))
```

### Per-Repository Panel

```promql
# Charts per repository (table)
helm_repo_charts_total

# Versions per repository (table)
helm_repo_versions_total

# Last scrape time (table)
helm_repo_last_scrape_success

# Scrape duration (graph)
rate(helm_repo_scrape_duration_seconds_sum[5m]) / 
rate(helm_repo_scrape_duration_seconds_count[5m])
```

### Chart Details Panel

```promql
# Top 20 charts by version count
topk(20, helm_repo_chart_versions)

# Recently updated charts (last 24h)
(time() - helm_repo_chart_age_newest_seconds) / 3600 < 24

# Stale charts (not updated in 90 days)
(time() - helm_repo_chart_age_newest_seconds) / 86400 > 90
```

## Tips

1. **Use `repository` label**: Filter all metrics by repository using `{repository="name"}`
2. **Aggregate**: Use `sum()`, `avg()`, `max()`, `min()` for cross-repository queries
3. **Regex**: Use `{repository=~"prod.*"}` for pattern matching
4. **Time calculations**: Convert timestamps to days with `/ 86400`, hours with `/ 3600`
5. **Rate functions**: Use `rate()` for counters and histograms over time
6. **Recording rules**: Create recording rules for frequently used complex queries

## Recording Rules Example

```yaml
groups:
  - name: helm_repo_exporter
    interval: 1m
    rules:
      # Total charts across all repositories
      - record: helm_s3:charts:total
        expr: sum(helm_repo_charts_total)
      
      # Average scrape duration per repository
      - record: helm_s3:scrape_duration:avg
        expr: |
          rate(helm_repo_scrape_duration_seconds_sum[5m]) / 
          rate(helm_repo_scrape_duration_seconds_count[5m])
      
      # Chart version distribution
      - record: helm_s3:chart_versions:sum_by_repository
        expr: sum by (repository) (helm_repo_chart_versions)
```

---

For more examples and best practices, see the [Prometheus documentation](https://prometheus.io/docs/prometheus/latest/querying/basics/).

