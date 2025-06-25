package main

import (
	"common"
	"gdcd/add-code-examples"
	"gdcd/types"
	"gdcd/utils"
)

func UpdateProjectReportForNewPage(page common.DocsPage, report types.ProjectReport) types.ProjectReport {
	report.Counter.IncomingCodeNodesCount += page.CodeNodesTotal
	report.Counter.IncomingLiteralIncludeCount += page.LiteralIncludesTotal
	report.Counter.IncomingIoCodeBlockCount += page.IoCodeBlocksTotal
	report.Counter.NewCodeNodesCount += page.CodeNodesTotal
	report.Counter.NewPagesCount += 1
	report = utils.ReportChanges(types.PageCreated, report, page.ID)
	if page.CodeNodesTotal > 0 {
		report = utils.ReportChanges(types.CodeExampleCreated, report, page.ID, page.CodeNodesTotal)
	}

	// Figure out how many of the page's code examples are new applied usage examples
	newAppliedUsageExampleCount := 0
	newCodeNodeCount := 0
	if page.Nodes != nil {
		for _, node := range *page.Nodes {
			if add_code_examples.IsNewAppliedUsageExample(node) {
				newAppliedUsageExampleCount++
			}
		}
		newCodeNodeCount = len(*page.Nodes)
	}
	report.Counter.NewAppliedUsageExamplesCount += newAppliedUsageExampleCount

	if newAppliedUsageExampleCount > 0 {
		report = utils.ReportChanges(types.AppliedUsageExampleAdded, report, page.ID, newAppliedUsageExampleCount)
	}

	if page.CodeNodesTotal != newCodeNodeCount {
		report = utils.ReportIssues(types.CodeNodeCountIssue, report, page.ID, page.CodeNodesTotal, newCodeNodeCount)
	}
	return report
}
