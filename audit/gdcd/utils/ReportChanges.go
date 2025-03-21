package utils

import (
	"fmt"
	"gdcd/types"
)

func ReportChanges(changeType types.ChangeType, report types.ProjectReport, stringArg string, counts ...int) types.ProjectReport {
	numberOfCounts := len(counts)
	var count1 int
	var count2 int
	switch numberOfCounts {
	case 1:
		count1 = counts[0]
	case 2:
		count1 = counts[0]
		count2 = counts[1]
	default:
		// Counts could be an empty array, in which case we do nothing
	}

	var message string
	switch changeType {
	case types.PageCreated:
		message = fmt.Sprintf("Page ID: %s", stringArg)
	case types.PageUpdated:
		message = fmt.Sprintf("Page ID: %s", stringArg)
	case types.PageRemoved:
		message = fmt.Sprintf("Page ID: %s", stringArg)
	case types.KeywordsUpdated:
		message = fmt.Sprintf("Page ID: %s", stringArg)
	case types.CodeExampleCreated:
		message = fmt.Sprintf("Page ID: %s, %d new code examples added", stringArg, count1)
	case types.CodeExampleUpdated:
		message = fmt.Sprintf("Page ID: %s, %d code examples updated", stringArg, count1)
	case types.CodeExampleRemoved:
		message = fmt.Sprintf("Page ID: %s, %d code examples removed", stringArg, count1)
	case types.CodeNodeCountChange:
		message = fmt.Sprintf("Page ID: %s, code node count was: %d, now %d", stringArg, count1, count2)
	case types.LiteralIncludeCountChange:
		message = fmt.Sprintf("Page ID: %s, literalinclude count was %d, now %d", stringArg, count1, count2)
	case types.IoCodeBlockCountChange:
		message = fmt.Sprintf("Page ID: %s, io-code-block count was %d, now %d", stringArg, count1, count2)
	case types.ProjectSummaryCodeNodeCountChange:
		message = fmt.Sprintf("Project %s: code node count from summary was %d, now %d", stringArg, count1, count2)
	case types.ProjectSummaryPageCountChange:
		message = fmt.Sprintf("Project %s: page count from summary was %d, now %d", stringArg, count1, count2)
	case types.AppliedUsageExampleAdded:
		message = fmt.Sprintf("Page ID: %s, %d new applied usage examples added", stringArg, count1)
	default:
		message = "Change type not handled in ReportChanges function"
	}

	change := types.Change{
		Type: changeType,
		Data: message,
	}
	report.Changes = append(report.Changes, change)
	return report
}
