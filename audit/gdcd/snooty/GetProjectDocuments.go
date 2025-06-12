package snooty

import (
	"bufio"
	"fmt"
	"gdcd/types"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

// GetProjectDocuments calls the Snooty Data API endpoint for the given project and branch, and gets an array of newline-delimited
// JSON blobs as the response (if successful). The JSON maps to an array of []types.PageWrapper, which we can unmarshal
// for further processing.
func GetProjectDocuments(docsProject types.DocsProjectDetails, client *http.Client) []types.PageWrapper {
	env := os.Getenv("APP_ENV")
	apiURL := fmt.Sprintf("https://snooty-data-api.mongodb.com/prod/projects/%s/%s/documents", docsProject.ProjectName, docsProject.ActiveBranch)
	var reader bufio.Reader

	if env == "testing" {
		var stubbedResponse []byte
		if docsProject.ProjectName == "spark-connector" {
			stubbedResponse = LoadJsonTestDataFromFile("spark-connector-project-documents-stub.json")
		} else if docsProject.ProjectName == "c" {
			stubbedResponse = LoadJsonTestDataFromFile("c-driver-project-documents-stub.json")
		}
		body := io.NopCloser(strings.NewReader(string(stubbedResponse)))
		reader = *bufio.NewReader(body)
	} else {
		resp, err := client.Get(apiURL)
		if err != nil {
			log.Fatalf("Failed to make request for docs project %s: %v", docsProject.ProjectName, err)
		}
		defer resp.Body.Close()

		// Check for HTTP error response
		if resp.StatusCode != http.StatusOK {
			log.Fatalf("Error: received status code %d for docs project %s", resp.StatusCode, docsProject.ProjectName)
		}
		log.Printf("\nSuccessfully retrieved a Snooty response for docs project %s. Deserializing to PageWrapper now.", docsProject.ProjectName)
		reader = *bufio.NewReader(resp.Body)
	}

	projectDocuments := ReadDocsForGitHubUser(reader)
	if len(projectDocuments) == 0 {
		log.Printf("\nNo docs found for project %s using url %s", docsProject.ProjectName, apiURL)
	}
	return projectDocuments
}
