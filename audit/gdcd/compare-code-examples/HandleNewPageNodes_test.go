package compare_code_examples

import (
	"context"
	add_code_examples "gdcd/add-code-examples"
	"gdcd/compare-code-examples/data"
	"gdcd/types"
	"github.com/tmc/langchaingo/llms/ollama"
	"log"
	"testing"
)

// NOTE: For these tests, I'm just confirming that we're creating the correct number of new code nodes. We don't need to
// make any assertions about the values because the code that actually creates the code nodes is tested in the 'add-code-examples' package.
func TestHandleNewPageNodesCreatesOneNode(t *testing.T) {
	astNode := data.GetNewASTNodes(1)
	astNodeWrapper := types.ASTNodeWrapper{
		InstancesOnPage: 1,
		Node:            astNode[0],
	}
	llm, err := ollama.New(ollama.WithModel(add_code_examples.MODEL))
	if err != nil {
		log.Fatalf("failed to connect to ollama: %v", err)
	}
	ctx := context.Background()
	codeNodes, _ := HandleNewPageNodes([]types.ASTNodeWrapper{astNodeWrapper}, llm, ctx, true)
	if len(codeNodes) != 1 {
		t.Errorf("FAILED: Should have 1 new code node")
	}
}

func TestHandleNewPageNodesCreatesMultipleNodes(t *testing.T) {
	astNodes := data.GetNewASTNodes(2)
	astNodeWrapper1 := types.ASTNodeWrapper{
		InstancesOnPage: 1,
		Node:            astNodes[0],
	}
	astNodeWrapper2 := types.ASTNodeWrapper{
		InstancesOnPage: 1,
		Node:            astNodes[1],
	}
	llm, err := ollama.New(ollama.WithModel(add_code_examples.MODEL))
	ctx := context.Background()
	if err != nil {
		log.Fatalf("failed to connect to ollama: %v", err)
	}
	codeNodes, _ := HandleNewPageNodes([]types.ASTNodeWrapper{astNodeWrapper1, astNodeWrapper2}, llm, ctx, true)
	if len(codeNodes) != 2 {
		t.Errorf("FAILED: Should have 2 new code node")
	}
}
