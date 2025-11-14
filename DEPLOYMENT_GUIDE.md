# Deployment Guide: Workflow Config

This guide walks you through deploying and testing the new workflow-based configuration.

---

## Prerequisites

- [x] Workflow implementation complete
- [x] Config validates successfully
- [ ] Access to `mongodb/docs-sample-apps` repo
- [ ] Access to Google Cloud project `github-copy-code-examples`
- [ ] `gcloud` CLI installed and authenticated

---

## Step 1: Upload Workflow Config to Source Repo

**Task:** Add the workflow config to `mongodb/docs-sample-apps`

### Option A: Use GitHub Web UI (Easiest)

1. Go to https://github.com/mongodb/docs-sample-apps
2. Click "Add file" → "Create new file"
3. Name it: `copier-config-workflow.yaml`
4. Copy the contents from `copier-config-workflow.yaml` in this repo
5. Commit directly to `main` branch with message: "Add workflow-based config for testing"

### Option B: Use Git CLI

```bash
# Clone the repo (if you don't have it)
git clone git@github.com:mongodb/docs-sample-apps.git
cd docs-sample-apps

# Copy the workflow config
cp /path/to/code-example-tooling/copier-config-workflow.yaml .

# Commit and push
git add copier-config-workflow.yaml
git commit -m "Add workflow-based config for testing"
git push origin main
```

**Verification:**
- [ ] File exists at https://github.com/mongodb/docs-sample-apps/blob/main/copier-config-workflow.yaml

---

## Step 2: Update App Configuration

**Task:** Point the app to use the workflow config

### Update env.yaml

```bash
cd examples-copier

# Copy production config to env.yaml (used for deployment)
cp configs/env.yaml.production env.yaml

# Edit env.yaml to use workflow config
# Change line 28 from:
#   CONFIG_FILE: "copier-config.yaml"
# To:
#   CONFIG_FILE: "copier-config-workflow.yaml"
```

**Quick command:**
```bash
cd examples-copier
cp configs/env.yaml.production env.yaml
sed -i '' 's/CONFIG_FILE: "copier-config.yaml"/CONFIG_FILE: "copier-config-workflow.yaml"/' env.yaml
```

**Verification:**
```bash
grep CONFIG_FILE env.yaml
# Should output: CONFIG_FILE: "copier-config-workflow.yaml"
```

---

## Step 3: Deploy to App Engine

**Task:** Deploy the updated app

### Build and Deploy

```bash
cd examples-copier

# Verify you're authenticated
gcloud auth list

# Set the project
gcloud config set project github-copy-code-examples

# Deploy (this takes 5-10 minutes)
gcloud app deploy app.yaml --quiet
```

**What happens during deployment:**
1. Code is uploaded to Google Cloud
2. Docker image is built with Go 1.23
3. App Engine creates new instances
4. Health checks verify the app is running
5. Traffic is switched to new version

**Verification:**
```bash
# Check deployment status
gcloud app versions list

# Check logs
gcloud app logs tail -s default

# Test health endpoint
curl https://github-copy-code-examples.uc.r.appspot.com/health
# Should return: {"status":"healthy","timestamp":"..."}
```

---

## Step 4: Create Test PR

**Task:** Create a test PR to verify workflow processing

### Create a Test Branch

```bash
cd docs-sample-apps  # Your local clone

# Create test branch
git checkout -b test-workflow-config

# Make a small change to trigger the copier
echo "# Test change" >> mflix/client/README.md

# Commit and push
git add mflix/client/README.md
git commit -m "Test: Trigger workflow processor"
git push origin test-workflow-config
```

### Create PR on GitHub

1. Go to https://github.com/mongodb/docs-sample-apps
2. Click "Compare & pull request" for your test branch
3. Title: "Test: Workflow config processor"
4. Body: "Testing new workflow-based configuration. Will merge to trigger copier."
5. Create the PR

### Merge the PR

1. Approve and merge the PR
2. This triggers the webhook to the copier app

---

## Step 5: Verify Results

**Task:** Check that the workflow processor worked correctly

### Check App Logs

```bash
# Watch logs in real-time
gcloud app logs tail -s default

# Look for these log messages:
# ✅ "processing files with workflows"
# ✅ "found matching workflows" (should show 3 workflows)
# ✅ "workflow processing complete"
```

**What to look for:**
```
INFO: processing files with workflows
  file_count: 1
  workflow_count: 3

INFO: found matching workflows
  webhook_repo: mongodb/docs-sample-apps
  matching_count: 3

INFO: workflow processing complete
  workflow_count: 3
```

### Check Target Repos

The copier should create PRs in these repos:

1. **mongodb/sample-app-java-mflix**
   - Go to: https://github.com/mongodb/sample-app-java-mflix/pulls
   - Look for PR: "Update MFlix application from docs-sample-apps"
   - Verify files: `client/README.md` should be updated

2. **mongodb/sample-app-nodejs-mflix**
   - Go to: https://github.com/mongodb/sample-app-nodejs-mflix/pulls
   - Look for PR with same title
   - Verify files: `client/README.md` should be updated

3. **mongodb/sample-app-python-mflix**
   - Go to: https://github.com/mongodb/sample-app-python-mflix/pulls
   - Look for PR with same title
   - Verify files: `client/README.md` should be updated

### Check Metrics

```bash
# Check metrics endpoint
curl https://github-copy-code-examples.uc.r.appspot.com/metrics
```

**Look for:**
- `webhooks_received` should increase by 1
- `files_matched` should increase by 3 (one per workflow)
- `files_uploaded` should increase by 3
- `webhooks_failed` should NOT increase

---

## Troubleshooting

### Issue: No PRs Created

**Check:**
1. App logs for errors: `gcloud app logs tail -s default`
2. Webhook delivery in GitHub:
   - Go to https://github.com/mongodb/docs-sample-apps/settings/hooks
   - Click on the webhook
   - Check "Recent Deliveries"
   - Look for 200 response (success) or error

**Common causes:**
- Webhook not configured
- App not deployed
- Config file not found
- Authentication issues

### Issue: Wrong Files Copied

**Check:**
1. Workflow transformations in config
2. App logs for transformation details
3. File paths in source repo

**Fix:**
- Update transformations in `copier-config-workflow.yaml`
- Redeploy app
- Create another test PR

### Issue: App Won't Deploy

**Check:**
1. `env.yaml` exists in `examples-copier/` directory
2. Go version in `app.yaml` matches installed version
3. Google Cloud project is correct

**Fix:**
```bash
# Verify env.yaml exists
ls -la examples-copier/env.yaml

# Check Go version
go version

# Verify project
gcloud config get-value project
```

---

## Success Criteria

- [x] Workflow config uploaded to source repo
- [x] App deployed successfully
- [x] Test PR merged
- [ ] App logs show "processing files with workflows"
- [ ] App logs show "found matching workflows: 3"
- [ ] PRs created in all 3 target repos
- [ ] Files copied correctly to target repos
- [ ] Metrics show successful uploads
- [ ] No errors in app logs

---

## Rollback Plan

If something goes wrong:

### Option 1: Switch Back to Legacy Config

```bash
cd examples-copier

# Edit env.yaml
sed -i '' 's/CONFIG_FILE: "copier-config-workflow.yaml"/CONFIG_FILE: "copier-config.yaml"/' env.yaml

# Redeploy
gcloud app deploy app.yaml --quiet
```

### Option 2: Roll Back to Previous Version

```bash
# List versions
gcloud app versions list

# Route traffic to previous version
gcloud app services set-traffic default --splits=<PREVIOUS_VERSION>=1
```

---

## Next Steps After Success

Once verified:

1. **Switch to workflow config permanently:**
   ```bash
   # In docs-sample-apps repo
   mv copier-config.yaml copier-config-legacy.yaml.backup
   mv copier-config-workflow.yaml copier-config.yaml
   
   # Update env.yaml back to:
   CONFIG_FILE: "copier-config.yaml"
   
   # Redeploy
   gcloud app deploy
   ```

2. **Add more apps** as workflows (15 min each)

3. **Install in other orgs** when ready (2-3 hours per org)

4. **Monitor for a week** to ensure stability

---

## Questions?

- Check `IMPLEMENTATION_COMPLETE.md` for architecture details
- Check `TROUBLESHOOTING.md` for common issues
- Check app logs: `gcloud app logs tail -s default`
- Check metrics: `curl https://github-copy-code-examples.uc.r.appspot.com/metrics`

