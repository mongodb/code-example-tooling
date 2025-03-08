package main

import (
	"snooty-api-parser/types"
	"time"
)

func MakeNewDocsPage(data types.PageWrapper, siteUrl string, projectName string, projectCounter types.ProjectCounts) (types.DocsPage, types.ProjectCounts) {
	incomingCodeNodes, incomingLiteralIncludeNodes, incomingIoCodeBlockNodes := GetCodeExamplesFromIncomingData(data.Data.AST)
	incomingCodeNodeCount := len(incomingCodeNodes)
	incomingLiteralIncludeNodeCount := len(incomingLiteralIncludeNodes)
	incomingIoCodeNodeCount := len(incomingIoCodeBlockNodes)
	projectCounter = IncrementProjectCountsForNewPage(incomingCodeNodeCount, incomingLiteralIncludeNodeCount, incomingIoCodeNodeCount, projectCounter)
	pageId := getPageId(data.Data.PageID)
	pageUrl := ConvertPageIdToProductionUrl(data.Data.PageID, siteUrl)
	var newCodeNodes []types.CodeNode
	for _, node := range incomingCodeNodes {
		newNode := MakeCodeNodeFromSnootyAST(node)
		newCodeNodes = append(newCodeNodes, newNode)
	}
	// TODO: Populate Product, Sub-Product and Languages for page
	return types.DocsPage{
		ID:                   pageId,
		CodeNodesTotal:       incomingCodeNodeCount,
		DateAdded:            time.Now(),
		DateLastUpdated:      time.Now(),
		IoCodeBlocksTotal:    incomingIoCodeNodeCount,
		Languages:            nil,
		LiteralIncludesTotal: incomingLiteralIncludeNodeCount,
		Nodes:                &newCodeNodes,
		PageURL:              pageUrl,
		ProjectName:          projectName,
		Product:              "",
		SubProduct:           "",
	}, projectCounter
}
