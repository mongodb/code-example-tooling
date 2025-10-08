# Test Payloads

This directory contains example webhook payloads for testing the examples-copier application.

## Files

### example-pr-merged.json
A complete example of a merged PR webhook payload with:
- Multiple file changes (added, modified, removed)
- Go and Python examples
- Database and auth categories
- Realistic file structure

## Usage

### Option 1: Use Example Payload

```bash
# Build the test tool
go build -o test-webhook ./cmd/test-webhook

# Send example payload
./test-webhook -payload test-payloads/example-pr-merged.json
```

### Option 2: Fetch Real PR Data

```bash
# Set GitHub token
export GITHUB_TOKEN=ghp_your_token_here

# Test with real PR
./test-webhook -pr 123 -owner myorg -repo myrepo
```

### Option 3: Use Helper Script

```bash
# Make script executable
chmod +x scripts/test-with-pr.sh

# Test with real PR (interactive)
./scripts/test-with-pr.sh 123 myorg myrepo
```

## Testing Scenarios

### Test Pattern Matching

Create custom payloads to test specific patterns:

**Test Regex Pattern:**
```json
{
  "files": [
    {
      "filename": "examples/go/database/connect.go",
      "status": "added"
    }
  ]
}
```

**Test Glob Pattern:**
```json
{
  "files": [
    {
      "filename": "examples/go/main.go",
      "status": "added"
    },
    {
      "filename": "examples/python/main.py",
      "status": "added"
    }
  ]
}
```

### Test Deprecation

```json
{
  "files": [
    {
      "filename": "examples/deprecated-example.go",
      "status": "removed"
    }
  ]
}
```

### Test Multiple Languages

```json
{
  "files": [
    {
      "filename": "examples/go/database/connect.go",
      "status": "added"
    },
    {
      "filename": "examples/python/database/connect.py",
      "status": "added"
    },
    {
      "filename": "examples/javascript/database/connect.js",
      "status": "added"
    }
  ]
}
```

## Creating Custom Payloads

1. Copy `example-pr-merged.json`
2. Modify the `files` array to match your test case
3. Update PR metadata as needed
4. Save with descriptive name (e.g., `test-go-examples.json`)

## Testing with Dry-Run Mode

Test without making actual commits:

```bash
# Start app in dry-run mode
DRY_RUN=true ./examples-copier &

# Send test webhook
./test-webhook -payload test-payloads/example-pr-merged.json

# Check logs for pattern matching and transformations
```

## Validating Results

After sending a test webhook:

1. **Check Application Logs**
   ```bash
   # Local
   tail -f logs/app.log
   
   # GCP
   gcloud app logs tail -s default
   ```

2. **Check Metrics**
   ```bash
   curl http://localhost:8080/metrics | jq
   ```

3. **Check Audit Logs** (if enabled)
   ```javascript
   db.audit_events.find().sort({timestamp: -1}).limit(10)
   ```

4. **Verify Pattern Matching**
   - Check which files were matched
   - Verify path transformations
   - Confirm message templating

## Common Test Cases

### Test Case 1: New Go Examples
```bash
./test-webhook -payload test-payloads/example-pr-merged.json
```
Expected: Files copied to target repo with transformed paths

### Test Case 2: Real PR from Production
```bash
export GITHUB_TOKEN=ghp_...
./scripts/test-with-pr.sh 456 mongodb docs-realm
```
Expected: Real PR data fetched and processed

### Test Case 3: Dry-Run Validation
```bash
DRY_RUN=true ./examples-copier &
./test-webhook -payload test-payloads/example-pr-merged.json
```
Expected: Processing logged but no commits made

## Troubleshooting

### Webhook Returns 401
- Check webhook secret matches
- Verify signature generation

### Files Not Matched
- Check pattern in config.yaml
- Use `config-validator test-pattern` to debug

### Path Transformation Wrong
- Use `config-validator test-transform` to debug
- Check variable names in template

### No Response
- Verify app is running
- Check webhook URL is correct
- Review application logs

