#!/bin/bash

# Test Slack notifications
# Usage: ./scripts/test-slack.sh [webhook-url]

set -e

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Check if webhook URL is provided
WEBHOOK_URL="${1:-$SLACK_WEBHOOK_URL}"

if [ -z "$WEBHOOK_URL" ]; then
    echo -e "${RED}Error: Slack webhook URL not provided${NC}"
    echo ""
    echo "Usage:"
    echo "  ./scripts/test-slack.sh <webhook-url>"
    echo ""
    echo "Or set environment variable:"
    echo "  export SLACK_WEBHOOK_URL='https://hooks.slack.com/services/...'"
    echo "  ./scripts/test-slack.sh"
    echo ""
    echo "Get your webhook URL from:"
    echo "  Slack → Apps → Incoming Webhooks → Add to Slack"
    exit 1
fi

echo -e "${BLUE}Testing Slack Notifications${NC}"
echo ""
echo -e "${YELLOW}Webhook URL: ${WEBHOOK_URL:0:50}...${NC}"
echo ""

# Test 1: Simple message
echo -e "${BLUE}Test 1: Sending simple test message...${NC}"
curl -X POST -H 'Content-type: application/json' \
    --data '{"text":"🧪 Test message from examples-copier"}' \
    "$WEBHOOK_URL"
echo ""
echo -e "${GREEN}✓ Simple message sent${NC}"
echo ""
sleep 2

# Test 2: PR Processed notification
echo -e "${BLUE}Test 2: Sending PR processed notification...${NC}"
curl -X POST -H 'Content-type: application/json' \
    --data '{
        "username": "Examples Copier",
        "icon_emoji": ":robot_face:",
        "attachments": [
            {
                "color": "good",
                "title": "✅ PR #42 Processed",
                "title_link": "https://github.com/mongodb/docs-code-examples/pull/42",
                "text": "Add Go database examples",
                "fields": [
                    {"title": "Repository", "value": "mongodb/docs-code-examples", "short": true},
                    {"title": "Files Matched", "value": "20", "short": true},
                    {"title": "Files Copied", "value": "18", "short": true},
                    {"title": "Files Failed", "value": "2", "short": true},
                    {"title": "Processing Time", "value": "5.2s", "short": true}
                ],
                "footer": "Examples Copier",
                "footer_icon": "https://github.githubassets.com/images/modules/logos_page/GitHub-Mark.png",
                "ts": '$(date +%s)'
            }
        ]
    }' \
    "$WEBHOOK_URL"
echo ""
echo -e "${GREEN}✓ PR processed notification sent${NC}"
echo ""
sleep 2

# Test 3: Error notification
echo -e "${BLUE}Test 3: Sending error notification...${NC}"
curl -X POST -H 'Content-type: application/json' \
    --data '{
        "username": "Examples Copier",
        "icon_emoji": ":robot_face:",
        "attachments": [
            {
                "color": "danger",
                "title": "❌ Error Occurred",
                "text": "An error occurred during config_load",
                "fields": [
                    {"title": "Operation", "value": "config_load", "short": true},
                    {"title": "Error", "value": "failed to retrieve config file: 404 Not Found", "short": false},
                    {"title": "Repository", "value": "mongodb/docs-code-examples", "short": true},
                    {"title": "PR Number", "value": "#42", "short": true}
                ],
                "footer": "Examples Copier",
                "footer_icon": "https://github.githubassets.com/images/modules/logos_page/GitHub-Mark.png",
                "ts": '$(date +%s)'
            }
        ]
    }' \
    "$WEBHOOK_URL"
echo ""
echo -e "${GREEN}✓ Error notification sent${NC}"
echo ""
sleep 2

# Test 4: Files copied notification
echo -e "${BLUE}Test 4: Sending files copied notification...${NC}"
curl -X POST -H 'Content-type: application/json' \
    --data '{
        "username": "Examples Copier",
        "icon_emoji": ":robot_face:",
        "attachments": [
            {
                "color": "good",
                "title": "📋 Files Copied from PR #42",
                "text": "```\n• generated-examples/test-project/cmd/main.go\n• generated-examples/test-project/internal/auth.go\n• generated-examples/test-project/configs/config.json\n... and 7 more```",
                "fields": [
                    {"title": "Source", "value": "mongodb/docs-code-examples", "short": true},
                    {"title": "Target", "value": "mongodb/target-repo", "short": true},
                    {"title": "Rule", "value": "Copy generated examples", "short": true},
                    {"title": "File Count", "value": "10", "short": true}
                ],
                "footer": "Examples Copier",
                "footer_icon": "https://github.githubassets.com/images/modules/logos_page/GitHub-Mark.png",
                "ts": '$(date +%s)'
            }
        ]
    }' \
    "$WEBHOOK_URL"
echo ""
echo -e "${GREEN}✓ Files copied notification sent${NC}"
echo ""
sleep 2

# Test 5: Deprecation notification
echo -e "${BLUE}Test 5: Sending deprecation notification...${NC}"
curl -X POST -H 'Content-type: application/json' \
    --data '{
        "username": "Examples Copier",
        "icon_emoji": ":robot_face:",
        "attachments": [
            {
                "color": "warning",
                "title": "⚠️ Files Deprecated from PR #42",
                "text": "```\n• old-examples/deprecated.go\n• old-examples/removed.py\n```",
                "fields": [
                    {"title": "Repository", "value": "mongodb/docs-code-examples", "short": true},
                    {"title": "File Count", "value": "2", "short": true}
                ],
                "footer": "Examples Copier",
                "footer_icon": "https://github.githubassets.com/images/modules/logos_page/GitHub-Mark.png",
                "ts": '$(date +%s)'
            }
        ]
    }' \
    "$WEBHOOK_URL"
echo ""
echo -e "${GREEN}✓ Deprecation notification sent${NC}"
echo ""

echo -e "${GREEN}=== All Tests Complete ===${NC}"
echo ""
echo -e "${YELLOW}Check your Slack channel for 5 test notifications:${NC}"
echo "  1. Simple test message"
echo "  2. PR processed notification (green)"
echo "  3. Error notification (red)"
echo "  4. Files copied notification (green)"
echo "  5. Deprecation notification (yellow)"
echo ""
echo -e "${BLUE}To use Slack notifications with the app:${NC}"
echo "  export SLACK_WEBHOOK_URL='$WEBHOOK_URL'"
echo "  CONFIG_FILE=copier-config.yaml make run-local-quick"
echo ""

