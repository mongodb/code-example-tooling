package snooty

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"snooty-api-parser/types"
)

func processLine(line []byte, gitHubUserName string) *types.PageWrapper {
	var generic map[string]interface{}
	if err := json.Unmarshal(line, &generic); err != nil {
		log.Fatalf("Failed to unmarshal line: %v", err)
	}
	typeField, ok := generic["type"].(string)
	if !ok {
		log.Fatalf("Type field is missing or not a string in line: %s", line)
	}
	// Process based on typeField
	switch typeField {
	case "timestamp":
		var timestamp types.TimestampData
		if err := json.Unmarshal(line, &timestamp); err != nil {
			log.Fatalf("Failed to unmarshal TimestampData: %v", err)
		}
		//fmt.Printf("Processed as TimestampData: %+v\n", timestamp)
	case "metadata":
		var metadata types.ProjectMetadataWrapper
		if err := json.Unmarshal(line, &metadata); err != nil {
			log.Fatalf("Failed to unmarshal ProjectMetadata: %v", err)
		}
		//fmt.Printf("Project ID is: %+v\n", metadata.Data.ID)
	case "page":
		var page types.PageWrapper
		if err := json.Unmarshal(line, &page); err != nil {
			log.Fatalf("Failed to unmarshal PageMetadata: %v", err)
		}
		//return &page
		if page.Data.GitHubUsername == gitHubUserName {
			return &page
		}
	case "asset":
		var fileAsset types.ProjectAsset
		if err := json.Unmarshal(line, &fileAsset); err != nil {
			log.Fatalf("Failed to unmarshal ProjectAsset: %v", err)
		}
		//fmt.Printf("Filename for asset is: %+v\n", fileAsset.Data.Filenames[0])
	default:
		log.Printf("Unknown type: %s\n", typeField)
	}
	return nil
}

func processLineWithoutGitHubUsername(line []byte) *types.PageWrapper {
	var generic map[string]interface{}
	if err := json.Unmarshal(line, &generic); err != nil {
		log.Fatalf("Failed to unmarshal line: %v", err)
	}
	typeField, ok := generic["type"].(string)
	if !ok {
		log.Fatalf("Type field is missing or not a string in line: %s", line)
	}
	// Process based on typeField
	switch typeField {
	case "timestamp":
		var timestamp types.TimestampData
		if err := json.Unmarshal(line, &timestamp); err != nil {
			log.Fatalf("Failed to unmarshal TimestampData: %v", err)
		}
		//fmt.Printf("Processed as TimestampData: %+v\n", timestamp)
	case "metadata":
		var metadata types.ProjectMetadataWrapper
		if err := json.Unmarshal(line, &metadata); err != nil {
			log.Fatalf("Failed to unmarshal ProjectMetadata: %v", err)
		}
		//fmt.Printf("Project ID is: %+v\n", metadata.Data.ID)
	case "page":
		var page types.PageWrapper
		if err := json.Unmarshal(line, &page); err != nil {
			log.Fatalf("Failed to unmarshal PageMetadata: %v", err)
		}
		return &page
	case "asset":
		var fileAsset types.ProjectAsset
		if err := json.Unmarshal(line, &fileAsset); err != nil {
			log.Fatalf("Failed to unmarshal ProjectAsset: %v", err)
		}
		//fmt.Printf("Filename for asset is: %+v\n", fileAsset.Data.Filenames[0])
	default:
		log.Printf("Unknown type: %s\n", typeField)
	}
	return nil
}

func ReadDocsForNetlify(reader bufio.Reader) []types.PageWrapper {
	var docsPages []types.PageWrapper
	gitHubUserNetlify := "netlify"
	for {
		line, err := reader.ReadBytes('\n') // Read until newline
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatalf("Error reading response: %v", err)
		}

		trimmedLine := bytes.TrimSpace(line)
		var maybePage *types.PageWrapper
		if len(trimmedLine) > 0 { // Process non-empty lines
			maybePage = processLine(trimmedLine, gitHubUserNetlify)
			if maybePage != nil {
				docsPages = append(docsPages, *maybePage)
			}
		}
	}
	return docsPages
}

func ReadDocsForDocsBuilderBot(reader bufio.Reader) []types.PageWrapper {
	var docsPages []types.PageWrapper
	for {
		line, err := reader.ReadBytes('\n') // Read until newline
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatalf("Error reading response: %v", err)
		}

		trimmedLine := bytes.TrimSpace(line)
		var maybePage *types.PageWrapper
		if len(trimmedLine) > 0 { // Process non-empty lines
			maybePage = processLineWithoutGitHubUsername(trimmedLine)
			if maybePage != nil {
				docsPages = append(docsPages, *maybePage)
			}
		}
	}
	return docsPages
}

func GetDocsPages(docsProject types.DocsProjectDetails, client *http.Client) []types.PageWrapper {
	projectsWithOldDeploy := []string{
		"spark-connector",
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
	}
	var docsPages []types.PageWrapper
	apiURL := fmt.Sprintf("https://snooty-data-api.mongodb.com/prod/projects/%s/%s/documents", docsProject.ProjectName, docsProject.ActiveBranch)
	// Make the HTTP GET request
	resp, err := client.Get(apiURL)
	if err != nil {
		log.Fatalf("Failed to make request for docs project %s: %v", docsProject.ProjectName, err)
	}
	defer resp.Body.Close()

	// Check for HTTP error response
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Error: received status code %d for docs project %s", docsProject.ProjectName, resp.StatusCode)
	}
	log.Printf("Successfully retrieved a Snooty response for docs project %s. Deserializing to pages now.", docsProject.ProjectName)
	reader := bufio.NewReader(resp.Body)
	if contains(projectsWithOldDeploy, docsProject.ProjectName) {
		docsPages = ReadDocsForDocsBuilderBot(*reader)
	} else {
		docsPages = ReadDocsForNetlify(*reader)
	}
	if len(docsPages) == 0 {
		log.Printf("No docs found for project %s using url %s", docsProject.ProjectName, apiURL)
	}
	return docsPages
}
