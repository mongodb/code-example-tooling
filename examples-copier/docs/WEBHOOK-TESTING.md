# Webhook Testing Guide

This guide explains how to test the examples-copier application with webhooks using real PR data or example payloads.

## Quick Start

### 1. Build the Test Tool

```bash
# Using Make
make test-webhook

# Or manually
go build -o test-webhook ./cmd/test-webhook
```

### 2. Test with Example Payload

```bash
# Send example payload to local server
./test-webhook -payload test-payloads/example-pr-merged.json

# See payload without sending
./test-webhook -payload test-payloads/example-pr-merged.json -dry-run
```

### 3. Test with Real PR Data

```bash
# Set GitHub token
export GITHUB_TOKEN=ghp_your_token_here

# Fetch and send real PR data
./test-webhook -pr 123 -owner myorg -repo myrepo

# Interactive testing with helper script
./scripts/test-with-pr.sh 123 myorg myrepo
```

## Testing Scenarios

### Scenario 1: Local Development Testing

Test your configuration changes locally before deploying:

```bash
# Terminal 1: Start app in dry-run mode
DRY_RUN=true ./examples-copier

# Terminal 2: Send test webhook
./test-webhook -payload test-payloads/example-pr-merged.json

# Check Terminal 1 for processing logs
```

**What to verify:**
- Files are matched by patterns
- Path transformations are correct
- Message templates render properly
- No errors in processing

### Scenario 2: Test with Real PR

Test with actual PR data from your repository:

```bash
# Set environment
export GITHUB_TOKEN=ghp_your_token_here
export REPO_OWNER=myorg
export REPO_NAME=myrepo

# Use helper script (interactive)
./scripts/test-with-pr.sh 456

# Or use test-webhook directly
./test-webhook -pr 456 -owner myorg -repo myrepo
```

**What to verify:**
- Real file paths match your patterns
- Actual PR metadata is used correctly
- All files from PR are processed

### Scenario 3: Test Against Staging

Test against your staging environment:

```bash
# Set staging URL
export WEBHOOK_URL=https://staging-myapp.appspot.com/webhook
export WEBHOOK_SECRET=your-staging-secret

# Test with real PR
./test-webhook -pr 123 -owner myorg -repo myrepo \
  -url $WEBHOOK_URL \
  -secret $WEBHOOK_SECRET
```

**What to verify:**
- Webhook signature verification works
- Staging environment processes correctly
- Audit logs are created (if enabled)
- Metrics are updated

### Scenario 4: Test Pattern Matching

Create custom payloads to test specific patterns:

```bash
# Create test payload
cat > test-go-only.json <<EOF
{
  "action": "closed",
  "number": 1,
  "pull_request": {
    "merged": true,
    "merge_commit_sha": "abc123"
  },
  "files": [
    {"filename": "examples/go/database/connect.go", "status": "added"},
    {"filename": "examples/go/auth/login.go", "status": "added"},
    {"filename": "examples/python/test.py", "status": "added"}
  ]
}
EOF

# Test
DRY_RUN=true ./examples-copier &
./test-webhook -payload test-go-only.json
```

**What to verify:**
- Only Go files are matched (if that's your pattern)
- Python files are ignored (if not in pattern)
- Variables are extracted correctly

### Scenario 5: Test Deprecation Tracking

Test file deletion handling:

```bash
# Create deprecation test payload
cat > test-deprecation.json <<EOF
{
  "action": "closed",
  "number": 1,
  "pull_request": {
    "merged": true,
    "merge_commit_sha": "abc123"
  },
  "files": [
    {"filename": "examples/old-example.go", "status": "removed"},
    {"filename": "examples/deprecated.go", "status": "removed"}
  ]
}
EOF

# Test
./test-webhook -payload test-deprecation.json
```

**What to verify:**
- Deleted files are tracked in deprecation file
- Audit logs show deprecation events
- Metrics count deprecated files

## Test Tool Options

### Command-Line Flags

```bash
-pr int         # PR number to fetch from GitHub
-owner string   # Repository owner (required with -pr)
-repo string    # Repository name (required with -pr)
-url string     # Webhook URL (default: http://localhost:8080/webhook)
-secret string  # Webhook secret for HMAC signature
-payload string # Path to custom payload JSON file
-dry-run        # Print payload without sending
-help           # Show help
```

### Environment Variables

```bash
GITHUB_TOKEN    # GitHub personal access token (for fetching PR data)
WEBHOOK_SECRET  # Default webhook secret (can be overridden with -secret)
REPO_OWNER      # Default repository owner
REPO_NAME       # Default repository name
WEBHOOK_URL     # Default webhook URL
```

## Helper Script

The `scripts/test-with-pr.sh` script provides an interactive testing experience:

```bash
# Make executable
chmod +x scripts/test-with-pr.sh

# Run with PR number
./scripts/test-with-pr.sh 123

# Or specify repo
./scripts/test-with-pr.sh 123 myorg myrepo
```

**Features:**
- Fetches PR metadata and displays it
- Confirms before sending
- Checks if app is running (for local testing)
- Shows helpful error messages
- Provides next steps after testing

## Creating Custom Test Payloads

### Minimal Payload

```json
{
  "action": "closed",
  "number": 1,
  "pull_request": {
    "merged": true,
    "merge_commit_sha": "abc123"
  },
  "files": [
    {"filename": "examples/test.go", "status": "added"}
  ]
}
```

### Complete Payload

See `test-payloads/example-pr-merged.json` for a complete example with:
- Multiple file changes (added, modified, removed)
- Full PR metadata
- Repository information
- Realistic file structure

### Testing Specific Features

**Test Regex Variables:**
```json
{
  "files": [
    {"filename": "examples/go/database/connect.go", "status": "added"}
  ]
}
```
With pattern: `^examples/(?P<lang>[^/]+)/(?P<category>[^/]+)/(?P<file>.+)$`
Should extract: `lang=go`, `category=database`, `file=connect.go`

**Test Multiple Languages:**
```json
{
  "files": [
    {"filename": "examples/go/main.go", "status": "added"},
    {"filename": "examples/python/main.py", "status": "added"},
    {"filename": "examples/javascript/main.js", "status": "added"}
  ]
}
```

**Test Path Transformations:**
```json
{
  "files": [
    {"filename": "examples/go/database/connect.go", "status": "added"}
  ]
}
```
With transform: `docs/${lang}/${category}/${file}`
Should produce: `docs/go/database/connect.go`

## Validation Checklist

After sending a test webhook, verify:

### Application Logs
```bash
# Local
tail -f logs/app.log

# GCP
gcloud app logs tail -s default
```

**Look for:**
- ✅ Webhook received
- ✅ Files matched by pattern
- ✅ Path transformations applied
- ✅ Variables extracted correctly
- ✅ No errors in processing

### Metrics Endpoint
```bash
curl http://localhost:8080/metrics | jq
```

**Verify:**
- ✅ `webhooks.received` incremented
- ✅ `webhooks.processed` incremented
- ✅ `files.matched` shows correct count
- ✅ `files.uploaded` updated (if not dry-run)

### Health Endpoint
```bash
curl http://localhost:8080/health | jq
```

**Verify:**
- ✅ Status is "healthy"
- ✅ GitHub authentication working
- ✅ Queue counts are correct

### Audit Logs (if enabled)
```javascript
// MongoDB query
db.audit_events.find().sort({timestamp: -1}).limit(10)
```

**Verify:**
- ✅ Event created for webhook
- ✅ Correct event type (copy/deprecation)
- ✅ File paths are correct
- ✅ Rule name matches config

## Troubleshooting

### Webhook Returns 401 Unauthorized

**Problem:** Signature verification failed

**Solution:**
```bash
# Make sure secret matches
./test-webhook -payload test.json -secret "correct-secret"

# Or disable signature check for testing
# (remove signature verification in code temporarily)
```

### Files Not Matched

**Problem:** Pattern doesn't match files

**Solution:**
```bash
# Test pattern with config-validator
./config-validator test-pattern \
  -type regex \
  -pattern "^examples/(?P<lang>[^/]+)/.*$" \
  -file "examples/go/main.go"

# Check config file pattern syntax
./config-validator validate -config config.yaml -v
```

### Path Transformation Wrong

**Problem:** Transformed path is incorrect

**Solution:**
```bash
# Test transformation
./config-validator test-transform \
  -template "docs/${lang}/${file}" \
  -file "examples/go/main.go" \
  -pattern "^examples/(?P<lang>[^/]+)/(?P<file>.+)$"

# Check variable names match
# Pattern: (?P<lang>...) -> Template: ${lang}
```

### No Response from Webhook

**Problem:** Webhook doesn't respond

**Solution:**
```bash
# Check if app is running
curl http://localhost:8080/health

# Check webhook URL
./test-webhook -payload test.json -url http://localhost:8080/webhook

# Check application logs for errors
```

### Real PR Fetch Fails

**Problem:** Can't fetch PR data from GitHub

**Solution:**
```bash
# Verify token is set
echo $GITHUB_TOKEN

# Test token manually
curl -H "Authorization: Bearer $GITHUB_TOKEN" \
  https://api.github.com/repos/owner/repo/pulls/123

# Check token permissions (needs repo read access)
```

## Best Practices

1. **Always test locally first** with dry-run mode
2. **Use real PR data** for realistic testing
3. **Create custom payloads** for edge cases
4. **Verify all metrics** after testing
5. **Check audit logs** to ensure tracking works
6. **Test with multiple file types** to verify patterns
7. **Test deprecation** by including removed files
8. **Use helper script** for interactive testing
9. **Keep test payloads** in version control
10. **Document test scenarios** for your specific use case

## Integration with CI/CD

Add webhook testing to your CI pipeline:

```yaml
# .github/workflows/test.yml
- name: Test webhook processing
  run: |
    # Start app in background
    DRY_RUN=true ./examples-copier &
    APP_PID=$!
    
    # Wait for app to start
    sleep 5
    
    # Run webhook tests
    ./test-webhook -payload test-payloads/example-pr-merged.json
    
    # Stop app
    kill $APP_PID
```

## Next Steps

After successful webhook testing:

1. Deploy to staging environment
2. Configure GitHub webhook in repository settings
3. Test with real PR merge
4. Monitor metrics and logs
5. Deploy to production
6. Set up alerts for failures

See [DEPLOYMENT-GUIDE.md](DEPLOYMENT-GUIDE.md) for deployment instructions.

