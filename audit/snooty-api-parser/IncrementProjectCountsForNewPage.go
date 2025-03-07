package main

import "snooty-api-parser/types"

func IncrementProjectCountsForNewPage(incomingCodeNodeCount int, incomingLiteralIncludeNodeCount int, incomingIoCodeBlockNodeCount int, projectCounter types.ProjectCounts) types.ProjectCounts {
	projectCounter.IncomingCodeNodesCount += incomingCodeNodeCount
	projectCounter.IncomingLiteralIncludeCount += incomingLiteralIncludeNodeCount
	projectCounter.IncomingIoCodeBlockCount += incomingIoCodeBlockNodeCount
	return projectCounter
}
