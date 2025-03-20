package db

import (
	"gdcd/types"
	"time"
)

func UpdateCollectionVersionDocument(existingSummaries types.CollectionReport, project types.DocsProjectDetails, report types.ProjectReport) types.CollectionReport {
	existingCollectionInfo := existingSummaries.Version[project.ActiveBranch]
	existingCollectionInfo.TotalPageCount = report.Counter.TotalCurrentPageCount
	existingCollectionInfo.TotalCodeCount = report.Counter.IncomingCodeNodesCount
	existingCollectionInfo.LastUpdatedAtUTC = time.Now().UTC()
	existingSummaries.Version[project.ActiveBranch] = existingCollectionInfo
	return existingSummaries
}
