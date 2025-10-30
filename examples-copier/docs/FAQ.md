# Frequently Asked Questions (FAQ)

Common questions about the examples-copier application.

## General Questions

### What is examples-copier?

Examples-copier is a GitHub app that automatically copies code examples and files from a source repository to one or more target repositories when pull requests are merged. It features advanced pattern matching, path transformations, and audit logging.

### Why use examples-copier?

- **Automate file synchronization** between repositories
- **Maintain consistency** across multiple documentation repos
- **Track changes** with audit logging
- **Flexible routing** with pattern matching
- **Transform paths** during copying

### What are the main features?

- Advanced pattern matching (prefix, glob, regex)
- Path transformations with variable substitution
- Multiple target repositories
- Flexible commit strategies (direct or PR)
- **Batch PRs** - Combine multiple rules into one PR per target repo
- **PR Template Integration** - Fetch and merge PR templates from target repos
- **File Exclusion** - Exclude patterns to filter out unwanted files
- Deprecation tracking
- MongoDB audit logging
- Health and metrics endpoints
- Slack notifications
- Dry-run mode for testing

## Configuration

### Do I need to use YAML configuration?

No. The app supports both YAML and legacy JSON configurations. YAML is recommended for new deployments because it supports advanced features like pattern matching and path transformations.

### Can I use multiple patterns?

Yes! You can define multiple copy rules, each with its own pattern and targets:

```yaml
copy_rules:
  - name: "Go examples"
    source_pattern:
      type: "regex"
      pattern: "^examples/go/(?P<file>.+)$"
    targets: [...]
  
  - name: "Python examples"
    source_pattern:
      type: "regex"
      pattern: "^examples/python/(?P<file>.+)$"
    targets: [...]
```

### Can one file match multiple rules?

Yes. A file can match multiple rules and be copied to multiple targets. This is useful for copying the same file to different repositories or branches.

### Where should I store the config file?

**For production:** Store `copier-config.yaml` in your source repository (the repo being monitored for PRs).

**For local testing:** Store `copier-config.yaml` in the examples-copier directory and set `CONFIG_FILE=copier-config.yaml`.

### How do I migrate from JSON to YAML?

Use the config-validator tool:

```bash
./config-validator convert -input config.json -output copier-config.yaml
```

The tool will automatically convert your legacy JSON configuration to the new YAML format while preserving all settings.

## Pattern Matching

### Which pattern type should I use?

- **Prefix** - Simple directory matching (e.g., `examples/`)
- **Glob** - Wildcard matching (e.g., `**/*.go`)
- **Regex** - Complex patterns with variable extraction (e.g., `^examples/(?P<lang>[^/]+)/.*$`)

See [Pattern Matching Guide](PATTERN-MATCHING-GUIDE.md) for details.

### How do I extract variables from file paths?

Use regex patterns with named capture groups:

```yaml
source_pattern:
  type: "regex"
  pattern: "^examples/(?P<lang>[^/]+)/(?P<category>[^/]+)/(?P<file>.+)$"
```

This extracts `lang`, `category`, and `file` variables that you can use in path transformations.

### What built-in variables are available?

- `${path}` - Full source file path
- `${filename}` - Just the filename
- `${dir}` - Directory path
- `${ext}` - File extension (with dot)
- `${name}` - Filename without extension

### How do I test my patterns?

Use the config-validator tool:

```bash
./config-validator test-pattern \
  -type regex \
  -pattern "^examples/(?P<lang>[^/]+)/(?P<file>.+)$" \
  -file "examples/go/main.go"
```

### Why aren't my files matching?

Common issues:
- Pattern doesn't match actual file paths
- Missing `^` or `$` anchors in regex
- Wrong pattern type
- Typos in the pattern

Check actual file paths in logs and test your pattern with config-validator.

## Path Transformation

### How do I transform file paths?

Use templates with variable substitution:

```yaml
path_transform: "docs/${lang}/${category}/${file}"
```

Variables come from pattern matching or built-in variables.

### Can I keep the same path?

Yes, use `${path}`:

```yaml
path_transform: "${path}"
```

### Can I flatten the directory structure?

Yes, use just the filename:

```yaml
path_transform: "all-examples/${filename}"
```

### How do I test path transformations?

```bash
./config-validator test-transform \
  -source "examples/go/main.go" \
  -template "docs/${lang}/${file}" \
  -vars "lang=go,file=main.go"
```

## Deployment

### What are the prerequisites?

- Go 1.23.4+
- GitHub App credentials
- Google Cloud project (for Secret Manager and logging)
- MongoDB Atlas (optional, for audit logging)

### Can I run it locally?

Yes! See [Local Testing](LOCAL-TESTING.md) for instructions.

### How do I deploy to Google Cloud?

See [Deployment Guide](DEPLOYMENT.md) for complete guide and [Deployment Checklist](DEPLOYMENT-CHECKLIST.md) for step-by-step instructions.

### Do I need MongoDB?

No, MongoDB is optional. It's used for audit logging. You can disable it:

```bash
export AUDIT_ENABLED=false
```

### Can I use it without Google Cloud?

The app uses Google Cloud Secret Manager for storing GitHub credentials. You could modify it to use environment variables instead, but this requires code changes.

## Testing

### How do I test locally?

1. Start the app in dry-run mode:
   ```bash
   DRY_RUN=true CONFIG_FILE=copier-config.yaml make run-local-quick
   ```

2. Send a test webhook:
   ```bash
   ./test-webhook -payload test-payloads/example-pr-merged.json
   ```

See [Local Testing](LOCAL-TESTING.md) for details.

### What is dry-run mode?

Dry-run mode processes webhooks and matches files but doesn't make actual commits or create PRs. It's perfect for testing configuration changes.

```bash
export DRY_RUN=true
```

### How do I test with real PR data?

Use the test-webhook tool:

```bash
export GITHUB_TOKEN=ghp_your_token
./test-webhook -pr 123 -owner myorg -repo myrepo
```

### How do I validate my configuration?

```bash
./config-validator validate -config copier-config.yaml -v
```

## Operations

### How do I monitor the application?

Use the health and metrics endpoints:

```bash
# Health check
curl http://localhost:8080/health

# Metrics
curl http://localhost:8080/metrics
```

### How do I enable Slack notifications?

Set the Slack webhook URL:

```bash
export SLACK_WEBHOOK_URL="https://hooks.slack.com/services/..."
```

See [Slack Notifications](SLACK-NOTIFICATIONS.md) for details.

### How do I view audit logs?

Query MongoDB:

```javascript
use code_copier
db.audit_events.find().sort({timestamp: -1}).limit(10).pretty()
```

### How do I troubleshoot issues?

1. Check [Troubleshooting Guide](TROUBLESHOOTING.md)
2. Enable debug logging: `export LOG_LEVEL=debug`
3. Check application logs
4. Use config-validator to test patterns

### Can I process PRs manually?

Yes, use the test-webhook tool:

```bash
./test-webhook -pr 123 -owner myorg -repo myrepo
```

## Features

### What commit strategies are supported?

- **Direct** - Commit directly to target branch
- **Pull Request** - Create a PR in target repo (with optional auto-merge)

```yaml
commit_strategy:
  type: "pull_request"
  pr_title: "Update examples"
  auto_merge: false
```

### Can I copy to multiple repositories?

Yes! Each rule can have multiple targets:

```yaml
targets:
  - repo: "org/docs-repo"
    branch: "main"
    path_transform: "examples/${file}"
  
  - repo: "org/website-repo"
    branch: "main"
    path_transform: "static/examples/${file}"
```

### How does deprecation tracking work?

When files are deleted in the source repo, they're tracked in a deprecation file in the target repo. This helps you identify files that should be removed.

### Can I customize commit messages?

Yes, use template variables:

```yaml
commit_strategy:
  type: "pull_request"
  commit_message: "Update ${lang} examples from PR #${pr_number}"
  pr_title: "Update ${lang} examples"
  pr_body: "Automated update (${file_count} files)"
```

### What variables are available for messages?

- `${rule_name}` - Name of the copy rule
- `${source_repo}` - Source repository
- `${target_repo}` - Target repository
- `${source_branch}` - Source branch
- `${target_branch}` - Target branch
- `${file_count}` - Number of files
- `${pr_number}` - PR number
- `${commit_sha}` - Commit SHA
- Plus any variables extracted from pattern matching

### How do I batch multiple rules into one PR?

Use `batch_by_repo: true` to combine all changes into one PR per target repository:

```yaml
batch_by_repo: true

batch_pr_config:
  pr_title: "Update from ${source_repo}"
  pr_body: |
    🤖 Automated update
    Files: ${file_count}  # Accurate count across all rules
  use_pr_template: true
  commit_message: "Update from ${source_repo} PR #${pr_number}"
```

**Benefits:**
- Single PR per target repo instead of multiple PRs
- Accurate `${file_count}` across all matched rules
- Easier review for related changes

### How do I use PR templates from target repos?

Set `use_pr_template: true` in your commit strategy or batch config:

```yaml
commit_strategy:
  type: "pull_request"
  pr_body: |
    🤖 Automated update
    Files: ${file_count}
  use_pr_template: true  # Fetches .github/pull_request_template.md
```

The service will:
1. Fetch the PR template from the target repo
2. Place the template content first (checklists, guidelines)
3. Add a separator (`---`)
4. Append your configured content (automation info)

This ensures reviewers see the target repo's review guidelines prominently.

### How do I exclude files from being copied?

Use `exclude_patterns` in your source pattern:

```yaml
source_pattern:
  type: "prefix"
  pattern: "examples/"
  exclude_patterns:
    - "\.gitignore$"      # Exclude .gitignore
    - "node_modules/"     # Exclude dependencies
    - "\.env$"            # Exclude .env files
    - "/dist/"            # Exclude build output
    - "\.test\.(js|ts)$"  # Exclude test files
```

**Common use cases:**
- Filter out configuration files (`.gitignore`, `.env`)
- Exclude dependencies (`node_modules/`, `vendor/`)
- Skip build artifacts (`/dist/`, `/build/`)
- Exclude test files (`*.test.js`, `*_test.go`)

## Performance

### How many files can it handle?

The app can handle hundreds of files per PR. Performance depends on:
- GitHub API rate limits
- Network latency
- Pattern complexity
- Number of targets

### Is it thread-safe?

Yes, the app uses proper synchronization for concurrent webhook processing.

### What are the rate limits?

GitHub API rate limits apply:
- 5,000 requests/hour for authenticated requests
- Lower limits for unauthenticated requests

## Security

### How are GitHub credentials stored?

GitHub App private key is stored in Google Cloud Secret Manager.

### How are webhooks authenticated?

Webhooks are authenticated using HMAC-SHA256 signature verification with a shared secret.

### Can I disable signature verification?

Yes, for local testing:

```bash
unset WEBHOOK_SECRET
```

**Never disable in production!**

### What permissions does the GitHub App need?

Minimum permissions:
- **Contents**: Read & Write (to read source files and write to target repos)
- **Pull Requests**: Read & Write (to create PRs)
- **Webhooks**: Read (to receive webhook events)

## Troubleshooting

### Files aren't being copied

Check:
1. Pattern matches the file paths
2. Configuration is valid
3. GitHub App has correct permissions
4. Webhook is configured correctly

See [Troubleshooting Guide](TROUBLESHOOTING.md) for details.

### Webhook returns 401

Check:
1. Webhook secret matches
2. Signature verification is working
3. For local testing, disable signature verification

### Application crashes

Check:
1. All required environment variables are set
2. MongoDB connection (if enabled)
3. Google Cloud credentials
4. Application logs for errors

