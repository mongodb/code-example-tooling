# Helper Scripts

Collection of helper scripts for testing and running the examples-copier application.

## Scripts

### run-local.sh

Start the examples-copier application locally with proper environment configuration.

**Usage:**
```bash
./scripts/run-local.sh
```

**What it does:**
- Loads environment variables from `configs/.env.local`
- Disables Google Cloud logging (uses stdout instead)
- Sets up local configuration
- Starts the application

**Example:**
```bash
# Start app locally
./scripts/run-local.sh

# In another terminal, test it
./test-webhook -payload test-payloads/example-pr-merged.json
```

**Environment:**
- Uses `configs/.env.local` for configuration
- Sets `COPIER_DISABLE_CLOUD_LOGGING=true`
- Sets `CONFIG_FILE=copier-config.yaml`

### test-and-check.sh

Send a test webhook and check the metrics.

**Usage:**
```bash
./scripts/test-and-check.sh
```

**What it does:**
1. Sends test webhook with example payload
2. Waits for processing
3. Fetches and displays metrics
4. Shows recent application logs

**Example:**
```bash
# Start app first
./scripts/run-local.sh

# In another terminal, test and check
./scripts/test-and-check.sh
```

**Output:**
```
Testing webhook with example payload...

✓ Loaded payload from test-payloads/example-pr-merged.json
✓ Response: 200 OK
✓ Webhook sent successfully

Webhook sent! Waiting 2 seconds for processing...

=== Metrics ===
{
  "webhooks": {
    "received": 1,
    "processed": 1,
    "failed": 0
  },
  "files": {
    "matched": 20,
    "uploaded": 0
  }
}

=== Recent Logs ===
[INFO] loaded config from local file
[INFO] retrieved changed files | {"count":21}
[INFO] processing files with pattern matching
[INFO] file matched pattern | {"file":"..."}
```

### test-slack.sh

Test Slack notifications by sending example messages.

**Usage:**
```bash
./scripts/test-slack.sh [webhook-url]
```

**Arguments:**
- `webhook-url` - Slack webhook URL (optional, uses `$SLACK_WEBHOOK_URL` if not provided)

**What it does:**
1. Sends simple test message
2. Sends PR processed notification
3. Sends error notification
4. Sends files copied notification
5. Sends deprecation notification

**Example:**
```bash
# Using environment variable
export SLACK_WEBHOOK_URL="https://hooks.slack.com/services/..."
./scripts/test-slack.sh

# Or pass URL directly
./scripts/test-slack.sh "https://hooks.slack.com/services/..."
```

**Output:**
```
Testing Slack Notifications

Webhook URL: https://hooks.slack.com/services/...

Test 1: Sending simple test message...
✓ Simple message sent

Test 2: Sending PR processed notification...
✓ PR processed notification sent

Test 3: Sending error notification...
✓ Error notification sent

Test 4: Sending files copied notification...
✓ Files copied notification sent

Test 5: Sending deprecation notification...
✓ Deprecation notification sent

=== All Tests Complete ===

Check your Slack channel for 5 test notifications
```

### test-with-pr.sh

Fetch real PR data from GitHub and send it to the webhook.

**Usage:**
```bash
./scripts/test-with-pr.sh <pr-number> <owner> <repo> [webhook-url]
```

**Arguments:**
- `pr-number` - Pull request number (required)
- `owner` - Repository owner (required)
- `repo` - Repository name (required)
- `webhook-url` - Webhook URL (optional, default: `http://localhost:8080/webhook`)

**Environment Variables:**
- `GITHUB_TOKEN` - GitHub personal access token (required)

**Example:**
```bash
# Set GitHub token
export GITHUB_TOKEN=ghp_your_token_here

# Test with real PR
./scripts/test-with-pr.sh 42 mongodb docs-code-examples

# Test with custom webhook URL
./scripts/test-with-pr.sh 42 mongodb docs-code-examples http://localhost:8080/webhook
```

**Output:**
```
Fetching PR #42 from mongodb/docs-code-examples...

✓ PR fetched successfully
✓ PR Title: Add Go database examples
✓ Files changed: 21
✓ Sending to webhook...
✓ Response: 200 OK

Check application logs for processing details.
```

## Common Workflows

### Local Development

```bash
# 1. Start app locally
./scripts/run-local.sh

# 2. In another terminal, test it
./scripts/test-and-check.sh

# 3. Check metrics
curl http://localhost:8080/metrics | jq
```

### Testing with Real Data

```bash
# 1. Start app
./scripts/run-local.sh

# 2. Set GitHub token
export GITHUB_TOKEN=ghp_...

# 3. Test with real PR
./scripts/test-with-pr.sh 42 myorg myrepo

# 4. Check results
./scripts/test-and-check.sh
```

### Testing Slack Integration

```bash
# 1. Test Slack webhook
export SLACK_WEBHOOK_URL="https://hooks.slack.com/services/..."
./scripts/test-slack.sh

# 2. Start app with Slack enabled
./scripts/run-local.sh

# 3. Send test webhook
./scripts/test-and-check.sh

# 4. Check Slack channel for notification
```

### Dry-Run Testing

```bash
# 1. Start app in dry-run mode
DRY_RUN=true ./scripts/run-local.sh

# 2. Test processing
./scripts/test-and-check.sh

# 3. Verify no commits were made (check logs)
```

## Script Details

### Environment Variables

All scripts respect these environment variables:

**Application:**
- `CONFIG_FILE` - Configuration file path
- `DRY_RUN` - Enable dry-run mode
- `LOG_LEVEL` - Logging level (debug, info, warn, error)
- `COPIER_DISABLE_CLOUD_LOGGING` - Disable Google Cloud logging

**GitHub:**
- `GITHUB_TOKEN` - GitHub personal access token
- `GITHUB_APP_ID` - GitHub App ID
- `GITHUB_INSTALLATION_ID` - GitHub Installation ID

**Slack:**
- `SLACK_WEBHOOK_URL` - Slack webhook URL
- `SLACK_CHANNEL` - Slack channel
- `SLACK_ENABLED` - Enable/disable Slack notifications

**MongoDB:**
- `MONGO_URI` - MongoDB connection string
- `AUDIT_ENABLED` - Enable/disable audit logging

### Exit Codes

All scripts use standard exit codes:
- `0` - Success
- `1` - Error occurred

### Error Handling

Scripts include error handling and will:
- Display clear error messages
- Exit with non-zero code on failure
- Provide troubleshooting hints

## Creating Custom Scripts

### Template

```bash
#!/bin/bash

# Script description
# Usage: ./my-script.sh [args]

set -e  # Exit on error

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Check prerequisites
if [ -z "$REQUIRED_VAR" ]; then
    echo -e "${RED}Error: REQUIRED_VAR not set${NC}"
    exit 1
fi

# Main logic
echo -e "${GREEN}Starting...${NC}"

# Do work
# ...

echo -e "${GREEN}Complete!${NC}"
```

### Best Practices

1. **Use `set -e`** to exit on errors
2. **Check prerequisites** before running
3. **Provide clear output** with colors
4. **Include usage instructions** in comments
5. **Handle errors gracefully**
6. **Make scripts executable**: `chmod +x script.sh`

## Troubleshooting

### Script Not Executable

**Error:**
```
Permission denied: ./scripts/run-local.sh
```

**Solution:**
```bash
chmod +x scripts/*.sh
```

### Environment Variables Not Set

**Error:**
```
Error: GITHUB_TOKEN not set
```

**Solution:**
```bash
export GITHUB_TOKEN=ghp_your_token_here
# Or add to configs/.env.local
```

### App Not Running

**Error:**
```
Connection refused
```

**Solution:**
```bash
# Start the app first
./scripts/run-local.sh

# Then run tests in another terminal
```

### Slack Webhook Fails

**Error:**
```
Error: invalid_payload
```

**Solution:**
```bash
# Verify webhook URL
echo $SLACK_WEBHOOK_URL

# Test directly
curl -X POST -H 'Content-type: application/json' \
  --data '{"text":"Test"}' \
  "$SLACK_WEBHOOK_URL"
```

## See Also

- [Local Testing Guide](../docs/LOCAL-TESTING.md) - Local development
- [Webhook Testing Guide](../docs/WEBHOOK-TESTING.md) - Testing webhooks
- [Quick Reference](../docs/QUICK-REFERENCE.md) - All commands
- [test-webhook Tool](../cmd/test-webhook/README.md) - Test webhook tool

