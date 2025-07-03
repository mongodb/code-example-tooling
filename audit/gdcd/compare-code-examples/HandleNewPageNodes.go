package compare_code_examples

import (
	"common"
	"context"
	"gdcd/snooty"
	"gdcd/types"

	"github.com/tmc/langchaingo/llms/ollama"
)

// HandleNewPageNodes creates a slice of new CodeNode objects from new ASTNode objects, and hands it back to the call site.
// We append all the "Handle" function results to an array, and overwrite the document in the DB with the updated code nodes.
func HandleNewPageNodes(newIncomingPageNodes []types.ASTNodeWrapper, llm *ollama.LLM, ctx context.Context, isDriversProject bool) ([]common.CodeNode, int) {
	newNodes := make([]common.CodeNode, 0)
	newCodeNodeCount := 0
	for _, incomingNode := range newIncomingPageNodes {
		newNode := snooty.MakeCodeNodeFromSnootyAST(incomingNode.Node, llm, ctx, isDriversProject)
		if incomingNode.InstancesOnPage > 1 {
			newNode.InstancesOnPage = incomingNode.InstancesOnPage
			newCodeNodeCount += incomingNode.InstancesOnPage
		} else {
			newCodeNodeCount++
		}
		newNodes = append(newNodes, newNode)
	}
	return newNodes, newCodeNodeCount
}
