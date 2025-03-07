package main

import "snooty-api-parser/types"

func IncrementProjectCountsForExistingPage(incomingCodeNodeCount int, incomingLiteralIncludeNodeCount int, incomingIoCodeBlockNodeCount int, existingPage types.DocsPage, projectCounter types.ProjectCounts) types.ProjectCounts {
	projectCounter.IncomingCodeNodesCount += incomingCodeNodeCount
	projectCounter.IncomingLiteralIncludeCount += incomingLiteralIncludeNodeCount
	projectCounter.IncomingIoCodeBlockCount += incomingIoCodeBlockNodeCount
	projectCounter.ExistingCodeNodesCount += existingPage.CodeNodesTotal
	projectCounter.ExistingLiteralIncludeCount += existingPage.LiteralIncludesTotal
	projectCounter.ExistingIoCodeBlockCount += existingPage.IoCodeBlocksTotal
	return projectCounter
}
