package services

import (
	"encoding/json"
	"fmt"
	"github.com/google/go-github/v48/github"
	"github.com/thompsch/app-tester/configs"
	. "github.com/thompsch/app-tester/types"
	"time"
)

func UpdateDeprecationFile() {
	content := retrieveJsonFile(configs.DeprecationFile)

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

	message := fmt.Sprintf("Updating %s.", configs.DeprecationFile)
	uploadDeprecationFileChanges(message, string(updatedJSON))
}

func uploadDeprecationFileChanges(message string, newDeprecationFileContents string) {
	client := GetRestClient()

	targetFileContent, _, _, err := client.Repositories.GetContents(ctx, configs.RepoOwner, configs.RepoName,
		configs.DeprecationFile, &github.RepositoryContentGetOptions{Ref: "main"})

	if err != nil {
		LogError(fmt.Sprintf("Error getting deprecation file contents: %v", err))
	}

	options := &github.RepositoryContentFileOptions{
		Message: github.String(message),
		Content: []byte(newDeprecationFileContents),
		Branch:  github.String("main"),
		Committer: &github.CommitAuthor{Name: github.String(configs.CommiterName),
			Email: github.String(configs.CommiterEmail)},
	}

	options.SHA = targetFileContent.SHA
	_, _, err = client.Repositories.UpdateFile(ctx, configs.RepoOwner, configs.RepoName, configs.DeprecationFile, options)
	if err != nil {
		LogError(fmt.Sprintf("Cannot update deprecation file: %v", err))
	}

	LogInfo(fmt.Sprintf("Deprecation file updated."))
}
