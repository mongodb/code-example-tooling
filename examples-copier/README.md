# GitHub Docs Code Example Copier

A GitHub app that automatically copies code examples and files from source repositories to target repositories when pull requests are merged. Features centralized configuration with distributed workflow management, $ref support for reusable components, advanced pattern matching, and comprehensive monitoring.

## Features

### Core Functionality
- **Main Config System** - Centralized configuration with distributed workflow management
- **Source Context Inference** - Workflows automatically inherit source repo/branch
- **$ref Support** - Reusable components for transformations, strategies, and excludes
- **Resilient Loading** - Continues processing when individual configs fail (logs warnings)
- **Automated File Copying** - Copies files from source to target repos on PR merge
- **Advanced Pattern Matching** - Prefix, glob, and regex patterns with variable extraction
- **Path Transformations** - Template-based path transformations with variable substitution
- **Flexible Commit Strategies** - Direct commits or pull requests with auto-merge
- **Deprecation Tracking** - Automatic tracking of deleted files

### Enhanced Features
- **Workflow References** - Local, remote (repo), or inline workflow configs
- **Default Precedence** - Workflow > Workflow config > Main config > System defaults
- **Message Templating** - Template-ized commit messages and PR titles
- **PR Template Integration** - Fetch and merge PR templates from target repos
- **File Exclusion** - Exclude patterns to filter out unwanted files
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

### Local Configuration

1. **Copy environment example file**

```bash
cp env.yaml.example env.yaml
```

2. **Set required environment variables**

```yaml
# GitHub Configuration
GITHUB_APP_ID: "123456"
INSTALLATION_ID: "789012"  # Optional fallback

# Config Repository (where main config lives)
CONFIG_REPO_OWNER: "your-org"
CONFIG_REPO_NAME: "config-repo"
CONFIG_REPO_BRANCH: "main"

# Main Config
MAIN_CONFIG_FILE: ".copier/workflows/main.yaml"
USE_MAIN_CONFIG: "true"

# Secret Manager References
GITHUB_APP_PRIVATE_KEY_SECRET_NAME: "projects/.../secrets/PEM/versions/latest"
WEBHOOK_SECRET_NAME: "projects/.../secrets/webhook-secret/versions/latest"

# Application Settings
WEBSERVER_PATH: "/events"
DEPRECATION_FILE: "deprecated_examples.json"
COMMITTER_NAME: "GitHub Copier App"
COMMITTER_EMAIL: "bot@mongodb.com"

# Feature Flags
AUDIT_ENABLED: "false"
METRICS_ENABLED: "true"
```

3. **Create main configuration file**

Create `.copier/workflows/main.yaml` in your config repository:

```yaml
# Main config with global defaults and workflow references
defaults:
  commit_strategy:
    type: "pull_request"
    auto_merge: false
  exclude:
    - "**/.env"
    - "**/node_modules/**"

workflow_configs:
  # Reference workflows in source repo
  - source: "repo"
    repo: "your-org/source-repo"
    branch: "main"
    path: ".copier/workflows/config.yaml"
    enabled: true
```

4. **Create workflow config in source repository**

Create `.copier/workflows/config.yaml` in your source repository:

```yaml
workflows:
  - name: "copy-examples"
    # source.repo and source.branch inherited from workflow config reference
    destination:
      repo: "your-org/target-repo"
      branch: "main"
    transformations:
      - move: { from: "examples", to: "docs/examples" }
    commit_strategy:
      type: "pull_request"
      pr_title: "Update code examples"
      use_pr_template: true
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

See [MAIN-CONFIG-README.md](configs/copier-config-examples/MAIN-CONFIG-README.md) for complete configuration documentation.

### Main Config Structure

The application uses a three-tier configuration system:

1. **Main Config** - Centralized defaults and workflow references
2. **Workflow Configs** - Collections of workflows (local, remote, or inline)
3. **Individual Workflows** - Specific source → destination mappings

### Transformation Types

#### Move Transformation
Move files from one directory to another:

```yaml
transformations:
  - move:
      from: "examples/go"
      to: "code/go"
```

Moves: `examples/go/main.go` → `code/go/main.go`

#### Copy Transformation
Copy a single file to a new location:

```yaml
transformations:
  - copy:
      from: "README.md"
      to: "docs/README.md"
```

Copies: `README.md` → `docs/README.md`

#### Glob Transformation
Wildcard matching with path transformation:

```yaml
transformations:
  - glob:
      pattern: "examples/*/main.go"
      transform: "code/${relative_path}"
```

Matches: `examples/go/main.go` → `code/examples/go/main.go`

#### Regex Transformation
Full regex with named capture groups:

```yaml
transformations:
  - regex:
      pattern: "^examples/(?P<lang>[^/]+)/(?P<file>.+)$"
      transform: "code/${lang}/${file}"
```

Matches: `examples/go/main.go` → `code/go/main.go` (extracts `lang=go`, `file=main.go`)

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
  use_pr_template: true  # Fetch and merge PR template from target repo
  auto_merge: true
```

### Advanced Features

#### $ref Support for Reusable Components

Extract common configurations into separate files:

```yaml
# Workflow config
workflows:
  - name: "mflix-java"
    destination:
      repo: "mongodb/sample-app-java-mflix"
      branch: "main"
    transformations:
      $ref: "../transformations/mflix-java.yaml"
    commit_strategy:
      $ref: "../strategies/mflix-pr-strategy.yaml"
    exclude:
      $ref: "../common/mflix-excludes.yaml"
```

#### Source Context Inference

Workflows automatically inherit source repo/branch from workflow config reference:

```yaml
# No need to specify source.repo and source.branch!
workflows:
  - name: "my-workflow"
    # source.repo and source.branch inherited automatically
    destination:
      repo: "mongodb/dest-repo"
      branch: "main"
    transformations:
      - move: { from: "src", to: "dest" }
```

#### PR Template Integration

Automatically fetch and merge PR templates from target repositories:

```yaml
commit_strategy:
  type: "pull_request"
  pr_body: "🤖 Automated update"
  use_pr_template: true  # Fetches .github/pull_request_template.md
```

#### File Exclusion

Exclude unwanted files at the workflow or workflow config level:

```yaml
exclude:
  - "**/.gitignore"
  - "**/node_modules/**"
  - "**/.env"
  - "**/dist/**"
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

See [DEPLOYMENT.md](./docs/DEPLOYMENT.md) for complete deployment guide.

### Google Cloud Run

```bash
cd examples-copier
./scripts/deploy-cloudrun.sh
```

### Docker

```bash
docker build -t examples-copier .
docker run -p 8080:8080 --env-file env.yaml examples-copier
```

## Security

- **Webhook Signature Verification** - HMAC-SHA256 validation
- **Secret Management** - Google Cloud Secret Manager
- **Least Privilege** - Minimal GitHub App permissions
- **Audit Trail** - Complete operation logging

## Documentation

### Getting Started

- **[Main Config README](configs/copier-config-examples/MAIN-CONFIG-README.md)** - Complete main config documentation
- **[Quick Start Guide](configs/copier-config-examples/QUICK-START-MAIN-CONFIG.md)** - Get started in 5 minutes
- **[Pattern Matching Guide](docs/PATTERN-MATCHING-GUIDE.md)** - Pattern matching with examples
- **[Local Testing](docs/LOCAL-TESTING.md)** - Test locally before deploying
- **[Deployment Guide](docs/DEPLOYMENT.md)** - Deploy to production

### Reference

- **[Architecture](docs/ARCHITECTURE.md)** - System design and components
- **[Troubleshooting](docs/TROUBLESHOOTING.md)** - Common issues and solutions
- **[FAQ](docs/FAQ.md)** - Frequently asked questions
- **[Deprecation Tracking](docs/DEPRECATION-TRACKING-EXPLAINED.md)** - How deprecation tracking works

### Features

- **[Slack Notifications](docs/SLACK-NOTIFICATIONS.md)** - Slack integration guide
- **[Webhook Testing](docs/WEBHOOK-TESTING.md)** - Test with real PR data

### Tools

- **[Scripts](scripts/README.md)** - Helper scripts for deployment and testing
