# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 0.1.x   | :white_check_mark: |

## Security Best Practices

### Authentication

The exporter supports multiple authentication methods for accessing Helm repositories:

#### Using Kubernetes Secrets (Recommended)

Store credentials securely in Kubernetes secrets:

```yaml
config:
  existingSecret:
    enabled: true
    name: "helm-repo-credentials"
```

Create the secret securely:
```bash
kubectl create secret generic helm-repo-credentials \
  --from-file=config.yaml=config.yaml \
  --namespace=your-namespace
```

#### ⚠️ Avoid: Inline Credentials

**NEVER** use inline credentials in values.yaml for production:

```yaml
# ❌ DO NOT DO THIS IN PRODUCTION
config:
  inline:
    enabled: true
    url: "https://charts.example.com/index.yaml"
    # Credentials should never be here
```

This is only acceptable for local development/testing with public repositories.

### Repository Access Permissions

Follow the principle of least privilege. The exporter only needs:

- Read access to `index.yaml` files
- No write permissions required
- No access to chart packages (only metadata)

Ensure your repository access controls are properly configured.

### Container Security

The default configuration follows security best practices:

- **Non-root user**: Runs as UID 65532 (nonroot)
- **Read-only root filesystem**: Prevents runtime modifications
- **No privilege escalation**: `allowPrivilegeEscalation: false`
- **Dropped capabilities**: All Linux capabilities dropped
- **Distroless base image**: Minimal attack surface

### Network Security

- The exporter only makes outbound HTTPS connections to configured repositories
- No inbound connections except for metrics/health endpoints
- Use NetworkPolicies to restrict traffic:

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: helm-repo-exporter
spec:
  podSelector:
    matchLabels:
      app.kubernetes.io/name: helm-repo-exporter
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: monitoring
    ports:
    - protocol: TCP
      port: 9571
  egress:
  - to:
    - namespaceSelector: {}
    ports:
    - protocol: TCP
      port: 53  # DNS
  - to:
    - podSelector: {}
    ports:
    - protocol: TCP
      port: 443  # HTTPS to repositories
```

### Secrets Management

For managing repository credentials, consider:

1. **External Secrets Operator**: Use [External Secrets Operator](https://external-secrets.io/)
2. **HashiCorp Vault**: Use [Vault Agent Injector](https://www.vaultproject.io/docs/platform/k8s)
3. **Sealed Secrets**: Use [Bitnami Sealed Secrets](https://github.com/bitnami-labs/sealed-secrets)
4. **Cloud Provider Secrets**: Use your cloud provider's secret management service

### Pod Security Standards

The exporter is compatible with the `restricted` Pod Security Standard:

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: monitoring
  labels:
    pod-security.kubernetes.io/enforce: restricted
    pod-security.kubernetes.io/audit: restricted
    pod-security.kubernetes.io/warn: restricted
```

## Reporting a Vulnerability

If you discover a security vulnerability, please follow these steps:

1. **Do NOT** open a public GitHub issue
2. Email the maintainers at: [security@example.com]
3. Include:
   - Description of the vulnerability
   - Steps to reproduce
   - Potential impact
   - Suggested fix (if any)

### Response Timeline

- **Initial Response**: Within 48 hours
- **Status Update**: Within 7 days
- **Fix Timeline**: Depends on severity
  - Critical: 7 days
  - High: 14 days
  - Medium: 30 days
  - Low: 60 days

## Security Updates

Security updates will be released as patch versions and announced via:

- GitHub Security Advisories
- Release notes
- README.md

## Security Checklist

Before deploying to production:

- [ ] Use external secrets manager for credentials
- [ ] Apply minimal repository access permissions
- [ ] Enable Pod Security Standards
- [ ] Configure resource limits
- [ ] Set up network policies
- [ ] Enable audit logging
- [ ] Configure monitoring and alerting
- [ ] Review and rotate credentials regularly
- [ ] Keep the exporter updated
- [ ] Scan container images for vulnerabilities
- [ ] Use HTTPS for all repository URLs
- [ ] Validate repository URLs and authentication

## Compliance

The exporter is designed to help meet common compliance requirements:

- **SOC 2**: Audit logging, access controls
- **PCI DSS**: Network isolation, minimal privileges
- **HIPAA**: Encryption in transit (HTTPS)
- **GDPR**: No PII collection

## Additional Resources

- [Kubernetes Security Best Practices](https://kubernetes.io/docs/concepts/security/)
- [CIS Kubernetes Benchmark](https://www.cisecurity.org/benchmark/kubernetes)
- [OWASP Kubernetes Security](https://owasp.org/www-project-kubernetes-top-ten/)

