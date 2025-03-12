package compare_code_examples

import (
	"context"
	"github.com/tmc/langchaingo/llms/ollama"
	"snooty-api-parser/types"
)

// MakeUpdatedCodeNodesArray takes the slices created in CompareExistingIncomingCodeExampleSlices and creates a new array
// of []types.CodeNode for inserting into Atlas. Because the code examples don't have a unique identifier to perform an
// effective upsert operation, we'll replace the entire array of code examples every time there are updates.
func MakeUpdatedCodeNodesArray(removedCodeNodes []types.CodeNode,
	unchangedPageNodes []types.ASTNode,
	unchangedPageNodesSha256CodeNodeLookup map[string]types.CodeNode,
	updatedPageNodes []types.ASTNode,
	updatedPageNodesSha256CodeNodeLookup map[string]types.CodeNode,
	newPageNodes []types.ASTNode, existingNodes []types.CodeNode,
	projectCounts types.ProjectCounts,
	pageId string,
	llm *ollama.LLM,
	ctx context.Context,
	isDriversProject bool) ([]types.CodeNode, types.ProjectCounts) {

	// Set up variables used by these functions
	unchangedCount := 0
	updatedCount := 0
	newCount := 0
	removedCount := 0
	aggregateCodeNodeChanges := make([]types.CodeNode, 0)
	existingHashCountMap := make(map[string]int)
	for _, existingNode := range existingNodes {
		existingHashCountMap[existingNode.Code]++
	}
	var unchangedCodeNodes []types.CodeNode

	// Call all the 'Handle' functions in sequence
	unchangedCodeNodes, unchangedCount, newCount, removedCount = HandleUnchangedPageNodes(existingHashCountMap, unchangedPageNodes, unchangedPageNodesSha256CodeNodeLookup, pageId)
	updatedCodeNodes := HandleUpdatedPageNodes(updatedPageNodes, updatedPageNodesSha256CodeNodeLookup)
	newCodeNodes := HandleNewPageNodes(newPageNodes, llm, ctx, isDriversProject)
	removedCodeNodesUpdatedForRemoval := HandleRemovedCodeNodes(removedCodeNodes)

	// Increment project counters
	updatedCount += len(updatedCodeNodes)
	newCount += len(newCodeNodes)
	removedCount += len(removedCodeNodesUpdatedForRemoval)
	projectCounts = IncrementProjectCounterForUpdatedCodeNodes(projectCounts, unchangedCount, updatedCount, newCount, removedCount)

	// Make the updated []types.CodeNode array
	aggregateCodeNodeChanges = append(aggregateCodeNodeChanges, unchangedCodeNodes...)
	aggregateCodeNodeChanges = append(aggregateCodeNodeChanges, updatedCodeNodes...)
	aggregateCodeNodeChanges = append(aggregateCodeNodeChanges, newCodeNodes...)
	aggregateCodeNodeChanges = append(aggregateCodeNodeChanges, removedCodeNodesUpdatedForRemoval...)

	return aggregateCodeNodeChanges, projectCounts
}
