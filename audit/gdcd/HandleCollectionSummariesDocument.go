package main

import (
	"common"
	"gdcd/db"
	"gdcd/types"
	"gdcd/utils"
)

func HandleCollectionSummariesDocument(project types.ProjectDetails, report types.ProjectReport) (common.CollectionReport, types.ProjectReport) {
	summaryDoc := db.GetAtlasProjectSummaryData(project.ProjectName)
	var latestCollectionInfo common.CollectionInfoView
	collectionVersionKey := ""
	// If we haven't audited this collection before, there will be no collection info document
	if summaryDoc == nil {
		return db.MakeSummariesDocument(project, report), report
	} else {
		// If we have retrieved a summary doc from the DB, it may contain more than one version
		elementIndex := 0
		for version, info := range summaryDoc.Version {
			// Iterate through the version list. Set the collection info and version key for the first value in the list.
			if elementIndex == 0 {
				latestCollectionInfo = info
				collectionVersionKey = version
				// If there is more than one version in the version list, increment the element index and continue checking versions
				if len(summaryDoc.Version) > 1 {
					elementIndex++
				}
			} else {
				// If we are looking at the 2nd or later element in the list, compare the LastUpdatedAtUTC date to the one
				// that is currently set. If this version has been updated more recently, set this version as the "latest"
				// collection info.
				if info.LastUpdatedAtUTC.After(latestCollectionInfo.LastUpdatedAtUTC) {
					latestCollectionInfo = info
					collectionVersionKey = version
					if elementIndex < len(summaryDoc.Version) {
						elementIndex++
					}
				}
			}
		}
	}
	if project.Version != collectionVersionKey {
		// If the active branch doesn't match the most recent version, need to make a new CollectionInfoView for this document
		updatedSummaryDoc := db.MakeNewCollectionVersionDocument(*summaryDoc, project, report)
		summaryDoc = &updatedSummaryDoc
	} else {
		// If the active branch does match the most recent version, just need to update this version document's last updated date and counts
		pageCountBeforeUpdating := summaryDoc.Version[project.Version].TotalPageCount
		updatedSummaryDoc := db.UpdateCollectionVersionDocument(*summaryDoc, project, report)
		summaryDoc = &updatedSummaryDoc

		// If we take the total pages from the last version of the summary (pageCountBeforeUpdating), add the count of
		// new pages, and subtract the count of removed pages, we should have the same number as
		// report.Counter.TotalCurrentPageCount. TotalCurrentPageCount is the length of the docs pages array we get from
		// Snooty, minus any pages where the deleted flag is true, to reflect the total current count of docs pages in the project.
		sumOfExpectedPages := pageCountBeforeUpdating + report.Counter.NewPagesCount - report.Counter.RemovedPagesCount
		if sumOfExpectedPages != report.Counter.TotalCurrentPageCount {
			report = utils.ReportIssues(types.PageCountIssue, report, project.ProjectName, sumOfExpectedPages, report.Counter.TotalCurrentPageCount)
		}
	}

	if latestCollectionInfo.TotalCodeCount != report.Counter.IncomingCodeNodesCount {
		report = utils.ReportChanges(types.ProjectSummaryCodeNodeCountChange, report, project.ProjectName, latestCollectionInfo.TotalCodeCount, report.Counter.IncomingCodeNodesCount)
	}
	sumOfExpectedCodeNodes := report.Counter.UpdatedCodeNodesCount + report.Counter.UnchangedCodeNodesCount + report.Counter.NewCodeNodesCount
	if sumOfExpectedCodeNodes != report.Counter.IncomingCodeNodesCount {
		report = utils.ReportIssues(types.CodeNodeCountIssue, report, project.ProjectName, sumOfExpectedCodeNodes, report.Counter.IncomingCodeNodesCount)
	}
	if latestCollectionInfo.TotalPageCount != report.Counter.TotalCurrentPageCount {
		report = utils.ReportChanges(types.ProjectSummaryPageCountChange, report, project.ProjectName, latestCollectionInfo.TotalPageCount, report.Counter.TotalCurrentPageCount)
	}
	return *summaryDoc, report
}
