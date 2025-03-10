package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"snooty-api-parser/snooty"
	"snooty-api-parser/types"
	"snooty-api-parser/utils"
	"strings"
	"time"
)

func getPageId(pageId string) string {
	parts := strings.Split(pageId, "/")
	// Check if the path has at least three parts to slice
	if len(parts) > 3 {
		// Omit the first three elements
		remainingParts := parts[3:]
		// Join the remaining parts back into a string with "/" separator
		result := strings.Join(remainingParts, "|")
		pageId = result
	} else {
		log.Println("The path", pageId, "does not have more than three parts to omit.")
	}
	return pageId
}

func main() {
	// Set up logging + a console display as this can take a long time
	startTime := time.Now()
	formattedTime := startTime.Format("2006-01-02 15:04:05")
	fmt.Println("Starting at ", formattedTime)
	file, err := os.Create("app.log")
	if err != nil {
		log.Print(err)
	}
	defer file.Close()
	log.SetOutput(file)

	// Set up the HTTP client to reuse across API calls
	client := &http.Client{
		Timeout: 30 * time.Second, // Set a timeout
	}
	// Uncomment to parse all projects
	//projectsToParse := snooty.GetProjects(client)

	// Uncomment to parse a single project during testing
	//sparkConnector := types.DocsProjectDetails{
	//	ProjectName:  "spark-connector",
	//	ActiveBranch: "v10.4",
	//	ProdUrl:      "https://mongodb.com/docs/spark-connector/current",
	//}
	//pyMongo := types.DocsProjectDetails{
	//	ProjectName:  "pymongo",
	//	ActiveBranch: "v4.11",
	//	ProdUrl:      "https://mongodb.com/docs/languages/python/pymongo-driver/current",
	//}
	//cDriver := types.DocsProjectDetails{
	//	ProjectName:  "c",
	//	ActiveBranch: "v1.30",
	//	ProdUrl:      "https://mongodb.com/docs/languages/c/c-driver/current",
	//}
	node := types.DocsProjectDetails{
		ProjectName:  "node",
		ActiveBranch: "v6.14",
		ProdUrl:      "https://mongodb.com/docs/drivers/node/current",
	}
	projectsToParse := []types.DocsProjectDetails{node}

	// Finish setting up console display to show progress during run
	totalProjects := len(projectsToParse)
	fmt.Printf("%d projects to parse\n", totalProjects)

	// Process docs pages for every project in the projectsToParse array
	firstProject := true
	for _, project := range projectsToParse {
		docsPages := snooty.GetProjectDocuments(project, client)
		log.Printf("Found %d docs pages for project %s\n", len(docsPages), project.ProjectName)
		if firstProject {
			utils.SetUpProgressDisplay(totalProjects, len(docsPages), project.ProjectName)
			firstProject = false
		} else {
			utils.SetNewSecondaryTarget(len(docsPages), project.ProjectName)
		}
		CheckDocsForUpdates(docsPages, project)
		utils.UpdatePrimaryTarget()
	}

	// Log some completion details to console
	endTime := time.Now()
	formattedTime = endTime.Format("2006-01-02 15:04:05")
	fmt.Println("\nCompleted at ", formattedTime)
	fmt.Println("Parsing projects took ", endTime.Sub(startTime))
}
