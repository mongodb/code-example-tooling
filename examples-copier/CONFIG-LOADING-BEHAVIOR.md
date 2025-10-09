# Config Loading Behavior - When Is Config Read?

## Quick Answer

**The config file is read from the SOURCE BRANCH (typically `main`), NOT from the merged PR.**

This means:
- ❌ **Config changes in a PR are NOT used** for that same PR
- ✅ **Config changes take effect** for the NEXT PR after they're merged
- ⚠️ **You cannot update config and files in the same PR** and have the new config apply

## How It Works

### Webhook Flow

When a PR is merged:

```
1. PR Merged → Webhook triggered
2. Authenticate with GitHub
3. Read config file from SOURCE BRANCH (main) ← IMPORTANT
4. Get changed files from the merged PR
5. Apply copy rules from config
6. Copy files to target repos
```

### Code Implementation

<augment_code_snippet path="examples-copier/services/github_read.go" mode="EXCERPT">
````go
func retrieveJsonFile(filePath string) string {
    client := GetRestClient()
    owner := os.Getenv(configs.RepoOwner)
    repo := os.Getenv(configs.RepoName)
    ctx := context.Background()
    fileContent, _, _, err :=
        client.Repositories.GetContents(ctx, owner, repo,
            filePath, &github.RepositoryContentGetOptions{
                Ref: os.Getenv(configs.SrcBranch),  // ← Reads from SRC_BRANCH (default: "main")
            })
    // ...
}
````
</augment_code_snippet>

**Key Point:** The `Ref` parameter is set to `os.Getenv(configs.SrcBranch)`, which defaults to `"main"`.

### Environment Configuration

<augment_code_snippet path="examples-copier/configs/environment.go" mode="EXCERPT">
````go
func NewConfig() *Config {
    return &Config{
        // ...
        SrcBranch: "main",  // ← Default branch to read config from
        // ...
    }
}
````
</augment_code_snippet>

**Default:** Config is always read from the `main` branch (or whatever `SRC_BRANCH` is set to).

## Scenarios

### Scenario 1: Update Config Only

**PR Contents:**
- Modified `copier-config.yaml` (new copy rule added)

**What Happens:**
1. PR merged
2. Webhook reads config from `main` (BEFORE this PR merged)
3. Old config is used (new rule NOT applied)
4. No files copied (because no other files changed)
5. Config changes are now in `main` for NEXT PR

**Result:** ✅ Config updated for future PRs

### Scenario 2: Update Config + Files in Same PR

**PR Contents:**
- Modified `copier-config.yaml` (new rule: copy `examples/new/*.go`)
- Added `examples/new/example.go`

**What Happens:**
1. PR merged
2. Webhook reads config from `main` (BEFORE this PR merged)
3. Old config is used (new rule NOT in effect yet)
4. `examples/new/example.go` does NOT match any rules
5. File is NOT copied ❌
6. Config changes are now in `main` for NEXT PR

**Result:** ⚠️ File NOT copied - need another PR to trigger copy

### Scenario 3: Update Files After Config Merged

**PR 1 Contents:**
- Modified `copier-config.yaml` (new rule: copy `examples/new/*.go`)

**PR 1 Result:**
- Config merged to `main`
- No files copied (no other files changed)

**PR 2 Contents:**
- Added `examples/new/example.go`

**PR 2 What Happens:**
1. PR merged
2. Webhook reads config from `main` (includes new rule from PR 1)
3. New config is used ✅
4. `examples/new/example.go` matches new rule
5. File is copied to target repo ✅

**Result:** ✅ File copied successfully

### Scenario 4: Update Existing Rule

**Current Config:**
```yaml
copy_rules:
  - name: "Go examples"
    source_pattern:
      type: "prefix"
      pattern: "examples/go"
    targets:
      - repo: "mongodb/docs"
        branch: "main"
        path_transform: "code/${relative_path}"
```

**PR Contents:**
- Modified `copier-config.yaml` (changed `path_transform` to `"docs/${relative_path}"`)
- Modified `examples/go/example.go`

**What Happens:**
1. PR merged
2. Webhook reads config from `main` (BEFORE this PR merged)
3. Old config is used (old path_transform: `"code/${relative_path}"`)
4. `examples/go/example.go` is copied to `code/go/example.go` (OLD path)
5. Config changes are now in `main` for NEXT PR

**Result:** ⚠️ File copied to OLD path - need another PR to copy to NEW path

## Why This Design?

### Stability

**Reason:** Ensures config is stable and tested before being used.

**Benefit:**
- Config changes are reviewed and merged first
- Next PR uses the reviewed config
- Reduces risk of broken config affecting file copying

### Predictability

**Reason:** Config state is known at webhook trigger time.

**Benefit:**
- No race conditions between config and file changes
- Clear separation: config changes vs file changes
- Easier to debug and understand behavior

### Simplicity

**Reason:** Always reads from a known branch (`main`).

**Benefit:**
- No need to check if config changed in PR
- No need to merge config changes before reading
- Simpler implementation

## Workarounds

### Option 1: Two-PR Workflow (Recommended)

**Step 1:** Update config
```bash
# PR 1: Update config only
git checkout -b update-config
# Edit copier-config.yaml
git add copier-config.yaml
git commit -m "Add new copy rule for X"
git push origin update-config
# Create PR, get approval, merge
```

**Step 2:** Add/modify files
```bash
# PR 2: Add files that use new config
git checkout -b add-files
# Add/modify files
git add examples/new/
git commit -m "Add new examples"
git push origin add-files
# Create PR, get approval, merge
# Files will be copied using new config ✅
```

### Option 2: Manual Trigger (If Supported)

If the tool supports manual triggering:
```bash
# After PR with config + files is merged
# Manually trigger webhook or re-run copier
# This will use the newly merged config
```

### Option 3: Empty Commit Trigger

**After config is merged:**
```bash
# Create empty commit to trigger webhook
git checkout -b trigger-copy
git commit --allow-empty -m "Trigger copy with new config"
git push origin trigger-copy
# Create PR, merge
# This will trigger webhook with new config
```

### Option 4: Modify File Again

**After config is merged:**
```bash
# Make a small change to the file
git checkout -b fix-copy
# Edit the file (add comment, fix typo, etc.)
git add examples/new/example.go
git commit -m "Trigger copy with updated config"
git push origin fix-copy
# Create PR, merge
# File will be copied with new config ✅
```

## Best Practices

### 1. Update Config First

**Always update config in a separate PR before adding files that use it.**

```
PR 1: Update copier-config.yaml
  ↓ (merge)
PR 2: Add files that match new rules
  ↓ (merge, files copied ✅)
```

### 2. Test Config Changes

**Use config-validator to test config before merging:**

```bash
cd examples-copier
./tools/config-validator/config-validator -config copier-config.yaml
```

### 3. Document Config Changes

**In PR description, note that files will be copied in NEXT PR:**

```markdown
## Changes
- Added new copy rule for `examples/python/*.py`

## Note
Files matching this rule will be copied starting with the NEXT PR after this is merged.
```

### 4. Plan Multi-Step Changes

**For complex changes, plan the sequence:**

```
Step 1: Update config (PR #123)
Step 2: Add new files (PR #124)
Step 3: Update existing files (PR #125)
```

### 5. Use Dry-Run for Testing

**Test config changes with dry-run mode:**

```bash
DRY_RUN=true ./examples-copier
# Check logs to see what would be copied
```

## Common Mistakes

### ❌ Mistake 1: Config + Files in Same PR

**Problem:**
```
PR: Update config + add files
Result: Files NOT copied (old config used)
```

**Solution:**
```
PR 1: Update config
PR 2: Add files (after PR 1 merged)
```

### ❌ Mistake 2: Expecting Immediate Effect

**Problem:**
```
Merge PR with config changes
Expect next file change to use new config immediately
```

**Reality:**
```
Config changes take effect for NEXT PR
Current PR uses OLD config
```

### ❌ Mistake 3: Not Testing Config

**Problem:**
```
Merge config with typo
Next PR fails to copy files
```

**Solution:**
```
Use config-validator before merging
Test with dry-run mode
```

## Future Enhancements

Potential improvements to config loading:

### 1. Read Config from Merged PR

**Idea:** Read config from the merged commit instead of `main`.

**Benefit:**
- Config changes apply immediately
- Single PR can update config + files

**Challenge:**
- More complex implementation
- Potential race conditions
- Harder to debug

### 2. Config Validation on PR

**Idea:** Validate config in PR before merge.

**Benefit:**
- Catch config errors before merge
- Prevent broken config from reaching `main`

**Implementation:**
- GitHub Action to validate config
- Block merge if validation fails

### 3. Config Change Detection

**Idea:** Detect if config changed in PR and warn user.

**Benefit:**
- User knows config won't apply to current PR
- Suggests two-PR workflow

**Implementation:**
- Check if `copier-config.yaml` in changed files
- Add comment to PR with warning

## Summary

**Current Behavior:**
- ✅ Config is read from `main` branch (or `SRC_BRANCH`)
- ❌ Config changes in PR do NOT apply to that PR
- ✅ Config changes apply to NEXT PR after merge

**Recommended Workflow:**
1. Update config in separate PR
2. Merge config PR
3. Add/modify files in subsequent PR
4. Files copied with new config ✅

**Key Takeaway:**
> Config changes require a two-PR workflow: one to update config, another to use the new config.

---

**See Also:**
- [Configuration Guide](docs/CONFIGURATION-GUIDE.md) - Complete config reference
- [Pattern Matching Guide](docs/PATTERN-MATCHING-GUIDE.md) - Pattern syntax
- [Troubleshooting](docs/TROUBLESHOOTING.md) - Common issues

