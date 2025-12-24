# Grafana Dashboard for Helm S3 Exporter

This guide explains how to import and use the Grafana dashboard for the Helm S3 Exporter.

## Dashboard Overview

The dashboard provides comprehensive visualization of your Helm repositories:

### Panels Included

1. **Overview Statistics** (Top Row)
   - Total Charts across all repositories
   - Total Versions
   - Total Repositories being monitored
   - Scrape Errors in the last hour

2. **Charts per Repository** (Time Series)
   - Track the number of charts in each repository over time
   - Separate line for each repository

3. **Versions per Repository** (Time Series)
   - Track the total number of chart versions per repository
   - Helps identify growing repositories

4. **Average Scrape Duration** (Time Series)
   - Monitor how long it takes to scrape each repository
   - Identify slow repositories
   - Shows mean and current values

5. **Scrape Error Rate** (Time Series)
   - Monitor scrape failures per repository
   - Helps identify problematic repositories

6. **Repository Statistics** (Table)
   - Repository name
   - Number of charts
   - Number of versions
   - Last successful scrape timestamp

7. **Top 20 Charts by Version Count** (Table)
   - Most versioned charts across all repositories
   - Sortable by repository, chart name, or version count

## Installation

### Prerequisites

- Grafana 8.0 or higher
- Prometheus datasource configured in Grafana
- Helm S3 Exporter running and being scraped by Prometheus

### Import Dashboard

#### Method 1: Via UI

1. Open Grafana
2. Go to **Dashboards** → **Import**
3. Click **Upload JSON file**
4. Select `examples/grafana-dashboard.json`
5. Select your Prometheus datasource
6. Click **Import**

#### Method 2: Via API

```bash
# Set your Grafana details
GRAFANA_URL="http://localhost:3000"
GRAFANA_API_KEY="your-api-key"

# Import the dashboard
curl -X POST \
  -H "Authorization: Bearer $GRAFANA_API_KEY" \
  -H "Content-Type: application/json" \
  -d @examples/grafana-dashboard.json \
  "$GRAFANA_URL/api/dashboards/db"
```

#### Method 3: ConfigMap (Kubernetes)

If using Grafana in Kubernetes with dashboard auto-discovery:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: helm-s3-exporter-dashboard
  namespace: monitoring
  labels:
    grafana_dashboard: "1"
data:
  helm-s3-exporter.json: |-
    # Paste contents of grafana-dashboard.json here
```

Then apply:

```bash
kubectl apply -f helm-s3-exporter-dashboard-configmap.yaml
```

### Configure Datasource

After importing, you may need to configure the Prometheus datasource:

1. Open the dashboard
2. Click the gear icon (⚙️) in the top right
3. Go to **Variables**
4. If `DS_PROMETHEUS` exists, set it to your Prometheus datasource
5. Save the dashboard

## Dashboard Features

### Auto-Refresh

The dashboard is configured to auto-refresh every 30 seconds. You can change this:

1. Click the refresh dropdown in the top right
2. Select your preferred interval

### Time Range

Default time range is **Last 6 hours**. To change:

1. Click the time range picker in the top right
2. Select a different range or enter a custom one

### Repository Filtering

Most panels support filtering by repository using Prometheus label filters:

1. Click on any panel's title
2. Select **Edit**
3. Modify the query to add `{repository="your-repo-name"}`
4. Apply changes

### Panel Customization

Each panel can be customized:

- **Clone**: Duplicate a panel to create variations
- **Edit**: Modify queries, thresholds, colors
- **Share**: Get a link or embed code
- **More**: Additional options like inspect data

## Example Queries

The dashboard uses these key queries:

### Total Metrics
```promql
# Total charts across all repositories
sum(helm_s3_charts_total)

# Total versions
sum(helm_s3_versions_total)

# Total repositories
count(count by (repository) (helm_s3_charts_total))
```

### Per-Repository Metrics
```promql
# Charts per repository
helm_s3_charts_total

# Versions per repository
helm_s3_versions_total

# Scrape duration (average)
rate(helm_s3_scrape_duration_seconds_sum[5m]) / 
rate(helm_s3_scrape_duration_seconds_count[5m])
```

### Error Monitoring
```promql
# Scrape errors (1 hour)
sum(increase(helm_s3_scrape_errors_total[1h]))

# Error rate per repository
rate(helm_s3_scrape_errors_total[5m])
```

### Top Charts
```promql
# Top 20 charts by version count
topk(20, helm_s3_chart_versions)
```

## Alerting

### Recommended Alerts

Create alerts based on dashboard panels:

#### High Error Rate
```promql
# Alert when any repository has errors
sum by (repository) (rate(helm_s3_scrape_errors_total[5m])) > 0
```

#### Slow Scrapes
```promql
# Alert when scrape duration exceeds 10 seconds
rate(helm_s3_scrape_duration_seconds_sum[5m]) / 
rate(helm_s3_scrape_duration_seconds_count[5m]) > 10
```

#### Stale Data
```promql
# Alert when last scrape was more than 15 minutes ago
(time() - helm_s3_last_scrape_success) > 900
```

#### No Charts Found
```promql
# Alert when repository has zero charts
helm_s3_charts_total == 0
```

### Creating Alerts in Grafana

1. Edit any panel
2. Go to the **Alert** tab
3. Click **Create Alert**
4. Configure conditions
5. Set notification channels
6. Save

## Troubleshooting

### Dashboard Shows No Data

**Issue**: All panels show "No data"

**Solutions**:
1. Verify Prometheus datasource is configured
2. Check that the exporter is running: `curl http://exporter:9571/metrics`
3. Verify Prometheus is scraping the exporter
4. Check Prometheus targets: `http://prometheus:9090/targets`
5. Verify metrics exist in Prometheus: Query `helm_s3_charts_total`

### Wrong Datasource

**Issue**: Dashboard shows datasource errors

**Solutions**:
1. Open dashboard settings
2. Go to **Variables** tab
3. Edit `DS_PROMETHEUS` variable
4. Select correct Prometheus datasource
5. Save dashboard

### Panels Show Partial Data

**Issue**: Some panels work, others don't

**Solutions**:
1. Check if specific metrics are missing: `curl http://exporter:9571/metrics | grep helm_s3`
2. Verify the exporter version supports all metrics
3. Check Prometheus scrape configuration
4. Look for errors in exporter logs

### Time Series Not Showing

**Issue**: Time series panels are empty but instant queries work

**Solutions**:
1. Prometheus needs historical data - wait for a few scrape intervals
2. Adjust time range to cover when data exists
3. Check Prometheus retention settings

## Customization Examples

### Add Repository Filter Variable

1. Go to dashboard settings
2. Click **Variables** → **Add variable**
3. Configure:
   - **Name**: `repository`
   - **Type**: Query
   - **Query**: `label_values(helm_s3_charts_total, repository)`
   - **Multi-value**: Yes
   - **Include All**: Yes
4. Save dashboard
5. Update panel queries to use `{repository=~"$repository"}`

### Create Custom Panel

Example: Show charts added in last 24 hours

1. Add new panel
2. Use query:
   ```promql
   sum by (repository) (
     helm_s3_chart_age_newest_seconds > (time() - 86400)
   )
   ```
3. Configure visualization
4. Save

### Change Color Scheme

1. Edit panel
2. Go to **Field** tab
3. Under **Color scheme**, select different palette
4. Apply to similar panels for consistency

## Best Practices

1. **Use Folders**: Organize dashboards in folders by team or function
2. **Tag Dashboards**: Add tags like `helm`, `charts`, `monitoring`
3. **Share Links**: Use shortened URLs to share specific time ranges
4. **Export Regularly**: Backup dashboard JSON to version control
5. **Document Changes**: Use dashboard version history
6. **Set Alerts**: Don't just visualize - alert on issues
7. **Use Variables**: Make dashboards reusable with variables

## Advanced Features

### Templating

Create dynamic dashboards using variables:

```promql
# Use in queries
helm_s3_charts_total{repository="$repository"}

# Chain variables
label_values(helm_s3_chart_versions{repository="$repository"}, chart)
```

### Annotations

Add events to your dashboard:

1. Dashboard settings → Annotations
2. Add new annotation
3. Query for events (deployments, incidents)
4. Events will appear on time series panels

### Dashboard Links

Add links to related dashboards:

1. Dashboard settings → Links
2. Add dashboard link
3. Link to related monitoring dashboards

## Performance Optimization

### For Large Deployments

If you have many repositories or charts:

1. **Limit Data Points**:
   ```promql
   helm_s3_charts_total[1h:1m]  # Resolution: 1 minute
   ```

2. **Use Recording Rules**:
   ```yaml
   - record: helm:charts:total
     expr: sum(helm_s3_charts_total)
   ```

3. **Reduce Refresh Rate**: Set to 1m or 5m instead of 30s

4. **Limit Time Range**: Use shorter ranges like 3h or 1h

5. **Use Instant Queries**: For tables, use instant queries instead of ranges

## Support

### Getting Help

1. Check [README.md](../README.md) for general help
2. Review [examples/prometheus-queries.md](prometheus-queries.md) for query help
3. Open GitHub issue with:
   - Dashboard JSON
   - Screenshot of issue
   - Prometheus query results

### Contributing

To improve the dashboard:

1. Export your customized dashboard
2. Submit a pull request with the updated JSON
3. Include screenshots showing improvements
4. Document new panels in this guide

## Additional Resources

- [Grafana Documentation](https://grafana.com/docs/)
- [Prometheus Query Examples](prometheus-queries.md)
- [Helm S3 Exporter README](../README.md)
- [Grafana Best Practices](https://grafana.com/docs/grafana/latest/best-practices/)

---

For more monitoring examples, see the [main README](../README.md).

