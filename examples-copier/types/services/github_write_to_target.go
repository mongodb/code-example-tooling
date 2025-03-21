package services

import (
	"context"
	"fmt"
	"github.com/google/go-github/v48/github"
	"github.com/pkg/errors"
	"github.com/thompsch/app-tester/configs"
	. "github.com/thompsch/app-tester/types"
)

// FilesToUpload is a map where the key is the repo name
// and the value is of type [UploadFileContent], which
// contains the target branch name and the collection of files
// to be uploaded.
var FilesToUpload map[UploadKey]UploadFileContent
var FilesToDeprecate map[string]Configs

// AddFilesToTargetRepoBranch adds new and modified files directly to the
// target branch without creating a new branch or PR. If you want more
// control over the process, use [AddFilesToTargetRepoViaPR] instead.
// ** IMPORTANT ** This needs to be thoroughly tested, since it has
// not been updated since other changes.
func AddFilesToTargetRepoBranch() {
	ctx := context.Background()
	client := GetRestClient()

	// *** For each repo ***
	for key, value := range FilesToUpload {
		// *** Create tree on target branch with all new/changed files ***
		treeSHA, err := addFilesToBranch(client, key, value.Content)
		if err != nil {
			LogCritical(fmt.Sprintf("Failed to add files to target branch: %v\n", err))
			continue
		}

		err = createCommit(ctx, client, key, treeSHA, "Add multiple files")
		if err != nil {
			LogCritical(fmt.Sprintf("Error creating commit: %v\n", err))
			continue
		}
	}
}

func AddFilesToTargetRepoViaPR() {
	ctx := context.Background()
	client := GetRestClient()

	// *** For each repo ***
	for key, value := range FilesToUpload {

		// *** Create a new Branch ***
		newBranchRef, nbe := createBranch(ctx, client, key.RepoName, value.TargetBranch)
		if nbe != nil {
			LogCritical(nbe.Error())
			return
		}

		// *** Create tree with multiple files ***
		treeSHA, err := addFilesToBranch(client, key, value.Content)
		if err != nil {
			LogCritical(fmt.Sprintf("Failed to add files to branch: %v\n", err))
			return
		}

		// *** Create a commit pointing to this tree ***
		err = createCommit(ctx, client, key, treeSHA, "Add multiple files")
		if err != nil {
			LogCritical(fmt.Sprintf("Error creating commit: %v\n", err))
			return
		}

		// *** Create a pull request ***
		pr, err := createPullRequest(ctx, client, key.RepoName, "temp_feature",
			"main", "Code Copy PR", "Adding multiple files to commit.")
		if err != nil {
			LogCritical(fmt.Sprintf("Error creating pull request: %v\n", err))
			return
		}
		LogInfo(fmt.Sprintf("Pull Request #%d created\n", pr.GetNumber()))

		// *** Merge PR ***
		mergeError := mergePR(ctx, client, key.RepoName, pr.GetNumber())
		if mergeError == nil {
			// *** Delete branch ***
			deleteBranchIfExists(ctx, client, key.RepoName, newBranchRef)
		}
	}
}

func addFilesToBranch(client *github.Client, key UploadKey,
	files []github.RepositoryContent) (*string, error) {

	entries := make(map[string]string)

	for _, file := range files {
		fileContent, _ := file.GetContent()
		entries[*file.Name] = fileContent
	}

	treeSHA, err := createCommitTree(ctx, client, key, entries)
	if err != nil {
		LogCritical(fmt.Sprintf("Error creating commit tree: %v\n", err))
		return nil, err
	}
	return &treeSHA, nil
}

// Creates a new branch from the base branch
func createBranch(ctx context.Context, client *github.Client, repo, newBranch string) (*github.Reference, error) {

	owner := configs.RepoOwner
	baseRef, _, err := client.Git.GetRef(ctx, owner, repo, "refs/heads/main")
	if err != nil {
		LogCritical(fmt.Sprintf("Failed to get 'main' newBranchRef: %s", err))
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

	LogInfo(fmt.Sprintf("Branch created successfully: %s on %s", newRef, repo))

	return newBranchRef, nil
}

func createCommitTree(ctx context.Context, client *github.Client, targetBranch UploadKey, files map[string]string) (string, error) {

	var entries []*github.TreeEntry

	for path, content := range files {
		entries = append(entries, &github.TreeEntry{
			Path:    github.String(path),
			Type:    github.String("blob"),
			Mode:    github.String("100644"),
			Content: github.String(content),
		})
	}
	LogInfo(fmt.Sprintf("Updating %s/%s", targetBranch.RepoName, targetBranch.BranchPath))
	targetRef, _, err := client.Git.GetRef(ctx, configs.RepoOwner, targetBranch.RepoName, targetBranch.BranchPath)

	if err != nil || targetRef == nil {
		if err == nil {
			err = errors.Errorf("targetRef is nil")
		}
		LogCritical(fmt.Sprintf("Failed to get ref for %s: %s\n", targetBranch.RepoName, err))
		return "", err
	}

	baseSHA := targetRef.Object.SHA
	tree, _, err := client.Git.CreateTree(ctx, configs.RepoOwner, targetBranch.RepoName, *baseSHA, entries)
	if err != nil {
		return "", fmt.Errorf("failed to create tree: %w", err)
	}
	return *tree.SHA, nil
}

func createCommit(ctx context.Context, client *github.Client,
	targetBranch UploadKey, treeSHA *string, message string) error {
	targetRef, _, err := client.Git.GetRef(ctx, configs.RepoOwner, targetBranch.RepoName, targetBranch.BranchPath)

	parentCommit := &github.Commit{
		SHA: targetRef.Object.SHA,
	}

	commit := &github.Commit{
		Message: github.String(message),
		Tree:    &github.Tree{SHA: treeSHA},
		Parents: []*github.Commit{parentCommit},
	}

	newCommit, _, err := client.Git.CreateCommit(ctx, configs.RepoOwner, targetBranch.RepoName, commit)
	if err != nil {
		LogError(fmt.Sprintf("Failed to create commit %s: %v", targetBranch.RepoName, err))
		return fmt.Errorf("could not create commit: %w", err)
	}

	targetRef.Object.SHA = newCommit.SHA

	_, _, err = client.Git.UpdateRef(ctx, configs.RepoOwner, targetBranch.RepoName, targetRef, false)
	if err != nil {
		return fmt.Errorf("failed to update ref to new commit: %w", err)
	}
	//fmt.Printf("Created commit %s\n", newCommit.GetSHA())
	return nil
}

func createPullRequest(ctx context.Context, client *github.Client, repo, head, base, title, body string) (*github.PullRequest, error) {
	pr := &github.NewPullRequest{
		Title: github.String(title),
		Head:  github.String(head),
		Base:  github.String(base),
		Body:  github.String(body),
	}
	newPR, _, err := client.PullRequests.Create(ctx, configs.RepoOwner, repo, pr)
	if err != nil {
		return nil, fmt.Errorf("could not create PR: %w", err)
	}
	return newPR, nil
}

func mergePR(ctx context.Context, client *github.Client, repo string, prNumber int) error {
	options := &github.PullRequestOptions{
		MergeMethod: "merge", // Other options: "squash" or "rebase"
	}
	result, _, err := client.PullRequests.Merge(ctx, configs.RepoOwner, repo, prNumber, "Merging the pull request", options)
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

func deleteBranchIfExists(backgroundContext context.Context, client *github.Client, repo string, ref *github.Reference) {

	owner := configs.RepoOwner
	if ref.GetRef() == "refs/heads/main" { //yes, this happened once...
		LogError("I refuse to delete branch 'main'.")
		return
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
