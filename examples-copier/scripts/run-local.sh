#!/bin/bash

# Run examples-copier locally with proper development settings
# This script sets up the environment for local testing

set -e

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${BLUE}Starting examples-copier in local development mode${NC}"
echo ""

# Check if binary exists
if [ ! -f "./examples-copier" ]; then
    echo -e "${YELLOW}Building examples-copier...${NC}"
    go build -o examples-copier .
    echo -e "${GREEN}✓ Built examples-copier${NC}"
fi

# Set local development environment
export COPIER_DISABLE_CLOUD_LOGGING=true
export DRY_RUN=true
export LOG_LEVEL=debug
export COPIER_DEBUG=true
export METRICS_ENABLED=true
export PORT=8080
export AUDIT_ENABLED=false

# Use config.json by default (or override)
export CONFIG_FILE=${CONFIG_FILE:-config.json}
export DEPRECATION_FILE=${DEPRECATION_FILE:-deprecated_examples.json}

# Load .env if it exists
if [ -f "configs/.env" ]; then
    echo -e "${BLUE}Loading configs/.env${NC}"
    set -a
    source configs/.env
    set +a
fi

# Show configuration
echo -e "${GREEN}Configuration:${NC}"
echo "  Dry Run:       ${DRY_RUN}"
echo "  Cloud Logging: ${COPIER_DISABLE_CLOUD_LOGGING}"
echo "  Audit Log:     ${AUDIT_ENABLED}"
echo "  Config File:   ${CONFIG_FILE}"
echo "  Port:          ${PORT}"
echo ""

# Check for GitHub token if needed for testing
if [ -z "$GITHUB_TOKEN" ] && [ -z "$GITHUB_APP_ID" ]; then
    echo -e "${YELLOW}Warning: No GitHub credentials set${NC}"
    echo "  For webhook testing with real PRs, set GITHUB_TOKEN"
    echo "  Get token from: https://github.com/settings/tokens"
    echo ""
fi

echo -e "${GREEN}Starting application...${NC}"
echo -e "${BLUE}Press Ctrl+C to stop${NC}"
echo ""

# Run the application
./examples-copier

