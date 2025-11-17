package services

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/google/go-github/v48/github"
	. "github.com/mongodb/code-example-tooling/code-copier/types"
)

// WorkflowProcessor processes workflows and applies transformations
type WorkflowProcessor interface {
	ProcessWorkflow(ctx context.Context, workflow Workflow, changedFiles []ChangedFile, prNumber int, sourceCommitSHA string) error
}

// workflowProcessor implements WorkflowProcessor
type workflowProcessor struct {
	patternMatcher   PatternMatcher
	pathTransformer  PathTransformer
	fileStateService FileStateService
	metricsCollector *MetricsCollector
	messageTemplater MessageTemplater
}

// NewWorkflowProcessor creates a new workflow processor
func NewWorkflowProcessor(
	patternMatcher PatternMatcher,
	pathTransformer PathTransformer,
	fileStateService FileStateService,
	metricsCollector *MetricsCollector,
	messageTemplater MessageTemplater,
) WorkflowProcessor {
	return &workflowProcessor{
		patternMatcher:   patternMatcher,
		pathTransformer:  pathTransformer,
		fileStateService: fileStateService,
		metricsCollector: metricsCollector,
		messageTemplater: messageTemplater,
	}
}

// ProcessWorkflow processes a single workflow
func (wp *workflowProcessor) ProcessWorkflow(
	ctx context.Context,
	workflow Workflow,
	changedFiles []ChangedFile,
	prNumber int,
	sourceCommitSHA string,
) error {
	LogInfoCtx(ctx, "Processing workflow", map[string]interface{}{
		"workflow_name":   workflow.Name,
		"source_repo":     workflow.Source.Repo,
		"destination_repo": workflow.Destination.Repo,
		"file_count":      len(changedFiles),
	})

	// Track files matched and skipped
	filesMatched := 0
	filesSkipped := 0

	// Process each changed file
	for _, file := range changedFiles {
		matched, err := wp.processFileForWorkflow(ctx, workflow, file, prNumber, sourceCommitSHA)
		if err != nil {
			LogErrorCtx(ctx, "Failed to process file for workflow", err, map[string]interface{}{
				"workflow_name": workflow.Name,
				"file_path":     file.Path,
			})
			continue
		}

		if matched {
			filesMatched++
		} else {
			filesSkipped++
		}
	}

	LogInfoCtx(ctx, "Workflow processing complete", map[string]interface{}{
		"workflow_name":  workflow.Name,
		"files_matched":  filesMatched,
		"files_skipped":  filesSkipped,
	})

	return nil
}

// processFileForWorkflow processes a single file for a workflow
func (wp *workflowProcessor) processFileForWorkflow(
	ctx context.Context,
	workflow Workflow,
	file ChangedFile,
	prNumber int,
	sourceCommitSHA string,
) (bool, error) {
	// Check if file is excluded
	if wp.isExcluded(file.Path, workflow.Exclude) {
		LogInfoCtx(ctx, "File excluded by workflow exclude patterns", map[string]interface{}{
			"workflow_name": workflow.Name,
			"file_path":     file.Path,
		})
		return false, nil
	}

	// Try each transformation until one matches
	for i, transformation := range workflow.Transformations {
		matched, targetPath, err := wp.applyTransformation(ctx, workflow, transformation, file.Path)
		if err != nil {
			return false, fmt.Errorf("transformation[%d]: %w", i, err)
		}

		if !matched {
			continue
		}

		// File matched this transformation
		LogInfoCtx(ctx, "File matched transformation", map[string]interface{}{
			"workflow_name":      workflow.Name,
			"transformation_idx": i,
			"transformation_type": transformation.GetType(),
			"source_path":        file.Path,
			"target_path":        targetPath,
		})

		// Handle file based on status
		if file.Status == "removed" {
			// Add to deprecation map
			wp.addToDeprecationMap(workflow, targetPath)
		} else {
			// Add to upload queue
			err := wp.addToUploadQueue(ctx, workflow, file, targetPath, prNumber, sourceCommitSHA)
			if err != nil {
				return false, fmt.Errorf("failed to queue file for upload: %w", err)
			}
		}

		return true, nil
	}

	// No transformation matched
	LogInfoCtx(ctx, "File did not match any transformation", map[string]interface{}{
		"workflow_name": workflow.Name,
		"file_path":     file.Path,
	})

	return false, nil
}

// applyTransformation applies a transformation to a file path
func (wp *workflowProcessor) applyTransformation(
	ctx context.Context,
	workflow Workflow,
	transformation Transformation,
	sourcePath string,
) (matched bool, targetPath string, err error) {
	switch transformation.GetType() {
	case TransformationTypeMove:
		return wp.applyMoveTransformation(transformation.Move, sourcePath)
	case TransformationTypeCopy:
		return wp.applyCopyTransformation(transformation.Copy, sourcePath)
	case TransformationTypeGlob:
		return wp.applyGlobTransformation(transformation.Glob, sourcePath)
	case TransformationTypeRegex:
		return wp.applyRegexTransformation(transformation.Regex, sourcePath)
	default:
		return false, "", fmt.Errorf("unknown transformation type: %s", transformation.GetType())
	}
}

// applyMoveTransformation applies a move transformation
func (wp *workflowProcessor) applyMoveTransformation(
	move *MoveTransform,
	sourcePath string,
) (matched bool, targetPath string, err error) {
	// Check if source path starts with the "from" prefix
	from := strings.TrimSuffix(move.From, "/")
	
	if sourcePath == from {
		// Exact match - move the file to the "to" path
		return true, move.To, nil
	}
	
	if strings.HasPrefix(sourcePath, from+"/") {
		// Path is under the "from" directory - preserve relative path
		relativePath := strings.TrimPrefix(sourcePath, from+"/")
		targetPath = filepath.Join(move.To, relativePath)
		return true, targetPath, nil
	}

	return false, "", nil
}

// applyCopyTransformation applies a copy transformation
func (wp *workflowProcessor) applyCopyTransformation(
	copy *CopyTransform,
	sourcePath string,
) (matched bool, targetPath string, err error) {
	// Copy only matches exact file path
	if sourcePath == copy.From {
		return true, copy.To, nil
	}
	return false, "", nil
}

// applyGlobTransformation applies a glob transformation
func (wp *workflowProcessor) applyGlobTransformation(
	glob *GlobTransform,
	sourcePath string,
) (matched bool, targetPath string, err error) {
	// Use doublestar for glob matching
	matched, err = doublestar.Match(glob.Pattern, sourcePath)
	if err != nil {
		return false, "", fmt.Errorf("invalid glob pattern: %w", err)
	}
	if !matched {
		return false, "", nil
	}

	// Extract variables for path transformation
	variables := wp.extractGlobVariables(glob.Pattern, sourcePath)

	// Apply path transformation using the correct signature
	targetPath, err = wp.pathTransformer.Transform(sourcePath, glob.Transform, variables)
	if err != nil {
		return false, "", fmt.Errorf("path transformation failed: %w", err)
	}

	return true, targetPath, nil
}

// applyRegexTransformation applies a regex transformation
func (wp *workflowProcessor) applyRegexTransformation(
	regex *RegexTransform,
	sourcePath string,
) (matched bool, targetPath string, err error) {
	// Use existing pattern matcher for regex
	sourcePattern := SourcePattern{
		Type:    PatternTypeRegex,
		Pattern: regex.Pattern,
	}
	
	matchResult := wp.patternMatcher.Match(sourcePath, sourcePattern)
	if !matchResult.Matched {
		return false, "", nil
	}

	// Apply path transformation with captured variables
	targetPath, err = wp.pathTransformer.Transform(sourcePath, regex.Transform, matchResult.Variables)
	if err != nil {
		return false, "", fmt.Errorf("path transformation failed: %w", err)
	}

	return true, targetPath, nil
}

// extractGlobVariables extracts variables from a glob pattern match
func (wp *workflowProcessor) extractGlobVariables(pattern, path string) map[string]string {
	variables := make(map[string]string)
	
	// Extract common variables
	// For pattern "mflix/server/**" matching "mflix/server/java-spring/src/main.java"
	// Extract relative_path = "java-spring/src/main.java"
	
	// Find the ** in the pattern
	starStarIdx := strings.Index(pattern, "**")
	if starStarIdx >= 0 {
		prefix := pattern[:starStarIdx]
		if strings.HasPrefix(path, prefix) {
			relativePath := strings.TrimPrefix(path, prefix)
			relativePath = strings.TrimPrefix(relativePath, "/")
			variables["relative_path"] = relativePath
		}
	}
	
	return variables
}

// isExcluded checks if a file path matches any exclude pattern
func (wp *workflowProcessor) isExcluded(path string, excludePatterns []string) bool {
	for _, pattern := range excludePatterns {
		matched, err := doublestar.Match(pattern, path)
		if err != nil {
			LogWarning(fmt.Sprintf("Invalid exclude pattern: %s: %v", pattern, err))
			continue
		}
		if matched {
			return true
		}
	}
	return false
}

// addToDeprecationMap adds a file to the deprecation map
func (wp *workflowProcessor) addToDeprecationMap(workflow Workflow, targetPath string) {
	deprecationFile := "deprecated_examples.json"
	if workflow.DeprecationCheck != nil && workflow.DeprecationCheck.File != "" {
		deprecationFile = workflow.DeprecationCheck.File
	}

	entry := DeprecatedFileEntry{
		FileName: targetPath,
		Repo:     workflow.Destination.Repo,
		Branch:   workflow.Destination.Branch,
	}

	wp.fileStateService.AddFileToDeprecate(deprecationFile, entry)
}

// addToUploadQueue adds a file to the upload queue
func (wp *workflowProcessor) addToUploadQueue(
	ctx context.Context,
	workflow Workflow,
	file ChangedFile,
	targetPath string,
	prNumber int,
	sourceCommitSHA string,
) error {
	// Parse source repo owner/name
	parts := strings.Split(workflow.Source.Repo, "/")
	if len(parts) != 2 {
		return fmt.Errorf("invalid source repo format: expected owner/repo, got: %s", workflow.Source.Repo)
	}
	sourceRepoOwner := parts[0]
	sourceRepoName := parts[1]

	// Fetch file content from source repository
	fileContent, err := RetrieveFileContentsWithConfigAndBranch(ctx, file.Path, sourceCommitSHA, sourceRepoOwner, sourceRepoName)
	if err != nil {
		return fmt.Errorf("failed to retrieve file content: %w", err)
	}

	// Update file name to target path
	fileContent.Name = github.String(targetPath)

	// Create upload key
	key := UploadKey{
		RepoName:   workflow.Destination.Repo,
		BranchPath: workflow.Destination.Branch,
	}

	// Get existing entries from FileStateService
	filesToUpload := wp.fileStateService.GetFilesToUpload()
	content, exists := filesToUpload[key]
	if !exists {
		content = UploadFileContent{
			Content:        []github.RepositoryContent{},
			CommitStrategy: CommitStrategy(getCommitStrategyType(workflow)),
			UsePRTemplate:  getUsePRTemplate(workflow),
			AutoMergePR:    getAutoMerge(workflow),
		}
	}

	// Add file to content
	content.Content = append(content.Content, *fileContent)

	// Render templates with message context
	msgCtx := NewMessageContext()
	msgCtx.SourceRepo = workflow.Source.Repo
	msgCtx.SourceBranch = workflow.Source.Branch
	msgCtx.TargetRepo = workflow.Destination.Repo
	msgCtx.TargetBranch = workflow.Destination.Branch
	msgCtx.PRNumber = prNumber
	msgCtx.CommitSHA = sourceCommitSHA
	msgCtx.FileCount = len(content.Content)

	// Render commit message
	if workflow.CommitStrategy != nil && workflow.CommitStrategy.CommitMessage != "" {
		content.CommitMessage = wp.messageTemplater.RenderCommitMessage(workflow.CommitStrategy.CommitMessage, msgCtx)
	} else {
		content.CommitMessage = fmt.Sprintf("Update from workflow: %s", workflow.Name)
	}

	// Render PR title
	if workflow.CommitStrategy != nil && workflow.CommitStrategy.PRTitle != "" {
		content.PRTitle = wp.messageTemplater.RenderPRTitle(workflow.CommitStrategy.PRTitle, msgCtx)
	} else {
		content.PRTitle = content.CommitMessage
	}

	// Render PR body
	if workflow.CommitStrategy != nil && workflow.CommitStrategy.PRBody != "" {
		content.PRBody = wp.messageTemplater.RenderPRBody(workflow.CommitStrategy.PRBody, msgCtx)
	}

	// Add back to FileStateService
	wp.fileStateService.AddFileToUpload(key, content)

	// Record metric (with zero duration since we're just queuing)
	if wp.metricsCollector != nil {
		wp.metricsCollector.RecordFileUploaded(0 * time.Second)
	}

	return nil
}

// Helper functions to extract config values

func getCommitStrategyType(workflow Workflow) string {
	if workflow.CommitStrategy != nil && workflow.CommitStrategy.Type != "" {
		return workflow.CommitStrategy.Type
	}
	return "pull_request" // default
}

func getCommitMessage(workflow Workflow) string {
	if workflow.CommitStrategy != nil && workflow.CommitStrategy.CommitMessage != "" {
		return workflow.CommitStrategy.CommitMessage
	}
	return fmt.Sprintf("Update from workflow: %s", workflow.Name)
}

func getPRTitle(workflow Workflow) string {
	if workflow.CommitStrategy != nil && workflow.CommitStrategy.PRTitle != "" {
		return workflow.CommitStrategy.PRTitle
	}
	return getCommitMessage(workflow)
}

func getPRBody(workflow Workflow) string {
	if workflow.CommitStrategy != nil && workflow.CommitStrategy.PRBody != "" {
		return workflow.CommitStrategy.PRBody
	}
	return ""
}

func getUsePRTemplate(workflow Workflow) bool {
	if workflow.CommitStrategy != nil {
		return workflow.CommitStrategy.UsePRTemplate
	}
	return false
}

func getAutoMerge(workflow Workflow) bool {
	if workflow.CommitStrategy != nil {
		return workflow.CommitStrategy.AutoMerge
	}
	return false
}
