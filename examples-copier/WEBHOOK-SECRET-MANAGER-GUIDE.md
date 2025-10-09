# Using Webhook Secret from Google Cloud Secret Manager

## Overview

The examples-copier application now supports loading the webhook secret from Google Cloud Secret Manager instead of hardcoding it in environment variables. This is **more secure** and follows best practices for secret management.

## Benefits

✅ **Security**: Secrets are encrypted at rest and in transit  
✅ **Audit Trail**: Secret Manager logs all access to secrets  
✅ **Rotation**: Easy to rotate secrets without redeploying  
✅ **Access Control**: Fine-grained IAM permissions  
✅ **No Hardcoding**: Secrets never appear in config files or version control  

## Quick Start

### 1. Store Webhook Secret in Secret Manager

```bash
# Generate a secure random secret (if you don't have one)
WEBHOOK_SECRET=$(openssl rand -hex 32)
echo "Generated webhook secret: $WEBHOOK_SECRET"

# Store in Secret Manager
echo -n "$WEBHOOK_SECRET" | gcloud secrets create webhook-secret \
  --data-file=- \
  --replication-policy="automatic"

# Verify it was created
gcloud secrets describe webhook-secret
```

### 2. Grant App Engine Access

```bash
# Get your project ID
PROJECT_ID=$(gcloud config get-value project)

# Grant App Engine service account access to the secret
gcloud secrets add-iam-policy-binding webhook-secret \
  --member="serviceAccount:${PROJECT_ID}@appspot.gserviceaccount.com" \
  --role="roles/secretmanager.secretAccessor"

# Verify permissions
gcloud secrets get-iam-policy webhook-secret
```

### 3. Configure env.yaml

Use `WEBHOOK_SECRET_NAME` instead of `WEBHOOK_SECRET`:

```yaml
env_variables:
  # ... other config ...
  
  # Use Secret Manager (RECOMMENDED)
  WEBHOOK_SECRET_NAME: "projects/1054147886816/secrets/webhook-secret/versions/latest"
  
  # DO NOT use direct secret in production
  # WEBHOOK_SECRET: "hardcoded-secret-here"
```

### 4. Deploy

```bash
./deploy.sh
```

## Configuration Options

The application supports **two ways** to provide the webhook secret:

### Option 1: Secret Manager (Recommended for Production)

**Environment Variable:** `WEBHOOK_SECRET_NAME`

**Format:** `projects/PROJECT_ID/secrets/SECRET_NAME/versions/VERSION`

**Example:**
```yaml
WEBHOOK_SECRET_NAME: "projects/1054147886816/secrets/webhook-secret/versions/latest"
```

**Pros:**
- ✅ Secure - secret never in config files
- ✅ Auditable - all access logged
- ✅ Rotatable - update secret without redeploying
- ✅ Access controlled - IAM permissions

**Cons:**
- ❌ Requires Secret Manager setup
- ❌ Requires IAM permissions

### Option 2: Direct Environment Variable (For Testing Only)

**Environment Variable:** `WEBHOOK_SECRET`

**Format:** Plain text string

**Example:**
```yaml
WEBHOOK_SECRET: "my-webhook-secret-123"
```

**Pros:**
- ✅ Simple - no Secret Manager needed
- ✅ Fast - no API calls

**Cons:**
- ❌ Insecure - secret in config file
- ❌ No audit trail
- ❌ Hard to rotate
- ❌ Risk of committing to version control

**⚠️ Use only for local development/testing!**

## How It Works

### Loading Priority

1. **Check `WEBHOOK_SECRET`** - If set, use it directly (no Secret Manager call)
2. **Check `WEBHOOK_SECRET_NAME`** - If set, load from Secret Manager
3. **Use default** - `projects/1054147886816/secrets/webhook-secret/versions/latest`

### Code Flow

```
app.go
  ├─> configs.LoadEnvironment()
  │     └─> Loads WEBHOOK_SECRET and WEBHOOK_SECRET_NAME from env
  │
  └─> services.LoadWebhookSecret(config)
        ├─> If config.WebhookSecret is set → use it
        └─> Else → load from Secret Manager using config.WebhookSecretName
              └─> Store in config.WebhookSecret
```

### Signature Verification

```
webhook_handler.go
  └─> ParseWebhookDataWithConfig()
        └─> verifySignatureFunc(sigHeader, payload, []byte(config.WebhookSecret))
              └─> HMAC-SHA256 verification
```

## Your Current Setup

Based on your Secret Manager output:

```yaml
name: projects/1054147886816/secrets/webhook-secret
createTime: '2025-10-06T17:56:10.467642Z'
replication:
  automatic: {}
```

### Your env.yaml Should Use:

```yaml
env_variables:
  GITHUB_APP_ID: "1166559"
  INSTALLATION_ID: "62138132"
  REPO_NAME: "docs-code-examples"
  REPO_OWNER: "mongodb"
  
  # GitHub App private key from Secret Manager
  GITHUB_APP_PRIVATE_KEY_SECRET_NAME: "projects/1054147886816/secrets/CODE_COPIER_PEM/versions/latest"
  
  # Webhook secret from Secret Manager (RECOMMENDED)
  WEBHOOK_SECRET_NAME: "projects/1054147886816/secrets/webhook-secret/versions/latest"
  
  # Other config...
  COMMITTER_NAME: "GitHub Copier App"
  COMMITTER_EMAIL: "bot@mongodb.com"
  PORT: "8080"
  WEBSERVER_PATH: "/events"
  CONFIG_FILE: "copier-config.yaml"
  DEPRECATION_FILE: "deprecated_examples.json"
  GOOGLE_PROJECT_ID: "1054147886816"
  GOOGLE_LOG_NAME: "code-copier-log"
```

## Testing Locally

For local testing, you can use `SKIP_SECRET_MANAGER=true`:

```bash
# In your .env.local file
SKIP_SECRET_MANAGER=true
WEBHOOK_SECRET="test-secret-123"
GITHUB_APP_PRIVATE_KEY="-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----"
```

Then run:
```bash
go run app.go -env .env.local
```

## Rotating the Webhook Secret

### Step 1: Create New Secret Version

```bash
# Generate new secret
NEW_SECRET=$(openssl rand -hex 32)

# Add new version to Secret Manager
echo -n "$NEW_SECRET" | gcloud secrets versions add webhook-secret \
  --data-file=-
```

### Step 2: Update GitHub Webhook

1. Go to your source repository on GitHub
2. Settings → Webhooks
3. Edit the webhook
4. Update the "Secret" field with the new secret
5. Save

### Step 3: Verify

The application will automatically use the latest version (no redeployment needed if using `versions/latest`).

```bash
# Test webhook delivery
# GitHub → Settings → Webhooks → Recent Deliveries → Redeliver
```

### Step 4: Disable Old Version (Optional)

```bash
# List versions
gcloud secrets versions list webhook-secret

# Disable old version
gcloud secrets versions disable VERSION_NUMBER --secret=webhook-secret
```

## Troubleshooting

### Error: "failed to load webhook secret"

**Cause:** Secret Manager client can't access the secret

**Solutions:**
1. Verify secret exists:
   ```bash
   gcloud secrets describe webhook-secret
   ```

2. Check IAM permissions:
   ```bash
   gcloud secrets get-iam-policy webhook-secret
   ```

3. Grant access:
   ```bash
   PROJECT_ID=$(gcloud config get-value project)
   gcloud secrets add-iam-policy-binding webhook-secret \
     --member="serviceAccount:${PROJECT_ID}@appspot.gserviceaccount.com" \
     --role="roles/secretmanager.secretAccessor"
   ```

### Error: "webhook signature verification failed"

**Cause:** Secret in Secret Manager doesn't match GitHub webhook secret

**Solutions:**
1. Get secret from Secret Manager:
   ```bash
   gcloud secrets versions access latest --secret=webhook-secret
   ```

2. Compare with GitHub webhook secret:
   - GitHub → Settings → Webhooks → Edit
   - Check the "Secret" field

3. Update one to match the other

### Error: "SKIP_SECRET_MANAGER=true but no WEBHOOK_SECRET set"

**Cause:** Testing locally without providing direct secret

**Solution:**
```bash
export WEBHOOK_SECRET="test-secret-123"
```

## Security Best Practices

### ✅ DO

- ✅ Use Secret Manager in production
- ✅ Use `versions/latest` for automatic rotation
- ✅ Grant minimal IAM permissions
- ✅ Rotate secrets regularly (every 90 days)
- ✅ Use different secrets for different environments
- ✅ Monitor Secret Manager audit logs

### ❌ DON'T

- ❌ Hardcode secrets in env.yaml for production
- ❌ Commit env.yaml to version control
- ❌ Share secrets via email or chat
- ❌ Use the same secret across multiple apps
- ❌ Grant broad IAM permissions
- ❌ Use weak secrets (use `openssl rand -hex 32`)

## Comparison with GitHub Private Key

Both secrets are now loaded from Secret Manager:

| Secret | Environment Variable | Secret Manager Path |
|--------|---------------------|---------------------|
| GitHub Private Key | `GITHUB_APP_PRIVATE_KEY_SECRET_NAME` | `projects/.../secrets/CODE_COPIER_PEM/versions/latest` |
| Webhook Secret | `WEBHOOK_SECRET_NAME` | `projects/.../secrets/webhook-secret/versions/latest` |

**Consistency:** Both use the same pattern for security and maintainability.

## Summary

**Before (Insecure):**
```yaml
WEBHOOK_SECRET: "hardcoded-secret-in-config-file"
```

**After (Secure):**
```yaml
WEBHOOK_SECRET_NAME: "projects/1054147886816/secrets/webhook-secret/versions/latest"
```

**Result:**
- ✅ Secret stored securely in Secret Manager
- ✅ No secrets in config files
- ✅ Easy rotation without redeployment
- ✅ Audit trail of all access
- ✅ Fine-grained access control

---

**Ready to deploy?** Your webhook secret is already in Secret Manager, so just update your env.yaml to use `WEBHOOK_SECRET_NAME` instead of `WEBHOOK_SECRET`! 🔒

