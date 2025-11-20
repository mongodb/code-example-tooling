# Main Config Architecture Guide

This guide explains the main config architecture that supports centralized configuration with distributed workflow definitions.

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Configuration Files](#configuration-files)
- [Getting Started](#getting-started)
- [Migration Guide](#migration-guide)
- [Best Practices](#best-practices)
- [Examples](#examples)

## Overview

The main config architecture introduces a hierarchical configuration system that separates global settings from workflow-specific configurations. This enables:

- **Centralized defaults** in a main config file
- **Distributed workflows** in source repositories
- **Reusable components** for transformations, strategies, and excludes
- **Clear ownership** of workflow configurations

## Architecture

### Three-Tier Configuration

```
Main Config (Central)
  ├── Global Defaults
  └── Workflow Config References
        ├── Local Workflow Configs (same repo)
        ├── Remote Workflow Configs (source repos)
        └── Inline Workflows (simple cases)
              └── Individual Workflows
```

### Default Precedence

Settings are applied in order of specificity (most specific wins):

1. **Individual Workflow** settings (highest priority)
2. **Workflow Config** defaults
3. **Main Config** defaults
4. **System** defaults (lowest priority)

## Configuration Files

### 1. Main Config File

**Location**: Specified in `env.yaml` as `MAIN_CONFIG_FILE`  
**Purpose**: Central configuration with global defaults and workflow references

```yaml
# main-config.yaml
defaults:
  commit_strategy:
    type: "pull_request"
    auto_merge: false
  exclude:
    - "**/.env"
    - "**/node_modules/**"

workflow_configs:
  - source: "repo"
    repo: "mongodb/docs-sample-apps"
    path: ".copier/workflows.yaml"
```

### 2. Workflow Config Files

**Location**: In source repositories (e.g., `.copier/workflows.yaml`)  
**Purpose**: Define workflows for a specific source repository

```yaml
# .copier/workflows.yaml
defaults:
  commit_strategy:
    type: "pull_request"

workflows:
  - name: "mflix-java"
    source:
      repo: "mongodb/docs-sample-apps"
      branch: "main"
    destination:
      repo: "mongodb/sample-app-java-mflix"
      branch: "main"
    transformations:
      - move: { from: "mflix/client", to: "client" }
```

### 3. Reusable Component Files

**Location**: In source repositories (e.g., `.copier/transformations/`)  
**Purpose**: Extract common configurations for reuse

```yaml
# .copier/transformations/mflix-java.yaml
- move: { from: "mflix/client", to: "client" }
- move: { from: "mflix/server/java-spring", to: "server" }
```

## Getting Started

### Step 1: Set Environment Variables

Add to your `env.yaml`:

```yaml
env_variables:
  # Main config settings
  MAIN_CONFIG_FILE: "main-config.yaml"
  USE_MAIN_CONFIG: "true"
  
  # Config repository
  CONFIG_REPO_OWNER: "mongodb"
  CONFIG_REPO_NAME: "code-copier-config"
  CONFIG_REPO_BRANCH: "main"
```

### Step 2: Create Main Config

Create `main-config.yaml` in your config repository:

```yaml
defaults:
  commit_strategy:
    type: "pull_request"
    auto_merge: false
  exclude:
    - "**/.env"
    - "**/.env.*"

workflow_configs:
  - source: "repo"
    repo: "mongodb/docs-sample-apps"
    branch: "main"
    path: ".copier/workflows.yaml"
```

### Step 3: Create Workflow Config in Source Repo

Create `.copier/workflows.yaml` in your source repository:

```yaml
workflows:
  - name: "my-workflow"
    source:
      repo: "mongodb/docs-sample-apps"
      branch: "main"
    destination:
      repo: "mongodb/my-destination"
      branch: "main"
    transformations:
      - move: { from: "src", to: "dest" }
```

### Step 4: Deploy and Test

1. Deploy the updated configuration
2. Trigger a webhook event (merge a PR in source repo)
3. Verify workflows execute correctly

## Best Practices

### 1. Organization

**Recommended directory structure**:

```
Config Repo (mongodb/code-copier-config):
  main-config.yaml
  workflows/
    mflix-workflows.yaml
    university-workflows.yaml

Source Repo (mongodb/docs-sample-apps):
  .copier/
    workflows.yaml
    transformations/
      mflix-java.yaml
      mflix-nodejs.yaml
    strategies/
      mflix-pr-strategy.yaml
    common/
      mflix-excludes.yaml
```

### 2. Workflow Config Placement

- **Centralized**: Use `source: "local"` for workflows managed by central team
- **Distributed**: Use `source: "repo"` for workflows managed by source repo teams
- **Simple**: Use `source: "inline"` for one-off or simple workflows

### 3. Reusable Components

Extract common configurations:

- **Transformations**: When multiple workflows use similar file mappings
- **Strategies**: When multiple workflows use the same PR format
- **Excludes**: When multiple workflows exclude the same patterns

### 4. Default Strategy

Set sensible defaults at each level:

- **Main config**: Organization-wide defaults
- **Workflow config**: Source repo defaults
- **Individual workflow**: Workflow-specific overrides

### 5. Testing

- Test workflow configs in source repo PRs
- Use `DRY_RUN=true` for testing without side effects
- Validate configurations before deploying

## Examples

See the example files in this directory:

- `main-config-example.yaml` - Complete main config example
- `source-repo-workflows-example.yaml` - Workflow config in source repo
- `reusable-components/` - Examples of reusable components
  - `transformations-example.yaml`
  - `strategy-example.yaml`
  - `excludes-example.yaml`

## Reference Syntax

### Workflow Config References

```yaml
# Local file in config repo
- source: "local"
  path: "workflows/my-workflows.yaml"

# Remote file in source repo
- source: "repo"
  repo: "owner/repo"
  branch: "main"
  path: ".copier/workflows.yaml"

# Inline workflows
- source: "inline"
  workflows:
    - name: "my-workflow"
      # ... workflow definition ...
```

### Component References

You can use `$ref` to reference external files for transformations, commit_strategy, and exclude patterns:

```yaml
# Reference transformations
transformations:
  $ref: "transformations/mflix-java.yaml"

# Reference strategy
commit_strategy:
  $ref: "strategies/mflix-pr-strategy.yaml"

# Reference excludes
exclude:
  $ref: "common/mflix-excludes.yaml"
```

**Benefits:**
- Share common configurations across multiple workflows
- Keep workflow configs clean and focused
- Organize related files in a logical directory structure

**Path Resolution:**
- Relative paths are resolved relative to the workflow config file
- Example: If your workflow config is at `.copier/workflows.yaml`, then `transformations/mflix-java.yaml` resolves to `.copier/transformations/mflix-java.yaml`

## Troubleshooting

### Config Not Loading

- Check `MAIN_CONFIG_FILE` is set correctly
- Verify `USE_MAIN_CONFIG=true`
- Check file exists in config repository
- Review logs for parsing errors

### Workflows Not Executing

- Verify workflow source repo matches webhook repo
- Check workflow config is referenced in main config
- Validate workflow config syntax
- Review logs for validation errors

### Authentication Issues

- Ensure GitHub App has access to all repos
- Verify installation IDs are correct
- Check app permissions

## Support

For questions or issues:

1. Check the example configurations
2. Review the logs for error messages
3. Validate your configuration syntax
4. Test with `DRY_RUN=true`


