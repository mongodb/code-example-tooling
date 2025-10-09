# Deployment Guide - Google Cloud App Engine

## Overview

This guide covers deploying the examples-copier application to Google Cloud App Engine (Flexible Environment).

## Prerequisites

### 1. Google Cloud SDK

Install the Google Cloud SDK if not already installed:

```bash
# macOS (using Homebrew)
brew install --cask google-cloud-sdk

# Or download from: https://cloud.google.com/sdk/docs/install
```

### 2. Authentication

Authenticate with Google Cloud:

```bash
# Login to your Google account
gcloud auth login

# Set application default credentials
gcloud auth application-default login
```

### 3. Project Configuration

Set your Google Cloud project:

```bash
# List available projects
gcloud projects list

# Set the project (replace with your project ID)
gcloud config set project YOUR_PROJECT_ID
```

### 4. Required APIs

Enable required Google Cloud APIs:

```bash
# Enable App Engine Admin API
gcloud services enable appengine.googleapis.com

# Enable Secret Manager API (for GitHub private key)
gcloud services enable secretmanager.googleapis.com

# Enable Cloud Logging API
gcloud services enable logging.googleapis.com
```

## Configuration

### 1. Environment Variables

The application uses environment variables for configuration. These are **NOT** stored in `app.yaml` for security reasons.

**Required Environment Variables:**
- `GITHUB_APP_ID` - GitHub App ID (numeric)
- `INSTALLATION_ID` - GitHub App Installation ID
- `REPO_NAME` - Source repository name
- `REPO_OWNER` - Source repository owner
- `GITHUB_APP_PRIVATE_KEY_SECRET_NAME` - GCP Secret Manager path to private key
- `WEBHOOK_SECRET` - GitHub webhook secret for signature validation

**Optional Environment Variables:**
- `PORT` - Server port (default: 8080)
- `WEBSERVER_PATH` - Webhook endpoint path (default: /webhook)
- `CONFIG_FILE` - Config file name (default: copier-config.yaml)
- `DEPRECATION_FILE` - Deprecation file name (default: deprecated_examples.json)
- `COMMITTER_NAME` - Git committer name (default: Copier Bot)
- `COMMITTER_EMAIL` - Git committer email (default: bot@example.com)
- `GOOGLE_PROJECT_ID` - GCP project ID for logging
- `GOOGLE_LOG_NAME` - Cloud Logging log name

### 2. Store GitHub Private Key in Secret Manager

The GitHub App private key must be stored in Google Cloud Secret Manager:

```bash
# Create secret from file
gcloud secrets create CODE_COPIER_PEM \
  --data-file=/path/to/your/private-key.pem \
  --replication-policy="automatic"

# Grant App Engine access to the secret
gcloud secrets add-iam-policy-binding CODE_COPIER_PEM \
  --member="serviceAccount:YOUR_PROJECT_ID@appspot.gserviceaccount.com" \
  --role="roles/secretmanager.secretAccessor"

# Get the full secret path (use this in GITHUB_APP_PRIVATE_KEY_SECRET_NAME)
echo "projects/$(gcloud config get-value project)/secrets/CODE_COPIER_PEM/versions/latest"
```

### 3. Set Environment Variables in App Engine

Environment variables are set during deployment using the `--env-vars-file` flag.

Create `env.yaml`:

```yaml
env_variables:
  GITHUB_APP_ID: "1166559"
  INSTALLATION_ID: "62138132"
  REPO_NAME: "docs-code-examples"
  REPO_OWNER: "mongodb"
  GITHUB_APP_PRIVATE_KEY_SECRET_NAME: "projects/YOUR_PROJECT_ID/secrets/CODE_COPIER_PEM/versions/latest"
  WEBHOOK_SECRET: "your-webhook-secret-here"
  COMMITTER_NAME: "GitHub Copier App"
  COMMITTER_EMAIL: "bot@mongodb.com"
  CONFIG_FILE: "copier-config.yaml"
  DEPRECATION_FILE: "deprecated_examples.json"
  WEBSERVER_PATH: "/events"
  GOOGLE_PROJECT_ID: "YOUR_PROJECT_ID"
  GOOGLE_LOG_NAME: "code-copier-log"
```

**⚠️ IMPORTANT:** Add `env.yaml` to `.gitignore` to prevent committing secrets!

## Deployment

### Quick Deployment

Use the provided deployment script:

```bash
cd examples-copier
./deploy.sh
```

### Manual Deployment

#### Step 1: Build and Test Locally

```bash
cd examples-copier

# Build the application
go build -o examples-copier .

# Run tests
go test ./...

# Test locally (optional)
go run app.go -env ./configs/.env.test
```

#### Step 2: Deploy to App Engine

```bash
# Deploy with environment variables
gcloud app deploy app.yaml --env-vars-file=env.yaml

# Or deploy without prompts
gcloud app deploy app.yaml --env-vars-file=env.yaml --quiet
```

#### Step 3: Verify Deployment

```bash
# View deployment status
gcloud app versions list

# View logs
gcloud app logs tail -s default

# Open the application in browser
gcloud app browse
```

### Deployment Options

**Deploy specific version:**
```bash
gcloud app deploy app.yaml --version=v1 --env-vars-file=env.yaml
```

**Deploy without promoting (traffic stays on current version):**
```bash
gcloud app deploy app.yaml --no-promote --env-vars-file=env.yaml
```

**Deploy and set traffic split:**
```bash
# Deploy new version
gcloud app deploy app.yaml --version=v2 --no-promote --env-vars-file=env.yaml

# Split traffic (50% to v1, 50% to v2)
gcloud app services set-traffic default --splits=v1=0.5,v2=0.5
```

## Post-Deployment

### 1. Update GitHub Webhook URL

After deployment, update the webhook URL in your source repository:

```
https://YOUR_PROJECT_ID.appspot.com/events
```

**Steps:**
1. Go to your source repository on GitHub
2. Settings → Webhooks
3. Edit the webhook
4. Update the Payload URL to your App Engine URL
5. Ensure Content type is `application/json`
6. Set the webhook secret (same as `WEBHOOK_SECRET` env var)
7. Select "Pull requests" event
8. Save webhook

### 2. Verify Webhook

Test the webhook:

```bash
# Trigger a test PR merge in your source repo
# Check App Engine logs for webhook receipt

gcloud app logs tail -s default
```

### 3. Monitor Application

**View logs:**
```bash
# Tail logs in real-time
gcloud app logs tail -s default

# View logs in Cloud Console
gcloud app logs read --limit=50
```

**View metrics:**
```bash
# Open Cloud Console
gcloud app open-console
```

## Troubleshooting

### Issue: Deployment Fails

**Check:**
- All required APIs are enabled
- `app.yaml` is valid
- Go version matches runtime version
- Dependencies are up to date

**Solution:**
```bash
# Update dependencies
go mod tidy

# Verify app.yaml
cat app.yaml

# Check enabled APIs
gcloud services list --enabled
```

### Issue: Application Not Receiving Webhooks

**Check:**
- Webhook URL is correct
- Webhook secret matches `WEBHOOK_SECRET` env var
- GitHub App has correct permissions
- Firewall rules allow GitHub IPs

**Solution:**
```bash
# Check logs for webhook errors
gcloud app logs tail -s default | grep webhook

# Test webhook manually using curl
curl -X POST https://YOUR_PROJECT_ID.appspot.com/events \
  -H "Content-Type: application/json" \
  -d '{"test": "data"}'
```

### Issue: Cannot Access Secret Manager

**Check:**
- Secret exists in Secret Manager
- App Engine service account has access
- Secret path is correct in env vars

**Solution:**
```bash
# List secrets
gcloud secrets list

# Check IAM permissions
gcloud secrets get-iam-policy CODE_COPIER_PEM

# Grant access if needed
gcloud secrets add-iam-policy-binding CODE_COPIER_PEM \
  --member="serviceAccount:YOUR_PROJECT_ID@appspot.gserviceaccount.com" \
  --role="roles/secretmanager.secretAccessor"
```

### Issue: High Memory Usage

**Check:**
- App Engine instance size
- Memory leaks in code
- Number of concurrent requests

**Solution:**
Update `app.yaml`:
```yaml
runtime: go
runtime_config:
  operating_system: "ubuntu22"
  runtime_version: "1.23"
env: flex

resources:
  cpu: 1
  memory_gb: 2
  disk_size_gb: 10

automatic_scaling:
  min_num_instances: 1
  max_num_instances: 5
  cool_down_period_sec: 120
  cpu_utilization:
    target_utilization: 0.6
```

## Rollback

If deployment fails or has issues, rollback to previous version:

```bash
# List versions
gcloud app versions list

# Set traffic to previous version
gcloud app services set-traffic default --splits=PREVIOUS_VERSION=1

# Delete bad version (optional)
gcloud app versions delete BAD_VERSION
```

## Cost Optimization

### 1. Use Minimum Instances

For low-traffic applications:

```yaml
automatic_scaling:
  min_num_instances: 1
  max_num_instances: 2
```

### 2. Use Standard Environment (if possible)

App Engine Standard is cheaper than Flexible, but requires code changes.

### 3. Monitor Costs

```bash
# View billing
gcloud billing accounts list

# Set budget alerts in Cloud Console
```

## Security Best Practices

### 1. Never Commit Secrets

Add to `.gitignore`:
```
env.yaml
*.pem
.env.production
```

### 2. Rotate Secrets Regularly

```bash
# Create new secret version
gcloud secrets versions add CODE_COPIER_PEM --data-file=/path/to/new-key.pem

# Update env var to use new version
# Redeploy application
```

### 3. Use IAM Roles

Grant minimum required permissions:
```bash
# App Engine service account should have:
# - Secret Manager Secret Accessor
# - Cloud Logging Writer
```

### 4. Enable VPC Service Controls (optional)

For additional security in production.

## Maintenance

### Update Application

```bash
# Pull latest code
git pull origin main

# Build and test
go build -o examples-copier .
go test ./...

# Deploy
gcloud app deploy app.yaml --env-vars-file=env.yaml
```

### Update Dependencies

```bash
# Update Go modules
go get -u ./...
go mod tidy

# Test
go test ./...

# Deploy
gcloud app deploy app.yaml --env-vars-file=env.yaml
```

### View Application Info

```bash
# App Engine info
gcloud app describe

# List services
gcloud app services list

# List versions
gcloud app versions list

# View instance info
gcloud app instances list
```

## Additional Resources

- [App Engine Go Documentation](https://cloud.google.com/appengine/docs/flexible/go)
- [Secret Manager Documentation](https://cloud.google.com/secret-manager/docs)
- [Cloud Logging Documentation](https://cloud.google.com/logging/docs)
- [GitHub Apps Documentation](https://docs.github.com/en/apps)

---

**See Also:**
- [README.md](README.md) - Application overview
- [Configuration Guide](CONFIG-LOADING-BEHAVIOR.md) - Config file details
- [Troubleshooting](TROUBLESHOOTING.md) - Common issues

