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
	statusDeleted       = "DELETED"
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
func RetrieveFileContentsWithConfigAndBranch(ctx context.Context, filePath string, branch string, config *configs.Config) (*github.RepositoryContent, error) {
	client := GetRestClient()

	fileContent, _, _, err := client.Repositories.GetContents(
		ctx,
		config.RepoOwner,
		config.RepoName,
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

	// Check if it's a merged PR event
	prEvt, ok := evt.(*github.PullRequestEvent)
	if !ok || prEvt.GetPullRequest() == nil {
		LogWarningCtx(ctx, "payload not pull_request event", nil)
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

	// Update config with actual repository from webhook
	config.RepoOwner = repoOwner
	config.RepoName = repoName

	// Load configuration using new loader
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

	// Set source repo in config if not set
	if yamlConfig.SourceRepo == "" {
		yamlConfig.SourceRepo = fmt.Sprintf("%s/%s", config.RepoOwner, config.RepoName)
	}

	// Validate webhook is from expected source repository
	webhookRepo := fmt.Sprintf("%s/%s", repoOwner, repoName)
	if webhookRepo != yamlConfig.SourceRepo {
		LogWarningCtx(ctx, "webhook from unexpected repository", map[string]interface{}{
			"webhook_repo":  webhookRepo,
			"expected_repo": yamlConfig.SourceRepo,
		})
		container.MetricsCollector.RecordWebhookFailed()
		return
	}

	// Get changed files from PR
	changedFiles, err := GetFilesChangedInPr(prNumber)
	if err != nil {
		LogAndReturnError(ctx, "get_files", "failed to get changed files", err)
		container.MetricsCollector.RecordWebhookFailed()

		// Send error notification to Slack
		container.SlackNotifier.NotifyError(ctx, &ErrorEvent{
			Operation:  "get_files",
			Error:      err,
			PRNumber:   prNumber,
			SourceRepo: yamlConfig.SourceRepo,
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

	// Process files with new pattern matching
	processFilesWithPatternMatching(ctx, prNumber, sourceCommitSHA, changedFiles, yamlConfig, config, container)

	// Upload queued files
	FilesToUpload = container.FileStateService.GetFilesToUpload()
	AddFilesToTargetRepoBranch()
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
		PRURL:          fmt.Sprintf("https://github.com/%s/pull/%d", yamlConfig.SourceRepo, prNumber),
		SourceRepo:     yamlConfig.SourceRepo,
		FilesMatched:   filesMatched,
		FilesCopied:    filesUploaded,
		FilesFailed:    filesFailed,
		ProcessingTime: processingTime,
	})
}

// processFilesWithPatternMatching processes changed files using the new pattern matching system
func processFilesWithPatternMatching(ctx context.Context, prNumber int, sourceCommitSHA string,
	changedFiles []types.ChangedFile, yamlConfig *types.YAMLConfig, config *configs.Config, container *ServiceContainer) {

	LogInfoCtx(ctx, "processing files with pattern matching", map[string]interface{}{
		"file_count": len(changedFiles),
		"rule_count": len(yamlConfig.CopyRules),
	})

	// Log first few files for debugging
	for i, file := range changedFiles {
		if i < 3 {
			LogInfoCtx(ctx, "sample file path", map[string]interface{}{
				"index": i,
				"path":  file.Path,
			})
		}
	}

	for _, file := range changedFiles {
		if err := ctx.Err(); err != nil {
			LogWebhookOperation(ctx, "file_iteration", "file iteration cancelled", err)
			return
		}

		// Try to match file against each rule
		for _, rule := range yamlConfig.CopyRules {
			if err := ctx.Err(); err != nil {
				LogWebhookOperation(ctx, "file_iteration", "file iteration cancelled", err)
				return
			}

			// Match file against pattern
			matchResult := container.PatternMatcher.Match(file.Path, rule.SourcePattern)
			if !matchResult.Matched {
				continue
			}

			// Record matched file
			container.MetricsCollector.RecordFileMatched()

			LogInfoCtx(ctx, "file matched pattern", map[string]interface{}{
				"file":      file.Path,
				"rule":      rule.Name,
				"pattern":   rule.SourcePattern.Pattern,
				"variables": matchResult.Variables,
			})

			// Process each target
			for _, target := range rule.Targets {
				processFileForTarget(ctx, prNumber, sourceCommitSHA, file, rule, target, matchResult.Variables, yamlConfig.SourceBranch, config, container)
			}
		}
	}
}

// processFileForTarget processes a single file for a specific target
func processFileForTarget(ctx context.Context, prNumber int, sourceCommitSHA string, file types.ChangedFile,
	rule types.CopyRule, target types.TargetConfig, variables map[string]string, sourceBranch string, config *configs.Config, container *ServiceContainer) {

	// Transform path
	targetPath, err := container.PathTransformer.Transform(file.Path, target.PathTransform, variables)
	if err != nil {
		LogErrorCtx(ctx, "failed to transform path", err,
			map[string]interface{}{
				"operation":   "path_transform",
				"source_path": file.Path,
				"template":    target.PathTransform,
			})
		return
	}

	// Handle deleted files
	if file.Status == statusDeleted {
		handleFileDeprecation(ctx, prNumber, sourceCommitSHA, file, rule, target, targetPath, sourceBranch, config, container)
		return
	}

	// Handle file copy
	handleFileCopyWithAudit(ctx, prNumber, sourceCommitSHA, file, rule, target, targetPath, variables, sourceBranch, config, container)
}

// handleFileCopyWithAudit handles file copying with audit logging
func handleFileCopyWithAudit(ctx context.Context, prNumber int, sourceCommitSHA string, file types.ChangedFile,
	rule types.CopyRule, target types.TargetConfig, targetPath string, variables map[string]string, sourceBranch string,
	config *configs.Config, container *ServiceContainer) {

	startTime := time.Now()
	sourceRepo := fmt.Sprintf("%s/%s", config.RepoOwner, config.RepoName)

	// Retrieve file content from the source commit SHA (the merge commit)
	// This ensures we fetch the exact version of the file that was merged
	fc, err := RetrieveFileContentsWithConfigAndBranch(ctx, file.Path, sourceCommitSHA, config)
	if err != nil {
		// Log error event
		container.AuditLogger.LogErrorEvent(ctx, &AuditEvent{
			RuleName:     rule.Name,
			SourceRepo:   sourceRepo,
			SourcePath:   file.Path,
			TargetRepo:   target.Repo,
			TargetPath:   targetPath,
			CommitSHA:    sourceCommitSHA,
			PRNumber:     prNumber,
			Success:      false,
			ErrorMessage: err.Error(),
			DurationMs:   time.Since(startTime).Milliseconds(),
		})
		container.MetricsCollector.RecordFileUploadFailed()
		LogFileOperation(ctx, "retrieve", file.Path, target.Repo, "failed to retrieve file", err)
		return
	}

	// Update file name to target path
	fc.Name = github.String(targetPath)

	// Queue file for upload
	queueFileForUploadWithStrategy(target, *fc, rule, variables, prNumber, sourceCommitSHA, sourceBranch, config, container)

	// Log successful copy event
	fileSize := int64(0)
	if fc.Content != nil {
		fileSize = int64(len(*fc.Content))
	}

	container.AuditLogger.LogCopyEvent(ctx, &AuditEvent{
		RuleName:   rule.Name,
		SourceRepo: sourceRepo,
		SourcePath: file.Path,
		TargetRepo: target.Repo,
		TargetPath: targetPath,
		CommitSHA:  sourceCommitSHA,
		PRNumber:   prNumber,
		Success:    true,
		DurationMs: time.Since(startTime).Milliseconds(),
		FileSize:   fileSize,
		AdditionalData: map[string]any{
			"variables": variables,
		},
	})

	container.MetricsCollector.RecordFileUploaded(time.Since(startTime))

	LogFileOperation(ctx, "queue_copy", file.Path, target.Repo, "file queued for copy", nil,
		map[string]interface{}{
			"target_path": targetPath,
			"rule":        rule.Name,
		})
}

// handleFileDeprecation handles file deprecation with audit logging
func handleFileDeprecation(ctx context.Context, prNumber int, sourceCommitSHA string, file types.ChangedFile,
	rule types.CopyRule, target types.TargetConfig, targetPath string, sourceBranch string, config *configs.Config, container *ServiceContainer) {

	sourceRepo := fmt.Sprintf("%s/%s", config.RepoOwner, config.RepoName)

	// Check if deprecation is enabled for this target
	if target.DeprecationCheck == nil || !target.DeprecationCheck.Enabled {
		return
	}

	// Add to deprecation queue
	addToDeprecationMapForTarget(targetPath, target, container.FileStateService)

	// Log deprecation event
	container.AuditLogger.LogDeprecationEvent(ctx, &AuditEvent{
		RuleName:   rule.Name,
		SourceRepo: sourceRepo,
		SourcePath: file.Path,
		TargetRepo: target.Repo,
		TargetPath: targetPath,
		CommitSHA:  sourceCommitSHA,
		PRNumber:   prNumber,
		Success:    true,
	})

	container.MetricsCollector.RecordFileDeprecated()

	LogFileOperation(ctx, "deprecate", file.Path, target.Repo, "file marked for deprecation", nil,
		map[string]interface{}{
			"target_path": targetPath,
			"rule":        rule.Name,
		})
}

// queueFileForUploadWithStrategy queues a file for upload with the appropriate strategy
func queueFileForUploadWithStrategy(target types.TargetConfig, file github.RepositoryContent,
	rule types.CopyRule, variables map[string]string, prNumber int, sourceCommitSHA string, sourceBranch string, config *configs.Config, container *ServiceContainer) {

	// Include rule name and commit strategy in the key to allow multiple rules
	// targeting the same repo/branch with different strategies
	commitStrategy := string(target.CommitStrategy.Type)
	if commitStrategy == "" {
		commitStrategy = "direct" // default
	}

	key := types.UploadKey{
		RepoName:       target.Repo,
		BranchPath:     "refs/heads/" + target.Branch,
		RuleName:       rule.Name,
		CommitStrategy: commitStrategy,
	}

	// Get existing entry or create new
	filesToUpload := container.FileStateService.GetFilesToUpload()
	entry, exists := filesToUpload[key]
	if !exists {
		entry = types.UploadFileContent{
			TargetBranch: target.Branch,
		}
	}

	// Set commit strategy
	entry.CommitStrategy = types.CommitStrategy(target.CommitStrategy.Type)
	entry.AutoMergePR = target.CommitStrategy.AutoMerge

	// Add file to content first so we can get accurate file count
	entry.Content = append(entry.Content, file)

	// Render commit message, PR title, and PR body using templates
	msgCtx := types.NewMessageContext()
	msgCtx.RuleName = rule.Name
	msgCtx.SourceRepo = fmt.Sprintf("%s/%s", config.RepoOwner, config.RepoName)
	msgCtx.SourceBranch = sourceBranch
	msgCtx.TargetRepo = target.Repo
	msgCtx.TargetBranch = target.Branch
	msgCtx.FileCount = len(entry.Content)
	msgCtx.PRNumber = prNumber
	msgCtx.CommitSHA = sourceCommitSHA
	msgCtx.Variables = variables

	if target.CommitStrategy.CommitMessage != "" {
		entry.CommitMessage = container.MessageTemplater.RenderCommitMessage(target.CommitStrategy.CommitMessage, msgCtx)
	}
	if target.CommitStrategy.PRTitle != "" {
		entry.PRTitle = container.MessageTemplater.RenderPRTitle(target.CommitStrategy.PRTitle, msgCtx)
	}
	if target.CommitStrategy.PRBody != "" {
		entry.PRBody = container.MessageTemplater.RenderPRBody(target.CommitStrategy.PRBody, msgCtx)
	}

	container.FileStateService.AddFileToUpload(key, entry)
}

// addToDeprecationMapForTarget adds a file to the deprecation map
func addToDeprecationMapForTarget(targetPath string, target types.TargetConfig, fileStateService FileStateService) {
	deprecationFile := "deprecated_examples.json"
	if target.DeprecationCheck != nil && target.DeprecationCheck.File != "" {
		deprecationFile = target.DeprecationCheck.File
	}

	entry := types.DeprecatedFileEntry{
		FileName: targetPath,
		Repo:     target.Repo,
		Branch:   target.Branch,
	}

	fileStateService.AddFileToDeprecate(deprecationFile, entry)
}
