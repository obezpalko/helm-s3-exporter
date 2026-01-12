# Docker Compose Deployment Guide

This guide explains how to deploy the Helm Repository Exporter using Docker Compose.

## Quick Reference

```bash
# Start the exporter
docker compose up -d

# Start with Prometheus and Grafana
docker compose --profile with-prometheus up -d

# View logs
docker compose logs -f

# Stop services
docker compose down

# Rebuild and restart
docker compose up -d --build

# Using Makefile
make compose-up          # Start services
make compose-down        # Stop services
make compose-logs        # View logs
make compose-up-full     # Start with Prometheus/Grafana
```

## Quick Start

### 1. Basic Setup

Run the exporter with the default configuration:

```bash
docker compose up -d
```

This will:
- Build the Docker image
- Start the exporter on port 9571
- Use the `config.yaml` file for configuration
- Enable the HTML dashboard at http://localhost:9571/charts
- Expose Prometheus metrics at http://localhost:9571/metrics

### 2. View Logs

```bash
docker compose logs -f helm-repo-exporter
```

### 3. Check Status

```bash
# Health check
curl http://localhost:9571/health

# View metrics
curl http://localhost:9571/metrics

# View HTML dashboard
open http://localhost:9571/charts
```

### 4. Stop the Service

```bash
docker compose down
```

## Configuration Options

### Option 1: Using Configuration File (Recommended)

Edit the `config.yaml` file to configure your repositories:

```yaml
repositories:
  - name: bitnami
    url: https://charts.bitnami.com/bitnami/index.yaml
    scanInterval: 5m

  - name: private-repo
    url: https://charts.example.com/index.yaml
    auth:
      basic:
        username: myuser
        password: mypassword

scanInterval: 5m
scanTimeout: 30s
metricsPort: "9571"
enableHTML: true
```

Then start the service:

```bash
docker compose up -d
```

### Option 2: Using Environment Variables

For a simple single-repository setup, you can use environment variables instead. Edit `compose.yaml`:

```yaml
services:
  helm-repo-exporter:
    environment:
      # Comment out CONFIG_FILE
      # - CONFIG_FILE=/config/config.yaml
      
      # Use these instead
      - INDEX_URL=https://charts.bitnami.com/bitnami/index.yaml
      - SCAN_INTERVAL=5m
      - SCAN_TIMEOUT=30s
      - ENABLE_HTML=true
```

## Running with Prometheus and Grafana

To run the complete monitoring stack with Prometheus and Grafana:

```bash
docker compose --profile with-prometheus up -d
```

This will start:
- **Helm Repository Exporter** on port 9571
- **Prometheus** on port 9090
- **Grafana** on port 3000

### Access the Services

| Service | URL | Credentials |
|---------|-----|-------------|
| Helm Repo Exporter | http://localhost:9571/charts | - |
| Prometheus | http://localhost:9090 | - |
| Grafana | http://localhost:3000 | admin / admin |

### Configure Grafana

1. Open Grafana at http://localhost:3000
2. Login with `admin` / `admin`
3. Add Prometheus data source:
   - Go to Configuration → Data Sources
   - Add Prometheus
   - URL: `http://prometheus:9090`
   - Click "Save & Test"
4. Import the dashboard from `examples/grafana-dashboard.json`

## Advanced Configuration

### Custom Ports

To change the exposed ports, edit `compose.yaml`:

```yaml
services:
  helm-repo-exporter:
    ports:
      - "8080:9571"  # Expose on port 8080 instead
```

### Using External Configuration

Mount your own configuration file:

```yaml
services:
  helm-repo-exporter:
    volumes:
      - /path/to/your/config.yaml:/config/config.yaml:ro
```

### Private Repositories with Credentials

For private repositories, update `config.yaml` with authentication:

```yaml
repositories:
  - name: private-repo
    url: https://charts.example.com/index.yaml
    auth:
      basic:
        username: ${REPO_USERNAME}
        password: ${REPO_PASSWORD}
```

Then create a `.env` file:

```bash
REPO_USERNAME=myuser
REPO_PASSWORD=mypassword
```

And update `compose.yaml`:

```yaml
services:
  helm-repo-exporter:
    env_file:
      - .env
```

### Resource Limits

Add resource limits to prevent excessive resource usage:

```yaml
services:
  helm-repo-exporter:
    deploy:
      resources:
        limits:
          cpus: '0.5'
          memory: 256M
        reservations:
          cpus: '0.25'
          memory: 128M
```

## Troubleshooting

### Check Container Status

```bash
docker compose ps
```

### View Logs

```bash
# All services
docker compose logs

# Specific service
docker compose logs helm-repo-exporter

# Follow logs
docker compose logs -f
```

### Restart Services

```bash
docker compose restart helm-repo-exporter
```

### Rebuild Image

If you've made changes to the code:

```bash
docker compose build --no-cache
docker compose up -d
```

### Health Check Failed

If the health check fails:

```bash
# Check if the service is responding
curl http://localhost:9571/health

# Check logs for errors
docker compose logs helm-repo-exporter

# Verify configuration
docker compose exec helm-repo-exporter cat /config/config.yaml
```

### Connection Issues

If you can't connect to repositories:

```bash
# Test from inside the container
docker compose exec helm-repo-exporter wget -O- https://charts.bitnami.com/bitnami/index.yaml

# Check DNS resolution
docker compose exec helm-repo-exporter nslookup charts.bitnami.com
```

## Maintenance

### Update to Latest Version

```bash
# Pull latest code
git pull

# Rebuild and restart
docker compose build
docker compose up -d
```

### Backup Configuration

```bash
# Backup configuration
cp config.yaml config.yaml.backup

# Backup Prometheus data
docker compose exec prometheus tar czf /tmp/prometheus-backup.tar.gz /prometheus
docker compose cp prometheus:/tmp/prometheus-backup.tar.gz ./prometheus-backup.tar.gz
```

### Clean Up

```bash
# Stop and remove containers
docker compose down

# Remove volumes (WARNING: This deletes all data)
docker compose down -v

# Remove images
docker compose down --rmi all
```

## Production Deployment

For production deployments, consider:

1. **Use secrets management** for credentials (Docker Secrets, Vault, etc.)
2. **Enable TLS** for the metrics endpoint
3. **Set up monitoring** and alerting
4. **Configure log rotation**
5. **Use persistent volumes** for Prometheus data
6. **Set resource limits**
7. **Enable automatic restarts**
8. **Use a reverse proxy** (nginx, traefik) for HTTPS

Example production `compose.yaml` additions:

```yaml
services:
  helm-repo-exporter:
    restart: always
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"
    deploy:
      resources:
        limits:
          cpus: '1'
          memory: 512M
```

## Examples

### Monitor Multiple Repositories

```yaml
# config.yaml
repositories:
  - name: bitnami
    url: https://charts.bitnami.com/bitnami/index.yaml
    scanInterval: 5m
  - name: prometheus
    url: https://prometheus-community.github.io/helm-charts/index.yaml
    scanInterval: 10m
  - name: grafana
    url: https://grafana.github.io/helm-charts/index.yaml
    scanInterval: 10m
```

### Different Scan Intervals

```yaml
# config.yaml
repositories:
  - name: critical-repo
    url: https://charts.example.com/critical/index.yaml
    scanInterval: 1m  # Check every minute
  - name: normal-repo
    url: https://charts.example.com/normal/index.yaml
    scanInterval: 10m  # Check every 10 minutes
```

### S3-backed Repositories

```yaml
# config.yaml
repositories:
  - name: s3-repo
    url: s3://my-bucket/helm-charts/index.yaml
    scanInterval: 5m
```

## Support

For more information:
- [Main README](README.md)
- [Configuration Guide](examples/CONFIGURATION_GUIDE.md)
- [Quick Start Guide](QUICKSTART.md)
- [Security Guidelines](SECURITY.md)
