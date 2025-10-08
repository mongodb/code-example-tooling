# Configuration Setup Guide

## The Issue You Encountered

```
[ERROR] failed to load config | {"error":"failed to retrieve config file: 
failed to get config file: GET https://api.github.com/repos/mongodb/docs-code-examples/contents/config.yaml?ref=main: 404 Not Found []"}
```

**Cause:** The app tries to fetch the config file from your **source repository** on GitHub, but it doesn't exist there yet.

## Solutions

### Option 1: Local Config File (For Testing - EASIEST)

The app now supports loading config from a **local file** for testing purposes.

**Steps:**

1. **Config file already created** - `config.yaml` exists in the `examples-copier/` directory

2. **Run the app** - It will automatically use the local file:
   ```bash
   make run-local-quick
   ```

3. **Test it:**
   ```bash
   # In another terminal
   ./test-webhook -payload test-payloads/example-pr-merged.json
   ```

**How it works:**
- App first tries to load `config.yaml` from the current directory
- If found, uses it (logs: "loaded config from local file")
- If not found, falls back to fetching from GitHub

### Option 2: Add Config to Source Repository (For Production)

For production use, the config should be in your source repository.

**Steps:**

1. **Find your source repository:**
   ```bash
   # From the error, your source repo is:
   # mongodb/docs-code-examples
   ```

2. **Copy the config file:**
   ```bash
   # Copy config.yaml to your source repo
   cp examples-copier/config.yaml /path/to/docs-code-examples/config.yaml
   ```

3. **Customize it** for your needs:
   ```yaml
   source_repo: "mongodb/docs-code-examples"  # Your source repo
   source_branch: "main"
   
   copy_rules:
     - name: "Copy examples"
       source_pattern:
         type: "prefix"
         pattern: "examples/"  # Adjust to your file structure
       targets:
         - repo: "mongodb/your-target-repo"  # Where to copy files
           branch: "main"
           path_transform: "docs/${path}"
   ```

4. **Commit and push:**
   ```bash
   cd /path/to/docs-code-examples
   git add config.yaml
   git commit -m "Add examples-copier configuration"
   git push origin main
   ```

5. **Deploy the app** - It will now fetch config from GitHub

## Configuration File Formats

### YAML Format (Recommended)

**File:** `config.yaml`

```yaml
source_repo: "mongodb/docs-code-examples"
source_branch: "main"

copy_rules:
  - name: "Copy Go examples"
    source_pattern:
      type: "regex"
      pattern: "^examples/(?P<lang>go)/(?P<category>[^/]+)/(?P<file>.+)$"
    targets:
      - repo: "mongodb/target-repo"
        branch: "main"
        path_transform: "docs/code-examples/${lang}/${category}/${file}"
        commit_strategy:
          type: "pull_request"
          commit_message: "Update ${category} examples"
          pr_title: "Update ${category} examples"
          auto_merge: false
```

### JSON Format (Legacy)

**File:** `config.json`

```json
[
  {
    "source_directory": "examples",
    "target_repo": "mongodb/target-repo",
    "target_branch": "main",
    "target_directory": "docs/code-examples",
    "recursive_copy": true,
    "copier_commit_strategy": "pr",
    "pr_title": "Update code examples",
    "commit_message": "Sync examples",
    "merge_without_review": false
  }
]
```

## Customizing Your Config

### 1. Set Your Repositories

```yaml
source_repo: "your-org/your-source-repo"  # Where examples come from

targets:
  - repo: "your-org/your-target-repo"     # Where to copy them
```

### 2. Define Patterns

**Match all files in a directory:**
```yaml
source_pattern:
  type: "prefix"
  pattern: "examples/"
```

**Match specific file types:**
```yaml
source_pattern:
  type: "regex"
  pattern: "^examples/.*\\.go$"  # Only .go files
```

**Extract variables from paths:**
```yaml
source_pattern:
  type: "regex"
  pattern: "^examples/(?P<lang>[^/]+)/(?P<category>[^/]+)/(?P<file>.+)$"
  # Extracts: lang, category, file
```

### 3. Transform Paths

```yaml
# Keep same structure
path_transform: "${path}"

# Add prefix
path_transform: "docs/${path}"

# Reorganize with variables
path_transform: "docs/${lang}/${category}/${file}"
```

### 4. Set Commit Strategy

**Direct commit:**
```yaml
commit_strategy:
  type: "direct"
  commit_message: "Update examples"
```

**Pull request:**
```yaml
commit_strategy:
  type: "pull_request"
  commit_message: "Update examples"
  pr_title: "Update ${category} examples"
  pr_body: "Automated update"
  auto_merge: false
```

## Validating Your Config

Before using, validate your configuration:

```bash
# Validate syntax and structure
./config-validator validate -config config.yaml -v

# Test pattern matching
./config-validator test-pattern \
  -type regex \
  -pattern "^examples/(?P<lang>[^/]+)/.*$" \
  -file "examples/go/main.go"

# Test path transformation
./config-validator test-transform \
  -template "docs/${lang}/${file}" \
  -file "examples/go/main.go" \
  -pattern "^examples/(?P<lang>[^/]+)/(?P<file>.+)$"
```

## Testing Your Config

### 1. Test Locally

```bash
# Terminal 1: Start app with local config
make run-local-quick

# Terminal 2: Send test webhook
./test-webhook -payload test-payloads/example-pr-merged.json
```

### 2. Check Logs

Look for:
```
[INFO] loaded config from local file | {"file":"config.yaml"}
[INFO] Loaded YAML configuration with 3 copy rules
[INFO] Pattern matched: examples/go/database/connect.go
[INFO]   → Transformed to: docs/code-examples/go/database/connect.go
```

### 3. Verify Metrics

```bash
curl http://localhost:8080/metrics | jq
```

## Environment Variables

```bash
# Specify config file name (default: config.json)
CONFIG_FILE=config.yaml

# For local testing
COPIER_DISABLE_CLOUD_LOGGING=true
DRY_RUN=true
```

## Troubleshooting

### Error: "config file is empty"
**Solution:** Make sure config.yaml has content

### Error: "config validation failed"
**Solution:** Run `./config-validator validate -config config.yaml -v`

### Error: "pattern doesn't match"
**Solution:** Test pattern with `./config-validator test-pattern`

### Files not being copied
**Solution:** Check that:
1. Pattern matches the file paths
2. Target repo is correct
3. Commit strategy is set

## Quick Start Checklist

- [ ] Config file created (`config.yaml` or `config.json`)
- [ ] Repositories set correctly (source and target)
- [ ] Patterns match your file structure
- [ ] Path transformations tested
- [ ] Config validated with `config-validator`
- [ ] Tested locally with dry-run mode
- [ ] Ready to deploy!

## Next Steps

1. **Test locally** with the provided config
2. **Customize** patterns and transformations for your needs
3. **Validate** with config-validator
4. **Test** with real PR data
5. **Deploy** to production
6. **Add config to source repo** for production use

See [LOCAL-TESTING.md](LOCAL-TESTING.md) for complete testing guide.

