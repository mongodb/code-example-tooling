package services

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/go-github/v48/github"
	"github.com/shurcooL/githubv4"
	"github.com/thompsch/app-tester/configs"
	. "github.com/thompsch/app-tester/types"
	"log"
)

var ctx = context.Background()

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

func GetFilesChangedInPr(pr_number int) ([]ChangedFile, error) {
	if InstallationAccessToken == "" {
		log.Println("No installation token provided")
		ConfigurePermissions()
	}

	var prQuery PullRequestQuery
	variables := map[string]interface{}{
		"owner":  githubv4.String(configs.RepoOwner),
		"name":   githubv4.String(configs.RepoName),
		"number": githubv4.Int(pr_number),
	}

	client := GetGraphQLClient()
	err := client.Query(context.Background(), &prQuery, variables)
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

func retrieveJsonFile(filePath string) string {
	client := GetRestClient()
	owner := configs.RepoOwner
	repo := configs.RepoName
	fileContent, _, _, err :=
		client.Repositories.GetContents(ctx, owner, repo,
			filePath, &github.RepositoryContentGetOptions{
				Ref: "main",
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

func RetrieveFileContents(filePath string) github.RepositoryContent {

	owner := configs.RepoOwner
	repo := configs.RepoName
	client := GetRestClient()

	fileContent, _, _, err :=
		client.Repositories.GetContents(ctx, owner, repo,
			filePath, &github.RepositoryContentGetOptions{
				Ref: "main",
			})

	if err != nil {
		LogCritical(fmt.Sprintf("Error getting file content: %v", err))
	}
	return *fileContent
}
