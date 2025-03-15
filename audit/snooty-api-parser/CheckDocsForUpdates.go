package main

import (
	"context"
	"fmt"
	"github.com/tmc/langchaingo/llms/ollama"
	"log"
	"snooty-api-parser/db"
	"snooty-api-parser/types"
	"snooty-api-parser/utils"
)

func CheckDocsForUpdates(docsPages []types.PageWrapper, project types.DocsProjectDetails, llm *ollama.LLM, ctx context.Context, report types.ProjectReport) {
	incomingPageIds := make(map[string]bool)
	incomingPageCount := len(docsPages)
	var newPageIds []string
	var newPages []types.DocsPage
	var updatedPages []types.DocsPage
	for _, page := range docsPages {
		atlasDocId := utils.ConvertSnootyPageIdToAtlasPageId(page.Data.PageID)
		incomingPageIds[atlasDocId] = true
		atlasDocument := db.GetAtlasPageData(project.ProjectName, atlasDocId)
		// If there is no existing document in Atlas that matches the page, we need to make a new page
		if atlasDocument == nil {
			var newPage types.DocsPage
			newPage, report = MakeNewDocsPage(page, project.ProdUrl, report, llm, ctx)
			newPageIds = append(newPageIds, atlasDocId)
			newPages = append(newPages, newPage)
		} else {
			// If there is an existing document in Atlas, update the existing page
			// If the code example counts are the same on the incoming page as they are on the existing page,
			// we treat that as an unchanged page and it does not return an updated page - it returns nil
			var updatedPage *types.DocsPage
			updatedPage, report = UpdateExistingDocsPage(*atlasDocument, page, report, llm, ctx)
			if updatedPage != nil {
				updatedPages = append(updatedPages, *updatedPage)
			}
		}
		utils.UpdateSecondaryTarget()
	}
	existingPageIds := db.GetAtlasPageIDs(project.ProjectName)
	var removedPageIds []string
	if existingPageIds != nil {
		for _, existingId := range existingPageIds {
			if !incomingPageIds[existingId] {
				removedPageIds = append(removedPageIds, existingId)
			}
		}
	}
	for _, removedPageId := range removedPageIds {
		atlasDocument := db.GetAtlasPageData(project.ProjectName, removedPageId)
		if atlasDocument == nil {
			var removedDocument *types.DocsPage
			removedDocument, report = RemoveExistingPage(atlasDocument, report)
			updatedPages = append(updatedPages, *removedDocument)
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

	// TODO: Move report updates out to a separate func
	if latestCollectionInfo.TotalCodeCount != report.Counter.IncomingCodeNodesCount {
		codeNodeCountChange := types.Change{
			Type: types.ProjectSummaryCodeNodeCountChange,
			Data: fmt.Sprintf("Project %s: code node count from summary was %d, now %d", project.ProjectName, latestCollectionInfo.TotalCodeCount, report.Counter.IncomingCodeNodesCount),
		}
		report.Changes = append(report.Changes, codeNodeCountChange)
	}
	sumOfExpectedCodeNodes := report.Counter.UpdatedCodeNodesCount + report.Counter.UnchangedCodeNodesCount + report.Counter.NewCodeNodesCount
	if sumOfExpectedCodeNodes != report.Counter.IncomingCodeNodesCount {
		codeNodeIssue := types.Issue{
			Type: types.CodeNodeCountIssue,
			Data: fmt.Sprintf("Project %s: expected code node sum %d, got %d - updatedCount: %d + unchangedCount %d + newCount %d", project.ProjectName, sumOfExpectedCodeNodes, report.Counter.IncomingCodeNodesCount, report.Counter.UpdatedCodeNodesCount, report.Counter.UnchangedCodeNodesCount, report.Counter.NewCodeNodesCount),
		}
		report.Issues = append(report.Issues, codeNodeIssue)
	}
	if latestCollectionInfo.TotalPageCount != incomingPageCount {
		pageCountChange := types.Change{
			Type: types.ProjectSummaryPageCountChange,
			Data: fmt.Sprintf("Project %s: page count from summary was %d, now %d", project.ProjectName, latestCollectionInfo.TotalPageCount, incomingPageCount),
		}
		report.Changes = append(report.Changes, pageCountChange)
	}
	sumOfExpectedPages := len(existingPageIds) + report.Counter.NewPagesCount - report.Counter.RemovedPagesCount
	if sumOfExpectedPages != incomingPageCount {
		pageCountIssue := types.Issue{
			Type: types.PageCountIssue,
			Data: fmt.Sprintf("Project %s: expected current pages from summing changes is %d, got %d", project.ProjectName, sumOfExpectedPages, incomingPageCount),
		}
		report.Issues = append(report.Issues, pageCountIssue)
	}
	if len(report.Changes) > 0 {
		log.Printf("\nProject changes for %s\n", project.ProjectName)
		for _, change := range report.Changes {
			log.Printf("%s: %s", change.Type.String(), change.Data.(string))
		}
	} else if len(report.Changes) == 0 {
		log.Printf("\nProject changes for %s\n", project.ProjectName)
		log.Println("No changes in project")
	}
	if len(report.Issues) > 0 {
		log.Printf("\nIssues with data in project %s\n", project.ProjectName)
		for _, issue := range report.Issues {
			log.Printf("%s: %s", issue.Type.String(), issue.Data.(string))
		}
	} else if len(report.Issues) == 0 {
		log.Printf("\nNo issues with data in project %s\n", project.ProjectName)
	}
}
