package main

import (
	"context"
	"fmt"
	"gdcd/add-code-examples"
	"gdcd/db"
	"gdcd/snooty"
	"gdcd/types"
	"gdcd/utils"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/tmc/langchaingo/llms/ollama"
)

func main() {
	// Set up logging + a console display to show progress
	// Logs are saved to a timestamped file in the logs directory, which is ignored by git
	// NOTE: the GDCD tool can take a long time to run (~1.5-2hrs, depending on your machine)
	startTime := time.Now()
	formattedTime := startTime.Format("2006-01-02 15:04:05")
	fmt.Println("Starting at ", formattedTime)

	logDir := "./logs"
	logFile, err := utils.InitLogger(logDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing logger: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Log file created:", logFile.Name())
	defer logFile.Close()

	// Determine the environment
	env := os.Getenv("APP_ENV")
	if env == "" {
		log.Fatal("APP_ENV is not set")
	}
	log.Println("Running the tool for APP_ENV:", env)
	// Load the appropriate .env file
	var envFile string
	switch env {
	case "development":
		envFile = ".env.development"
	case "production":
		envFile = ".env.production"
	default:
		log.Fatalf("Unknown environment: %s", env)
	}
	// Load the .env file
	err = godotenv.Load(envFile)
	if err != nil {
		log.Fatalf("Error loading %s file", envFile)
	}

	// Set up the HTTP client to reuse across API calls
	client := &http.Client{
		Timeout: 30 * time.Second, // Set a timeout
	}
	// Uncomment to parse all projects
	projectsToParse := snooty.GetProjects(client)

	// Uncomment to parse a single project during testing
	// compass := types.ProjectDetails{
	//	ProjectName:  "compass",
	//	Version: "master",
	//	ProdUrl:      "https://mongodb.com/docs/compass/current",
	// }
	// opsManager := types.ProjectDetails{
	//	ProjectName:  "ops-manager",
	//	Version: "v8.0",
	//	ProdUrl:      "https://mongodb.com/docs/ops-manager/current",
	// }
	// cloudManager := types.ProjectDetails{
	//	ProjectName:  "cloud-manager",
	//	Version: "master",
	//	ProdUrl:      "https://mongodb.com/docs/cloud-manager/",
	// }
	// projectsToParse := []types.ProjectDetails{compass}

	// Finish setting up console display to show progress during run
	totalProjects := len(projectsToParse)
	fmt.Printf("%d projects to parse\n", totalProjects)

	// Initialize the LLM
	ctx := context.Background()
	llm, err := ollama.New(ollama.WithModel(add_code_examples.MODEL))
	if err != nil {
		log.Fatalf("failed to connect to ollama: %v", err)
	}

	// Backup the current database
	db.BackUpDb()

	// Process pages for every project in the projectsToParse array
	firstProject := true
	for _, project := range projectsToParse {
		// Get pages from the API
		pages := snooty.GetProjectPages(project, client)
		pageCount := len(pages)
		log.Printf("Found %d docs pages for project %s\n", pageCount, project.ProjectName)
		report := types.ProjectReport{
			ProjectName: project.ProjectName,
			Changes:     nil,
			Issues:      nil,
			Counter: types.ProjectCounts{
				TotalCurrentPageCount: pageCount,
			},
		}
		if pageCount > 0 {
			if firstProject {
				utils.SetUpProgressDisplay(totalProjects, pageCount, project.ProjectName)
				firstProject = false
			} else {
				utils.SetNewSecondaryTarget(pageCount, project.ProjectName)
			}
			CheckPagesForUpdates(pages, project, llm, ctx, report)
			utils.UpdatePrimaryTarget()
		} else {
			report = utils.ReportIssues(types.PagesNotFoundIssue, report, project.ProjectName)
			LogReportForProject(project.ProjectName, report)
			utils.UpdatePrimaryTarget()
		}
	}
	utils.FinishPrintingProgressIndicators()

	// Log some completion details to console
	endTime := time.Now()
	formattedTime = endTime.Format("2006-01-02 15:04:05")
	fmt.Println("Completed at ", formattedTime)
	fmt.Println("Parsing projects took ", endTime.Sub(startTime))
}
