package compare_code_examples

import (
	"common"
	"context"
	add_code_examples "gdcd/add-code-examples"
	"gdcd/types"

	"github.com/tmc/langchaingo/llms/ollama"
)

// MakeUpdatedCodeNodesArray takes the slices created in CompareExistingIncomingCodeExampleSlices and creates a new array
// of []common.CodeNode for inserting into Atlas. Because the code examples don't have a unique identifier to perform an
// effective upsert operation, we'll replace the entire array of code examples every time there are updates.
func MakeUpdatedCodeNodesArray(removedCodeNodes []common.CodeNode,
	existingRemovedCodeNodes []common.CodeNode,
	unchangedPageNodes []common.CodeNode,
	updatedPageNodes []types.ASTNode,
	updatedPageNodesSha256CodeNodeLookup map[string]common.CodeNode,
	newPageNodes []types.ASTNode,
	incomingCount int,
	report types.ProjectReport,
	pageId string,
	llm *ollama.LLM,
	ctx context.Context,
	isDriversProject bool) ([]common.CodeNode, types.ProjectReport) {

	// Set up variables used by these functions
	aggregateCodeNodeChanges := make([]common.CodeNode, 0)
	newAppliedUsageExampleCounts := 0

	// Call all the 'Handle' functions in sequence
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
	aggregateCodeNodeChanges = append(aggregateCodeNodeChanges, unchangedPageNodes...)
	aggregateCodeNodeChanges = append(aggregateCodeNodeChanges, updatedCodeNodes...)
	aggregateCodeNodeChanges = append(aggregateCodeNodeChanges, newCodeNodes...)
	aggregateCodeNodeChanges = append(aggregateCodeNodeChanges, removedCodeNodesUpdatedForRemoval...)

	// Increment project counters
	report = UpdateProjectReportForUpdatedCodeNodes(report, pageId, incomingCount, len(unchangedPageNodes), len(updatedCodeNodes), len(newCodeNodes), len(removedCodeNodesUpdatedForRemoval), len(aggregateCodeNodeChanges), newAppliedUsageExampleCounts)
	// We don't want to report on any of the removed code nodes that were already on the page, but we do want to keep them
	// around, so append them to the Nodes array after adding and reporting on the new stuff
	aggregateCodeNodeChanges = append(aggregateCodeNodeChanges, existingRemovedCodeNodes...)
	return aggregateCodeNodeChanges, report
}
