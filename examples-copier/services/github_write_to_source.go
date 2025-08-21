package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/google/go-github/v48/github"
	"github.com/mongodb/code-example-tooling/code-copier/configs"
	. "github.com/mongodb/code-example-tooling/code-copier/types"
)

func UpdateDeprecationFile() {
	content := retrieveJsonFile(os.Getenv(configs.DeprecationFile))

	var deprecationFile DeprecationFile
	err := json.Unmarshal([]byte(content), &deprecationFile)
	if err != nil {
		LogError(fmt.Sprintf("Failed to unmarshal %s: %v", configs.ConfigFile, err))
	}

	for key, value := range FilesToDeprecate {
		newDeprecatedFileEntry := DeprecatedFileEntry{
			FileName:  key,
			Repo:      value.TargetRepo,
			Branch:    value.TargetBranch,
			DeletedOn: time.Now().Format(time.RFC3339),
		}
		deprecationFile = append(deprecationFile, newDeprecatedFileEntry)
	}

	updatedJSON, err := json.MarshalIndent(deprecationFile, "", "  ")
	if err != nil {
		LogError(fmt.Sprintf("Error marshaling JSON: %v", err))
	}

	message := fmt.Sprintf("Updating %s.", os.Getenv(configs.DeprecationFile))
	uploadDeprecationFileChanges(message, string(updatedJSON))
}

func uploadDeprecationFileChanges(message string, newDeprecationFileContents string) {
	client := GetRestClient()
	ctx := context.Background()

	targetFileContent, _, _, err := client.Repositories.GetContents(ctx, os.Getenv(configs.RepoOwner), os.Getenv(configs.RepoName),
		os.Getenv(configs.DeprecationFile), &github.RepositoryContentGetOptions{Ref: os.Getenv(configs.SrcBranch)})

	if err != nil {
		LogError(fmt.Sprintf("Error getting deprecation file contents: %v", err))
	}

	options := &github.RepositoryContentFileOptions{
		Message: github.String(message),
		Content: []byte(newDeprecationFileContents),
		Branch:  github.String(os.Getenv(configs.SrcBranch)),
		Committer: &github.CommitAuthor{Name: github.String(os.Getenv(configs.CommiterName)),
			Email: github.String(os.Getenv(configs.CommiterEmail))},
	}

	options.SHA = targetFileContent.SHA
	_, _, err = client.Repositories.UpdateFile(ctx, os.Getenv(configs.RepoOwner), os.Getenv(configs.RepoName), os.Getenv(configs.DeprecationFile), options)
	if err != nil {
		LogError(fmt.Sprintf("Cannot update deprecation file: %v", err))
	}

	LogInfo(fmt.Sprintf("Deprecation file updated."))
}
