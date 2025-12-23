# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 0.1.x   | :white_check_mark: |

## Security Best Practices

### Authentication

#### Recommended: IAM Roles for Service Accounts (IRSA)

For production deployments on AWS EKS, always use IRSA:

```yaml
auth:
  useIAMRole: true

serviceAccount:
  create: true
  annotations:
    eks.amazonaws.com/role-arn: arn:aws:iam::ACCOUNT:role/ROLE_NAME
```

**Benefits:**
- No static credentials in cluster
- Automatic credential rotation
- Fine-grained IAM permissions
- Audit trail via CloudTrail

#### Using Existing Secrets

If you must use static credentials, use Kubernetes secrets:

```yaml
auth:
  useIAMRole: false
  existingSecret: "aws-credentials"
```

Create the secret securely:
```bash
kubectl create secret generic aws-credentials \
  --from-literal=AWS_ACCESS_KEY_ID=xxx \
  --from-literal=AWS_SECRET_ACCESS_KEY=yyy \
  --namespace=your-namespace
```

#### ⚠️ Avoid: Inline Credentials

**NEVER** use inline credentials in values.yaml for production:

```yaml
# ❌ DO NOT DO THIS IN PRODUCTION
auth:
  credentials:
    accessKeyId: "AKIAIOSFODNN7EXAMPLE"
    secretAccessKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
```

This is only acceptable for local development/testing.

### IAM Permissions

Follow the principle of least privilege. The exporter only needs:

- `s3:ListBucket` on the bucket
- `s3:GetObject` on objects in the bucket

Example minimal IAM policy:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": ["s3:ListBucket"],
      "Resource": "arn:aws:s3:::your-bucket"
    },
    {
      "Effect": "Allow",
      "Action": ["s3:GetObject"],
      "Resource": "arn:aws:s3:::your-bucket/*"
    }
  ]
}
```

### Container Security

The default configuration follows security best practices:

- **Non-root user**: Runs as UID 65532 (nonroot)
- **Read-only root filesystem**: Prevents runtime modifications
- **No privilege escalation**: `allowPrivilegeEscalation: false`
- **Dropped capabilities**: All Linux capabilities dropped
- **Distroless base image**: Minimal attack surface

### Network Security

- The exporter only makes outbound connections to S3
- No inbound connections except for metrics/health endpoints
- Use NetworkPolicies to restrict traffic:

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: helm-s3-exporter
spec:
  podSelector:
    matchLabels:
      app.kubernetes.io/name: helm-s3-exporter
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
      port: 443  # HTTPS to S3
```

### Secrets Management

For managing AWS credentials, consider:

1. **AWS Secrets Manager**: Use [External Secrets Operator](https://external-secrets.io/)
2. **HashiCorp Vault**: Use [Vault Agent Injector](https://www.vaultproject.io/docs/platform/k8s)
3. **Sealed Secrets**: Use [Bitnami Sealed Secrets](https://github.com/bitnami-labs/sealed-secrets)

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

- [ ] Use IRSA or external secrets manager
- [ ] Apply minimal IAM permissions
- [ ] Enable Pod Security Standards
- [ ] Configure resource limits
- [ ] Set up network policies
- [ ] Enable audit logging
- [ ] Configure monitoring and alerting
- [ ] Review and rotate credentials regularly
- [ ] Keep the exporter updated
- [ ] Scan container images for vulnerabilities

## Compliance

The exporter is designed to help meet common compliance requirements:

- **SOC 2**: Audit logging, access controls
- **PCI DSS**: Network isolation, minimal privileges
- **HIPAA**: Encryption in transit (S3 HTTPS)
- **GDPR**: No PII collection

## Additional Resources

- [AWS EKS Best Practices - Security](https://aws.github.io/aws-eks-best-practices/security/docs/)
- [Kubernetes Security Best Practices](https://kubernetes.io/docs/concepts/security/)
- [CIS Kubernetes Benchmark](https://www.cisecurity.org/benchmark/kubernetes)

