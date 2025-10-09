# Local Testing Guide

This guide explains how to run and test the examples-copier application locally without requiring Google Cloud or MongoDB.

## Quick Start

### Option 1: Use the Helper Script (Recommended)

```bash
# Build and run in local development mode
make run-local

# Or directly
./scripts/run-local.sh
```

### Option 2: Quick Command

```bash
# One-liner for quick testing
COPIER_DISABLE_CLOUD_LOGGING=true DRY_RUN=true ./examples-copier
```

### Option 3: Use Makefile

```bash
# Build and run with local settings
make run-local-quick
```

## Setup for Local Testing

### 1. Create Local Environment File

```bash
# Copy the local template
cp configs/.env.local configs/.env

# Edit with your values (optional)
nano configs/.env
```

### 2. Minimal Configuration

For basic local testing, you only need:

```bash
# configs/.env
COPIER_DISABLE_CLOUD_LOGGING=true
DRY_RUN=true
CONFIG_FILE=config.json
```

### 3. For Testing with Real PRs

Add a GitHub token:

```bash
# Get token from: https://github.com/settings/tokens
# Required scope: repo (read access)

# Add to configs/.env
GITHUB_TOKEN=ghp_your_token_here
```

## Running Locally

### Start the Application

```bash
# Terminal 1: Start the app
make run-local-quick

# You should see:
# ╔════════════════════════════════════════════════════════════════╗
# ║  GitHub Code Example Copier                                    ║
# ╠════════════════════════════════════════════════════════════════╣
# ║  Port:         8080                                            ║
# ║  Webhook Path: /events                                         ║
# ║  Config File:  copier-config.example.yaml                      ║
# ║  Dry Run:      true                                            ║
# ║  Audit Log:    false                                           ║
# ║  Metrics:      true                                            ║
# ╚════════════════════════════════════════════════════════════════╝
```

### Test with Webhook

```bash
# Terminal 2: Send test webhook (automatically fetches webhook secret)
make test-webhook-example

# Or send webhook manually with secret
export WEBHOOK_SECRET=$(gcloud secrets versions access latest --secret=webhook-secret)
./test-webhook -payload test-payloads/example-pr-merged.json -secret "$WEBHOOK_SECRET"

# Or test with real PR
export GITHUB_TOKEN=ghp_...
export WEBHOOK_SECRET=$(gcloud secrets versions access latest --secret=webhook-secret)
./test-webhook -pr 456 -owner mongodb -repo docs-realm -secret "$WEBHOOK_SECRET"
```

## What Happens in Local Mode

### ✅ What Works

- ✅ Webhook processing
- ✅ Pattern matching
- ✅ Path transformations
- ✅ Message templating
- ✅ File state management
- ✅ Metrics collection
- ✅ Health checks
- ✅ Logging to stdout

### ❌ What's Disabled (in Dry-Run)

- ❌ Actual commits to GitHub
- ❌ Creating pull requests
- ❌ Uploading files
- ❌ Google Cloud Logging (uses stdout instead)
- ❌ MongoDB audit logging (unless you enable it)

### 📊 What You Can Verify

1. **Pattern Matching**
   - Check logs to see which files matched
   - Verify patterns work correctly

2. **Path Transformations**
   - See transformed paths in logs
   - Verify variables are extracted

3. **Message Templates**
   - See rendered commit messages
   - Verify PR titles are correct

4. **Configuration**
   - Validate config file loads
   - Check for errors

## Testing Scenarios

### Scenario 1: Test Configuration Changes

```bash
# 1. Edit your config file
nano config.json  # or copier-config.yaml

# 2. Validate it
./config-validator validate -config config.json -v

# 3. Start app
make run-local

# 4. Send test webhook
./test-webhook -payload test-payloads/example-pr-merged.json

# 5. Check logs to verify changes work
```

### Scenario 2: Test with Real PR

```bash
# 1. Set GitHub token
export GITHUB_TOKEN=ghp_your_token_here

# 2. Start app in one terminal
make run-local

# 3. In another terminal, test with real PR
./test-webhook -pr 456 -owner mongodb -repo docs-realm

# 4. Watch Terminal 1 for processing logs
```

### Scenario 3: Test Pattern Matching

```bash
# 1. Create custom test payload
cat > test-my-pattern.json <<EOF
{
  "action": "closed",
  "pull_request": {"merged": true, "merge_commit_sha": "abc"},
  "files": [
    {"filename": "examples/go/database/connect.go", "status": "added"},
    {"filename": "examples/python/auth/login.py", "status": "added"}
  ]
}
EOF

# 2. Start app
make run-local

# 3. Send test
./test-webhook -payload test-my-pattern.json

# 4. Verify in logs which files matched
```

## Checking Results

### View Logs

Logs go to stdout when cloud logging is disabled:

```bash
# You'll see logs like:
[INFO] Webhook received: pull_request event
[INFO] PR #42 merged: "Add Go database examples"
[INFO] Processing 5 files from PR
[DEBUG] Testing pattern: ^examples/(?P<lang>[^/]+)/(?P<category>[^/]+)/.*$
[INFO] Pattern matched: examples/go/database/connect.go
[INFO]   → Transformed to: docs/go/database/connect.go
[INFO]   → Variables: lang=go, category=database
[DRY-RUN] Would create commit with 2 files
[DRY-RUN] Would create PR: "Update database examples"
```

### Check Metrics

```bash
curl http://localhost:8080/metrics | jq
```

**Output:**
```json
{
  "webhooks": {
    "received": 1,
    "processed": 1,
    "failed": 0,
    "success_rate": 100
  },
  "files": {
    "matched": 2,
    "uploaded": 0,
    "upload_failed": 0
  }
}
```

### Check Health

```bash
curl http://localhost:8080/health | jq
```

**Output:**
```json
{
  "status": "healthy",
  "started": true,
  "github": {
    "status": "healthy",
    "authenticated": true
  },
  "queues": {
    "upload_count": 0,
    "deprecation_count": 0
  },
  "uptime": "5m30s"
}
```

## Environment Variables for Local Testing

### Required (Minimal)

```bash
COPIER_DISABLE_CLOUD_LOGGING=true  # Use stdout instead of GCP
DRY_RUN=true                       # Don't make actual commits
```

### Recommended

```bash
LOG_LEVEL=debug                    # Detailed logging
COPIER_DEBUG=true                  # Extra debug info
METRICS_ENABLED=true               # Enable /metrics endpoint
CONFIG_FILE=config.json            # Your config file
```

### Optional (for Real PR Testing)

```bash
GITHUB_TOKEN=ghp_...               # For fetching real PRs
REPO_OWNER=mongodb                 # Default repo owner
REPO_NAME=docs-realm               # Default repo name
```

### Optional (for Audit Logging)

```bash
AUDIT_ENABLED=true                 # Enable audit logging
MONGO_URI=mongodb://localhost:27017  # Local MongoDB
# Or use MongoDB Atlas:
# MONGO_URI=mongodb+srv://user:pass@cluster.mongodb.net
AUDIT_DATABASE=code_copier_dev
AUDIT_COLLECTION=audit_events
```

## Troubleshooting

### Error: "projects/GOOGLE_CLOUD_PROJECT_ID is not a valid resource name"

**Problem:** Cloud logging is enabled but GCP_PROJECT_ID is not set

**Solution:**
```bash
# Disable cloud logging for local testing
COPIER_DISABLE_CLOUD_LOGGING=true ./examples-copier
```

### Error: "connection refused" when sending webhook

**Problem:** Application is not running, or you're trying to run both in the same terminal

**Solution:**
```bash
# Terminal 1: Start the app (this blocks the terminal)
make run-local-quick

# Terminal 2: In a NEW terminal window, send the webhook
cd examples-copier
make test-webhook-example

# Or manually:
export WEBHOOK_SECRET=$(gcloud secrets versions access latest --secret=webhook-secret)
./test-webhook -payload test-payloads/example-pr-merged.json -secret "$WEBHOOK_SECRET"
```

**Note:** The `make test-webhook-example` command requires the server to be running in a separate terminal. You cannot run both commands in the same terminal unless you background the server process.

### Error: "GITHUB_TOKEN environment variable not set"

**Problem:** Trying to fetch real PR without token

**Solution:**
```bash
# Get token from https://github.com/settings/tokens
export GITHUB_TOKEN=ghp_your_token_here

# Then try again
./test-webhook -pr 456 -owner mongodb -repo docs-realm
```

### No files matched in logs

**Problem:** Pattern doesn't match the files

**Solution:**
```bash
# Test your pattern
./config-validator test-pattern \
  -type regex \
  -pattern "^examples/(?P<lang>[^/]+)/.*$" \
  -file "examples/go/main.go"

# Check config file
./config-validator validate -config config.json -v
```

## Complete Testing Workflow

### Full Local Testing Cycle

```bash
# 1. Build everything
make build

# 2. Validate configuration
./config-validator validate -config config.json -v

# 3. Test pattern matching
./config-validator test-pattern \
  -type regex \
  -pattern "^examples/(?P<lang>[^/]+)/(?P<category>[^/]+)/.*$" \
  -file "examples/go/database/connect.go"

# 4. Start app in Terminal 1
make run-local

# 5. In Terminal 2, test with example payload
./test-webhook -payload test-payloads/example-pr-merged.json

# 6. Check metrics
curl http://localhost:8080/metrics | jq

# 7. Test with real PR (if you have GITHUB_TOKEN)
export GITHUB_TOKEN=ghp_...
./test-webhook -pr 456 -owner mongodb -repo docs-realm

# 8. Review logs in Terminal 1

# 9. Stop app (Ctrl+C in Terminal 1)
```

## Tips for Effective Local Testing

1. **Always start with dry-run mode** - Never test with real commits locally
2. **Use debug logging** - Set `LOG_LEVEL=debug` to see everything
3. **Test patterns first** - Use `config-validator` before running the app
4. **Create custom payloads** - Test specific scenarios
5. **Check metrics** - Verify counts are correct
6. **Use real PR data** - Most realistic testing
7. **Keep test payloads** - Save them for regression testing
8. **Monitor logs** - Watch for errors or unexpected behavior

## Next Steps

After successful local testing:

1. ✅ Patterns match correctly
2. ✅ Transformations work as expected
3. ✅ Messages render properly
4. ✅ No errors in processing

Then you can:

1. Deploy to staging environment
2. Test with real webhooks from GitHub
3. Monitor metrics and audit logs
4. Deploy to production

See [DEPLOYMENT.md](DEPLOYMENT.md) for deployment instructions.

## Quick Reference

```bash
# Terminal 1: Start app locally
make run-local-quick

# Terminal 2: Test with example (auto-fetches webhook secret)
make test-webhook-example

# Or test manually with webhook secret
export WEBHOOK_SECRET=$(gcloud secrets versions access latest --secret=webhook-secret)
./test-webhook -payload test-payloads/example-pr-merged.json -secret "$WEBHOOK_SECRET"

# Test with real PR
export GITHUB_TOKEN=ghp_...
export WEBHOOK_SECRET=$(gcloud secrets versions access latest --secret=webhook-secret)
./test-webhook -pr 456 -owner mongodb -repo docs-realm -secret "$WEBHOOK_SECRET"

# Check metrics
curl http://localhost:8080/metrics | jq

# Check health
curl http://localhost:8080/health | jq

# Validate config
./config-validator validate -config copier-config.yaml -v
```

