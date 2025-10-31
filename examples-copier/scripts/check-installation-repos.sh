#!/bin/bash

# Script to check which repositories a GitHub App installation has access to
# Requires: curl, jq, gcloud

set -e

echo "🔍 Checking GitHub App Installation Repositories"
echo "================================================"
echo ""

# Check if jq is installed
if ! command -v jq &> /dev/null; then
    echo "❌ jq is required but not installed"
    echo "Install: brew install jq"
    exit 1
fi

# Get the installation ID from env.yaml
INSTALLATION_ID=$(grep "INSTALLATION_ID:" env.yaml | grep -v "#" | awk '{print $2}' | tr -d '"')

if [ -z "$INSTALLATION_ID" ]; then
    echo "❌ INSTALLATION_ID not found in env.yaml"
    exit 1
fi

echo "Installation ID: $INSTALLATION_ID"
echo ""

# Get the GitHub App private key from Secret Manager
echo "📥 Retrieving GitHub App private key from Secret Manager..."
PEM_KEY=$(gcloud secrets versions access latest --secret=CODE_COPIER_PEM)

if [ -z "$PEM_KEY" ]; then
    echo "❌ Failed to retrieve private key"
    exit 1
fi

echo "✅ Private key retrieved"
echo ""

# Get the GitHub App ID from env.yaml
APP_ID=$(grep "GITHUB_APP_ID:" env.yaml | awk '{print $2}' | tr -d '"')

if [ -z "$APP_ID" ]; then
    echo "❌ GITHUB_APP_ID not found in env.yaml"
    exit 1
fi

echo "GitHub App ID: $APP_ID"
echo ""

# Generate JWT token (simplified - requires ruby)
echo "🔐 Generating JWT token..."

# Save PEM key to temp file
TMP_PEM=$(mktemp)
echo "$PEM_KEY" > "$TMP_PEM"

# Generate JWT using ruby
JWT=$(ruby -rjwt -rjson -e "
  private_key = OpenSSL::PKey::RSA.new(File.read('$TMP_PEM'))
  payload = {
    iat: Time.now.to_i - 60,
    exp: Time.now.to_i + (10 * 60),
    iss: '$APP_ID'
  }
  puts JWT.encode(payload, private_key, 'RS256')
" 2>/dev/null)

# Clean up temp file
rm -f "$TMP_PEM"

if [ -z "$JWT" ]; then
    echo "❌ Failed to generate JWT token"
    echo "Note: This script requires ruby with jwt gem installed"
    echo "Install: gem install jwt"
    exit 1
fi

echo "✅ JWT token generated"
echo ""

# Get installation access token
echo "🔑 Getting installation access token..."
INSTALL_TOKEN_RESPONSE=$(curl -s -X POST \
  -H "Authorization: Bearer $JWT" \
  -H "Accept: application/vnd.github+json" \
  "https://api.github.com/app/installations/$INSTALLATION_ID/access_tokens")

INSTALL_TOKEN=$(echo "$INSTALL_TOKEN_RESPONSE" | jq -r '.token')

if [ "$INSTALL_TOKEN" == "null" ] || [ -z "$INSTALL_TOKEN" ]; then
    echo "❌ Failed to get installation access token"
    echo "Response:"
    echo "$INSTALL_TOKEN_RESPONSE" | jq .
    exit 1
fi

echo "✅ Installation access token obtained"
echo ""

# Get installation details
echo "📋 Installation Details:"
echo "------------------------"
INSTALL_INFO=$(curl -s \
  -H "Authorization: Bearer $JWT" \
  -H "Accept: application/vnd.github+json" \
  "https://api.github.com/app/installations/$INSTALLATION_ID")

ACCOUNT=$(echo "$INSTALL_INFO" | jq -r '.account.login')
ACCOUNT_TYPE=$(echo "$INSTALL_INFO" | jq -r '.account.type')
REPO_SELECTION=$(echo "$INSTALL_INFO" | jq -r '.repository_selection')

echo "Account: $ACCOUNT ($ACCOUNT_TYPE)"
echo "Repository Selection: $REPO_SELECTION"
echo ""

# Get list of repositories
echo "📚 Accessible Repositories:"
echo "---------------------------"

if [ "$REPO_SELECTION" == "all" ]; then
    echo "✅ Installation has access to ALL repositories in $ACCOUNT"
    echo ""
    echo "Fetching repository list..."
    REPOS=$(curl -s \
      -H "Authorization: token $INSTALL_TOKEN" \
      -H "Accept: application/vnd.github+json" \
      "https://api.github.com/installation/repositories?per_page=100")
    
    echo "$REPOS" | jq -r '.repositories[] | "  - \(.full_name)"'
    
    TOTAL=$(echo "$REPOS" | jq -r '.total_count')
    echo ""
    echo "Total: $TOTAL repositories"
else
    echo "✅ Installation has access to SELECTED repositories"
    echo ""
    REPOS=$(curl -s \
      -H "Authorization: token $INSTALL_TOKEN" \
      -H "Accept: application/vnd.github+json" \
      "https://api.github.com/installation/repositories?per_page=100")
    
    echo "$REPOS" | jq -r '.repositories[] | "  - \(.full_name)"'
    
    TOTAL=$(echo "$REPOS" | jq -r '.total_count')
    echo ""
    echo "Total: $TOTAL repositories"
fi

echo ""
echo "✅ Done!"

