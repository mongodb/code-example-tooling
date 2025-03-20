package compare_code_examples

import (
	"gdcd/types"
	"gdcd/utils"
)

func UpdateProjectReportForUpdatedCodeNodes(report types.ProjectReport, pageId string, incomingCount int, existingCount int, unchangedCount int, updatedCount int, newCount int, removedCount int, aggregateCodeNodeChangeCount int, newAppliedUsageExampleCount int) types.ProjectReport {
	if newCount > 0 {
		report.Counter.NewCodeNodesCount += newCount
		report = utils.ReportChanges(types.CodeExampleCreated, report, pageId, newCount)
	}
	if unchangedCount > 0 {
		report.Counter.UnchangedCodeNodesCount += unchangedCount
	}
	if updatedCount > 0 {
		report.Counter.UpdatedCodeNodesCount += updatedCount
		report = utils.ReportChanges(types.CodeExampleUpdated, report, pageId, updatedCount)
	}
	if removedCount > 0 {
		report.Counter.RemovedCodeNodesCount += removedCount
		report = utils.ReportChanges(types.CodeExampleRemoved, report, pageId, removedCount)
	}
	if newAppliedUsageExampleCount > 0 {
		report.Counter.NewAppliedUsageExamplesCount += newAppliedUsageExampleCount
		report = utils.ReportChanges(types.AppliedUsageExampleAdded, report, pageId, newAppliedUsageExampleCount)
	}
	countForIncomingChanges := unchangedCount + updatedCount + newCount
	if countForIncomingChanges != incomingCount {
		report = utils.ReportIssues(types.CodeNodeCountIssue, report, pageId, countForIncomingChanges, incomingCount)
	}
	countFromExisting := aggregateCodeNodeChangeCount - removedCount
	if countFromExisting != incomingCount {
		report = utils.ReportIssues(types.CodeNodeCountIssue, report, pageId, countFromExisting, incomingCount)
	}
	return report
}
