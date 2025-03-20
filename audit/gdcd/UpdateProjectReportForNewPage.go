package main

import (
	"gdcd/types"
	"gdcd/utils"
)

func UpdateProjectReportForNewPage(incomingCodeNodeCount int, incomingLiteralIncludeNodeCount int, incomingIoCodeBlockNodeCount int, newCodeNodes int, newAppliedUsageExampleCount int, pageId string, report types.ProjectReport) types.ProjectReport {
	report.Counter.IncomingCodeNodesCount += incomingCodeNodeCount
	report.Counter.IncomingLiteralIncludeCount += incomingLiteralIncludeNodeCount
	report.Counter.IncomingIoCodeBlockCount += incomingIoCodeBlockNodeCount
	report.Counter.NewCodeNodesCount += newCodeNodes
	report.Counter.NewAppliedUsageExamplesCount += newAppliedUsageExampleCount
	report.Counter.NewPagesCount += 1
	report = utils.ReportChanges(types.PageCreated, report, pageId)
	if newCodeNodes > 0 {
		report = utils.ReportChanges(types.CodeExampleCreated, report, pageId, newCodeNodes)
	}

	if newAppliedUsageExampleCount > 0 {
		report = utils.ReportChanges(types.AppliedUsageExampleAdded, report, pageId, newAppliedUsageExampleCount)
	}

	newCodeNodeCount := newCodeNodes
	if incomingCodeNodeCount != newCodeNodeCount {
		report = utils.ReportIssues(types.CodeNodeCountIssue, report, pageId, incomingCodeNodeCount, newCodeNodeCount)
	}
	return report
}
