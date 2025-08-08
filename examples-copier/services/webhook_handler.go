package services

import (
	"encoding/json"
	"fmt"
	"github.com/google/go-github/v48/github"
	. "github.com/mongodb/code-example-tooling/code-copier/types"
	"io"
	"net/http"
	"path/filepath"
	"strings"
)

func ParseWebhookData(w http.ResponseWriter, r *http.Request) {
	GlobalContext.SetContext(r.Context())
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			LogInfo(fmt.Sprintf("Error closing ReadCloser %v", err))
		}
	}(r.Body)

	input, err := io.ReadAll(r.Body)
	if err != nil {
		LogCritical(fmt.Sprintf("Fail when parsing webhook: %v", err))
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(input, &payload); err != nil {
		LogError(fmt.Sprintf("Error unmarshalling outer JSON: %v", err))
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	pullRequest, ok := payload["pull_request"].(map[string]interface{})
	if !ok {
		LogWarning("Error asserting pull_request as map[string]interface{}")
		http.Error(w, "Invalid webhook payload format", http.StatusBadRequest)
		return
	}

	number, exists := pullRequest["number"]
	if !exists {
		LogWarning("Key 'number' missing in the JSON input")
		http.Error(w, "Missing required fields in payload", http.StatusBadRequest)
		return
	}

	numberFloat, ok := number.(float64)
	if !ok {
		LogWarning("Error asserting number as float64")
		http.Error(w, "Invalid number format in payload", http.StatusBadRequest)
		return
	}
	numberAsInt := int(numberFloat)

	state, ok := pullRequest["state"].(string)
	if !ok {
		LogWarning("Error asserting state as string")
		http.Error(w, "Invalid state format in payload", http.StatusBadRequest)
		return
	}

	merged, ok := pullRequest["merged"].(bool)
	if !ok {
		LogWarning("Error asserting merged as bool")
		http.Error(w, "Invalid merged format in payload", http.StatusBadRequest)
		return
	}

	if state == "closed" && merged {
		LogInfo(fmt.Sprintf("PR %d was merged and closed.", numberAsInt))
		LogInfo("--Start--")
		if err = HandlePrClosedEvent(numberAsInt); err != nil {
			LogError(fmt.Sprintf("Failed to handle PR closed event: %v", err))
			http.Error(w, "Failed to process webhook", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

func HandlePrClosedEvent(prNumber int) error {
	if InstallationAccessToken == "" {
		ConfigurePermissions()
	}

	configFile, configError := RetrieveAndParseConfigFile()
	if configError != nil {
		LogError(fmt.Sprintf("Failed to retrieve and parse config file: %v", configError))
		return fmt.Errorf("config file error: %w", configError)
	}

	changedFiles, changedFilesError := GetFilesChangedInPr(prNumber)
	if changedFilesError != nil {
		LogError(fmt.Sprintf("Failed to get files changed in PR %d: %v", prNumber, changedFilesError))
		return fmt.Errorf("failed to get changed files: %w", changedFilesError)
	}

	err := iterateFilesForCopy(changedFiles, configFile)
	if err != nil {
		return err
	}
	AddFilesToTargetRepoBranch()
	UpdateDeprecationFile()
	LogInfo("--Done--")
	return nil
}

// iterateFilesForCopy takes a splice of ChangedFiles from a PR, and the config file.
// It iterates through the file list to see if the source path matches one
// of the defined source paths in the config file, and if so, calls [addToRepoAndFilesMap]
func iterateFilesForCopy(changedFiles []ChangedFile, configFile ConfigFileType) error {
	var totalFileCount int32
	var uploadedCount int32

	for _, file := range changedFiles {
		totalFileCount++
		for _, config := range configFile {
			matches := false
			var relativePath string

			if config.RecursiveCopy {
				// Recursive mode - check if path starts with source directory
				if strings.HasPrefix(file.Path, config.SourceDirectory) {
					matches = true
					var err error
					relativePath, err = filepath.Rel(config.SourceDirectory, file.Path)
					if err != nil {
						return fmt.Errorf("failed to determine relative path for %s: %w", file.Path, err)
					}
				}
			} else {
				// Non-recursive mode - exact directory match only
				justPath := filepath.Dir(file.Path)
				if config.SourceDirectory == justPath {
					matches = true
					relativePath = filepath.Base(file.Path)
				}
			}

			if matches {
				target := filepath.Join(config.TargetDirectory, relativePath)

				if file.Status == "DELETED" {
					LogInfo(fmt.Sprintf("File %s has been deleted. Adding to the deprecation file.", target))
					addToDeprecationMap(target, config)
				} else {
					LogInfo(fmt.Sprintf("Found file %s to copy to %s/%s on branch %s",
						file.Path, config.TargetRepo, target, config.TargetBranch))
					fileContent, err := RetrieveFileContents(file.Path)
					if err != nil {
						return fmt.Errorf("failed to retrieve contents for %s: %w", file.Path, err)
					}
					addToRepoAndFilesMap(config.TargetRepo, config.TargetBranch, fileContent)
				}
				uploadedCount++
			}
		}
	}
	return nil
}

func addToDeprecationMap(target string, config Configs) {
	if FilesToDeprecate == nil {
		FilesToDeprecate = make(map[string]Configs)
	}
	FilesToDeprecate[target] = config
}

func addToRepoAndFilesMap(repoName, targetBranch string, file github.RepositoryContent) {
	if FilesToUpload == nil {
		FilesToUpload = make(map[UploadKey]UploadFileContent)
	}
	key := UploadKey{RepoName: repoName, BranchPath: fmt.Sprintf("%s%s", "refs/heads/", targetBranch)}
	if entry, exists := FilesToUpload[key]; exists {
		entry.Content = append(entry.Content, file)
		FilesToUpload[key] = entry
	} else {
		var fileContent = UploadFileContent{}
		fileContent.TargetBranch = targetBranch
		fileContent.Content = []github.RepositoryContent{file}
		FilesToUpload[key] = fileContent
	}
}
