# test-webhook

Command-line tool for testing the examples-copier webhook endpoint with example or real PR data.

## Overview

The `test-webhook` tool helps you:
- Test webhook processing locally
- Use example payloads for testing
- Fetch real PR data from GitHub
- Debug webhook issues
- Validate configuration with real data

## Installation

```bash
cd examples-copier
go build -o test-webhook ./cmd/test-webhook
```

## Usage

### Test with Example Payload

Send a pre-made example payload to the webhook endpoint.

**Usage:**
```bash
./test-webhook -payload <file> [-url <url>]
```

**Options:**
- `-payload` - Path to JSON payload file (required)
- `-url` - Webhook URL (default: `http://localhost:8080/webhook`)

**Example:**

```bash
# Use example payload
./test-webhook -payload test-payloads/example-pr-merged.json

# Use custom URL
./test-webhook -payload test-payloads/example-pr-merged.json \
  -url http://localhost:8080/webhook
```

**Output:**
```
Testing webhook with example payload...

✓ Loaded payload from test-payloads/example-pr-merged.json
✓ Response: 200 OK
✓ Webhook sent successfully

Check application logs for processing details.
```

### Test with Real PR Data

Fetch real PR data from GitHub and send it to the webhook.

**Usage:**
```bash
./test-webhook -pr <number> -owner <owner> -repo <repo> [-url <url>]
```

**Options:**
- `-pr` - Pull request number (required)
- `-owner` - Repository owner (required)
- `-repo` - Repository name (required)
- `-url` - Webhook URL (default: `http://localhost:8080/webhook`)

**Environment Variables:**
- `GITHUB_TOKEN` - GitHub personal access token (required for real PR data)

**Example:**

```bash
# Set GitHub token
export GITHUB_TOKEN=ghp_your_token_here

# Test with real PR
./test-webhook -pr 42 -owner mongodb -repo docs-code-examples

# Test with custom URL
./test-webhook -pr 42 -owner mongodb -repo docs-code-examples \
  -url http://localhost:8080/webhook
```

**Output:**
```
Fetching PR data from GitHub...

✓ Fetched PR #42 from mongodb/docs-code-examples
✓ PR Title: Add Go database examples
✓ Files changed: 21
✓ Response: 200 OK
✓ Webhook sent successfully

Check application logs for processing details.
```

## Common Use Cases

### Local Testing

Test your configuration locally before deploying:

```bash
# 1. Start app in dry-run mode
DRY_RUN=true CONFIG_FILE=copier-config.yaml make run-local-quick

# 2. In another terminal, send test webhook
./test-webhook -payload test-payloads/example-pr-merged.json

# 3. Check logs
tail -f logs/app.log
```

### Testing Pattern Matching

Test if your patterns match real PR files:

```bash
# 1. Start app with your config
CONFIG_FILE=copier-config.yaml make run-local-quick

# 2. Send webhook with real PR data
export GITHUB_TOKEN=ghp_...
./test-webhook -pr 42 -owner myorg -repo myrepo

# 3. Check metrics
curl http://localhost:8080/metrics | jq '.files'
```

### Testing Path Transformations

Verify files are copied to correct locations:

```bash
# 1. Start app in dry-run mode
DRY_RUN=true CONFIG_FILE=copier-config.yaml ./examples-copier &

# 2. Send test webhook
./test-webhook -payload test-payloads/example-pr-merged.json

# 3. Check logs for transformed paths
grep "transformed path" logs/app.log
```

### Testing Slack Notifications

Test Slack integration:

```bash
# 1. Start app with Slack enabled
export SLACK_WEBHOOK_URL="https://hooks.slack.com/services/..."
CONFIG_FILE=copier-config.yaml ./examples-copier &

# 2. Send test webhook
./test-webhook -payload test-payloads/example-pr-merged.json

# 3. Check Slack channel for notification
```

### Debugging Webhook Issues

Debug webhook processing:

```bash
# 1. Enable debug logging
export LOG_LEVEL=debug
CONFIG_FILE=copier-config.yaml ./examples-copier &

# 2. Send test webhook
./test-webhook -payload test-payloads/example-pr-merged.json

# 3. Review detailed logs
grep "DEBUG" logs/app.log
```

## Example Payloads

The `test-payloads/` directory contains example webhook payloads:

### example-pr-merged.json

A complete merged PR payload with:
- Multiple file changes (added, modified, removed)
- Various file types and paths
- Realistic PR metadata

**Usage:**
```bash
./test-webhook -payload test-payloads/example-pr-merged.json
```

### Creating Custom Payloads

Create custom payloads for specific test scenarios:

```bash
# Copy example
cp test-payloads/example-pr-merged.json test-payloads/my-test.json

# Edit to match your test case
vim test-payloads/my-test.json

# Test with custom payload
./test-webhook -payload test-payloads/my-test.json
```

**Example custom payload:**
```json
{
  "action": "closed",
  "pull_request": {
    "number": 123,
    "merged": true,
    "merge_commit_sha": "abc123",
    "head": {
      "sha": "def456"
    }
  },
  "repository": {
    "full_name": "myorg/myrepo"
  }
}
```

## Testing Workflow

### Complete Testing Workflow

```bash
# 1. Validate configuration
./config-validator validate -config copier-config.yaml -v

# 2. Test pattern matching
./config-validator test-pattern \
  -type regex \
  -pattern "^examples/(?P<lang>[^/]+)/(?P<file>.+)$" \
  -file "examples/go/main.go"

# 3. Start app in dry-run mode
DRY_RUN=true CONFIG_FILE=copier-config.yaml ./examples-copier &

# 4. Test with example payload
./test-webhook -payload test-payloads/example-pr-merged.json

# 5. Check metrics
curl http://localhost:8080/metrics | jq

# 6. Test with real PR
export GITHUB_TOKEN=ghp_...
./test-webhook -pr 42 -owner myorg -repo myrepo

# 7. Review logs
grep "matched" logs/app.log
```

## Troubleshooting

### Connection Refused

**Error:**
```
Error: connection refused
```

**Solution:** Ensure the app is running:
```bash
curl http://localhost:8080/health
```

### 401 Unauthorized

**Error:**
```
Response: 401 Unauthorized
```

**Solution:** Disable webhook signature verification for testing:
```bash
unset WEBHOOK_SECRET
./examples-copier &
```

### 404 Not Found

**Error:**
```
Response: 404 Not Found
```

**Solution:** Check the webhook URL:
```bash
# Default is /webhook
./test-webhook -payload test.json -url http://localhost:8080/webhook
```

### GitHub API Rate Limit

**Error:**
```
Error: GitHub API rate limit exceeded
```

**Solution:**
- Wait for rate limit reset
- Use authenticated requests with `GITHUB_TOKEN`
- Use example payloads instead of real PR data

### Invalid Payload

**Error:**
```
Error: invalid JSON payload
```

**Solution:** Validate your JSON:
```bash
cat test-payloads/my-test.json | jq
```

## Advanced Usage

### Testing Multiple PRs

```bash
# Create script to test multiple PRs
cat > test-multiple-prs.sh << 'EOF'
#!/bin/bash
export GITHUB_TOKEN=ghp_...

for pr in 42 43 44 45; do
  echo "Testing PR #$pr..."
  ./test-webhook -pr $pr -owner myorg -repo myrepo
  sleep 2
done
EOF

chmod +x test-multiple-prs.sh
./test-multiple-prs.sh
```

### Automated Testing

```bash
# Create test script
cat > run-tests.sh << 'EOF'
#!/bin/bash
set -e

echo "Starting app..."
DRY_RUN=true CONFIG_FILE=copier-config.yaml ./examples-copier &
APP_PID=$!
sleep 2

echo "Running tests..."
./test-webhook -payload test-payloads/example-pr-merged.json

echo "Checking metrics..."
curl -s http://localhost:8080/metrics | jq '.files.matched'

echo "Stopping app..."
kill $APP_PID

echo "Tests complete!"
EOF

chmod +x run-tests.sh
./run-tests.sh
```

### Integration with CI/CD

```yaml
# .github/workflows/test.yml
name: Test Examples Copier

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      
      - name: Set up Go
        uses: actions/setup-go@v2
        with:
          go-version: 1.23.4
      
      - name: Build
        run: |
          go build -o examples-copier .
          go build -o test-webhook ./cmd/test-webhook
      
      - name: Test
        run: |
          DRY_RUN=true CONFIG_FILE=copier-config.yaml ./examples-copier &
          sleep 2
          ./test-webhook -payload test-payloads/example-pr-merged.json
```

## Exit Codes

- `0` - Success
- `1` - Error occurred

## See Also

- [Webhook Testing Guide](../../docs/WEBHOOK-TESTING.md) - Comprehensive testing guide
- [Local Testing](../../docs/LOCAL-TESTING.md) - Local development
- [Test Payloads](../../test-payloads/README.md) - Example payloads
- [Quick Reference](../../docs/QUICK-REFERENCE.md) - All commands

