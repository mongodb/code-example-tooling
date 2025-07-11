package main

import (
	"common"
	"context"
	"fmt"
	"gdcd/db"
	"gdcd/snooty"
	"gdcd/types"
	"gdcd/utils"
	"time"

	"github.com/tmc/langchaingo/llms/ollama"
)

// CheckPagesForUpdates takes the slice of incoming pages for a given project that we got from the Snooty Data API, plus
// other things initialized in main() that are needed here. We iterate through the pages in the project, checking for
// things that need to be added, removed, or updated. We compile a report for the project, which we're currently outputting
// to a log file on the local file system. Then, we perform a batch update with all the changes for this project.
func CheckPagesForUpdates(pages []types.PageWrapper, project types.ProjectDetails, llm *ollama.LLM, ctx context.Context, report types.ProjectReport) {
	incomingPageIdsMatchingExistingPages := make(map[string]bool)
	incomingDeletedPageCount := 0

	// When a page doesn't match one in the DB, it could be either net new or a moved page. Hold it in a temp array
	// for comparison
	var maybeNewPages []types.NewOrMovedPage
	var newPages []types.NewOrMovedPage
	var newPageDBEntries []common.DocsPage
	var movedPages []types.NewOrMovedPage
	var updatedPages []common.DocsPage
	for _, page := range pages {
		// The Snooty Data API returns pages that may have been deleted. If the page is deleted, we want to check and see
		// if it exists already in the DB, and delete it if it does. If we haven't already made an entry for it, we
		// don't need to do anything else.
		if page.Data.Deleted {
			report = HandleDeletedIncomingPages(project.ProjectName, page, report)
			incomingDeletedPageCount++
			utils.UpdateSecondaryTarget()
		} else {
			maybeExistingPage := CheckForExistingPage(project.ProjectName, page)
			if maybeExistingPage != nil {
				// If there is an existing document in Atlas, update the existing page
				// If the code example counts are the same on the incoming page as they are on the existing page,
				// we treat that as an unchanged page and it does not return an updated page - it returns nil
				incomingPageIdsMatchingExistingPages[maybeExistingPage.ID] = true
				var updatedPage *common.DocsPage
				updatedPage, report = UpdateExistingPage(*maybeExistingPage, page, report, llm, ctx)
				if updatedPage != nil {
					updatedPages = append(updatedPages, *updatedPage)
				}
				utils.UpdateSecondaryTarget()
			} else {
				// If there is no existing document in Atlas that matches the page, we need to make a new page. BUT!
				// It might actually be a new or moved page. So store it in a temp `maybeNewPages` slice so we can compare
				// it against removed pages later and potentially call it a "moved" page, instead.
				newOrMovedPage := getNewOrMovedPageDetails(page.Data)
				maybeNewPages = append(maybeNewPages, newOrMovedPage)
			}
		}
	}

	// After iterating through the incoming pages from the Snooty Data API, we need to figure out if any of the page IDs
	// we had in the DB are not coming in from the incoming response. If so, those pages are either moved or removed.
	report, newPages, movedPages = db.HandleMissingPageIds(project.ProjectName, incomingPageIdsMatchingExistingPages, maybeNewPages, report)

	// If we have new pages, create the corresponding DocsPage and increment the project report for them
	if newPages != nil {
		for _, page := range newPages {
			newPage := MakeNewPage(page.PageData, project.ProjectName, project.ProdUrl, llm, ctx)
			newPageDBEntries = append(newPageDBEntries, newPage)
			report = UpdateProjectReportForNewPage(newPage, report)
			utils.UpdateSecondaryTarget()
		}
	}

	// If we have moved pages, handle them
	if movedPages != nil {
		for _, page := range movedPages {
			var movedPage common.DocsPage
			oldPage := db.GetAtlasPageData(project.ProjectName, page.OldPageId)

			if oldPage != nil {
				movedPage = *oldPage
				movedPage.ID = page.NewPageId
				newPageUrl := utils.ConvertAtlasPageIdToProductionUrl(page.NewPageId, project.ProdUrl)
				movedPage.DateLastUpdated = time.Now()
				movedPage.PageURL = newPageUrl
			} else {
				movedPage = MakeNewPage(page.PageData, project.ProjectName, project.ProdUrl, llm, ctx)
				movedPage.DateAdded = page.DateAdded
			}

			// Remove the old page from the DB
			db.RemovePageFromAtlas(project.ProjectName, page.OldPageId)

			// Append the "moved" page to the `newPageDBEntries` array. Because the page ID doesn't match the old one,
			// we write it to the DB as a new page. Because we just deleted the old page, it works out to the same count
			// and provides the up-to-date data in the DB.
			newPageDBEntries = append(newPageDBEntries, movedPage)

			incomingAstCodeNodes, incomingAstLiteralIncludeNodes, incomingAstIoCodeBlockNodes := snooty.GetCodeExamplesFromIncomingData(page.PageData.AST)
			incomingAstCodeNodeCount := len(incomingAstCodeNodes)
			incomingAstLiteralIncludeNodesCount := len(incomingAstLiteralIncludeNodes)
			incomingAstIoCodeBlockNodesCount := len(incomingAstIoCodeBlockNodes)
			// Update the project counts for the "existing" page
			report = IncrementProjectCountsForExistingPage(incomingAstCodeNodeCount, incomingAstLiteralIncludeNodesCount, incomingAstIoCodeBlockNodesCount, movedPage, report)

			// Report it in the logs as a moved page
			stringMessageForReport := fmt.Sprintf("Old page ID: %s, new page ID: %s", page.OldPageId, page.NewPageId)
			report = utils.ReportChanges(types.PageMoved, report, stringMessageForReport)
			if movedPage.CodeNodesTotal != incomingAstCodeNodeCount {
				utils.ReportIssues(types.CodeNodeCountIssue, report, page.NewPageId, page.CodeNodeCount, len(incomingAstCodeNodes))
			}
			utils.UpdateSecondaryTarget()
		}
	}

	// Get the existing "summaries" document from the DB, and update it.
	var summaryDoc common.CollectionReport

	// Adjust the total page count we're getting from Snooty to remove any 'deleted' pages - we don't want to count or track those
	report.Counter.TotalCurrentPageCount = report.Counter.TotalCurrentPageCount - incomingDeletedPageCount
	summaryDoc, report = HandleCollectionSummariesDocument(project, report)

	// Output the project report to the log
	LogReportForProject(project.ProjectName, report)

	// At this point, we have all the new and updated pages and an updated summary. Write updates to Atlas.
	db.BatchUpdateCollection(project.ProjectName, newPageDBEntries, updatedPages, summaryDoc)
}

func getNewOrMovedPageDetails(metadata types.PageMetadata) types.NewOrMovedPage {
	incomingCodeNodes, incomingLiteralIncludeNodes, incomingIoCodeBlockNodes := snooty.GetCodeExamplesFromIncomingData(metadata.AST)
	pageId := utils.ConvertSnootyPageIdToAtlasPageId(metadata.PageID)
	return types.NewOrMovedPage{
		PageId:              pageId,
		CodeNodeCount:       len(incomingCodeNodes),
		LiteralIncludeCount: len(incomingLiteralIncludeNodes),
		IoCodeBlockCount:    len(incomingIoCodeBlockNodes),
		PageData:            metadata,
	}
}
