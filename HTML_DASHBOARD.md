# HTML Dashboard Features

The Helm S3 Exporter includes an optional HTML dashboard for visualizing chart information.

## Features

### 📊 Overview Statistics
- Total number of charts
- Total number of versions
- Oldest chart date
- Newest chart date

### 🔍 Real-Time Filtering
- **Filter by Chart Name**: Type to instantly filter charts
- **Keyboard Shortcuts**:
  - Press `/` to focus the search box
  - Press `Escape` to clear the search and unfocus
- **Live Counter**: Shows how many charts match your filter

### 📦 Chart Information
Each chart card displays:
- Chart icon (if available)
- Chart name
- Description
- Number of versions
- Oldest version date
- Newest version date

### 📋 Expandable Version Lists
- Click on the version badge to expand/collapse the version list
- Each version shows:
  - Version number
  - Release date
  - **Download link** (direct link to the chart package)

## Enabling the Dashboard

### Method 1: Inline Configuration

```yaml
# values.yaml
config:
  inline:
    enabled: true
    url: "https://charts.bitnami.com/bitnami/index.yaml"

html:
  enabled: true
  path: /charts
```

### Method 2: Config File

```yaml
# config.yaml
repositories:
  - name: bitnami
    url: https://charts.bitnami.com/bitnami/index.yaml

enableHTML: true
htmlPath: /charts
```

### Method 3: Helm Chart

```bash
helm install my-exporter ./charts/helm-s3-exporter \
  --set config.inline.enabled=true \
  --set config.inline.url=https://charts.bitnami.com/bitnami/index.yaml \
  --set html.enabled=true
```

## Accessing the Dashboard

### Local Development

```bash
# Run the exporter
CONFIG_FILE=config.yaml ./bin/exporter

# Open in browser
open http://localhost:9571/charts
```

### Kubernetes

```bash
# Port forward
kubectl port-forward -n monitoring svc/helm-s3-exporter 9571:9571

# Open in browser
open http://localhost:9571/charts
```

### Ingress

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: helm-s3-exporter
  namespace: monitoring
spec:
  rules:
  - host: helm-charts.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: helm-s3-exporter
            port:
              number: 9571
```

Then access at: `https://helm-charts.example.com/charts`

## Usage Examples

### Filtering Charts

1. **Find specific chart**:
   - Type `nginx` in the filter box
   - Only charts with "nginx" in the name will be shown

2. **Find charts by prefix**:
   - Type `cert-` to find all cert-manager related charts

3. **Clear filter**:
   - Press `Escape` or clear the text box

### Viewing Versions

1. **Expand versions**:
   - Click on the version badge (e.g., "127 versions")
   - The list will expand showing all versions

2. **Download a specific version**:
   - Expand the versions list
   - Click the "Download" link next to the version you want
   - The chart package will be downloaded

3. **Collapse versions**:
   - Click the version badge again to collapse

## Screenshots

### Main Dashboard
```
┌─────────────────────────────────────────────────────────────┐
│ 🚀 Helm Repository Dashboard                                │
│ Generated: 2025-12-24 10:00:00 UTC                          │
└─────────────────────────────────────────────────────────────┘

┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐
│   144    │ │  18,330  │ │2022-01-20│ │2025-12-23│
│  Charts  │ │ Versions │ │  Oldest  │ │  Newest  │
└──────────┘ └──────────┘ └──────────┘ └──────────┘

┌─────────────────────────────────────────────────────────────┐
│ Filter by Chart Name: [nginx_________________]              │
│ Showing 5 of 144 charts                                     │
└─────────────────────────────────────────────────────────────┘

📦 Available Charts

┌─────────────────────────────────────────────────────────────┐
│ [icon] nginx                                                 │
│ NGINX Open Source is a web server...                        │
│ [127 versions ▼] Oldest: 2020-01-15  Newest: 2025-12-20    │
└─────────────────────────────────────────────────────────────┘
```

### Expanded Versions
```
┌─────────────────────────────────────────────────────────────┐
│ [icon] nginx                                                 │
│ NGINX Open Source is a web server...                        │
│ [127 versions ▲] Oldest: 2020-01-15  Newest: 2025-12-20    │
│                                                              │
│ ┌────────────────────────────────────────────────────────┐ │
│ │ 18.2.5 • 2025-12-20                      [Download]    │ │
│ │ 18.2.4 • 2025-12-15                      [Download]    │ │
│ │ 18.2.3 • 2025-12-10                      [Download]    │ │
│ │ 18.2.2 • 2025-12-05                      [Download]    │ │
│ │ ...                                                     │ │
│ └────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

## Customization

### Change HTML Path

```yaml
html:
  enabled: true
  path: /dashboard  # Default is /charts
```

### Disable HTML Dashboard

```yaml
html:
  enabled: false
```

## Technical Details

### Data Source
- The dashboard displays aggregated data from all configured repositories
- Data is updated each time a repository is scraped
- For multiple repositories, the dashboard shows the combined view

### Performance
- Filtering is done client-side using JavaScript
- No server requests are made during filtering
- Expandable lists use CSS transitions for smooth animations

### Browser Compatibility
- Modern browsers (Chrome, Firefox, Safari, Edge)
- Responsive design works on mobile devices
- No external dependencies (pure HTML/CSS/JS)

## Troubleshooting

### Dashboard Not Accessible

**Issue**: Cannot access the dashboard at the configured path

**Solutions**:
1. Verify HTML is enabled: `html.enabled=true`
2. Check the correct path: Default is `/charts`
3. Verify the exporter is running: `curl http://localhost:9571/health`
4. Check logs: `kubectl logs -n monitoring deployment/helm-s3-exporter`

### No Data Shown

**Issue**: Dashboard shows "No data available yet"

**Solutions**:
1. Wait for the initial scrape to complete (check logs)
2. Verify repository configuration is correct
3. Check for scrape errors in metrics: `curl http://localhost:9571/metrics | grep error`

### Download Links Not Working

**Issue**: Download links return 404 or don't work

**Solutions**:
1. Verify the repository's `index.yaml` contains valid URLs
2. Check if the chart packages are actually hosted at those URLs
3. Some repositories may require authentication for downloads

### Filter Not Working

**Issue**: Typing in the filter box doesn't filter charts

**Solutions**:
1. Check browser console for JavaScript errors
2. Try refreshing the page
3. Verify you're using a modern browser

## Advanced Usage

### Bookmarklet for Quick Access

Create a bookmarklet to quickly open the dashboard:

```javascript
javascript:(function(){window.open('http://localhost:9571/charts','_blank')})()
```

### Embedding in Other Dashboards

The HTML dashboard can be embedded in an iframe:

```html
<iframe src="http://helm-s3-exporter:9571/charts" 
        width="100%" 
        height="800px" 
        frameborder="0">
</iframe>
```

### Custom Styling

The dashboard uses inline CSS. To customize:
1. Fork the repository
2. Edit `internal/web/html.go`
3. Modify the `<style>` section in `htmlTemplate`
4. Rebuild and deploy

## Future Enhancements

Potential future improvements:
- [ ] Multi-repository view with repository tabs
- [ ] Chart comparison feature
- [ ] Version diff viewer
- [ ] Search history
- [ ] Export to CSV/JSON
- [ ] Dark mode toggle
- [ ] Chart dependency visualization
- [ ] Version changelog display

## Feedback

If you have suggestions for the HTML dashboard, please:
1. Open an issue on GitHub
2. Describe the feature or improvement
3. Include screenshots or mockups if applicable

---

For more information, see the [main README](README.md).

