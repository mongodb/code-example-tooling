# Configuration Guide

Complete guide to configuring the examples-copier application.

## Table of Contents

- [Overview](#overview)
- [Configuration File Structure](#configuration-file-structure)
- [Top-Level Fields](#top-level-fields)
- [Copy Rules](#copy-rules)
- [Source Patterns](#source-patterns)
- [Target Configuration](#target-configuration)
- [Commit Strategies](#commit-strategies)
- [Deprecation Tracking](#deprecation-tracking)
- [Built-in Variables](#built-in-variables)
- [Complete Examples](#complete-examples)
- [Validation](#validation)
- [Best Practices](#best-practices)
- [Pattern Matching Cheatsheet](#pattern-matching-cheat-sheet)

## Overview

The examples-copier uses a YAML configuration file (default: `copier-config.yaml`) to define how files are copied from a source repository to one or more target repositories.

**Key Features:**
- Pattern matching (prefix, glob, regex)
- Path transformation with variables
- Multiple targets per rule
- Flexible commit strategies
- Deprecation tracking
- Template-based messages

## Configuration File Structure

```yaml
# Top-level configuration
source_repo: "owner/source-repository"
source_branch: "main"
batch_by_repo: false  # Optional: batch all changes into one PR per target repo

# Copy rules define what to copy and where
copy_rules:
  - name: "rule-name"
    source_pattern:
      type: "prefix|glob|regex"
      pattern: "pattern-string"
    targets:
      - repo: "owner/target-repository"
        branch: "main"
        path_transform: "target/path/${variable}"
        commit_strategy:
          type: "direct|pull_request|batch"
          # ... strategy options
        deprecation_check:
          enabled: true
          file: "deprecated_examples.json"
```

## Top-Level Fields

### source_repo

**Type:** String (required)  
**Format:** `owner/repository`

The source repository where files are copied from.

```yaml
source_repo: "mongodb/docs-code-examples"
```

### source_branch

**Type:** String (optional)
**Default:** `"main"`

The branch to copy files from.

```yaml
source_branch: "main"
```

### batch_by_repo

**Type:** Boolean (optional)
**Default:** `false`

When `true`, all changes from a single source PR are batched into **one pull request per target repository**, regardless of how many copy rules match files.

When `false` (default), each copy rule creates a **separate pull request** in the target repository.

**Example - Separate PRs per rule (default):**
```yaml
batch_by_repo: false  # or omit this field

copy_rules:
  - name: "copy-client"
    # ... matches 5 files
  - name: "copy-server"
    # ... matches 3 files
  - name: "copy-readme"
    # ... matches 1 file

# Result: 3 separate PRs in the target repo
```

**Example - Single batched PR:**
```yaml
batch_by_repo: true

copy_rules:
  - name: "copy-client"
    # ... matches 5 files
  - name: "copy-server"
    # ... matches 3 files
  - name: "copy-readme"
    # ... matches 1 file

# Result: 1 PR containing all 9 files in the target repo
```

**Use Cases:**
- ✅ **Use `batch_by_repo: true`** when you want all related changes in a single PR for easier review
- ✅ **Use `batch_by_repo: false`** when different rules need separate review processes or different reviewers

**Note:** When batching is enabled, use `batch_pr_config` (see below) to customize PR metadata, or a generic title/body will be generated.

### batch_pr_config

**Type:** Object (optional)
**Used when:** `batch_by_repo: true`

Defines PR metadata (title, body, commit message) for batched pull requests. This allows you to customize the PR with accurate file counts and custom messaging.

**Fields:**
- `pr_title` - (optional) PR title template
- `pr_body` - (optional) PR body template
- `commit_message` - (optional) Commit message template
- `use_pr_template` - (optional) Fetch and merge PR template from target repo (default: false)

**Available template variables:**
- `${source_repo}` - Source repository (e.g., "owner/repo")
- `${target_repo}` - Target repository
- `${source_branch}` - Source branch name
- `${target_branch}` - Target branch name
- `${file_count}` - **Accurate** total number of files in the batched PR
- `${pr_number}` - Source PR number
- `${commit_sha}` - Source commit SHA

**Example:**
```yaml
source_repo: "mongodb/code-examples"
source_branch: "main"
batch_by_repo: true

batch_pr_config:
  pr_title: "Update code examples from ${source_repo}"
  pr_body: |
    🤖 Automated update of code examples

    **Source Information:**
    - Repository: ${source_repo}
    - PR: #${pr_number}
    - Commit: ${commit_sha}

    **Changes:**
    - Total files: ${file_count}
    - Target branch: ${target_branch}
  commit_message: "Update examples from ${source_repo} PR #${pr_number}"
  use_pr_template: true  # Fetch PR template from target repos

copy_rules:
  - name: "copy-client"
    # ... rule config
  - name: "copy-server"
    # ... rule config
```

**Default behavior (if `batch_pr_config` is not specified):**
```yaml
# Default PR title:
"Update files from owner/repo PR #123"

# Default PR body:
"Automated update from owner/repo

Source PR: #123
Commit: abc1234
Files: 42"
```

### copy_rules

**Type:** Array (required)
**Minimum:** 1 rule

List of copy rules that define what files to copy and where.

```yaml
copy_rules:
  - name: "first-rule"
    # ... rule configuration
  - name: "second-rule"
    # ... rule configuration
```

## Copy Rules

Each copy rule defines a pattern to match files and one or more targets to copy them to.

### name

**Type:** String (required)

Descriptive name for the rule. Used in logs and metrics.

```yaml
name: "Copy Go examples"
```

### source_pattern

**Type:** Object (required)

Defines how to match source files. See [Source Patterns](#source-patterns).

### targets

**Type:** Array (required)  
**Minimum:** 1 target

List of target repositories and configurations. See [Target Configuration](#target-configuration).

## Source Patterns

Source patterns define which files to match in the source repository.

### Pattern Types

#### 1. Prefix Pattern

Matches files that start with a specific prefix.

```yaml
source_pattern:
  type: "prefix"
  pattern: "examples/go"
```

**Matches:**
- `examples/go/main.go` ✓
- `examples/go/database/connect.go` ✓
- `examples/python/main.py` ✗

**Variables Extracted:**
- `matched_prefix` - The matched prefix
- `relative_path` - Path after the prefix

**Example:**
```yaml
# File: examples/go/database/connect.go
# Variables:
#   matched_prefix: "examples/go"
#   relative_path: "database/connect.go"
```

#### 2. Glob Pattern

Matches files using wildcard patterns.

```yaml
source_pattern:
  type: "glob"
  pattern: "examples/*/*.go"
```

**Wildcards:**
- `*` - Matches any characters except `/`
- `**` - Matches any characters including `/`
- `?` - Matches single character

**Examples:**
```yaml
# Match all Go files in any language directory
pattern: "examples/*/*.go"

# Match all files in examples directory (recursive)
pattern: "examples/**/*"

# Match specific file types
pattern: "examples/**/*.{go,py,js}"
```

**Variables Extracted:**
- `matched_pattern` - The pattern that matched

#### 3. Regex Pattern

Matches files using regular expressions with named capture groups.

```yaml
source_pattern:
  type: "regex"
  pattern: "^examples/(?P<lang>[^/]+)/(?P<category>[^/]+)/(?P<file>.+)$"
```

**Named Capture Groups:**
Use `(?P<name>...)` syntax to extract variables.

**Example:**
```yaml
# File: examples/go/database/connect.go
# Pattern: ^examples/(?P<lang>[^/]+)/(?P<category>[^/]+)/(?P<file>.+)$
# Variables:
#   lang: "go"
#   category: "database"
#   file: "connect.go"
```

**Common Patterns:**
```yaml
# Language and file
pattern: "^examples/(?P<lang>[^/]+)/(?P<file>.+)$"

# Language, category, and file
pattern: "^examples/(?P<lang>[^/]+)/(?P<category>[^/]+)/(?P<file>.+)$"

# Version-specific examples
pattern: "^examples/v(?P<version>[0-9]+)/(?P<lang>[^/]+)/(?P<file>.+)$"

# Optional segments
pattern: "^examples/(?P<lang>[^/]+)(/(?P<category>[^/]+))?/(?P<file>[^/]+)$"
```

### Excluding Files with `exclude_patterns`

**Type:** Array of strings (optional)
**Format:** Go-compatible regex patterns

You can exclude specific files from being matched by adding `exclude_patterns` to any source pattern. This is useful for filtering out files like `.gitignore`, `.env`, `node_modules`, build artifacts, etc.

**Important:** Exclude patterns use **Go regex syntax** (no negative lookahead `(?!...)`).

#### Basic Example

```yaml
source_pattern:
  type: "prefix"
  pattern: "examples/"
  exclude_patterns:
    - "\.gitignore$"      # Exclude .gitignore files
    - "\.env$"            # Exclude .env files
    - "node_modules/"     # Exclude node_modules directory
```

#### How It Works

1. **Main pattern matches first** - The file must match the main pattern (`type` and `pattern`)
2. **Then exclusions are checked** - If the file matches any `exclude_patterns`, it's excluded
3. **Result** - File is only copied if it matches the main pattern AND doesn't match any exclusions

#### Examples by Pattern Type

**Prefix Pattern with Exclusions:**
```yaml
- name: "copy-examples-no-config"
  source_pattern:
    type: "prefix"
    pattern: "examples/"
    exclude_patterns:
      - "\.gitignore$"
      - "\.env$"
      - "/node_modules/"
      - "/dist/"
      - "/build/"
  targets:
    - repo: "mongodb/docs"
      branch: "main"
      path_transform: "code-examples/${relative_path}"
```

**Regex Pattern with Exclusions:**
```yaml
- name: "java-server-no-tests"
  source_pattern:
    type: "regex"
    pattern: "^mflix/server/java-spring/(?P<file>.+)$"
    exclude_patterns:
      - "/test/"           # Exclude test directories
      - "Test\.java$"      # Exclude test files
      - "\.gitignore$"     # Exclude .gitignore
  targets:
    - repo: "mongodb/sample-app-java"
      branch: "main"
      path_transform: "server/${file}"
```

**Glob Pattern with Exclusions:**
```yaml
- name: "js-files-no-minified"
  source_pattern:
    type: "glob"
    pattern: "examples/**/*.js"
    exclude_patterns:
      - "\.min\.js$"       # Exclude minified files
      - "\.test\.js$"      # Exclude test files
  targets:
    - repo: "mongodb/docs"
      branch: "main"
      path_transform: "code/${matched_pattern}"
```

#### Common Exclusion Patterns

```yaml
# Exclude hidden files (starting with .)
exclude_patterns:
  - "/\\.[^/]+$"

# Exclude build artifacts
exclude_patterns:
  - "/dist/"
  - "/build/"
  - "\.min\\.(js|css)$"

# Exclude dependencies
exclude_patterns:
  - "node_modules/"
  - "vendor/"
  - "__pycache__/"

# Exclude config files
exclude_patterns:
  - "\.gitignore$"
  - "\.env$"
  - "\.env\\..*$"
  - "config\\.local\\."

# Exclude test files
exclude_patterns:
  - "/test/"
  - "/tests/"
  - "Test\\.java$"
  - "_test\\.go$"
  - "\\.test\\.(js|ts)$"
  - "\\.spec\\.(js|ts)$"

# Exclude documentation
exclude_patterns:
  - "README\\.md$"
  - "\\.md$"
  - "/docs/"
```

#### Regex Syntax Notes

**✅ Supported (Go regex):**
- Character classes: `[abc]`, `[a-z]`, `[^abc]`
- Quantifiers: `*`, `+`, `?`, `{n}`, `{n,}`, `{n,m}`
- Anchors: `^` (start), `$` (end)
- Alternation: `(js|ts|jsx|tsx)`
- Escaping: `\.`, `\(`, `\[`, etc.

**❌ Not Supported:**
- Negative lookahead: `(?!...)` - Use multiple patterns instead
- Lookbehind: `(?<=...)`, `(?<!...)`
- Named groups in exclusions (not needed)

#### Multiple Exclusions

You can specify multiple exclusion patterns. A file is excluded if it matches **any** of them:

```yaml
source_pattern:
  type: "prefix"
  pattern: "mflix/"
  exclude_patterns:
    - "\.gitignore$"           # OR
    - "\.env$"                 # OR
    - "node_modules/"          # OR
    - "/dist/"                 # OR
    - "\.min\\.js$"            # OR
    - "README\\.md$"           # Any match = excluded
```

#### Validation

Exclude patterns are validated when the config is loaded:
- ✅ Must be valid Go regex syntax
- ✅ Cannot be empty strings
- ❌ Invalid regex will cause config validation to fail

## Target Configuration

Defines where and how to copy matched files.

### repo

**Type:** String (required)  
**Format:** `owner/repository`

Target repository to copy files to.

```yaml
repo: "mongodb/docs"
```

### branch

**Type:** String (optional)  
**Default:** `"main"`

Target branch to commit to.

```yaml
branch: "main"
```

### path_transform

**Type:** String (required)

Template for transforming source paths to target paths. Uses `${variable}` syntax.

```yaml
path_transform: "docs/code-examples/${lang}/${file}"
```

**Available Variables:**
- Pattern-extracted variables (from regex named groups)
- Built-in variables (see [Built-in Variables](#built-in-variables))

**Examples:**
```yaml
# Use relative path from prefix match
path_transform: "docs/${relative_path}"

# Use regex-extracted variables
path_transform: "code/${lang}/${category}/${file}"

# Use built-in variables
path_transform: "examples/${dir}/${filename}"

# Combine multiple variables
path_transform: "v${version}/${lang}/examples/${file}"
```

### commit_strategy

**Type:** Object (optional)  
**Default:** `type: "direct"`

Defines how to commit changes. See [Commit Strategies](#commit-strategies).

### deprecation_check

**Type:** Object (optional)

Enables deprecation tracking. See [Deprecation Tracking](#deprecation-tracking).

## Commit Strategies

### Direct Commit

Commits directly to the target branch.

```yaml
commit_strategy:
  type: "direct"
  commit_message: "Update examples from ${source_repo}"
```

**Fields:**
- `type` - Must be `"direct"`
- `commit_message` - (optional) Commit message template

**Use When:**
- You have direct commit access
- Changes don't require review
- Automated updates to documentation

### Pull Request

Creates a pull request for changes.

```yaml
commit_strategy:
  type: "pull_request"
  pr_title: "Update ${lang} examples"
  pr_body: |
    Automated update of ${lang} examples

    Files updated: ${file_count}
    Source: ${source_repo}
    PR: #${pr_number}
  use_pr_template: false
  auto_merge: false
```

**Fields:**
- `type` - Must be `"pull_request"`
- `pr_title` - (optional) PR title template
- `pr_body` - (optional) PR body template
- `use_pr_template` - (optional) Fetch and merge PR template from target repo (default: false)
- `auto_merge` - (optional) Auto-merge if checks pass (default: false)
- `commit_message` - (optional) Commit message template

**Use When:**
- Changes require review
- You want CI checks to run
- Multiple approvers needed

#### Using PR Templates

When `use_pr_template: true`, the service will:
1. Fetch the PR template from the target repository (`.github/pull_request_template.md`)
2. Merge it with your configured `pr_body`
3. Create the PR with the combined content as the **actual PR description** (not a comment)

**Example:**

```yaml
commit_strategy:
  type: "pull_request"
  pr_title: "Update ${lang} examples"
  pr_body: |
    🤖 **Automated Update**

    - Files: ${file_count}
    - Source: ${source_repo}
    - PR: #${pr_number}
  use_pr_template: true  # Fetch template from target repo
  auto_merge: false
```

**Result:** The PR description will contain the target repo's PR template first, followed by your configured content:

```markdown
## Checklist (from target repo's template)

- [ ] Tests added
- [ ] Documentation updated
- [ ] Breaking changes documented

---

🤖 **Automated Update**

- Files: 10
- Source: mongodb/code-examples
- PR: #42
```

**Template Locations Checked (in order):**
1. `.github/pull_request_template.md`
2. `.github/PULL_REQUEST_TEMPLATE.md`
3. `docs/pull_request_template.md`
4. `PULL_REQUEST_TEMPLATE.md`
5. `pull_request_template.md`

**Notes:**
- If no template is found, only the configured `pr_body` is used
- **The PR template appears first**, followed by a separator (`---`), then your configured body
- This ensures the target repo's review guidelines and checklists are prominently displayed
- Templates are fetched from the target repository's branch
- If template fetching fails, a warning is logged but the PR is still created with your configured body
- Works with both individual rules and `batch_pr_config` (when `batch_by_repo: true`)

### Batch Commit

Batches multiple files into fewer commits.

```yaml
commit_strategy:
  type: "batch"
  batch_size: 50
  commit_message: "Update ${file_count} example files"
```

**Fields:**
- `type` - Must be `"batch"`
- `batch_size` - (optional) Files per commit (default: 100)
- `commit_message` - (optional) Commit message template

**Use When:**
- Copying many files
- Want to reduce commit noise
- Grouping related changes

## Deprecation Tracking

Track deprecated files for cleanup.

```yaml
deprecation_check:
  enabled: true
  file: "deprecated_examples.json"
```

**Fields:**
- `enabled` - Enable deprecation tracking
- `file` - (optional) Deprecation file name (default: `deprecated_examples.json`)

**How It Works:**
1. Tracks files copied to target repository
2. Detects when files are removed from source
3. Adds removed files to deprecation file
4. Allows cleanup of obsolete files

**Deprecation File Format:**
```json
{
  "deprecated_files": [
    {
      "path": "docs/examples/old-file.go",
      "deprecated_at": "2024-01-15T10:30:00Z",
      "reason": "Removed from source repository"
    }
  ]
}
```

## Built-in Variables

Available in all path transformations and message templates.

### Path Variables

| Variable      | Description                 | Example               |
|---------------|-----------------------------|-----------------------|
| `${path}`     | Full source path            | `examples/go/main.go` |
| `${filename}` | File name with extension    | `main.go`             |
| `${name}`     | File name without extension | `main`                |
| `${dir}`      | Directory path              | `examples/go`         |
| `${ext}`      | File extension              | `go`                  |

### Pattern Variables

Variables extracted from pattern matching:

**Prefix Pattern:**
- `${matched_prefix}` - The matched prefix
- `${relative_path}` - Path after prefix

**Glob Pattern:**
- `${matched_pattern}` - The pattern that matched

**Regex Pattern:**
- Any named capture group: `${group_name}`

### Message Variables

Available in commit messages, PR titles, and PR body templates:

| Variable           | Description                   | Example                     |
|--------------------|-------------------------------|-----------------------------|
| `${rule_name}`     | Name of the copy rule         | `java-aggregation-examples` |
| `${source_repo}`   | Source repository             | `mongodb/aggregation-tasks` |
| `${target_repo}`   | Target repository             | `mongodb/vector-search`     |
| `${source_branch}` | Source branch                 | `main`                      |
| `${target_branch}` | Target branch                 | `main`                      |
| `${file_count}`    | Number of files               | `3`                         |
| `${pr_number}`     | PR number that triggered copy | `42`                        |
| `${commit_sha}`    | Source commit SHA             | `abc123def456`              |

**Example Usage:**
```yaml
commit_strategy:
  type: "pull_request"
  pr_title: "Update ${lang} examples"
  pr_body: |
    Automated update of ${lang} examples

    **Details:**
    - Rule: ${rule_name}
    - Source: ${source_repo}
    - Files updated: ${file_count}
    - Source PR: #${pr_number}
```

## Complete Examples

### Example 1: Simple Prefix Match

Copy all Go examples to docs repository.

```yaml
source_repo: "mongodb/code-examples"
source_branch: "main"

copy_rules:
  - name: "Copy Go examples"
    source_pattern:
      type: "prefix"
      pattern: "examples/go"
    targets:
      - repo: "mongodb/docs"
        branch: "main"
        path_transform: "source/code/${relative_path}"
        commit_strategy:
          type: "direct"
          commit_message: "Update Go examples"
```

### Example 2: Multi-Language with Regex

Copy examples for multiple languages with categorization.

```yaml
source_repo: "mongodb/code-examples"
source_branch: "main"

copy_rules:
  - name: "Language examples"
    source_pattern:
      type: "regex"
      pattern: "^examples/(?P<lang>[^/]+)/(?P<category>[^/]+)/(?P<file>.+)$"
    targets:
      - repo: "mongodb/docs"
        branch: "main"
        path_transform: "source/code/${lang}/${category}/${file}"
        commit_strategy:
          type: "pull_request"
          pr_title: "Update ${lang} ${category} examples"
          pr_body: |
            Automated update of ${lang} examples
            
            Category: ${category}
            Files: ${file_count}
            Source PR: #${pr_number}
          auto_merge: false
        deprecation_check:
          enabled: true
```

### Example 3: Multiple Targets

Copy same files to multiple repositories.

```yaml
source_repo: "mongodb/code-examples"
source_branch: "main"

copy_rules:
  - name: "Shared examples"
    source_pattern:
      type: "regex"
      pattern: "^shared/(?P<lang>[^/]+)/(?P<file>.+)$"
    targets:
      # Target 1: Main docs
      - repo: "mongodb/docs"
        branch: "main"
        path_transform: "examples/${lang}/${file}"
        commit_strategy:
          type: "direct"
      
      # Target 2: Tutorials
      - repo: "mongodb/tutorials"
        branch: "main"
        path_transform: "code/${lang}/${file}"
        commit_strategy:
          type: "pull_request"
          pr_title: "Update ${lang} examples"
      
      # Target 3: API reference
      - repo: "mongodb/api-docs"
        branch: "main"
        path_transform: "reference/${lang}/examples/${file}"
        commit_strategy:
          type: "direct"
```

### Example 4: Version-Specific Examples

Handle versioned examples with different targets.

```yaml
source_repo: "mongodb/code-examples"
source_branch: "main"

copy_rules:
  - name: "Versioned examples"
    source_pattern:
      type: "regex"
      pattern: "^examples/v(?P<version>[0-9]+)/(?P<lang>[^/]+)/(?P<file>.+)$"
    targets:
      - repo: "mongodb/docs"
        branch: "v${version}"
        path_transform: "source/code/${lang}/${file}"
        commit_strategy:
          type: "pull_request"
          pr_title: "Update v${version} ${lang} examples"
          pr_body: |
            Update ${lang} examples for version ${version}

            Files: ${file_count}
            Source: ${source_repo}
```

### Example 5: Glob Pattern with File Type Filtering

Copy specific file types using glob patterns.

```yaml
source_repo: "mongodb/code-examples"
source_branch: "main"

copy_rules:
  - name: "Go source files"
    source_pattern:
      type: "glob"
      pattern: "examples/**/*.go"
    targets:
      - repo: "mongodb/docs"
        branch: "main"
        path_transform: "code/go/${path}"
        commit_strategy:
          type: "batch"
          batch_size: 25
          commit_message: "Update ${file_count} Go example files"
```

## Validation

### Validate Configuration

Use the `config-validator` tool to validate your configuration:

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
  -source "examples/go/main.go" \
  -template "docs/${lang}/${file}" \
  -var "lang=go" \
  -var "file=main.go"
```

### Validation Rules

**Required Fields:**
- `source_repo` - Must be in `owner/repo` format
- `copy_rules` - At least one rule required
- `copy_rules[].name` - Each rule must have a name
- `copy_rules[].source_pattern.type` - Must be `prefix`, `glob`, or `regex`
- `copy_rules[].source_pattern.pattern` - Pattern string required
- `copy_rules[].targets` - At least one target required
- `targets[].repo` - Must be in `owner/repo` format
- `targets[].path_transform` - Template string required

**Optional Fields with Defaults:**
- `source_branch` - Defaults to `"main"`
- `targets[].branch` - Defaults to `"main"`
- `targets[].commit_strategy.type` - Defaults to `"direct"`
- `deprecation_check.file` - Defaults to `"deprecated_examples.json"`

**Validation Errors:**

```bash
# Missing required field
Error: copy_rules[0]: name is required

# Invalid pattern type
Error: copy_rules[0].source_pattern: invalid pattern type: invalid (must be prefix, glob, or regex)

# Invalid commit strategy
Error: copy_rules[0].targets[0].commit_strategy: invalid type: invalid (must be direct, pull_request, or batch)

# Invalid regex pattern
Error: copy_rules[0].source_pattern: invalid regex pattern: missing closing )
```

## Best Practices

### 1. Use Descriptive Rule Names

```yaml
# Good
name: "Copy Go database examples to docs"

# Bad
name: "rule1"
```

### 2. Start Simple, Then Add Complexity

```yaml
# Start with prefix patterns
source_pattern:
  type: "prefix"
  pattern: "examples/go"

# Graduate to regex when needed
source_pattern:
  type: "regex"
  pattern: "^examples/(?P<lang>[^/]+)/(?P<file>.+)$"
```

### 3. Use Specific Patterns

```yaml
# Good - specific pattern
pattern: "^examples/(?P<lang>go|python|java)/(?P<file>.+\\.(?P<ext>go|py|java))$"

# Risky - too broad
pattern: "^examples/.+$"
```

### 4. Test Patterns Before Deploying

```bash
# Test locally first
./config-validator test-pattern \
  -type regex \
  -pattern "^examples/(?P<lang>[^/]+)/(?P<file>.+)$" \
  -file "examples/go/main.go"

# Validate entire config
./config-validator validate -config copier-config.yaml -v
```

### 5. Use Pull Requests for Important Changes

```yaml
# For production docs
commit_strategy:
  type: "pull_request"
  auto_merge: false

# For staging/dev
commit_strategy:
  type: "direct"
```

### 6. Enable Deprecation Tracking

```yaml
deprecation_check:
  enabled: true
  file: "deprecated_examples.json"
```

### 7. Use Meaningful Commit Messages

```yaml
commit_strategy:
  type: "pull_request"
  pr_title: "Update ${lang} examples - ${file_count} files"
  pr_body: |
    ## Summary
    Automated update of ${lang} code examples

    ## Details
    - Files updated: ${file_count}
    - Source: ${source_repo}
    - Source PR: #${pr_number}
    - Commit: ${commit_sha}

    ## Review Checklist
    - [ ] Examples compile/run correctly
    - [ ] Documentation is up to date
    - [ ] No breaking changes
```

### 8. Organize Rules Logically

```yaml
copy_rules:
  # Group by language
  - name: "Go examples"
    # ...

  - name: "Python examples"
    # ...

  # Or group by target
  - name: "Main docs - all languages"
    # ...

  - name: "Tutorials - all languages"
    # ...
```

### 9. Use Environment-Specific Configs

```bash
# Development
copier-config.dev.yaml

# Staging
copier-config.staging.yaml

# Production
copier-config.yaml
```

### 10. Document Your Configuration

```yaml
# Add comments to explain complex patterns
copy_rules:
  # This rule copies Go examples from the generated-examples directory
  # to the main docs repository. It extracts the project name and
  # preserves the directory structure.
  - name: "Generated Go examples"
    source_pattern:
      type: "regex"
      # Pattern: generated-examples/{project}/{rest-of-path}
      pattern: "^generated-examples/(?P<project>[^/]+)/(?P<rest>.+)$"
    targets:
      - repo: "mongodb/docs"
        branch: "main"
        # Transform: examples/{project}/{rest-of-path}
        path_transform: "examples/${project}/${rest}"
```

## Configuration File Location

### Default Location

The application looks for `copier-config.yaml` in:
1. Current directory
2. Source repository (fetched from GitHub)

### Custom Location

Use the `CONFIG_FILE` environment variable:

```bash
# Use custom config file
export CONFIG_FILE=my-config.yaml
./examples-copier

# Use environment-specific config
export CONFIG_FILE=copier-config.production.yaml
./examples-copier
```

### Local vs Remote

**Local File (for testing):**
```bash
# Create local config
cp configs/copier-config.example.yaml copier-config.yaml

# Edit and test
vim copier-config.yaml
./examples-copier
```

**Remote File (for production):**
```bash
# Add config to source repository
git add copier-config.yaml
git commit -m "Add copier configuration"
git push origin main

# Application fetches from GitHub
./examples-copier
```

## Troubleshooting

### Config Not Found

**Error:**
```
[ERROR] failed to load config | {"error":"failed to retrieve config file: 404 Not Found"}
```

**Solutions:**
1. Create local config file: `copier-config.yaml`
2. Add config to source repository
3. Check `CONFIG_FILE` environment variable
4. Verify file name matches exactly

### Invalid Pattern

**Error:**
```
Error: copy_rules[0].source_pattern: invalid regex pattern
```

**Solutions:**
1. Test pattern with `config-validator`
2. Check regex syntax
3. Escape special characters
4. Use raw strings for complex patterns

### Path Transform Failed

**Error:**
```
[ERROR] failed to transform path | {"error":"variable not found: lang"}
```

**Solutions:**
1. Verify variable is extracted by pattern
2. Check variable name spelling
3. Test with `config-validator test-transform`
4. Use built-in variables if pattern variables unavailable

### Validation Failed

**Error:**
```
Error: copy_rules[0]: name is required
```

**Solutions:**
1. Run `config-validator validate -config copier-config.yaml -v`
2. Check all required fields are present
3. Verify YAML syntax is correct
4. Check indentation (YAML is whitespace-sensitive)

# Pattern Matching Cheat Sheet

Quick reference for pattern matching in examples-copier.

## Pattern Types at a Glance

| Type       | Use When                              | Example                         | Extracts Variables?           |
|------------|---------------------------------------|---------------------------------|-------------------------------|
| **Prefix** | Simple directory matching             | `examples/`                     | ✅ Yes (prefix, relative_path) |
| **Glob**   | Wildcard matching                     | `**/*.go`                       | ❌ No                          |
| **Regex**  | Complex patterns, variable extraction | `^examples/(?P<lang>[^/]+)/.*$` | ✅ Yes (custom)                |

## Prefix Patterns

### Syntax
```yaml
source_pattern:
  type: "prefix"
  pattern: "examples/"
```

### Examples
| Pattern     | Matches               | Doesn't Match          |
|-------------|-----------------------|------------------------|
| `examples/` | `examples/go/main.go` | `src/examples/test.go` |
| `src/`      | `src/main.go`         | `examples/src/test.go` |
| `docs/api/` | `docs/api/readme.md`  | `docs/guide/api.md`    |

### Variables
- `${matched_prefix}` - The matched prefix
- `${relative_path}` - Path after the prefix

## Glob Patterns

### Wildcards
| Symbol | Matches                 | Example                     |
|--------|-------------------------|-----------------------------|
| `*`    | Any characters (no `/`) | `*.go` → `main.go`          |
| `**`   | Any directories         | `**/*.go` → `a/b/c/main.go` |
| `?`    | Single character        | `test?.go` → `test1.go`     |

### Examples
| Pattern            | Matches                | Doesn't Match |
|--------------------|------------------------|---------------|
| `*.go`             | `main.go`              | `src/main.go` |
| `**/*.go`          | `a/b/c/main.go`        | `main.py`     |
| `examples/**/*.js` | `examples/node/app.js` | `src/app.js`  |
| `test?.go`         | `test1.go`, `testA.go` | `test12.go`   |

## Regex Patterns

### Common Building Blocks

| Pattern      | Matches                     | Example                |
|--------------|-----------------------------|------------------------|
| `[^/]+`      | One or more non-slash chars | Directory or file name |
| `.+`         | One or more any chars       | Rest of path           |
| `.*`         | Zero or more any chars      | Optional content       |
| `[0-9]+`     | One or more digits          | Version numbers        |
| `(foo\|bar)` | Either foo or bar           | Specific values        |
| `\.go$`      | Ends with .go               | File extension         |
| `^examples/` | Starts with examples/       | Path prefix            |

### Named Capture Groups

```regex
(?P<name>pattern)
```

**Example:**
```regex
^examples/(?P<lang>[^/]+)/(?P<file>.+)$
```

Extracts:
- `lang` from first directory
- `file` from rest of path

### Common Patterns

#### Language + File
```regex
^examples/(?P<lang>[^/]+)/(?P<file>.+)$
```
- `examples/go/main.go` → `lang=go, file=main.go`

#### Language + Category + File
```regex
^examples/(?P<lang>[^/]+)/(?P<category>[^/]+)/(?P<file>.+)$
```
- `examples/go/database/connect.go` → `lang=go, category=database, file=connect.go`

#### Project + Rest
```regex
^generated-examples/(?P<project>[^/]+)/(?P<rest>.+)$
```
- `generated-examples/app/cmd/main.go` → `project=app, rest=cmd/main.go`

#### Version Support
```regex
^examples/(?P<lang>[^/]+)/(?P<version>v[0-9]+\\.x)/(?P<file>.+)$
```
- `examples/node/v6.x/app.js` → `lang=node, version=v6.x, file=app.js`

#### Type + Language + File
```regex
^source/examples/(?P<type>generated|manual)/(?P<lang>[^/]+)/(?P<file>.+)$
```
- `source/examples/generated/node/app.js` → `type=generated, lang=node, file=app.js`

## Path Transformation

### Syntax
```yaml
path_transform: "docs/${lang}/${file}"
```

### Built-in Variables

| Variable      | Value for `examples/go/database/connect.go` |
|---------------|---------------------------------------------|
| `${path}`     | `examples/go/database/connect.go`           |
| `${filename}` | `connect.go`                                |
| `${dir}`      | `examples/go/database`                      |
| `${ext}`      | `.go`                                       |
| `${name}`     | `connect`                                   |

### Common Transformations

| Transform                          | Input                    | Output                     |
|------------------------------------|--------------------------|----------------------------|
| `${path}`                          | `examples/go/main.go`    | `examples/go/main.go`      |
| `docs/${path}`                     | `examples/go/main.go`    | `docs/examples/go/main.go` |
| `docs/${relative_path}`            | `examples/go/main.go`    | `docs/go/main.go`          |
| `${lang}/${file}`                  | `examples/go/main.go`    | `go/main.go`               |
| `docs/${lang}/${category}/${file}` | `examples/go/db/conn.go` | `docs/go/db/conn.go`       |

## Complete Examples

### Example 1: Simple Copy
```yaml
source_pattern:
  type: "prefix"
  pattern: "examples/"
targets:
  - path_transform: "docs/${path}"
```
**Result:** `examples/go/main.go` → `docs/examples/go/main.go`

### Example 2: Language-Based
```yaml
source_pattern:
  type: "regex"
  pattern: "^examples/(?P<lang>[^/]+)/(?P<file>.+)$"
targets:
  - path_transform: "docs/code-examples/${lang}/${file}"
```
**Result:** `examples/go/main.go` → `docs/code-examples/go/main.go`

### Example 3: Categorized
```yaml
source_pattern:
  type: "regex"
  pattern: "^examples/(?P<lang>[^/]+)/(?P<category>[^/]+)/(?P<file>.+)$"
targets:
  - path_transform: "docs/${lang}/${category}/${file}"
```
**Result:** `examples/go/database/connect.go` → `docs/go/database/connect.go`

### Example 4: Glob for Extensions
```yaml
source_pattern:
  type: "glob"
  pattern: "examples/**/*.go"
targets:
  - path_transform: "docs/${path}"
```
**Result:** `examples/go/auth/login.go` → `docs/examples/go/auth/login.go`

### Example 5: Project-Based
```yaml
source_pattern:
  type: "regex"
  pattern: "^generated-examples/(?P<project>[^/]+)/(?P<rest>.+)$"
targets:
  - path_transform: "examples/${project}/${rest}"
```
**Result:** `generated-examples/app/cmd/main.go` → `examples/app/cmd/main.go`

## Testing Commands

### Test Pattern
```bash
./config-validator test-pattern \
  -type regex \
  -pattern "^examples/(?P<lang>[^/]+)/(?P<file>.+)$" \
  -file "examples/go/main.go"
```

### Test Transform
```bash
./config-validator test-transform \
  -source "examples/go/main.go" \
  -template "docs/${lang}/${file}" \
  -vars "lang=go,file=main.go"
```

### Validate Config
```bash
./config-validator validate -config copier-config.yaml -v
```

## Decision Tree

```
What do you need?
│
├─ Copy entire directory tree
│  └─ Use PREFIX pattern
│     pattern: "examples/"
│     transform: "docs/${path}"
│
├─ Match by file extension
│  └─ Use GLOB pattern
│     pattern: "**/*.go"
│     transform: "docs/${path}"
│
├─ Extract language from path
│  └─ Use REGEX pattern
│     pattern: "^examples/(?P<lang>[^/]+)/(?P<file>.+)$"
│     transform: "docs/${lang}/${file}"
│
└─ Complex matching with multiple variables
   └─ Use REGEX pattern
      pattern: "^examples/(?P<lang>[^/]+)/(?P<category>[^/]+)/(?P<file>.+)$"
      transform: "docs/${lang}/${category}/${file}"
```

## Common Mistakes

### ❌ Missing Anchors
```yaml
# Wrong - matches partial paths
pattern: "examples/(?P<lang>[^/]+)/(?P<file>.+)"

# Right - matches full path
pattern: "^examples/(?P<lang>[^/]+)/(?P<file>.+)$"
```

### ❌ Wrong Character Class
```yaml
# Wrong - .+ matches slashes too
pattern: "^examples/(?P<lang>.+)/(?P<file>.+)$"
# Right - [^/]+ doesn't match slashes
pattern: "^examples/(?P<lang>[^/]+)/(?P<file>.+)$"
```

### ❌ Unnamed Groups
```yaml
# Wrong - doesn't extract variables
pattern: "^examples/([^/]+)/(.+)$"

# Right - named groups extract variables
pattern: "^examples/(?P<lang>[^/]+)/(?P<file>.+)$"
```

### ❌ Variable Name Mismatch
```yaml
# Pattern extracts "lang"
pattern: "^examples/(?P<lang>[^/]+)/(?P<file>.+)$"

# Wrong - uses "language"
path_transform: "docs/${language}/${file}"

# Right - uses "lang"
path_transform: "docs/${lang}/${file}"
```

## Tips

1. **Start simple** - Use prefix, then add regex when needed
2. **Test first** - Use `config-validator` before deploying
3. **Use anchors** - Always use `^` and `$` in regex
4. **Be specific** - Use `[^/]+` instead of `.+` for directories
5. **Name clearly** - Use descriptive variable names like `lang`, not `a`
6. **Check logs** - Look for "sample file path" to see actual paths

## See Also

- [Pattern Matching Guide](PATTERN-MATCHING-GUIDE.md) - Detailed pattern matching documentation
- [Quick Reference](../QUICK-REFERENCE.md) - Command reference
- [Deployment Guide](DEPLOYMENT.md) - Deploying the application
- [Architecture](ARCHITECTURE.md) - System architecture overview

---

**Need Help?**
- See [Troubleshooting Guide](TROUBLESHOOTING.md)
- See [FAQ](FAQ.md)
- Check example configs in `configs/copier-config.example.yaml`

