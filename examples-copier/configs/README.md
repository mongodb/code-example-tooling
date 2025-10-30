# Environment Configuration Guide

This directory contains **template files** for different deployment scenarios. Copy the appropriate template to create your working configuration file.

## Template Files Overview

| Template File         | Purpose                               | Use Case                        |
|-----------------------|---------------------------------------|---------------------------------|
| `env.yaml.example`    | Complete reference with all variables | First-time setup, documentation |
| `env.yaml.production` | Production-ready template             | Quick deployment to production  |
| `.env.local.example`  | Local development template            | Local testing and development   |

**Note:** Your actual working files (`env.yaml`, `.env`) should be created from these templates and are gitignored to protect secrets.

---

## env.yaml.example

**Location:** `configs/env.yaml.example`

**Purpose:** Comprehensive reference showing ALL possible environment variables

**Contents:**
- ✅ All 30+ supported variables
- ✅ Detailed comments for each variable
- ✅ Default values shown
- ✅ Security best practices
- ✅ Deployment notes
- ✅ Local development tips

**Use this when:**
- Setting up for the first time
- Need to understand all available options
- Want to see what features are available
- Need reference documentation

---

## env.yaml.production

**Location:** `configs/env.yaml.production`

**Purpose:** Production-ready template with sensible defaults

**Contents:**
- ✅ Required variables pre-filled with MongoDB values
- ✅ Secret Manager references (recommended approach)
- ✅ Essential settings only
- ✅ Audit logging enabled
- ✅ Metrics enabled
- ✅ Production-optimized

**Use this when:**
- Deploying to production quickly
- Want a minimal, clean configuration
- Using Secret Manager (recommended)
- Don't need advanced features

**NOT included:**
- Slack notifications (optional)
- MongoDB direct URI (use Secret Manager instead)
- Advanced options (rarely needed)

---

## .env.local.example

**Location:** `configs/.env.local.example`

**Purpose:** Local development template (traditional .env format)

**Contents:**
- ✅ .env format (KEY=value)
- ✅ Suitable for local testing
- ✅ Works with godotenv
- ✅ Can use direct secrets (not Secret Manager)

**Use this when:**
- Developing locally
- Testing without Google Cloud
- Using `go run` or local server
- Don't want to set up Secret Manager

**Format difference:**
```bash
# .env format (for local development)
GITHUB_APP_ID=123456
REPO_OWNER=mongodb
REPO_NAME=docs-code-examples
```

vs

```yaml
# env.yaml format (for App Engine deployment)
env_variables:
  GITHUB_APP_ID: "123456"
  REPO_OWNER: "mongodb"
  REPO_NAME: "docs-code-examples"
```

---

## Deployment Targets

This service supports **two Google Cloud deployment options**:

### App Engine (Flexible Environment)

**Config file:** `env.yaml` (with `env_variables:` wrapper)

**Format:**
```yaml
env_variables:
  GITHUB_APP_ID: "123456"
  REPO_OWNER: "mongodb"
```

**Deploy:**
```bash
cp configs/env.yaml.production env.yaml
# Edit env.yaml with your values
gcloud app deploy app.yaml  # Includes env.yaml automatically
```

**Best for:** Long-running services, always-on applications

---

### Cloud Run (Serverless Containers)

**Config file:** `env-cloudrun.yaml` (plain YAML, no wrapper)

**Format:**
```yaml
GITHUB_APP_ID: "123456"
REPO_OWNER: "mongodb"
```

**Deploy:**
```bash
cp configs/env.yaml.production env-cloudrun.yaml
# Remove the 'env_variables:' wrapper
# Edit env-cloudrun.yaml with your values
gcloud run deploy examples-copier --source . --env-vars-file=env-cloudrun.yaml
```

**Best for:** Cost-effective, scales to zero, serverless

---

## Usage Scenarios

### Scenario 1: First-Time Production Deployment

**Recommended:** `env.yaml.production`

```bash
# Quick start
cp configs/env.yaml.production env.yaml
nano env.yaml  # Update PROJECT_NUMBER and values
./scripts/grant-secret-access.sh
gcloud app deploy app.yaml  # env.yaml is included via 'includes' directive
```

**Why:** Pre-configured with production best practices, minimal setup required.

---

### Scenario 2: Need to Understand All Options

**Recommended:** `env.yaml.example`

```bash
# Reference all options
cat configs/env.yaml.example

# Copy and customize
cp configs/env.yaml.example env.yaml
nano env.yaml  # Enable features you need
```

**Why:** Shows all available features with detailed explanations.

---

### Scenario 3: Local Development

**Recommended:** `.env.local.example`

```bash
# Local development
cp configs/.env.local.example configs/.env
nano configs/.env  # Add your values

# Run locally
go run app.go -env configs/.env
```

**Why:** Simpler format for local testing, no Secret Manager required.

---

### Scenario 4: Custom Production Setup

**Recommended:** Start with `env.yaml.example`, customize

```bash
# Start with full reference
cp configs/env.yaml.example env.yaml

# Enable features you need
nano env.yaml
# - Enable Slack notifications
# - Configure custom MongoDB settings
# - Set custom defaults

# Deploy
gcloud app deploy app.yaml  # env.yaml is included via 'includes' directive
```

**Why:** Need advanced features not in production template.

---

## Migration Guide

### From .env to env.yaml

Use the conversion script:

```bash
./scripts/convert-env-to-yaml.sh configs/.env env.yaml
```

Or manually convert:

```bash
# .env format:
GITHUB_APP_ID=123456
REPO_OWNER=mongodb

# env.yaml format:
env_variables:
  GITHUB_APP_ID: "123456"
  REPO_OWNER: "mongodb"
```

### Between App Engine and Cloud Run formats

Use the format conversion script:

```bash
# Convert App Engine → Cloud Run
./scripts/convert-env-format.sh to-cloudrun env.yaml env-cloudrun.yaml

# Convert Cloud Run → App Engine
./scripts/convert-env-format.sh to-appengine env-cloudrun.yaml env.yaml
```

**Key difference:**
- **App Engine**: Requires `env_variables:` wrapper with 2-space indentation
- **Cloud Run**: Plain YAML without wrapper

### From env.yaml.production to env.yaml.example

```bash
# Start with production template
cp configs/env.yaml.production env.yaml

# Add optional features from example
# Compare files and add what you need:
diff configs/env.yaml.production configs/env.yaml.example
```

---

## Best Practices

### ✅ DO

- **Use `env.yaml.production` for quick production deployment**
- **Use `env.yaml.example` as reference documentation**
- **Use `.env.local.example` for local development**
- **Add `env.yaml` and `.env` to `.gitignore`**
- **Use Secret Manager for production secrets**
- **Keep comments in your env.yaml for team documentation**

### ❌ DON'T

- **Don't commit `env.yaml` or `.env` with actual secrets**
- **Don't use direct secrets in production (use Secret Manager)**
- **Don't mix .env and env.yaml formats**
- **Don't remove all comments (they help future you)**
- **Don't use `env.yaml.production` as-is (update values first)**

---

## File Locations

```
examples-copier/
├── configs/
│   ├── env.yaml.example          # ← Complete reference (all variables)
│   ├── env.yaml.production       # ← Production template (essential only)
│   └── .env.local.example        # ← Local development template
├── env.yaml                      # ← App Engine config (create from template, gitignored)
├── env-cloudrun.yaml             # ← Cloud Run config (create from template, gitignored)
└── .env                          # ← Local config (create from template, gitignored)
```

**Working files (not in repo):**
- `env.yaml` - App Engine deployment (YAML with `env_variables:` wrapper)
- `env-cloudrun.yaml` - Cloud Run deployment (plain YAML, no wrapper)
- `.env` - Local development (KEY=value format)

---

## Quick Reference

**Need to deploy quickly?**
→ Use `env.yaml.production`

**Need to understand all options?**
→ Read `env.yaml.example`

**Need to develop locally?**
→ Use `.env.local.example`

**Need advanced features?**
→ Start with `env.yaml.example`, customize

**Need to convert formats?**
→ Use `./scripts/convert-env-to-yaml.sh`

---

## See Also

- [CONFIGURATION-GUIDE.md](../docs/CONFIGURATION-GUIDE.md) - Variable validation and reference
- [DEPLOYMENT.md](../docs/DEPLOYMENT.md) - Complete deployment guide
- [LOCAL-TESTING.md](../docs/LOCAL-TESTING.md) - Local development guide

