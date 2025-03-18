package compare_code_examples

import (
	"fmt"
	"snooty-api-parser/types"
)

func UpdateProjectReportForUpdatedCodeNodes(report types.ProjectReport, pageId string, incomingCount int, existingCount int, unchangedCount int, updatedCount int, newCount int, removedCount int, aggregateCodeNodeChangeCount int, newAppliedUsageExampleCount int) types.ProjectReport {
	if newCount > 0 {
		report.Counter.NewCodeNodesCount += newCount
		newChange := types.Change{
			Type: types.CodeExampleCreated,
			Data: fmt.Sprintf("Page ID: %s, created %d new code examples", pageId, newCount),
		}
		report.Changes = append(report.Changes, newChange)
	}
	if unchangedCount > 0 {
		report.Counter.UnchangedCodeNodesCount += unchangedCount
	}
	if updatedCount > 0 {
		report.Counter.UpdatedCodeNodesCount += updatedCount
		removedChange := types.Change{
			Type: types.CodeExampleUpdated,
			Data: fmt.Sprintf("Page ID: %s, updated %d code examples", pageId, updatedCount),
		}
		report.Changes = append(report.Changes, removedChange)
	}
	if removedCount > 0 {
		report.Counter.RemovedCodeNodesCount += removedCount
		removedChange := types.Change{
			Type: types.CodeExampleRemoved,
			Data: fmt.Sprintf("Page ID: %s, removed %d code examples", pageId, removedCount),
		}
		report.Changes = append(report.Changes, removedChange)
	}
	if newAppliedUsageExampleCount > 0 {
		report.Counter.NewAppliedUsageExamplesCount += newAppliedUsageExampleCount
		newAppliedUsageExamplesChange := types.Change{
			Type: types.AppliedUsageExampleAdded,
			Data: fmt.Sprintf("Page ID: %s, %d new applied usage examples", pageId, newAppliedUsageExampleCount),
		}
		report.Changes = append(report.Changes, newAppliedUsageExamplesChange)
	}
	countForIncomingChanges := unchangedCount + updatedCount + newCount
	if countForIncomingChanges != incomingCount {
		issue := types.Issue{
			Type: types.CodeNodeCountIssue,
			Data: fmt.Sprintf("Page ID: %s, unchanged count %d + new count %d + updated count %d  = sum %d, != incoming count %d", pageId, unchangedCount, newCount, updatedCount, countForIncomingChanges, incomingCount),
		}
		report.Issues = append(report.Issues, issue)
	}
	countFromExisting := aggregateCodeNodeChangeCount - removedCount
	if countFromExisting != incomingCount {
		issue := types.Issue{
			Type: types.CodeNodeCountIssue,
			Data: fmt.Sprintf("Page ID: %s, aggregate change count %d - removed count %d = sum %d, != incoming count %d", pageId, aggregateCodeNodeChangeCount, removedCount, countFromExisting, incomingCount),
		}
		report.Issues = append(report.Issues, issue)
	}
	return report
}
