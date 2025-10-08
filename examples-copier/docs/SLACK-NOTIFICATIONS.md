# Slack Notifications

The examples-copier supports sending notifications to Slack when PRs are processed, files are copied, or errors occur.

## Features

- ✅ **PR Processed Notifications** - Get notified when a PR is successfully processed
- ✅ **Error Notifications** - Get alerted when errors occur
- ✅ **Files Copied Notifications** - See which files were copied
- ✅ **Deprecation Notifications** - Track when files are deprecated
- ✅ **Rich Formatting** - Color-coded messages with detailed information
- ✅ **Customizable** - Configure channel, username, and icon

## Setup

### 1. Create a Slack Incoming Webhook

1. Go to your Slack workspace settings
2. Navigate to **Apps** → **Manage** → **Custom Integrations** → **Incoming Webhooks**
3. Click **Add to Slack**
4. Select the channel where you want notifications
5. Copy the **Webhook URL** (looks like `https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXX`)

### 2. Configure Environment Variables

Add these environment variables to your configuration:

```bash
# Required: Slack webhook URL
SLACK_WEBHOOK_URL="https://hooks.slack.com/services/YOUR/WEBHOOK/URL"

# Optional: Customize notification settings
SLACK_CHANNEL="#code-examples"           # Default: #code-examples
SLACK_USERNAME="Examples Copier"         # Default: Examples Copier
SLACK_ICON_EMOJI=":robot_face:"          # Default: :robot_face:
SLACK_ENABLED=true                       # Default: true if webhook URL is set
```

### 3. Test the Integration

Run the app and trigger a webhook:

```bash
# Start the app with Slack enabled
SLACK_WEBHOOK_URL="https://hooks.slack.com/services/..." \
CONFIG_FILE=copier-config.yaml \
make run-local-quick

# Send a test webhook
./test-webhook -payload test-payloads/example-pr-merged.json
```

You should see a notification in your Slack channel!

## Notification Types

### 1. PR Processed Notification

Sent when a PR is successfully processed.

**Includes:**
- PR number and title
- Link to the PR
- Repository name
- Files matched, copied, and failed counts
- Processing time

**Color:**
- 🟢 Green - All files copied successfully
- 🟡 Yellow - Some files failed

**Example:**
```
✅ PR #42 Processed
Add Go database examples

Repository: mongodb/docs-code-examples
Files Matched: 20
Files Copied: 18
Files Failed: 2
Processing Time: 5.2s
```

### 2. Error Notification

Sent when an error occurs during processing.

**Includes:**
- Operation that failed
- Error message
- Repository name
- PR number (if applicable)

**Color:** 🔴 Red

**Example:**
```
❌ Error Occurred
An error occurred during config_load

Operation: config_load
Error: failed to retrieve config file: 404 Not Found
Repository: mongodb/docs-code-examples
PR Number: #42
```

### 3. Files Copied Notification

Sent when files are successfully copied to a target repository.

**Includes:**
- PR number
- Source and target repositories
- Rule name that matched
- File count
- List of files (up to 10, then "... and X more")

**Color:** 🟢 Green

**Example:**
```
📋 Files Copied from PR #42

• generated-examples/test-project/cmd/main.go
• generated-examples/test-project/internal/auth.go
... and 8 more

Source: mongodb/docs-code-examples
Target: mongodb/target-repo
Rule: Copy generated examples
File Count: 10
```

### 4. Deprecation Notification

Sent when files are marked as deprecated (deleted from source).

**Includes:**
- PR number
- Repository name
- File count
- List of deprecated files

**Color:** 🟡 Yellow

**Example:**
```
⚠️ Files Deprecated from PR #42

• old-examples/deprecated.go
• old-examples/removed.py

Repository: mongodb/docs-code-examples
File Count: 2
```

## Configuration Options

### Environment Variables

| Variable            | Description                  | Default                   | Required |
|---------------------|------------------------------|---------------------------|----------|
| `SLACK_WEBHOOK_URL` | Slack incoming webhook URL   | -                         | Yes      |
| `SLACK_CHANNEL`     | Channel to post to           | `#code-examples`          | No       |
| `SLACK_USERNAME`    | Bot username                 | `Examples Copier`         | No       |
| `SLACK_ICON_EMOJI`  | Bot icon emoji               | `:robot_face:`            | No       |
| `SLACK_ENABLED`     | Enable/disable notifications | `true` if webhook URL set | No       |

### Disabling Notifications

To disable Slack notifications:

```bash
# Option 1: Don't set SLACK_WEBHOOK_URL
# (notifications will be automatically disabled)

# Option 2: Explicitly disable
SLACK_ENABLED=false
```

## Customization

### Custom Channel

Override the default channel for specific notifications:

```bash
SLACK_CHANNEL="#deployments"
```

### Custom Username and Icon

Personalize the bot appearance:

```bash
SLACK_USERNAME="Code Copier Bot"
SLACK_ICON_EMOJI=":package:"
```

Available emoji options:
- `:robot_face:` - 🤖 Robot
- `:package:` - 📦 Package
- `:rocket:` - 🚀 Rocket
- `:gear:` - ⚙️ Gear
- `:bell:` - 🔔 Bell
- `:clipboard:` - 📋 Clipboard

## Testing

### Test with Example Payload

```bash
# Set your Slack webhook URL
export SLACK_WEBHOOK_URL="https://hooks.slack.com/services/..."

# Start the app
CONFIG_FILE=copier-config.yaml make run-local-quick

# Send test webhook
./test-webhook -payload test-payloads/example-pr-merged.json
```

### Test with Real PR

```bash
export SLACK_WEBHOOK_URL="https://hooks.slack.com/services/..."
export GITHUB_TOKEN="ghp_your_token"

./test-webhook -pr 42 -owner mongodb -repo docs-code-examples
```

## Troubleshooting

### Notifications Not Appearing

1. **Check webhook URL is set:**
   ```bash
   echo $SLACK_WEBHOOK_URL
   ```

2. **Check Slack is enabled:**
   ```bash
   echo $SLACK_ENABLED
   ```

3. **Check app logs for errors:**
   ```
   [ERROR] failed to send slack message: ...
   ```

4. **Verify webhook URL is valid:**
   - Should start with `https://hooks.slack.com/services/`
   - Test it with curl:
     ```bash
     curl -X POST -H 'Content-type: application/json' \
       --data '{"text":"Test message"}' \
       $SLACK_WEBHOOK_URL
     ```

### Wrong Channel

If notifications go to the wrong channel:

1. **Check SLACK_CHANNEL environment variable:**
   ```bash
   echo $SLACK_CHANNEL
   ```

2. **Note:** The webhook URL has a default channel configured in Slack. The `SLACK_CHANNEL` variable can override this, but the bot must have permission to post to that channel.

### Notifications Too Noisy

To reduce notification frequency:

1. **Disable specific notification types** (requires code changes)
2. **Use a dedicated channel** for copier notifications
3. **Adjust Slack channel notification settings**

## Production Deployment

### Google Cloud Run

Add environment variables to your Cloud Run service:

```bash
gcloud run services update examples-copier \
  --set-env-vars="SLACK_WEBHOOK_URL=https://hooks.slack.com/services/..." \
  --set-env-vars="SLACK_CHANNEL=#code-examples"
```

### Docker

Add to your `docker-compose.yml`:

```yaml
services:
  examples-copier:
    environment:
      - SLACK_WEBHOOK_URL=https://hooks.slack.com/services/...
      - SLACK_CHANNEL=#code-examples
      - SLACK_USERNAME=Examples Copier
      - SLACK_ICON_EMOJI=:robot_face:
```

## Security Considerations

1. **Keep webhook URL secret** - It allows posting to your Slack workspace
2. **Use environment variables** - Don't commit webhook URLs to git
3. **Rotate webhooks periodically** - Create new webhooks if compromised
4. **Limit channel permissions** - Use a dedicated channel for bot notifications

## Examples

### Minimal Configuration

```bash
SLACK_WEBHOOK_URL="https://hooks.slack.com/services/T00/B00/XXX"
```

### Full Configuration

```bash
SLACK_WEBHOOK_URL="https://hooks.slack.com/services/T00/B00/XXX"
SLACK_CHANNEL="#deployments"
SLACK_USERNAME="Code Examples Bot"
SLACK_ICON_EMOJI=":package:"
SLACK_ENABLED=true
```

### Disable in Development

```bash
# Don't set SLACK_WEBHOOK_URL in development
# Or explicitly disable:
SLACK_ENABLED=false
```

## See Also

- [Slack Incoming Webhooks Documentation](https://api.slack.com/messaging/webhooks)
- [Slack Message Formatting](https://api.slack.com/reference/surfaces/formatting)
- [LOCAL-TESTING.md](LOCAL-TESTING.md) - Local testing guide

