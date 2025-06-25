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

// GetProjectPages calls the Snooty Data API endpoint for the given project and branch, and gets an array of newline-delimited
// JSON blobs as the response (if successful). The JSON maps to an array of []types.PageWrapper, which we can unmarshal
// for further processing.
func GetProjectPages(project types.ProjectDetails, client *http.Client) []types.PageWrapper {
	env := os.Getenv("APP_ENV")
	apiURL := fmt.Sprintf("https://snooty-data-api.mongodb.com/prod/projects/%s/%s/documents", project.ProjectName, project.ActiveBranch)
	var reader bufio.Reader

	if env == "testing" {
		var stubbedResponse []byte
		if project.ProjectName == "spark-connector" {
			stubbedResponse = LoadJsonTestDataFromFile("spark-connector-project-documents-stub.json")
		} else if project.ProjectName == "c" {
			stubbedResponse = LoadJsonTestDataFromFile("c-driver-project-documents-stub.json")
		}
		body := io.NopCloser(strings.NewReader(string(stubbedResponse)))
		reader = *bufio.NewReader(body)
	} else {
		resp, err := client.Get(apiURL)
		if err != nil {
			log.Fatalf("Failed to make request for project %s: %v", project.ProjectName, err)
		}
		defer resp.Body.Close()

		// Check for HTTP error response
		if resp.StatusCode != http.StatusOK {
			log.Fatalf("Error: received status code %d for project %s", resp.StatusCode, project.ProjectName)
		}
		log.Printf("\nSuccessfully retrieved a Snooty response for project %s. Deserializing to PageWrapper now.", project.ProjectName)
		reader = *bufio.NewReader(resp.Body)
	}

	projectDocuments := ReadPagesForGitHubUser(reader)
	if len(projectDocuments) == 0 {
		log.Printf("\nNo pages found for project %s using url %s", project.ProjectName, apiURL)
	}
	return projectDocuments
}
