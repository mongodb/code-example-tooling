package main

import (
	"snooty-api-parser/types"
	"snooty-api-parser/utils"
	"time"
)

func RemoveExistingPage(page *types.DocsPage, report types.ProjectReport) (*types.DocsPage, types.ProjectReport) {
	page.IsRemoved = true
	page.DateRemoved = time.Now()
	nodeIterationCount := 0
	var removedNodeCount int
	if page.Nodes != nil {
		removedNodeCount = len(*page.Nodes)
		updatedCodeNodes := make([]types.CodeNode, 0)
		// Mark the code example nodes on the page as removed
		for _, node := range *page.Nodes {
			node.DateRemoved = time.Now()
			node.IsRemoved = true
			updatedCodeNodes = append(updatedCodeNodes, node)
			nodeIterationCount++
		}
		page.Nodes = &updatedCodeNodes

		// Set total nodes, literalincludes and io-code-blocks to 0
		page.CodeNodesTotal = 0
		page.LiteralIncludesTotal = 0
		page.IoCodeBlocksTotal = 0

		// Update the languages array to zero out counts for languages the page
		updatedLanguagesArray := MakeLanguagesArray([]types.CodeNode{}, []types.ASTNode{}, []types.ASTNode{})
		page.Languages = updatedLanguagesArray

		// Update the date_last_updated time
		page.DateLastUpdated = time.Now()
	}

	// Update report for removed page
	report.Counter.RemovedPagesCount += 1
	report.Counter.RemovedCodeNodesCount += nodeIterationCount
	report = utils.ReportChanges(types.PageRemoved, report, page.ID)
	if nodeIterationCount != removedNodeCount {
		report = utils.ReportIssues(types.CodeNodeCountIssue, report, page.ID, nodeIterationCount, removedNodeCount)
	}
	return page, report
}
