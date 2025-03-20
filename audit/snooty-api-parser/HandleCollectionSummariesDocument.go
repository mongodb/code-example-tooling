package main

import (
	"snooty-api-parser/db"
	"snooty-api-parser/types"
	"snooty-api-parser/utils"
)

func HandleCollectionSummariesDocument(project types.DocsProjectDetails, report types.ProjectReport, incomingPageCount int) (types.CollectionReport, types.ProjectReport) {
	summaryDoc := db.GetAtlasProjectSummaryData(project.ProjectName)
	var latestCollectionInfo types.CollectionInfoView
	collectionVersionKey := ""
	// If we haven't audited this collection before, there will be no collection info document
	if summaryDoc == nil {
		summaryDocPointer := db.MakeSummariesDocument(project, report)
		summaryDoc = &summaryDocPointer
		// If we're making a new summaries document, the only key is the current active branch
		latestCollectionInfo = summaryDoc.Version[project.ActiveBranch]
		collectionVersionKey = project.ActiveBranch
	} else {
		// If we have retrieved a summary doc from the DB, it may contain more than one version
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
	var pageCountBeforeUpdating int
	if project.ActiveBranch != collectionVersionKey {
		// If the active branch doesn't match the most recent version, need to make a new CollectionInfoView for this document
		updatedSummaryDoc := db.MakeNewCollectionVersionDocument(*summaryDoc, project, report)
		summaryDoc = &updatedSummaryDoc
	} else {
		// If the active branch does match the most recent version, just need to update this version document's last updated date and counts
		pageCountBeforeUpdating = summaryDoc.Version[project.ActiveBranch].TotalPageCount
		updatedSummaryDoc := db.UpdateCollectionVersionDocument(*summaryDoc, project, report)
		summaryDoc = &updatedSummaryDoc
	}

	if latestCollectionInfo.TotalCodeCount != report.Counter.IncomingCodeNodesCount {
		report = utils.ReportChanges(types.ProjectSummaryCodeNodeCountChange, report, project.ProjectName, latestCollectionInfo.TotalCodeCount, report.Counter.IncomingCodeNodesCount)
	}
	sumOfExpectedCodeNodes := report.Counter.UpdatedCodeNodesCount + report.Counter.UnchangedCodeNodesCount + report.Counter.NewCodeNodesCount
	if sumOfExpectedCodeNodes != report.Counter.IncomingCodeNodesCount {
		report = utils.ReportIssues(types.CodeNodeCountIssue, report, project.ProjectName, sumOfExpectedCodeNodes, report.Counter.IncomingCodeNodesCount)
	}
	if latestCollectionInfo.TotalPageCount != incomingPageCount {
		report = utils.ReportChanges(types.ProjectSummaryPageCountChange, report, project.ProjectName, latestCollectionInfo.TotalPageCount, incomingPageCount)
	}
	sumOfExpectedPages := pageCountBeforeUpdating + report.Counter.NewPagesCount - report.Counter.RemovedPagesCount
	if sumOfExpectedPages != incomingPageCount {
		report = utils.ReportIssues(types.PageCountIssue, report, project.ProjectName, sumOfExpectedPages, incomingPageCount)
	}
	return *summaryDoc, report
}
