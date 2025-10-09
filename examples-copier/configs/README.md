# Environment Files Comparison

Overview of the different environment configuration files and when to use each.

## Files Overview

| File                  | Purpose                               | Use Case                        |
|-----------------------|---------------------------------------|---------------------------------|
| `env.yaml.example`    | Complete reference with all variables | First-time setup, documentation |
| `env.yaml.production` | Production-ready template             | Quick deployment to production  |
| `.env.example`        | Local development template            | Local testing and development   |

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

## .env.example.new

**Location:** `configs/.env.example.new`

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

**Recommended:** `.env.example.new`

```bash
# Local development
cp configs/.env.example.new configs/.env
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
- **Use `.env.example.new` for local development**
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
│   └── .env.example              # ← Local development template
├── env.yaml                      # ← Your actual config (gitignored)
└── .env                          # ← Your local config (gitignored)
```

---

## Quick Reference

**Need to deploy quickly?**
→ Use `env.yaml.production`

**Need to understand all options?**
→ Read `env.yaml.example`

**Need to develop locally?**
→ Use `.env.example.new`

**Need advanced features?**
→ Start with `env.yaml.example`, customize

**Need to convert formats?**
→ Use `./scripts/convert-env-to-yaml.sh`

---

## See Also

- [CONFIGURATION-GUIDE.md](../docs/CONFIGURATION-GUIDE.md) - Variable validation and reference
- [DEPLOYMENT.md](../docs/DEPLOYMENT.md) - Complete deployment guide
- [DEPLOYMENT-CHECKLIST.md](../docs/DEPLOYMENT-CHECKLIST.md) - Step-by-step checklist
- [LOCAL-TESTING.md](../docs/LOCAL-TESTING.md) - Local development guide

