#!/bin/bash

# Test webhook and check results
# Usage: ./scripts/test-and-check.sh

set -e

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${BLUE}Testing webhook with example payload...${NC}"
echo ""

# Send webhook
./test-webhook -payload test-payloads/example-pr-merged.json

echo ""
echo -e "${GREEN}Webhook sent! Waiting 2 seconds for processing...${NC}"
sleep 2

echo ""
echo -e "${BLUE}=== Metrics ===${NC}"
curl -s http://localhost:8080/metrics | jq '.'

echo ""
echo -e "${BLUE}=== Health ===${NC}"
curl -s http://localhost:8080/health | jq '.'

echo ""
echo -e "${GREEN}Test complete!${NC}"
echo ""
echo -e "${YELLOW}Check the app logs (Terminal 1) for detailed processing information${NC}"

