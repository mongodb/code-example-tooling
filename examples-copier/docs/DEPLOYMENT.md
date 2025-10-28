# Deployment Guide

Complete guide for deploying the GitHub Code Example Copier to Google Cloud App Engine with Secret Manager.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Architecture Overview](#architecture-overview)
- [Secret Manager Setup](#secret-manager-setup)
- [Configuration](#configuration)
- [Deployment](#deployment)
- [Post-Deployment](#post-deployment)
- [Monitoring](#monitoring)
- [Troubleshooting](#troubleshooting)
- [Deployment Checklist](#deployment-checklist)

## Prerequisites

### Required Tools

- **Go 1.23+** - For local development and testing
- **Google Cloud SDK** - For deployment
- **GitHub App** - With appropriate permissions
- **MongoDB Atlas** (optional) - For audit logging

### Required Accounts & Access

- Google Cloud project with billing enabled
- GitHub organization admin access (to create/configure GitHub App)
- MongoDB Atlas account (if using audit logging)

### Install Google Cloud SDK

```bash
# macOS
brew install --cask google-cloud-sdk

# Verify installation
gcloud --version
```

### Authenticate with Google Cloud

```bash
# Login to Google Cloud
gcloud auth login

# Set application default credentials
gcloud auth application-default login

# Set your project
gcloud config set project YOUR_PROJECT_ID

# Verify
gcloud config get-value project
```

## Architecture Overview

### Components

```
┌─────────────────────────────────────────────────────────────┐
│                     GitHub Repository                       │
│                  (docs-code-examples)                       │
└────────────────────┬────────────────────────────────────────┘
                     │ Webhook (PR merged)
                     ↓
┌─────────────────────────────────────────────────────────────┐
│              Google Cloud App Engine                        │
│  ┌──────────────────────────────────────────────────────┐   │
│  │  examples-copier Application                         │   │
│  │  - Receives webhook                                  │   │
│  │  - Validates signature                               │   │
│  │  - Loads config from source repo                     │   │
│  │  - Matches files against patterns                    │   │
│  │  - Copies to target repos                            │   │
│  └──────────────────────────────────────────────────────┘   │
└────────────┬────────────────────────────┬───────────────────┘
             │                            │
             ↓                            ↓
┌────────────────────────┐   ┌───────────────────────────────┐
│  Secret Manager        │   │  Cloud Logging                │
│  - GitHub private key  │   │  - Application logs           │
│  - Webhook secret      │   │  - Webhook events             │
│  - MongoDB URI         │   │  - Error tracking             │
└────────────────────────┘   └───────────────────────────────┘
             │
             ↓
┌────────────────────────┐
│  MongoDB Atlas         │
│  - Audit events        │
│  - Metrics             │
└────────────────────────┘
```

### Environment: Flexible vs Standard

This application uses **App Engine Flexible Environment**:

**app.yaml:**
```yaml
runtime: go
runtime_config:
  operating_system: "ubuntu22"
  runtime_version: "1.23"
env: flex  # ← Flexible Environment
```

**Key differences:**
- Environment variables in **separate file** (`env.yaml`) included via `includes` directive
- Deployment: `gcloud app deploy app.yaml`
- Better Secret Manager integration
- More flexible runtime configuration

## Secret Manager Setup

### Why Secret Manager?

- **Security**: Secrets encrypted at rest and in transit  
- **Audit Trail**: All access logged  
- **Rotation**: Update secrets without redeployment  
- **Access Control**: Fine-grained IAM permissions  
- **No Hardcoding**: Secrets never in config files or version control  

### Enable Secret Manager API

```bash
gcloud services enable secretmanager.googleapis.com
```

### Store Secrets

#### 1. GitHub App Private Key

```bash
# Store your GitHub App private key
gcloud secrets create CODE_COPIER_PEM \
  --data-file=/path/to/your/private-key.pem \
  --replication-policy="automatic"
```

#### 2. Webhook Secret

```bash
# Generate a secure webhook secret
WEBHOOK_SECRET=$(openssl rand -hex 32)
echo "Generated: $WEBHOOK_SECRET"

# Store in Secret Manager
echo -n "$WEBHOOK_SECRET" | gcloud secrets create webhook-secret \
  --data-file=- \
  --replication-policy="automatic"

# Save this value - you'll need it for GitHub webhook configuration
```

#### 3. MongoDB URI (Optional - for audit logging)

```bash
# Store MongoDB connection string
echo -n "mongodb+srv://user:pass@cluster.mongodb.net/dbname" | \
  gcloud secrets create mongo-uri \
  --data-file=- \
  --replication-policy="automatic"
```

### Grant App Engine Access

```bash
# Get your project number
PROJECT_NUMBER=$(gcloud projects describe $(gcloud config get-value project) --format="value(projectNumber)")

# App Engine service account
SERVICE_ACCOUNT="${PROJECT_NUMBER}@appspot.gserviceaccount.com"

# Grant access to each secret
gcloud secrets add-iam-policy-binding CODE_COPIER_PEM \
  --member="serviceAccount:${SERVICE_ACCOUNT}" \
  --role="roles/secretmanager.secretAccessor"

gcloud secrets add-iam-policy-binding webhook-secret \
  --member="serviceAccount:${SERVICE_ACCOUNT}" \
  --role="roles/secretmanager.secretAccessor"

gcloud secrets add-iam-policy-binding mongo-uri \
  --member="serviceAccount:${SERVICE_ACCOUNT}" \
  --role="roles/secretmanager.secretAccessor"
```

**Or use the provided script:**
```bash
cd examples-copier
./scripts/grant-secret-access.sh
```

### Verify Secrets

```bash
# List all secrets
gcloud secrets list

# View secret metadata
gcloud secrets describe CODE_COPIER_PEM

# Verify IAM permissions
gcloud secrets get-iam-policy CODE_COPIER_PEM
```

## Configuration

### Create env.yaml

The `env.yaml` file contains environment variables for App Engine deployment.

```bash
cd examples-copier

# Copy from production template
cp configs/env.yaml.production env.yaml

# Or convert from .env file
./scripts/convert-env-to-yaml.sh configs/.env env.yaml

# Edit if needed
nano env.yaml
```

### env.yaml Structure

**Important Notes:**
- Do NOT set `PORT` in `env.yaml` - App Engine Flexible automatically sets this
- The application defaults to port 8080 for local development
- Secret Manager references must include `/versions/latest` or a specific version number

```yaml
env_variables:
  # =============================================================================
  # GitHub Configuration (Non-sensitive)
  # =============================================================================
  GITHUB_APP_ID: "YOUR_APP_ID"
  INSTALLATION_ID: "YOUR_INSTALLATION_ID"
  REPO_OWNER: "your-org"
  REPO_NAME: "your-repo"
  SRC_BRANCH: "main"
  
  # =============================================================================
  # Secret Manager References (Sensitive - SECURE!)
  # =============================================================================
  GITHUB_APP_PRIVATE_KEY_SECRET_NAME: "projects/PROJECT_NUMBER/secrets/CODE_COPIER_PEM/versions/latest"
  WEBHOOK_SECRET_NAME: "projects/PROJECT_NUMBER/secrets/webhook-secret/versions/latest"
  MONGO_URI_SECRET_NAME: "projects/PROJECT_NUMBER/secrets/mongo-uri/versions/latest"
  
  # =============================================================================
  # Application Settings
  # =============================================================================
  # PORT: "8080"                                   # DO NOT SET - App Engine sets this automatically
  WEBSERVER_PATH: "/events"
  CONFIG_FILE: "copier-config.yaml"
  DEPRECATION_FILE: "deprecated_examples.json"
  
  # =============================================================================
  # Committer Information
  # =============================================================================
  COMMITTER_NAME: "GitHub Copier App"
  COMMITTER_EMAIL: "bot@example.com"
  
  # =============================================================================
  # Google Cloud Configuration
  # =============================================================================
  GOOGLE_PROJECT_ID: "your-project-id"
  GOOGLE_LOG_NAME: "code-copier-log"
  
  # =============================================================================
  # Feature Flags
  # =============================================================================
  AUDIT_ENABLED: "true"
  METRICS_ENABLED: "true"
  # DRY_RUN: "false"
```

### Important Notes

**✅ DO:**
- Use Secret Manager references (`*_SECRET_NAME` variables)
- Keep `env.yaml` in `.gitignore`
- Use `env.yaml.production` as template

**❌ DON'T:**
- Put actual secrets in `env.yaml` (use `*_SECRET_NAME` instead)
- Commit `env.yaml` to version control
- Share `env.yaml` via email/chat

### How Secrets Are Loaded

```
Application Startup:
1. Load env.yaml → environment variables
2. Read WEBHOOK_SECRET_NAME from env
3. Call Secret Manager API to get actual secret
4. Store in config.WebhookSecret
5. Use for webhook signature validation
```

**Code flow:**
```go
// app.go
config, _ := configs.LoadEnvironment(envFile)
services.LoadWebhookSecret(config)  // Loads from Secret Manager
services.LoadMongoURI(config)       // Loads from Secret Manager
```

## Deployment

### Pre-Deployment Checklist

- [ ] Secrets created in Secret Manager
- [ ] IAM permissions granted to App Engine
- [ ] `env.yaml` created and configured
- [ ] `env.yaml` in `.gitignore`
- [ ] `app.yaml` uses Flexible Environment

### Deploy to App Engine

```bash
cd examples-copier

# Deploy (env.yaml is included via 'includes' directive in app.yaml)
gcloud app deploy app.yaml

# Or specify project
gcloud app deploy app.yaml --project=your-project-id
```

### Verify Deployment

```bash
# Check deployment status
gcloud app versions list

# Get app URL
APP_URL=$(gcloud app describe --format="value(defaultHostname)")
echo "App URL: https://${APP_URL}"

# View logs
gcloud app logs tail -s default
```

### Test Health Endpoint

```bash
# Test health
curl https://${APP_URL}/health

# Expected response:
# {
#   "status": "healthy",
#   "started": true,
#   "github": {
#     "status": "healthy",
#     "authenticated": true
#   },
#   "queues": {
#     "upload_count": 0,
#     "deprecation_count": 0
#   },
#   "uptime": "5m30s"
# }
```

## Post-Deployment

### Configure GitHub Webhook

1. **Navigate to repository settings**
   - Go to: `https://github.com/YOUR_ORG/YOUR_REPO/settings/hooks`

2. **Add or edit webhook**
   - **Payload URL:** `https://YOUR_APP.appspot.com/events`
   - **Content type:** `application/json`
   - **Secret:** (the webhook secret from Secret Manager)
   - **Events:** Select "Pull requests"
   - **Active:** ✓ Checked

3. **Get webhook secret from Secret Manager**
   ```bash
   gcloud secrets versions access latest --secret=webhook-secret
   ```

4. **Save webhook**

### Test Webhook

**Option A: Merge a test PR**
```bash
# Create and merge a test PR
# Watch logs for webhook receipt
gcloud app logs tail -s default | grep webhook
```

**Option B: Redeliver from GitHub**
1. Go to webhook settings
2. Click "Recent Deliveries"
3. Click on a delivery
4. Click "Redeliver"
5. Watch logs

### Verify Functionality

```bash
# Check logs for successful processing
gcloud app logs read --limit=50

# Look for:
# ✅ "Starting web server on port :8080"
# ✅ "webhook received"
# ✅ "Config file loaded successfully"
# ✅ "file matched pattern"
# ✅ "Copied file to target repo"

# Should NOT see:
# ❌ "failed to load webhook secret"
# ❌ "failed to load MongoDB URI"
# ❌ "webhook signature verification failed"
```

## Monitoring

### View Logs

```bash
# Real-time logs
gcloud app logs tail -s default

# Recent logs
gcloud app logs read --limit=100

# Filter for errors
gcloud app logs read --limit=100 | grep ERROR

# Filter for webhooks
gcloud app logs read --limit=100 | grep webhook
```

### Check Metrics

```bash
# Metrics endpoint
curl https://YOUR_APP.appspot.com/metrics

# Response includes:
# - webhooks_received
# - webhooks_processed
# - files_matched
# - files_uploaded
# - processing_time (p50, p95, p99)
```

### Audit Logging (if enabled)

Query MongoDB for audit events:

```javascript
// Connect to MongoDB
mongosh "mongodb+srv://..."

// Recent events
db.audit_events.find().sort({timestamp: -1}).limit(10)

// Failed operations
db.audit_events.find({success: false})

// Statistics by rule
db.audit_events.aggregate([
  {$match: {event_type: "copy"}},
  {$group: {
    _id: "$rule_name",
    count: {$sum: 1}
  }}
])
```

--- 

# Deployment Checklist

Quick reference checklist for deploying the GitHub Code Example Copier to Google Cloud App Engine.

## 📋 Pre-Deployment

### ☐ 1. Prerequisites Installed

```bash
# Verify Go
go version  # Should be 1.23+

# Verify gcloud
gcloud --version

# Verify authentication
gcloud auth list
```

### ☐ 2. Google Cloud Project Setup

```bash
# Set project
gcloud config set project YOUR_PROJECT_ID

# Verify
gcloud config get-value project

# Enable required APIs
gcloud services enable secretmanager.googleapis.com
gcloud services enable appengine.googleapis.com
```

### ☐ 3. Secrets in Secret Manager

```bash
# List secrets
gcloud secrets list

# Expected secrets:
# ✅ CODE_COPIER_PEM  - GitHub App private key
# ✅ webhook-secret   - Webhook signature validation
# ✅ mongo-uri        - MongoDB connection (optional)
```

**If secrets don't exist, create them:**

```bash
# GitHub private key
gcloud secrets create CODE_COPIER_PEM \
  --data-file=/path/to/private-key.pem \
  --replication-policy="automatic"

# Webhook secret
WEBHOOK_SECRET=$(openssl rand -hex 32)
echo -n "$WEBHOOK_SECRET" | gcloud secrets create webhook-secret \
  --data-file=- \
  --replication-policy="automatic"
echo "Save this: $WEBHOOK_SECRET"

# MongoDB URI (optional)
echo -n "mongodb+srv://..." | gcloud secrets create mongo-uri \
  --data-file=- \
  --replication-policy="automatic"
```

### ☐ 4. Grant IAM Permissions

```bash
# Run the grant script
cd examples-copier
./scripts/grant-secret-access.sh
```

**Or manually:**

```bash
PROJECT_NUMBER=$(gcloud projects describe $(gcloud config get-value project) --format="value(projectNumber)")
SERVICE_ACCOUNT="${PROJECT_NUMBER}@appspot.gserviceaccount.com"

gcloud secrets add-iam-policy-binding CODE_COPIER_PEM \
  --member="serviceAccount:${SERVICE_ACCOUNT}" \
  --role="roles/secretmanager.secretAccessor"

gcloud secrets add-iam-policy-binding webhook-secret \
  --member="serviceAccount:${SERVICE_ACCOUNT}" \
  --role="roles/secretmanager.secretAccessor"

gcloud secrets add-iam-policy-binding mongo-uri \
  --member="serviceAccount:${SERVICE_ACCOUNT}" \
  --role="roles/secretmanager.secretAccessor"
```

**Verify:**
```bash
gcloud secrets get-iam-policy CODE_COPIER_PEM | grep @appspot
gcloud secrets get-iam-policy webhook-secret | grep @appspot
gcloud secrets get-iam-policy mongo-uri | grep @appspot
```

### ☐ 5. Create env.yaml

```bash
cd examples-copier

# Copy from template
cp configs/env.yaml.production env.yaml

# Or convert from .env
./scripts/convert-env-to-yaml.sh configs/.env env.yaml

# Edit with your values if needed
nano env.yaml
```

**Required changes in env.yaml:**
- `GITHUB_APP_ID` - Your GitHub App ID
- `INSTALLATION_ID` - Your installation ID
- `REPO_OWNER` - Source repository owner
- `REPO_NAME` - Source repository name
- `GITHUB_APP_PRIVATE_KEY_SECRET_NAME` - Update project number
- `WEBHOOK_SECRET_NAME` - Update project number
- `MONGO_URI_SECRET_NAME` - Update project number (if using audit logging)
- `GOOGLE_PROJECT_ID` - Your Google Cloud project ID

### ☐ 6. Verify env.yaml in .gitignore

```bash
# Check
grep "env.yaml" .gitignore

# If not found, add it
echo "env.yaml" >> .gitignore
```

### ☐ 7. Verify app.yaml Configuration

```bash
cat app.yaml
```

**Should contain:**
```yaml
runtime: go
runtime_config:
  operating_system: "ubuntu22"
  runtime_version: "1.23"
env: flex
```

**Should NOT contain:**
- ❌ `env_variables:` section (those go in env.yaml)

---

## 🚀 Deployment

### ☐ 8. Deploy to App Engine

```bash
cd examples-copier

# Deploy (env.yaml is included via 'includes' directive in app.yaml)
gcloud app deploy app.yaml
```

**Expected output:**
```
Updating service [default]...done.
Setting traffic split for service [default]...done.
Deployed service [default] to [https://YOUR_APP.appspot.com]
```

### ☐ 9. Verify Deployment

```bash
# Check versions
gcloud app versions list

# Get app URL
APP_URL=$(gcloud app describe --format="value(defaultHostname)")
echo "App URL: https://${APP_URL}"
```

### ☐ 10. Check Logs

```bash
# View real-time logs
gcloud app logs tail -s default
```

**Look for:**
- ✅ "Starting web server on port :8080"
- ✅ No errors about secrets
- ✅ No "failed to load webhook secret"
- ✅ No "failed to load MongoDB URI"

**Should NOT see:**
- ❌ "failed to load webhook secret"
- ❌ "failed to load MongoDB URI"
- ❌ "SKIP_SECRET_MANAGER=true"

### ☐ 11. Test Health Endpoint

```bash
# Get app URL
APP_URL=$(gcloud app describe --format="value(defaultHostname)")

# Test health
curl https://${APP_URL}/health
```

**Expected response:**
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
  "uptime": "1m23s"
}
```

---

## 🔗 GitHub Webhook Configuration

### ☐ 12. Get Webhook Secret

```bash
# Get the webhook secret value
gcloud secrets versions access latest --secret=webhook-secret
```

**Save this value** - you'll need it for GitHub webhook configuration.

### ☐ 13. Configure GitHub Webhook

1. **Go to repository settings**
    - URL: `https://github.com/YOUR_ORG/YOUR_REPO/settings/hooks`

2. **Add or edit webhook**
    - **Payload URL:** `https://YOUR_APP.appspot.com/events`
    - **Content type:** `application/json`
    - **Secret:** (paste the value from step 12)
    - **SSL verification:** Enable SSL verification
    - **Events:** Select "Pull requests"
    - **Active:** ✓ Checked

3. **Save webhook**

### ☐ 14. Test Webhook

**Option A: Redeliver existing webhook**
1. Go to webhook settings
2. Click "Recent Deliveries"
3. Click on a delivery
4. Click "Redeliver"

**Option B: Create test PR**
1. Create a test PR in your source repository
2. Merge it
3. Watch logs for webhook receipt

```bash
# Watch logs
gcloud app logs tail -s default | grep webhook
```

---

## ✅ Post-Deployment Verification

### ☐ 15. Verify Secrets Loaded

```bash
# Check logs for secret loading
gcloud app logs read --limit=100 | grep -i "secret"
```

**Should NOT see:**
- ❌ "failed to load webhook secret"
- ❌ "failed to load MongoDB URI"

### ☐ 16. Verify Webhook Signature Validation

```bash
# Watch logs during webhook delivery
gcloud app logs tail -s default
```

**Look for:**
- ✅ "webhook received"
- ✅ "signature verified"
- ✅ "processing webhook"

**Should NOT see:**
- ❌ "webhook signature verification failed"
- ❌ "invalid signature"

### ☐ 17. Verify File Copying

```bash
# Watch logs during PR merge
gcloud app logs tail -s default
```

**Look for:**
- ✅ "Config file loaded successfully"
- ✅ "file matched pattern"
- ✅ "Copied file to target repo"

### ☐ 18. Verify Audit Logging (if enabled)

```bash
# Connect to MongoDB
mongosh "YOUR_MONGO_URI"

# Check for recent events
db.audit_events.find().sort({timestamp: -1}).limit(5)
```

### ☐ 19. Verify Metrics (if enabled)

```bash
# Check metrics endpoint
curl https://YOUR_APP.appspot.com/metrics
```

**Expected response:**
```json
{
  "webhooks": {
    "received": 1,
    "processed": 1,
    "failed": 0
  },
  "files": {
    "matched": 5,
    "uploaded": 5,
    "failed": 0
  }
}
```

### ☐ 20. Security Verification

```bash
# Verify env.yaml doesn't contain actual secrets
cat env.yaml | grep -E "BEGIN|mongodb\+srv|ghp_"
# Should return NOTHING (only Secret Manager paths)

# Verify env.yaml is not committed
git status | grep env.yaml
# Should show: nothing to commit (or untracked)

# Verify IAM permissions
gcloud secrets get-iam-policy CODE_COPIER_PEM | grep @appspot
gcloud secrets get-iam-policy webhook-secret | grep @appspot
# Should see the service account
```

---

## 🐛 Troubleshooting

### Error: "failed to load webhook secret"

**Cause:** Secret Manager access denied

**Fix:**
```bash
./scripts/grant-secret-access.sh
```

### Error: "webhook signature verification failed"

**Cause:** Secret in Secret Manager doesn't match GitHub webhook secret

**Fix:**
```bash
# Get secret from Secret Manager
gcloud secrets versions access latest --secret=webhook-secret

# Update GitHub webhook with this value
# OR update Secret Manager with GitHub's value
```

### Error: "MONGO_URI is required when audit logging is enabled"

**Cause:** Audit logging enabled but MongoDB URI not loaded

**Fix:**
```bash
# Option 1: Disable audit logging
# In env.yaml: AUDIT_ENABLED: "false"

# Option 2: Ensure MONGO_URI_SECRET_NAME is set
# In env.yaml: MONGO_URI_SECRET_NAME: "projects/.../secrets/mongo-uri/versions/latest"

# Redeploy
gcloud app deploy app.yaml
```

### Error: "Config file not found"

**Cause:** `copier-config.yaml` missing from source repository

**Fix:**
```bash
# Add copier-config.yaml to your source repository
# See documentation for config file format
```

---

## 📊 Success Criteria

All items should be ✅:

- ✅ Deployment completes without errors
- ✅ App Engine is running
- ✅ Health endpoint returns 200 OK
- ✅ Logs show no secret loading errors
- ✅ Webhook receives PR events
- ✅ Webhook signature validation works
- ✅ Files are copied to target repos
- ✅ Audit events logged (if enabled)
- ✅ Metrics available (if enabled)
- ✅ No secrets in config files
- ✅ env.yaml not in version control

---

## 🎉 You're Done!

Your application is deployed with:
- ✅ All secrets in Secret Manager (secure!)
- ✅ No hardcoded secrets in config files
- ✅ Easy secret rotation (just update in Secret Manager)
- ✅ Audit trail of secret access
- ✅ Fine-grained IAM permissions

**Next steps:**
1. Monitor logs for first few PRs
2. Verify files are copied correctly
3. Set up alerts (optional)
4. Document any custom configuration

---

## 📚 Quick Reference

```bash
# Deploy
gcloud app deploy app.yaml

# View logs
gcloud app logs tail -s default

# Check health
curl https://YOUR_APP.appspot.com/health

# Check metrics
curl https://YOUR_APP.appspot.com/metrics

# List secrets
gcloud secrets list

# Get secret value
gcloud secrets versions access latest --secret=SECRET_NAME

# Grant access
./scripts/grant-secret-access.sh

# Rollback
gcloud app versions list
gcloud app services set-traffic default --splits=PREVIOUS_VERSION=1
```
---

## Troubleshooting

### Common Issues

| Error | Cause | Solution |
|-------|-------|----------|
| "failed to load webhook secret" | Secret Manager access denied | Run `./grant-secret-access.sh` |
| "webhook signature verification failed" | Secret mismatch | Verify secret matches GitHub webhook |
| "MONGO_URI is required" | Audit enabled but no URI | Set `MONGO_URI_SECRET_NAME` or disable audit |
| "Config file not found" | Missing copier-config.yaml | Add config file to source repo |

### Quick Fixes

```bash
# Grant secret access
./scripts/grant-secret-access.sh

# View secret value
gcloud secrets versions access latest --secret=webhook-secret

# Disable audit logging
# In env.yaml: AUDIT_ENABLED: "false"

# Redeploy
gcloud app deploy app.yaml
```

## Next Steps

1. **Monitor first few PRs** - Watch logs to ensure files are copied correctly
2. **Set up alerts** (optional) - Configure Cloud Monitoring alerts
3. **Document custom config** - Add notes about your specific setup
4. **Plan secret rotation** - Schedule regular secret updates

---

**See also:**
- [FAQ.md](FAQ.md) - Frequently asked questions
- [../WEBHOOK-SECRET-MANAGER-GUIDE.md](../WEBHOOK-SECRET-MANAGER-GUIDE.md) - Secret Manager details
- [../ENV-FILES-EXPLAINED.md](../ENV-FILES-EXPLAINED.md) - Environment file explanation

