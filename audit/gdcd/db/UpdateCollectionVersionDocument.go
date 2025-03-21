package db

import (
	"common"
	"gdcd/types"
	"time"
)

func UpdateCollectionVersionDocument(existingSummaries common.CollectionReport, project types.DocsProjectDetails, report types.ProjectReport) common.CollectionReport {
	existingCollectionInfo := existingSummaries.Version[project.ActiveBranch]
	existingCollectionInfo.TotalPageCount = report.Counter.TotalCurrentPageCount
	existingCollectionInfo.TotalCodeCount = report.Counter.IncomingCodeNodesCount
	existingCollectionInfo.LastUpdatedAtUTC = time.Now().UTC()
	existingSummaries.Version[project.ActiveBranch] = existingCollectionInfo
	return existingSummaries
}
