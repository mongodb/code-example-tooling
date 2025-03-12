package compare_code_examples

import (
	"context"
	"github.com/tmc/langchaingo/llms/ollama"
	"snooty-api-parser/snooty"
	"snooty-api-parser/types"
)

// HandleNewPageNodes creates a slice of new CodeNode objects from new ASTNode objects, and hands it back to the call site.
// We append all the "Handle" function results to an array, and overwrite the document in the DB with the updated code nodes.
func HandleNewPageNodes(newIncomingPageNodes []types.ASTNode, llm *ollama.LLM, ctx context.Context, isDriversProject bool) []types.CodeNode {
	newNodes := make([]types.CodeNode, 0)
	for _, incomingNode := range newIncomingPageNodes {
		newNode := snooty.MakeCodeNodeFromSnootyAST(incomingNode, llm, ctx, isDriversProject)
		newNodes = append(newNodes, newNode)
	}
	return newNodes
}
