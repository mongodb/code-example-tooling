# Examples Copier Architecture

This document describes the architecture and design of the examples-copier application, including its core components, pattern matching system, configuration management, deprecation tracking, and operational features.

## Core Architecture

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

## Features

### 1. Enhanced Pattern Matching

**Files Created:**
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

### 2. Path Transformations

**Files Created:**
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

### 3. Deprecation Tracking

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

### 4. YAML Configuration Support

**Files Created:**
- `types/config.go` - New configuration types
- `services/config_loader.go` - Configuration loader with YAML/JSON support
- `configs/copier-config.example.yaml` - Example YAML configuration

**Capabilities:**
- Native YAML support with `gopkg.in/yaml.v3`
- Backward compatible JSON support
- Automatic legacy config conversion
- Comprehensive validation
- Default value handling

**Configuration Structure:**
```yaml
source_repo: "mongodb/docs-code-examples"
source_branch: "main"

copy_rules:
  - name: "go-examples"
    source_pattern:
      type: "glob"
      pattern: "examples/**/*.go"
    targets:
      - repo: "mongodb/docs"
        branch: "main"
        path_transform: "code/${filename}"
        commit_strategy:
          type: "pull_request"  # or "direct"
          pr_title: "Update Go examples"
          pr_body: "Automated update"
          auto_merge: false
        deprecation_check:
          enabled: true
          file: "deprecated_examples.json"
```

### 5. Template Engine for Messages

**Files Created:**
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

### 6. MongoDB Audit Logging

**Files Created:**
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

### 7. Health Check and Metrics Endpoints

**Files Created:**
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

### 8. CLI Validation Tool

**Files Created:**
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
config-validator init -template basic -output my-copier-config.yaml

# Convert between formats
config-validator convert -input config.json -output copier-config.yaml
```

### 9. Development/Testing Features

**Features:**
- **Dry Run Mode**: `DRY_RUN="true"` - No actual changes made
- **Non-main Branch Support**: Configure any target branch
- **Enhanced Logging**: Structured logging with context (JSON format)
- **Metrics Collection**: Optional metrics tracking
- **Context-aware Operations**: All operations support context cancellation

**Logging Features:**
- Structured JSON logs with contextual information
- Operation tracking with elapsed time
- File status logging (ADDED, MODIFIED, DELETED)
- Deprecation event logging
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

### Basic YAML Config
```yaml
source_repo: "mongodb/docs-code-examples"
source_branch: "main"

copy_rules:
  - name: "go-examples"
    source_pattern:
      type: "prefix"
      pattern: "examples/go"
    targets:
      - repo: "mongodb/docs"
        branch: "main"
        path_transform: "code/go/${relative_path}"
        commit_strategy:
          type: "direct"
          commit_message: "Update Go examples from ${source_repo}"
```

### Advanced Regex Config with Deprecation
```yaml
source_repo: "mongodb/docs-code-examples"
source_branch: "main"

copy_rules:
  - name: "language-examples"
    source_pattern:
      type: "regex"
      pattern: "^examples/(?P<lang>[^/]+)/(?P<file>.+)$"
    targets:
      - repo: "mongodb/docs"
        branch: "main"
        path_transform: "code/${lang}/${file}"
        commit_strategy:
          type: "pull_request"
          pr_title: "Update ${lang} examples"
          pr_body: "Updated ${file_count} ${lang} files from ${source_repo}"
          auto_merge: false
        deprecation_check:
          enabled: true
          file: "deprecated_examples.json"
```

### Multi-Target Config
```yaml
source_repo: "mongodb/aggregation-examples"
source_branch: "main"

copy_rules:
  # Java examples
  - name: "java-examples"
    source_pattern:
      type: "regex"
      pattern: "^java/(?P<file>.+\\.java)$"
    targets:
      - repo: "mongodb/docs"
        branch: "main"
        path_transform: "java/${file}"
        commit_strategy:
          type: "pull_request"
          pr_title: "Update Java examples"
          auto_merge: false
        deprecation_check:
          enabled: true
          file: "deprecated_examples.json"

  # Node.js examples
  - name: "nodejs-examples"
    source_pattern:
      type: "regex"
      pattern: "^nodejs/(?P<file>.+\\.(js|ts))$"
    targets:
      - repo: "mongodb/docs"
        branch: "main"
        path_transform: "node/${file}"
        commit_strategy:
          type: "pull_request"
          pr_title: "Update Node.js examples"
          auto_merge: true
        deprecation_check:
          enabled: true
          file: "deprecated_examples.json"

  # Python examples
  - name: "python-examples"
    source_pattern:
      type: "regex"
      pattern: "^python/(?P<file>.+\\.py)$"
    targets:
      - repo: "mongodb/docs"
        branch: "main"
        path_transform: "python/${file}"
        commit_strategy:
          type: "direct"
          commit_message: "Update Python examples"
        deprecation_check:
          enabled: true
          file: "deprecated_examples.json"
```

## Key Benefits

1. **Flexible Pattern Matching**: Regex patterns with variable extraction and multiple pattern types
2. **Better Developer Experience**: YAML configs are more readable and maintainable
3. **Observable**: Health checks, metrics, and comprehensive audit logging
4. **Testable**: CLI tools for validation and testing, dry-run mode
5. **Production Ready**: Thread-safe operations, proper error handling, monitoring
6. **Deprecation Tracking**: Automatic detection and tracking of deleted files
7. **Batch Operations**: Efficient batching of multiple files per target
8. **Template Engine**: Dynamic message generation with variables

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

**Platform**: Google Cloud App Engine (Flexible Environment)

**Environment Variables:**
```bash
# Required
REPO_OWNER="mongodb"
REPO_NAME="docs-code-examples"
SRC_BRANCH="main"
GITHUB_TOKEN="ghp_..."
WEBHOOK_SECRET="..."

# Optional
AUDIT_ENABLED="true"
MONGO_URI="mongodb+srv://..."
DRY_RUN="false"
CONFIG_FILE="copier-config.yaml"
```

**Health Monitoring:**
- `/health` endpoint for liveness checks
- `/metrics` endpoint for monitoring
- Structured JSON logs for analysis

## Breaking Changes

None - the refactoring maintains backward compatibility with existing JSON configs through automatic conversion.

## Future Enhancements

Potential improvements documented in codebase:

1. **Automatic Cleanup PRs** - Create PRs to remove deprecated files from targets
2. **Expiration Dates** - Auto-remove deprecation entries after X days
3. **Cleanup Verification** - Check if deprecated files still exist in targets
4. **Batch Cleanup Tool** - CLI tool to clean up all deprecated files
5. **Notifications** - Alert when deprecation file grows large
6. **Retry Logic** - Automatic retry for failed GitHub API calls
7. **Rate Limiting** - Respect GitHub API rate limits
8. **Webhook Queue** - Queue webhooks for processing during high load

