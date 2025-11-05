#!/bin/bash
# Deploy examples-copier to Google Cloud Run
# Usage: ./scripts/deploy-cloudrun.sh [region]

set -e

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# Get the project root (parent of scripts directory)
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Configuration
SERVICE_NAME="examples-copier"
REGION="${1:-us-central1}"
ENV_FILE="$PROJECT_ROOT/env-cloudrun.yaml"

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${BLUE}╔════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║         Deploying examples-copier to Cloud Run                ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════════════════════════╝${NC}"
echo ""

# Check if env-cloudrun.yaml exists
if [ ! -f "$ENV_FILE" ]; then
    echo -e "${YELLOW}⚠️  Warning: $ENV_FILE not found${NC}"
    echo "Create it from a template:"
    echo "  cp configs/env.yaml.production $ENV_FILE"
    echo "  # Edit with your values"
    exit 1
fi

# Get current project
PROJECT=$(gcloud config get-value project 2>/dev/null)
if [ -z "$PROJECT" ]; then
    echo -e "${YELLOW}⚠️  No Google Cloud project set${NC}"
    echo "Set your project:"
    echo "  gcloud config set project YOUR_PROJECT_ID"
    exit 1
fi

echo -e "${GREEN}📦 Configuration:${NC}"
echo "   Service:  $SERVICE_NAME"
echo "   Region:   $REGION"
echo "   Project:  $PROJECT"
echo "   Env File: $ENV_FILE"
echo ""

# Confirm deployment
read -p "Deploy to Cloud Run? (y/N) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Deployment cancelled"
    exit 0
fi

echo ""
echo -e "${BLUE}🚀 Deploying...${NC}"
echo ""

# Change to project root for deployment
cd "$PROJECT_ROOT"

# Deploy to Cloud Run using Dockerfile
# Note: Using --source with Dockerfile to ensure it uses Docker build, not buildpacks
gcloud run deploy "$SERVICE_NAME" \
  --source . \
  --region "$REGION" \
  --env-vars-file="$ENV_FILE" \
  --allow-unauthenticated \
  --max-instances=10 \
  --cpu=1 \
  --memory=512Mi \
  --timeout=300s \
  --concurrency=80 \
  --port=8080 \
  --platform=managed

echo ""
echo -e "${GREEN}✅ Deployment complete!${NC}"
echo ""

# Get service URL
SERVICE_URL=$(gcloud run services describe "$SERVICE_NAME" \
  --region="$REGION" \
  --format="value(status.url)" 2>/dev/null)

if [ -n "$SERVICE_URL" ]; then
    echo -e "${GREEN}🌐 Service URL:${NC}"
    echo "   $SERVICE_URL"
    echo ""
    echo -e "${BLUE}📋 Next steps:${NC}"
    echo "   1. Test health endpoint:"
    echo "      curl $SERVICE_URL/health"
    echo ""
    echo "   2. View logs:"
    echo "      gcloud run services logs read $SERVICE_NAME --region=$REGION --limit=50"
    echo ""
    echo "   3. Configure GitHub webhook:"
    echo "      Payload URL: $SERVICE_URL/events"
    echo "      Secret: (from Secret Manager)"
fi

