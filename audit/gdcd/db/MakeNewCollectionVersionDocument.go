package db

import (
	"gdcd/types"
	"time"
)

func MakeNewCollectionVersionDocument(existingSummaries types.CollectionReport, project types.DocsProjectDetails, report types.ProjectReport) types.CollectionReport {
	collectionInfo := types.CollectionInfoView{
		TotalPageCount:   report.Counter.TotalCurrentPageCount,
		TotalCodeCount:   report.Counter.IncomingCodeNodesCount,
		LastUpdatedAtUTC: time.Now().UTC(),
	}
	existingSummaries.Version[project.ActiveBranch] = collectionInfo
	return existingSummaries
}
