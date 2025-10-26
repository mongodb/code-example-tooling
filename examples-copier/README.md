# GitHub Docs Code Example Copier

A GitHub app that automatically copies code examples and files from a source repository to one or more target repositories when pull requests are merged. Features advanced pattern matching, path transformations, audit logging, and comprehensive monitoring.

## Features

### Core Functionality
- **Automated File Copying** - Copies files from source to target repos on PR merge
- **Advanced Pattern Matching** - Prefix, glob, and regex patterns with variable extraction
- **Path Transformations** - Template-based path transformations with variable substitution
- **Multiple Targets** - Copy files to multiple repositories and branches
- **Flexible Commit Strategies** - Direct commits or pull requests with auto-merge
- **Deprecation Tracking** - Automatic tracking of deleted files

### Enhanced Features
- **YAML Configuration** - Modern YAML config with JSON backward compatibility
- **Message Templating** - Template-ized commit messages and PR titles
- **Audit Logging** - MongoDB-based event tracking for all operations
- **Health & Metrics** - `/health` and `/metrics` endpoints for monitoring
- **Development Tools** - Dry-run mode, CLI validation, enhanced logging
- **Thread-Safe** - Concurrent webhook processing with proper state management

## 🚀 Quick Start

### Prerequisites

- Go 1.23.4+
- GitHub App credentials
- Google Cloud project (for Secret Manager and logging)
- MongoDB Atlas (optional, for audit logging)

### Installation

```bash
# Clone the repository
git clone https://github.com/your-org/code-example-tooling.git
cd code-example-tooling/examples-copier

# Install dependencies
go mod download

# Build the application
go build -o examples-copier .

# Build CLI tools
go build -o config-validator ./cmd/config-validator
```

### Configuration

1. **Copy environment example file**

```bash
# For local development
cp configs/.env.local.example configs/.env

# Or for YAML-based configuration
cp configs/env.yaml.example env.yaml
```

2. **Set required environment variables**

```bash
# GitHub Configuration
REPO_OWNER=your-org
REPO_NAME=your-repo
SRC_BRANCH=main
GITHUB_APP_ID=123456
GITHUB_INSTALLATION_ID=789012

# Google Cloud
GCP_PROJECT_ID=your-project
PEM_KEY_NAME=projects/123/secrets/<name>/versions/latest
WEBHOOK_SECRET_NAME=projects/123/secrets/webhook-secret

# Application Settings
PORT=8080
CONFIG_FILE=copier-config.yaml
DEPRECATION_FILE=deprecated_examples.json

# Optional: MongoDB Audit Logging
AUDIT_ENABLED=true
MONGO_URI=mongodb+srv://user:pass@cluster.mongodb.net
AUDIT_DATABASE=code_copier
AUDIT_COLLECTION=audit_events

# Optional: Development Features
DRY_RUN=false
METRICS_ENABLED=true
```

3. **Create configuration file**

Create `copier-config.yaml` in your source repository:

```yaml
source_repo: "your-org/source-repo"
source_branch: "main"

copy_rules:
  - name: "Copy Go examples"
    source_pattern:
      type: "regex"
      pattern: "^examples/(?P<lang>[^/]+)/(?P<category>[^/]+)/(?P<file>.+)$"
    targets:
      - repo: "your-org/target-repo"
        branch: "main"
        path_transform: "docs/examples/${lang}/${category}/${file}"
        commit_strategy:
          type: "pull_request"
          commit_message: "Update ${category} examples from ${lang}"
          pr_title: "Update ${category} examples"
          auto_merge: false
        deprecation_check:
          enabled: true
          file: "deprecated_examples.json"
```

### Running the Application

```bash
# Run with default settings
./examples-copier

# Run with custom environment file
./examples-copier -env ./configs/.env.production

# Run in dry-run mode (no actual commits)
./examples-copier -dry-run

# Validate configuration only
./examples-copier -validate
```

## Configuration

### Pattern Types

#### Prefix Pattern
Simple string prefix matching:

```yaml
source_pattern:
  type: "prefix"
  pattern: "examples/go/"
```

Matches: `examples/go/main.go`, `examples/go/database/connect.go`

#### Glob Pattern
Wildcard matching with `*` and `?`:

```yaml
source_pattern:
  type: "glob"
  pattern: "examples/*/main.go"
```

Matches: `examples/go/main.go`, `examples/python/main.go`

#### Regex Pattern
Full regex with named capture groups:

```yaml
source_pattern:
  type: "regex"
  pattern: "^examples/(?P<lang>[^/]+)/(?P<file>.+)$"
```

Matches: `examples/go/main.go` (extracts `lang=go`, `file=main.go`)

### Path Transformations

Transform source paths to target paths using variables:

```yaml
path_transform: "docs/${lang}/${category}/${file}"
```

**Built-in Variables:**
- `${path}` - Full source path
- `${filename}` - File name only
- `${dir}` - Directory path
- `${ext}` - File extension

**Custom Variables:**
- Any named groups from regex patterns
- Example: `(?P<lang>[^/]+)` creates `${lang}`

### Commit Strategies

#### Direct Commit
```yaml
commit_strategy:
  type: "direct"
  commit_message: "Update examples from ${source_repo}"
```

#### Pull Request
```yaml
commit_strategy:
  type: "pull_request"
  commit_message: "Update examples"
  pr_title: "Update ${category} examples"
  pr_body: "Automated update from ${source_repo}"
  auto_merge: true
```

### Message Templates

Use variables in commit messages and PR titles:

```yaml
commit_message: "Update ${category} examples from ${lang}"
pr_title: "Update ${category} examples"
```

**Available Variables:**
- `${rule_name}` - Name of the copy rule
- `${source_repo}` - Source repository
- `${target_repo}` - Target repository
- `${source_branch}` - Source branch
- `${target_branch}` - Target branch
- `${file_count}` - Number of files being copied
- Any custom variables from pattern matching

## CLI Tools

### Config Validator

Validate and test configurations before deployment:

```bash
# Validate config file
./config-validator validate -config copier-config.yaml -v

# Test pattern matching
./config-validator test-pattern \
  -type regex \
  -pattern "^examples/(?P<lang>[^/]+)/(?P<file>.+)$" \
  -file "examples/go/main.go"

# Test path transformation
./config-validator test-transform \
  -template "docs/${lang}/${file}" \
  -file "examples/go/main.go" \
  -pattern "^examples/(?P<lang>[^/]+)/(?P<file>.+)$"

# Initialize new config from template
./config-validator init -output copier-config.yaml

# Convert between formats
./config-validator convert -input config.json -output copier-config.yaml
```

## Monitoring

### Health Endpoint

Check application health:

```bash
curl http://localhost:8080/health
```

Response:
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
  "uptime": "1h23m45s"
}
```

### Metrics Endpoint

Get performance metrics:

```bash
curl http://localhost:8080/metrics
```

Response:
```json
{
  "webhooks": {
    "received": 42,
    "processed": 40,
    "failed": 2,
    "success_rate": 95.24,
    "processing_time": {
      "avg_ms": 234.5,
      "p50_ms": 200,
      "p95_ms": 450,
      "p99_ms": 890
    }
  },
  "files": {
    "matched": 150,
    "uploaded": 145,
    "upload_failed": 5,
    "deprecated": 3,
    "upload_success_rate": 96.67
  }
}
```

## Audit Logging

When enabled, all operations are logged to MongoDB:

```javascript
// Query recent copy events
db.audit_events.find({
  event_type: "copy",
  success: true
}).sort({timestamp: -1}).limit(10)

// Find failed operations
db.audit_events.find({
  success: false
}).sort({timestamp: -1})

// Statistics by rule
db.audit_events.aggregate([
  {$match: {event_type: "copy"}},
  {$group: {
    _id: "$rule_name",
    count: {$sum: 1},
    avg_duration: {$avg: "$duration_ms"}
  }}
])
```

## Testing

### Run Unit Tests

```bash
# Run all tests
go test ./services -v

# Run specific test suite
go test ./services -v -run TestPatternMatcher

# Run with coverage
go test ./services -cover
go test ./services -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## Development

### Dry-Run Mode

Test without making actual changes:

```bash
DRY_RUN=true ./examples-copier
```

In dry-run mode:
- Webhooks are processed
- Files are matched and transformed
- Audit events are logged
- **NO actual commits or PRs are created**

### Enhanced Logging

Enable detailed logging:

```bash
LOG_LEVEL=debug ./examples-copier
# or
COPIER_DEBUG=true ./examples-copier
```

## Architecture

### Project Structure

```
examples-copier/
├── app.go                    # Main application entry point
├── cmd/
│   ├── config-validator/     # CLI validation tool
│   └── test-webhook/         # Webhook testing tool
├── configs/
│   ├── environment.go        # Environment configuration
│   ├── .env.local.example    # Local environment template
│   ├── env.yaml.example      # YAML environment template
│   └── copier-config.example.yaml # Config template
├── services/
│   ├── pattern_matcher.go    # Pattern matching engine
│   ├── config_loader.go      # Config loading & validation
│   ├── audit_logger.go       # MongoDB audit logging
│   ├── health_metrics.go     # Health & metrics endpoints
│   ├── file_state_service.go # Thread-safe state management
│   ├── service_container.go  # Dependency injection
│   ├── webhook_handler_new.go # Webhook handler
│   ├── github_auth.go        # GitHub authentication
│   ├── github_read.go        # GitHub read operations
│   ├── github_write_to_target.go # GitHub write operations
│   └── slack_notifier.go     # Slack notifications
├── types/
│   ├── config.go             # Configuration types
│   └── types.go              # Core types
└── docs/
    ├── ARCHITECTURE.md       # Architecture overview
    ├── CONFIGURATION-GUIDE.md # Complete config reference
    ├── DEPLOYMENT.md         # Deployment guide
    ├── FAQ.md                # Frequently asked questions
    └── ...                   # Additional documentation
```

### Service Container

The application uses dependency injection for clean architecture:

```go
container := NewServiceContainer(config)
// All services initialized and wired together
```

## Deployment

See [DEPLOYMENT.md](./docs/DEPLOYMENT.md) for complete deployment guide and [DEPLOYMENT-CHECKLIST.md](./docs/DEPLOYMENT-CHECKLIST.md) for step-by-step checklist.

### Google Cloud App Engine

```bash
gcloud app deploy
```

### Docker

```bash
docker build -t examples-copier .
docker run -p 8080:8080 --env-file .env examples-copier
```

## Security

- **Webhook Signature Verification** - HMAC-SHA256 validation
- **Secret Management** - Google Cloud Secret Manager
- **Least Privilege** - Minimal GitHub App permissions
- **Audit Trail** - Complete operation logging

## Documentation

### Getting Started

- **[Configuration Guide](docs/CONFIGURATION-GUIDE.md)** - Complete configuration reference
- **[Pattern Matching Guide](docs/PATTERN-MATCHING-GUIDE.md)** - Pattern matching with examples
- **[Local Testing](docs/LOCAL-TESTING.md)** - Test locally before deploying
- **[Deployment Guide](docs/DEPLOYMENT.md)** - Deploy to production

### Reference

- **[Architecture](docs/ARCHITECTURE.md)** - System design and components
- **[Troubleshooting](docs/TROUBLESHOOTING.md)** - Common issues and solutions
- **[FAQ](docs/FAQ.md)** - Frequently asked questions
- **[Debug Logging](docs/DEBUG-LOGGING.md)** - Debug logging configuration
- **[Deprecation Tracking](docs/DEPRECATION-TRACKING-EXPLAINED.md)** - How deprecation tracking works

### Features

- **[Slack Notifications](docs/SLACK-NOTIFICATIONS.md)** - Slack integration guide
- **[Webhook Testing](docs/WEBHOOK-TESTING.md)** - Test with real PR data

### Tools

- **[config-validator](cmd/config-validator/README.md)** - Validate and test configurations
- **[test-webhook](cmd/test-webhook/README.md)** - Test webhook processing
- **[Scripts](scripts/README.md)** - Helper scripts
