package main

import (
	"context"
	"github.com/tmc/langchaingo/llms/ollama"
	"snooty-api-parser/db"
	"snooty-api-parser/types"
	"snooty-api-parser/utils"
)

// CheckDocsForUpdates takes the slice of incoming pages for a given project that we got from the Snooty Data API, plus
// other things initialized in main() that are needed here. We iterate through the pages in the project, checking for
// things that need to be added, removed, or updated. We compile a report for the project, which we're currently outputting
// to a log file on the local file system. Then, we perform a batch update with all the changes for this project.
func CheckDocsForUpdates(docsPages []types.PageWrapper, project types.DocsProjectDetails, llm *ollama.LLM, ctx context.Context, report types.ProjectReport) {
	incomingPageIdsMatchingExistingPages := make(map[string]bool)
	incomingPageCount := len(docsPages)
	incomingDeletedPageCount := 0
	var newPageIds []string
	var newPages []types.DocsPage
	var updatedPages []types.DocsPage
	for _, page := range docsPages {
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
				var updatedPage *types.DocsPage
				updatedPage, report = UpdateExistingDocsPage(*maybeExistingPage, page, report, llm, ctx)
				if updatedPage != nil {
					updatedPages = append(updatedPages, *updatedPage)
				}
			} else {
				// If there is no existing document in Atlas that matches the page, we need to make a new page
				var newPage types.DocsPage
				newPage, report = MakeNewDocsPage(page, project.ProdUrl, report, llm, ctx)
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
	var summaryDoc types.CollectionReport
	summaryDoc, report = HandleCollectionSummariesDocument(project, report, incomingPageCount)

	// Output the project report to the log
	LogReportForProject(project.ProjectName, report)

	// At this point, we have all the new and updated pages and an updated summary. Write updates to Atlas.
	db.BatchUpdateCollection(project.ProjectName, newPages, updatedPages, summaryDoc)
}
