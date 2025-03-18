package compare_code_examples

import (
	"context"
	"github.com/tmc/langchaingo/llms/ollama"
	add_code_examples "snooty-api-parser/add-code-examples"
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
	newPageNodes []types.ASTNode,
	existingNodes []types.CodeNode,
	incomingCount int,
	report types.ProjectReport,
	pageId string,
	llm *ollama.LLM,
	ctx context.Context,
	isDriversProject bool) ([]types.CodeNode, types.ProjectReport) {

	// Set up variables used by these functions
	existingNodeCount := len(existingNodes)
	aggregateCodeNodeChanges := make([]types.CodeNode, 0)
	existingHashCountMap := make(map[string]int)
	for _, existingNode := range existingNodes {
		existingHashCountMap[existingNode.Code]++
	}
	var unchangedCodeNodes []types.CodeNode
	newAppliedUsageExampleCounts := 0

	// Call all the 'Handle' functions in sequence
	unchangedCodeNodes = HandleUnchangedPageNodes(existingHashCountMap, unchangedPageNodes, unchangedPageNodesSha256CodeNodeLookup)
	updatedCodeNodes := HandleUpdatedPageNodes(updatedPageNodes, updatedPageNodesSha256CodeNodeLookup)
	newCodeNodes := HandleNewPageNodes(newPageNodes, llm, ctx, isDriversProject)
	removedCodeNodesUpdatedForRemoval := HandleRemovedCodeNodes(removedCodeNodes)
	
	if len(newCodeNodes) > 0 {
		for _, node := range newCodeNodes {
			if add_code_examples.IsNewAppliedUsageExample(node) {
				newAppliedUsageExampleCounts++
			}
		}
	}

	// Make the updated []types.CodeNode array
	aggregateCodeNodeChanges = append(aggregateCodeNodeChanges, unchangedCodeNodes...)
	aggregateCodeNodeChanges = append(aggregateCodeNodeChanges, updatedCodeNodes...)
	aggregateCodeNodeChanges = append(aggregateCodeNodeChanges, newCodeNodes...)
	aggregateCodeNodeChanges = append(aggregateCodeNodeChanges, removedCodeNodesUpdatedForRemoval...)

	// Increment project counters
	report = UpdateProjectReportForUpdatedCodeNodes(report, pageId, incomingCount, existingNodeCount, len(unchangedCodeNodes), len(updatedCodeNodes), len(newCodeNodes), len(removedCodeNodesUpdatedForRemoval), len(aggregateCodeNodeChanges), newAppliedUsageExampleCounts)
	return aggregateCodeNodeChanges, report
}
