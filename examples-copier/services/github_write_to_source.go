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
	// Early return if there are no files to deprecate - prevents blank commits
	if len(FilesToDeprecate) == 0 {
		LogInfo("No deprecated files to record; skipping deprecation file update")
		return
	}

	// Fetch the deprecation file from the repository
	client := GetRestClient()
	ctx := context.Background()

	fileContent, _, _, err := client.Repositories.GetContents(
		ctx,
		os.Getenv(configs.ConfigRepoOwner),
		os.Getenv(configs.ConfigRepoName),
		os.Getenv(configs.DeprecationFile),
		&github.RepositoryContentGetOptions{
			Ref: os.Getenv(configs.ConfigRepoBranch),
		},
	)
	if err != nil {
		LogError(fmt.Sprintf("Error getting deprecation file: %v", err))
		return
	}

	content, err := fileContent.GetContent()
	if err != nil {
		LogError(fmt.Sprintf("Error decoding deprecation file: %v", err))
		return
	}

	var deprecationFile DeprecationFile
	err = json.Unmarshal([]byte(content), &deprecationFile)
	if err != nil {
		LogError(fmt.Sprintf("Failed to unmarshal %s: %v", configs.DeprecationFile, err))
		return
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

	LogInfo(fmt.Sprintf("Successfully updated %s with %d entries", os.Getenv(configs.DeprecationFile), len(FilesToDeprecate)))
}

func uploadDeprecationFileChanges(message string, newDeprecationFileContents string) {
	client := GetRestClient()
	ctx := context.Background()

	targetFileContent, _, _, err := client.Repositories.GetContents(ctx, os.Getenv(configs.ConfigRepoOwner), os.Getenv(configs.ConfigRepoName),
		os.Getenv(configs.DeprecationFile), &github.RepositoryContentGetOptions{Ref: os.Getenv(configs.ConfigRepoBranch)})

	if err != nil {
		LogError(fmt.Sprintf("Error getting deprecation file contents: %v", err))
	}

	options := &github.RepositoryContentFileOptions{
		Message: github.String(message),
		Content: []byte(newDeprecationFileContents),
		Branch:  github.String(os.Getenv(configs.ConfigRepoBranch)),
		Committer: &github.CommitAuthor{Name: github.String(os.Getenv(configs.CommitterName)),
			Email: github.String(os.Getenv(configs.CommitterEmail))},
	}

	options.SHA = targetFileContent.SHA
	_, _, err = client.Repositories.UpdateFile(ctx, os.Getenv(configs.ConfigRepoOwner), os.Getenv(configs.ConfigRepoName), os.Getenv(configs.DeprecationFile), options)
	if err != nil {
		LogError(fmt.Sprintf("Cannot update deprecation file: %v", err))
	}

	LogInfo(fmt.Sprintf("Deprecation file updated."))
}
