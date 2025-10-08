# Deployment Guide

This guide walks you through deploying the refactored examples-copier application with all new features.

## ✅ Integration Complete

The following features have been successfully integrated:

- ✅ Enhanced pattern matching (prefix, glob, regex)
- ✅ Path transformations with variable substitution
- ✅ YAML configuration support (with JSON backward compatibility)
- ✅ MongoDB audit logging
- ✅ Health and metrics endpoints
- ✅ Template-ized commit messages and PR titles
- ✅ Dry-run mode
- ✅ CLI validation tool
- ✅ ServiceContainer architecture

## Prerequisites

1. **Go 1.23.4+** installed
2. **MongoDB Atlas** account (for audit logging)
3. **GitHub App** credentials
4. **Google Cloud** project (for Secret Manager and logging)

## Step 1: Build the Application

```bash
cd examples-copier

# Build main application
go build -o examples-copier .

# Build CLI validator
go build -o config-validator ./cmd/config-validator

# Verify builds
./examples-copier -help
./config-validator -help
```

## Step 2: Configure Environment

Create or update your `.env` file:

```bash
# Copy example
cp configs/.env.example.new configs/.env

# Edit with your values
vim configs/.env
```

### Required Environment Variables

```bash
# GitHub Configuration
REPO_OWNER=your-org
REPO_NAME=your-repo
SRC_BRANCH=main
GITHUB_APP_ID=123456
GITHUB_INSTALLATION_ID=789012

# Google Cloud
GCP_PROJECT_ID=your-project
PEM_KEY_NAME=projects/123/secrets/CODE_COPIER_PEM/versions/latest

# Application Settings
PORT=8080
WEBSERVER_PATH=/webhook
CONFIG_FILE=config.yaml
DEPRECATION_FILE=deprecated_examples.json

# New Features
DRY_RUN=false
AUDIT_ENABLED=true
METRICS_ENABLED=true
WEBHOOK_SECRET=your-webhook-secret

# MongoDB (for audit logging)
MONGO_URI=mongodb+srv://user:pass@cluster.mongodb.net
AUDIT_DATABASE=code_copier
AUDIT_COLLECTION=audit_events
```

## Step 3: Create YAML Configuration

Create `config.yaml` in your repository:

```yaml
source_repo: "your-org/source-repo"
source_branch: "main"

copy_rules:
  - name: "Copy Go examples"
    source_pattern:
      type: "regex"
      pattern: "^examples/go/(?P<category>[^/]+)/(?P<file>.+)$"
    targets:
      - repo: "your-org/target-repo"
        branch: "main"
        path_transform: "docs/examples/${category}/${file}"
        commit_strategy:
          type: "pull_request"
          commit_message: "Update ${category} examples from source"
          pr_title: "Update ${category} examples"
          auto_merge: false
        deprecation_check:
          enabled: true
          file: "deprecated_examples.json"
```

### Validate Configuration

```bash
# Validate config file
./config-validator validate -config config.yaml -v

# Test pattern matching
./config-validator test-pattern \
  -type regex \
  -pattern "^examples/go/(?P<category>[^/]+)/(?P<file>.+)$" \
  -file "examples/go/database/connect.go"

# Test path transformation
./config-validator test-transform \
  -template "docs/examples/${category}/${file}" \
  -file "examples/go/database/connect.go" \
  -pattern "^examples/go/(?P<category>[^/]+)/(?P<file>.+)$"
```

## Step 4: Test with Dry-Run Mode

Before deploying to production, test with dry-run mode:

```bash
# Enable dry-run in .env
DRY_RUN=true ./examples-copier -env ./configs/.env
```

In dry-run mode:
- Webhooks are processed
- Files are matched and transformed
- Audit events are logged
- **NO actual commits or PRs are created**

## Step 5: Deploy to Google Cloud App Engine

### Update `app.yaml`

```yaml
runtime: go123
env: standard

env_variables:
  REPO_OWNER: "your-org"
  REPO_NAME: "your-repo"
  CONFIG_FILE: "config.yaml"
  AUDIT_ENABLED: "true"
  METRICS_ENABLED: "true"
  MONGO_URI: "mongodb+srv://..."
  # ... other variables

handlers:
  - url: /.*
    script: auto
    secure: always
```

### Deploy

```bash
# Deploy to App Engine
gcloud app deploy

# View logs
gcloud app logs tail -s default

# Check health
curl https://your-app.appspot.com/health
```

## Step 6: Configure GitHub Webhook

1. Go to your repository settings
2. Navigate to **Webhooks** → **Add webhook**
3. Set **Payload URL**: `https://your-app.appspot.com/webhook`
4. Set **Content type**: `application/json`
5. Set **Secret**: (your WEBHOOK_SECRET value)
6. Select events: **Pull requests**
7. Click **Add webhook**

## Step 7: Monitor and Verify

### Check Health Endpoint

```bash
curl https://your-app.appspot.com/health
```

Expected response:
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
  "uptime": "1h23m45s"
}
```

### Check Metrics Endpoint

```bash
curl https://your-app.appspot.com/metrics
```

Expected response:
```json
{
  "webhooks": {
    "received": 42,
    "processed": 40,
    "failed": 2,
    "success_rate": 95.24
  },
  "files": {
    "matched": 150,
    "uploaded": 145,
    "failed": 5,
    "deprecated": 3
  },
  "processing_time": {
    "p50": 234,
    "p95": 567,
    "p99": 890
  }
}
```

### Query Audit Logs

Connect to MongoDB and query audit events:

```javascript
// Recent successful copies
db.audit_events.find({
  event_type: "copy",
  success: true
}).sort({timestamp: -1}).limit(10)

// Failed operations
db.audit_events.find({
  success: false
}).sort({timestamp: -1})

// Statistics by rule
db.audit_events.aggregate([
  {$match: {event_type: "copy"}},
  {$group: {
    _id: "$rule_name",
    count: {$sum: 1},
    avg_duration: {$avg: "$duration_ms"}
  }}
])
```

## Step 8: Gradual Rollout

### Phase 1: Test with One Rule

Start with a single, simple copy rule:

```yaml
copy_rules:
  - name: "Test rule"
    source_pattern:
      type: "prefix"
      pattern: "test/examples/"
    targets:
      - repo: "your-org/test-repo"
        branch: "test-branch"
        path_transform: "${path}"
```

### Phase 2: Add More Rules

Gradually add more complex rules with regex patterns and transformations.

### Phase 3: Enable Auto-Merge

Once confident, enable auto-merge for specific rules:

```yaml
commit_strategy:
  type: "pull_request"
  auto_merge: true
```

## Troubleshooting

### Issue: Config validation fails

```bash
# Check config syntax
./config-validator validate -config config.yaml -v

# Test specific patterns
./config-validator test-pattern -type regex -pattern "..." -file "..."
```

### Issue: Files not matching

Check the audit logs for match attempts:

```javascript
db.audit_events.find({
  source_path: "your/file/path.go"
})
```

### Issue: MongoDB connection fails

```bash
# Test connection
mongosh "mongodb+srv://user:pass@cluster.mongodb.net/code_copier"

# Check environment variable
echo $MONGO_URI
```

### Issue: Webhook signature verification fails

```bash
# Verify webhook secret matches
echo $WEBHOOK_SECRET

# Check GitHub webhook delivery logs
# Go to Settings → Webhooks → Recent Deliveries
```

## Rollback Plan

If issues arise:

1. **Disable webhook** in GitHub repository settings
2. **Revert to previous version** using `gcloud app versions list` and `gcloud app services set-traffic`
3. **Check audit logs** to identify what was changed
4. **Fix configuration** and redeploy

## Performance Tuning

### Optimize Pattern Matching

- Use **prefix** patterns for simple directory matching (fastest)
- Use **glob** patterns for wildcard matching (medium)
- Use **regex** patterns only when necessary (slowest)

### Batch Operations

Group multiple file changes into single commits/PRs:

```yaml
commit_strategy:
  type: "pull_request"
  # All files matching this rule will be in one PR
```

### MongoDB Indexing

Ensure indexes exist for common queries:

```javascript
db.audit_events.createIndex({timestamp: -1})
db.audit_events.createIndex({rule_name: 1, timestamp: -1})
db.audit_events.createIndex({success: 1, timestamp: -1})
```

## Next Steps

1. ✅ Application deployed and running
2. ✅ Webhooks configured
3. ✅ Monitoring in place
4. 📝 Update main README.md with new features
5. 🧪 Write unit tests for new functionality
6. 📊 Set up alerting for failed operations

## Support

For issues or questions:
- Check `REFACTORING-SUMMARY.md` for feature documentation
- Review `INTEGRATION-GUIDE.md` for technical details
- Check audit logs in MongoDB
- Review application logs in Google Cloud

