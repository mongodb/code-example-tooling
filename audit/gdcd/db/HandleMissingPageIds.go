package db

import (
	"gdcd/types"
	"gdcd/utils"
	"strings"
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
			// If the page IDs share a page name, they might be the same page
			pageIdsOverlap := checkIfPageIdsOverlap(existingPage.ID, maybeNewPage.PageId)

			// If the count of code examples is exactly the same, *and* that count is not 0, they might be the same page
			codeNodeCountMatches := maybeNewPage.CodeNodeCount == existingPage.CodeNodesTotal && maybeNewPage.CodeNodeCount != 0

			// To be more precise, also check if the count of literalincludes and io-code-blocks match
			literalIncludeCountMatches := maybeNewPage.LiteralIncludeCount == existingPage.LiteralIncludesTotal
			ioCodeBlockCountMatches := maybeNewPage.IoCodeBlockCount == existingPage.IoCodeBlocksTotal

			// If the page name shares common elements, all three counts match, and the code node count is not 0,
			// consider it a moved page instead of new & removed pages
			if pageIdsOverlap && codeNodeCountMatches && literalIncludeCountMatches && ioCodeBlockCountMatches {
				maybeNewPage.NewPageId = maybeNewPage.PageId
				maybeNewPage.OldPageId = existingPage.ID
				maybeNewPage.DateAdded = existingPage.DateAdded
				movedPages = append(movedPages, maybeNewPage)

				// If we find a match, we can remove it from the `maybeNewPages` slice so we don't attempt to match it again
				// Anything left in the `maybeNewPages` slice after comparing all the maybe removed pages is net new, so
				// we'll pass it back to the call site to handle it as a new page
				maybeNewPages = removeMovedPage(maybeNewPages, index)
				pageIsMoved = true

				// We've found a match, so we can skip the rest of the `maybeNewPages` for this `maybeRemovedPageId`
				break
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

// checkIfPageIdsOverlap does a couple of types of comparison between the old page ID and the new page ID to determine
// if they "match".
func checkIfPageIdsOverlap(oldPageId string, newPageId string) bool {
	// First, we want to get the page title. Split by `|`. The final element will be the page title.
	// i.e. in the page ID `tutorial|create-mongodb-user-for-cluster` - the final element after the `|`,
	// `create-mongodb-user-for-cluster` - is what we're considering the page title
	oldPageName := getPageTitleFromId(oldPageId)
	newPageName := getPageTitleFromId(newPageId)
	newPageSegments := getExtendedPageTitleFromId(newPageName)

	// The simplest case is a restructure that moves the pages from one directory to another without any changes.
	// If the page name is an exact match, we can return true, because the page title overlaps 100%
	if oldPageName == newPageName {
		return true
		// In some cases, the page may have become a title page for a section, and may now have pages below it. Check
		// if the old page name is up a directory level.
	} else if contains(newPageSegments, oldPageName) {
		return true
	} else {
		// If it's not a 1:1 move the page without changing the title situation, we can compare the page titles to try
		// to figure out if it has enough overlap to be effectively the same page title
		return pageNamesHaveCommonElements(oldPageName, newPageName)
	}
}

func getPageTitleFromId(pageId string) string {
	parts := strings.Split(pageId, "|")

	// Get the last element
	if len(parts) > 0 {
		lastElement := parts[len(parts)-1] // Access the last index
		return lastElement
	} else {
		return ""
	}
}

func getExtendedPageTitleFromId(pageId string) []string {
	parts := strings.Split(pageId, "|")

	var titleSegments []string
	// Get the last element
	if len(parts) > 0 {
		lastElement := parts[len(parts)-1] // Access the last index
		titleSegments = append(titleSegments, lastElement)
	}
	// If there are multiple elements, get the second-to-last element. This may contain something that _used_ to match
	// the page ID when we are now nesting pages below it
	if len(parts) > 1 {
		secondToLastElement := parts[len(parts)-2]
		titleSegments = append(titleSegments, secondToLastElement)
	}
	return titleSegments
}

func pageNamesHaveCommonElements(oldPageName string, newPageName string) bool {
	// Split the page names by `-` to get the words in the name for common comparison
	oldPageNameParts := strings.Split(oldPageName, "-")
	newPageNameParts := strings.Split(newPageName, "-")

	// We don't want to count irrelevant words for this comparison, so compare elements against these words and omit
	// them from being counted as an overlap
	ignoreWords := []string{"and", "or", "by", "for", "the", "in"}

	oldPageNameElements := make(map[string]bool)
	for _, element := range oldPageNameParts {
		oldPageNameElements[element] = true // Mark the presence of each element in the map
	}

	// Compare with `newPageNameParts` and count common elements
	commonCount := 0
	for _, value := range newPageNameParts {
		if oldPageNameElements[value] { // Check if the element exists in the map
			// Confirm the element isn't one of the ignore words
			if !contains(ignoreWords, value) {
				// If it's not an ignore word, consider it a common element
				commonCount++
			}
		}
	}

	if commonCount > 0 {
		return true
	} else {
		return false
	}
}

func contains(slice []string, str string) bool {
	for _, value := range slice {
		if value == str {
			return true // Return true if the string is found
		}
	}
	return false // Return false if the string is not found
}

func removeMovedPage(maybeNewPages []types.NewOrMovedPage, index int) []types.NewOrMovedPage {
	return append(maybeNewPages[:index], maybeNewPages[index+1:]...)
}
