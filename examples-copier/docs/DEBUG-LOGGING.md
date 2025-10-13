# Debug Logging Guide

This guide explains how to enable and use debug logging in the Examples Copier application.

## Overview

The Examples Copier supports configurable logging levels to help with development, troubleshooting, and debugging. By default, the application logs at the INFO level, but you can enable DEBUG logging for more verbose output.

## Environment Variables

### LOG_LEVEL

**Purpose:** Set the logging level for the application

**Values:**
- `info` (default) - Standard operational logs
- `debug` - Verbose debug logs with detailed operation information

**Example:**
```bash
LOG_LEVEL="debug"
```

### COPIER_DEBUG

**Purpose:** Alternative way to enable debug mode

**Values:**
- `true` - Enable debug logging
- `false` (default) - Standard logging

**Example:**
```bash
COPIER_DEBUG="true"
```

**Note:** Either `LOG_LEVEL="debug"` OR `COPIER_DEBUG="true"` will enable debug logging. You only need to set one.

### COPIER_DISABLE_CLOUD_LOGGING

**Purpose:** Disable Google Cloud Logging (useful for local development)

**Values:**
- `true` - Disable GCP logging, only log to stdout
- `false` (default) - Enable GCP logging if configured

**Example:**
```bash
COPIER_DISABLE_CLOUD_LOGGING="true"
```

**Use case:** When developing locally, you may not want logs sent to Google Cloud. This flag keeps all logs local.

---

## How It Works

### Code Implementation

The logging system is implemented in `services/logger.go`:

```go
// LogDebug writes debug logs only when LOG_LEVEL=debug or COPIER_DEBUG=true.
func LogDebug(message string) {
    if !isDebugEnabled() {
        return
    }
    // Mirror to GCP as info if available, plus prefix to stdout
    if googleInfoLogger != nil && gcpLoggingEnabled {
        googleInfoLogger.Println("[DEBUG] " + message)
    }
    log.Println("[DEBUG] " + message)
}

func isDebugEnabled() bool {
    if strings.EqualFold(os.Getenv("LOG_LEVEL"), "debug") {
        return true
    }
    return strings.EqualFold(os.Getenv("COPIER_DEBUG"), "true")
}

func isCloudLoggingDisabled() bool {
    return strings.EqualFold(os.Getenv("COPIER_DISABLE_CLOUD_LOGGING"), "true")
}
```

### Log Levels

The application supports the following log levels:

| Level | Function | When to Use | Example |
|-------|----------|-------------|---------|
| **DEBUG** | `LogDebug()` | Detailed operation logs, file matching, API calls | `[DEBUG] Matched file: src/example.js` |
| **INFO** | `LogInfo()` | Standard operational logs | `[INFO] Processing webhook event` |
| **WARN** | `LogWarning()` | Warning conditions | `[WARN] File not found, skipping` |
| **ERROR** | `LogError()` | Error conditions | `[ERROR] Failed to create PR` |
| **CRITICAL** | `LogCritical()` | Critical failures | `[CRITICAL] Database connection failed` |

---

## Usage Examples

### Local Development with Debug Logging

**Using .env file:**
```bash
# configs/.env
LOG_LEVEL="debug"
COPIER_DISABLE_CLOUD_LOGGING="true"
DRY_RUN="true"
```

**Using environment variables:**
```bash
export LOG_LEVEL=debug
export COPIER_DISABLE_CLOUD_LOGGING=true
export DRY_RUN=true
go run app.go
```

### Production with Debug Logging (Temporary)

**env.yaml:**
```yaml
env_variables:
  LOG_LEVEL: "debug"
  # ... other variables
```

**Deploy:**
```bash
gcloud app deploy app.yaml  # env.yaml is included via 'includes' directive
```

**Important:** Remember to disable debug logging after troubleshooting to reduce log volume and costs.

### Local Development without Cloud Logging

```bash
# configs/.env
COPIER_DISABLE_CLOUD_LOGGING="true"
```

This keeps all logs local (stdout only), which is faster and doesn't require GCP credentials.

---

## What Gets Logged at DEBUG Level?

When debug logging is enabled, you'll see additional information about:

### 1. **File Matching Operations**
```
[DEBUG] Checking pattern: src/**/*.js
[DEBUG] Matched file: src/examples/example1.js
[DEBUG] Excluded file: src/tests/test.js (matches exclude pattern)
```

### 2. **GitHub API Calls**
```
[DEBUG] Fetching file from GitHub: src/example.js
[DEBUG] Creating PR for target repo: mongodb/docs-code-examples
[DEBUG] GitHub API response: 200 OK
```

### 3. **Configuration Loading**
```
[DEBUG] Loading config file: copier-config.yaml
[DEBUG] Found 5 copy rules
[DEBUG] Rule 1: Copy src/**/*.js to examples/
```

### 4. **Webhook Processing**
```
[DEBUG] Received webhook event: pull_request
[DEBUG] PR action: closed
[DEBUG] PR merged: true
[DEBUG] Processing 3 changed files
```

### 5. **Pattern Matching**
```
[DEBUG] Testing pattern: src/**/*.{js,ts}
[DEBUG] File matches: true
[DEBUG] Applying transformations: 2
```

---

## Best Practices

### ✅ DO

- **Enable debug logging when troubleshooting issues**
  ```bash
  LOG_LEVEL="debug"
  ```

- **Disable cloud logging for local development**
  ```bash
  COPIER_DISABLE_CLOUD_LOGGING="true"
  ```

- **Use debug logging with dry run mode for testing**
  ```bash
  LOG_LEVEL="debug"
  DRY_RUN="true"
  ```

- **Disable debug logging in production after troubleshooting**
  - High log volume can increase costs
  - May expose sensitive information

### ❌ DON'T

- **Don't leave debug logging enabled in production long-term**
  - Increases log volume and storage costs
  - May impact performance
  - Can expose internal implementation details

- **Don't rely on debug logs for critical monitoring**
  - Use INFO/WARN/ERROR levels for operational monitoring
  - Debug logs may be disabled in production

- **Don't log sensitive data even in debug mode**
  - The code already avoids logging secrets
  - Be careful when adding new debug logs

---

## Troubleshooting

### Debug Logs Not Appearing

**Problem:** Set `LOG_LEVEL="debug"` but not seeing debug logs

**Solutions:**

1. **Check the variable is set correctly:**
   ```bash
   echo $LOG_LEVEL
   # Should output: debug
   ```

2. **Try the alternative flag:**
   ```bash
   COPIER_DEBUG="true"
   ```

3. **Check case sensitivity:**
   ```bash
   # Both work (case-insensitive):
   LOG_LEVEL="debug"
   LOG_LEVEL="DEBUG"
   ```

4. **Verify the code is calling LogDebug():**
   - Not all operations have debug logs
   - Check `services/logger.go` for `LogDebug()` calls

### Logs Not Going to Google Cloud

**Problem:** Logs appear in stdout but not in Google Cloud Logging

**Solutions:**

1. **Check if cloud logging is disabled:**
   ```bash
   # Remove or set to false:
   # COPIER_DISABLE_CLOUD_LOGGING="true"
   ```

2. **Verify GCP credentials:**
   ```bash
   gcloud auth application-default login
   ```

3. **Check project ID is set:**
   ```bash
   GOOGLE_CLOUD_PROJECT_ID="your-project-id"
   ```

4. **Check log name is set:**
   ```bash
   COPIER_LOG_NAME="code-copier-log"
   ```

### Too Many Logs

**Problem:** Debug logging produces too much output

**Solutions:**

1. **Disable debug logging:**
   ```bash
   # Remove or comment out:
   # LOG_LEVEL="debug"
   # COPIER_DEBUG="true"
   ```

2. **Use grep to filter:**
   ```bash
   # Show only errors:
   go run app.go 2>&1 | grep ERROR
   
   # Show only specific operations:
   go run app.go 2>&1 | grep "pattern matching"
   ```

3. **Redirect to file:**
   ```bash
   go run app.go > debug.log 2>&1
   ```

---

## Configuration Examples

### Example 1: Local Development (Recommended)

```bash
# configs/.env
LOG_LEVEL="debug"
COPIER_DISABLE_CLOUD_LOGGING="true"
DRY_RUN="true"
AUDIT_ENABLED="false"
METRICS_ENABLED="true"
```

**Why:**
- Debug logs help understand what's happening
- No cloud logging keeps it fast and local
- Dry run prevents accidental changes
- No audit logging (simpler setup)

### Example 2: Production Troubleshooting

```yaml
# env.yaml
env_variables:
  LOG_LEVEL: "debug"
  GOOGLE_CLOUD_PROJECT_ID: "your-project-id"
  COPIER_LOG_NAME: "code-copier-log"
  # ... other variables
```

**Why:**
- Temporarily enable debug for troubleshooting
- Logs go to Cloud Logging for analysis
- Remember to disable after fixing issue

### Example 3: Local with Cloud Logging

```bash
# configs/.env
LOG_LEVEL="debug"
GOOGLE_CLOUD_PROJECT_ID="your-project-id"
COPIER_LOG_NAME="code-copier-log-dev"
# COPIER_DISABLE_CLOUD_LOGGING not set (defaults to false)
```

**Why:**
- Test cloud logging integration locally
- Separate log name for dev environment
- Useful for testing logging infrastructure

---

## See Also

- [LOCAL-TESTING.md](LOCAL-TESTING.md) - Local development guide
- [TROUBLESHOOTING.md](TROUBLESHOOTING.md) - General troubleshooting
- [CONFIGURATION-GUIDE.md](CONFIGURATION-GUIDE.md) - Complete configuration reference
- [../configs/env.yaml.example](../configs/env.yaml.example) - All environment variables
- [../configs/.env.example](../configs/.env.example) - Local development template

