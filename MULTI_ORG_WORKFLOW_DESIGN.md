# Multi-Org + Workflow Design Proposal

## Executive Summary

You want to implement **two major enhancements**:
1. **Simplified workflow-based config** (reduce complexity by 70%)
2. **Multi-org installation support** (install app in multiple GitHub orgs)

**Good news:** We can design both together and avoid rework! 🎉

---

## Current Multi-Org Support

Your app **already has partial multi-org support**:

<augment_code_snippet path="examples-copier/services/github_auth.go" mode="EXCERPT">
````go
// GetRestClientForOrg returns a GitHub REST API client authenticated for a specific organization
func GetRestClientForOrg(org string) (*github.Client, error) {
    // Check if we have a cached token for this org
    if token, ok := installationTokenCache[org]; ok && token != "" {
        // ... return cached client
    }
    
    // Get installation ID for the organization
    installationID, err := getInstallationIDForOrg(org)
    // ... get token and create client
}
````
</augment_code_snippet>

**What works:**
- ✅ Can authenticate to different orgs dynamically
- ✅ Caches installation tokens per org
- ✅ Auto-discovers installation IDs

**What's missing:**
- ❌ Config assumes single source repo (`source_repo` field)
- ❌ Webhook handler validates against single source
- ❌ No way to configure multiple source repos

---

## Design Considerations for Multi-Org

### 1. **Config Structure: Per-Workflow vs Global**

**Option A: Global Source (Current)**
```yaml
source_repo: "mongodb/docs-sample-apps"  # ❌ Single source only
workflows:
  - name: "mflix-java"
    destination: { repo: "mongodb/sample-app-java" }
```

**Option B: Per-Workflow Source (Recommended)**
```yaml
workflows:
  - name: "mflix-java"
    source:
      repo: "mongodb/docs-sample-apps"
      branch: "main"
    destination:
      repo: "mongodb/sample-app-java"
      branch: "main"
    transformations: [...]
  
  - name: "university-python"
    source:
      repo: "10gen/university-content"  # ✅ Different org!
      branch: "main"
    destination:
      repo: "mongodb-university/python-course"
      branch: "main"
    transformations: [...]
```

**Why Option B?**
- ✅ Each workflow can have different source repo
- ✅ Supports multi-org installations naturally
- ✅ More flexible (different branches per workflow)
- ✅ Clearer intent (explicit source/destination)
- ✅ Matches Copybara's design

---

### 2. **Webhook Routing: Single vs Multiple Sources**

**Current Behavior:**
```go
// Validate webhook is from expected source repository
webhookRepo := fmt.Sprintf("%s/%s", repoOwner, repoName)
if webhookRepo != yamlConfig.SourceRepo {
    // ❌ Reject webhook
    return
}
```

**Problem:** Only accepts webhooks from ONE source repo.

**Solution: Workflow-Based Routing**
```go
// Find all workflows that match this source repo
matchingWorkflows := []Workflow{}
for _, workflow := range yamlConfig.Workflows {
    if workflow.Source.Repo == webhookRepo {
        matchingWorkflows = append(matchingWorkflows, workflow)
    }
}

if len(matchingWorkflows) == 0 {
    LogWarning("No workflows configured for source repo: " + webhookRepo)
    return
}

// Process files for each matching workflow
for _, workflow := range matchingWorkflows {
    processWorkflow(ctx, workflow, changedFiles)
}
```

**Benefits:**
- ✅ Accepts webhooks from multiple source repos
- ✅ Each source repo can have multiple workflows
- ✅ Workflows are independent (different orgs, different targets)

---

### 3. **Installation ID Management**

**Current Approach: Auto-Discovery**
```go
// GetRestClientForOrg auto-discovers installation ID
client, err := GetRestClientForOrg("mongodb")
```

**This works great!** No config changes needed.

**Optional Enhancement: Explicit Installation IDs**
```yaml
# For cases where auto-discovery doesn't work
workflows:
  - name: "mflix-java"
    source:
      repo: "mongodb/docs-sample-apps"
      installation_id: "12345678"  # Optional: override auto-discovery
    destination:
      repo: "mongodb/sample-app-java"
      installation_id: "12345678"  # Optional: override auto-discovery
```

**Recommendation:** Keep auto-discovery, add optional override for edge cases.

---

### 4. **Config Location: Where to Store Config File**

**Current:** Config stored in source repo (`mongodb/docs-sample-apps`)

**Problem with Multi-Org:**
- If you have 3 source repos in 3 orgs, where does config live?
- Each source repo would need its own config (duplication)
- Changing config requires PRs to multiple repos

**Solution A: Centralized Config Repo (Recommended)**
```yaml
# Store config in a dedicated repo: mongodb/code-copier-config
# App reads config from this repo on startup and caches it

workflows:
  # Workflows from mongodb org
  - name: "mflix-java"
    source: { repo: "mongodb/docs-sample-apps" }
    destination: { repo: "mongodb/sample-app-java" }
  
  # Workflows from 10gen org
  - name: "university-python"
    source: { repo: "10gen/university-content" }
    destination: { repo: "mongodb-university/python-course" }
  
  # Workflows from mongodb-university org
  - name: "course-materials"
    source: { repo: "mongodb-university/course-source" }
    destination: { repo: "mongodb-university/course-public" }
```

**Benefits:**
- ✅ Single source of truth
- ✅ One PR to update all workflows
- ✅ Easy to see all copy operations
- ✅ Can use different org for config repo

**Solution B: Per-Source Config (Alternative)**
```yaml
# Each source repo has its own config file
# mongodb/docs-sample-apps/copier-config.yaml
workflows:
  - name: "mflix-java"
    # source is implicit (this repo)
    destination: { repo: "mongodb/sample-app-java" }
```

**Tradeoffs:**
- ✅ Config lives with source code
- ❌ Harder to see all workflows
- ❌ Duplication if multiple sources target same destination

**Recommendation:** Use **Solution A (Centralized Config)** for multi-org.

---

## Proposed Config Format (Multi-Org + Workflows)

### Full Example

```yaml
# Stored in: mongodb/code-copier-config (or any repo you choose)

# Global defaults (optional)
defaults:
  commit_strategy:
    type: "pull_request"
    auto_merge: false
  deprecation_check:
    enabled: true

# Workflows: each defines source → destination mapping
workflows:
  # ============================================================================
  # MongoDB Org: docs-sample-apps → sample app repos
  # ============================================================================
  
  - name: "mflix-java"
    source:
      repo: "mongodb/docs-sample-apps"
      branch: "main"
    destination:
      repo: "mongodb/sample-app-java-mflix"
      branch: "main"
    transformations:
      - move: { from: "mflix/client", to: "client" }
      - move: { from: "mflix/server/java-spring", to: "server" }
      - copy: { from: "mflix/README-JAVA-SPRING.md", to: "README.md" }
  
  - name: "mflix-nodejs"
    source:
      repo: "mongodb/docs-sample-apps"
      branch: "main"
    destination:
      repo: "mongodb/sample-app-nodejs-mflix"
      branch: "main"
    transformations:
      - move: { from: "mflix/client", to: "client" }
      - move: { from: "mflix/server/js-express", to: "server" }
      - copy: { from: "mflix/README-JAVASCRIPT-EXPRESS.md", to: "README.md" }
  
  # ============================================================================
  # 10gen Org: internal content → public repos
  # ============================================================================
  
  - name: "university-python-course"
    source:
      repo: "10gen/university-content"
      branch: "main"
    destination:
      repo: "mongodb-university/python-course"
      branch: "main"
    transformations:
      - move: { from: "courses/python/public", to: "content" }
      - copy: { from: "courses/python/README.md", to: "README.md" }
    exclude:
      - "**/internal/**"
      - "**/.env"
  
  - name: "university-java-course"
    source:
      repo: "10gen/university-content"
      branch: "main"
    destination:
      repo: "mongodb-university/java-course"
      branch: "main"
    transformations:
      - move: { from: "courses/java/public", to: "content" }
  
  # ============================================================================
  # MongoDB University Org: course source → course public
  # ============================================================================
  
  - name: "course-materials-sync"
    source:
      repo: "mongodb-university/course-source"
      branch: "main"
    destination:
      repo: "mongodb-university/course-public"
      branch: "main"
    transformations:
      - move: { from: "materials", to: "public" }
    commit_strategy:
      type: "direct"  # Override default
```

### Key Features

1. **Per-Workflow Source** - Each workflow specifies its source repo
2. **Multi-Org Support** - Workflows can span multiple orgs
3. **Flexible Routing** - Webhook from any source repo triggers its workflows
4. **Centralized Config** - Single file defines all copy operations
5. **Auto-Discovery** - Installation IDs discovered automatically per org

---

## Implementation Changes

### 1. Update Config Types

```go
// types/config.go

type YAMLConfig struct {
    // Remove global source (breaking change, but worth it)
    // SourceRepo   string     `yaml:"source_repo,omitempty"`
    // SourceBranch string     `yaml:"source_branch,omitempty"`
    
    // Legacy format (backward compatible during transition)
    CopyRules []CopyRule `yaml:"copy_rules,omitempty"`
    
    // New workflow format
    Workflows []Workflow `yaml:"workflows,omitempty"`
    Defaults  *Defaults  `yaml:"defaults,omitempty"`
}

type Workflow struct {
    Name            string           `yaml:"name"`
    Source          Source           `yaml:"source"`           // NEW!
    Destination     Destination      `yaml:"destination"`
    Transformations []Transformation `yaml:"transformations"`
    Exclude         []string         `yaml:"exclude,omitempty"`
    CommitStrategy  *CommitStrategyConfig `yaml:"commit_strategy,omitempty"`
}

type Source struct {
    Repo           string `yaml:"repo"`
    Branch         string `yaml:"branch,omitempty"`         // defaults to "main"
    InstallationID string `yaml:"installation_id,omitempty"` // optional override
}

type Destination struct {
    Repo           string `yaml:"repo"`
    Branch         string `yaml:"branch,omitempty"`         // defaults to "main"
    InstallationID string `yaml:"installation_id,omitempty"` // optional override
}
```

### 2. Update Webhook Handler

```go
// services/webhook_handler_new.go

func HandleWebhook(w http.ResponseWriter, r *http.Request) {
    // ... parse webhook ...
    
    webhookRepo := fmt.Sprintf("%s/%s", repoOwner, repoName)
    
    // Find workflows that match this source repo
    matchingWorkflows := []types.Workflow{}
    for _, workflow := range yamlConfig.Workflows {
        if workflow.Source.Repo == webhookRepo {
            matchingWorkflows = append(matchingWorkflows, workflow)
        }
    }
    
    if len(matchingWorkflows) == 0 {
        LogWarning("No workflows configured for source: " + webhookRepo)
        w.WriteHeader(http.StatusNoContent)
        return
    }
    
    LogInfo(fmt.Sprintf("Found %d workflows for source: %s", 
        len(matchingWorkflows), webhookRepo))
    
    // Process each workflow
    for _, workflow := range matchingWorkflows {
        processWorkflow(ctx, workflow, changedFiles, container)
    }
}
```

### 3. Update Workflow Processor

```go
// services/workflow_processor.go

func (wp *WorkflowProcessor) ProcessWorkflow(
    ctx context.Context,
    workflow types.Workflow,
    changedFiles []types.ChangedFile,
) error {
    
    // Get source org from workflow
    sourceOrg, _ := parseRepoPath(workflow.Source.Repo)
    
    // Get destination org from workflow
    destOrg, _ := parseRepoPath(workflow.Destination.Repo)
    
    // Authenticate to source org (for reading files)
    sourceClient, err := GetRestClientForOrg(sourceOrg)
    if err != nil {
        return fmt.Errorf("failed to auth to source org %s: %w", sourceOrg, err)
    }
    
    // Authenticate to destination org (for writing files)
    destClient, err := GetRestClientForOrg(destOrg)
    if err != nil {
        return fmt.Errorf("failed to auth to dest org %s: %w", destOrg, err)
    }
    
    // Process transformations...
    for _, transformation := range workflow.Transformations {
        // ... match files and transform paths ...
    }
    
    return nil
}
```

---

## Migration Path

### Phase 1: Add Workflow Support (Backward Compatible)

```yaml
# Support BOTH formats during transition

# Old format (still works)
source_repo: "mongodb/docs-sample-apps"
copy_rules: [...]

# New format (preferred)
workflows: [...]
```

### Phase 2: Add Multi-Org Support

```yaml
# Workflows can now have different sources
workflows:
  - name: "workflow-1"
    source: { repo: "mongodb/repo1" }
    destination: { repo: "mongodb/target1" }
  
  - name: "workflow-2"
    source: { repo: "10gen/repo2" }  # Different org!
    destination: { repo: "mongodb-university/target2" }
```

### Phase 3: Move Config to Centralized Repo

```bash
# Create config repo
gh repo create mongodb/code-copier-config --private

# Move config file
mv copier-config.yaml → mongodb/code-copier-config/config.yaml

# Update app to read from config repo
CONFIG_REPO=mongodb/code-copier-config
```

### Phase 4: Deprecate Old Format (6 months later)

Remove support for `source_repo` and `copy_rules`.

---

## Benefits Summary

| Feature | Current | With Workflows | With Multi-Org |
|---------|---------|----------------|----------------|
| **Config size** | 300+ lines | 90 lines | 90 lines |
| **Source repos** | 1 | 1 | Unlimited ✅ |
| **Target orgs** | Multiple | Multiple | Multiple ✅ |
| **Config location** | Source repo | Source repo | Centralized ✅ |
| **Webhook routing** | Single source | Single source | Multiple sources ✅ |
| **Installation IDs** | Manual | Manual | Auto-discovery ✅ |

---

## Recommendation

**✅ Implement workflows with per-workflow source from the start.**

**Why:**
1. **Avoid rework** - Don't implement global source, then change to per-workflow
2. **Future-proof** - Multi-org support built-in from day 1
3. **Better design** - Matches Copybara's approach
4. **Same effort** - No extra work to add `source:` field now

**Timeline:**
- Week 1: Implement workflow types with `source` field
- Week 2: Update webhook handler for multi-source routing
- Week 3: Test with multiple orgs
- Week 4: Move config to centralized repo (optional)

---

## Next Steps

1. **Approve this design** - Does per-workflow source work for you?
2. **Choose config location** - Centralized repo or per-source?
3. **Start implementation** - I can build this in 2-3 weeks
4. **Install app in multiple orgs** - mongodb, 10gen, mongodb-university
5. **Test and deploy** - Verify multi-org workflows work

**Want me to start implementing this?** 🚀

