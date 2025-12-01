# Code Copier Workflows

This directory contains workflow configurations for automatically copying code examples to destination repositories.

## For Writers: Quick Overview

**What is this?** An automated system that copies your code examples to other repositories when you merge a PR.

**What do I need to know?**
1. When you **merge a PR** in this repo, the copier automatically runs
2. Files are **matched against patterns** in `.copier/workflows.yaml`
3. Matched files are **copied to destination repositories**
4. A **PR is created** in each destination repository (usually)
5. Someone needs to **review and merge** the destination PRs (unless auto-merge is enabled)

**What happens to deleted files?**
- They are **NOT automatically deleted** from destination repositories
- They are tracked in `deprecated_examples.json` for manual cleanup

**Do I need to do anything?**
- Usually no! Just merge your PR as normal
- Check destination repositories to ensure PRs are created
- If something goes wrong, contact the DevEx team

**Where are the logs?**
```bash
gcloud app logs read --limit=100 | grep "your-repo-name"
```

## Quick Start

### File Location

Place your workflow configuration at: `.copier/workflows.yaml`

### Basic Workflow Structure

```yaml
workflows:
  - name: "my-workflow"
    destination:
      repo: "mongodb/destination-repo"
      branch: "main"
    transformations:
      - move:
          from: "source/path"
          to: "destination/path"
```

### Important: Path Resolution

**All file paths in transformations are relative to the repository root**, not to the config file location.

Even though your config is at `.copier/workflows.yaml`, patterns and paths are matched against the full repository path:

```yaml
# ✅ Correct - paths from repository root
transformations:
  - move:
      from: "examples/go"
      to: "code"

# ❌ Wrong - don't use relative paths like ../
transformations:
  - move:
      from: "../examples/go"  # Don't do this!
      to: "code"
```

## Adding a New Workflow

1. **Edit `.copier/workflows.yaml`** in your repository

2. **Add a new workflow entry:**

```yaml
workflows:
  - name: "my-new-workflow"
    destination:
      repo: "mongodb/my-destination-repo"
      branch: "main"
    transformations:
      - move:
          from: "examples/my-code"
          to: "code"
```

3. **Commit and push** - the workflow is now active!

## Modifying an Existing Workflow

Simply edit the workflow in `.copier/workflows.yaml` and commit your changes. The updated configuration will be used for the next PR merge.

## Common Transformation Types

### Move Directory

Copy all files from one directory to another:

```yaml
transformations:
  - move:
      from: "examples/go"
      to: "code/go"
```

### Copy Single File

Copy a specific file:

```yaml
transformations:
  - copy:
      from: "README.md"
      to: "docs/README.md"
```

### Match with Wildcards

Use glob patterns for flexible matching:

```yaml
transformations:
  - glob:
      pattern: "examples/**/*.go"
      transform: "code/${relative_path}"
```

## Commit Strategies

Choose how files are committed to destination repositories:

### Pull Request (Recommended)

Creates a PR in the destination repository for review:

```yaml
commit_strategy:
  type: "pull_request"
  pr_title: "Update examples from ${source_repo}"
  pr_body: |
    Automated update from source repository.

    Source PR: #${pr_number}
    Commit: ${commit_sha}
  auto_merge: false  # Requires manual review and merge
```

**When to use:**
- When destination repo requires code review
- When you want to run CI/CD tests before merging
- When changes need approval (most common)

### Pull Request with Auto-Merge

Creates a PR and automatically merges it if there are no conflicts:

```yaml
commit_strategy:
  type: "pull_request"
  auto_merge: true  # Automatically merges if no conflicts
```

**When to use:**
- When destination repo has automated tests that must pass
- When you trust the source content completely
- When you want fast propagation with safety checks

### Direct Commit

Commits directly to the destination branch (no PR):

```yaml
commit_strategy:
  type: "direct"
  commit_message: "Update examples from ${source_repo}"
```

**When to use:**
- When destination repo doesn't require review
- When you need immediate updates
- When you have full confidence in the source content

**⚠️ Warning:** Direct commits bypass code review and CI checks!

## Available Variables

Use these variables in PR titles, bodies, and commit messages:

- `${source_repo}` - Source repository name
- `${source_branch}` - Source branch name
- `${pr_number}` - Source PR number
- `${commit_sha}` - Source commit SHA
- `${file_count}` - Number of files changed

Use these in path transformations:

- `${relative_path}` - Path relative to the matched pattern
- `${path}` - Full source file path
- `${filename}` - Just the filename
- `${dir}` - Directory path
- `${ext}` - File extension

## Excluding Files

Prevent certain files from being copied:

```yaml
workflows:
  - name: "my-workflow"
    destination:
      repo: "mongodb/destination-repo"
      branch: "main"
    transformations:
      - move: { from: "src", to: "dest" }
    exclude:
      - "**/.env"           # Environment files
      - "**/node_modules/**" # Dependencies
      - "**/*.test.js"      # Test files
      - "**/.DS_Store"      # Mac system files
```

**Common exclusions:**
- Environment files: `**/.env`, `**/.env.*`
- Dependencies: `**/node_modules/**`, `**/vendor/**`
- Build artifacts: `**/dist/**`, `**/build/**`
- Test files: `**/*.test.*`, `**/tests/**`
- System files: `**/.DS_Store`, `**/Thumbs.db`

## Setting Defaults

Apply settings to all workflows in this file:

```yaml
defaults:
  commit_strategy:
    type: "pull_request"
    auto_merge: false
  exclude:
    - "**/.env"
    - "**/node_modules/**"

workflows:
  - name: "workflow-1"
    # inherits defaults
    destination:
      repo: "mongodb/dest-1"
    transformations:
      - move: { from: "src", to: "dest" }
  
  - name: "workflow-2"
    # inherits defaults
    destination:
      repo: "mongodb/dest-2"
    transformations:
      - move: { from: "examples", to: "code" }
```

## Testing Your Configuration

Before committing, you can validate your configuration:

```bash
# Validate syntax
./config-validator validate -config .copier/workflows.yaml

# Test a pattern match
./config-validator test-pattern \
  -type glob \
  -pattern "examples/**/*.go" \
  -file "examples/database/connect.go"
```

## How It Works

### When You Merge a PR

1. **You merge a PR** in this repository
2. **GitHub sends a webhook** to the copier application
3. **Copier loads your workflows** from `.copier/workflows.yaml`
4. **Files are matched** against transformation patterns
5. **Files are copied** to destination repositories
6. **PRs are created** in destination repositories (or committed directly)

### What Triggers the Copier?

**✅ Triggers:**
- Merging a PR (action: `closed` + `merged: true`)

**❌ Does NOT trigger:**
- Opening a PR
- Updating a PR
- Closing a PR without merging
- Direct commits to main branch (no PR)
- Draft PRs

### What Happens to Different File Types?

**Added or Modified Files:**
- Matched against transformation patterns
- Copied to destination repository
- Included in destination PR

**Deleted Files:**
- Matched against transformation patterns
- Added to deprecation tracking file (if enabled)
- **NOT automatically deleted** from destination
- Manual cleanup required (see Deprecation Tracking below)

## Understanding Destination PRs

### What Gets Created in the Destination Repository?

When you merge a PR in this repository, the copier creates a PR in each destination repository:

**PR Structure:**
- **Branch name**: `copier/YYYYMMDD-HHMMSS` (e.g., `copier/20240115-143022`)
- **Base branch**: The branch specified in your workflow (usually `main`)
- **Title**: From your `pr_title` configuration
- **Body**: From your `pr_body` configuration
- **Files**: All matched files from your source PR

**Example:**
```
Source PR: mongodb/docs-code-examples #123
  ↓
Destination PR: mongodb/go-examples-repo #45
  Branch: copier/20240115-143022
  Files: 5 files changed
  Status: Open (awaiting review)
```

### What Happens After the Destination PR is Created?

**If `auto_merge: false` (default):**
1. PR is created and left open
2. Destination repo maintainers review the PR
3. CI/CD tests run (if configured)
4. Someone manually merges the PR

**If `auto_merge: true`:**
1. PR is created
2. Copier waits for GitHub to compute mergeability
3. If no conflicts: PR is automatically merged
4. If conflicts: PR is left open for manual resolution
5. Temporary branch is deleted after merge

**If `type: "direct"`:**
1. No PR is created
2. Files are committed directly to the target branch
3. No review process

## Deprecation Tracking

When you delete files from this repository, the copier can track them for cleanup in destination repositories.

### How It Works

1. **You delete a file** and merge the PR
2. **Copier detects the deletion** and matches it against patterns
3. **File is added** to `deprecated_examples.json` in this repository
4. **You manually clean up** the file from destination repositories

### Enable Deprecation Tracking

```yaml
workflows:
  - name: "my-workflow"
    destination:
      repo: "mongodb/destination-repo"
      branch: "main"
    transformations:
      - move: { from: "examples", to: "code" }
    deprecation_check:
      enabled: true
      file: "deprecated_examples.json"  # Optional: custom filename
```

### Deprecation File Format

The deprecation file is stored in **this repository** (source):

```json
[
  {
    "filename": "code/go/old-example.go",
    "repo": "mongodb/destination-repo",
    "branch": "main",
    "deleted_on": "2024-01-15T10:30:00Z"
  }
]
```

### Cleanup Process

1. **Review** `deprecated_examples.json` periodically
2. **Create PRs** in destination repositories to remove deprecated files
3. **Remove entries** from `deprecated_examples.json` after cleanup

**Note:** Files are **NOT automatically deleted** from destination repositories. Deprecation tracking only records what needs to be cleaned up.

## Need Help?

- **Full Documentation**: [Code Example Tooling Repository](https://github.com/mongodb/code-example-tooling)
- **Configuration Examples**: See `examples-copier/configs/copier-config-examples/`
- **Pattern Matching Guide**: See `examples-copier/docs/PATTERN-MATCHING-GUIDE.md`
- **Main Config Architecture**: See `examples-copier/configs/copier-config-examples/MAIN-CONFIG-README.md`
- **Deprecation Tracking**: See `examples-copier/docs/DEPRECATION-TRACKING-EXPLAINED.md`

## Example: Complete Workflow

```yaml
# .copier/workflows.yaml

defaults:
  commit_strategy:
    type: "pull_request"
    auto_merge: false
  exclude:
    - "**/.env"
    - "**/node_modules/**"

workflows:
  - name: "go-examples"
    destination:
      repo: "mongodb/go-examples-repo"
      branch: "main"
    transformations:
      - move:
          from: "examples/go"
          to: "code"
    commit_strategy:
      pr_title: "Update Go examples from ${source_repo}"
      pr_body: |
        Automated update of Go code examples.
        
        **Source**: ${source_repo} (PR #${pr_number})
        **Commit**: ${commit_sha}
        **Files**: ${file_count} changed
    deprecation_check:
      enabled: true
      file: "deprecated_examples.json"
```

## Troubleshooting

### My PR merged but files weren't copied

**Check:**
1. Was it a merged PR? (not just closed)
2. Do the changed files match your transformation patterns?
3. Check the copier logs (see below)
4. Verify `.copier/workflows.yaml` is valid YAML

### How do I view the logs?

```bash
# View recent logs
gcloud app logs read --limit=100

# Search for your PR
gcloud app logs read --limit=200 | grep "PR #123"

# Search for your repo
gcloud app logs read --limit=200 | grep "your-repo-name"
```

### How do I test my configuration?

```bash
# Validate YAML syntax
./config-validator validate -config .copier/workflows.yaml

# Test a pattern match
./config-validator test-pattern \
  -type glob \
  -pattern "examples/**/*.go" \
  -file "examples/database/connect.go"
```

### Files are being copied to the wrong location

**Check:**
- Are your paths relative to repository root? (not relative to config file)
- Is your `transform` pattern correct?
- Test with `config-validator` tool

### I want to copy to multiple destinations

Create multiple workflows, one for each destination:

```yaml
workflows:
  - name: "copy-to-docs"
    destination:
      repo: "mongodb/docs"
    transformations:
      - move: { from: "examples", to: "code" }

  - name: "copy-to-website"
    destination:
      repo: "mongodb/website"
    transformations:
      - move: { from: "examples", to: "static/examples" }
```

## Questions?

Contact the Developer Experience team or open an issue in the [code-example-tooling repository](https://github.com/mongodb/code-example-tooling/issues).

