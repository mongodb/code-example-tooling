package compare_code_examples

import "snooty-api-parser/types"

func MakeUpdatedCodeNodesArray(removedCodeNodes []types.CodeNode, unchangedPageNodes []types.ASTNode, unchangedPageNodesSha256CodeNodeLookup map[string]types.CodeNode, updatedPageNodes []types.ASTNode, updatedPageNodesSha256CodeNodeLookup map[string]types.CodeNode, newPageNodes []types.ASTNode, existingNodes []types.CodeNode, projectCounts types.ProjectCounts, pageId string) ([]types.CodeNode, types.ProjectCounts) {
	aggregateCodeNodeChanges := make([]types.CodeNode, 0)
	existingHashCountMap := make(map[string]int)
	for _, existingNode := range existingNodes {
		existingHashCountMap[existingNode.Code]++
	}

	// Call all the 'Handle' functions in sequence
	unchangedCodeNodes, unchangedCount, newCount, removedCount := HandleUnchangedPageNodes(existingHashCountMap, unchangedPageNodes, unchangedPageNodesSha256CodeNodeLookup, pageId)
	projectCounts.ExistingCodeNodesCount += unchangedCount
	projectCounts.NewCodeNodesCount += newCount
	projectCounts.RemovedCodeNodesCount += removedCount

	updatedCodeNodes := HandleUpdatedPageNodes(updatedPageNodes, updatedPageNodesSha256CodeNodeLookup)
	projectCounts.UpdatedCodeNodesCount += len(updatedCodeNodes)
	newCodeNodes := HandleNewPageNodes(newPageNodes)
	projectCounts.NewCodeNodesCount += len(newCodeNodes)
	removedCodeNodesUpdatedForRemoval := HandleRemovedCodeNodes(removedCodeNodes)
	projectCounts.RemovedCodeNodesCount += len(removedCodeNodesUpdatedForRemoval)

	aggregateCodeNodeChanges = append(aggregateCodeNodeChanges, unchangedCodeNodes...)
	aggregateCodeNodeChanges = append(aggregateCodeNodeChanges, updatedCodeNodes...)
	aggregateCodeNodeChanges = append(aggregateCodeNodeChanges, newCodeNodes...)
	aggregateCodeNodeChanges = append(aggregateCodeNodeChanges, removedCodeNodesUpdatedForRemoval...)
	return aggregateCodeNodeChanges, projectCounts
}
