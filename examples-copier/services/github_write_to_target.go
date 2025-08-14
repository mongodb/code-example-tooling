package services

import (
	"context"
	"fmt"
	"github.com/google/go-github/v48/github"
	"github.com/mongodb/code-example-tooling/code-copier/configs"
	. "github.com/mongodb/code-example-tooling/code-copier/types"
	"github.com/pkg/errors"
	"log"
	"os"
	"strings"
	"time"
)

// FilesToUpload is a map where the key is the repo name
// and the value is of type [UploadFileContent], which
// contains the target branch name and the collection of files
// to be uploaded.
var FilesToUpload map[UploadKey]UploadFileContent
var FilesToDeprecate map[string]Configs

// commitStrategy returns the commit strategy based on the environment variable.
// Supported values are "direct" (commits directly to the target branch) and "pr" (creates a pull request - default).
// If an unsupported value is provided, it defaults to "pr".
func commitStrategy(c Configs) string {
	switch v := c.CopierCommitStrategy; v {
	case "direct", "pr":
		return v
	}
	return "pr"
}

// findConfig returns the first entry matching repoName or zero-value
func findConfig(cfgs ConfigFileType, repoName string) Configs {
	for _, c := range cfgs {
		if c.TargetRepo == repoName {
			return c
		}
	}
	return Configs{}
}

// repoOwner returns the repository owner from environment variables.
func repoOwner() string { return os.Getenv(configs.RepoOwner) }

// AddFilesToTargetRepoBranch uploads files to the target repository branch
// using the specified commit strategy (direct or via pull request).
func AddFilesToTargetRepoBranch(cfgs ConfigFileType) {
	ctx := GlobalContext.GetContext()
	client := GetRestClient()

	for key, value := range FilesToUpload {
		cfg := findConfig(cfgs, key.RepoName)
		switch commitStrategy(cfg) {
		case "direct": // commits directly to the target branch
			LogInfo(fmt.Sprintf("Using direct commit strategy for %s on branch %s", key.RepoName, key.BranchPath))
			if err := addFilesToBranch(ctx, client, key, value.Content, "Add multiple files"); err != nil {
				LogCritical(fmt.Sprintf("Failed to add files to target branch: %v\n", err))
			}
		default: // "pr" or any other value defaults to PR strategy
			if err := addFilesViaPR(ctx, client, key, value.Content, "Add multiple files", cfg.MergeWithoutReview); err != nil {
				LogCritical(fmt.Sprintf("Failed via PR path: %v\n", err))
			}
		}
	}
}

// createPullRequest opens a pull request from head to base in the specified repository.
func createPullRequest(ctx context.Context, client *github.Client, repo, head, base, title, body string) (*github.PullRequest, error) {
	owner := repoOwner()
	pr := &github.NewPullRequest{
		Title: github.String(title),
		Head:  github.String(head), // for same-repo branches, just "branch"; for forks, use "owner:branch"
		Base:  github.String(base), // e.g. "main"
		Body:  github.String(body),
	}
	created, _, err := client.PullRequests.Create(ctx, owner, repo, pr)
	if err != nil {
		return nil, fmt.Errorf("could not create PR: %w", err)
	}
	return created, nil
}

// addFilesViaPR creates a temporary branch, commits files to it, opens a pull request, and optionally merges it automatically.
func addFilesViaPR(ctx context.Context, client *github.Client, key UploadKey,
	files []github.RepositoryContent, message string, mergeWithoutReview bool,
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
	if err = createCommit(ctx, client, tempKey, baseSHA, treeSHA, message); err != nil {
		return fmt.Errorf("create commit on temp branch: %w", err)
	}

	// 3) Create PR from temp branch to base branch
	base := strings.TrimPrefix(key.BranchPath, "refs/heads/")
	pr, err := createPullRequest(ctx, client, key.RepoName, tempBranch, base, message, "")
	if err != nil {
		return fmt.Errorf("create PR: %w", err)
	}

	// 4) Optionally merge the PR without review if MergeWithoutReview is true
	LogInfo(fmt.Sprintf("PR created: #%d from %s to %s", pr.GetNumber(), tempBranch, base))
	LogInfo(fmt.Sprintf("PR URL: %s", pr.GetHTMLURL()))
	if mergeWithoutReview {
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
	owner := repoOwner()

	// Use provided base branch or default to "main"
	base := "main"
	if len(baseBranch) > 0 && baseBranch[0] != "" {
		base = baseBranch[0]
	}

	baseRef, _, err := client.Git.GetRef(ctx, owner, repo, "refs/heads/"+base)
	if err != nil {
		LogCritical(fmt.Sprintf("Failed to get '%s' baseRef: %s", base, err))
		return nil, err
	}

	// *** Check if branch (newBranchRef) already exists and delete it ***
	newBranchRef, _, err := client.Git.GetRef(ctx, owner, repo, fmt.Sprintf("%s%s", "refs/heads/", newBranch))
	deleteBranchIfExists(ctx, client, repo, newBranchRef)

	newRef := &github.Reference{
		Ref: github.String(fmt.Sprintf("%s%s", "refs/heads/", newBranch)),
		Object: &github.GitObject{
			SHA: baseRef.Object.SHA,
		},
	}

	newBranchRef, _, err = client.Git.CreateRef(ctx, owner, repo, newRef)
	if err != nil {
		LogCritical(fmt.Sprintf("Failed to create newBranchRef %s:  %s", newRef, err))
		return nil, err
	}

	LogInfo(fmt.Sprintf("Branch created successfully: %s on %s (from %s)", newRef, repo, base))

	return newBranchRef, nil
}

// createCommitTree looks up the branch ref once, then builds a tree on top of that base commit.
func createCommitTree(ctx context.Context, client *github.Client, targetBranch UploadKey,
	files map[string]string) (treeSHA string, baseSHA string, err error) {

	owner := repoOwner()

	// 1) Get current ref (ONE GET)
	ref, _, err := client.Git.GetRef(ctx, owner, targetBranch.RepoName, targetBranch.BranchPath)
	if err != nil || ref == nil {
		if err == nil {
			err = errors.Errorf("targetRef is nil")
		}
		LogCritical(fmt.Sprintf("Failed to get ref for %s: %v\n", targetBranch.RepoName, err))
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
	tree, _, err := client.Git.CreateTree(ctx, owner, targetBranch.RepoName, baseSHA, treeEntries)
	if err != nil {
		return "", "", fmt.Errorf("failed to create tree: %w", err)
	}
	return tree.GetSHA(), baseSHA, nil
}

// createCommit makes the commit using the provided baseSHA, and updates the branch ref to the new commit.
func createCommit(ctx context.Context, client *github.Client, targetBranch UploadKey,
	baseSHA string, treeSHA string, message string) error {

	owner := repoOwner()

	parent := &github.Commit{SHA: github.String(baseSHA)}
	commit := &github.Commit{
		Message: github.String(message),
		Tree:    &github.Tree{SHA: github.String(treeSHA)},
		Parents: []*github.Commit{parent},
	}

	newCommit, _, err := client.Git.CreateCommit(ctx, owner, targetBranch.RepoName, commit)
	if err != nil {
		return fmt.Errorf("could not create commit: %w", err)
	}

	// Update branch ref directly (no second GET)
	ref := &github.Reference{
		Ref:    github.String(targetBranch.BranchPath), // e.g., "refs/heads/main"
		Object: &github.GitObject{SHA: github.String(newCommit.GetSHA())},
	}
	if _, _, err := client.Git.UpdateRef(ctx, owner, targetBranch.RepoName, ref, false); err != nil {
		return fmt.Errorf("failed to update ref to new commit: %w", err)
	}
	return nil
}

// mergePR merges the specified pull request in the given repository.
func mergePR(ctx context.Context, client *github.Client, repo string, prNumber int) error {
	owner := repoOwner()

	options := &github.PullRequestOptions{
		MergeMethod: "merge", // Other options: "squash" or "rebase"
	}
	result, _, err := client.PullRequests.Merge(ctx, owner, repo, prNumber, "Merging the pull request", options)
	if err != nil {
		LogCritical(fmt.Sprintf("Failed to merge PR: %v\n", err))
		return err
	}
	if result.GetMerged() {
		LogInfo(fmt.Sprintf("Successfully merged PR #%d\n", prNumber))
		return nil
	} else {
		LogError(fmt.Sprintf("Failed to merge PR #%d: %s", prNumber, result.GetMessage()))
		return fmt.Errorf("failed to merge PR #%d: %s", prNumber, result.GetMessage())
	}
}

// deleteBranchIfExists deletes the specified branch if it exists, except for 'main'.
func deleteBranchIfExists(backgroundContext context.Context, client *github.Client, repo string, ref *github.Reference) {

	owner := repoOwner()
	if ref.GetRef() == "refs/heads/main" {
		LogError("I refuse to delete branch 'main'.")
		log.Fatal()
	}

	LogInfo(fmt.Sprintf("Deleting branch %s on %s", ref.GetRef(), repo))
	_, _, err := client.Git.GetRef(backgroundContext, owner, repo, ref.GetRef())

	if err == nil { // Branch exists (there was no error fetching it)
		_, err = client.Git.DeleteRef(backgroundContext, owner, repo, ref.GetRef())
		if err != nil {
			LogCritical(fmt.Sprintf("Error deleting branch: %v\n", err))
		}
	}
}
