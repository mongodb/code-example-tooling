#!/bin/bash

# Test if the GitHub App can access the configured repository
# This script checks the recent logs for 401 errors

set -e

echo "🔍 Testing GitHub Repository Access"
echo "===================================="
echo ""

# Get configuration from env.yaml
REPO_OWNER=$(grep "REPO_OWNER:" env.yaml | grep -v "#" | awk '{print $2}' | tr -d '"')
REPO_NAME=$(grep "REPO_NAME:" env.yaml | grep -v "#" | awk '{print $2}' | tr -d '"')
INSTALLATION_ID=$(grep "INSTALLATION_ID:" env.yaml | grep -v "#" | awk '{print $2}' | tr -d '"')

echo "Configuration:"
echo "  Repository: $REPO_OWNER/$REPO_NAME"
echo "  Installation ID: $INSTALLATION_ID"
echo ""

# Check health endpoint
echo "📊 Checking application health..."
HEALTH=$(curl -s https://github-copy-code-examples.ue.r.appspot.com/health)
AUTH_STATUS=$(echo "$HEALTH" | python3 -c "import sys, json; print(json.load(sys.stdin)['github']['authenticated'])")

if [ "$AUTH_STATUS" == "True" ]; then
    echo "✅ GitHub authentication is working"
else
    echo "❌ GitHub authentication is NOT working"
    exit 1
fi
echo ""

# Check recent logs for 401 errors
echo "🔍 Checking recent logs for 401 errors..."
RECENT_ERRORS=$(gcloud logging read "resource.type=gae_app AND severity>=ERROR AND textPayload=~'401 Bad credentials'" --limit=5 --format="value(timestamp,textPayload)" --freshness=30m 2>/dev/null)

if [ -z "$RECENT_ERRORS" ]; then
    echo "✅ No recent 401 errors found!"
    echo ""
    echo "🎉 GitHub App can successfully access the repository!"
else
    echo "❌ Found recent 401 errors:"
    echo ""
    echo "$RECENT_ERRORS"
    echo ""
    echo "This means the GitHub App cannot access one or more repositories."
    echo ""
    echo "Possible causes:"
    echo "1. GitHub App is not installed on the repository"
    echo "2. Installation ID doesn't match the repository"
    echo "3. GitHub App doesn't have 'Contents' read permission"
    echo ""
    echo "To fix:"
    echo "1. Go to: https://github.com/settings/installations"
    echo "2. Find your GitHub App installation"
    echo "3. Make sure $REPO_OWNER/$REPO_NAME is in the list of accessible repositories"
    echo "4. If not, click 'Configure' and add it"
fi

echo ""
echo "📋 Summary"
echo "=========="
echo "Repository: $REPO_OWNER/$REPO_NAME"
echo "Installation ID: $INSTALLATION_ID"
echo "Authentication: $AUTH_STATUS"

if [ -z "$RECENT_ERRORS" ]; then
    echo "Status: ✅ WORKING"
    exit 0
else
    echo "Status: ❌ NEEDS ATTENTION"
    exit 1
fi

