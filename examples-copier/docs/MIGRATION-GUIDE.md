# Migration Guide

This guide helps you migrate from the legacy JSON configuration to the new YAML configuration with enhanced features.

## Overview

The refactored examples-copier application maintains **backward compatibility** with legacy JSON configs while offering powerful new features through YAML configuration.

## Migration Options

### Option 1: Continue Using Legacy Config (No Changes Required)

Your existing `config.json` will continue to work without any changes:

```json
[
  {
    "source_directory": "examples",
    "target_repo": "org/target",
    "target_branch": "main",
    "target_directory": "docs",
    "recursive_copy": true,
    "copier_commit_strategy": "pr",
    "pr_title": "Update docs",
    "commit_message": "Sync from source",
    "merge_without_review": false
  }
]
```

**What happens:**
- Legacy config is automatically converted to new format internally
- All existing functionality continues to work
- No migration required

### Option 2: Migrate to YAML (Recommended)

Migrate to YAML to access new features like pattern matching, path transformations, and message templating.

## Step-by-Step Migration

### Step 1: Convert Your Config

Use the CLI tool to convert your existing config:

```bash
./config-validator convert -input config.json -output copier-config.yaml
```

### Step 2: Review the Converted Config

The tool generates a YAML config with equivalent functionality:

**Before (JSON):**
```json
[
  {
    "source_directory": "examples",
    "target_repo": "org/target",
    "target_branch": "main",
    "target_directory": "docs",
    "recursive_copy": true
  }
]
```

**After (YAML):**
```yaml
source_repo: "org/source"
source_branch: "main"

copy_rules:
  - name: "legacy-rule-0"
    source_pattern:
      type: "prefix"
      pattern: "examples"
    targets:
      - repo: "org/target"
        branch: "main"
        path_transform: "docs/${path}"
        commit_strategy:
          type: "direct"
```

### Step 3: Enhance with New Features

Now you can add advanced features:

#### Add Pattern Matching

Replace simple prefix patterns with regex for more control:

```yaml
copy_rules:
  - name: "Copy Go examples"
    source_pattern:
      type: "regex"
      pattern: "^examples/(?P<lang>[^/]+)/(?P<category>[^/]+)/(?P<file>.+)$"
    targets:
      - repo: "org/target"
        branch: "main"
        path_transform: "docs/${lang}/${category}/${file}"
```

#### Add Message Templates

Use variables in commit messages:

```yaml
commit_strategy:
  type: "pull_request"
  commit_message: "Update ${category} examples from ${lang}"
  pr_title: "Update ${category} examples"
  auto_merge: false
```

#### Add Deprecation Tracking

Enable automatic deprecation tracking:

```yaml
deprecation_check:
  enabled: true
  file: "deprecated_examples.json"
```

### Step 4: Validate Your Config

Test your new config before deploying:

```bash
# Validate syntax and structure
./config-validator validate -config copier-config.yaml -v

# Test pattern matching
./config-validator test-pattern \
  -type regex \
  -pattern "^examples/(?P<lang>[^/]+)/(?P<file>.+)$" \
  -file "examples/go/main.go"

# Test path transformation
./config-validator test-transform \
  -template "docs/${lang}/${file}" \
  -file "examples/go/main.go" \
  -pattern "^examples/(?P<lang>[^/]+)/(?P<file>.+)$"
```

### Step 5: Test in Dry-Run Mode

Test the new config without making actual changes:

```bash
DRY_RUN=true ./examples-copier -env ./configs/.env
```

Monitor the logs to ensure files are matched and transformed correctly.

### Step 6: Deploy

Once validated, update your environment to use the new config:

```bash
# Update CONFIG_FILE in .env
CONFIG_FILE=copier-config.yaml

# Deploy
./examples-copier
```

## Feature Comparison

### Legacy JSON Config

**Capabilities:**
- ✅ Simple directory-based copying
- ✅ Recursive or non-recursive copy
- ✅ Direct commit or PR strategy
- ✅ Custom commit messages and PR titles
- ✅ Auto-merge PRs

**Limitations:**
- ❌ No pattern matching (only directory prefixes)
- ❌ No path transformations
- ❌ No variable substitution in messages
- ❌ No regex support
- ❌ Limited flexibility

### New YAML Config

**All legacy features plus:**
- ✅ **Pattern Matching** - Prefix, glob, and regex patterns
- ✅ **Path Transformations** - Template-based with variables
- ✅ **Message Templating** - Dynamic commit messages and PR titles
- ✅ **Variable Extraction** - Named groups from regex patterns
- ✅ **Multiple Targets** - Copy to multiple repos with different transforms
- ✅ **Deprecation Tracking** - Automatic tracking per target
- ✅ **Audit Logging** - MongoDB-based event tracking
- ✅ **Enhanced Validation** - Comprehensive config validation

## Migration Examples

### Example 1: Simple Directory Copy

**Legacy JSON:**
```json
[
  {
    "source_directory": "examples/go",
    "target_repo": "org/docs",
    "target_branch": "main",
    "target_directory": "code-examples/go",
    "recursive_copy": true
  }
]
```

**New YAML (Basic):**
```yaml
source_repo: "org/source"
copy_rules:
  - name: "Copy Go examples"
    source_pattern:
      type: "prefix"
      pattern: "examples/go/"
    targets:
      - repo: "org/docs"
        branch: "main"
        path_transform: "code-examples/go/${path}"
```

**New YAML (Enhanced):**
```yaml
source_repo: "org/source"
copy_rules:
  - name: "Copy Go examples"
    source_pattern:
      type: "regex"
      pattern: "^examples/go/(?P<category>[^/]+)/(?P<file>.+)$"
    targets:
      - repo: "org/docs"
        branch: "main"
        path_transform: "code-examples/go/${category}/${file}"
        commit_strategy:
          type: "pull_request"
          commit_message: "Update Go ${category} examples"
          pr_title: "Update Go ${category} examples"
```

### Example 2: Multiple Targets

**Legacy JSON:**
```json
[
  {
    "source_directory": "examples",
    "target_repo": "org/docs-v1",
    "target_directory": "examples",
    "recursive_copy": true
  },
  {
    "source_directory": "examples",
    "target_repo": "org/docs-v2",
    "target_directory": "code-samples",
    "recursive_copy": true
  }
]
```

**New YAML:**
```yaml
source_repo: "org/source"
copy_rules:
  - name: "Copy examples to multiple targets"
    source_pattern:
      type: "prefix"
      pattern: "examples/"
    targets:
      - repo: "org/docs-v1"
        branch: "main"
        path_transform: "examples/${path}"
      - repo: "org/docs-v2"
        branch: "main"
        path_transform: "code-samples/${path}"
```

### Example 3: Language-Specific Routing

**New YAML (Not possible with legacy):**
```yaml
source_repo: "org/source"
copy_rules:
  - name: "Route by language"
    source_pattern:
      type: "regex"
      pattern: "^examples/(?P<lang>[^/]+)/(?P<rest>.+)$"
    targets:
      - repo: "org/go-docs"
        branch: "main"
        path_transform: "examples/${rest}"
        condition: "${lang} == 'go'"
      - repo: "org/python-docs"
        branch: "main"
        path_transform: "examples/${rest}"
        condition: "${lang} == 'python'"
```

## Environment Variables

### New Variables

Add these to your `.env` file for new features:

```bash
# Audit Logging (Optional)
AUDIT_ENABLED=true
MONGO_URI=mongodb+srv://user:pass@cluster.mongodb.net
AUDIT_DATABASE=code_copier
AUDIT_COLLECTION=audit_events

# Metrics (Optional)
METRICS_ENABLED=true

# Development (Optional)
DRY_RUN=false
LOG_LEVEL=info
```

### Deprecated Variables

These still work but are superseded by config file settings:

- `COPIER_COMMIT_STRATEGY` - Use `commit_strategy.type` in config
- `DEFAULT_COMMIT_MESSAGE` - Use `commit_strategy.commit_message` in config
- `DEFAULT_RECURSIVE_COPY` - Pattern matching replaces this
- `DEFAULT_PR_MERGE` - Use `commit_strategy.auto_merge` in config

## Rollback Plan

If you need to rollback:

1. **Restore legacy config:**
   ```bash
   CONFIG_FILE=config.json ./examples-copier
   ```

2. **Use legacy README:**
   ```bash
   cat README.legacy.md
   ```

3. **Disable new features:**
   ```bash
   AUDIT_ENABLED=false
   METRICS_ENABLED=false
   ```

## Testing Your Migration

### 1. Validate Config
```bash
./config-validator validate -config copier-config.yaml -v
```

### 2. Test Pattern Matching
```bash
./config-validator test-pattern \
  -type regex \
  -pattern "your-pattern" \
  -file "test/file/path.go"
```

### 3. Dry-Run Test
```bash
DRY_RUN=true ./examples-copier
```

### 4. Monitor Metrics
```bash
curl http://localhost:8080/metrics
```

### 5. Check Audit Logs
```javascript
db.audit_events.find().sort({timestamp: -1}).limit(10)
```

## Getting Help

- **Documentation**: See [README.md](../README.md) for complete feature documentation
- **Examples**: Check [configs/config.example.yaml](configs/config.example.yaml)
- **CLI Help**: Run `./config-validator -help`

## Summary

- ✅ **Backward Compatible** - Legacy configs continue to work
- ✅ **Gradual Migration** - Migrate at your own pace
- ✅ **Enhanced Features** - Access powerful new capabilities
- ✅ **Easy Rollback** - Can revert if needed
- ✅ **Well Tested** - 51 unit tests covering all features

The migration is designed to be smooth and risk-free. Start with validation and dry-run testing before deploying to production.

