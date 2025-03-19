package utils

import (
	"fmt"
	"snooty-api-parser/types"
)

func ReportIssues(issueType types.IssueType, report types.ProjectReport, stringArg string, counts ...int) types.ProjectReport {
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
	switch issueType {
	case types.PagesNotFoundIssue:
		message = fmt.Sprintf("No documents found for project %s", stringArg)
	case types.CodeNodeCountIssue:
		message = fmt.Sprintf("Project %s: expected %d code nodes, got %d", stringArg, count1, count2)
	case types.PageCountIssue:
		message = fmt.Sprintf("Project %s: expected current pages from summing changes is %d, got %d", stringArg, count1, count2)
	default:
		message = "Change type not handled in ReportChanges function"
	}

	issue := types.Issue{
		Type: issueType,
		Data: message,
	}
	report.Issues = append(report.Issues, issue)
	return report
}
