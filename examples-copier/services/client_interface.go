package services

import (
	"context"
	"github.com/google/go-github/v48/github"
)

// GitHubServiceInterface defines methods from GitHub API
type GitHubServiceInterface interface {
	GetContents(ctx context.Context, owner, repo, path string, opt *github.RepositoryContentGetOptions) (*github.RepositoryContent, []*github.RepositoryContent, *github.Response, error)
	CreateOrUpdateFile(ctx context.Context, owner, repo, path string, opts *github.RepositoryContentFileOptions) (*github.RepositoryContentResponse, *github.Response, error)
}

// GitHubWrapper adapts the standard GitHub client to our interface
type GitHubWrapper struct {
	Client *github.RepositoriesService
}

// GetContents wraps the GitHub API method
func (w *GitHubWrapper) GetContents(ctx context.Context, owner, repo, path string, opt *github.RepositoryContentGetOptions) (*github.RepositoryContent, []*github.RepositoryContent, *github.Response, error) {
	return w.Client.GetContents(ctx, owner, repo, path, opt)
}

// CreateOrUpdateFile implements the missing method by checking if file exists first
func (w *GitHubWrapper) CreateOrUpdateFile(ctx context.Context, owner, repo, path string, opts *github.RepositoryContentFileOptions) (*github.RepositoryContentResponse, *github.Response, error) {
	var ref string
	if opts.Branch != nil {
		ref = *opts.Branch
	}

	content, _, resp, err := w.Client.GetContents(ctx, owner, repo, path, &github.RepositoryContentGetOptions{Ref: ref})

	if err != nil && resp.StatusCode == 404 {
		// File doesn't exist, create it
		return w.Client.CreateFile(ctx, owner, repo, path, opts)
	} else if err != nil {
		// Some other error occurred
		return nil, resp, err
	}

	// File exists, update it
	opts.SHA = content.SHA
	return w.Client.UpdateFile(ctx, owner, repo, path, opts)
}
