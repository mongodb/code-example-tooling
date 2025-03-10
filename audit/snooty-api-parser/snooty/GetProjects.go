package snooty

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"snooty-api-parser/types"
)

func contains(slice []string, target string) bool {
	for _, s := range slice {
		if s == target {
			return true // The target string is found
		}
	}
	return false // The target string is not found
}

func GetProjects(client *http.Client) []types.DocsProjectDetails {
	env := os.Getenv("APP_ENV")
	var response types.Response
	if env == "production" {
		apiURL := "https://snooty-data-api.mongodb.com/prod/projects/"
		resp, err := client.Get(apiURL)
		if err != nil {
			log.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		// Check for HTTP error response
		if resp.StatusCode != http.StatusOK {
			log.Fatalf("Error: received status code %d", resp.StatusCode)
		}

		// Read the response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatalf("failed to read response body: %s", err)
		}
		err = json.Unmarshal(body, &response)
		if err != nil {
			log.Fatalf("failed to unmarshal JSON: %s", err)
		}
	} else {
		stubbedResponse := LoadJsonTestDataFromFile("projects-stub.json")
		err := json.Unmarshal(stubbedResponse, &response)
		if err != nil {
			log.Fatalf("failed to unmarshal JSON: %s", err)
		}
	}

	ignoreProjectNames := []string{
		"atlas-open-service-broker",
		"realm",
		"docs-app-services",
		"datalake",
		"intellij",
		"landing",
		"mongodb-vscode",
		"visual-studio-extension",
		"guides",
		"atlas-app-services",
		"drivers",
		"mongoid-railsmdb",
	}

	var collectionsToParse []types.DocsProjectDetails
	for _, docsProject := range response.Data {
		var activeBranch string
		var prodUrl string
		if !contains(ignoreProjectNames, docsProject.Project) {
			for _, branch := range docsProject.Branches {
				if branch.Active && branch.IsStableBranch {
					activeBranch = branch.GitBranchName
					prodUrl = branch.FullUrl
					break
				}
			}
			collectionDetails := types.DocsProjectDetails{
				ProjectName:  docsProject.Project,
				ActiveBranch: activeBranch,
				ProdUrl:      prodUrl,
			}
			collectionsToParse = append(collectionsToParse, collectionDetails)
		}
	}
	log.Println("Found ", len(collectionsToParse), "collections to parse from the Snooty Data API")
	return collectionsToParse
}
