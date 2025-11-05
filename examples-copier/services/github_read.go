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
func GetFilesChangedInPr(pr_number int) ([]ChangedFile, error) {
	if InstallationAccessToken == "" {
		log.Println("No installation token provided")
		ConfigurePermissions()
	}

	var prQuery PullRequestQuery
	variables := map[string]interface{}{
		"owner":  githubv4.String(os.Getenv(configs.RepoOwner)),
		"name":   githubv4.String(os.Getenv(configs.RepoName)),
		"number": githubv4.Int(pr_number),
	}

	client := GetGraphQLClient()
	ctx := context.Background()
	err := client.Query(ctx, &prQuery, variables)
	if err != nil {
		LogCritical(fmt.Sprintf("Failed to execute query GetFilesChanged: %v", err))
		return nil, err
	}

	var changedFiles []ChangedFile
	for _, edge := range prQuery.Repository.PullRequest.Files.Edges {
		changedFiles = append(changedFiles, ChangedFile{
			Path:      string(edge.Node.Path),
			Additions: int(edge.Node.Additions),
			Deletions: int(edge.Node.Deletions),
			Status:    string(edge.Node.ChangeType),
		})
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

// RetrieveFileContents fetches the contents of a file from the repository at the specified path.
// It returns a github.RepositoryContent object containing the file details.
func RetrieveFileContents(filePath string) (github.RepositoryContent, error) {
	owner := os.Getenv(configs.RepoOwner)
	repo := os.Getenv(configs.RepoName)
	client := GetRestClient()
	ctx := context.Background()

	fileContent, _, _, err :=
		client.Repositories.GetContents(ctx, owner, repo,
			filePath, &github.RepositoryContentGetOptions{
				Ref: os.Getenv(configs.SrcBranch),
			})

	if err != nil {
		LogCritical(fmt.Sprintf("Error getting file content: %v", err))
	}
	return *fileContent, nil
}
