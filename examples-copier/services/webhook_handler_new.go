package services

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/go-github/v48/github"
	"github.com/mongodb/code-example-tooling/code-copier/configs"
	"github.com/mongodb/code-example-tooling/code-copier/types"
)

const (
	maxWebhookBodyBytes = 1 << 20 // 1MB
	// GitHub GraphQL API returns file status in uppercase for the ChangeType field
	// Possible values: ADDED, MODIFIED, DELETED, RENAMED, COPIED, CHANGED
	statusDeleted = "DELETED"
)

// simpleVerifySignature verifies the webhook signature
func simpleVerifySignature(sigHeader string, body, secret []byte) bool {
	if sigHeader == "" {
		return false
	}

	// Remove "sha256=" prefix
	if !strings.HasPrefix(sigHeader, "sha256=") {
		return false
	}
	signature := sigHeader[7:]

	// Compute HMAC
	mac := hmac.New(sha256.New, secret)
	mac.Write(body)
	expectedMAC := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expectedMAC))
}

// RetrieveFileContentsWithConfigAndBranch fetches file contents from a specific branch
func RetrieveFileContentsWithConfigAndBranch(ctx context.Context, filePath string, branch string, repoOwner string, repoName string) (*github.RepositoryContent, error) {
	client := GetRestClient()

	fileContent, _, _, err := client.Repositories.GetContents(
		ctx,
		repoOwner,
		repoName,
		filePath,
		&github.RepositoryContentGetOptions{
			Ref: branch,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get file content: %w", err)
	}

	return fileContent, nil
}

// HandleWebhookWithContainer handles incoming GitHub webhook requests using the service container
func HandleWebhookWithContainer(w http.ResponseWriter, r *http.Request, config *configs.Config, container *ServiceContainer) {
	startTime := time.Now()
	ctx := r.Context()

	LogInfoCtx(ctx, "webhook handler started", map[string]interface{}{
		"elapsed_ms": time.Since(startTime).Milliseconds(),
	})

	// Read and validate webhook payload
	limited := io.LimitReader(r.Body, maxWebhookBodyBytes)
	payload, err := io.ReadAll(limited)
	if err != nil {
		LogWebhookOperation(ctx, "read_body", "failed to read webhook body", err)
		container.MetricsCollector.RecordWebhookFailed()
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	eventType := r.Header.Get("X-GitHub-Event")
	if eventType == "" {
		LogWebhookOperation(ctx, "missing_event", "missing X-GitHub-Event header", nil)
		container.MetricsCollector.RecordWebhookFailed()
		http.Error(w, "missing event type", http.StatusBadRequest)
		return
	}

	LogInfoCtx(ctx, "payload read", map[string]interface{}{
		"elapsed_ms": time.Since(startTime).Milliseconds(),
		"size_bytes": len(payload),
	})

	// Verify webhook signature
	if config.WebhookSecret != "" {
		sigHeader := r.Header.Get("X-Hub-Signature-256")
		if !simpleVerifySignature(sigHeader, payload, []byte(config.WebhookSecret)) {
			LogWebhookOperation(ctx, "signature_verification", "webhook signature verification failed", nil)
			container.MetricsCollector.RecordWebhookFailed()
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		LogInfoCtx(ctx, "signature verified", map[string]interface{}{
			"elapsed_ms": time.Since(startTime).Milliseconds(),
		})
	}

	// Parse webhook event
	evt, err := github.ParseWebHook(eventType, payload)
	if err != nil {
		LogWebhookOperation(ctx, "parse_payload", "failed to parse webhook payload", err,
			map[string]interface{}{"event_type": eventType})
		container.MetricsCollector.RecordWebhookFailed()
		http.Error(w, "bad webhook", http.StatusBadRequest)
		return
	}

	// Check if it's a pull_request event
	prEvt, ok := evt.(*github.PullRequestEvent)
	if !ok || prEvt.GetPullRequest() == nil {
		// Record ignored webhook with event type
		container.MetricsCollector.RecordWebhookIgnored(eventType)

		// Log with event type for better debugging
		LogInfoCtx(ctx, "ignoring non-pull_request event", map[string]interface{}{
			"event_type": eventType,
			"size_bytes": len(payload),
		})
		w.WriteHeader(http.StatusNoContent)
		return
	}

	action := prEvt.GetAction()
	merged := prEvt.GetPullRequest().GetMerged()

	LogInfoCtx(ctx, "PR event received", map[string]interface{}{
		"action": action,
		"merged": merged,
	})

	if !(action == "closed" && merged) {
		LogInfoCtx(ctx, "skipping non-merged PR", map[string]interface{}{
			"action": action,
			"merged": merged,
		})
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Process the merged PR
	prNumber := prEvt.GetPullRequest().GetNumber()
	sourceCommitSHA := prEvt.GetPullRequest().GetMergeCommitSHA()

	// Extract repository info from webhook payload
	repo := prEvt.GetRepo()
	if repo == nil {
		LogWarningCtx(ctx, "webhook missing repository info", nil)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	repoOwner := repo.GetOwner().GetLogin()
	repoName := repo.GetName()

	LogInfoCtx(ctx, "processing merged PR", map[string]interface{}{
		"pr_number":  prNumber,
		"sha":        sourceCommitSHA,
		"repo":       fmt.Sprintf("%s/%s", repoOwner, repoName),
		"elapsed_ms": time.Since(startTime).Milliseconds(),
	})

	// Respond immediately to avoid GitHub webhook timeout
	LogInfoCtx(ctx, "sending immediate response", map[string]interface{}{
		"elapsed_ms": time.Since(startTime).Milliseconds(),
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(`{"status":"accepted"}`))

	LogInfoCtx(ctx, "response sent", map[string]interface{}{
		"elapsed_ms": time.Since(startTime).Milliseconds(),
	})

	// Flush the response immediately
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
		LogInfoCtx(ctx, "response flushed", map[string]interface{}{
			"elapsed_ms": time.Since(startTime).Milliseconds(),
		})
	}

	// Process asynchronously in background with a new context
	// Don't use the request context as it will be cancelled when the request completes
	bgCtx := context.Background()
	go handleMergedPRWithContainer(bgCtx, prNumber, sourceCommitSHA, repoOwner, repoName, config, container)
}

// handleMergedPRWithContainer processes a merged PR using the new pattern matching system
func handleMergedPRWithContainer(ctx context.Context, prNumber int, sourceCommitSHA string, repoOwner string, repoName string, config *configs.Config, container *ServiceContainer) {
	startTime := time.Now()

	// Configure GitHub permissions
	if InstallationAccessToken == "" {
		ConfigurePermissions()
	}

	// Load configuration using new loader
	// Note: config.ConfigRepoOwner and config.ConfigRepoName are already set from env.yaml
	// The webhook repoOwner/repoName are used for matching workflows, not for loading config
	yamlConfig, err := container.ConfigLoader.LoadConfig(ctx, config)
	if err != nil {
		LogAndReturnError(ctx, "config_load", "failed to load config", err)
		container.MetricsCollector.RecordWebhookFailed()

		// Send error notification to Slack
		container.SlackNotifier.NotifyError(ctx, &ErrorEvent{
			Operation:  "config_load",
			Error:      err,
			PRNumber:   prNumber,
			SourceRepo: fmt.Sprintf("%s/%s", repoOwner, repoName),
		})
		return
	}

	// Find workflows matching this source repo
	webhookRepo := fmt.Sprintf("%s/%s", repoOwner, repoName)
	matchingWorkflows := []types.Workflow{}
	for _, workflow := range yamlConfig.Workflows {
		if workflow.Source.Repo == webhookRepo {
			matchingWorkflows = append(matchingWorkflows, workflow)
		}
	}

	if len(matchingWorkflows) == 0 {
		LogWarningCtx(ctx, "no workflows configured for source repository", map[string]interface{}{
			"webhook_repo":   webhookRepo,
			"workflow_count": len(yamlConfig.Workflows),
		})
		container.MetricsCollector.RecordWebhookFailed()
		return
	}

	LogInfoCtx(ctx, "found matching workflows", map[string]interface{}{
		"webhook_repo":   webhookRepo,
		"matching_count": len(matchingWorkflows),
	})

	// Store matching workflows for processing
	yamlConfig.Workflows = matchingWorkflows

	// Get changed files from PR (from the source repository that triggered the webhook)
	changedFiles, err := GetFilesChangedInPr(repoOwner, repoName, prNumber)
	if err != nil {
		LogAndReturnError(ctx, "get_files", "failed to get changed files", err)
		container.MetricsCollector.RecordWebhookFailed()

		// Send error notification to Slack
		container.SlackNotifier.NotifyError(ctx, &ErrorEvent{
			Operation:  "get_files",
			Error:      err,
			PRNumber:   prNumber,
			SourceRepo: webhookRepo,
		})
		return
	}

	LogInfoCtx(ctx, "retrieved changed files", map[string]interface{}{
		"count": len(changedFiles),
	})

	// Track metrics before processing
	filesMatchedBefore := container.MetricsCollector.GetFilesMatched()
	filesUploadedBefore := container.MetricsCollector.GetFilesUploaded()
	filesFailedBefore := container.MetricsCollector.GetFilesUploadFailed()

	// Process files with workflow processor
	processFilesWithWorkflows(ctx, prNumber, sourceCommitSHA, changedFiles, yamlConfig, container)

	// Upload queued files
	FilesToUpload = container.FileStateService.GetFilesToUpload()
	AddFilesToTargetRepoBranchWithFetcher(container.PRTemplateFetcher, container.MetricsCollector)
	container.FileStateService.ClearFilesToUpload()

	// Update deprecation file - copy from FileStateService to global map for legacy function
	deprecationMap := container.FileStateService.GetFilesToDeprecate()
	FilesToDeprecate = make(map[string]types.Configs)
	for _, entry := range deprecationMap {
		FilesToDeprecate[entry.FileName] = types.Configs{
			TargetRepo:   entry.Repo,
			TargetBranch: entry.Branch,
		}
	}
	UpdateDeprecationFile()
	container.FileStateService.ClearFilesToDeprecate()

	// Calculate metrics after processing
	filesMatched := container.MetricsCollector.GetFilesMatched() - filesMatchedBefore
	filesUploaded := container.MetricsCollector.GetFilesUploaded() - filesUploadedBefore
	filesFailed := container.MetricsCollector.GetFilesUploadFailed() - filesFailedBefore
	processingTime := time.Since(startTime)

	LogInfoCtx(ctx, "--Done--", map[string]interface{}{
		"pr_number": prNumber,
		"sha":       sourceCommitSHA,
	})

	// Send success notification to Slack
	container.SlackNotifier.NotifyPRProcessed(ctx, &PRProcessedEvent{
		PRNumber:       prNumber,
		PRTitle:        fmt.Sprintf("PR #%d", prNumber), // TODO: Get actual PR title from GitHub
		PRURL:          fmt.Sprintf("https://github.com/%s/pull/%d", webhookRepo, prNumber),
		SourceRepo:     webhookRepo,
		FilesMatched:   filesMatched,
		FilesCopied:    filesUploaded,
		FilesFailed:    filesFailed,
		ProcessingTime: processingTime,
	})
}

// processFilesWithWorkflows processes changed files using the workflow system
func processFilesWithWorkflows(ctx context.Context, prNumber int, sourceCommitSHA string,
	changedFiles []types.ChangedFile, yamlConfig *types.YAMLConfig, container *ServiceContainer) {

	LogInfoCtx(ctx, "processing files with workflows", map[string]interface{}{
		"file_count":     len(changedFiles),
		"workflow_count": len(yamlConfig.Workflows),
	})

	// Create workflow processor
	workflowProcessor := NewWorkflowProcessor(
		container.PatternMatcher,
		container.PathTransformer,
		container.FileStateService,
		container.MetricsCollector,
		container.MessageTemplater,
	)

	// Process each workflow
	for _, workflow := range yamlConfig.Workflows {
		if err := ctx.Err(); err != nil {
			LogWebhookOperation(ctx, "workflow_processing", "workflow processing cancelled", err)
			return
		}

		err := workflowProcessor.ProcessWorkflow(ctx, workflow, changedFiles, prNumber, sourceCommitSHA)
		if err != nil {
			LogErrorCtx(ctx, "failed to process workflow", err, map[string]interface{}{
				"workflow_name": workflow.Name,
			})
			// Continue processing other workflows
			continue
		}
	}

	LogInfoCtx(ctx, "workflow processing complete", map[string]interface{}{
		"workflow_count": len(yamlConfig.Workflows),
	})
}


