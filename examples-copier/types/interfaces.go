package types

import (
	"context"
	"github.com/google/go-github/v48/github"
)

type GitHubService interface {
	GetContents(ctx context.Context, owner, repo, path string,
		opt *github.RepositoryContentGetOptions) (fileContent *github.RepositoryContent,
		directoryContent []*github.RepositoryContent, response *github.Response, err error)
}
