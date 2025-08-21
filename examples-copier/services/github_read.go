package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/google/go-github/v48/github"
	"github.com/mongodb/code-example-tooling/code-copier/configs"
	. "github.com/mongodb/code-example-tooling/code-copier/types"
	"github.com/shurcooL/githubv4"
)

// RetrieveAndParseConfigFile fetches the configuration file from the repository
// and unmarshals its JSON content into a ConfigFileType structure.
func RetrieveAndParseConfigFile() (ConfigFileType, error) {
	content := retrieveJsonFile(configs.ConfigFile)
	if content == "" {
		return nil, &github.Error{Message: "Config File Not Found or is empty"}
	}
	var configFile ConfigFileType
	err := json.Unmarshal([]byte(content), &configFile)
	if err != nil {
		LogError(fmt.Sprintf("Failed to unmarshal %s: %v", configs.ConfigFile, err))
		return nil, err
	}
	return configFile, nil
}

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
	return changedFiles, nil
}

// retrieveJsonFile fetches the content of a JSON file from the specified path in the repository.
// It returns the file content as a string.
func retrieveJsonFile(filePath string) string {
	client := GetRestClient()
	owner := os.Getenv(configs.RepoOwner)
	repo := os.Getenv(configs.RepoName)
	ctx := context.Background()
	fileContent, _, _, err :=
		client.Repositories.GetContents(ctx, owner, repo,
			filePath, &github.RepositoryContentGetOptions{
				Ref: os.Getenv(configs.SrcBranch),
			})
	if err != nil {
		LogCritical(fmt.Sprintf("Error getting file content: %v", err))
		return ""
	}

	content, err := fileContent.GetContent()
	if err != nil {
		LogCritical(fmt.Sprintf("Error decoding file content: %v", err))
		return ""
	}
	return content
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
