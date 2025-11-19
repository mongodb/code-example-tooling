# Quick Start: Main Config Architecture

Get started with the new main config architecture in 5 minutes.

## What is Main Config?

Main config is a centralized configuration system that:
- Stores global defaults in one place
- References workflow configs in source repositories
- Supports reusable components
- Enables distributed workflow management

## Quick Setup

### 1. Update env.yaml

Add these variables to your `env.yaml`:

```yaml
env_variables:
  # Enable main config
  MAIN_CONFIG_FILE: "main-config.yaml"
  USE_MAIN_CONFIG: "true"
  
  # Config repository (where main config lives)
  CONFIG_REPO_OWNER: "mongodb"
  CONFIG_REPO_NAME: "code-copier-config"
  CONFIG_REPO_BRANCH: "main"
```

### 2. Create Main Config

Create `main-config.yaml` in your config repository:

```yaml
# Global defaults
defaults:
  commit_strategy:
    type: "pull_request"
    auto_merge: false
  exclude:
    - "**/.env"
    - "**/.env.*"

# Workflow references
workflow_configs:
  # Reference workflow config in source repo
  - source: "repo"
    repo: "mongodb/docs-sample-apps"
    branch: "main"
    path: ".copier/workflows.yaml"
```

### 3. Create Workflow Config in Source Repo

Create `.copier/workflows.yaml` in your source repository:

```yaml
workflows:
  - name: "my-first-workflow"
    source:
      repo: "mongodb/docs-sample-apps"
      branch: "main"
    destination:
      repo: "mongodb/my-destination"
      branch: "main"
    transformations:
      - move:
          from: "examples"
          to: "code-examples"
```

### 4. Deploy and Test

```bash
# Deploy the app
gcloud app deploy app.yaml

# Test by merging a PR in your source repo
# The workflow should execute automatically
```

## Common Patterns

### Pattern 1: Centralized Workflows

Keep all workflows in the config repo:

```yaml
# main-config.yaml
workflow_configs:
  - source: "local"
    path: "workflows/mflix-workflows.yaml"
  - source: "local"
    path: "workflows/university-workflows.yaml"
```

**Use when**: Central team manages all workflows

### Pattern 2: Distributed Workflows

Keep workflows in source repos:

```yaml
# main-config.yaml
workflow_configs:
  - source: "repo"
    repo: "mongodb/docs-sample-apps"
    path: ".copier/workflows.yaml"
  - source: "repo"
    repo: "10gen/university-content"
    path: ".copier/workflows.yaml"
```

**Use when**: Source repo teams manage their own workflows

### Pattern 3: Hybrid Approach

Mix centralized and distributed:

```yaml
# main-config.yaml
workflow_configs:
  # Centralized (managed by central team)
  - source: "local"
    path: "workflows/critical-workflows.yaml"
  
  # Distributed (managed by source teams)
  - source: "repo"
    repo: "mongodb/docs-sample-apps"
    path: ".copier/workflows.yaml"
  
  # Inline (simple one-offs)
  - source: "inline"
    workflows:
      - name: "simple-copy"
        source:
          repo: "mongodb/docs"
          branch: "main"
        destination:
          repo: "mongodb/docs-public"
          branch: "main"
        transformations:
          - move: { from: "examples", to: "public-examples" }
```

**Use when**: You need flexibility

## Directory Structure

### Recommended Structure

```
Config Repo (mongodb/code-copier-config):
├── main-config.yaml              # Main config file
└── workflows/                    # Centralized workflows (optional)
    ├── mflix-workflows.yaml
    └── university-workflows.yaml

Source Repo (mongodb/docs-sample-apps):
└── .copier/                      # Workflow config directory
    ├── workflows.yaml            # Workflow definitions
    ├── transformations/          # Reusable transformations
    │   ├── mflix-java.yaml
    │   └── mflix-nodejs.yaml
    ├── strategies/               # Reusable strategies
    │   └── mflix-pr-strategy.yaml
    └── common/                   # Common configs
        └── mflix-excludes.yaml
```

## Default Precedence

Settings are applied from least to most specific:

```
System Defaults
    ↓
Main Config Defaults
    ↓
Workflow Config Defaults
    ↓
Individual Workflow Settings (wins)
```

Example:

```yaml
# main-config.yaml
defaults:
  commit_strategy:
    type: "pull_request"
    auto_merge: false  # Global default

# .copier/workflows.yaml
defaults:
  commit_strategy:
    auto_merge: true  # Overrides main config

workflows:
  - name: "my-workflow"
    commit_strategy:
      auto_merge: false  # Overrides workflow config default
```

## Troubleshooting

### Config Not Loading

**Problem**: App can't find main config file

**Solution**:
1. Check `MAIN_CONFIG_FILE` is set correctly
2. Verify file exists in config repository
3. Check `CONFIG_REPO_OWNER` and `CONFIG_REPO_NAME`
4. Review app logs for errors

### Workflows Not Executing

**Problem**: Workflows don't run when PR is merged

**Solution**:
1. Verify workflow `source.repo` matches webhook repo
2. Check workflow config is referenced in main config
3. Validate YAML syntax
4. Check app logs for validation errors

### Authentication Errors

**Problem**: Can't access source or destination repos

**Solution**:
1. Verify GitHub App has access to all repos
2. Check app is installed in all required orgs
3. Verify app permissions include repo read/write

## Next Steps

1. **Read the full guide**: See `MAIN-CONFIG-README.md`
2. **Review examples**: Check `main-config-example.yaml`
3. **Test locally**: Use `DRY_RUN=true` for testing
4. **Add more workflows**: Expand your configuration
5. **Use reusable components**: Extract common configs

## Migration from Legacy Format

Already using `copier-config.yaml`? You have options:

### Option 1: Keep Legacy Format
Don't change anything - legacy format still works!

### Option 2: Migrate Gradually
1. Set `MAIN_CONFIG_FILE` to new file
2. Use inline workflows initially
3. Gradually move to separate files

### Option 3: Full Migration
1. Create main config with workflow references
2. Move workflows to source repos
3. Update env.yaml
4. Test thoroughly
5. Deploy

## Support

- **Documentation**: `MAIN-CONFIG-README.md`
- **Examples**: `main-config-example.yaml`, `source-repo-workflows-example.yaml`
- **Reusable Components**: `reusable-components/` directory

## Key Benefits

✅ **Separation of Concerns** - Each repo manages its own workflows  
✅ **Scalability** - Works for monorepos with many workflows  
✅ **Flexibility** - Mix centralized and distributed configs  
✅ **Discoverability** - Configs live near source code  
✅ **Maintainability** - Update workflows without touching main config  

## Common Use Cases

### Use Case 1: Monorepo with Many Workflows

**Problem**: 50+ workflows in one config file  
**Solution**: Split into multiple workflow config files

```yaml
workflow_configs:
  - source: "local"
    path: "workflows/mflix-workflows.yaml"  # 10 workflows
  - source: "local"
    path: "workflows/university-workflows.yaml"  # 15 workflows
  - source: "local"
    path: "workflows/docs-workflows.yaml"  # 25 workflows
```

### Use Case 2: Multiple Source Repos

**Problem**: Workflows for different source repos mixed together  
**Solution**: Each source repo has its own workflow config

```yaml
workflow_configs:
  - source: "repo"
    repo: "mongodb/docs-sample-apps"
    path: ".copier/workflows.yaml"
  - source: "repo"
    repo: "10gen/university-content"
    path: ".copier/workflows.yaml"
```

### Use Case 3: Team Ownership

**Problem**: Multiple teams need to manage workflows  
**Solution**: Each team manages workflows in their source repo

```yaml
workflow_configs:
  # Team A's workflows
  - source: "repo"
    repo: "mongodb/team-a-repo"
    path: ".copier/workflows.yaml"
  
  # Team B's workflows
  - source: "repo"
    repo: "mongodb/team-b-repo"
    path: ".copier/workflows.yaml"
```

---

**Ready to get started?** Follow the steps above and you'll be up and running in minutes!

