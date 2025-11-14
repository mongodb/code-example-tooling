# Deployment Checklist: Workflow Config

Quick checklist for deploying the workflow-based configuration.

---

## Pre-Deployment

- [ ] Workflow config validates: `cd examples-copier && go run test-workflow-config.go ../copier-config-workflow.yaml`
- [ ] App compiles: `cd examples-copier && go build .`
- [ ] Have access to `mongodb/docs-mongodb-internal` repo (config repo)
- [ ] Have access to Google Cloud project `github-copy-code-examples`
- [ ] `gcloud` CLI installed: `gcloud --version`
- [ ] Authenticated to Google Cloud: `gcloud auth list`

---

## Step 1: Upload Config (5 minutes)

**Upload `copier-config-workflow.yaml` to mongodb/docs-mongodb-internal**

⚠️ **Important:** The config goes in `mongodb/docs-mongodb-internal` (the config repo), NOT in `mongodb/docs-sample-apps` (the source repo).

### Quick Method (GitHub Web UI):
1. [ ] Go to https://github.com/mongodb/docs-mongodb-internal
2. [ ] Click "Add file" → "Create new file"
3. [ ] Name: `copier-config-workflow.yaml`
4. [ ] Copy contents from local file: `copier-config-workflow.yaml`
5. [ ] Commit to `main` branch with message: "Add workflow-based config for testing"

### Verify:
```bash
curl -s https://raw.githubusercontent.com/mongodb/docs-mongodb-internal/main/copier-config-workflow.yaml | head -5
# Should show the workflow config
```

---

## Step 2: Configure App (2 minutes)

**Update app to use workflow config**

```bash
cd examples-copier

# Create env.yaml from production template
cp configs/env.yaml.production env.yaml

# Update CONFIG_FILE to use workflow config
sed -i '' 's/CONFIG_FILE: "copier-config.yaml"/CONFIG_FILE: "copier-config-workflow.yaml"/' env.yaml

# Verify
grep CONFIG_FILE env.yaml
```

**Expected output:**
```
CONFIG_FILE: "copier-config-workflow.yaml"
```

- [ ] `env.yaml` created
- [ ] `CONFIG_FILE` points to `copier-config-workflow.yaml`

---

## Step 3: Deploy (10 minutes)

**Deploy to Google Cloud App Engine**

```bash
cd examples-copier

# Set project
gcloud config set project github-copy-code-examples

# Deploy
gcloud app deploy app.yaml --quiet
```

**Wait for deployment to complete** (5-10 minutes)

### Verify Deployment:
```bash
# Check health endpoint
curl https://github-copy-code-examples.uc.r.appspot.com/health

# Should return:
# {"status":"healthy","timestamp":"2024-..."}
```

- [ ] Deployment successful
- [ ] Health endpoint returns `{"status":"healthy"}`
- [ ] No errors in logs: `gcloud app logs tail -s default`

---

## Step 4: Test (10 minutes)

**Create test PR to trigger workflow processing**

### In `mongodb/docs-sample-apps`:

```bash
# Create test branch
git checkout -b test-workflow-config

# Make a small change
echo "# Test workflow config" >> mflix/client/README.md

# Commit and push
git add mflix/client/README.md
git commit -m "Test: Trigger workflow processor"
git push origin test-workflow-config
```

### On GitHub:
1. [ ] Create PR from `test-workflow-config` branch
2. [ ] Title: "Test: Workflow config processor"
3. [ ] Merge the PR

---

## Step 5: Verify (5 minutes)

**Check that workflows processed correctly**

### Check App Logs:
```bash
gcloud app logs tail -s default
```

**Look for:**
- [ ] `"processing files with workflows"`
- [ ] `"found matching workflows"` with `matching_count: 3`
- [ ] `"workflow processing complete"`
- [ ] No errors

### Check Target Repos:

**PRs should be created in:**
1. [ ] https://github.com/mongodb/sample-app-java-mflix/pulls
2. [ ] https://github.com/mongodb/sample-app-nodejs-mflix/pulls
3. [ ] https://github.com/mongodb/sample-app-python-mflix/pulls

**Each PR should:**
- [ ] Have title: "Update MFlix application from docs-sample-apps"
- [ ] Include `client/README.md` with test change
- [ ] Have proper PR body with source info

### Check Metrics:
```bash
curl https://github-copy-code-examples.uc.r.appspot.com/metrics
```

- [ ] `webhooks_received` increased by 1
- [ ] `files_matched` increased by 3
- [ ] `files_uploaded` increased by 3
- [ ] `webhooks_failed` did NOT increase

---

## Success! ✅

If all checks pass:
- [x] Workflow config is working
- [x] Multi-org support is ready
- [x] Ready for production use

---

## Next Steps

### Immediate:
- [ ] Monitor for 24 hours
- [ ] Check a few more PRs to ensure consistency

### This Week:
- [ ] Switch to workflow config permanently:
  ```bash
  # In docs-sample-apps repo
  mv copier-config.yaml copier-config-legacy.yaml.backup
  mv copier-config-workflow.yaml copier-config.yaml
  
  # Update env.yaml
  sed -i '' 's/CONFIG_FILE: "copier-config-workflow.yaml"/CONFIG_FILE: "copier-config.yaml"/' env.yaml
  
  # Redeploy
  gcloud app deploy
  ```

### When Ready:
- [ ] Add more apps as workflows
- [ ] Install in additional orgs (10gen, mongodb-university)
- [ ] Consider centralized config repo

---

## Rollback (If Needed)

**If something goes wrong:**

```bash
cd examples-copier

# Switch back to legacy config
sed -i '' 's/CONFIG_FILE: "copier-config-workflow.yaml"/CONFIG_FILE: "copier-config.yaml"/' env.yaml

# Redeploy
gcloud app deploy app.yaml --quiet
```

---

## Troubleshooting

### No PRs Created?
1. Check webhook delivery: https://github.com/mongodb/docs-sample-apps/settings/hooks
2. Check app logs: `gcloud app logs tail -s default`
3. Verify config file exists in source repo

### Wrong Files Copied?
1. Check transformation definitions in config
2. Review app logs for transformation details
3. Update config and redeploy

### Deployment Failed?
1. Verify `env.yaml` exists: `ls -la examples-copier/env.yaml`
2. Check Go version: `go version` (should be 1.23+)
3. Verify project: `gcloud config get-value project`

---

## Time Estimate

- **Pre-deployment:** 5 minutes
- **Step 1 (Upload):** 5 minutes
- **Step 2 (Configure):** 2 minutes
- **Step 3 (Deploy):** 10 minutes
- **Step 4 (Test):** 10 minutes
- **Step 5 (Verify):** 5 minutes

**Total: ~40 minutes**

---

## Support

- **Detailed guide:** See `DEPLOYMENT_GUIDE.md`
- **Architecture:** See `IMPLEMENTATION_COMPLETE.md`
- **Logs:** `gcloud app logs tail -s default`
- **Metrics:** `curl https://github-copy-code-examples.uc.r.appspot.com/metrics`

