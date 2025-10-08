# Examples Copier Architecture

This document describes the architecture and design of the examples-copier application, including its core components, pattern matching system, configuration management, and operational features.

## Features

### Enhanced Pattern Matching

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

### Path Transformations

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

### YAML Configuration Support

**Files Created:**
- `types/config.go` - New configuration types
- `services/config_loader.go` - Configuration loader with YAML/JSON support
- `configs/config.example.yaml` - Example YAML configuration

**Capabilities:**
- Native YAML support with `gopkg.in/yaml.v3`
- Backward compatible JSON support
- Automatic legacy config conversion
- Comprehensive validation
- Default value handling

**Example:**
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
```

### Template Engine for Messages

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

### MongoDB Audit Logging

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

### Health Check and Metrics Endpoints

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
  "uptime": "2h15m30s"
}
```

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

### CLI Validation Tool

**Files Created:**
- `cmd/config-validator/main.go` - CLI tool for configuration management

**Commands:**

```bash
# Validate configuration
config-validator validate -config copier-copier-config.yaml -v

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

### 8. Development/Testing Features

**Features:**
- **Dry Run Mode**: `DRY_RUN="true"` - No actual changes made
- **Non-main Branch Support**: Configure any target branch
- **Enhanced Logging**: Structured logging with context
- **Metrics Collection**: Optional metrics tracking

## Usage Examples

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
```

### Advanced Regex Config
```yaml
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
          pr_body: "Updated ${file_count} ${lang} files"
```

## Benefits

1. **More Flexible**: Regex patterns with variable extraction
2. **Better DX**: YAML configs are more readable and maintainable
3. **Observable**: Health checks, metrics, and audit logging
4. **Testable**: CLI tools for validation and testing
5. **Production Ready**: Dry-run mode, proper error handling, monitoring

## Breaking Changes

None - the refactoring maintains backward compatibility with existing JSON configs through automatic conversion.

