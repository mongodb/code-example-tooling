package snooty

import (
	"encoding/json"
	"gdcd/types"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
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

func getLastSegment(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		log.Printf("ERROR: failed to parse URL: %s\n", err)
	}
	u.Path = strings.TrimSuffix(u.Path, "/")
	seg := path.Base(u.Path)
	return seg
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
			log.Fatalf("ERROR: Failed to get the projects list from the Snooty Data API: %v", err)
		}
		defer resp.Body.Close()

		// Check for HTTP error response
		if resp.StatusCode != http.StatusOK {
			log.Fatalf("ERROR: when requesting the projects list from the Snooty Data API, received status code %d", resp.StatusCode)
		}

		// Read the response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatalf("ERROR: failed to read response body from Snooty Data API projects list: %s", err)
		}
		err = json.Unmarshal(body, &response)
		if err != nil {
			log.Fatalf("ERROR: failed to unmarshal JSON from the Snooty Data API projects list: %s", err)
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
		"mms-docs",
		"visual-studio-extension",
		"guides",
		"atlas-app-services",
		"drivers",
		"mongoid-railsmdb",
		"cluster-sync", // The Snooty Data API currently lists `cluster-sync` and `mongosync` as independent projects. We don't want to process twice, so ignore the `cluster-sync` entry.
	}

	var collectionsToParse []types.ProjectDetails
	for _, docsProject := range response.Data {
		var version string
		var prodUrl string
		if !contains(ignoreProjectNames, docsProject.Project) {
			for _, branch := range docsProject.Branches {
				if branch.Active && branch.IsStableBranch {
					// Some of the FullUrl fields have trailing slashes, and some don't. When we use the ProdUrl to make
					// the PageUrl, we add a slash, so we need to remove a trailing slash if one exists here so we don't
					// have double slashes.
					urlWithNoTrailingSlash := removeTrailingSlash(branch.FullUrl)

					// If the docs project has only one branch, we can assume it's unversioned and just use "main" as
					// the version to fetch documents. i.e. https://www.mongodb.com/docs/mongodb-shell/
					if len(docsProject.Branches) == 1 {
						version = "main"
					} else {
						// If the docs project has more than one branch, we need to use the active, stable branch name's
						// last segment of the URL as the version to fetch documents. i.e. https://www.mongodb.com/docs/atlas/operator/current/
						lastSegment := getLastSegment(urlWithNoTrailingSlash)
						version = lastSegment
					}
					prodUrl = urlWithNoTrailingSlash
					break
				}
			}
			// If the project does not have an active, stable branch, we don't want to try to get the project details
			if version != "" {
				collectionDetails := types.ProjectDetails{
					ProjectName: docsProject.Project,
					Version:     version,
					ProdUrl:     prodUrl,
				}
				collectionsToParse = append(collectionsToParse, collectionDetails)
			} else {
				log.Printf("Skipping project %s because it does not have an active, stable branch", docsProject.Project)
			}
		}
	}
	log.Println("Found ", len(collectionsToParse), "collections to parse from the Snooty Data API")
	return collectionsToParse
}
