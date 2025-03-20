package main

import (
	"snooty-api-parser/db"
	"snooty-api-parser/types"
	"snooty-api-parser/utils"
)

// HandleDeletedIncomingPages checks whether a page that has the `"deleted":true` flag when it comes in from the Snooty Data API
// has a corresponding page in Atlas. If it does, we delete it.
func HandleDeletedIncomingPages(collectionName string, deletedPage types.PageWrapper, report types.ProjectReport) types.ProjectReport {
	maybeAtlasId := utils.ConvertSnootyPageIdToAtlasPageId(deletedPage.Data.PageID)
	maybeAtlasDocument := db.GetAtlasPageData(collectionName, maybeAtlasId)
	if maybeAtlasDocument != nil {
		codeNodeCount := maybeAtlasDocument.CodeNodesTotal
		literalIncludeCount := maybeAtlasDocument.LiteralIncludesTotal
		ioCodeBlockCount := maybeAtlasDocument.IoCodeBlocksTotal
		pageRemoved := db.RemovePageFromAtlas(collectionName, maybeAtlasDocument.ID)
		if pageRemoved {
			report.Counter.RemovedPagesCount += 1
			report = utils.ReportChanges(types.PageRemoved, report, maybeAtlasDocument.ID)
			if codeNodeCount > 0 {
				report = utils.ReportChanges(types.CodeExampleRemoved, report, maybeAtlasDocument.ID, codeNodeCount)
				report.Counter.RemovedCodeNodesCount += codeNodeCount
			}
			if literalIncludeCount > 0 {
				report = utils.ReportChanges(types.LiteralIncludeCountChange, report, maybeAtlasDocument.ID, literalIncludeCount, 0)
			}
			if ioCodeBlockCount > 0 {
				report = utils.ReportChanges(types.IoCodeBlockCountChange, report, maybeAtlasDocument.ID, ioCodeBlockCount, 0)
			}
		} else {
			report = utils.ReportIssues(types.PageNotRemovedIssue, report, maybeAtlasDocument.ID)
		}
	}
	return report
}
