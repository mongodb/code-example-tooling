package main

import (
	"common"
	"gdcd/types"
)

func IncrementProjectCountsForExistingPage(incomingCodeNodeCount int, incomingLiteralIncludeNodeCount int, incomingIoCodeBlockNodeCount int, existingPage common.DocsPage, report types.ProjectReport) types.ProjectReport {
	report.Counter.IncomingCodeNodesCount += incomingCodeNodeCount
	report.Counter.IncomingLiteralIncludeCount += incomingLiteralIncludeNodeCount
	report.Counter.IncomingIoCodeBlockCount += incomingIoCodeBlockNodeCount
	report.Counter.ExistingCodeNodesCount += existingPage.CodeNodesTotal
	report.Counter.ExistingLiteralIncludeCount += existingPage.LiteralIncludesTotal
	report.Counter.ExistingIoCodeBlockCount += existingPage.IoCodeBlocksTotal
	return report
}
