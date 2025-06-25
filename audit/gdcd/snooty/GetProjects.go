package snooty

import (
	"encoding/json"
	"gdcd/types"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

func contains(slice []string, target string) bool {
	for _, s := range slice {
		if s == target {
			return true // The target string is found
		}
	}
	return false // The target string is not found
}

func removeTrailingSlash(input string) string {
	if strings.HasSuffix(input, "/") {
		return input[:len(input)-1]
	}
	return input
}

func GetProjects(client *http.Client) []types.ProjectDetails {
	env := os.Getenv("APP_ENV")
	var response types.Response
	if env == "testing" {
		stubbedResponse := LoadJsonTestDataFromFile("projects-stub.json")
		err := json.Unmarshal(stubbedResponse, &response)
		if err != nil {
			log.Fatalf("failed to unmarshal JSON: %s", err)
		}
	} else {
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

	var collectionsToParse []types.ProjectDetails
	for _, docsProject := range response.Data {
		var activeBranch string
		var prodUrl string
		if !contains(ignoreProjectNames, docsProject.Project) {
			for _, branch := range docsProject.Branches {
				if branch.Active && branch.IsStableBranch {
					// Some of the FullUrl fields have trailing slashes, and some don't. When we use the ProdUrl to make
					// the PageUrl, we add a slash, so we need to remove a trailing slash if one exists here so we don't
					// have double slashes.
					urlWithNoTrailingSlash := removeTrailingSlash(branch.FullUrl)
					activeBranch = branch.GitBranchName
					prodUrl = urlWithNoTrailingSlash
					break
				}
			}
			collectionDetails := types.ProjectDetails{
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
