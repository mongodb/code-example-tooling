# Troubleshooting Guide

Common issues and solutions for the examples-copier application.

## Table of Contents

- [Configuration Issues](#configuration-issues)
- [Pattern Matching Issues](#pattern-matching-issues)
- [Webhook Issues](#webhook-issues)
- [Deployment Issues](#deployment-issues)
- [Slack Notification Issues](#slack-notification-issues)
- [Performance Issues](#performance-issues)
- [Debugging Tips](#debugging-tips)

## Configuration Issues

### Config File Not Found

**Error:**
```
[ERROR] failed to load config | {"error":"failed to retrieve config file: 404 Not Found"}
```

**Cause:** App tries to fetch config from GitHub but file doesn't exist.

**Solutions:**

1. **For local testing** - Create `copier-config.yaml` in the app directory:
   ```bash
   cp config.example.yaml copier-config.yaml
   # Edit copier-config.yaml with your settings
   CONFIG_FILE=copier-config.yaml make run-local-quick
   ```

2. **For production** - Add config to your source repository:
   ```bash
   # Copy config to source repo
   cp copier-config.yaml /path/to/source-repo/copier-config.yaml
   cd /path/to/source-repo
   git add copier-config.yaml
   git commit -m "Add examples-copier config"
   git push
   ```

### Invalid Configuration

**Error:**
```
[ERROR] config validation failed | {"error":"source_repo is required"}
```

**Solution:** Validate your config:
```bash
./config-validator validate -config copier-config.yaml -v
```

Fix the reported errors and validate again.

### Environment Variables Not Set

**Error:**
```
[ERROR] missing required environment variable: GITHUB_APP_ID
```

**Solution:**
```bash
# Check which variables are set
env | grep -E "(GITHUB|REPO|GCP)"

# Set missing variables
export GITHUB_APP_ID=123456
export GITHUB_INSTALLATION_ID=789012

# Or use .env file
cp configs/.env.example configs/.env
# Edit .env with your values
source configs/.env
```

## Pattern Matching Issues

### Files Not Matching Pattern

**Symptom:** Metrics show `"matched": 0` even though files were changed.

**Debug Steps:**

1. **Check actual file paths:**
   ```bash
   # Look for "sample file path" in logs
   grep "sample file path" logs/app.log
   ```

2. **Test your pattern:**
   ```bash
   ./config-validator test-pattern \
     -type regex \
     -pattern "^examples/(?P<lang>[^/]+)/(?P<file>.+)$" \
     -file "examples/go/main.go"
   ```

3. **Common issues:**
   - Missing `^` or `$` anchors in regex
   - Wrong pattern type (prefix vs glob vs regex)
   - Pattern doesn't match actual file structure
   - Typos in the pattern

**Example Fix:**
```yaml
# ❌ Wrong - doesn't match actual paths
pattern: "^examples/(?P<lang>go)/(?P<file>.+)$"

# ✅ Right - matches any language
pattern: "^examples/(?P<lang>[^/]+)/(?P<file>.+)$"
```

### Variables Not Extracted

**Symptom:** Path transformation fails or uses wrong paths.

**Debug Steps:**

1. **Check for named capture groups:**
   ```yaml
   # ❌ Wrong - unnamed groups
   pattern: "^examples/([^/]+)/(.+)$"
   
   # ✅ Right - named groups
   pattern: "^examples/(?P<lang>[^/]+)/(?P<file>.+)$"
   ```

2. **Test variable extraction:**
   ```bash
   ./config-validator test-pattern \
     -type regex \
     -pattern "^examples/(?P<lang>[^/]+)/(?P<file>.+)$" \
     -file "examples/go/main.go"
   ```

3. **Verify variable names match:**
   ```yaml
   # Pattern extracts "lang"
   pattern: "^examples/(?P<lang>[^/]+)/(?P<file>.+)$"
   
   # Transform must use "lang" (not "language")
   path_transform: "docs/${lang}/${file}"
   ```

### Path Transformation Fails

**Symptom:** Files copied to wrong location.

**Debug Steps:**

1. **Test transformation:**
   ```bash
   ./config-validator test-transform \
     -source "examples/go/main.go" \
     -template "docs/${lang}/${file}" \
     -vars "lang=go,file=main.go"
   ```

2. **Check variable names:**
   - Variables in template must match extracted variables
   - Use built-in variables: `${path}`, `${filename}`, `${dir}`, `${ext}`

3. **Verify template syntax:**
   ```yaml
   # ✅ Correct
   path_transform: "docs/${lang}/${file}"
   
   # ❌ Wrong - missing ${}
   path_transform: "docs/lang/file"
   ```

## Webhook Issues

### Webhook Returns 401 Unauthorized

**Cause:** Webhook signature verification failed.

**Solutions:**

1. **Check webhook secret:**
   ```bash
   echo $WEBHOOK_SECRET
   # Should match GitHub webhook secret
   ```

2. **For local testing, disable signature verification:**
   ```bash
   unset WEBHOOK_SECRET
   # Or set to empty
   export WEBHOOK_SECRET=""
   ```

3. **Verify GitHub webhook configuration:**
   - Go to GitHub → Settings → Webhooks
   - Check secret matches `WEBHOOK_SECRET`
   - Ensure content type is `application/json`

### Webhook Returns 500 Internal Server Error

**Debug Steps:**

1. **Check application logs:**
   ```bash
   # Local
   tail -f logs/app.log
   
   # GCP
   gcloud app logs tail -s default
   ```

2. **Check for common errors:**
   - MongoDB connection failed
   - GitHub API rate limit
   - Invalid configuration
   - Missing environment variables

3. **Test with dry-run mode:**
   ```bash
   DRY_RUN=true ./examples-copier
   ```

### No Response from Webhook

**Debug Steps:**

1. **Verify app is running:**
   ```bash
   curl http://localhost:8080/health
   ```

2. **Check webhook URL:**
   ```bash
   # Should be: http://your-domain:8080/webhook
   # Or whatever WEBSERVER_PATH is set to
   ```

3. **Test with curl:**
   ```bash
   ./test-webhook -payload test-payloads/example-pr-merged.json
   ```

## Deployment Issues

### Google Cloud Logging Error

**Error:**
```
[ERROR] failed to create cloud logging client: invalid project ID
```

**Solution:** Disable cloud logging for local testing:
```bash
export COPIER_DISABLE_CLOUD_LOGGING=true
./examples-copier
```

### MongoDB Connection Failed

**Error:**
```
[ERROR] failed to connect to MongoDB: connection timeout
```

**Solutions:**

1. **Check MongoDB URI:**
   ```bash
   echo $MONGO_URI
   # Should be: mongodb+srv://user:pass@cluster.mongodb.net
   ```

2. **Verify IP whitelist:**
   - Go to MongoDB Atlas → Network Access
   - Add your IP or use `0.0.0.0/0` for testing

3. **Disable audit logging for testing:**
   ```bash
   export AUDIT_ENABLED=false
   ```

### GitHub API Rate Limit

**Error:**
```
[ERROR] GitHub API rate limit exceeded
```

**Solutions:**

1. **Check rate limit status:**
   ```bash
   curl -H "Authorization: token $GITHUB_TOKEN" \
     https://api.github.com/rate_limit
   ```

2. **Wait for reset** or **use authenticated requests**

3. **For testing, use dry-run mode:**
   ```bash
   DRY_RUN=true ./examples-copier
   ```

## Slack Notification Issues

### No Slack Notifications

**Debug Steps:**

1. **Check webhook URL:**
   ```bash
   echo $SLACK_WEBHOOK_URL
   # Should start with: https://hooks.slack.com/services/
   ```

2. **Test webhook directly:**
   ```bash
   curl -X POST -H 'Content-type: application/json' \
     --data '{"text":"Test"}' \
     "$SLACK_WEBHOOK_URL"
   ```

3. **Check if Slack is enabled:**
   ```bash
   echo $SLACK_ENABLED
   # Should be: true
   ```

4. **Check application logs:**
   ```bash
   grep "slack" logs/app.log
   ```

### Slack Notifications in Wrong Channel

**Solution:**

The webhook URL determines the default channel. To override:
```bash
export SLACK_CHANNEL="#your-channel"
```

**Note:** The bot must have permission to post to that channel.

### Slack Notifications Too Noisy

**Solutions:**

1. **Use dedicated channel:**
   ```bash
   export SLACK_CHANNEL="#copier-notifications"
   ```

2. **Adjust Slack channel notification settings**

3. **Disable in development:**
   ```bash
   export SLACK_ENABLED=false
   ```

## Performance Issues

### Slow Webhook Processing

**Debug Steps:**

1. **Check metrics:**
   ```bash
   curl http://localhost:8080/metrics | jq '.webhooks.processing_time'
   ```

2. **Common causes:**
   - Large number of files in PR
   - Slow GitHub API responses
   - MongoDB connection latency
   - Complex regex patterns

3. **Optimize:**
   - Use simpler patterns when possible
   - Enable connection pooling
   - Increase timeout values

### High Memory Usage

**Debug Steps:**

1. **Check for memory leaks:**
   ```bash
   # Monitor memory usage
   top -p $(pgrep examples-copier)
   ```

2. **Common causes:**
   - Large file contents in memory
   - Unclosed connections
   - Goroutine leaks

3. **Solutions:**
   - Restart application periodically
   - Review audit logs for patterns
   - Check for stuck webhooks

## Debugging Tips

### Enable Debug Logging

```bash
export LOG_LEVEL=debug
./examples-copier
```

### Check Health Endpoint

```bash
curl http://localhost:8080/health | jq
```

### Check Metrics

```bash
curl http://localhost:8080/metrics | jq
```

### Test Configuration

```bash
./config-validator validate -config copier-config.yaml -v
```

### Test Pattern Matching

```bash
./config-validator test-pattern \
  -type regex \
  -pattern "YOUR_PATTERN" \
  -file "ACTUAL_FILE_PATH"
```

### Test Path Transformation

```bash
./config-validator test-transform \
  -source "SOURCE_PATH" \
  -template "TEMPLATE" \
  -vars "key1=value1,key2=value2"
```

### Test with Dry-Run Mode

```bash
DRY_RUN=true ./examples-copier &
./test-webhook -payload test-payloads/example-pr-merged.json
```

### Check Audit Logs

```javascript
// MongoDB shell
use code_copier
db.audit_events.find().sort({timestamp: -1}).limit(10).pretty()
```

### Trace a Specific Request

```bash
# Look for request ID in logs
grep "request_id=abc123" logs/app.log
```

## Getting Help

If you can't resolve the issue:

1. **Check [FAQ](FAQ.md)** for common questions
2. **Review [documentation](../README.md)** for your use case
3. **Search existing issues** on GitHub
4. **Open a new issue** with:
   - Error message
   - Configuration (sanitized)
   - Steps to reproduce
   - Logs (sanitized)

## See Also

- [FAQ](FAQ.md) - Frequently asked questions
- [Pattern Matching Guide](PATTERN-MATCHING-GUIDE.md) - Pattern matching help
- [Local Testing](LOCAL-TESTING.md) - Testing locally

