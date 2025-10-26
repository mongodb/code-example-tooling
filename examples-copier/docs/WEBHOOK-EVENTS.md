# Webhook Events Guide

## Overview

The examples-copier application receives GitHub webhook events and processes them to copy code examples between repositories. This document explains which events are processed and which are ignored.

## Supported Events

### Pull Request Events (`pull_request`)

**Status:** ✅ **Processed**

The application **only** processes `pull_request` events with the following criteria:

- **Action:** `closed`
- **Merged:** `true`

All other pull request actions are ignored:
- `opened` - PR created but not merged
- `synchronize` - PR updated with new commits
- `edited` - PR title/description changed
- `labeled` - Labels added/removed
- `review_requested` - Reviewers requested
- etc.

**Example Log Output:**
```
[INFO] PR event received | {"action":"closed","merged":true}
[INFO] processing merged PR | {"pr_number":123,"repo":"owner/repo","sha":"abc123"}
```

## Ignored Events

The following GitHub webhook events are **intentionally ignored** and will not trigger any processing:

### Common Ignored Events

| Event Type | Description | Why Ignored |
|------------|-------------|-------------|
| `ping` | GitHub webhook test | Not a code change |
| `push` | Direct push to branch | Only process merged PRs |
| `installation` | App installed/uninstalled | Not relevant to copying |
| `installation_repositories` | Repos added/removed from app | Not relevant to copying |
| `repository` | Repository created/deleted | Not relevant to copying |
| `workflow_run` | GitHub Actions workflow | Not relevant to copying |
| `check_run` | CI check completed | Not relevant to copying |
| `status` | Commit status updated | Not relevant to copying |

**Example Log Output:**
```
[INFO] ignoring non-pull_request event | {"event_type":"ping","size_bytes":7233}
```

## Monitoring Webhook Events

### Viewing Metrics

Check the `/metrics` endpoint to see webhook event statistics:

```bash
curl https://your-app.appspot.com/metrics | jq '.webhooks'
```

**Example Response:**
```json
{
  "received": 150,
  "processed": 45,
  "failed": 2,
  "ignored": 103,
  "event_types": {
    "pull_request": 45,
    "ping": 5,
    "push": 50,
    "workflow_run": 48
  },
  "success_rate": 95.74,
  "processing_time": {
    "avg_ms": 1250,
    "min_ms": 450,
    "max_ms": 3200,
    "p50_ms": 1100,
    "p95_ms": 2800,
    "p99_ms": 3100
  }
}
```

### Understanding the Metrics

- **`received`**: Total webhooks received (all event types)
- **`processed`**: Successfully processed merged PRs
- **`failed`**: Webhooks that encountered errors
- **`ignored`**: Non-PR events or non-merged PRs
- **`event_types`**: Breakdown by GitHub event type
- **`success_rate`**: Percentage of received webhooks successfully processed

### Viewing Logs

**Local Development:**
```bash
# Watch application logs
tail -f logs/app.log | grep "event_type"
```

**Google Cloud Platform:**
```bash
# View recent logs
gcloud app logs tail -s default | grep "event_type"

# Filter for ignored events
gcloud app logs tail -s default | grep "ignoring non-pull_request"
```

## Configuring GitHub Webhooks

### Recommended Configuration

When setting up the GitHub webhook in your repository settings:

1. **Payload URL:** `https://your-app.appspot.com/events`
2. **Content type:** `application/json`
3. **Secret:** (use your webhook secret)
4. **Events:** Select **"Pull requests"** only

### Why Select Only "Pull requests"?

While the application safely ignores other event types, selecting only "Pull requests" reduces unnecessary webhook traffic and makes monitoring clearer.

**Benefits:**
- ✅ Reduces network traffic
- ✅ Reduces log noise
- ✅ Easier to monitor and debug
- ✅ Lower webhook delivery quota usage

### If You Need Multiple Event Types

If your webhook is shared with other systems that need different events, it's safe to enable additional event types. The examples-copier will simply ignore them.

## Troubleshooting

### High Number of Ignored Events

**Symptom:** Metrics show many ignored events

**Possible Causes:**
1. **Webhook configured for all events** - Reconfigure to only send `pull_request` events
2. **Multiple webhooks configured** - Check repository settings for duplicate webhooks
3. **Shared webhook** - Other systems may be using the same endpoint

**Solution:**
```bash
# Check webhook configuration
# Go to: https://github.com/YOUR_ORG/YOUR_REPO/settings/hooks

# Verify only "Pull requests" is selected
```

### No Events Being Processed

**Symptom:** `processed` count is 0, but `ignored` count is high

**Possible Causes:**
1. **PRs not being merged** - Only merged PRs are processed
2. **Wrong event type** - Verify webhook sends `pull_request` events
3. **Configuration error** - Check copier-config.yaml exists and is valid

**Solution:**
```bash
# Check recent webhook deliveries in GitHub
# Go to: https://github.com/YOUR_ORG/YOUR_REPO/settings/hooks/WEBHOOK_ID

# Look for:
# - Event type: pull_request
# - Action: closed
# - Merged: true
```

### Unexpected Event Types

**Symptom:** Seeing event types you didn't expect

**Common Scenarios:**
1. **`ping` events** - GitHub sends these when webhook is created/edited (normal)
2. **`push` events** - Someone may have enabled this in webhook settings
3. **`workflow_run` events** - GitHub Actions workflows triggering webhooks

**Solution:**
Review and update webhook configuration to only send necessary events.

## Best Practices

### 1. Monitor Event Type Distribution

Regularly check the `event_types` breakdown in metrics:

```bash
curl https://your-app.appspot.com/metrics | jq '.webhooks.event_types'
```

**Expected Distribution:**
- Most events should be `pull_request`
- Occasional `ping` events are normal
- High numbers of other types suggest misconfiguration

### 2. Set Up Alerts

Configure alerts for:
- High `failed` count
- Low `success_rate` (< 90%)
- Unexpected event types appearing

### 3. Regular Audits

Periodically review:
- GitHub webhook configuration
- Application logs for ignored events
- Metrics trends over time

## Related Documentation

- [DEPLOYMENT.md](DEPLOYMENT.md) - Webhook configuration during deployment
- [WEBHOOK-TESTING.md](WEBHOOK-TESTING.md) - Testing webhook processing
- [MONITORING.md](MONITORING.md) - Monitoring and alerting setup

## Summary

- ✅ **Only merged PRs are processed**
- ✅ **All other events are safely ignored**
- ✅ **Metrics track all event types**
- ✅ **Configure webhook to send only `pull_request` events for best results**
- ✅ **Monitor `/metrics` endpoint to understand webhook traffic**

