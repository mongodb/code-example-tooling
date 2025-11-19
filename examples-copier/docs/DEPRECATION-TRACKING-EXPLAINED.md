# Deprecation Tracking - How It Works

Deprecation tracking automatically detects when files are **deleted** from the source repository and records them in a deprecation file for later cleanup in target repositories.

## How It Works

### 1. File Deletion Detection

When a PR is merged in the source repository:

1. **Webhook triggers** - Application receives PR merged event
2. **Changed files retrieved** - Gets list of all files in the PR
3. **Status checked** - Each file has a status: `ADDED`, `MODIFIED`, or `DELETED`
4. **Deleted files identified** - Files with `DELETED` status are flagged

### 2. Pattern Matching

For each deleted file:

1. **Pattern matching runs** - File path is matched against copy rules
2. **Target path calculated** - Determines where file exists in target repo
3. **Deprecation queue updated** - File added to deprecation queue

### 3. Deprecation File Update

After processing all files:

1. **Check deprecation queue** - If queue is empty, no commit is made
2. **Fetch existing file** - Retrieves current `deprecated_examples.json`
3. **Merge entries** - Adds new deprecated files to existing list
4. **Commit to source repo** - Updates deprecation file with new entries

## Code Flow

```
PR Merged Event
    ↓
Get Changed Files
    ↓
For each file:
    ├─ Status = DELETED?
    │   ├─ Yes → Match against patterns
    │   │         ↓
    │   │      Add to deprecation queue
    │   │
    │   └─ No → Process for copying
    ↓
Check deprecation queue
    ├─ Empty? → Skip update (NO COMMIT)
    └─ Has files? → Update deprecation file
```

## Blank Commit Protection

### ✅ YES - Protected Against Blank Commits

The implementation has built-in protection against blank commits:

````go
func UpdateDeprecationFile() {
    // ✅ Early return if there are no files to deprecate - prevents blank commits
    if len(FilesToDeprecate) == 0 {
        LogInfo("No deprecated files to record; skipping deprecation file update")
        return  // ← NO COMMIT MADE
    }

    // ... rest of update logic ...
}
````

**Protection:**
- ✅ Checks if deprecation queue is empty
- ✅ Returns early if nothing to deprecate
- ✅ **No commit is made** if queue is empty
- ✅ Logs skip message for visibility


## Deprecation File Format

The deprecation file is a JSON array stored in the **source repository**:

```json
[
  {
    "filename": "docs/examples/old-example.go",
    "repo": "mongodb/docs",
    "branch": "main",
    "deleted_on": "2024-01-15T10:30:00Z"
  },
  {
    "filename": "docs/examples/deprecated-file.py",
    "repo": "mongodb/tutorials",
    "branch": "main",
    "deleted_on": "2024-01-20T14:22:00Z"
  }
]
```

**Fields:**
- `filename` - Path in target repository where file exists
- `repo` - Target repository (e.g., `mongodb/docs`)
- `branch` - Target branch (e.g., `main`)
- `deleted_on` - Timestamp when file was deleted from source

## Configuration

Enable deprecation tracking in your workflow config:

```yaml
workflows:
  - name: "Copy Go examples"
    source:
      repo: "mongodb/source-repo"
      branch: "main"
    destination:
      repo: "mongodb/docs"
      branch: "main"
    transformations:
      - move:
          from: "examples/go"
          to: "code"
    deprecation_check:
      enabled: true                      # ← Enable tracking
      file: "deprecated_examples.json"   # ← Optional: custom filename
```

**Options:**
- `enabled` - Set to `true` to enable tracking
- `file` - (Optional) Custom deprecation file name (default: `deprecated_examples.json`)

## Use Cases

### 1. Automatic Cleanup Tracking

**Scenario:** You delete an old example from source repo

**What Happens:**
1. File deleted in source PR
2. PR merged → webhook triggered
3. Deleted file matched against patterns
4. Target path calculated
5. Entry added to deprecation file
6. **Manual cleanup** - You can now clean up target repos

### 2. Audit Trail

**Scenario:** Need to know what files were removed and when

**What Happens:**
- Deprecation file provides complete history
- Timestamp of deletion
- Which target repos are affected
- Which branch contains the file

### 3. Batch Cleanup

**Scenario:** Clean up multiple target repositories

**What Happens:**
1. Review deprecation file
2. See all files that need cleanup
3. Create cleanup PRs for each target repo
4. Remove entries from deprecation file after cleanup

## Workflow Example

### Step 1: Delete File from Source

```bash
# In source repo (mongodb/docs-code-examples)
git rm examples/go/old-example.go
git commit -m "Remove outdated example"
git push origin feature-branch

# Create and merge PR
```

### Step 2: Webhook Processes Deletion

```
[INFO] processing merged PR | {"pr_number": 123}
[INFO] retrieved changed files | {"count": 1}
[INFO] file matched pattern | {"file": "examples/go/old-example.go", "status": "DELETED"}
[INFO] file marked for deprecation | {"target_path": "docs/code/go/old-example.go"}
[INFO] successfully updated deprecated_examples.json with 1 entries
```

### Step 3: Review Deprecation File

```json
[
  {
    "filename": "docs/code/go/old-example.go",
    "repo": "mongodb/docs",
    "branch": "main",
    "deleted_on": "2024-01-15T10:30:00Z"
  }
]
```

### Step 4: Manual Cleanup

```bash
# In target repo (mongodb/docs)
git rm docs/code/go/old-example.go
git commit -m "Remove deprecated example"
git push origin cleanup-branch

# Create PR for review
```

### Step 5: Update Deprecation File

After cleanup, remove the entry from `deprecated_examples.json`.

## Dry-Run Mode

In dry-run mode, deprecation tracking is **simulated**:

```bash
DRY_RUN=true ./examples-copier
```

**Output:**
```
[DRY_RUN] Would update deprecation file with the following entries
[DRY_RUN] file: docs/code/go/old-example.go repo: mongodb/docs branch: main
```

**No commits are made** - only logs what would happen.

## Best Practices

### 1. Enable for Production Targets

```yaml
# Enable for important repositories
targets:
  - repo: "mongodb/docs"
    deprecation_check:
      enabled: true  # ← Track deletions
```

### 2. Regular Cleanup

- Review deprecation file weekly/monthly
- Create cleanup PRs for target repos
- Remove entries after cleanup

### 3. Monitor Deprecation File Size

- Large deprecation file = cleanup needed
- Set up alerts for file size
- Automate cleanup where possible

### 4. Use Different Files per Target

```yaml
targets:
  - repo: "mongodb/docs"
    deprecation_check:
      enabled: true
      file: "deprecated_docs.json"
  
  - repo: "mongodb/tutorials"
    deprecation_check:
      enabled: true
      file: "deprecated_tutorials.json"
```

## Limitations

### 1. Manual Cleanup Required

- Deprecation file only **tracks** deletions
- Does **not** automatically delete from target repos
- Manual cleanup PRs still needed

### 2. Source Repository Only

- Deprecation file is stored in **source repository**
- Not stored in target repositories
- Requires access to source repo to see deprecations

### 3. No Automatic Expiration

- Entries remain until manually removed
- No automatic cleanup after X days
- File can grow large over time

## Future Enhancements

Potential improvements:

1. **Automatic Cleanup PRs** - Create PRs to remove deprecated files
2. **Expiration Dates** - Auto-remove entries after X days
3. **Cleanup Verification** - Check if file still exists in target
4. **Batch Cleanup Tool** - CLI tool to clean up all deprecated files
5. **Notifications** - Alert when deprecation file grows large

## Summary

**How Deprecation Tracking Works:**
1. ✅ Detects deleted files in source PR
2. ✅ Matches against copy rules patterns
3. ✅ Calculates target repository paths
4. ✅ Records in deprecation file (source repo)
5. ✅ **Protected against blank commits**

**Use Cases:**
- Track deleted files for cleanup
- Audit trail of removals
- Coordinate cleanup across multiple repos

**Limitations:**
- Manual cleanup still required
- No automatic deletion
- File can grow over time

---

**See Also:**
- [Configuration Guide](CONFIGURATION-GUIDE.md) - Deprecation configuration
- [Architecture](ARCHITECTURE.md) - System design
- [Troubleshooting](TROUBLESHOOTING.md) - Common issues

