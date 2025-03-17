package main

import "snooty-api-parser/types"

func IncrementProjectCountsForNewPage(incomingCodeNodeCount int, incomingLiteralIncludeNodeCount int, incomingIoCodeBlockNodeCount int, newCodeNodes int, projectCounter types.ProjectCounts) types.ProjectCounts {
	projectCounter.IncomingCodeNodesCount += incomingCodeNodeCount
	projectCounter.IncomingLiteralIncludeCount += incomingLiteralIncludeNodeCount
	projectCounter.IncomingIoCodeBlockCount += incomingIoCodeBlockNodeCount
	projectCounter.NewCodeNodesCount += newCodeNodes
	return projectCounter
}
