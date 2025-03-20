package db

import (
	"gdcd/types"
	"time"
)

func MakeSummariesDocument(project types.DocsProjectDetails, report types.ProjectReport) types.CollectionReport {
	collectionInfo := types.CollectionInfoView{
		TotalPageCount:   report.Counter.NewPagesCount,
		TotalCodeCount:   report.Counter.NewCodeNodesCount,
		LastUpdatedAtUTC: time.Now().UTC(),
	}
	versionMap := make(map[string]types.CollectionInfoView)
	versionMap[project.ActiveBranch] = collectionInfo
	collectionReport := types.CollectionReport{
		ID:      "summaries",
		Version: versionMap,
	}
	return collectionReport
}
