package db

import (
	"gdcd/types"
	"gdcd/utils"
)

// HandleMissingPageIds gets a list of all the page IDs from Atlas, compares each page ID against incoming ones coming
// in from Snooty, and tries to figure out whether existing IDs that do not match incoming ones are moved pages or removed
// pages. If the page is removed, we delete it from the DB. We pass moved and new pages back to the call site for further
// handling.
func HandleMissingPageIds(collectionName string, incomingPageIds map[string]bool, maybeNewPages []types.NewOrMovedPage, report types.ProjectReport) (types.ProjectReport, []types.NewOrMovedPage, []types.NewOrMovedPage) {
	var movedPages []types.NewOrMovedPage
	// Get a slice of all the page IDs for pages that are currently in Atlas
	existingPageIds := GetAtlasPageIDs(collectionName)
	// If we don't get any page IDs from Atlas, just return the unmodified report
	if existingPageIds == nil {
		return report, maybeNewPages, movedPages
	}
	// Compare the pages that are currently in Atlas with pages coming in from the Snooty Data API. If the page exists
	// in Atlas but isn't coming in from the Snooty Data API, grab the ID so we can remove the page in Atlas.
	var maybeRemovedPageIds []string
	for _, existingId := range existingPageIds {
		if incomingPageIds[existingId] {
			// If the page ID in Atlas matches an incoming page ID from Snooty matches, skip the rest of the loop
			continue
		}
		// If an existing ID in Atlas does not match any of the pages coming in from Snooty, add the ID to a list of pages that might be removed
		maybeRemovedPageIds = append(maybeRemovedPageIds, existingId)
	}

	var pageIdsToDelete []string

	// A page ID that isn't an exact match for one coming in from Snooty could be either a moved page or a removed page
	for _, maybeRemovedPageId := range maybeRemovedPageIds {
		existingPage := GetAtlasPageData(collectionName, maybeRemovedPageId)
		pageIsMoved := false

		// Compare the removed page against the unaccounted for pages in the collection. An incoming page that
		// does not have a matching page ID could be either moved or new. If the count of code examples, literalincludes,
		// and io-code-blocks exactly matches a removed page, we'll call it "moved" instead of "new"
		for index, maybeNewPage := range maybeNewPages {
			// If the count of code examples is exactly the same, *and* that count is not 0, they might be the same page
			codeNodeCountMatches := maybeNewPage.CodeNodeCount == existingPage.CodeNodesTotal && maybeNewPage.CodeNodeCount != 0

			// To be more precise, also check if the count of literalincludes and io-code-blocks match
			literalIncludeCountMatches := maybeNewPage.LiteralIncludeCount == existingPage.LiteralIncludesTotal
			ioCodeBlockCountMatches := maybeNewPage.IoCodeBlockCount == existingPage.IoCodeBlocksTotal

			// If all three counts match, and the code node count is not 0, consider it a moved page instead of new & removed pages
			if codeNodeCountMatches && literalIncludeCountMatches && ioCodeBlockCountMatches {
				maybeNewPage.NewPageId = maybeNewPage.PageId
				maybeNewPage.OldPageId = existingPage.ID
				maybeNewPage.PageData.DateAdded = existingPage.DateAdded
				movedPages = append(movedPages, maybeNewPage)

				// If we find a match, we can remove it from the `maybeNewPages` slice so we don't attempt to match it again
				// Anything left in the `maybeNewPages` slice after comparing all the maybe removed pages is net new, so
				// we'll pass it back to the call site to handle it as a new page
				maybeNewPages = removeMovedPage(maybeNewPages, index)
				pageIsMoved = true

				// We've found a match, so we can skip the rest of the `maybeNewPages` for this `maybeRemovedPageId`
				continue
			}
		}

		// If we have gone through all the maybe new pages, and none is an exact match in code example counts, consider
		// it a removed page
		if !pageIsMoved {
			pageIdsToDelete = append(pageIdsToDelete, maybeRemovedPageId)
		}
	}

	// Handle all the removed pages
	for _, pageIdToDelete := range pageIdsToDelete {
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

	// Anything left in the `maybeNewPages` slice at this point is net new, so we'll handle it back at the call site
	// Anything in the `movedPages` slice is moved, which we'll also handle back at the call site
	return report, maybeNewPages, movedPages
}

func removeMovedPage(maybeNewPages []types.NewOrMovedPage, index int) []types.NewOrMovedPage {
	return append(maybeNewPages[:index], maybeNewPages[index+1:]...)
}
