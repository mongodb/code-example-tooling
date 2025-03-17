package db

import (
	"snooty-api-parser/types"
	"time"
)

func MakeNewCollectionVersionDocument(existingSummaries types.CollectionReport, project types.DocsProjectDetails, report types.ProjectReport) types.CollectionReport {
	collectionInfo := types.CollectionInfoView{
		TotalPageCount:   report.Counter.NewPagesCount,
		TotalCodeCount:   report.Counter.NewCodeNodesCount,
		LastUpdatedAtUTC: time.Now().UTC(),
	}
	existingSummaries.Version[project.ActiveBranch] = collectionInfo
	return existingSummaries
}
