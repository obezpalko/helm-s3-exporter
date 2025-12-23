# GitHub Container Registry (GHCR) Usage

This project publishes Docker images to GitHub Container Registry (ghcr.io) automatically on every release.

## Pulling Images

### Latest Version

```bash
docker pull ghcr.io/obezpalko/helm-s3-exporter:latest
```

### Specific Version

```bash
docker pull ghcr.io/obezpalko/helm-s3-exporter:v0.1.0
```

### Major/Minor Versions

```bash
# Pull latest patch version of v0.1.x
docker pull ghcr.io/obezpalko/helm-s3-exporter:0.1

# Pull latest minor version of v0.x.x
docker pull ghcr.io/obezpalko/helm-s3-exporter:0
```

## Using in Kubernetes

### With Helm Chart (Recommended)

The Helm chart already uses GHCR by default:

```bash
helm install my-exporter ./charts/helm-s3-exporter \
  --set s3.bucket=my-bucket
```

### Direct Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: helm-s3-exporter
spec:
  replicas: 1
  selector:
    matchLabels:
      app: helm-s3-exporter
  template:
    metadata:
      labels:
        app: helm-s3-exporter
    spec:
      containers:
      - name: exporter
        image: ghcr.io/obezpalko/helm-s3-exporter:latest
        ports:
        - containerPort: 9571
        env:
        - name: S3_BUCKET
          value: "my-helm-charts"
        - name: S3_REGION
          value: "us-east-1"
```

## Multi-Platform Support

Images are built for multiple architectures:
- **linux/amd64** (x86_64) - Standard Intel/AMD servers
- **linux/arm64** (aarch64) - ARM servers (AWS Graviton, Apple Silicon, etc.)

Docker automatically pulls the correct architecture for your system.

## Image Tags

| Tag Format | Example | Description |
|------------|---------|-------------|
| `latest` | `latest` | Latest stable release |
| `vX.Y.Z` | `v0.1.0` | Specific version |
| `X.Y` | `0.1` | Latest patch of minor version |
| `X` | `0` | Latest minor of major version |

## Using with Docker Compose

```yaml
version: '3.8'
services:
  exporter:
    image: ghcr.io/obezpalko/helm-s3-exporter:latest
    ports:
      - "9571:9571"
    environment:
      - S3_BUCKET=my-helm-charts
      - S3_REGION=us-east-1
      - AWS_ACCESS_KEY_ID=${AWS_ACCESS_KEY_ID}
      - AWS_SECRET_ACCESS_KEY=${AWS_SECRET_ACCESS_KEY}
```

## For Private Repositories

If the image becomes private, you'll need to authenticate:

### Docker Login

```bash
# Generate a GitHub Personal Access Token with read:packages scope
export CR_PAT=YOUR_TOKEN

echo $CR_PAT | docker login ghcr.io -u USERNAME --password-stdin
```

### Kubernetes Pull Secret

```bash
kubectl create secret docker-registry ghcr-secret \
  --docker-server=ghcr.io \
  --docker-username=USERNAME \
  --docker-password=YOUR_TOKEN \
  --docker-email=YOUR_EMAIL

# Use in deployment
kubectl patch serviceaccount default -p '{"imagePullSecrets": [{"name": "ghcr-secret"}]}'
```

Or in Helm values:

```yaml
imagePullSecrets:
  - name: ghcr-secret
```

## Advantages of GHCR

✅ **No additional setup** - Works automatically with GitHub Actions  
✅ **Free for public repos** - Unlimited storage and bandwidth  
✅ **Integrated with GitHub** - Same authentication as your repo  
✅ **Automatic versioning** - Tags match your Git releases  
✅ **Multi-platform builds** - Supports amd64 and arm64  
✅ **Built-in caching** - Faster builds with layer caching  

## Verifying Images

Check available tags:

```bash
# Using GitHub API
curl -H "Authorization: Bearer YOUR_TOKEN" \
  https://api.github.com/user/packages/container/helm-s3-exporter/versions
```

Or visit: https://github.com/obezpalko/helm-s3-exporter/pkgs/container/helm-s3-exporter

## Troubleshooting

### Error: "denied: permission_denied"

**Solution**: The image might be private. Follow the "Private Repositories" section above.

### Error: "manifest unknown"

**Solution**: The requested tag doesn't exist. Check available tags on GitHub.

### Multi-platform issues

If you're having issues with ARM builds:

```bash
# Pull specific platform
docker pull --platform linux/amd64 ghcr.io/obezpalko/helm-s3-exporter:latest
```

## Security

- Images are scanned for vulnerabilities
- Built from scratch (distroless base)
- Runs as non-root user (UID 65532)
- Read-only root filesystem

## References

- [GitHub Container Registry Documentation](https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry)
- [Docker Multi-platform Images](https://docs.docker.com/build/building/multi-platform/)

