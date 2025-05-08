package db

import (
	"gdcd/types"
	"gdcd/utils"
)

func HandleMissingPageIds(collectionName string, incomingPageIds map[string]bool, report types.ProjectReport) types.ProjectReport {
	// Get a slice of all the page IDs for pages that are currently in Atlas
	existingPageIds := GetAtlasPageIDs(collectionName)
	// If we don't get any page IDs from Atlas, just return the unmodified report
	if existingPageIds == nil {
		return report
	}
	// Compare the pages that are currently in Atlas with pages coming in from the Snooty Data API. If the page exists
	// in Atlas but isn't coming in from the Snooty Data API, grab the ID so we can remove the page in Atlas.
	// TODO: There may be a logic issue here. When we could not retrieve the page ID from the DB; the page was getting
	//  deleted. That suggests some logic is backward here, but I can't see a logic issue. Revisit if this still appears
	//  to be a problem now that the DB retrieval func has retry logic. (And/or add testing for this!)
	var pageIdsWithNoMatchingSnootyPage []string
	for _, existingId := range existingPageIds {
		if incomingPageIds[existingId] {
			// If the page ID in Atlas matches an incoming page ID from Snooty matches, skip the rest of the loop
			continue
		}
		// If an existing ID in Atlas does not match any of the pages coming in from Snooty, add the ID to a list of pages we should delete
		pageIdsWithNoMatchingSnootyPage = append(pageIdsWithNoMatchingSnootyPage, existingId)
	}
	for _, pageIdToDelete := range pageIdsWithNoMatchingSnootyPage {
		// We want to report details for the page we're about to delete, so we need to pull up the page to get the details
		existingPage := GetAtlasPageData(collectionName, pageIdToDelete)
		codeNodeCount := existingPage.CodeNodesTotal
		literalIncludeCount := existingPage.LiteralIncludesTotal
		ioCodeBlockCount := existingPage.IoCodeBlocksTotal
		pageRemoved := RemovePageFromAtlas(collectionName, pageIdToDelete)
		if pageRemoved {
			report.Counter.RemovedPagesCount += 1
			report = utils.ReportChanges(types.PageRemoved, report, pageIdToDelete)
			if codeNodeCount > 0 {
				report = utils.ReportChanges(types.CodeExampleRemoved, report, pageIdToDelete, codeNodeCount)
				report.Counter.RemovedCodeNodesCount += codeNodeCount
			}
			if literalIncludeCount > 0 {
				report = utils.ReportChanges(types.LiteralIncludeCountChange, report, pageIdToDelete, literalIncludeCount, 0)
			}
			if ioCodeBlockCount > 0 {
				report = utils.ReportChanges(types.IoCodeBlockCountChange, report, pageIdToDelete, ioCodeBlockCount, 0)
			}
		} else {
			report = utils.ReportIssues(types.PageNotRemovedIssue, report, pageIdToDelete)
		}
	}
	return report
}
