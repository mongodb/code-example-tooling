#!/bin/bash
# Convert .env file to env.yaml format for Google Cloud App Engine

set -e

# Default input/output files
INPUT_FILE="${1:-.env}"
OUTPUT_FILE="${2:-env.yaml}"

if [ ! -f "$INPUT_FILE" ]; then
    echo "Error: Input file '$INPUT_FILE' not found"
    echo "Usage: $0 [input-file] [output-file]"
    echo "Example: $0 .env.production env.yaml"
    exit 1
fi

echo "Converting $INPUT_FILE to $OUTPUT_FILE..."

# Start the YAML file
echo "env_variables:" > "$OUTPUT_FILE"

# Read the .env file and convert to YAML
while IFS= read -r line || [ -n "$line" ]; do
    # Skip empty lines and comments
    if [[ -z "$line" ]] || [[ "$line" =~ ^[[:space:]]*# ]]; then
        continue
    fi
    
    # Extract key and value
    if [[ "$line" =~ ^([^=]+)=(.*)$ ]]; then
        key="${BASH_REMATCH[1]}"
        value="${BASH_REMATCH[2]}"
        
        # Remove leading/trailing whitespace from key
        key=$(echo "$key" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')
        
        # Remove quotes from value if present
        value=$(echo "$value" | sed 's/^["'\'']\(.*\)["'\'']$/\1/')
        
        # Write to YAML file with proper indentation
        echo "  $key: \"$value\"" >> "$OUTPUT_FILE"
    fi
done < "$INPUT_FILE"

echo "✅ Conversion complete: $OUTPUT_FILE"
echo ""
echo "⚠️  IMPORTANT: Review $OUTPUT_FILE before deploying!"
echo "   - Verify all values are correct"
echo "   - Check for sensitive data"
echo "   - Ensure $OUTPUT_FILE is in .gitignore"

