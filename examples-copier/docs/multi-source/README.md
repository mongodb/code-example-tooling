# Multi-Source Repository Support

## Overview

This feature enables the examples-copier to monitor and process webhooks from **multiple source repositories** across **multiple GitHub organizations** using a **centralized configuration** approach.

### Use Case

Perfect for teams managing code examples across multiple repositories and organizations:

```
Sources (monitored repos):
├── 10gen/docs-mongodb-internal
├── mongodb/docs-sample-apps
└── mongodb/docs-code-examples

Targets (destination repos):
├── mongodb/docs
├── mongodb/docs-realm
├── mongodb/developer-hub
└── 10gen/docs-mongodb-internal
```

### Key Features

✅ **Centralized Configuration** - One config file manages all sources  
✅ **Multi-Organization Support** - Works across mongodb, 10gen, mongodb-university orgs  
✅ **Cross-Org Copying** - Copy from mongodb → 10gen or vice versa  
✅ **Single Deployment** - One app instance handles all sources  
✅ **100% Backward Compatible** - Existing single-source configs still work  

## Quick Start

### 1. Configuration Repository Setup

Store your config in a dedicated repository:

```
Repository: mongodb-university/code-example-tooling
File: copier-config.yaml
```

### 2. Environment Variables

```bash
# Config Repository
CONFIG_REPO_OWNER=mongodb-university
CONFIG_REPO_NAME=code-example-tooling
CONFIG_FILE=copier-config.yaml

# GitHub App Installations (one per org)
MONGODB_INSTALLATION_ID=<from-mongodb-org>
TENGEN_INSTALLATION_ID=<from-10gen-org>
MONGODB_UNIVERSITY_INSTALLATION_ID=<from-mongodb-university-org>
```

### 3. Example Configuration

```yaml
# File: mongodb-university/code-example-tooling/copier-config.yaml

sources:
  # Source from 10gen org
  - repo: "10gen/docs-mongodb-internal"
    branch: "main"
    installation_id: "${TENGEN_INSTALLATION_ID}"
    copy_rules:
      - name: "internal-to-public"
        source_pattern:
          type: "prefix"
          pattern: "examples/"
        targets:
          - repo: "mongodb/docs"
            branch: "main"
            path_transform: "source/code/${relative_path}"
            commit_strategy:
              type: "pull_request"
              pr_title: "Update examples from internal docs"

  # Source from mongodb org
  - repo: "mongodb/docs-code-examples"
    branch: "main"
    installation_id: "${MONGODB_INSTALLATION_ID}"
    copy_rules:
      - name: "examples-to-internal"
        source_pattern:
          type: "prefix"
          pattern: "public/"
        targets:
          - repo: "10gen/docs-mongodb-internal"
            branch: "main"
            path_transform: "external-examples/${relative_path}"
            commit_strategy:
              type: "direct"
```

### 4. GitHub App Installation

Install the GitHub App in **all three organizations**:

1. **mongodb** - for mongodb/* repos (source and target)
2. **10gen** - for 10gen/* repos (source and target)
3. **mongodb-university** - for the config repo

## Documentation

| Document | Purpose |
|----------|---------|
| **[Implementation Plan](MULTI-SOURCE-IMPLEMENTATION-PLAN.md)** | Detailed implementation guide for developers |
| **[Technical Spec](MULTI-SOURCE-TECHNICAL-SPEC.md)** | Technical specifications and architecture |
| **[Migration Guide](MULTI-SOURCE-MIGRATION-GUIDE.md)** | How to migrate from single-source to multi-source |
| **[Quick Reference](MULTI-SOURCE-QUICK-REFERENCE.md)** | Common tasks and troubleshooting |

## Architecture

### Centralized Configuration Approach

```
Config Repo (mongodb-university/code-example-tooling)
    │
    ├─ copier-config.yaml (manages all sources)
    │
    ├─ Sources:
    │  ├─ 10gen/docs-mongodb-internal
    │  ├─ mongodb/docs-sample-apps
    │  └─ mongodb/docs-code-examples
    │
    └─ Targets:
       ├─ mongodb/docs
       ├─ mongodb/docs-realm
       ├─ mongodb/developer-hub
       └─ 10gen/docs-mongodb-internal
```

### Webhook Flow

```
1. Webhook arrives from mongodb/docs-code-examples
   ↓
2. App loads config from mongodb-university/code-example-tooling
   ↓
3. Router identifies source repo in config
   ↓
4. Switches to MONGODB_INSTALLATION_ID
   ↓
5. Reads changed files from source
   ↓
6. For each target:
   - Switches to target org's installation ID
   - Writes files to target repo
```

## Key Differences from Original Plan

This implementation focuses on **centralized configuration** for a **single team** managing multiple repos across organizations:

| Feature | This Implementation | Original Plan |
|---------|-------------------|---------------|
| **Config Storage** | Centralized (one file) | Centralized OR distributed |
| **Config Location** | Dedicated repo (3rd org) | Source repo or central |
| **Use Case** | Single team, multi-org | General purpose |
| **Complexity** | Simplified | Full-featured |
| **Multi-Tenant** | No (not needed) | Future enhancement |

## Benefits

### For MongoDB Docs Team

1. **Single Source of Truth** - All copy rules in one config file
2. **Easy to Understand** - See all flows at a glance
3. **Centralized Management** - No need to update multiple repos
4. **Cross-Org Support** - Built-in support for mongodb ↔ 10gen flows
5. **Simple Deployment** - One app instance for everything

### Operational

1. **Reduced Infrastructure** - One deployment instead of multiple
2. **Unified Monitoring** - All metrics and logs in one place
3. **Easier Debugging** - Single config to check
4. **Better Visibility** - See all copy operations together

## Implementation Status

| Component | Status |
|-----------|--------|
| Documentation | ✅ Complete |
| Implementation Plan | ✅ Complete |
| Technical Spec | ✅ Complete |
| Migration Guide | ✅ Complete |
| Code Implementation | ⏳ Pending |
| Testing | ⏳ Pending |
| Deployment | ⏳ Pending |

## Next Steps

1. Review the [Implementation Plan](MULTI-SOURCE-IMPLEMENTATION-PLAN.md)
2. Set up GitHub App installations in all three orgs
3. Create config repository structure
4. Begin implementation (Phase 1: Core Infrastructure)
5. Test with staging environment
6. Deploy to production

## Support

For questions or issues:

1. Check the [Quick Reference](MULTI-SOURCE-QUICK-REFERENCE.md)
2. Review the [Migration Guide](MULTI-SOURCE-MIGRATION-GUIDE.md) FAQ
3. Consult the [Technical Spec](MULTI-SOURCE-TECHNICAL-SPEC.md)

---

**Configuration Approach**: Centralized  
**Target Use Case**: MongoDB Docs Team (mongodb, 10gen, mongodb-university orgs)  
**Status**: Ready for Implementation  
**Last Updated**: 2025-10-15

