# Testing Enhancements Summary

## Overview

Extended the examples-copier testing capabilities with comprehensive webhook testing tools that support both example payloads and real PR data.

## New Testing Tools

### 1. Test Webhook CLI Tool

**Location:** `cmd/test-webhook/main.go`

**Features:**
- Send test webhooks to local or remote endpoints
- Fetch real PR data from GitHub API
- Use custom payload files
- Generate HMAC signatures for authentication
- Dry-run mode to preview payloads

**Usage:**
```bash
# Build
go build -o test-webhook ./cmd/test-webhook

# Use example payload
./test-webhook

# Use custom payload
./test-webhook -payload test-payloads/example-pr-merged.json

# Fetch real PR data
./test-webhook -pr 123 -owner myorg -repo myrepo

# Test against production
./test-webhook -pr 123 -owner myorg -repo myrepo \
  -url https://myapp.appspot.com/webhook \
  -secret "webhook-secret"

# Dry-run (see payload without sending)
./test-webhook -pr 123 -owner myorg -repo myrepo -dry-run
```

### 2. Interactive Test Script

**Location:** `scripts/test-with-pr.sh`

**Features:**
- Interactive PR testing workflow
- Fetches and displays PR metadata
- Confirms before sending webhook
- Checks if application is running
- Helpful error messages and guidance

**Usage:**
```bash
# Make executable
chmod +x scripts/test-with-pr.sh

# Test with PR (uses REPO_OWNER and REPO_NAME from env)
./scripts/test-with-pr.sh 123

# Test with specific repo
./scripts/test-with-pr.sh 123 myorg myrepo

# Test against production
WEBHOOK_URL=https://myapp.appspot.com/webhook ./scripts/test-with-pr.sh 123
```

### 3. Example Test Payloads

**Location:** `test-payloads/`

**Files:**
- `example-pr-merged.json` - Complete example with multiple file types
- `README.md` - Documentation for test payloads

**Example payload includes:**
- Multiple file changes (added, modified, removed)
- Go and Python examples
- Database and auth categories
- Realistic file structure for testing patterns

### 4. Makefile Targets

**Location:** `Makefile`

**New targets:**
```bash
# Build all tools
make build

# Run unit tests
make test-unit

# Test with example payload
make test-webhook-example

# Test with real PR
make test-webhook-pr PR=123 OWNER=myorg REPO=myrepo

# Quick test cycle
make quick-test

# Full test cycle
make full-test
```

## Testing Workflows

### Workflow 1: Local Development

```bash
# Terminal 1: Start app in dry-run mode
DRY_RUN=true ./examples-copier

# Terminal 2: Send test webhook
./test-webhook -payload test-payloads/example-pr-merged.json

# Verify in Terminal 1:
# - Files matched by patterns
# - Path transformations correct
# - Message templates rendered
# - No errors
```

### Workflow 2: Test with Real PR

```bash
# Set GitHub token
export GITHUB_TOKEN=ghp_your_token_here

# Interactive testing
./scripts/test-with-pr.sh 456 myorg myrepo

# Or direct
./test-webhook -pr 456 -owner myorg -repo myrepo

# Verify:
# - Real file paths match patterns
# - Actual PR metadata used
# - All files processed
```

### Workflow 3: Staging Environment

```bash
# Test against staging
./test-webhook -pr 123 -owner myorg -repo myrepo \
  -url https://staging-app.appspot.com/webhook \
  -secret "staging-secret"

# Verify:
# - Webhook signature works
# - Staging processes correctly
# - Audit logs created
# - Metrics updated
```

### Workflow 4: Pattern Testing

```bash
# Create custom payload for specific pattern
cat > test-pattern.json <<EOF
{
  "action": "closed",
  "pull_request": {"merged": true, "merge_commit_sha": "abc"},
  "files": [
    {"filename": "examples/go/database/connect.go", "status": "added"}
  ]
}
EOF

# Test
DRY_RUN=true ./examples-copier &
./test-webhook -payload test-pattern.json

# Verify pattern matching and transformations
```

## Documentation

### New Documentation Files

1. **WEBHOOK-TESTING.md** (300 lines)
   - Comprehensive webhook testing guide
   - Testing scenarios and workflows
   - Troubleshooting guide
   - Best practices
   - CI/CD integration examples

2. **test-payloads/README.md**
   - Test payload documentation
   - Usage examples
   - Creating custom payloads
   - Testing scenarios

3. **Updated README.md**
   - Added webhook testing section
   - Examples for all three testing methods
   - Links to detailed documentation

4. **Updated QUICK-REFERENCE.md**
   - Added webhook testing commands
   - Test tool options
   - Common testing patterns

## Key Features

### Real PR Data Fetching

The test tool can fetch real PR data from GitHub:

```bash
# Fetches PR metadata and files
./test-webhook -pr 123 -owner myorg -repo myrepo
```

**What it fetches:**
- PR number, state, merged status
- Merge commit SHA
- Head and base branch info
- All files changed in the PR
- File statuses (added, modified, removed)

### Webhook Signature Generation

Automatically generates HMAC-SHA256 signatures:

```bash
./test-webhook -payload test.json -secret "my-secret"
```

**Signature:**
- Uses same algorithm as GitHub
- Adds `X-Hub-Signature-256` header
- Enables testing of signature verification

### Dry-Run Preview

Preview payloads before sending:

```bash
./test-webhook -pr 123 -owner myorg -repo myrepo -dry-run
```

**Output:**
- Pretty-printed JSON payload
- Shows exactly what would be sent
- Useful for debugging patterns

### Interactive Testing

Helper script provides guided testing:

```bash
./scripts/test-with-pr.sh 123
```

**Features:**
- Fetches and displays PR info
- Confirms before sending
- Checks if app is running
- Shows next steps

## Testing Capabilities

### What You Can Test

1. **Pattern Matching**
   - Prefix patterns
   - Glob patterns
   - Regex patterns with variables

2. **Path Transformations**
   - Template rendering
   - Variable substitution
   - Built-in and custom variables

3. **Message Templating**
   - Commit messages
   - PR titles and bodies
   - Variable extraction

4. **Deprecation Tracking**
   - Deleted file handling
   - Deprecation file updates
   - Audit logging

5. **Multiple Targets**
   - Different repos
   - Different branches
   - Different transformations

6. **Commit Strategies**
   - Direct commits
   - Pull requests
   - Auto-merge behavior

### Validation Points

After testing, verify:

✅ **Application Logs**
- Webhook received
- Files matched
- Transformations applied
- No errors

✅ **Metrics Endpoint**
- Webhooks received/processed
- Files matched/uploaded
- Success rates

✅ **Health Endpoint**
- Status healthy
- GitHub authenticated
- Queue counts correct

✅ **Audit Logs** (if enabled)
- Events created
- Correct event types
- Accurate metadata

## Integration with Existing Tests

### Unit Tests (51 tests)
- Pattern matching
- Config loading
- File state management
- Metrics collection

### Webhook Tests (NEW)
- End-to-end webhook processing
- Real PR data testing
- Pattern matching validation
- Transformation verification

### Combined Testing
```bash
# Run all tests
make test-unit                    # Unit tests
make test-webhook-example         # Webhook with example
make test-webhook-pr PR=123 ...   # Webhook with real PR
```

## Environment Variables

### For Test Tool

```bash
GITHUB_TOKEN      # GitHub API token (for fetching PRs)
WEBHOOK_SECRET    # Default webhook secret
REPO_OWNER        # Default repository owner
REPO_NAME         # Default repository name
WEBHOOK_URL       # Default webhook endpoint
```

### For Application

```bash
DRY_RUN=true      # Test without making commits
LOG_LEVEL=debug   # Detailed logging
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
10. **Document test scenarios** for your use case

## Troubleshooting

### Common Issues

**Webhook returns 401:**
- Check webhook secret matches
- Verify signature generation

**Files not matched:**
- Test pattern with `config-validator test-pattern`
- Check pattern syntax in config

**Path transformation wrong:**
- Test with `config-validator test-transform`
- Verify variable names match

**Can't fetch PR:**
- Check `GITHUB_TOKEN` is set
- Verify token has repo read access
- Check PR number and repo are correct

## Next Steps

1. **Test your configuration** with example payloads
2. **Test with real PRs** from your repository
3. **Validate patterns** match your file structure
4. **Verify transformations** produce correct paths
5. **Test in staging** before production
6. **Set up CI/CD** with webhook tests
7. **Deploy to production** with confidence

## Summary

The webhook testing enhancements provide:

✅ **Three testing methods:**
- Example payloads
- Real PR data
- Interactive script

✅ **Complete tooling:**
- CLI test tool
- Helper scripts
- Example payloads
- Makefile targets

✅ **Comprehensive documentation:**
- Webhook testing guide
- Quick reference
- Troubleshooting
- Best practices

✅ **Full validation:**
- Logs
- Metrics
- Health checks
- Audit logs

The testing infrastructure is production-ready and enables confident deployment of configuration changes! 🚀

