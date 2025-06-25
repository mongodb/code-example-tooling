package db

import (
	"common"
	"gdcd/types"
	"time"
)

func MakeNewCollectionVersionDocument(existingSummaries common.CollectionReport, project types.ProjectDetails, report types.ProjectReport) common.CollectionReport {
	collectionInfo := common.CollectionInfoView{
		TotalPageCount:   report.Counter.TotalCurrentPageCount,
		TotalCodeCount:   report.Counter.IncomingCodeNodesCount,
		LastUpdatedAtUTC: time.Now().UTC(),
	}
	existingSummaries.Version[project.ActiveBranch] = collectionInfo
	return existingSummaries
}
