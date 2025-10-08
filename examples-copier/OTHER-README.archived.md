# GitHub Code Example Copier

A production-ready GitHub App that automatically copies code examples between repositories when pull requests are merged. Built with Go, it provides intelligent pattern matching, path transformation, batch uploads, and deprecation tracking.

## Features

- **Automatic Code Copying**: Copies code examples when PRs are merged
- **Pattern Matching**: Supports prefix, glob, and regex patterns with variable extraction
- **Path Transformation**: Transform source paths to target paths with variable substitution
- **Batch Uploads**: Efficient batch uploads using GitHub Tree API (single commit for multiple files)
- **Deprecation Tracking**: Automatically tracks deleted files in deprecation lists
- **Multiple Targets**: Copy files to multiple repositories with different transformations
- **Commit Strategies**: Direct commits, pull requests, or batch commits
- **Security**: Webhook signature verification, rate limiting, input validation
- **Health Checks**: Built-in health and info endpoints for monitoring
- **CLI Tools**: Management CLI with metrics, audit logs, live dashboard, and configuration utilities

## Table of Contents

- [Quick Start](#quick-start)
- [Installation](#installation)
- [Configuration](#configuration)
- [Pattern Matching](#pattern-matching)
- [Path Transformations](#path-transformations)
- [Commit Strategies](#commit-strategies)
- [Deprecation Tracking](#deprecation-tracking)
- [CLI Tools](#cli-tools)
- [Deployment](#deployment)
- [API Endpoints](#api-endpoints)
- [Development](#development)
- [Testing](#testing)
- [Security](#security)
- [Troubleshooting](#troubleshooting)

## 🚀 Quick Start

### Prerequisites

- Go 1.23.4 or later
- GitHub App with appropriate permissions
- Google Cloud Project (for Secret Manager)

### Installation

```bash
# Clone the repository
git clone https://github.com/mongodb/code-example-tooling.git
cd code-example-tooling/examples-copier

# Build the application
go build -o examples-copier .

# Build the CLI tools
go build -o copier-cli ./cmd/copier-cli
go build -o config-tool ./cmd/config-tool

# Or build all at once using make
make build-all
```

### Configuration

1. Create a `.env` file:

```bash
# GitHub App Configuration
GITHUB_APP_CLIENT_ID=your-app-id
INSTALLATION_ID=your-installation-id
GITHUB_APP_PRIVATE_KEY_SECRET_NAME=projects/your-project/secrets/github-key/versions/latest
WEBHOOK_SECRET=your-webhook-secret

# Repository Configuration
REPO_OWNER=mongodb
REPO_NAME=docs-examples

# File Configuration
CONFIG_FILE=copier-config.yaml
DEPRECATION_FILE=deprecated_examples.json
CONFIG_BRANCH=main  # Branch to fetch config from

# Server Configuration
PORT=8080
LOG_LEVEL=info

# Google Cloud Configuration
GOOGLE_PROJECT_ID=your-project-id

# Audit Logging (Optional)
AUDIT_ENABLED=false
MONGO_URI=mongodb://localhost:27017
AUDIT_DATABASE=copier_audit
AUDIT_COLLECTION=events
```

2. Create a configuration file:

```bash
# Initialize from template
./config-tool init -template basic -output copier-config.yaml

# Or create manually
cat > copier-config.yaml << EOF
source_repo: "mongodb/docs-examples"
source_branch: "main"

copy_rules:
  - name: "go-examples"
    source_pattern:
      type: "glob"
      pattern: "examples/go/**/*.go"
    targets:
      - repo: "mongodb/docs"
        branch: "main"
        path_transform: "source/code-examples/go/\${path}"
EOF
```

3. Run the application:

```bash
./examples-copier
```

## 💻 Command-Line Interface

### Available Flags

| Flag         | Type   | Description                                      | Example                        |
|--------------|--------|--------------------------------------------------|--------------------------------|
| `-env`       | string | Path to environment file                         | `-env ./configs/.env`          |
| `-config`    | string | Path to YAML config file (overrides CONFIG_FILE) | `-config /path/to/config.yaml` |
| `-port`      | string | HTTP server port (overrides PORT)                | `-port 8080`                   |
| `-log-level` | string | Log level: debug, info, warn, error              | `-log-level debug`             |
| `-dry-run`   | bool   | Enable dry-run mode (no changes)                 | `-dry-run`                     |
| `-validate`  | bool   | Validate configuration and exit                  | `-validate`                    |
| `-version`   | bool   | Show version information                         | `-version`                     |
| `-help`      | bool   | Show help information                            | `-help`                        |

### Common Usage Examples

```bash
# Show help
./examples-copier -help

# Show version
./examples-copier -version

# Validate configuration
./examples-copier -validate

# Start with custom environment file
./examples-copier -env /path/to/.env

# Start on custom port
./examples-copier -port 8080

# Enable debug logging
./examples-copier -log-level debug

# Dry-run mode (no actual changes)
./examples-copier -dry-run

# Combine multiple flags
./examples-copier -env .env.test -port 8080 -log-level debug -dry-run
```

### Configuration Priority

Command-line flags override environment variables:

1. **Command-line flags** (highest priority)
2. **Environment variables**
3. **.env file**
4. **Default values** (lowest priority)

Example:
```bash
# PORT in .env is 3000, but command-line overrides it to 8080
./examples-copier -port 8080
```

### Startup Banner

When the application starts, it displays a banner with configuration summary:

```
╔════════════════════════════════════════════════════════════════╗
║  GitHub Code Example Copier v1.0.0                             ║
╠════════════════════════════════════════════════════════════════╣
║  Port:         3000                                            ║
║  Webhook Path: /webhook                                        ║
║  Log Level:    info                                            ║
╚════════════════════════════════════════════════════════════════╝
```

## Configuration

> **📖 For complete configuration documentation, see [docs/CONFIGURATION.md](docs/CONFIGURATION.md)**
> This section covers repository configuration files. For environment variables and deployment configuration, see the master configuration guide.

### Configuration File Structure

The configuration file defines how files are copied between repositories.

**YAML Example:**

```yaml
source_repo: "mongodb/docs-examples"
source_branch: "main"

copy_rules:
  - name: "go-examples"
    source_pattern:
      type: "regex"
      pattern: "^examples/(?P<lang>go)/(?P<category>[^/]+)/(?P<file>.+)$"
    targets:
      - repo: "mongodb/docs"
        branch: "main"
        path_transform: "source/code-examples/${lang}/${category}/${file}"
        commit_strategy:
          type: "direct"
      - repo: "mongodb/tutorials"
        branch: "main"
        path_transform: "examples/${lang}/${file}"
        commit_strategy:
          type: "pull_request"
          pr_title: "Update ${lang} examples"
          pr_body: "Automated update from source repository"
        deprecation_check:
          enabled: true
          file: "deprecated_examples.json"
```

### Configuration Fields

- **source_repo**: Source repository (owner/name)
- **source_branch**: Default source branch
- **copy_rules**: Array of copy rules
    - **name**: Rule name (for logging)
    - **source_pattern**: Pattern configuration
        - **type**: Pattern type (prefix, glob, regex)
        - **pattern**: Pattern string
    - **targets**: Array of target configurations
        - **repo**: Target repository (owner/name)
        - **branch**: Target branch
        - **path_transform**: Path transformation template
        - **commit_strategy**: Commit strategy configuration
        - **deprecation_check**: Deprecation check configuration

## Pattern Matching

### Pattern Types

#### 1. Prefix Pattern

Matches files that start with a specific prefix.

```yaml
source_pattern:
  type: "prefix"
  pattern: "examples/go/"
```

Matches:
- `examples/go/main.go` ✓
- `examples/go/auth/login.go` ✓
- `examples/python/main.py` ✗

#### 2. Glob Pattern

Matches files using glob syntax (supports `*`, `**`, `?`).

```yaml
source_pattern:
  type: "glob"
  pattern: "examples/**/*.go"
```

Matches:
- `examples/go/main.go` ✓
- `examples/go/auth/login.go` ✓
- `examples/python/main.py` ✗

#### 3. Regex Pattern

Matches files using regular expressions with named capture groups.

```yaml
source_pattern:
  type: "regex"
  pattern: "^examples/(?P<lang>[^/]+)/(?P<file>.+)$"
```

Matches:
- `examples/go/main.go` ✓ (lang=go, file=main.go)
- `examples/python/auth.py` ✓ (lang=python, file=auth.py)

### Testing Patterns

Use the CLI tool to test patterns:

```bash
# Test glob pattern
./config-tool test-pattern \
  -pattern "examples/**/*.go" \
  -type glob \
  -file "examples/go/main.go"

# Test regex pattern
./config-tool test-pattern \
  -pattern "^examples/(?P<lang>[^/]+)/(?P<file>.+)$" \
  -type regex \
  -file "examples/go/main.go"
```

## Path Transformations

Path transformations use variable substitution to transform source paths to target paths.

### Built-in Variables

- `${path}`: Full file path
- `${filename}`: File name only
- `${dir}`: Directory path
- `${ext}`: File extension

### Regex Variables

When using regex patterns with named groups, extracted variables are available:

```yaml
source_pattern:
  type: "regex"
  pattern: "^examples/(?P<lang>[^/]+)/(?P<category>[^/]+)/(?P<file>.+)$"

targets:
  - repo: "mongodb/docs"
    path_transform: "source/code-examples/${lang}/${category}/${file}"
```

**Example:**
- Source: `examples/go/auth/login.go`
- Variables: `lang=go`, `category=auth`, `file=login.go`
- Target: `source/code-examples/go/auth/login.go`

## Commit Strategies

### Direct Commit

Commits directly to the target branch.

```yaml
commit_strategy:
  type: "direct"
  commit_message: "Update code examples"
```

### Pull Request

Creates a pull request for review.

```yaml
commit_strategy:
  type: "pull_request"
  pr_title: "Update ${lang} examples"
  pr_body: "Automated update from source repository"
```

### Batch Commit

Groups multiple files into a single commit.

```yaml
commit_strategy:
  type: "batch"
  batch_size: 100
  commit_message: "Batch update code examples"
```

## Deprecation Tracking

When files are deleted from the source repository, they can be automatically tracked in a deprecation file. This provides a historical record of removed files without deleting them from target repositories.

**Key Principle:** The copier **does not delete files** from target repositories. Instead, it records deletions in a deprecation file in the **source repository**.

### Configuration

Enable deprecation tracking per target repository:

```yaml
targets:
  - repo: "mongodb/docs"
    branch: "main"
    path_transform: "source/code-examples/${lang}/${file}"
    deprecation_check:
      enabled: true
      file: "deprecated_examples.json"  # Optional, defaults to deprecated_examples.json
```

### Deprecation File Format

The deprecation file is stored as JSON in the **source repository**:

```json
{
  "files": [
    {
      "path": "examples/old/deprecated.go",
      "reason": "File removed from source repository (rule: go-examples)",
      "since": "2025-10-06"
    }
  ]
}
```

### How It Works

1. **Detection**: When a PR is merged, deleted files are identified (status: "removed")
2. **Matching**: Deleted files are matched against copy rules
3. **Tracking**: Matched files are added to the deprecation queue
4. **Update**: The deprecation file in the source repository is updated with new entries
5. **Preservation**: Target repositories keep their files (no deletions occur)

### Field Descriptions

| Field    | Type   | Description                                               |
|----------|--------|-----------------------------------------------------------|
| `path`   | string | The file path that was deleted from the source repository |
| `reason` | string | Why it was deprecated (auto-generated with rule name)     |
| `since`  | string | Date of deprecation in ISO 8601 format (YYYY-MM-DD)       |

### Example Use Case

```
1. Developer deletes examples/go/auth/basic-auth.go
2. PR is merged to main branch
3. Webhook triggers the copier
4. File matches the "go-examples" rule
5. Deprecation file is updated in source repo
6. Target repositories (mongodb/docs) keep their copies
7. Documentation teams can see the file was removed and update references
```

### Monitoring

View deprecation events using the CLI:

```bash
# View recent deprecations
copier-cli audit search --query "event_type:deprecation" --limit 50

# View deprecation statistics
copier-cli audit stats --since 7d
```

For complete details, see [Deprecation Tracking Documentation](docs/DEPRECATION-TRACKING.md).

## CLI Tools

The project includes two CLI tools for managing and monitoring the application:

### copier-cli - Management & Monitoring CLI

A comprehensive CLI for managing and monitoring the running application.

```bash
# Build the CLI
make build-copier-cli

# Or build manually
go build -o copier-cli ./cmd/copier-cli
```

#### Key Features

- **Real-time Metrics** - Query and monitor application metrics
- **Audit Logging** - Query MongoDB audit logs and statistics
- **Live Dashboard** - Interactive dashboard with metrics and audit data
- **Health Checks** - Monitor application health and service status
- **Configuration Management** - Validate, reload, and compare configurations
- **Multi-Environment Support** - Manage local, staging, and production environments

#### Quick Start

```bash
# Check application health
copier-cli health

# View current metrics
copier-cli metrics status

# Launch live dashboard
copier-cli dashboard

# View recent audit events (requires MongoDB)
export MONGO_URI="mongodb://localhost:27017"
copier-cli audit recent --limit 20

# Compare metrics between environments
copier-cli metrics compare \
  --env1 http://localhost:3000 \
  --env2 https://staging.run.app
```

See [cmd/copier-cli/README.md](cmd/copier-cli/README.md) for complete documentation.

### config-tool - Configuration Utilities

A utility for managing configuration files (legacy tool, consider using `copier-cli config` commands).

#### Validate Configuration

```bash
./config-tool validate -config copier-config.yaml -v
```

#### Convert Between Formats

```bash
# YAML to JSON
./config-tool convert -input config.yaml -output config.json

# JSON to YAML
./config-tool convert -input config.json -output config.yaml
```

#### Initialize Configuration

```bash
# Basic template
./config-tool init -template basic -output copier-config.yaml

# Advanced template
./config-tool init -template advanced -output copier-config.yaml -format yaml
```

#### Test Pattern Matching

```bash
./config-tool test-pattern \
  -pattern "examples/**/*.go" \
  -type glob \
  -file "examples/go/main.go"
```

## Deployment

### Docker

```dockerfile
FROM golang:1.23.4-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o examples-copier .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/examples-copier .
EXPOSE 8080
CMD ["./examples-copier"]
```

### Google Cloud Run

```bash
gcloud run deploy examples-copier \
  --source . \
  --platform managed \
  --region us-central1 \
  --allow-unauthenticated
```

## API Endpoints

### GET /

Service description and endpoint list.

### GET /health

Health check endpoint.

**Response:**
```json
{
  "status": "healthy",
  "started": true,
  "github": {"status": "healthy"},
  "queues": {
    "upload_count": 0,
    "deprecation_count": 0
  }
}
```

### GET /info

Service information.

**Response:**
```json
{
  "version": "1.0.0",
  "app_id": "123456",
  "repo_owner": "mongodb",
  "repo_name": "docs-examples",
  "started": true
}
```

### GET /metrics

Application metrics endpoint (enabled by default).

**Response:**
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
      "max_ms": 3200.8,
      "p50_ms": 1100.0,
      "p95_ms": 2500.0,
      "p99_ms": 3000.0
    }
  },
  "files": {
    "matched": 320,
    "uploaded": 310,
    "upload_failed": 5,
    "deprecated": 5,
    "upload_success_rate": 98.41,
    "upload_time": {
      "avg_ms": 850.3,
      "min_ms": 200.1,
      "max_ms": 2100.5,
      "p50_ms": 750.0,
      "p95_ms": 1800.0,
      "p99_ms": 2000.0
    }
  },
  "github_api": {
    "calls": 1250,
    "errors": 12,
    "error_rate": 0.96,
    "rate_limit": {
      "remaining": 4850,
      "reset_at": "2025-10-04T13:45:00Z"
    }
  },
  "queues": {
    "upload_queue_size": 5,
    "deprecation_queue_size": 2,
    "retry_queue_size": 1
  },
  "system": {
    "uptime_seconds": 86400
  }
}
```

**Configuration:**
```bash
# Enable/disable metrics collection (default: true)
METRICS_ENABLED=true
```

### POST /webhook

GitHub webhook handler.

**Headers:**
- `X-GitHub-Event`: Event type
- `X-GitHub-Delivery`: Delivery ID
- `X-Hub-Signature-256`: HMAC signature

## Audit Logging

The application supports optional audit logging to MongoDB for tracking all file copy operations, deprecations, and errors.

### Configuration

Enable audit logging by setting these environment variables:

```bash
AUDIT_ENABLED=true
MONGO_URI=mongodb://localhost:27017
AUDIT_DATABASE=copier_audit
AUDIT_COLLECTION=events
```

### Event Types

**Copy Events** - Logged when files are successfully copied or when copy operations fail:
- Source repository and path
- Target repository and path
- Rule name
- Commit SHA
- PR number
- File size
- Duration (milliseconds)
- Success/failure status
- Error message (if failed)

**Deprecation Events** - Logged when files are marked for deprecation:
- Source repository and path
- Rule name
- Commit SHA
- PR number

**Error Events** - Logged when operations fail:
- Source repository and path
- Target repository (if applicable)
- Error message
- Duration

### Querying Audit Logs

The audit logger provides methods for querying events:

- `GetRecentEvents(limit)` - Get recent events
- `GetFailedEvents(limit)` - Get failed operations
- `GetEventsByRule(ruleName, limit)` - Get events for a specific rule
- `GetStatsByRule()` - Get statistics grouped by rule
- `GetDailyVolume(days)` - Get daily copy volume statistics

### MongoDB Indexes

The following indexes are automatically created:
- `timestamp` (descending) - For time-based queries
- `event_type` - For filtering by event type
- `rule_name` - For rule-specific queries
- `success` - For filtering successful/failed operations
- `source_repo` - For source repository queries

## Testing

```bash
# Run all tests
go test ./... -cover

# Run specific package tests
go test ./services -v

# Run with coverage report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## Security

- **Webhook Signature Verification**: HMAC SHA-256
- **Rate Limiting**: 100 requests/hour per delivery ID
- **Input Validation**: XSS, SQL injection, command injection protection
- **Secret Management**: Google Cloud Secret Manager
- **GitHub App Authentication**: JWT tokens with expiration

## Troubleshooting

See [TROUBLESHOOTING.md](TROUBLESHOOTING.md) for common issues and solutions.

## Documentation

- [Architecture](./docs/ARCHITECTURE.md)
- [Security](./docs/SECURITY.md)
- [Pattern Matching Guide](./docs/PATTERN-MATCHING-GUIDE.md)
- [CLI Guide](./docs/CLI-GUIDE.md)

## License

Apache 2.0. See [LICENSE](LICENSE).

