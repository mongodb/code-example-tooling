package services

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/go-github/v48/github"
	"github.com/mongodb/code-example-tooling/code-copier/configs"
	. "github.com/mongodb/code-example-tooling/code-copier/types"
	"github.com/pkg/errors"
)

// FilesToUpload is a map where the key is the repo name
// and the value is of type [UploadFileContent], which
// contains the target branch name and the collection of files
// to be uploaded.
var FilesToUpload map[UploadKey]UploadFileContent
var FilesToDeprecate map[string]Configs



// repoOwner returns the repository owner from environment variables.
func repoOwner() string { return os.Getenv(configs.RepoOwner) }

// parseRepoPath parses a repository path in the format "owner/repo" and returns owner and repo separately.
// If the path doesn't contain a slash, it returns the source repo owner from env and the path as repo name.
func parseRepoPath(repoPath string) (owner, repo string) {
	parts := strings.Split(repoPath, "/")
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	// Fallback to source repo owner if no slash found (backward compatibility)
	return repoOwner(), repoPath
}

// normalizeRepoName ensures a repository name includes the owner prefix.
// If the repo name already has an owner (contains "/"), returns it as-is.
// Otherwise, prepends the default repo owner from environment.
func normalizeRepoName(repoName string) string {
	if strings.Contains(repoName, "/") {
		return repoName
	}
	return repoOwner() + "/" + repoName
}

// AddFilesToTargetRepoBranch uploads files to the target repository branch
// using the specified commit strategy (direct or via pull request).
func AddFilesToTargetRepoBranch() {
	AddFilesToTargetRepoBranchWithFetcher(nil, nil)
}

// AddFilesToTargetRepoBranchWithFetcher uploads files to the target repository branch
// using the specified commit strategy (direct or via pull request).
// If prTemplateFetcher is provided, it will be used to fetch PR templates when use_pr_template is true.
// If metricsCollector is provided, it will be used to record upload failures.
func AddFilesToTargetRepoBranchWithFetcher(prTemplateFetcher PRTemplateFetcher, metricsCollector *MetricsCollector) {
	ctx := context.Background()

	for key, value := range FilesToUpload {
		// Parse the repository to get the organization
		owner, _ := parseRepoPath(key.RepoName)

		// Get a client authenticated for this organization
		client, err := GetRestClientForOrg(owner)
		if err != nil {
			LogCritical(fmt.Sprintf("Failed to get GitHub client for org %s: %v", owner, err))
			// Record failure for each file in this batch
			if metricsCollector != nil {
				for range value.Content {
					metricsCollector.RecordFileUploadFailed()
				}
			}
			continue
		}

		// Determine commit strategy from value (set by pattern-matching system)
		strategy := string(value.CommitStrategy)
		if strategy == "" {
			strategy = "direct" // default
		}

		// Get commit message from value or use default
		commitMsg := value.CommitMessage
		if strings.TrimSpace(commitMsg) == "" {
			commitMsg = os.Getenv(configs.DefaultCommitMessage)
			if strings.TrimSpace(commitMsg) == "" {
				commitMsg = configs.NewConfig().DefaultCommitMessage
			}
		}

		// Get PR title from value or use commit message
		prTitle := value.PRTitle
		if strings.TrimSpace(prTitle) == "" {
			prTitle = commitMsg
		}

		// Get PR body from value
		prBody := value.PRBody

		// Fetch and merge PR template if requested
		if value.UsePRTemplate && prTemplateFetcher != nil && strategy != "direct" {
			targetBranch := strings.TrimPrefix(key.BranchPath, "refs/heads/")
			template, err := prTemplateFetcher.FetchPRTemplate(ctx, client, key.RepoName, targetBranch)
			if err != nil {
				LogWarning(fmt.Sprintf("Failed to fetch PR template for %s: %v", key.RepoName, err))
			} else if template != "" {
				// Merge configured body with template
				prBody = MergePRBodyWithTemplate(prBody, template)
				LogInfo(fmt.Sprintf("Merged PR template for %s", key.RepoName))
			}
		}

		// Get auto-merge setting from value
		mergeWithoutReview := value.AutoMergePR

		switch strategy {
		case "direct": // commits directly to the target branch
			LogInfo(fmt.Sprintf("Using direct commit strategy for %s on branch %s", key.RepoName, key.BranchPath))
			if err := addFilesToBranch(ctx, client, key, value.Content, commitMsg); err != nil {
				LogCritical(fmt.Sprintf("Failed to add files to target branch: %v\n", err))
				// Record failure for each file in this batch
				if metricsCollector != nil {
					for range value.Content {
						metricsCollector.RecordFileUploadFailed()
					}
				}
			}
		default: // "pr" or "pull_request" strategy
			LogInfo(fmt.Sprintf("Using PR commit strategy for %s on branch %s (auto_merge=%v)", key.RepoName, key.BranchPath, mergeWithoutReview))
			if err := addFilesViaPR(ctx, client, key, value.Content, commitMsg, prTitle, prBody, mergeWithoutReview); err != nil {
				LogCritical(fmt.Sprintf("Failed via PR path: %v\n", err))
				// Record failure for each file in this batch
				if metricsCollector != nil {
					for range value.Content {
						metricsCollector.RecordFileUploadFailed()
					}
				}
			}
		}
	}
}

// createPullRequest opens a pull request from head to base in the specified repository.
func createPullRequest(ctx context.Context, client *github.Client, repo, head, base, title, body string) (*github.PullRequest, error) {
	owner, repoName := parseRepoPath(repo)
	pr := &github.NewPullRequest{
		Title: github.String(title),
		Head:  github.String(head), // for same-repo branches, just "branch"; for forks, use "owner:branch"
		Base:  github.String(base), // e.g. "main"
		Body:  github.String(body),
	}
	created, _, err := client.PullRequests.Create(ctx, owner, repoName, pr)
	if err != nil {
		return nil, fmt.Errorf("could not create PR: %w", err)
	}
	return created, nil
}

// addFilesViaPR creates a temporary branch, commits files to it using the provided commitMessage,
// opens a pull request with prTitle and prBody, and optionally merges it automatically.
func addFilesViaPR(ctx context.Context, client *github.Client, key UploadKey,
	files []github.RepositoryContent, commitMessage string, prTitle string, prBody string, mergeWithoutReview bool,
) error {
	tempBranch := "copier/" + time.Now().UTC().Format("20060102-150405")

	// 1) Create branch off the target branch specified in key.BranchPath or default to "main"
	baseBranch := strings.TrimPrefix(key.BranchPath, "refs/heads/")
	newRef, err := createBranch(ctx, client, key.RepoName, tempBranch, baseBranch)
	if err != nil {
		return fmt.Errorf("create branch: %w", err)
	}
	_ = newRef // we just need it created; ref is not reused directly

	// 2) Commit files to temp branch
	entries := make(map[string]string, len(files))
	for _, f := range files {
		content, _ := f.GetContent()
		entries[f.GetName()] = content
	}

	tempKey := UploadKey{RepoName: key.RepoName, BranchPath: "refs/heads/" + tempBranch}
	treeSHA, baseSHA, err := createCommitTree(ctx, client, tempKey, entries)
	if err != nil {
		return fmt.Errorf("create tree on temp branch: %w", err)
	}
	if err = createCommit(ctx, client, tempKey, baseSHA, treeSHA, commitMessage); err != nil {
		return fmt.Errorf("create commit on temp branch: %w", err)
	}

	// 3) Create PR from temp branch to base branch
	base := strings.TrimPrefix(key.BranchPath, "refs/heads/")
	pr, err := createPullRequest(ctx, client, key.RepoName, tempBranch, base, prTitle, prBody)
	if err != nil {
		return fmt.Errorf("create PR: %w", err)
	}

	// 4) Optionally merge the PR without review if MergeWithoutReview is true
	LogInfo(fmt.Sprintf("PR created: #%d from %s to %s", pr.GetNumber(), tempBranch, base))
	LogInfo(fmt.Sprintf("PR URL: %s", pr.GetHTMLURL()))
	if mergeWithoutReview {
		// Poll PR for mergeability; GitHub may take a moment to compute it
		// Get polling configuration from environment or use defaults
		cfg := configs.NewConfig()
		maxAttempts := cfg.PRMergePollMaxAttempts
		if envAttempts := os.Getenv(configs.PRMergePollMaxAttempts); envAttempts != "" {
			if parsed, err := parseIntWithDefault(envAttempts, maxAttempts); err == nil {
				maxAttempts = parsed
			}
		}

		pollInterval := cfg.PRMergePollInterval
		if envInterval := os.Getenv(configs.PRMergePollInterval); envInterval != "" {
			if parsed, err := parseIntWithDefault(envInterval, pollInterval); err == nil {
				pollInterval = parsed
			}
		}

		var mergeable *bool
		var mergeableState string
		owner, repoName := parseRepoPath(key.RepoName)
		for i := 0; i < maxAttempts; i++ {
			current, _, gerr := client.PullRequests.Get(ctx, owner, repoName, pr.GetNumber())
			if gerr == nil && current != nil {
				mergeable = current.Mergeable
				mergeableState = current.GetMergeableState()
				if mergeable != nil { // computed
					break
				}
			}
			time.Sleep(time.Duration(pollInterval) * time.Millisecond)
		}
		if mergeable != nil && !*mergeable || strings.EqualFold(mergeableState, "dirty") {
			LogWarning(fmt.Sprintf("PR #%d is not mergeable (state=%s). Likely merge conflicts. Leaving PR open for manual resolution.", pr.GetNumber(), mergeableState))
			return fmt.Errorf("pull request #%d has merge conflicts (state=%s)", pr.GetNumber(), mergeableState)
		}
		if err = mergePR(ctx, client, key.RepoName, pr.GetNumber()); err != nil {
			return fmt.Errorf("merge PR: %w", err)
		}
		deleteBranchIfExists(ctx, client, key.RepoName, &github.Reference{Ref: github.String("refs/heads/" + tempBranch)})
	} else {
		LogInfo(fmt.Sprintf("PR created and awaiting review: #%d", pr.GetNumber()))
	}
	return nil
}

// addFilesToBranch builds a tree, creates a commit, and updates the ref (direct to target branch)
func addFilesToBranch(ctx context.Context, client *github.Client, key UploadKey,
	files []github.RepositoryContent, message string) error {

	entries := make(map[string]string, len(files))
	for _, f := range files {
		content, _ := f.GetContent()
		entries[f.GetName()] = content
	}

	treeSHA, baseSHA, err := createCommitTree(ctx, client, key, entries)
	if err != nil {
		LogCritical(fmt.Sprintf("Error creating commit tree: %v\n", err))
		return err
	}
	if err := createCommit(ctx, client, key, baseSHA, treeSHA, message); err != nil {
		LogCritical(fmt.Sprintf("Error creating commit: %v\n", err))
		return err
	}
	return nil
}

// createBranch creates a new branch from the specified base branch (defaults to 'main') and deletes it first if it already exists.
func createBranch(ctx context.Context, client *github.Client, repo, newBranch string, baseBranch ...string) (*github.Reference, error) {
	// Normalize repo name for consistent logging and operations
	normalizedRepo := normalizeRepoName(repo)
	owner, repoName := parseRepoPath(normalizedRepo)

	// Use provided base branch or default to "main"
	base := "main"
	if len(baseBranch) > 0 && baseBranch[0] != "" {
		base = baseBranch[0]
	}

	baseRef, _, err := client.Git.GetRef(ctx, owner, repoName, "refs/heads/"+base)
	if err != nil {
		LogCritical(fmt.Sprintf("Failed to get '%s' baseRef: %s", base, err))
		return nil, err
	}

	// *** Check if branch (newBranchRef) already exists and delete it ***
	newBranchRef, _, err := client.Git.GetRef(ctx, owner, repoName, fmt.Sprintf("%s%s", "refs/heads/", newBranch))
	deleteBranchIfExists(ctx, client, normalizedRepo, newBranchRef)

	newRef := &github.Reference{
		Ref: github.String(fmt.Sprintf("%s%s", "refs/heads/", newBranch)),
		Object: &github.GitObject{
			SHA: baseRef.Object.SHA,
		},
	}

	newBranchRef, _, err = client.Git.CreateRef(ctx, owner, repoName, newRef)
	if err != nil {
		LogCritical(fmt.Sprintf("Failed to create newBranchRef %s:  %s", newRef, err))
		return nil, err
	}

	LogInfo(fmt.Sprintf("Branch created successfully: %s on %s (from %s)", newRef, normalizedRepo, base))

	return newBranchRef, nil
}

// createCommitTree looks up the branch ref once, then builds a tree on top of that base commit.
func createCommitTree(ctx context.Context, client *github.Client, targetBranch UploadKey,
	files map[string]string) (treeSHA string, baseSHA string, err error) {

	// Normalize repo name for consistent logging
	normalizedRepo := normalizeRepoName(targetBranch.RepoName)
	owner, repoName := parseRepoPath(normalizedRepo)
	LogInfo(fmt.Sprintf("DEBUG createCommitTree: targetBranch.RepoName=%q, normalized=%q, parsed owner=%q, repoName=%q",
		targetBranch.RepoName, normalizedRepo, owner, repoName))

	// 1) Get current ref with retry logic to handle GitHub API eventual consistency
	// When a branch is just created, it may take a moment to be visible
	var ref *github.Reference

	// Get retry configuration from environment or use defaults
	cfg := configs.NewConfig()
	maxRetries := cfg.GitHubAPIMaxRetries
	if envRetries := os.Getenv(configs.GitHubAPIMaxRetries); envRetries != "" {
		if parsed, err := parseIntWithDefault(envRetries, maxRetries); err == nil {
			maxRetries = parsed
		}
	}

	initialRetryDelay := cfg.GitHubAPIInitialRetryDelay
	if envDelay := os.Getenv(configs.GitHubAPIInitialRetryDelay); envDelay != "" {
		if parsed, err := parseIntWithDefault(envDelay, initialRetryDelay); err == nil {
			initialRetryDelay = parsed
		}
	}

	retryDelay := time.Duration(initialRetryDelay) * time.Millisecond

	for attempt := 1; attempt <= maxRetries; attempt++ {
		ref, _, err = client.Git.GetRef(ctx, owner, repoName, targetBranch.BranchPath)
		if err == nil && ref != nil {
			break // Success
		}

		if attempt < maxRetries {
			LogWarning(fmt.Sprintf("Failed to get ref for %s (attempt %d/%d): %v. Retrying in %v...",
				normalizedRepo, attempt, maxRetries, err, retryDelay))
			time.Sleep(retryDelay)
			retryDelay *= 2 // Exponential backoff
		}
	}

	if err != nil || ref == nil {
		if err == nil {
			err = errors.Errorf("targetRef is nil after %d attempts", maxRetries)
		}
		LogCritical(fmt.Sprintf("Failed to get ref for %s after %d attempts: %v\n", normalizedRepo, maxRetries, err))
		return "", "", err
	}
	baseSHA = ref.GetObject().GetSHA()

	// 2) Build tree entries
	var treeEntries []*github.TreeEntry
	for path, content := range files {
		treeEntries = append(treeEntries, &github.TreeEntry{
			Path:    github.String(path),
			Type:    github.String("blob"),
			Mode:    github.String("100644"),
			Content: github.String(content),
		})
	}

	// 3) Create tree on top of baseSHA
	tree, _, err := client.Git.CreateTree(ctx, owner, repoName, baseSHA, treeEntries)
	if err != nil {
		return "", "", fmt.Errorf("failed to create tree: %w", err)
	}
	return tree.GetSHA(), baseSHA, nil
}

// createCommit makes the commit using the provided baseSHA, and updates the branch ref to the new commit.
func createCommit(ctx context.Context, client *github.Client, targetBranch UploadKey,
	baseSHA string, treeSHA string, message string) error {

	owner, repoName := parseRepoPath(targetBranch.RepoName)

	parent := &github.Commit{SHA: github.String(baseSHA)}
	commit := &github.Commit{
		Message: github.String(message),
		Tree:    &github.Tree{SHA: github.String(treeSHA)},
		Parents: []*github.Commit{parent},
	}

	newCommit, _, err := client.Git.CreateCommit(ctx, owner, repoName, commit)
	if err != nil {
		return fmt.Errorf("could not create commit: %w", err)
	}

	// Update branch ref directly (no second GET)
	ref := &github.Reference{
		Ref:    github.String(targetBranch.BranchPath), // e.g., "refs/heads/main"
		Object: &github.GitObject{SHA: github.String(newCommit.GetSHA())},
	}
	if _, _, err := client.Git.UpdateRef(ctx, owner, repoName, ref, false); err != nil {
		// Detect non-fast-forward / conflict scenarios and provide a clearer error
		if eresp, ok := err.(*github.ErrorResponse); ok {
			if eresp.Response != nil && eresp.Response.StatusCode == http.StatusUnprocessableEntity {
				return fmt.Errorf("failed to update ref: non-fast-forward (possible conflict). Consider using PR strategy: %w", err)
			}
		}
		return fmt.Errorf("failed to update ref to new commit: %w", err)
	}
	return nil
}

// mergePR merges the specified pull request in the given repository.
func mergePR(ctx context.Context, client *github.Client, repo string, pr_number int) error {
	owner, repoName := parseRepoPath(repo)

	options := &github.PullRequestOptions{
		MergeMethod: "merge", // Other options: "squash" or "rebase"
	}
	result, _, err := client.PullRequests.Merge(ctx, owner, repoName, pr_number, "Merging the pull request", options)
	if err != nil {
		LogCritical(fmt.Sprintf("Failed to merge PR: %v\n", err))
		return err
	}
	if result.GetMerged() {
		LogInfo(fmt.Sprintf("Successfully merged PR #%d\n", pr_number))
		return nil
	} else {
		LogError(fmt.Sprintf("Failed to merge PR #%d: %s", pr_number, result.GetMessage()))
		return fmt.Errorf("failed to merge PR #%d: %s", pr_number, result.GetMessage())
	}
}

// deleteBranchIfExists deletes the specified branch if it exists, except for 'main'.
func deleteBranchIfExists(backgroundContext context.Context, client *github.Client, repo string, ref *github.Reference) {
	// Early return if ref is nil (branch doesn't exist)
	if ref == nil {
		return
	}

	// Normalize repo name for consistent logging
	normalizedRepo := normalizeRepoName(repo)
	owner, repoName := parseRepoPath(normalizedRepo)

	if ref.GetRef() == "refs/heads/main" {
		LogError("I refuse to delete branch 'main'.")
		log.Fatal()
	}

	LogInfo(fmt.Sprintf("Deleting branch %s on %s", ref.GetRef(), normalizedRepo))
	_, _, err := client.Git.GetRef(backgroundContext, owner, repoName, ref.GetRef())

	if err == nil { // Branch exists (there was no error fetching it)
		_, err = client.Git.DeleteRef(backgroundContext, owner, repoName, ref.GetRef())
		if err != nil {
			LogCritical(fmt.Sprintf("Error deleting branch: %v\n", err))
		}
	}
}

// DeleteBranchIfExistsExported is an exported wrapper for testing deleteBranchIfExists
func DeleteBranchIfExistsExported(ctx context.Context, client *github.Client, repo string, ref *github.Reference) {
	deleteBranchIfExists(ctx, client, repo, ref)
}

// parseIntWithDefault parses a string to int, returning defaultValue on error
func parseIntWithDefault(s string, defaultValue int) (int, error) {
	var result int
	if _, err := fmt.Sscanf(s, "%d", &result); err != nil {
		return defaultValue, err
	}
	return result, nil
}
