package main

import (
	"common"
	"context"
	"gdcd/db"
	"gdcd/types"
	"gdcd/utils"

	"github.com/tmc/langchaingo/llms/ollama"
)

// CheckPagesForUpdates takes the slice of incoming pages for a given project that we got from the Snooty Data API, plus
// other things initialized in main() that are needed here. We iterate through the pages in the project, checking for
// things that need to be added, removed, or updated. We compile a report for the project, which we're currently outputting
// to a log file on the local file system. Then, we perform a batch update with all the changes for this project.
func CheckPagesForUpdates(pages []types.PageWrapper, project types.ProjectDetails, llm *ollama.LLM, ctx context.Context, report types.ProjectReport) {
	incomingPageIdsMatchingExistingPages := make(map[string]bool)
	incomingDeletedPageCount := 0
	var newPageIds []string
	var newPages []common.DocsPage
	var updatedPages []common.DocsPage
	for _, page := range pages {
		// The Snooty Data API returns pages that may have been deleted. If the page is deleted, we want to check and see
		// if it exists already in the DB, and delete it if it does. If we haven't already made an entry for it, we
		// don't need to do anything else.
		if page.Data.Deleted {
			report = HandleDeletedIncomingPages(project.ProjectName, page, report)
			incomingDeletedPageCount++
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
			} else {
				// If there is no existing document in Atlas that matches the page, we need to make a new page
				var newPage common.DocsPage
				newPage, report = MakeNewPage(page, project.ProdUrl, report, llm, ctx)
				newPageIds = append(newPageIds, newPage.ID)
				newPages = append(newPages, newPage)
			}
		}
		utils.UpdateSecondaryTarget()
	}

	// After iterating through the incoming pages from the Snooty Data API, we need to figure out if any of the page IDs
	// we had in the DB are not coming in from the incoming response. If so, we should delete those entries.
	report = db.HandleMissingPageIds(project.ProjectName, incomingPageIdsMatchingExistingPages, report)

	// Get the existing "summaries" document from the DB, and update it.
	var summaryDoc common.CollectionReport

	// Adjust the total page count we're getting from Snooty to remove any 'deleted' pages - we don't want to count or track those
	report.Counter.TotalCurrentPageCount = report.Counter.TotalCurrentPageCount - incomingDeletedPageCount
	summaryDoc, report = HandleCollectionSummariesDocument(project, report)

	// Output the project report to the log
	LogReportForProject(project.ProjectName, report)

	// At this point, we have all the new and updated pages and an updated summary. Write updates to Atlas.
	db.BatchUpdateCollection(project.ProjectName, newPages, updatedPages, summaryDoc)
}
