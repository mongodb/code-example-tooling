# Code Copier Workflows

This directory contains workflow configurations for automatically copying code examples to destination repositories.

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

## Customizing PR Details

Add custom PR titles and descriptions:

```yaml
workflows:
  - name: "my-workflow"
    destination:
      repo: "mongodb/destination-repo"
      branch: "main"
    transformations:
      - move: { from: "src", to: "dest" }
    commit_strategy:
      type: "pull_request"
      pr_title: "Update examples from ${source_repo}"
      pr_body: |
        Automated update from source repository.
        
        Source PR: #${pr_number}
        Commit: ${commit_sha}
      auto_merge: false
```

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
      - "**/.env"
      - "**/node_modules/**"
      - "**/*.test.js"
```

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

1. **You merge a PR** in this repository
2. **Copier detects the merge** via webhook
3. **Workflows are matched** based on changed files
4. **Files are copied** to destination repositories
5. **PRs are created** in destination repositories

## Need Help?

- **Full Documentation**: [Code Example Tooling Repository](https://github.com/mongodb/code-example-tooling)
- **Configuration Examples**: See `examples-copier/configs/copier-config-examples/`
- **Pattern Matching Guide**: See `examples-copier/docs/PATTERN-MATCHING-GUIDE.md`
- **Main Config Architecture**: See `examples-copier/configs/copier-config-examples/MAIN-CONFIG-README.md`

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

## Questions?

Contact the Developer Experience team or open an issue in the [code-example-tooling repository](https://github.com/mongodb/code-example-tooling/issues).

