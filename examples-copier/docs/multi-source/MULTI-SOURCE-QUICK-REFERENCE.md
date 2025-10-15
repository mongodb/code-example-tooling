# Multi-Source Support - Quick Reference Guide

## Overview

This guide provides quick reference information for working with multi-source repository configurations.

## Configuration Format

### Single Source (Legacy)

```yaml
source_repo: "mongodb/docs-code-examples"
source_branch: "main"
copy_rules:
  - name: "example"
    # ... rules
```

### Multi-Source (New)

```yaml
sources:
  - repo: "mongodb/docs-code-examples"
    branch: "main"
    installation_id: "12345678"  # Optional
    copy_rules:
      - name: "example"
        # ... rules
```

## Key Concepts

### Source Repository
- The repository being monitored for changes
- Identified by `owner/repo` format (e.g., `mongodb/docs-code-examples`)
- Each source can have its own copy rules

### Installation ID
- GitHub App installation identifier
- Different organizations require different installation IDs
- Optional: defaults to `INSTALLATION_ID` environment variable

### Copy Rules
- Define which files to copy and where
- Each source can have multiple copy rules
- Rules are evaluated independently per source

## Common Tasks

### Add a New Source Repository

```yaml
sources:
  # Existing sources...
  
  # Add new source
  - repo: "mongodb/new-repo"
    branch: "main"
    installation_id: "99887766"
    copy_rules:
      - name: "new-rule"
        source_pattern:
          type: "prefix"
          pattern: "examples/"
        targets:
          - repo: "mongodb/target"
            branch: "main"
            path_transform: "code/${path}"
            commit_strategy:
              type: "pull_request"
              pr_title: "Update examples"
              auto_merge: false
```

### Configure Multiple Targets

```yaml
sources:
  - repo: "mongodb/source"
    branch: "main"
    copy_rules:
      - name: "multi-target"
        source_pattern:
          type: "glob"
          pattern: "**/*.go"
        targets:
          # Target 1
          - repo: "mongodb/target1"
            branch: "main"
            path_transform: "examples/${filename}"
            commit_strategy:
              type: "direct"
          
          # Target 2
          - repo: "mongodb/target2"
            branch: "develop"
            path_transform: "code/${filename}"
            commit_strategy:
              type: "pull_request"
              pr_title: "Update examples"
              auto_merge: false
```

### Set Global Defaults

```yaml
sources:
  - repo: "mongodb/source1"
    # ... config
  - repo: "mongodb/source2"
    # ... config

# Apply to all sources unless overridden
defaults:
  commit_strategy:
    type: "pull_request"
    auto_merge: false
  deprecation_check:
    enabled: true
    file: "deprecated_examples.json"
```

### Cross-Organization Copying

```yaml
sources:
  # Source from mongodb org
  - repo: "mongodb/public-examples"
    branch: "main"
    installation_id: "11111111"
    copy_rules:
      - name: "to-internal"
        source_pattern:
          type: "prefix"
          pattern: "public/"
        targets:
          # Target in 10gen org (requires different installation)
          - repo: "10gen/internal-docs"
            branch: "main"
            path_transform: "examples/${path}"
            commit_strategy:
              type: "direct"
```

## Validation

### Validate Configuration

```bash
# Validate syntax and logic
./config-validator validate -config copier-config.yaml -v

# Check specific source
./config-validator validate-source \
  -config copier-config.yaml \
  -source "mongodb/docs-code-examples"
```

### Test Pattern Matching

```bash
# Test if a file matches patterns
./config-validator test-pattern \
  -config copier-config.yaml \
  -source "mongodb/docs-code-examples" \
  -file "examples/go/main.go"
```

### Test Path Transformation

```bash
# Test path transformation
./config-validator test-transform \
  -config copier-config.yaml \
  -source "mongodb/docs-code-examples" \
  -file "examples/go/main.go"
```

## Monitoring

### Health Check

```bash
# Check application health
curl http://localhost:8080/health | jq

# Check specific source
curl http://localhost:8080/health | jq '.sources["mongodb/docs-code-examples"]'
```

### Metrics

```bash
# Get all metrics
curl http://localhost:8080/metrics | jq

# Get metrics for specific source
curl http://localhost:8080/metrics | jq '.by_source["mongodb/docs-code-examples"]'
```

### Logs

```bash
# Filter logs by source
gcloud app logs read --filter='jsonPayload.source_repo="mongodb/docs-code-examples"'

# Filter by operation
gcloud app logs read --filter='jsonPayload.operation="webhook_received"'
```

## Troubleshooting

### Webhook Not Processing

**Check 1: Is source configured?**
```bash
./config-validator list-sources -config copier-config.yaml
```

**Check 2: Is webhook signature valid?**
```bash
# Check logs for signature validation errors
gcloud app logs read --filter='jsonPayload.error=~"signature"'
```

**Check 3: Is installation ID correct?**
```bash
# Verify installation ID
curl -H "Authorization: Bearer YOUR_JWT" \
  https://api.github.com/app/installations
```

### Files Not Copying

**Check 1: Do files match patterns?**
```bash
./config-validator test-pattern \
  -config copier-config.yaml \
  -source "mongodb/source" \
  -file "path/to/file.go"
```

**Check 2: Is path transformation correct?**
```bash
./config-validator test-transform \
  -config copier-config.yaml \
  -source "mongodb/source" \
  -file "path/to/file.go"
```

**Check 3: Check audit logs**
```bash
# Query MongoDB audit logs
db.audit_events.find({
  source_repo: "mongodb/source",
  success: false
}).sort({timestamp: -1}).limit(10)
```

### Installation Authentication Errors

**Check 1: Verify installation ID**
```yaml
sources:
  - repo: "mongodb/source"
    installation_id: "12345678"  # Verify this is correct
```

**Check 2: Check token expiry**
```bash
# Tokens are cached for 1 hour
# Check logs for token refresh
gcloud app logs read --filter='jsonPayload.operation="token_refresh"'
```

**Check 3: Verify app permissions**
- Go to GitHub App settings
- Check installation has required permissions
- Verify app is installed on the repository

## Environment Variables

### Required

```bash
# GitHub App Configuration
GITHUB_APP_ID=123456
INSTALLATION_ID=12345678  # Default installation ID

# Google Cloud
GCP_PROJECT_ID=your-project
PEM_KEY_NAME=projects/123/secrets/pem/versions/latest
WEBHOOK_SECRET_NAME=projects/123/secrets/webhook/versions/latest

# Application
PORT=8080
CONFIG_FILE=copier-config.yaml
```

### Optional

```bash
# Dry Run Mode
DRY_RUN=false

# Audit Logging
AUDIT_ENABLED=true
MONGO_URI=mongodb+srv://...
AUDIT_DATABASE=copier_audit
AUDIT_COLLECTION=events

# Metrics
METRICS_ENABLED=true

# Slack Notifications
SLACK_WEBHOOK_URL=https://hooks.slack.com/...
SLACK_CHANNEL=#copier-alerts
```

## Best Practices

### 1. Use Descriptive Rule Names

```yaml
# Good
- name: "go-examples-to-docs"

# Bad
- name: "rule1"
```

### 2. Test Before Deploying

```bash
# Always validate
./config-validator validate -config copier-config.yaml -v

# Test in dry-run mode
DRY_RUN=true ./examples-copier
```

### 3. Monitor Per Source

```yaml
# Enable metrics for each source
sources:
  - repo: "mongodb/source"
    settings:
      enabled: true
      # Monitor this source specifically
```

### 4. Use Pull Requests for Production

```yaml
# Safer for production
commit_strategy:
  type: "pull_request"
  auto_merge: false  # Require review
```

### 5. Enable Deprecation Tracking

```yaml
# Track deleted files
deprecation_check:
  enabled: true
  file: "deprecated_examples.json"
```

### 6. Set Appropriate Timeouts

```yaml
sources:
  - repo: "mongodb/large-repo"
    settings:
      timeout_seconds: 300  # 5 minutes for large repos
```

### 7. Use Rate Limiting

```yaml
sources:
  - repo: "mongodb/high-volume-repo"
    settings:
      rate_limit:
        max_webhooks_per_minute: 10
        max_concurrent: 3
```

## Migration Checklist

- [ ] Backup current configuration
- [ ] Convert to multi-source format
- [ ] Validate new configuration
- [ ] Test in dry-run mode
- [ ] Deploy to staging
- [ ] Test with real webhooks
- [ ] Monitor metrics and logs
- [ ] Deploy to production
- [ ] Decommission old deployments

## Quick Commands

```bash
# Validate config
./config-validator validate -config copier-config.yaml -v

# Convert legacy to multi-source
./config-validator convert-to-multi-source \
  -input copier-config.yaml \
  -output copier-config-multi.yaml

# Test pattern matching
./config-validator test-pattern \
  -config copier-config.yaml \
  -source "mongodb/source" \
  -file "examples/go/main.go"

# Dry run
DRY_RUN=true ./examples-copier

# Check health
curl http://localhost:8080/health | jq

# Get metrics
curl http://localhost:8080/metrics | jq

# View logs
gcloud app logs tail -s default

# Deploy
gcloud app deploy
```

## Support Resources

- [Implementation Plan](MULTI-SOURCE-IMPLEMENTATION-PLAN.md)
- [Technical Specification](MULTI-SOURCE-TECHNICAL-SPEC.md)
- [Migration Guide](MULTI-SOURCE-MIGRATION-GUIDE.md)
- [Configuration Guide](CONFIGURATION-GUIDE.md)
- [Troubleshooting Guide](TROUBLESHOOTING.md)

## Common Patterns

### Pattern 1: Single Source, Multiple Targets

```yaml
sources:
  - repo: "mongodb/source"
    branch: "main"
    copy_rules:
      - name: "to-multiple-targets"
        source_pattern:
          type: "glob"
          pattern: "**/*.go"
        targets:
          - repo: "mongodb/target1"
            # ... config
          - repo: "mongodb/target2"
            # ... config
          - repo: "mongodb/target3"
            # ... config
```

### Pattern 2: Multiple Sources, Single Target

```yaml
sources:
  - repo: "mongodb/source1"
    branch: "main"
    copy_rules:
      - name: "from-source1"
        source_pattern:
          type: "prefix"
          pattern: "examples/"
        targets:
          - repo: "mongodb/target"
            path_transform: "source1/${path}"
            # ... config
  
  - repo: "mongodb/source2"
    branch: "main"
    copy_rules:
      - name: "from-source2"
        source_pattern:
          type: "prefix"
          pattern: "examples/"
        targets:
          - repo: "mongodb/target"
            path_transform: "source2/${path}"
            # ... config
```

### Pattern 3: Cross-Organization with Different Strategies

```yaml
sources:
  # Public repo - use PRs
  - repo: "mongodb/public-examples"
    branch: "main"
    installation_id: "11111111"
    copy_rules:
      - name: "public-to-docs"
        source_pattern:
          type: "prefix"
          pattern: "examples/"
        targets:
          - repo: "mongodb/docs"
            branch: "main"
            path_transform: "code/${path}"
            commit_strategy:
              type: "pull_request"
              auto_merge: false
  
  # Internal repo - direct commits
  - repo: "10gen/internal-examples"
    branch: "main"
    installation_id: "22222222"
    copy_rules:
      - name: "internal-to-docs"
        source_pattern:
          type: "prefix"
          pattern: "examples/"
        targets:
          - repo: "10gen/internal-docs"
            branch: "main"
            path_transform: "code/${path}"
            commit_strategy:
              type: "direct"
```

