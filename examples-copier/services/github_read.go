package services

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/google/go-github/v48/github"
	"github.com/mongodb/code-example-tooling/code-copier/configs"
	. "github.com/mongodb/code-example-tooling/code-copier/types"
	"github.com/shurcooL/githubv4"
)

// GetFilesChangedInPr retrieves the list of files changed in a specified pull request.
// It returns a slice of ChangedFile structures containing details about each changed file.
// Parameters:
//   - owner: The repository owner (e.g., "cbullinger")
//   - repo: The repository name (e.g., "aggregation-tasks")
//   - pr_number: The pull request number
func GetFilesChangedInPr(owner string, repo string, pr_number int) ([]ChangedFile, error) {
	if InstallationAccessToken == "" {
		log.Println("No installation token provided")
		ConfigurePermissions()
	}

	client := GetGraphQLClient()
	ctx := context.Background()

	var changedFiles []ChangedFile
	var cursor *githubv4.String = nil
	hasNextPage := true

	// Paginate through all files
	for hasNextPage {
		var prQuery PullRequestQuery
		variables := map[string]interface{}{
			"owner":  githubv4.String(owner),
			"name":   githubv4.String(repo),
			"number": githubv4.Int(pr_number),
			"cursor": cursor,
		}

		err := client.Query(ctx, &prQuery, variables)
		if err != nil {
			LogCritical(fmt.Sprintf("Failed to execute query GetFilesChanged: %v", err))
			return nil, err
		}

		// Append files from this page
		for _, edge := range prQuery.Repository.PullRequest.Files.Edges {
			changedFiles = append(changedFiles, ChangedFile{
				Path:      string(edge.Node.Path),
				Additions: int(edge.Node.Additions),
				Deletions: int(edge.Node.Deletions),
				Status:    string(edge.Node.ChangeType),
			})
		}

		// Check if there are more pages
		hasNextPage = prQuery.Repository.PullRequest.Files.PageInfo.HasNextPage
		if hasNextPage {
			cursor = &prQuery.Repository.PullRequest.Files.PageInfo.EndCursor
		}
	}

	LogInfo(fmt.Sprintf("PR has %d changed files.", len(changedFiles)))

	// Log all files for debugging (especially to see if server files are included)
	LogInfo("=== ALL FILES FROM GRAPHQL API ===")
	for i, file := range changedFiles {
		LogInfo(fmt.Sprintf("  [%d] %s (status: %s)", i, file.Path, file.Status))
	}
	LogInfo("=== END FILE LIST ===")

	// Count files by directory for debugging
	clientCount := 0
	serverCount := 0
	otherCount := 0
	for _, file := range changedFiles {
		if len(file.Path) >= 13 && file.Path[:13] == "mflix/client/" {
			clientCount++
		} else if len(file.Path) >= 13 && file.Path[:13] == "mflix/server/" {
			serverCount++
		} else {
			otherCount++
		}
	}
	LogInfo(fmt.Sprintf("File breakdown: client=%d, server=%d, other=%d", clientCount, serverCount, otherCount))

	return changedFiles, nil
}

// RetrieveFileContents fetches the contents of a file from the config repository at the specified path.
// It returns a github.RepositoryContent object containing the file details.
func RetrieveFileContents(filePath string) (github.RepositoryContent, error) {
	owner := os.Getenv(configs.ConfigRepoOwner)
	repo := os.Getenv(configs.ConfigRepoName)
	client := GetRestClient()
	ctx := context.Background()

	fileContent, _, _, err :=
		client.Repositories.GetContents(ctx, owner, repo,
			filePath, &github.RepositoryContentGetOptions{
				Ref: os.Getenv(configs.ConfigRepoBranch),
			})

	if err != nil {
		LogCritical(fmt.Sprintf("Error getting file content: %v", err))
	}
	return *fileContent, nil
}
