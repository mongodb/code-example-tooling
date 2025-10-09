#!/bin/bash

# Deployment script for examples-copier to Google Cloud App Engine
# Usage: ./deploy.sh [options]
#
# Options:
#   --project PROJECT_ID    Set GCP project ID
#   --version VERSION       Set version name (default: auto-generated)
#   --no-promote           Deploy without promoting to receive traffic
#   --quiet                Skip confirmation prompts
#   --env-file FILE        Path to env.yaml file (default: env.yaml)
#   --help                 Show this help message

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
PROJECT_ID=""
VERSION=""
PROMOTE="true"
QUIET="false"
ENV_FILE="env.yaml"

# Parse command line arguments
while [[ $# -gt 0 ]]; do
  case $1 in
    --project)
      PROJECT_ID="$2"
      shift 2
      ;;
    --version)
      VERSION="$2"
      shift 2
      ;;
    --no-promote)
      PROMOTE="false"
      shift
      ;;
    --quiet)
      QUIET="true"
      shift
      ;;
    --env-file)
      ENV_FILE="$2"
      shift 2
      ;;
    --help)
      echo "Usage: ./deploy.sh [options]"
      echo ""
      echo "Options:"
      echo "  --project PROJECT_ID    Set GCP project ID"
      echo "  --version VERSION       Set version name (default: auto-generated)"
      echo "  --no-promote           Deploy without promoting to receive traffic"
      echo "  --quiet                Skip confirmation prompts"
      echo "  --env-file FILE        Path to env.yaml file (default: env.yaml)"
      echo "  --help                 Show this help message"
      exit 0
      ;;
    *)
      echo -e "${RED}Unknown option: $1${NC}"
      echo "Use --help for usage information"
      exit 1
      ;;
  esac
done

# Function to print colored messages
print_info() {
  echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
  echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
  echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
  echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check if command exists
command_exists() {
  command -v "$1" >/dev/null 2>&1
}

# Check prerequisites
print_info "Checking prerequisites..."

# Check if gcloud is installed
if ! command_exists gcloud; then
  print_error "gcloud CLI is not installed"
  echo "Install from: https://cloud.google.com/sdk/docs/install"
  exit 1
fi

# Check if go is installed
if ! command_exists go; then
  print_error "Go is not installed"
  echo "Install from: https://golang.org/dl/"
  exit 1
fi

print_success "Prerequisites check passed"

# Get current project if not specified
if [ -z "$PROJECT_ID" ]; then
  PROJECT_ID=$(gcloud config get-value project 2>/dev/null)
  if [ -z "$PROJECT_ID" ]; then
    print_error "No GCP project configured"
    echo "Set project with: gcloud config set project PROJECT_ID"
    echo "Or use: ./deploy.sh --project PROJECT_ID"
    exit 1
  fi
fi

print_info "Using GCP project: $PROJECT_ID"

# Check if env.yaml exists
if [ ! -f "$ENV_FILE" ]; then
  print_error "Environment file not found: $ENV_FILE"
  echo ""
  echo "Create $ENV_FILE with required environment variables:"
  echo ""
  cat << 'EOF'
env_variables:
  GITHUB_APP_ID: "your-app-id"
  INSTALLATION_ID: "your-installation-id"
  REPO_NAME: "your-repo-name"
  REPO_OWNER: "your-repo-owner"
  GITHUB_APP_PRIVATE_KEY_SECRET_NAME: "projects/PROJECT_ID/secrets/SECRET_NAME/versions/latest"
  WEBHOOK_SECRET: "your-webhook-secret"
  COMMITTER_NAME: "GitHub Copier App"
  COMMITTER_EMAIL: "bot@example.com"
  CONFIG_FILE: "copier-config.yaml"
  DEPRECATION_FILE: "deprecated_examples.json"
  WEBSERVER_PATH: "/events"
EOF
  echo ""
  exit 1
fi

print_success "Environment file found: $ENV_FILE"

# Check if app.yaml exists
if [ ! -f "app.yaml" ]; then
  print_error "app.yaml not found in current directory"
  echo "Run this script from the examples-copier directory"
  exit 1
fi

# Build the application
print_info "Building application..."
if go build -o examples-copier .; then
  print_success "Build successful"
else
  print_error "Build failed"
  exit 1
fi

# Run tests
print_info "Running tests..."
if go test ./... -v; then
  print_success "All tests passed"
else
  print_warning "Some tests failed - continuing anyway"
fi

# Show deployment summary
echo ""
echo "========================================="
echo "Deployment Summary"
echo "========================================="
echo "Project:      $PROJECT_ID"
echo "Version:      ${VERSION:-auto-generated}"
echo "Promote:      $PROMOTE"
echo "Env File:     $ENV_FILE"
echo "========================================="
echo ""

# Confirm deployment
if [ "$QUIET" != "true" ]; then
  read -p "Continue with deployment? (y/N) " -n 1 -r
  echo
  if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    print_info "Deployment cancelled"
    exit 0
  fi
fi

# Build gcloud command
DEPLOY_CMD="gcloud app deploy app.yaml --env-vars-file=$ENV_FILE --project=$PROJECT_ID"

if [ -n "$VERSION" ]; then
  DEPLOY_CMD="$DEPLOY_CMD --version=$VERSION"
fi

if [ "$PROMOTE" != "true" ]; then
  DEPLOY_CMD="$DEPLOY_CMD --no-promote"
fi

if [ "$QUIET" = "true" ]; then
  DEPLOY_CMD="$DEPLOY_CMD --quiet"
fi

# Deploy to App Engine
print_info "Deploying to App Engine..."
echo "Command: $DEPLOY_CMD"
echo ""

if eval "$DEPLOY_CMD"; then
  print_success "Deployment successful!"
  echo ""
  
  # Get app URL
  APP_URL=$(gcloud app describe --project=$PROJECT_ID --format="value(defaultHostname)" 2>/dev/null)
  if [ -n "$APP_URL" ]; then
    print_info "Application URL: https://$APP_URL"
    print_info "Webhook URL: https://$APP_URL/events"
  fi
  
  echo ""
  print_info "Next steps:"
  echo "  1. Update GitHub webhook URL to: https://$APP_URL/events"
  echo "  2. Verify webhook secret matches WEBHOOK_SECRET in env.yaml"
  echo "  3. Test webhook by merging a PR in source repository"
  echo "  4. Monitor logs: gcloud app logs tail -s default --project=$PROJECT_ID"
  echo ""
  
  # Ask if user wants to view logs
  if [ "$QUIET" != "true" ]; then
    read -p "View application logs? (y/N) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
      gcloud app logs tail -s default --project=$PROJECT_ID
    fi
  fi
else
  print_error "Deployment failed"
  echo ""
  echo "Troubleshooting:"
  echo "  1. Check that all required APIs are enabled:"
  echo "     gcloud services enable appengine.googleapis.com --project=$PROJECT_ID"
  echo "     gcloud services enable secretmanager.googleapis.com --project=$PROJECT_ID"
  echo "  2. Verify env.yaml contains all required variables"
  echo "  3. Check deployment logs for specific errors"
  echo "  4. See DEPLOYMENT-GUIDE.md for detailed troubleshooting"
  exit 1
fi

