# Deployment Summary

## Quick Start

To deploy the latest version of examples-copier to Google Cloud App Engine:

```bash
cd examples-copier

# 1. Set up environment file
cp env.yaml.example env.yaml
# Edit env.yaml with your values

# 2. Deploy
./deploy.sh
```

## What Was Created

### 1. Deployment Documentation

**DEPLOYMENT-GUIDE.md** (300+ lines)
- Complete deployment guide
- Prerequisites and setup
- Configuration instructions
- Deployment steps
- Post-deployment tasks
- Troubleshooting
- Security best practices
- Maintenance procedures

**DEPLOYMENT-CHECKLIST.md** (300+ lines)
- Step-by-step checklist
- Pre-deployment tasks
- Deployment commands
- Post-deployment verification
- Rollback procedures
- Quick command reference
- Success criteria

### 2. Deployment Script

**deploy.sh** (executable)
- Automated deployment script
- Prerequisites checking
- Build and test
- Deployment to App Engine
- Post-deployment verification
- Options:
  - `--project PROJECT_ID` - Set GCP project
  - `--version VERSION` - Set version name
  - `--no-promote` - Deploy without promoting
  - `--quiet` - Skip prompts
  - `--env-file FILE` - Custom env file path
  - `--help` - Show help

### 3. Configuration Files

**env.yaml.example**
- Example environment variables
- All required and optional settings
- Documentation for each variable
- Security notes

**.gitignore**
- Prevents committing secrets
- Excludes env.yaml, .env files
- Excludes private keys
- Standard Go ignores

## Deployment Steps

### Prerequisites

1. **Install Google Cloud SDK**
   ```bash
   brew install --cask google-cloud-sdk
   ```

2. **Authenticate**
   ```bash
   gcloud auth login
   gcloud auth application-default login
   ```

3. **Set Project**
   ```bash
   gcloud config set project YOUR_PROJECT_ID
   ```

4. **Enable APIs**
   ```bash
   gcloud services enable appengine.googleapis.com
   gcloud services enable secretmanager.googleapis.com
   gcloud services enable logging.googleapis.com
   ```

### Configuration

1. **Store GitHub Private Key**
   ```bash
   gcloud secrets create CODE_COPIER_PEM \
     --data-file=/path/to/private-key.pem \
     --replication-policy="automatic"
   
   PROJECT_ID=$(gcloud config get-value project)
   gcloud secrets add-iam-policy-binding CODE_COPIER_PEM \
     --member="serviceAccount:${PROJECT_ID}@appspot.gserviceaccount.com" \
     --role="roles/secretmanager.secretAccessor"
   ```

2. **Create env.yaml**
   ```bash
   cp env.yaml.example env.yaml
   # Edit with your values
   ```

   **Required values:**
   - `GITHUB_APP_ID`
   - `INSTALLATION_ID`
   - `REPO_NAME`
   - `REPO_OWNER`
   - `GITHUB_APP_PRIVATE_KEY_SECRET_NAME`
   - `WEBHOOK_SECRET`
   - `GOOGLE_PROJECT_ID`

### Deploy

**Option 1: Using script (recommended)**
```bash
./deploy.sh
```

**Option 2: Manual**
```bash
gcloud app deploy app.yaml --env-vars-file=env.yaml
```

**Option 3: Specific version**
```bash
./deploy.sh --version=v2 --no-promote
```

### Post-Deployment

1. **Get App URL**
   ```bash
   gcloud app describe --format="value(defaultHostname)"
   ```

2. **Update GitHub Webhook**
   - URL: `https://YOUR_PROJECT_ID.appspot.com/events`
   - Content type: `application/json`
   - Secret: (same as `WEBHOOK_SECRET`)
   - Events: Pull requests

3. **Verify**
   ```bash
   gcloud app logs tail -s default
   ```

## Current Deployment Configuration

### App Engine Settings (app.yaml)

```yaml
runtime: go
runtime_config:
  operating_system: "ubuntu22"
  runtime_version: "1.23"
env: flex
```

**Environment:** Flexible Environment  
**Runtime:** Go 1.23  
**OS:** Ubuntu 22.04  

### Default Configuration

**Port:** 8080  
**Webhook Path:** /events  
**Config File:** copier-config.yaml  
**Deprecation File:** deprecated_examples.json  
**Source Branch:** main  

## Monitoring

### View Logs

```bash
# Real-time logs
gcloud app logs tail -s default

# Recent logs
gcloud app logs read --limit=50

# Filter for errors
gcloud app logs read --limit=100 | grep ERROR

# Filter for webhooks
gcloud app logs read --limit=100 | grep webhook
```

### View Metrics

```bash
# Open Cloud Console
gcloud app open-console

# View in browser
# Go to: App Engine → Dashboard
```

## Rollback

If deployment has issues:

```bash
# List versions
gcloud app versions list

# Route traffic to previous version
gcloud app services set-traffic default --splits=PREVIOUS_VERSION=1

# Delete bad version
gcloud app versions delete BAD_VERSION
```

## Security Notes

### ⚠️ Important

1. **Never commit secrets**
   - env.yaml is in .gitignore
   - Never commit .env files
   - Never commit .pem files

2. **Use Secret Manager**
   - GitHub private key stored in Secret Manager
   - Not in environment variables
   - Not in code

3. **Rotate secrets regularly**
   - Update GitHub App private key
   - Update webhook secret
   - Update in both GitHub and env.yaml

4. **Minimum permissions**
   - App Engine service account has minimal IAM roles
   - Only Secret Manager Secret Accessor
   - Only Cloud Logging Writer

## Troubleshooting

### Deployment Fails

```bash
# Check APIs
gcloud services list --enabled

# Verify build
go build -o examples-copier .

# Check env.yaml
cat env.yaml
```

### Webhooks Not Received

```bash
# Verify URL
gcloud app describe --format="value(defaultHostname)"

# Check logs
gcloud app logs tail -s default | grep webhook

# Test webhook from GitHub
# Settings → Webhooks → Recent Deliveries → Redeliver
```

### Cannot Access Secrets

```bash
# List secrets
gcloud secrets list

# Check permissions
gcloud secrets get-iam-policy CODE_COPIER_PEM

# Grant access
PROJECT_ID=$(gcloud config get-value project)
gcloud secrets add-iam-policy-binding CODE_COPIER_PEM \
  --member="serviceAccount:${PROJECT_ID}@appspot.gserviceaccount.com" \
  --role="roles/secretmanager.secretAccessor"
```

## Files Created

```
examples-copier/
├── DEPLOYMENT-GUIDE.md          # Complete deployment guide
├── DEPLOYMENT-CHECKLIST.md      # Step-by-step checklist
├── DEPLOYMENT-SUMMARY.md        # This file
├── deploy.sh                    # Deployment script (executable)
├── env.yaml.example             # Example environment variables
├── .gitignore                   # Git ignore file (includes env.yaml)
└── app.yaml                     # App Engine configuration (existing)
```

## Next Steps

1. **Review Documentation**
   - Read DEPLOYMENT-GUIDE.md for details
   - Follow DEPLOYMENT-CHECKLIST.md step-by-step

2. **Prepare Environment**
   - Create env.yaml from example
   - Store GitHub private key in Secret Manager
   - Enable required APIs

3. **Deploy**
   - Run `./deploy.sh`
   - Update GitHub webhook
   - Test with a PR

4. **Monitor**
   - Watch logs for first few PRs
   - Verify files are copied correctly
   - Check for errors

## Support

**Documentation:**
- [DEPLOYMENT-GUIDE.md](DEPLOYMENT-GUIDE.md) - Detailed guide
- [DEPLOYMENT-CHECKLIST.md](DEPLOYMENT-CHECKLIST.md) - Checklist
- [README.md](README.md) - Application overview
- [CONFIG-LOADING-BEHAVIOR.md](CONFIG-LOADING-BEHAVIOR.md) - Config details

**Google Cloud:**
- [App Engine Documentation](https://cloud.google.com/appengine/docs/flexible/go)
- [Secret Manager Documentation](https://cloud.google.com/secret-manager/docs)
- [Cloud Logging Documentation](https://cloud.google.com/logging/docs)

**GitHub:**
- [GitHub Apps Documentation](https://docs.github.com/en/apps)
- [Webhooks Documentation](https://docs.github.com/en/webhooks)

---

**Ready to deploy?** Run `./deploy.sh` to get started!

