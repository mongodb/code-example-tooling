# GitHub Metrics Collection Scheduling

## Overview

The GitHub metrics collection job is designed to run **every 14 days** to collect repository metrics from GitHub. This interval ensures we capture metrics within GitHub's 14-day data retention window while avoiding unnecessary API calls.

## How It Works

### File-Based Tracking

The system uses a file-based approach to track when the job last ran successfully:

1. **Persistent Storage**: A Kubernetes persistent volume is mounted at `/data` to store the last run timestamp
2. **Last Run File**: The file `/data/last-run.json` contains the timestamp of the last successful run
3. **Automatic Checking**: Each time the cronjob executes, it checks if 14 days have passed since the last run
4. **Skip Logic**: If less than 14 days have passed, the job exits successfully without collecting metrics

### Cronjob Schedule

The Kubernetes cronjob is configured to run **every Sunday at 8am UTC** (`0 8 * * 0`).

- The job runs weekly, but the application logic determines whether to actually collect metrics
- This approach is more reliable than trying to schedule exactly every 14 days with cron syntax
- If a run is missed (e.g., due to maintenance), the next weekly run will catch it

### Example Timeline

```
Week 1, Sunday: Job runs → Collects metrics → Records timestamp
Week 2, Sunday: Job runs → Checks timestamp → Skips (only 7 days)
Week 3, Sunday: Job runs → Checks timestamp → Collects metrics (14+ days) → Records timestamp
Week 4, Sunday: Job runs → Checks timestamp → Skips (only 7 days)
Week 5, Sunday: Job runs → Checks timestamp → Collects metrics (14+ days) → Records timestamp
```

## Implementation Details

### Last Run Tracker Module

The `last-run-tracker.js` module provides three main functions:

#### `shouldRunMetricsCollection()`
Checks if 14 days have passed since the last run.

**Returns:**
```javascript
{
  shouldRun: boolean,        // true if 14+ days have passed
  lastRun: Date|null,        // timestamp of last run
  daysSinceLastRun: number|null  // days since last run
}
```

#### `recordSuccessfulRun()`
Records the current timestamp as the last successful run.

#### `getLastRunInfo()`
Gets information about the last run without checking if we should run (useful for debugging).

### Last Run File Format

The `/data/last-run.json` file contains:

```json
{
  "lastRun": "2025-12-03T08:00:00.000Z",
  "timestamp": 1733212800000
}
```

## Configuration

### Cronjob Configuration (`cronjobs.yml`)

```yaml
persistence:
  enabled: true
  storageClass: "standard"
  accessMode: ReadWriteOnce
  size: 1Gi
  mountPath: /data

cronJobs:
  - name: github-metrics-collection
    schedule: "0 8 * * 0"  # Every Sunday at 8am UTC
    command:
      - node
      - index.js
```

### Key Configuration Points

- **Persistent Volume**: Required to maintain state between cronjob executions
- **Mount Path**: `/data` - where the last run file is stored
- **Schedule**: Weekly execution allows the application to decide when to run
- **Exit Code**: The job exits with code 0 (success) even when skipping, so Kubernetes doesn't mark it as failed

## Monitoring

### Checking Last Run Status

To check when the job last ran, you can:

1. **View the last-run file** in the persistent volume:
   ```bash
   kubectl exec -it <pod-name> -n docs -- cat /data/last-run.json
   ```

2. **Check job logs** for skip messages:
   ```bash
   kubectl logs -n docs -l job-name=github-metrics-collection --tail=50
   ```

### Expected Log Output

**When running:**
```
No previous run detected. This is the first run.
Starting metrics collection...
✓ Metrics collection completed successfully
✓ Recorded successful run at 2025-12-03T08:00:00.000Z
```

**When skipping:**
```
Last run: 2025-12-03T08:00:00.000Z
Days since last run: 7
✗ Only 7 days have passed. Skipping metrics collection.
  Next run should occur in approximately 7 days.
Skipping metrics collection - not enough time has passed since last run.
Last run was 7 days ago on 2025-12-03T08:00:00.000Z
```

## Troubleshooting

### Job Never Runs

If the job keeps skipping even though 14+ days have passed:

1. Check the last-run file timestamp:
   ```bash
   kubectl exec -it <pod-name> -n docs -- cat /data/last-run.json
   ```

2. Manually delete the file to force a run:
   ```bash
   kubectl exec -it <pod-name> -n docs -- rm /data/last-run.json
   ```

### Persistent Volume Issues

If the persistent volume isn't working:

1. Check if the PVC is bound:
   ```bash
   kubectl get pvc -n docs
   ```

2. Check pod events for volume mount errors:
   ```bash
   kubectl describe pod <pod-name> -n docs
   ```

### Force a Run

To force the job to run immediately regardless of the last run time:

1. Delete the last-run file:
   ```bash
   kubectl exec -it <pod-name> -n docs -- rm /data/last-run.json
   ```

2. Manually trigger the cronjob:
   ```bash
   kubectl create job --from=cronjob/github-metrics-collection manual-run-$(date +%s) -n docs
   ```

## Benefits of This Approach

1. **Reliable 14-day interval**: Ensures metrics are collected every 14 days without complex cron syntax
2. **Resilient to missed runs**: If a run is missed, the next execution will catch it
3. **Simple to monitor**: Clear log messages indicate whether the job ran or skipped
4. **Easy to override**: Can force a run by deleting the last-run file
5. **Kubernetes-native**: Uses persistent volumes for state management
6. **No external dependencies**: Doesn't require a database or external service to track state

