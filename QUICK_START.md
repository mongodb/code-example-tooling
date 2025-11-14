# Quick Start: Deploy Workflow Config

**Time:** 40 minutes  
**Goal:** Deploy and test the new workflow-based configuration

---

## 🎯 The 5 Steps

### Step 1: Upload Config to Config Repo (5 min)

**Where:** `mongodb/docs-mongodb-internal` (the config repo)

**How:**
1. Go to https://github.com/mongodb/docs-mongodb-internal
2. Click "Add file" → "Create new file"
3. Name it: `copier-config-workflow.yaml`
4. Copy the entire contents from your local `copier-config-workflow.yaml` file
5. Commit message: "Add workflow-based config for testing"
6. Commit directly to `main` branch

**Verify:**
```bash
curl -s https://raw.githubusercontent.com/mongodb/docs-mongodb-internal/main/copier-config-workflow.yaml | head -10
```

---

### Step 2: Configure App (2 min)

**Where:** Your local `code-example-tooling` repo

```bash
cd examples-copier

# Create env.yaml from production template
cp configs/env.yaml.production env.yaml

# Update CONFIG_FILE to use workflow config
sed -i '' 's/CONFIG_FILE: "copier-config.yaml"/CONFIG_FILE: "copier-config-workflow.yaml"/' env.yaml

# Verify the change
grep CONFIG_FILE env.yaml
# Should output: CONFIG_FILE: "copier-config-workflow.yaml"
```

---

### Step 3: Deploy to App Engine (10 min)

```bash
cd examples-copier

# Set the Google Cloud project
gcloud config set project github-copy-code-examples

# Deploy (takes 5-10 minutes)
gcloud app deploy app.yaml --quiet
```

**Wait for deployment to complete...**

**Verify:**
```bash
# Test health endpoint
curl https://github-copy-code-examples.uc.r.appspot.com/health

# Should return:
# {"status":"healthy","timestamp":"2024-11-14T..."}
```

---

### Step 4: Create Test PR (10 min)

**Where:** `mongodb/docs-sample-apps` (the source repo)

```bash
# Clone or navigate to docs-sample-apps repo
cd /path/to/docs-sample-apps

# Create test branch
git checkout main
git pull
git checkout -b test-workflow-config

# Make a small change to trigger the copier
echo "# Test workflow config - $(date)" >> mflix/client/README.md

# Commit and push
git add mflix/client/README.md
git commit -m "Test: Trigger workflow processor"
git push origin test-workflow-config
```

**On GitHub:**
1. Go to https://github.com/mongodb/docs-sample-apps
2. Click "Compare & pull request"
3. Title: "Test: Workflow config processor"
4. Body: "Testing new workflow-based configuration"
5. Create the PR
6. **Merge the PR** (this triggers the webhook)

---

### Step 5: Verify Results (5 min)

#### A. Check App Logs

```bash
gcloud app logs tail -s default
```

**Look for these messages:**
```
✅ "processing files with workflows"
✅ "found matching workflows" with matching_count: 3
✅ "workflow processing complete"
```

#### B. Check Target Repos for PRs

The copier should create PRs in these 3 repos:

1. **Java:** https://github.com/mongodb/sample-app-java-mflix/pulls
2. **Node.js:** https://github.com/mongodb/sample-app-nodejs-mflix/pulls
3. **Python:** https://github.com/mongodb/sample-app-python-mflix/pulls

**Each PR should have:**
- Title: "Update MFlix application from docs-sample-apps"
- File: `client/README.md` with your test change
- Proper PR body with source information

#### C. Check Metrics

```bash
curl https://github-copy-code-examples.uc.r.appspot.com/metrics
```

**Look for:**
- `webhooks_received` increased by 1
- `files_matched` increased by 3
- `files_uploaded` increased by 3

---

## ✅ Success Criteria

- [ ] Config uploaded to `mongodb/docs-mongodb-internal`
- [ ] App deployed successfully
- [ ] Health endpoint returns healthy
- [ ] Test PR merged in `mongodb/docs-sample-apps`
- [ ] Logs show "processing files with workflows"
- [ ] 3 PRs created in target repos
- [ ] Each PR has correct files
- [ ] Metrics show successful uploads

---

## 🆘 If Something Goes Wrong

### Rollback to Legacy Config

```bash
cd examples-copier

# Switch back to old config
sed -i '' 's/CONFIG_FILE: "copier-config-workflow.yaml"/CONFIG_FILE: "copier-config.yaml"/' env.yaml

# Redeploy
gcloud app deploy app.yaml --quiet
```

### Common Issues

**Issue: Config file not found**
- Check that `copier-config-workflow.yaml` exists in `mongodb/docs-mongodb-internal`
- Verify the file name is exactly `copier-config-workflow.yaml`

**Issue: No PRs created**
- Check webhook delivery: https://github.com/mongodb/docs-sample-apps/settings/hooks
- Check app logs for errors: `gcloud app logs tail -s default`

**Issue: Deployment failed**
- Verify `env.yaml` exists: `ls -la examples-copier/env.yaml`
- Check you're authenticated: `gcloud auth list`

---

## 📊 What Happens When You Merge the Test PR

```
1. GitHub sends webhook to your app
   ↓
2. App receives webhook from mongodb/docs-sample-apps
   ↓
3. App loads copier-config-workflow.yaml from mongodb/docs-mongodb-internal
   ↓
4. App finds 3 workflows matching the source repo:
   - mflix-java
   - mflix-nodejs
   - mflix-python
   ↓
5. App processes each workflow:
   - Applies transformations (move/copy)
   - Queues files for upload
   ↓
6. App creates 3 PRs in target repos:
   - mongodb/sample-app-java-mflix
   - mongodb/sample-app-nodejs-mflix
   - mongodb/sample-app-python-mflix
   ↓
7. Done! ✅
```

---

## 🎉 After Success

Once everything works:

1. **Monitor for 24 hours** - Make sure a few more PRs work correctly

2. **Switch to workflow config permanently:**
   ```bash
   # In mongodb/docs-mongodb-internal repo
   # Rename copier-config.yaml to copier-config-legacy.yaml.backup
   # Rename copier-config-workflow.yaml to copier-config.yaml
   
   # Update env.yaml
   sed -i '' 's/CONFIG_FILE: "copier-config-workflow.yaml"/CONFIG_FILE: "copier-config.yaml"/' env.yaml
   
   # Redeploy
   gcloud app deploy
   ```

3. **Add more apps** as workflows (15 min each)

4. **Install in other orgs** when ready (2-3 hours per org)

---

## 📚 Need More Help?

- **Detailed guide:** `DEPLOYMENT_GUIDE.md`
- **Checklist format:** `DEPLOYMENT_CHECKLIST.md`
- **Architecture:** `IMPLEMENTATION_COMPLETE.md`
- **Check logs:** `gcloud app logs tail -s default`
- **Check metrics:** `curl https://github-copy-code-examples.uc.r.appspot.com/metrics`

---

## ⏱️ Time Breakdown

- Step 1 (Upload): 5 minutes
- Step 2 (Configure): 2 minutes
- Step 3 (Deploy): 10 minutes
- Step 4 (Test PR): 10 minutes
- Step 5 (Verify): 5 minutes

**Total: ~40 minutes**

---

**Ready? Start with Step 1!** 🚀

