#!/bin/bash

# Convert between App Engine (env.yaml) and Cloud Run (env-cloudrun.yaml) formats
#
# Usage:
#   ./convert-env-format.sh to-cloudrun env.yaml env-cloudrun.yaml
#   ./convert-env-format.sh to-appengine env-cloudrun.yaml env.yaml

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Print usage
usage() {
    echo "Convert between App Engine and Cloud Run environment file formats"
    echo ""
    echo "Usage:"
    echo "  $0 to-cloudrun <input-env.yaml> <output-env-cloudrun.yaml>"
    echo "  $0 to-appengine <input-env-cloudrun.yaml> <output-env.yaml>"
    echo ""
    echo "Examples:"
    echo "  # Convert App Engine format to Cloud Run format"
    echo "  $0 to-cloudrun env.yaml env-cloudrun.yaml"
    echo ""
    echo "  # Convert Cloud Run format to App Engine format"
    echo "  $0 to-appengine env-cloudrun.yaml env.yaml"
    echo ""
    echo "Formats:"
    echo "  App Engine:   env_variables: wrapper with indented keys"
    echo "  Cloud Run:    Plain YAML without wrapper"
    exit 1
}

# Check arguments
if [ $# -ne 3 ]; then
    usage
fi

COMMAND=$1
INPUT=$2
OUTPUT=$3

# Validate command
if [ "$COMMAND" != "to-cloudrun" ] && [ "$COMMAND" != "to-appengine" ]; then
    echo -e "${RED}Error: Invalid command '$COMMAND'${NC}"
    echo "Must be 'to-cloudrun' or 'to-appengine'"
    usage
fi

# Check input file exists
if [ ! -f "$INPUT" ]; then
    echo -e "${RED}Error: Input file '$INPUT' not found${NC}"
    exit 1
fi

# Check if output file exists
if [ -f "$OUTPUT" ]; then
    echo -e "${YELLOW}Warning: Output file '$OUTPUT' already exists${NC}"
    read -p "Overwrite? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Aborted"
        exit 1
    fi
fi

# Convert to Cloud Run format (remove env_variables wrapper and unindent)
if [ "$COMMAND" = "to-cloudrun" ]; then
    echo -e "${BLUE}Converting App Engine format to Cloud Run format...${NC}"
    
    # Remove 'env_variables:' line and unindent by 2 spaces
    sed '/^env_variables:/d' "$INPUT" | sed 's/^  //' > "$OUTPUT"
    
    echo -e "${GREEN}✓ Converted to Cloud Run format: $OUTPUT${NC}"
    echo ""
    echo "Deploy with:"
    echo "  gcloud run deploy examples-copier --source . --env-vars-file=$OUTPUT"
fi

# Convert to App Engine format (add env_variables wrapper and indent)
if [ "$COMMAND" = "to-appengine" ]; then
    echo -e "${BLUE}Converting Cloud Run format to App Engine format...${NC}"
    
    # Add 'env_variables:' header and indent all lines by 2 spaces
    echo "env_variables:" > "$OUTPUT"
    sed 's/^/  /' "$INPUT" >> "$OUTPUT"
    
    echo -e "${GREEN}✓ Converted to App Engine format: $OUTPUT${NC}"
    echo ""
    echo "Deploy with:"
    echo "  gcloud app deploy app.yaml  # Includes $OUTPUT automatically"
fi

echo ""
echo -e "${YELLOW}Note: Review the output file before deploying!${NC}"

