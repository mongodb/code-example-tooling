package services

import (
	"encoding/json"
	"fmt"
	"github.com/google/go-github/v48/github"
	. "github.com/thompsch/app-tester/types"
	"io"
	"log"
	"net/http"
	"path/filepath"
)

func ParseWebhookData(w http.ResponseWriter, r *http.Request) {
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			LogInfo(fmt.Sprintf("Error closing ReadCloser %v", err))
		}
	}(r.Body)

	input, err := io.ReadAll(r.Body)
	if err != nil {
		LogCritical(fmt.Sprintf("Fail when parsing webhook: %v", err))
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(input, &payload); err != nil {
		LogError(fmt.Sprintf("Error unmarshalling outer JSON: %v", err))
	}

	pullRequest, ok := payload["pull_request"].(map[string]interface{})
	if !ok {
		LogWarning(fmt.Sprintf("Error asserting pull_request as map[string]interface{}"))
	}

	number, exists := pullRequest["number"]
	if !exists {
		LogWarning(fmt.Sprint("Key 'number' missing in the JSON input"))
	}

	numberAsInt := int(number.(float64))

	state, ok := pullRequest["state"].(string)
	if !ok {
		LogWarning(fmt.Sprintf("Error asserting state as string"))
	}
	merged, ok := pullRequest["merged"].(bool)
	if !ok {
		log.Println("Error asserting merged as bool")
	}

	if state == "closed" && merged {
		LogInfo(fmt.Sprintf("PR %d was merged and closed.", numberAsInt))
		LogInfo("--Start--")
		HandlePrClosedEvent(numberAsInt)
	}
}

func HandlePrClosedEvent(prNumber int) {
	if InstallationAccessToken == "" {
		ConfigurePermissions()
	}
	configFile, configError := RetrieveAndParseConfigFile()
	if configError == nil {
		changedFiles, changedFilesError := GetFilesChangedInPr(prNumber)
		if changedFilesError == nil {
			iterateFilesForCopy(changedFiles, configFile)
			AddFilesToTargetRepoBranch()
			UpdateDeprecationFile()
			LogInfo("--Done--")
		}
	}
}

// iterateFilesForCopy takes a splice of ChangedFiles from a PR, and the config file.
// It iterates through the file list to see if the source path matches one
// of the defined source paths in the config file, and if so, calls [addToRepoAndFilesMap]
func iterateFilesForCopy(changedFiles []ChangedFile, configFile ConfigFileType) {
	var totalFileCount int32
	var uploadedCount int32

	for _, file := range changedFiles {
		totalFileCount++
		for _, config := range configFile {
			justPath := filepath.Dir(file.Path)
			justFile := filepath.Base(file.Path)
			if config.SourceDirectory == justPath {
				target := fmt.Sprintf("%s/%s", config.TargetDirectory, justFile)

				if file.Status == "DELETED" {
					LogInfo(fmt.Sprintf("File %s has been deleted. Adding to the deprecation file.", target))
					addToDeprecationMap(target, config)
				} else {
					LogInfo(fmt.Sprintf("Found file %s to copy to %s/%s on branch %s",
						file.Path, config.TargetRepo, config.TargetDirectory, config.TargetBranch))
					addToRepoAndFilesMap(config.TargetRepo, config.TargetBranch, RetrieveFileContents(file.Path))
				}
				uploadedCount++
			}
		}
	}
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
