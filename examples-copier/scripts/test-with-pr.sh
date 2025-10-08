#!/bin/bash

# Test webhook with real PR data
# Usage: ./scripts/test-with-pr.sh <pr-number> [owner] [repo]

set -e

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Default values from environment or config
PR_NUMBER=${1:-}
OWNER=${2:-${REPO_OWNER:-}}
REPO=${3:-${REPO_NAME:-}}
WEBHOOK_URL=${WEBHOOK_URL:-http://localhost:8080/webhook}
WEBHOOK_SECRET=${WEBHOOK_SECRET:-}

# Help message
if [ "$1" == "-h" ] || [ "$1" == "--help" ] || [ -z "$PR_NUMBER" ]; then
    echo "Test Webhook with Real PR Data"
    echo ""
    echo "Usage: $0 <pr-number> [owner] [repo]"
    echo ""
    echo "Arguments:"
    echo "  pr-number    PR number to test with (required)"
    echo "  owner        Repository owner (default: \$REPO_OWNER)"
    echo "  repo         Repository name (default: \$REPO_NAME)"
    echo ""
    echo "Environment Variables:"
    echo "  GITHUB_TOKEN      GitHub token for API access (required)"
    echo "  WEBHOOK_URL       Webhook endpoint (default: http://localhost:8080/webhook)"
    echo "  WEBHOOK_SECRET    Webhook secret for signature"
    echo "  REPO_OWNER        Default repository owner"
    echo "  REPO_NAME         Default repository name"
    echo ""
    echo "Examples:"
    echo "  # Test with PR #123 (uses REPO_OWNER and REPO_NAME from env)"
    echo "  $0 123"
    echo ""
    echo "  # Test with specific repo"
    echo "  $0 123 myorg myrepo"
    echo ""
    echo "  # Test against production"
    echo "  WEBHOOK_URL=https://myapp.appspot.com/webhook $0 123"
    echo ""
    exit 0
fi

# Validate inputs
if [ -z "$OWNER" ]; then
    echo -e "${RED}Error: Repository owner not specified${NC}"
    echo "Set REPO_OWNER environment variable or pass as second argument"
    exit 1
fi

if [ -z "$REPO" ]; then
    echo -e "${RED}Error: Repository name not specified${NC}"
    echo "Set REPO_NAME environment variable or pass as third argument"
    exit 1
fi

if [ -z "$GITHUB_TOKEN" ]; then
    echo -e "${RED}Error: GITHUB_TOKEN environment variable not set${NC}"
    echo "Get a token from: https://github.com/settings/tokens"
    exit 1
fi

# Build test-webhook tool if needed
if [ ! -f "./test-webhook" ]; then
    echo -e "${BLUE}Building test-webhook tool...${NC}"
    go build -o test-webhook ./cmd/test-webhook
    echo -e "${GREEN}✓ Built test-webhook${NC}"
fi

# Check if app is running (if testing locally)
if [[ "$WEBHOOK_URL" == http://localhost* ]]; then
    echo -e "${BLUE}Checking if application is running...${NC}"
    if ! curl -s -f "$WEBHOOK_URL" > /dev/null 2>&1; then
        echo -e "${YELLOW}Warning: Application doesn't seem to be running at $WEBHOOK_URL${NC}"
        echo -e "${YELLOW}Start it with: DRY_RUN=true ./examples-copier${NC}"
        read -p "Continue anyway? (y/N) " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            exit 1
        fi
    else
        echo -e "${GREEN}✓ Application is running${NC}"
    fi
fi

# Fetch and display PR info
echo -e "${BLUE}Fetching PR #$PR_NUMBER from $OWNER/$REPO...${NC}"

PR_DATA=$(curl -s -H "Authorization: Bearer $GITHUB_TOKEN" \
    "https://api.github.com/repos/$OWNER/$REPO/pulls/$PR_NUMBER")

PR_TITLE=$(echo "$PR_DATA" | grep -o '"title": *"[^"]*"' | head -1 | sed 's/"title": *"\(.*\)"/\1/')
PR_STATE=$(echo "$PR_DATA" | grep -o '"state": *"[^"]*"' | head -1 | sed 's/"state": *"\(.*\)"/\1/')
PR_MERGED=$(echo "$PR_DATA" | grep -o '"merged": *[^,]*' | head -1 | sed 's/"merged": *\(.*\)/\1/')

echo -e "${GREEN}✓ Found PR #$PR_NUMBER${NC}"
echo -e "  Title: $PR_TITLE"
echo -e "  State: $PR_STATE"
echo -e "  Merged: $PR_MERGED"

# Ask for confirmation
echo ""
echo -e "${YELLOW}Ready to send webhook:${NC}"
echo -e "  PR: #$PR_NUMBER ($OWNER/$REPO)"
echo -e "  URL: $WEBHOOK_URL"
if [ -n "$WEBHOOK_SECRET" ]; then
    echo -e "  Secret: ${WEBHOOK_SECRET:0:10}... (configured)"
else
    echo -e "  Secret: ${YELLOW}(none - signature will not be verified)${NC}"
fi
echo ""
read -p "Send webhook? (y/N) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Cancelled"
    exit 0
fi

# Send webhook
echo -e "${BLUE}Sending webhook...${NC}"

SECRET_FLAG=""
if [ -n "$WEBHOOK_SECRET" ]; then
    SECRET_FLAG="-secret $WEBHOOK_SECRET"
fi

./test-webhook \
    -pr "$PR_NUMBER" \
    -owner "$OWNER" \
    -repo "$REPO" \
    -url "$WEBHOOK_URL" \
    $SECRET_FLAG

echo ""
echo -e "${GREEN}✓ Test complete!${NC}"
echo ""
echo "Check application logs for processing details:"
echo "  Local: Check terminal output"
echo "  GCP: gcloud app logs tail -s default"

