package db

import (
	"gdcd/types"
	"gdcd/utils"
)

func HandleMissingPageIds(collectionName string, incomingPageIds map[string]bool, report types.ProjectReport) types.ProjectReport {
	// Get a slice of all the page IDs for pages that are currently in Atlas
	existingPageIds := GetAtlasPageIDs(collectionName)
	var missingPageIds []string
	if existingPageIds != nil {
		// Compare the pages that are currently in Atlas with pages coming in from the Snooty Data API. If the page exists
		// in Atlas but isn't coming in from the Snooty Data API, grab the ID so we can remove the page in Atlas.
		for _, existingId := range existingPageIds {
			if !incomingPageIds[existingId] {
				missingPageIds = append(missingPageIds, existingId)
			}
		}
	}
	for _, missingPageId := range missingPageIds {
		// We want to report details for the page we're about to delete, so we need to pull up the page to get the details
		existingPage := GetAtlasPageData(collectionName, missingPageId)
		codeNodeCount := existingPage.CodeNodesTotal
		literalIncludeCount := existingPage.LiteralIncludesTotal
		ioCodeBlockCount := existingPage.IoCodeBlocksTotal
		pageRemoved := RemovePageFromAtlas(collectionName, missingPageId)
		if pageRemoved {
			report.Counter.RemovedPagesCount += 1
			report = utils.ReportChanges(types.PageRemoved, report, missingPageId)
			if codeNodeCount > 0 {
				report = utils.ReportChanges(types.CodeExampleRemoved, report, missingPageId, codeNodeCount)
				report.Counter.RemovedCodeNodesCount += codeNodeCount
			}
			if literalIncludeCount > 0 {
				report = utils.ReportChanges(types.LiteralIncludeCountChange, report, missingPageId, literalIncludeCount, 0)
			}
			if ioCodeBlockCount > 0 {
				report = utils.ReportChanges(types.IoCodeBlockCountChange, report, missingPageId, ioCodeBlockCount, 0)
			}
		} else {
			report = utils.ReportIssues(types.PageNotRemovedIssue, report, missingPageId)
		}
	}
	return report
}
