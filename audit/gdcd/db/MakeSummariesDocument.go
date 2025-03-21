package db

import (
	"common"
	"gdcd/types"
	"time"
)

func MakeSummariesDocument(project types.DocsProjectDetails, report types.ProjectReport) common.CollectionReport {
	collectionInfo := common.CollectionInfoView{
		TotalPageCount:   report.Counter.NewPagesCount,
		TotalCodeCount:   report.Counter.NewCodeNodesCount,
		LastUpdatedAtUTC: time.Now().UTC(),
	}
	versionMap := make(map[string]common.CollectionInfoView)
	versionMap[project.ActiveBranch] = collectionInfo
	collectionReport := common.CollectionReport{
		ID:      "summaries",
		Version: versionMap,
	}
	return collectionReport
}
