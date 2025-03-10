package main

import (
	"log"
	"snooty-api-parser/compare-code-examples"
	"snooty-api-parser/snooty"
	"snooty-api-parser/types"
	"time"
)

func UpdateExistingDocsPage(existingPage types.DocsPage, data types.PageWrapper, projectCounter types.ProjectCounts) (*types.DocsPage, types.ProjectCounts) {
	atlasDocCodeNodeCount := existingPage.CodeNodesTotal
	incomingCodeNodes, incomingLiteralIncludeNodes, incomingIoCodeBlockNodes := snooty.GetCodeExamplesFromIncomingData(data.Data.AST)
	incomingCodeNodePageCount := len(incomingCodeNodes)
	projectCounter = IncrementProjectCountsForExistingPage(len(incomingCodeNodes), len(incomingLiteralIncludeNodes), len(incomingIoCodeBlockNodes), existingPage, projectCounter)
	if incomingCodeNodePageCount == atlasDocCodeNodeCount {
		// The page doesn't have any changes - don't bother returning the page, but do return the updated project counter
		return nil, projectCounter
	}
	updatedDocsPage := existingPage

	// If examples exist already and we are getting no incoming examples from the API, the existing examples have been removed from the incoming page
	if existingPage.Nodes != nil && incomingCodeNodePageCount == 0 {
		removedNodeCount := len(*existingPage.Nodes)
		// Mark all nodes as removed
		updatedCodeNodes := make([]types.CodeNode, 0)
		for _, node := range *existingPage.Nodes {
			node.DateRemoved = time.Now()
			node.IsRemoved = true
			updatedCodeNodes = append(updatedCodeNodes, node)
		}
		updatedDocsPage.Nodes = &updatedCodeNodes
		projectCounter.RemovedCodeNodesCount += removedNodeCount
		log.Printf("Info: removed examples: %d, now: 0, page: %s.\n", removedNodeCount, existingPage.ID)
		// TODO: Set all node and language counts to 0
		updatedDocsPage.DateLastUpdated = time.Now()
	} else if existingPage.Nodes == nil && incomingCodeNodePageCount > 0 {
		// There are no existing code examples - they're all new - so just make new code examples
		newCodeNodes := make([]types.CodeNode, 0)
		for _, snootyNode := range incomingCodeNodes {
			newNode := snooty.MakeCodeNodeFromSnootyAST(snootyNode)
			newCodeNodes = append(newCodeNodes, newNode)
		}
		updatedDocsPage.Nodes = &newCodeNodes
		updatedDocsPage.CodeNodesTotal = len(newCodeNodes)
		updatedDocsPage.LiteralIncludesTotal = len(incomingLiteralIncludeNodes)
		updatedDocsPage.IoCodeBlocksTotal = len(incomingIoCodeBlockNodes)
		updatedDocsPage.DateLastUpdated = time.Now()
		projectCounter.NewCodeNodesCount += len(newCodeNodes)
		log.Printf("Info: new examples: %d, prev: 0, page: %s.\n", len(newCodeNodes), existingPage.ID)
		// TODO: still need to update lang counts and page totals for updated nodes array
	} else if existingPage.Nodes == nil && incomingCodeNodePageCount == 0 {
		// No code examples to deal with here - just return nil and the empty project counter
		return nil, projectCounter
	} else {
		var updatedCodeNodes []types.CodeNode
		//updatedCodeNodes, projectCounter = CompareCodeNodesForPage(*existingPage.Nodes, incomingCodeNodes, projectCounter, existingPage.ID)
		updatedCodeNodes, projectCounter = compare_code_examples.CompareExistingIncomingCodeExampleSlices(*existingPage.Nodes, incomingCodeNodes, projectCounter, existingPage.ID)
		updatedDocsPage.Nodes = &updatedCodeNodes
		updatedDocsPage.DateLastUpdated = time.Now()
		// TODO: still need to update lang counts and page totals for updated nodes array
	}
	// TODO: Figure out how we're going to do updates. Should I make a batch update object here or append it to an array of docs pages or...?
	return &updatedDocsPage, projectCounter
}
