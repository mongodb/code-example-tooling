package snooty

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"snooty-api-parser/types"
	"strings"
)

// GetProjectDocuments calls the Snooty Data API endpoint for the given project and branch, and gets an array of newline-delimited
// JSON blobs as the response (if successful). The JSON maps to an array of []types.PageWrapper, which we can unmarshal
// for further processing.
func GetProjectDocuments(docsProject types.DocsProjectDetails, client *http.Client) []types.PageWrapper {
	env := os.Getenv("APP_ENV")
	apiURL := fmt.Sprintf("https://snooty-data-api.mongodb.com/prod/projects/%s/%s/documents", docsProject.ProjectName, docsProject.ActiveBranch)
	var reader bufio.Reader
	if env == "production" {
		resp, err := client.Get(apiURL)
		if err != nil {
			log.Fatalf("Failed to make request for docs project %s: %v", docsProject.ProjectName, err)
		}
		defer resp.Body.Close()

		// Check for HTTP error response
		if resp.StatusCode != http.StatusOK {
			log.Fatalf("Error: received status code %d for docs project %s", resp.StatusCode, docsProject.ProjectName)
		}
		log.Printf("Successfully retrieved a Snooty response for docs project %s. Deserializing to PageWrapper now.", docsProject.ProjectName)
		reader = *bufio.NewReader(resp.Body)
	} else {
		var stubbedResponse []byte
		if docsProject.ProjectName == "spark-connector" {
			stubbedResponse = LoadJsonTestDataFromFile("spark-connector-project-documents-stub.json")
		} else if docsProject.ProjectName == "c" {
			stubbedResponse = LoadJsonTestDataFromFile("c-driver-project-documents-stub.json")
		}
		body := io.NopCloser(strings.NewReader(string(stubbedResponse)))
		reader = *bufio.NewReader(body)
	}

	// A DOP bug has introduced duplicate pages for some projects - one page with a GitHub username "netlify", and one with
	// a GitHub username "docs-builder-bot". We don't want to double-count pages/code examples, so this logic should only
	// count pages once depending on the GitHub username. Not all projects have duplicate usernames - I've created a
	// manual mapping here in `projectsWithOldDeploy` for whether to process as "docs-builder-bot" or "netlify".
	// When this DOP ticket is resolved, we can remove this logic: https://jira.mongodb.org/browse/DOP-5440
	projectsWithOldDeploy := []string{
		"spark-connector",
		"charts",
		"bi-connector",
		"cloudgov",
		"docs",
		"compass",
		"database-tools",
		"java",
		"atlas-cli",
		"cluster-sync",
		"ruby-driver",
		"csharp",
		"rust",
		"entity-framework",
		"atlas-operator",
		"cpp-driver",
		"scala",
		"pymongo-arrow",
		"mongodb-shell",
		"mongocli",
		"cloud-docs",
		"mongoid",
		"kotlin",
		"docs-relational-migrator",
		"ops-manager",
		"cloud-manager",
		"laravel",
		"pymongo",
		"c",
		"kotlin-sync",
		"atlas-architecture",
		"django",
	}

	var projectDocuments []types.PageWrapper
	if contains(projectsWithOldDeploy, docsProject.ProjectName) {
		projectDocuments = ReadDocsForGitHubUser(reader, GitHubUsernameDocsBuilderBot)
	} else {
		projectDocuments = ReadDocsForGitHubUser(reader, GitHubUsernameNetlify)
	}
	if len(projectDocuments) == 0 {
		log.Printf("No docs found for project %s using url %s", docsProject.ProjectName, apiURL)
	}
	return projectDocuments
}
