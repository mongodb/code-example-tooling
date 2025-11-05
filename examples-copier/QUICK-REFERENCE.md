# Quick Reference Guide

## Command Line

### Application

```bash
# Run with default settings
./examples-copier

# Run with custom environment
./examples-copier -env ./configs/.env.production

# Dry-run mode (no actual commits)
./examples-copier -dry-run

# Validate configuration only
./examples-copier -validate

# Show help
./examples-copier -help
```

### CLI Validator

```bash
# Validate config
./config-validator validate -config copier-config.yaml -v

# Test pattern
./config-validator test-pattern -type regex -pattern "^examples/(?P<lang>[^/]+)/.*$" -file "examples/go/main.go"

# Test transformation
./config-validator test-transform -source "examples/go/main.go" -template "docs/${lang}/${file}" -vars "lang=go,file=main.go"

# Initialize new config
./config-validator init -output copier-config.yaml

# Convert formats
./config-validator convert -input config.json -output copier-config.yaml
```

## Configuration Patterns

### Prefix Pattern
```yaml
source_pattern:
  type: "prefix"
  pattern: "examples/go/"
```

### Glob Pattern
```yaml
source_pattern:
  type: "glob"
  pattern: "examples/*/main.go"
```

### Regex Pattern
```yaml
source_pattern:
  type: "regex"
  pattern: "^examples/(?P<lang>[^/]+)/(?P<file>.+)$"
```

### Pattern with Exclusions
```yaml
source_pattern:
  type: "prefix"
  pattern: "examples/"
  exclude_patterns:
    - "\.gitignore$"
    - "node_modules/"
    - "\.env$"
    - "/dist/"
```

## Path Transformations

### Built-in Variables
- `${path}` - Full source path
- `${filename}` - File name only
- `${dir}` - Directory path
- `${ext}` - File extension

### Examples
```yaml
# Keep same path
path_transform: "${path}"

# Change directory
path_transform: "docs/${path}"

# Reorganize structure
path_transform: "docs/${lang}/${category}/${filename}"

# Change extension
path_transform: "${dir}/${filename}.md"
```

## Commit Strategies

### Direct Commit
```yaml
commit_strategy:
  type: "direct"
  commit_message: "Update examples"
```

### Pull Request
```yaml
commit_strategy:
  type: "pull_request"
  commit_message: "Update examples"
  pr_title: "Update code examples"
  pr_body: "Automated update"
  use_pr_template: true  # Fetch PR template from target repo
  auto_merge: true
```

### Batch PRs by Repository
```yaml
batch_by_repo: true

batch_pr_config:
  pr_title: "Update from ${source_repo}"
  pr_body: |
    🤖 Automated update
    Files: ${file_count}
  use_pr_template: true
  commit_message: "Update from ${source_repo} PR #${pr_number}"
```

## Advanced Features

### Exclude Patterns
Exclude unwanted files from being copied:

```yaml
source_pattern:
  type: "prefix"
  pattern: "examples/"
  exclude_patterns:
    - "\.gitignore$"      # Exclude .gitignore
    - "node_modules/"     # Exclude dependencies
    - "\.env$"            # Exclude .env files
    - "\.env\\..*$"       # Exclude .env.local, .env.production, etc.
    - "/dist/"            # Exclude build output
    - "/build/"           # Exclude build artifacts
    - "\.test\.(js|ts)$"  # Exclude test files
```

### PR Template Integration
Fetch and merge PR templates from target repos:

```yaml
commit_strategy:
  type: "pull_request"
  pr_body: |
    🤖 Automated update
    Files: ${file_count}
  use_pr_template: true  # Fetches .github/pull_request_template.md
```

**Result:** PR description shows:
1. Target repo's PR template (checklists, guidelines)
2. Separator (`---`)
3. Your configured content (automation info)

### Batch Configuration
When `batch_by_repo: true`, use `batch_pr_config` for accurate file counts:

```yaml
batch_by_repo: true

batch_pr_config:
  pr_title: "Update from ${source_repo}"
  pr_body: |
    Files: ${file_count}  # Accurate count across all rules
    Source: ${source_repo} PR #${pr_number}
  use_pr_template: true
```

## Message Templates

### Available Variables
- `${rule_name}` - Copy rule name (e.g., "java-aggregation-examples")
- `${source_repo}` - Source repository (e.g., "mongodb/aggregation-tasks")
- `${target_repo}` - Target repository (e.g., "mongodb/vector-search")
- `${source_branch}` - Source branch (e.g., "main")
- `${target_branch}` - Target branch (e.g., "main")
- `${file_count}` - Number of files (e.g., "3")
- `${pr_number}` - Source PR number (e.g., "42")
- `${commit_sha}` - Source commit SHA (e.g., "abc123")
- Custom variables from regex patterns (e.g., `${lang}`, `${file}`)

### Examples
```yaml
commit_message: "Update ${category} examples from ${lang}"
pr_title: "Update ${lang} examples"
pr_body: |
  Files updated: ${file_count} using ${rule_name} match pattern

  Source: ${source_repo}
  PR: #${pr_number}
```

## API Endpoints

### Health Check
```bash
curl http://localhost:8080/health
```

### Metrics
```bash
curl http://localhost:8080/metrics
```

### Webhook
```bash
curl -X POST http://localhost:8080/webhook \
  -H "Content-Type: application/json" \
  -H "X-Hub-Signature-256: sha256=..." \
  -d @webhook-payload.json
```

## Environment Variables

### Required
```bash
REPO_OWNER=your-org
REPO_NAME=your-repo
GITHUB_APP_ID=123456
GITHUB_INSTALLATION_ID=789012
GCP_PROJECT_ID=your-project
PEM_KEY_NAME=projects/123/secrets/KEY/versions/latest
```

### Optional
```bash
# Application
PORT=8080
CONFIG_FILE=copier-config.yaml
DEPRECATION_FILE=deprecated_examples.json
DRY_RUN=false

# Logging
LOG_LEVEL=info
COPIER_DEBUG=false
COPIER_DISABLE_CLOUD_LOGGING=false

# Audit
AUDIT_ENABLED=true
MONGO_URI=mongodb+srv://...
AUDIT_DATABASE=code_copier
AUDIT_COLLECTION=audit_events

# Metrics
METRICS_ENABLED=true

# Webhook
WEBHOOK_SECRET=your-secret
```

## MongoDB Queries

### Recent Events
```javascript
db.audit_events.find().sort({timestamp: -1}).limit(10)
```

### Failed Operations
```javascript
db.audit_events.find({success: false}).sort({timestamp: -1})
```

### Events by Rule
```javascript
db.audit_events.find({rule_name: "Copy Go examples"})
```

### Statistics
```javascript
db.audit_events.aggregate([
  {$match: {event_type: "copy"}},
  {$group: {
    _id: "$rule_name",
    count: {$sum: 1},
    avg_duration: {$avg: "$duration_ms"}
  }}
])
```

### Success Rate
```javascript
db.audit_events.aggregate([
  {$group: {
    _id: "$success",
    count: {$sum: 1}
  }}
])
```

## Testing

### Run Unit Tests
```bash
# All tests
go test ./services -v

# Specific test
go test ./services -v -run TestPatternMatcher

# With coverage
go test ./services -cover
```

### Test with Webhooks

#### Option 1: Use Example Payload
```bash
# Build test tool
go build -o test-webhook ./cmd/test-webhook

# Send example payload
./test-webhook -payload test-payloads/example-pr-merged.json

# Dry-run (see payload without sending)
./test-webhook -payload test-payloads/example-pr-merged.json -dry-run
```

#### Option 2: Use Real PR Data
```bash
# Set GitHub token
export GITHUB_TOKEN=ghp_your_token_here

# Fetch and send real PR data
./test-webhook -pr 123 -owner myorg -repo myrepo

# Test against production
./test-webhook -pr 123 -owner myorg -repo myrepo \
  -url https://myapp.appspot.com/webhook \
  -secret "my-webhook-secret"
```

#### Option 3: Use Helper Script (Interactive)
```bash
# Make executable
chmod +x scripts/test-with-pr.sh

# Run interactive test
./scripts/test-with-pr.sh 123 myorg myrepo
```

### Test in Dry-Run Mode
```bash
# Start app in dry-run mode
DRY_RUN=true ./examples-copier &

# Send test webhook
./test-webhook -pr 123 -owner myorg -repo myrepo

# Check logs (no actual commits made)
```

### Build
```bash
# Main application
go build -o examples-copier .

# CLI validator
go build -o config-validator ./cmd/config-validator

# Test webhook tool
go build -o test-webhook ./cmd/test-webhook

# All tools
go build -o examples-copier . && \
go build -o config-validator ./cmd/config-validator && \
go build -o test-webhook ./cmd/test-webhook
```

## Common Patterns

### Copy All Go Files
```yaml
source_pattern:
  type: "regex"
  pattern: "^examples/.*\\.go$"
targets:
  - repo: "org/docs"
    path_transform: "code/${path}"
```

### Organize by Language
```yaml
source_pattern:
  type: "regex"
  pattern: "^examples/(?P<lang>[^/]+)/(?P<rest>.+)$"
targets:
  - repo: "org/docs"
    path_transform: "languages/${lang}/${rest}"
```

### Multiple Targets with Different Transforms
```yaml
source_pattern:
  type: "prefix"
  pattern: "examples/"
targets:
  - repo: "org/docs-v1"
    path_transform: "examples/${path}"
  - repo: "org/docs-v2"
    path_transform: "code-samples/${path}"
```

### Conditional Copying (by file type)
```yaml
source_pattern:
  type: "regex"
  pattern: "^examples/.*\\.(?P<ext>go|py|js)$"
targets:
  - repo: "org/docs"
    path_transform: "code/${ext}/${filename}"
```

## Troubleshooting

### Check Logs
```bash
# Application logs
gcloud app logs tail -s default

# Local logs
LOG_LEVEL=debug ./examples-copier
```

### Validate Config
```bash
./config-validator validate -config copier-config.yaml -v
```

### Test Pattern Matching
```bash
./config-validator test-pattern \
  -type regex \
  -pattern "your-pattern" \
  -file "test/file.go"
```

### Dry Run
```bash
DRY_RUN=true ./examples-copier
```

### Check Health
```bash
curl http://localhost:8080/health
```

### Check Metrics
```bash
curl http://localhost:8080/metrics | jq
```

## Deployment 

### Google Cloud Quick Commands

```bash
# Deploy (env.yaml is included via 'includes' directive in app.yaml)
gcloud app deploy app.yaml

# View logs
gcloud app logs tail -s default

# Check health
curl https://github-copy-code-examples.appspot.com/health

# List secrets
gcloud secrets list

# Grant access
./grant-secret-access.sh
```



## File Locations

```
examples-copier/
├── README.md                 # Main documentation
├── QUICK-REFERENCE.md        # This file
├── docs/
│   ├── ARCHITECTURE.md       # Architecture overview
│   ├── CONFIGURATION-GUIDE.md # Complete config reference
│   ├── DEPLOYMENT.md         # Deployment guide
│   ├── DEPLOYMENT-CHECKLIST.md  # Deployment checklist
│   ├── FAQ.md                # Frequently asked questions
│   ├── LOCAL-TESTING.md      # Local testing guide
│   ├── PATTERN-MATCHING-GUIDE.md # Pattern matching guide
│   ├── PATTERN-MATCHING-CHEATSHEET.md # Quick pattern reference
│   ├── TROUBLESHOOTING.md    # Troubleshooting guide
│   └── WEBHOOK-TESTING.md    # Webhook testing guide
├── configs/
│   ├── .env                  # Environment config
│   ├── env.yaml.example      # Environment template
│   └── copier-config.example.yaml # Config template
└── cmd/
    ├── config-validator/     # CLI validation tool
    └── test-webhook/         # Webhook testing tool
```

## Quick Start Checklist

- [ ] Clone repository
- [ ] Copy `configs/.env.local.example` to `configs/.env`
- [ ] Set required environment variables
- [ ] Create `copier-config.yaml` in source repo
- [ ] Validate config: `./config-validator validate -config copier-config.yaml`
- [ ] Test in dry-run: `DRY_RUN=true ./examples-copier`
- [ ] Deploy: `./examples-copier`
- [ ] Configure GitHub webhook
- [ ] Monitor: `curl http://localhost:8080/health`

## Support

- **Documentation**: [README.md](README.md)
- **Configuration**: [Configuration Guide](./docs/CONFIGURATION-GUIDE.md)
- **Deployment**: [Deployment Guide](./docs/DEPLOYMENT.md)
- **Troubleshooting**: [Troubleshooting Guide](./docs/TROUBLESHOOTING.md)
- **FAQ**: [Frequently Asked Questions](./docs/FAQ.md)

