# Local Testing - Quick Summary

## The Problem You Had

```bash
DRY_RUN=true ./examples-copier
# Error: projects/GOOGLE_CLOUD_PROJECT_ID is not a valid resource name
```

**Cause:** App tried to use Google Cloud Logging without a valid GCP project ID.

## The Solution

### Quick Fix (One Command)

```bash
COPIER_DISABLE_CLOUD_LOGGING=true DRY_RUN=true ./examples-copier
```

### Better Solution (Use Helper Script)

```bash
# One time setup
chmod +x scripts/run-local.sh

# Run anytime
./scripts/run-local.sh

# Or with make
make run-local-quick
```

## Complete Testing Workflow

### Terminal 1: Start the App

```bash
# Quick method
make run-local-quick

# You should see:
# ╔════════════════════════════════════════════════════════════════╗
# ║  GitHub Code Example Copier                                    ║
# ╠════════════════════════════════════════════════════════════════╣
# ║  Port:         8080                                            ║
# ║  Webhook Path: /webhook                                        ║
# ║  Config File:  config.json                                     ║
# ║  Dry Run:      true                                            ║
# ║  Audit Log:    false                                           ║
# ║  Metrics:      true                                            ║
# ╚════════════════════════════════════════════════════════════════╝
# 
# [INFO] Starting web server on port :8080
# ✓ No errors!
```

### Terminal 2: Test with Webhooks

#### Option A: Test with Example Payload

```bash
./test-webhook -payload test-payloads/example-pr-merged.json
```

#### Option B: Test with Real PR

```bash
# Set your GitHub token (one time)
export GITHUB_TOKEN=ghp_your_token_here

# Test with real PR
./test-webhook -pr 456 -owner mongodb -repo docs-realm
```

#### Option C: Interactive Testing

```bash
./scripts/test-with-pr.sh 456 mongodb docs-realm
```

### What You'll See

**In Terminal 1 (App Logs):**
```
[INFO] Webhook received: pull_request event
[INFO] PR #456 merged: "Add new Go examples"
[INFO] Processing 5 files from PR
[INFO] Pattern matched: examples/go/database/connect.go
[INFO]   → Transformed to: docs/go/database/connect.go
[INFO]   → Variables: lang=go, category=database, file=connect.go
[INFO]   → Commit message: "Update database examples from go"
[DRY-RUN] Would create commit with 2 files
[DRY-RUN] Would create PR: "Update examples"
[INFO] Metrics updated: files_matched=2
```

**In Terminal 2 (Test Tool):**
```
✓ Fetched PR #456 from mongodb/docs-realm
✓ Added signature: sha256=abc123...
✓ Webhook sent successfully to http://localhost:8080/webhook
✓ Response: 200 OK
```

## What Gets Tested

When you combine **DRY_RUN mode** with **real PR data**, you validate:

### ✅ Pattern Matching
- Your patterns match the actual files from the PR
- Variables are extracted correctly
- Files are filtered as expected

### ✅ Path Transformations
- Paths are transformed correctly
- Variables substitute properly
- Target paths are what you expect

### ✅ Message Templating
- Commit messages render correctly
- PR titles format as expected
- Variables work in templates

### ✅ Configuration
- Config file is valid
- All rules work
- No errors in processing

### ❌ What Doesn't Happen (Dry-Run)
- No actual commits to GitHub
- No PRs created
- No files uploaded
- No changes to any repository

## Validation Checklist

After testing, verify:

- [ ] **Logs show files matched** - Check Terminal 1
- [ ] **Path transformations correct** - See transformed paths in logs
- [ ] **Messages render properly** - Check commit messages
- [ ] **No errors** - No red error messages
- [ ] **Metrics updated** - `curl http://localhost:8080/metrics | jq`
- [ ] **Health check passes** - `curl http://localhost:8080/health | jq`

## Files Created for Local Testing

```
examples-copier/
├── configs/
│   └── .env.local              # Local environment template
├── scripts/
│   ├── run-local.sh            # Helper script to run locally
│   └── test-with-pr.sh         # Interactive PR testing
├── test-payloads/
│   ├── example-pr-merged.json  # Example webhook payload
│   └── README.md               # Payload documentation
├── LOCAL-TESTING.md            # Complete local testing guide
├── LOCAL-TESTING-SUMMARY.md    # This file
└── Makefile                    # Updated with local targets
```

## Environment Variables Explained

### For Running Locally

```bash
COPIER_DISABLE_CLOUD_LOGGING=true  # ← Fixes your error!
DRY_RUN=true                       # No actual commits
LOG_LEVEL=debug                    # Detailed logs
METRICS_ENABLED=true               # Enable /metrics
```

### For Testing with Real PRs

```bash
GITHUB_TOKEN=ghp_...               # Get from github.com/settings/tokens
                                   # Required scope: repo (read)
```

### Optional

```bash
AUDIT_ENABLED=false                # Disable MongoDB (for local)
CONFIG_FILE=config.json            # Your config file
PORT=8080                          # Server port
```

## Quick Commands Reference

```bash
# Build tools
make build

# Start app locally
make run-local-quick

# Test with example
./test-webhook -payload test-payloads/example-pr-merged.json

# Test with real PR
export GITHUB_TOKEN=ghp_...
./test-webhook -pr 456 -owner mongodb -repo docs-realm

# Check metrics
curl http://localhost:8080/metrics | jq

# Check health
curl http://localhost:8080/health | jq

# Validate config
./config-validator validate -config config.json -v
```

## Common Issues & Solutions

### Issue: Cloud logging error
```
logging client: rpc error: code = InvalidArgument desc = projects/GOOGLE_CLOUD_PROJECT_ID...
```
**Solution:** `COPIER_DISABLE_CLOUD_LOGGING=true`

### Issue: Connection refused
```
Error sending webhook: dial tcp: connect: connection refused
```
**Solution:** Make sure app is running in Terminal 1

### Issue: Can't fetch PR
```
Error: GITHUB_TOKEN environment variable not set
```
**Solution:** `export GITHUB_TOKEN=ghp_your_token_here`

### Issue: Files not matched
```
[INFO] Processing 5 files from PR
[INFO] No files matched any patterns
```
**Solution:** Test your pattern with `config-validator test-pattern`

## Next Steps

After successful local testing:

1. ✅ Verify patterns match your files
2. ✅ Confirm transformations are correct
3. ✅ Check messages render properly
4. ✅ No errors in processing

Then:

1. Deploy to staging
2. Test with real GitHub webhooks
3. Monitor metrics and logs
4. Deploy to production

## Documentation

- **[LOCAL-TESTING.md](LOCAL-TESTING.md)** - Complete local testing guide
- **[WEBHOOK-TESTING.md](WEBHOOK-TESTING.md)** - Webhook testing details
- **[QUICK-REFERENCE.md](QUICK-REFERENCE.md)** - Command reference
- **[README.md](README.md)** - Main documentation

## Summary

**Your original command:**
```bash
DRY_RUN=true ./examples-copier
# ❌ Error: Cloud logging issue
```

**Fixed command:**
```bash
COPIER_DISABLE_CLOUD_LOGGING=true DRY_RUN=true ./examples-copier
# ✅ Works perfectly!
```

**Even better:**
```bash
make run-local-quick
# ✅ Sets everything up correctly
```

**Then test:**
```bash
# Terminal 2
./test-webhook -pr 456 -owner mongodb -repo docs-realm
# ✅ Tests with real PR data in dry-run mode
```

**Result:** You can test everything locally without:
- ❌ Google Cloud credentials
- ❌ MongoDB setup
- ❌ Making actual commits
- ❌ Deploying anywhere

But you still validate:
- ✅ Pattern matching works
- ✅ Path transformations correct
- ✅ Message templates render
- ✅ Configuration is valid
- ✅ Real PR data processes correctly

Perfect for development and testing! 🎯

