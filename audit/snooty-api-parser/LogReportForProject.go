package main

import (
	"log"
	"snooty-api-parser/types"
)

func LogReportForProject(projectName string, report types.ProjectReport) {
	if len(report.Changes) > 0 {
		log.Printf("\nProject changes for %s\n", projectName)
		for _, change := range report.Changes {
			log.Printf("%s: %s", change.Type.String(), change.Data.(string))
		}
	} else if len(report.Changes) == 0 {
		log.Printf("\nProject changes for %s\n", projectName)
		log.Println("No changes in project")
	}
	if len(report.Issues) > 0 {
		log.Printf("\nIssues with data in project %s\n", projectName)
		for _, issue := range report.Issues {
			log.Printf("%s: %s", issue.Type.String(), issue.Data.(string))
		}
	} else if len(report.Issues) == 0 {
		log.Printf("\nNo issues with data in project %s\n", projectName)
	}
	if report.Counter.NewAppliedUsageExamplesCount > 0 {
		log.Printf("\nNew applied usage examples for %s: %d\n", projectName, report.Counter.NewAppliedUsageExamplesCount)
	}
}
