package main

import (
	"fmt"
	"log"
	"snooty-api-parser/db"
	"snooty-api-parser/types"
	"snooty-api-parser/utils"
)

func CheckDocsForUpdates(docsPages []types.PageWrapper, project types.DocsProjectDetails) {
	projectCounter := types.ProjectCounts{}
	incomingPageIds := make(map[string]bool)
	var newPageIds []string
	for _, page := range docsPages {
		var newPage types.DocsPage
		var updatedPage *types.DocsPage
		atlasDocId := getPageId(page.Data.PageID)
		incomingPageIds[atlasDocId] = true
		atlasDocument := db.GetAtlasPageData(project.ProjectName, atlasDocId)
		// If there is no existing document in Atlas that matches the page, we need to make a new page
		if atlasDocument == nil {
			newPage, projectCounter = MakeNewDocsPage(page, project.ProdUrl, project.ProjectName, projectCounter)
			newPageIds = append(newPageIds, atlasDocId)
			log.Printf("Info: new docs page for %s: \n", atlasDocId)
			log.Printf("Page: %+v\n", newPage)
		} else {
			// If there is an existing document in Atlas, update the existing page
			// If the code example counts are the same on the incoming page as they are on the existing page,
			// we treat that as an unchanged page and it does not return an updated page - it returns nil
			projectCounter.ExistingCodeNodesCount += atlasDocument.CodeNodesTotal
			projectCounter.ExistingLiteralIncludeCount += atlasDocument.LiteralIncludesTotal
			projectCounter.ExistingIoCodeBlockCount += atlasDocument.IoCodeBlocksTotal
			updatedPage, projectCounter = UpdateExistingDocsPage(*atlasDocument, page, projectCounter)
			if updatedPage != nil {
				log.Println("Updated page ", updatedPage.ID)
			}
		}
		utils.UpdateSecondaryTarget()
	}
	// TODO: Figure out what to output to the log about the project details
	var projectReport []string
	existingPageIds := db.GetAtlasPageIDs(project.ProjectName)
	var removedPageIds []string
	if existingPageIds != nil {
		for _, existingId := range existingPageIds {
			if !incomingPageIds[existingId] {
				removedPageIds = append(removedPageIds, existingId)
			}
		}
	}
	summaryDoc := db.GetAtlasProjectSummaryData(project.ProjectName)
	var latestCollectionInfo types.CollectionInfoView
	collectionVersionKey := ""
	if summaryDoc != nil {
		elementIndex := 0
		for version, info := range summaryDoc.Version {
			if elementIndex == 0 {
				latestCollectionInfo = info
				collectionVersionKey = version
				if len(summaryDoc.Version) > 1 {
					elementIndex++
				}
			} else {
				if info.LastUpdatedAtUTC.After(latestCollectionInfo.LastUpdatedAtUTC) {
					latestCollectionInfo = info
					collectionVersionKey = version
					if elementIndex > len(summaryDoc.Version) {
						elementIndex++
					}
				}
			}
		}
	}
	if project.ActiveBranch != collectionVersionKey {
		// TODO: If the active branch doesn't match the most recent version, need to make a whole new version document
	} else {
		// TODO: If the active branch does match the most recent version, just need to update this version document's last updated date and counts
	}
	if latestCollectionInfo.TotalCodeCount != projectCounter.IncomingCodeNodesCount {
		projectReport = append(projectReport,
			fmt.Sprintf("Info: project code count from summary doc: was %d, got %d", latestCollectionInfo.TotalCodeCount, projectCounter.IncomingCodeNodesCount))
	}
	if latestCollectionInfo.TotalPageCount != len(docsPages) {
		projectReport = append(projectReport,
			fmt.Sprintf("Info: page count from summary doc: was %d, got %d", latestCollectionInfo.TotalPageCount, len(docsPages)))
	}
	if len(removedPageIds) > 0 {
		projectReport = append(projectReport, fmt.Sprintf("Info: the following pages were removed from the docs:"))
		for _, removedPageId := range removedPageIds {
			projectReport = append(projectReport, fmt.Sprintf("Info: removed page ID: %s", removedPageId))
		}
	}
	if len(newPageIds) > 0 {
		projectReport = append(projectReport, fmt.Sprintf("Info: the following new pages were added to the docs:"))
		for _, newPageId := range newPageIds {
			projectReport = append(projectReport, fmt.Sprintf("Info: new page ID: %s", newPageId))
		}
	}
	runningCount := 0
	if projectCounter.RemovedCodeNodesCount > 0 {
		projectReport = append(projectReport, fmt.Sprintf("Info: %d removed examples", projectCounter.RemovedCodeNodesCount))
		runningCount = latestCollectionInfo.TotalCodeCount - projectCounter.RemovedCodeNodesCount
	}
	if projectCounter.NewCodeNodesCount > 0 {
		projectReport = append(projectReport, fmt.Sprintf("Info: %d new examples", projectCounter.NewCodeNodesCount))
		runningCount = latestCollectionInfo.TotalCodeCount + projectCounter.NewCodeNodesCount
	}
	if runningCount != 0 && runningCount != projectCounter.IncomingCodeNodesCount {
		projectReport = append(projectReport, fmt.Sprintf("ISSUE: Node counts don't match: should be %d, got %d", runningCount, projectCounter.IncomingCodeNodesCount))
	}
	if projectCounter.IncomingCodeNodesCount != projectCounter.ExistingCodeNodesCount {
		projectReport = append(projectReport, fmt.Sprintf("Info: code node count change: was %d, got %d", projectCounter.ExistingCodeNodesCount, projectCounter.IncomingCodeNodesCount))
	}
	if projectCounter.IncomingLiteralIncludeCount != projectCounter.ExistingLiteralIncludeCount {
		projectReport = append(projectReport, fmt.Sprintf("Info: literalinclude count change: was %d, got %d", projectCounter.ExistingLiteralIncludeCount, projectCounter.IncomingLiteralIncludeCount))
	}
	if projectCounter.IncomingIoCodeBlockCount != projectCounter.ExistingIoCodeBlockCount {
		projectReport = append(projectReport, fmt.Sprintf("Info: io-code-block count change: was %d, got %d", projectCounter.ExistingIoCodeBlockCount, projectCounter.IncomingIoCodeBlockCount))
	}
	if len(projectReport) > 0 {
		log.Printf("\nProject report for %s\n", project.ProjectName)
		for _, line := range projectReport {
			log.Println(line)
		}
		log.Printf("\n")
	} else {
		log.Printf("\nProject report for %s\n", project.ProjectName)
		log.Println("No counts changed in project")
		log.Printf("\n")
	}
}
