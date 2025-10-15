# Migration Guide: Single Source to Multi-Source Configuration

This guide helps you migrate from the legacy single-source configuration format to the new multi-source format.

## Table of Contents

- [Overview](#overview)
- [Backward Compatibility](#backward-compatibility)
- [Migration Steps](#migration-steps)
- [Configuration Comparison](#configuration-comparison)
- [Testing Your Migration](#testing-your-migration)
- [Rollback Plan](#rollback-plan)
- [FAQ](#faq)

## Overview

The multi-source feature allows the examples-copier to monitor and process webhooks from multiple source repositories in a single deployment. This eliminates the need to run separate copier instances for different source repositories.

### Benefits of Multi-Source

- **Simplified Deployment**: One instance handles multiple source repositories
- **Centralized Configuration**: Manage all copy rules in one place
- **Better Resource Utilization**: Shared infrastructure for all sources
- **Consistent Monitoring**: Unified metrics and audit logging
- **Cross-Organization Support**: Handle repos from different GitHub organizations

## Backward Compatibility

**Good News**: The new multi-source format is 100% backward compatible with existing configurations.

- ✅ Existing single-source configs continue to work without changes
- ✅ No breaking changes to the configuration schema
- ✅ Automatic detection of legacy vs. new format
- ✅ Gradual migration path available

## Migration Steps

### Step 1: Assess Your Current Setup

First, identify all the source repositories you're currently monitoring:

```bash
# List all your current copier deployments
# Each deployment typically monitors one source repository
```

**Example Current State:**
- Deployment 1: Monitors `mongodb/docs-code-examples`
- Deployment 2: Monitors `mongodb/atlas-examples`
- Deployment 3: Monitors `10gen/internal-examples`

### Step 2: Backup Current Configuration

```bash
# Backup your current configuration
cp copier-config.yaml copier-config.yaml.backup

# Backup environment variables
cp .env .env.backup
```

### Step 3: Convert Configuration Format

#### Option A: Manual Conversion

**Before (Single Source):**
```yaml
source_repo: "mongodb/docs-code-examples"
source_branch: "main"

copy_rules:
  - name: "go-examples"
    source_pattern:
      type: "prefix"
      pattern: "examples/go/"
    targets:
      - repo: "mongodb/docs"
        branch: "main"
        path_transform: "code/go/${path}"
        commit_strategy:
          type: "pull_request"
          pr_title: "Update Go examples"
          auto_merge: false
```

**After (Multi-Source):**
```yaml
sources:
  - repo: "mongodb/docs-code-examples"
    branch: "main"
    # Optional: Add installation_id if different from default
    # installation_id: "12345678"
    
    copy_rules:
      - name: "go-examples"
        source_pattern:
          type: "prefix"
          pattern: "examples/go/"
        targets:
          - repo: "mongodb/docs"
            branch: "main"
            path_transform: "code/go/${path}"
            commit_strategy:
              type: "pull_request"
              pr_title: "Update Go examples"
              auto_merge: false
```

#### Option B: Automated Conversion (Recommended)

Use the config-validator tool to automatically convert your configuration:

```bash
# Convert single-source to multi-source format
./config-validator convert-to-multi-source \
  -input copier-config.yaml \
  -output copier-config-multi.yaml

# Validate the new configuration
./config-validator validate -config copier-config-multi.yaml -v
```

### Step 4: Consolidate Multiple Deployments

If you have multiple copier deployments, consolidate them into one multi-source config:

```yaml
sources:
  # Source 1: From deployment 1
  - repo: "mongodb/docs-code-examples"
    branch: "main"
    installation_id: "12345678"
    copy_rules:
      # ... copy rules from deployment 1
  
  # Source 2: From deployment 2
  - repo: "mongodb/atlas-examples"
    branch: "main"
    installation_id: "87654321"
    copy_rules:
      # ... copy rules from deployment 2
  
  # Source 3: From deployment 3
  - repo: "10gen/internal-examples"
    branch: "main"
    installation_id: "11223344"
    copy_rules:
      # ... copy rules from deployment 3
```

### Step 5: Update Environment Variables

Update your `.env` file to support multiple installations:

```bash
# Before (single installation)
INSTALLATION_ID=12345678

# After (default installation + optional per-source)
INSTALLATION_ID=12345678  # Default/fallback installation ID

# Note: Per-source installation IDs are now in the config file
# under each source's installation_id field
```

### Step 6: Update GitHub App Installations

Ensure your GitHub App is installed on all source repositories:

1. Go to your GitHub App settings
2. Install the app on each source repository's organization
3. Note the installation ID for each organization
4. Add installation IDs to your config file

```bash
# Get installation IDs
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  https://api.github.com/app/installations
```

### Step 7: Validate Configuration

Before deploying, validate your new configuration:

```bash
# Validate configuration syntax and logic
./config-validator validate -config copier-config-multi.yaml -v

# Test pattern matching
./config-validator test-pattern \
  -config copier-config-multi.yaml \
  -source "mongodb/docs-code-examples" \
  -file "examples/go/main.go"

# Dry-run test
./examples-copier -config copier-config-multi.yaml -dry-run
```

### Step 8: Deploy and Test

1. **Deploy to staging first**:
```bash
# Deploy to staging environment
gcloud app deploy --project=your-staging-project
```

2. **Test with real webhooks**:
```bash
# Use the test-webhook tool
./test-webhook -config copier-config-multi.yaml \
  -payload test-payloads/example-pr-merged.json
```

3. **Monitor logs**:
```bash
# Watch application logs
gcloud app logs tail -s default
```

4. **Verify metrics**:
```bash
# Check health endpoint
curl https://your-app.appspot.com/health

# Check metrics endpoint
curl https://your-app.appspot.com/metrics
```

### Step 9: Production Deployment

Once validated in staging:

```bash
# Deploy to production
gcloud app deploy --project=your-production-project

# Monitor for issues
gcloud app logs tail -s default --project=your-production-project
```

### Step 10: Decommission Old Deployments

After confirming the multi-source deployment works:

1. Monitor for 24-48 hours
2. Verify all source repositories are being processed
3. Check audit logs for any errors
4. Decommission old single-source deployments

## Configuration Comparison

### Single Source (Legacy)

```yaml
source_repo: "mongodb/docs-code-examples"
source_branch: "main"

copy_rules:
  - name: "example-rule"
    source_pattern:
      type: "prefix"
      pattern: "examples/"
    targets:
      - repo: "mongodb/docs"
        branch: "main"
        path_transform: "code/${path}"
        commit_strategy:
          type: "direct"
```

### Multi-Source (New)

```yaml
sources:
  - repo: "mongodb/docs-code-examples"
    branch: "main"
    installation_id: "12345678"  # Optional
    copy_rules:
      - name: "example-rule"
        source_pattern:
          type: "prefix"
          pattern: "examples/"
        targets:
          - repo: "mongodb/docs"
            branch: "main"
            path_transform: "code/${path}"
            commit_strategy:
              type: "direct"

# Optional: Global defaults
defaults:
  commit_strategy:
    type: "pull_request"
    auto_merge: false
  deprecation_check:
    enabled: true
```

### Hybrid (Both Formats Supported)

The application automatically detects which format you're using:

```go
// Automatic detection logic
if config.SourceRepo != "" {
    // Legacy single-source format
    processSingleSource(config)
} else if len(config.Sources) > 0 {
    // New multi-source format
    processMultiSource(config)
}
```

## Testing Your Migration

### Test Checklist

- [ ] Configuration validates successfully
- [ ] Pattern matching works for all sources
- [ ] Path transformations are correct
- [ ] Webhooks route to correct source config
- [ ] GitHub authentication works for all installations
- [ ] Files are copied to correct target repositories
- [ ] Deprecation tracking works (if enabled)
- [ ] Metrics show data for all sources
- [ ] Audit logs contain source repository info
- [ ] Slack notifications work (if enabled)

### Test Commands

```bash
# 1. Validate configuration
./config-validator validate -config copier-config-multi.yaml -v

# 2. Test pattern matching for each source
./config-validator test-pattern \
  -config copier-config-multi.yaml \
  -source "mongodb/docs-code-examples" \
  -file "examples/go/main.go"

# 3. Dry-run mode
DRY_RUN=true ./examples-copier -config copier-config-multi.yaml

# 4. Test with webhook payload
./test-webhook -config copier-config-multi.yaml \
  -payload test-payloads/multi-source-webhook.json

# 5. Check health
curl http://localhost:8080/health

# 6. Check metrics
curl http://localhost:8080/metrics
```

## Rollback Plan

If you encounter issues after migration:

### Quick Rollback

```bash
# 1. Restore backup configuration
cp copier-config.yaml.backup copier-config.yaml
cp .env.backup .env

# 2. Redeploy previous version
gcloud app deploy --version=previous-version

# 3. Route traffic back
gcloud app services set-traffic default --splits=previous-version=1
```

### Gradual Rollback

```bash
# Route 50% traffic to old version
gcloud app services set-traffic default \
  --splits=new-version=0.5,previous-version=0.5

# Monitor and adjust as needed
```

## FAQ

### Q: Do I need to migrate immediately?

**A:** No. The legacy single-source format is fully supported and will continue to work. Migrate when you need to monitor multiple source repositories or want to consolidate deployments.

### Q: Can I mix legacy and new formats?

**A:** No. Each configuration file must use either the legacy format OR the new format, not both. However, you can have different deployments using different formats.

### Q: What happens if I don't specify installation_id?

**A:** The application will use the default `INSTALLATION_ID` from environment variables. This works fine if all your source repositories are in the same organization.

### Q: Can I gradually migrate one source at a time?

**A:** Yes. You can start with one source in the new format and add more sources over time. Keep your old deployments running until all sources are migrated.

### Q: How do I test without affecting production?

**A:** Use dry-run mode (`DRY_RUN=true`) to test configuration without making actual commits. Also test in a staging environment first.

### Q: What if a webhook comes from an unknown source?

**A:** The application will log a warning and return a 204 No Content response. No processing will occur. Check your configuration to ensure all expected sources are listed.

### Q: Can different sources target the same repository?

**A:** Yes! Multiple sources can target the same repository with different copy rules. The application handles this correctly.

### Q: How are metrics tracked for multiple sources?

**A:** Metrics are tracked both globally and per-source. Use the `/metrics` endpoint to see breakdown by source repository.

## Support

If you encounter issues during migration:

1. Check the [Troubleshooting Guide](TROUBLESHOOTING.md)
2. Review application logs for errors
3. Use the config-validator tool to identify issues
4. Consult the [Multi-Source Implementation Plan](MULTI-SOURCE-IMPLEMENTATION-PLAN.md)

## Next Steps

After successful migration:

1. Monitor metrics and audit logs
2. Optimize copy rules for performance
3. Consider enabling additional features (Slack notifications, etc.)
4. Document your specific configuration for your team
5. Set up alerts for failures

