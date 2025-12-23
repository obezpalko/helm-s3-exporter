#!/bin/bash
# Script to setup IRSA (IAM Roles for Service Accounts) for Helm S3 Exporter
# 
# Prerequisites:
# - AWS CLI configured
# - eksctl installed
# - kubectl configured for your EKS cluster
# 
# Usage:
#   ./setup-irsa.sh <cluster-name> <region> <s3-bucket-name> <namespace>
#
# Example:
#   ./setup-irsa.sh my-cluster us-west-2 my-helm-charts monitoring

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

if [ $# -ne 4 ]; then
    echo -e "${RED}Error: Invalid number of arguments${NC}"
    echo "Usage: $0 <cluster-name> <region> <s3-bucket-name> <namespace>"
    exit 1
fi

CLUSTER_NAME=$1
REGION=$2
S3_BUCKET=$3
NAMESPACE=$4
SERVICE_ACCOUNT="helm-s3-exporter"
ROLE_NAME="helm-s3-exporter-role"

echo -e "${GREEN}Setting up IRSA for Helm S3 Exporter${NC}"
echo "Cluster: $CLUSTER_NAME"
echo "Region: $REGION"
echo "S3 Bucket: $S3_BUCKET"
echo "Namespace: $NAMESPACE"
echo ""

# Create namespace if it doesn't exist
echo -e "${YELLOW}Creating namespace $NAMESPACE if it doesn't exist...${NC}"
kubectl create namespace $NAMESPACE --dry-run=client -o yaml | kubectl apply -f -

# Create IAM policy
echo -e "${YELLOW}Creating IAM policy...${NC}"
POLICY_ARN=$(aws iam create-policy \
    --policy-name ${ROLE_NAME}-policy \
    --policy-document "{
        \"Version\": \"2012-10-17\",
        \"Statement\": [
            {
                \"Effect\": \"Allow\",
                \"Action\": [
                    \"s3:ListBucket\"
                ],
                \"Resource\": \"arn:aws:s3:::${S3_BUCKET}\"
            },
            {
                \"Effect\": \"Allow\",
                \"Action\": [
                    \"s3:GetObject\",
                    \"s3:GetObjectVersion\"
                ],
                \"Resource\": \"arn:aws:s3:::${S3_BUCKET}/*\"
            }
        ]
    }" \
    --query 'Policy.Arn' \
    --output text 2>/dev/null || \
    aws iam list-policies --query "Policies[?PolicyName=='${ROLE_NAME}-policy'].Arn" --output text)

echo -e "${GREEN}Policy ARN: $POLICY_ARN${NC}"

# Create service account with IRSA
echo -e "${YELLOW}Creating service account with IRSA...${NC}"
eksctl create iamserviceaccount \
    --cluster=$CLUSTER_NAME \
    --region=$REGION \
    --namespace=$NAMESPACE \
    --name=$SERVICE_ACCOUNT \
    --attach-policy-arn=$POLICY_ARN \
    --role-name=$ROLE_NAME \
    --approve \
    --override-existing-serviceaccounts

echo -e "${GREEN}IRSA setup completed successfully!${NC}"
echo ""
echo -e "${YELLOW}Next steps:${NC}"
echo "1. Install the Helm chart:"
echo "   helm install helm-s3-exporter ./charts/helm-s3-exporter \\"
echo "     --namespace $NAMESPACE \\"
echo "     --set s3.bucket=$S3_BUCKET \\"
echo "     --set s3.region=$REGION \\"
echo "     --set serviceAccount.create=false \\"
echo "     --set serviceAccount.name=$SERVICE_ACCOUNT"
echo ""
echo "2. Verify the deployment:"
echo "   kubectl get pods -n $NAMESPACE"
echo "   kubectl logs -n $NAMESPACE -l app.kubernetes.io/name=helm-s3-exporter"

