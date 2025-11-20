# Examples Copier Architecture

This document describes the architecture and design of the examples-copier application, including its core components, main config system, pattern matching, configuration management, deprecation tracking, and operational features.

## Core Architecture

### Main Config System

The application uses a **centralized main config** with **distributed workflow configs**:

**Files:**
- `services/main_config_loader.go` - Main config loading and reference resolution
- `types/config.go` - Configuration types including MainConfig and WorkflowConfigRef

**Key Features:**
- **Centralized Defaults** - Global defaults in main config file
- **Distributed Workflows** - Workflow configs in source repositories
- **Three Reference Types**:
  - `inline` - Workflows embedded directly in main config
  - `local` - Workflow configs in same repo as main config
  - `repo` - Workflow configs in source repositories
- **Source Context Inference** - Workflows automatically inherit source.repo and source.branch from workflow config reference
- **$ref Support** - Reference external files for transformations, commit_strategy, and exclude patterns
- **Resilient Loading** - Continues processing when individual workflow configs fail to load (logs warnings instead of failing)

**Configuration Structure:**
```yaml
# Main config (.copier/workflows/main.yaml in config repo)
defaults:
  commit_strategy:
    type: "pull_request"
    auto_merge: false

workflow_configs:
  - source: "repo"
    repo: "mongodb/docs-sample-apps"
    branch: "main"
    path: ".copier/workflows/config.yaml"
    enabled: true
```

**Benefits:**
- Separation of concerns - each repo manages its own workflows
- Scalability - works for monorepos with many workflows
- Flexibility - mix centralized and distributed configs
- Discoverability - configs live near source code
- Maintainability - update workflows without touching main config

### Service Container Pattern

The application uses a **Service Container** to manage dependencies and provide thread-safe access to shared services:

**Files:**
- `services/webhook_handler_new.go` - ServiceContainer struct and initialization

**Components:**
- `FileStateService` - Thread-safe state management for files to upload/deprecate
- `PatternMatcher` - Pattern matching engine
- `MessageTemplater` - Template rendering for messages
- `AuditLogger` - MongoDB audit logging
- `MetricsCollector` - Metrics tracking

**Benefits:**
- Dependency injection for testability
- Thread-safe operations with mutex locks
- Clean separation of concerns
- Easy to mock for testing

### File State Management

**Files:**
- `services/file_state_service.go` - FileStateService interface and implementation

**Capabilities:**
- Thread-safe file queuing with `sync.RWMutex`
- Separate queues for uploads and deprecations
- Composite keys to prevent collisions
- Copy-on-read to prevent external modification

**Upload Key Structure:**
```go
type UploadKey struct {
    RepoName       string  // Target repository
    BranchPath     string  // Target branch
    RuleName       string  // Rule name (allows multiple rules per repo)
    CommitStrategy string  // "direct" or "pull_request"
}
```

**Deprecation Key Structure:**
- Composite key: `{repo}:{targetPath}` (e.g., `mongodb/docs:code/example.go`)
- Ensures uniqueness when multiple files are deprecated to the same deprecation file
- Prevents map key collisions

## Key Features

### 1. Main Config with Workflow References

**Files:**
- `services/main_config_loader.go` - Main config loading and workflow reference resolution
- `types/config.go` - MainConfig and WorkflowConfigRef types

**Capabilities:**
- **Three-tier configuration**: Main config → Workflow configs → Individual workflows
- **Default precedence**: Workflow > Workflow config > Main config > System defaults
- **Workflow config references**: Local, remote (repo), or inline workflows
- **Source context inference**: Workflows inherit source.repo/branch from workflow config reference
- **Resilient loading**: Logs warnings for missing configs and continues processing
- **Validation**: Comprehensive validation at each level

**Example:**
```yaml
# Main config
defaults:
  commit_strategy:
    type: "pull_request"
    auto_merge: false

workflow_configs:
  - source: "repo"
    repo: "mongodb/docs-sample-apps"
    path: ".copier/workflows/config.yaml"
```

### 2. $ref Support for Reusable Components

**Files:**
- `services/main_config_loader.go` - Reference resolution logic
- `types/config.go` - RefOrValue types for $ref support

**Capabilities:**
- **Transformations references**: Extract common file mappings
- **Strategy references**: Reuse PR strategies across workflows
- **Exclude references**: Share exclude patterns
- **Relative paths**: Resolved relative to workflow config file
- **Repo references**: `repo://owner/repo/path/file.yaml@branch` format

**Example:**
```yaml
workflows:
  - name: "mflix-java"
    transformations:
      $ref: "../transformations/mflix-java.yaml"
    commit_strategy:
      $ref: "../strategies/mflix-pr-strategy.yaml"
    exclude:
      $ref: "../common/mflix-excludes.yaml"
```

### 3. Enhanced Pattern Matching

**Files:**
- `services/pattern_matcher.go` - Pattern matching engine

**Capabilities:**
- **Prefix Matching**: Simple prefix-based file matching
- **Glob Matching**: Supports `*`, `**`, and `?` wildcards
- **Regex Matching**: Full regex support with named capture groups for variable extraction

**Example:**
```yaml
source_pattern:
  type: "regex"
  pattern: "^examples/(?P<lang>[^/]+)/(?P<category>[^/]+)/(?P<file>.+)$"
```

### 4. Path Transformations

**Files:**
- `services/pattern_matcher.go` (PathTransformer interface)

**Capabilities:**
- Variable substitution with `${variable}` syntax
- Built-in variables: `${path}`, `${filename}`, `${dir}`, `${ext}`
- Custom variables from regex capture groups
- Template validation with error reporting

**Example:**
```yaml
path_transform: "source/code-examples/${lang}/${category}/${file}"
```

### 5. Deprecation Tracking

**Files:**
- `services/webhook_handler_new.go` - Deprecation detection and queuing
- `services/github_write_to_source.go` - Deprecation file updates
- `services/file_state_service.go` - Deprecation queue management

**How It Works:**

1. **Detection**: When a PR is merged, files with status `DELETED` are identified
2. **Pattern Matching**: Deleted files are matched against copy rules
3. **Path Calculation**: Target repository paths are calculated using path transforms
4. **Queuing**: Files are added to deprecation queue with composite key `{repo}:{targetPath}`
5. **File Update**: Deprecation file in source repository is updated with all entries

**Key Implementation Details:**

**Composite Key Fix (Critical):**
```go
// Use composite key to prevent collisions when multiple files
// are deprecated to the same deprecation file
key := target.Repo + ":" + targetPath
fileStateService.AddFileToDeprecate(key, entry)
```

**Why Composite Keys?**
- Multiple rules can target the same deprecation file
- Without composite keys, entries would overwrite each other in the map
- Example: 3 files (Java, Node.js, Python) all using `deprecated_examples.json`
- With simple key: Only 1 entry survives (last one wins)
- With composite key: All 3 entries preserved

**Deprecation File Format:**
```json
[
  {
    "filename": "code/example.go",
    "repo": "mongodb/docs",
    "branch": "main",
    "deleted_on": "2025-10-26T18:34:43Z"
  }
]
```

**Configuration:**
```yaml
targets:
  - repo: "mongodb/docs"
    branch: "main"
    deprecation_check:
      enabled: true
      file: "deprecated_examples.json"  # Optional, defaults to deprecated_examples.json
```

**Protection Against Empty Commits:**
- Checks if deprecation queue is empty before updating file
- Returns early if no files to deprecate
- Prevents blank commits to source repository

### 6. YAML Configuration Support

**Files:**
- `types/config.go` - Configuration types with $ref support
- `services/config_loader.go` - Configuration loader
- `services/main_config_loader.go` - Main config loader with reference resolution

**Capabilities:**
- Native YAML support with `gopkg.in/yaml.v3`
- Custom unmarshaling for $ref support
- Comprehensive validation
- Default value handling
- Reference resolution (relative paths and repo:// format)

### 7. Template Engine for Messages

**Files:**
- `services/pattern_matcher.go` (MessageTemplater interface)
- `types/config.go` (MessageContext)

**Capabilities:**
- Template variables in commit messages, PR titles, and PR bodies
- Built-in context variables: `${rule_name}`, `${source_repo}`, `${target_repo}`, `${file_count}`, `${pr_number}`, `${commit_sha}`
- Custom variables from pattern matching
- Fallback to sensible defaults

**Example:**
```yaml
commit_strategy:
  type: "pull_request"
  pr_title: "Update ${lang} examples"
  pr_body: "Automated update of ${lang} examples (${file_count} files)"
```

### 8. MongoDB Audit Logging

**Files:**
- `services/audit_logger.go` - MongoDB audit logger

**Capabilities:**
- Event logging to MongoDB Atlas
- Event types: copy, deprecation, error
- Automatic indexing for performance
- Query methods: recent events, failed events, events by rule
- Statistics: by rule, daily volume
- No-op implementation when disabled

**Configuration:**
```bash
AUDIT_ENABLED="true"
MONGO_URI="mongodb+srv://user:pass@cluster.mongodb.net"
AUDIT_DATABASE="copier_audit"
AUDIT_COLLECTION="events"
```

**Event Structure:**
```json
{
  "timestamp": "2025-10-08T10:30:00Z",
  "event_type": "copy",
  "rule_name": "go-examples",
  "source_repo": "mongodb/docs-code-examples",
  "source_path": "examples/go/main.go",
  "target_repo": "mongodb/docs",
  "target_path": "code/go/main.go",
  "commit_sha": "abc123",
  "pr_number": 42,
  "success": true,
  "duration_ms": 1250,
  "file_size": 2048
}
```

**Integration:**
- Logs copy operations with success/failure status
- Logs deprecation events when files are deleted
- Logs errors with full context
- Thread-safe operation through ServiceContainer

### 9. Health Check and Metrics Endpoints

**Files:**
- `services/health_metrics.go` - Health and metrics implementation

**Endpoints:**

#### GET /health
Returns application health status:
```json
{
  "status": "healthy",
  "started": true,
  "github": {
    "status": "healthy",
    "authenticated": true
  },
  "queues": {
    "upload_count": 0,
    "deprecation_count": 0
  },
  "audit_logger": {
    "status": "healthy",
    "connected": true
  },
  "uptime": "2h15m30s"
}
```

**Health Check Features:**
- GitHub authentication verification
- Queue status (upload and deprecation)
- Audit logger connection status
- Application uptime tracking

#### GET /metrics
Returns detailed metrics:
```json
{
  "webhooks": {
    "received": 150,
    "processed": 145,
    "failed": 5,
    "success_rate": 96.67,
    "processing_time": {
      "avg_ms": 1250.5,
      "min_ms": 450.2,
      "max_ms": 3200.8
    }
  },
  "files": {
    "matched": 320,
    "uploaded": 310,
    "upload_failed": 5,
    "deprecated": 5,
    "upload_success_rate": 98.41
  },
  "github_api": {
    "calls": 1250,
    "errors": 12,
    "error_rate": 0.96
  }
}
```

**Metrics Tracking:**
- Webhook processing statistics
- File operation counters (matched, uploaded, deprecated, failed)
- GitHub API call tracking
- Success rates and error rates

### 10. CLI Validation Tool

**Files:**
- `cmd/config-validator/main.go` - CLI tool for configuration management

**Commands:**

```bash
# Validate configuration
config-validator validate -config copier-config.yaml -v

# Test pattern matching
config-validator test-pattern \
  -type glob \
  -pattern "examples/**/*.go" \
  -file "examples/go/main.go"

# Test path transformation
config-validator test-transform \
  -source "examples/go/main.go" \
  -template "code/${filename}"

# Initialize new config from template
config-validator init -template basic -output my-workflow-config.yaml
```

### 11. Development/Testing Features

**Features:**
- **Dry Run Mode**: `DRY_RUN="true"` - No actual changes made
- **Non-main Branch Support**: Configure any target branch
- **Enhanced Logging**: Structured logging with context
- **Metrics Collection**: Optional metrics tracking
- **Context-aware Operations**: All operations support context cancellation
- **Resilient Config Loading**: Continues processing when individual configs fail

**Logging Features:**
- Structured logs with contextual information
- Operation tracking with elapsed time
- File status logging (ADDED, MODIFIED, DELETED)
- Deprecation event logging
- Warning logs for missing configs (non-fatal)
- Error logging with full context

## Webhook Processing Flow

### High-Level Flow

1. **Webhook Received** → Verify signature and parse payload
2. **PR Validation** → Check if PR is merged
3. **File Retrieval** → Get changed files from GitHub GraphQL API
4. **Pattern Matching** → Match files against copy rules
5. **File Processing** → Handle copies and deprecations
6. **Queue Processing** → Upload files and update deprecation file
7. **Metrics & Audit** → Record metrics and log events

### Detailed Processing Steps

#### 1. File Status Detection
```go
// GitHub GraphQL API returns file status
type ChangedFile struct {
    Path      string
    Status    string  // "ADDED", "MODIFIED", "DELETED", "RENAMED", etc.
    Additions int
    Deletions int
}
```

#### 2. Pattern Matching
- Each file is tested against all copy rules
- First matching rule wins
- Variables extracted from regex capture groups
- Path transformation applied

#### 3. File Routing
```go
if file.Status == "DELETED" {
    // Route to deprecation handler
    handleFileDeprecation(...)
} else {
    // Route to copy handler
    handleFileCopyWithAudit(...)
}
```

#### 4. Queue Management
- Files queued with composite keys to prevent collisions
- Upload queue: `{repo}:{branch}:{rule}:{strategy}`
- Deprecation queue: `{repo}:{targetPath}`
- Thread-safe operations with mutex locks

#### 5. Batch Operations
- All files for same target are batched together
- Single commit per target repository
- Single PR per target (if using PR strategy)
- Deprecation file updated once with all entries

## Configuration Examples

### Main Config with Workflow References
```yaml
# .copier/workflows/main.yaml (in config repo)
defaults:
  commit_strategy:
    type: "pull_request"
    auto_merge: false
  exclude:
    - "**/.env"
    - "**/node_modules/**"

workflow_configs:
  # Workflows in source repo
  - source: "repo"
    repo: "mongodb/docs-sample-apps"
    branch: "main"
    path: ".copier/workflows/config.yaml"
    enabled: true

  # Local workflows in config repo
  - source: "local"
    path: "workflows/internal-workflows.yaml"
    enabled: true

  # Inline workflow for simple cases
  - source: "inline"
    workflows:
      - name: "simple-copy"
        source:
          repo: "mongodb/source-repo"
          branch: "main"
        destination:
          repo: "mongodb/dest-repo"
          branch: "main"
        transformations:
          - move: { from: "src", to: "dest" }
```

### Workflow Config in Source Repo
```yaml
# .copier/workflows/config.yaml (in source repo)
defaults:
  commit_strategy:
    type: "pull_request"
    auto_merge: false
  deprecation_check:
    enabled: true

workflows:
  - name: "mflix-java"
    # source.repo and source.branch inherited from workflow config reference
    destination:
      repo: "mongodb/sample-app-java-mflix"
      branch: "main"
    transformations:
      - move: { from: "mflix/client", to: "client" }
      - move: { from: "mflix/server/java-spring", to: "server" }
    commit_strategy:
      $ref: "../strategies/mflix-pr-strategy.yaml"
```

### Reusable Strategy File
```yaml
# .copier/strategies/mflix-pr-strategy.yaml
type: "pull_request"
pr_title: "🤖 Automated update from source repo"
pr_body: |
  This PR was automatically generated by the code copier app.

  **Files updated:** ${file_count}
  **Source:** ${source_repo}
use_pr_template: true
auto_merge: false
```

## Key Benefits

1. **Centralized Configuration**: Main config with distributed workflow management
2. **Source Context Inference**: Workflows automatically inherit source repo/branch
3. **Reusable Components**: $ref support for transformations, strategies, and excludes
4. **Resilient Loading**: Continues processing when individual configs fail
5. **Flexible Pattern Matching**: Regex patterns with variable extraction
6. **Observable**: Health checks, metrics, and comprehensive audit logging
7. **Testable**: CLI tools for validation and testing, dry-run mode
8. **Production Ready**: Thread-safe operations, proper error handling, monitoring
9. **Deprecation Tracking**: Automatic detection and tracking of deleted files
10. **Template Engine**: Dynamic message generation with variables

## Thread Safety

The application is designed for concurrent operations:

- **FileStateService**: Thread-safe with `sync.RWMutex`
- **MetricsCollector**: Thread-safe counters
- **AuditLogger**: Thread-safe MongoDB operations
- **ServiceContainer**: Immutable after initialization

## Error Handling

- Context-aware cancellation support
- Graceful degradation (audit logging optional)
- Detailed error logging with full context
- Metrics tracking for failed operations
- No-op implementations for optional features

## Performance Considerations

- **Batch Operations**: Multiple files committed in single operation
- **Composite Keys**: Prevent map collisions and overwrites
- **Copy-on-Read**: FileStateService returns copies to prevent external modification
- **GraphQL API**: Efficient file retrieval with single query
- **Mutex Locks**: Read/write locks for optimal concurrency

## Deployment

**Platform**: Google Cloud Run

**Environment Variables:**
```yaml
# GitHub Configuration
GITHUB_APP_ID: "1166559"
INSTALLATION_ID: "62138132"  # Optional fallback

# Config Repository
CONFIG_REPO_OWNER: "mongodb"
CONFIG_REPO_NAME: "code-example-tooling"
CONFIG_REPO_BRANCH: "main"

# Main Config
MAIN_CONFIG_FILE: ".copier/workflows/main.yaml"
USE_MAIN_CONFIG: "true"

# Secret Manager References
GITHUB_APP_PRIVATE_KEY_SECRET_NAME: "projects/.../secrets/CODE_COPIER_PEM/versions/latest"
WEBHOOK_SECRET_NAME: "projects/.../secrets/webhook-secret/versions/latest"
MONGO_URI_SECRET_NAME: "projects/.../secrets/mongo-uri/versions/latest"

# Application Settings
WEBSERVER_PATH: "/events"
DEPRECATION_FILE: "deprecated_examples.json"
COMMITTER_NAME: "GitHub Copier App"
COMMITTER_EMAIL: "bot@mongodb.com"

# Feature Flags
AUDIT_ENABLED: "false"
METRICS_ENABLED: "true"
```

**Health Monitoring:**
- `/health` endpoint for liveness checks
- `/metrics` endpoint for monitoring
- Structured logs for analysis

## Future Enhancements

Potential improvements:

1. **Automatic Cleanup PRs** - Create PRs to remove deprecated files from targets
2. **Expiration Dates** - Auto-remove deprecation entries after X days
3. **Config Validation CLI** - Enhanced validation tool
4. **Retry Logic** - Automatic retry for failed GitHub API calls
5. **Rate Limiting** - Respect GitHub API rate limits

