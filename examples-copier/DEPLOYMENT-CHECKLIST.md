g# Deployment Checklist

Quick checklist for deploying examples-copier to Google Cloud App Engine.

## Pre-Deployment

### ☐ 1. Install Prerequisites

```bash
# Install Google Cloud SDK (if not installed)
brew install --cask google-cloud-sdk

# Verify installation
gcloud --version
go version
```

### ☐ 2. Authenticate with Google Cloud

```bash
# Login
gcloud auth login

# Set application default credentials
gcloud auth application-default login
```

### ☐ 3. Set GCP Project

```bash
# List projects
gcloud projects list

# Set project
gcloud config set project YOUR_PROJECT_ID

# Verify
gcloud config get-value project
```

### ☐ 4. Enable Required APIs

```bash
# Enable App Engine
gcloud services enable appengine.googleapis.com

# Enable Secret Manager
gcloud services enable secretmanager.googleapis.com

# Enable Cloud Logging
gcloud services enable logging.googleapis.com

# Verify
gcloud services list --enabled | grep -E "appengine|secretmanager|logging"
```

### ☐ 5. Store GitHub Private Key in Secret Manager

```bash
# Create secret
gcloud secrets create CODE_COPIER_PEM \
  --data-file=/path/to/your/private-key.pem \
  --replication-policy="automatic"

# Grant App Engine access
PROJECT_ID=$(gcloud config get-value project)
gcloud secrets add-iam-policy-binding CODE_COPIER_PEM \
  --member="serviceAccount:${PROJECT_ID}@appspot.gserviceaccount.com" \
  --role="roles/secretmanager.secretAccessor"

# Get secret path (copy this for env.yaml)
echo "projects/${PROJECT_ID}/secrets/CODE_COPIER_PEM/versions/latest"
```

### ☐ 6. Create env.yaml

```bash
# Copy example
cp env.yaml.example env.yaml

# Edit with your values
nano env.yaml  # or use your preferred editor
```

**Required values to update:**
- `GITHUB_APP_ID`
- `INSTALLATION_ID`
- `REPO_NAME`
- `REPO_OWNER`
- `GITHUB_APP_PRIVATE_KEY_SECRET_NAME`
- `WEBHOOK_SECRET`
- `GOOGLE_PROJECT_ID`

### ☐ 7. Verify Configuration

```bash
# Check env.yaml exists
ls -l env.yaml

# Verify it's in .gitignore
grep "env.yaml" .gitignore
```

### ☐ 8. Build and Test

```bash
# Build
go build -o examples-copier .

# Run tests
go test ./...

# Test locally (optional)
go run app.go -env ./configs/.env.test
```

## Deployment

### ☐ 9. Deploy to App Engine

**Option A: Using deployment script (recommended)**

```bash
./deploy.sh
```

**Option B: Manual deployment**

```bash
gcloud app deploy app.yaml --env-vars-file=env.yaml
```

**Option C: Deploy specific version without promoting**

```bash
./deploy.sh --version=v2 --no-promote
```

### ☐ 10. Verify Deployment

```bash
# Check deployment status
gcloud app versions list

# Get app URL
gcloud app describe --format="value(defaultHostname)"

# View logs
gcloud app logs tail -s default
```

## Post-Deployment

### ☐ 11. Update GitHub Webhook

1. Go to source repository on GitHub
2. Settings → Webhooks
3. Edit existing webhook or create new one
4. Update Payload URL: `https://YOUR_PROJECT_ID.appspot.com/events`
5. Content type: `application/json`
6. Secret: (same as `WEBHOOK_SECRET` in env.yaml)
7. Events: Select "Pull requests"
8. Active: ✓ Checked
9. Save webhook

### ☐ 12. Test Webhook

```bash
# Method 1: Merge a test PR in source repository
# Watch logs for webhook receipt
gcloud app logs tail -s default

# Method 2: Send test webhook from GitHub
# Go to webhook settings → Recent Deliveries → Redeliver
```

### ☐ 13. Verify Application is Working

**Check logs for:**
- ✓ Webhook received
- ✓ Config file loaded
- ✓ Files copied to target repos
- ✓ No errors

```bash
# View recent logs
gcloud app logs read --limit=50

# Filter for errors
gcloud app logs read --limit=100 | grep ERROR

# Filter for webhook events
gcloud app logs read --limit=100 | grep webhook
```

### ☐ 14. Monitor Application

```bash
# Real-time logs
gcloud app logs tail -s default

# Open Cloud Console
gcloud app open-console

# View metrics
# Go to Cloud Console → App Engine → Dashboard
```

## Rollback (if needed)

### ☐ If Deployment Has Issues

```bash
# List versions
gcloud app versions list

# Route traffic to previous version
gcloud app services set-traffic default --splits=PREVIOUS_VERSION=1

# Delete bad version
gcloud app versions delete BAD_VERSION
```

## Troubleshooting

### Issue: Deployment fails

```bash
# Check APIs are enabled
gcloud services list --enabled

# Verify app.yaml is valid
cat app.yaml

# Check for build errors
go build -o examples-copier .
```

### Issue: Webhooks not received

```bash
# Check webhook URL is correct
gcloud app describe --format="value(defaultHostname)"

# Verify webhook secret matches
# Compare GitHub webhook secret with WEBHOOK_SECRET in env.yaml

# Check logs for errors
gcloud app logs tail -s default | grep webhook
```

### Issue: Cannot access secrets

```bash
# Verify secret exists
gcloud secrets list

# Check IAM permissions
gcloud secrets get-iam-policy CODE_COPIER_PEM

# Grant access if needed
PROJECT_ID=$(gcloud config get-value project)
gcloud secrets add-iam-policy-binding CODE_COPIER_PEM \
  --member="serviceAccount:${PROJECT_ID}@appspot.gserviceaccount.com" \
  --role="roles/secretmanager.secretAccessor"
```

### Issue: Files not being copied

```bash
# Check config file in source repo
# Verify copier-config.yaml exists and is valid

# Check logs for pattern matching
gcloud app logs read --limit=100 | grep "file matched"

# Verify GitHub App has write access to target repos
```

## Quick Commands Reference

```bash
# Deploy
./deploy.sh

# Deploy specific version
./deploy.sh --version=v2

# Deploy without promoting
./deploy.sh --version=v2 --no-promote

# View logs
gcloud app logs tail -s default

# List versions
gcloud app versions list

# Set traffic split
gcloud app services set-traffic default --splits=v1=0.5,v2=0.5

# Delete version
gcloud app versions delete VERSION_ID

# Open Cloud Console
gcloud app open-console

# Get app URL
gcloud app describe --format="value(defaultHostname)"
```

## Environment Variables Quick Reference

**Required:**
- `GITHUB_APP_ID` - GitHub App ID
- `INSTALLATION_ID` - Installation ID
- `REPO_NAME` - Source repo name
- `REPO_OWNER` - Source repo owner
- `GITHUB_APP_PRIVATE_KEY_SECRET_NAME` - Secret Manager path
- `WEBHOOK_SECRET` - Webhook secret

**Optional:**
- `PORT` - Server port (default: 8080)
- `WEBSERVER_PATH` - Webhook path (default: /webhook)
- `CONFIG_FILE` - Config file (default: copier-config.yaml)
- `DEPRECATION_FILE` - Deprecation file (default: deprecated_examples.json)
- `COMMITTER_NAME` - Git committer name
- `COMMITTER_EMAIL` - Git committer email
- `GOOGLE_PROJECT_ID` - GCP project for logging
- `GOOGLE_LOG_NAME` - Log name

## Success Criteria

✅ Deployment completes without errors  
✅ Application is accessible at App Engine URL  
✅ Webhook receives PR events from GitHub  
✅ Config file is loaded successfully  
✅ Files are copied to target repositories  
✅ Logs show no errors  
✅ Deprecation tracking works (if enabled)  

## Next Steps After Deployment

1. Monitor logs for first few PRs
2. Verify files are copied correctly
3. Check target repositories for commits
4. Set up monitoring/alerting (optional)
5. Document any custom configuration
6. Share webhook URL with team

---

**See Also:**
- [DEPLOYMENT-GUIDE.md](DEPLOYMENT-GUIDE.md) - Detailed deployment guide
- [README.md](README.md) - Application overview
- [CONFIG-LOADING-BEHAVIOR.md](CONFIG-LOADING-BEHAVIOR.md) - Config details

