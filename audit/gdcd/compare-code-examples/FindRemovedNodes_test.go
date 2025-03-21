package compare_code_examples

import (
	"common"
	"gdcd/compare-code-examples/data"
	"gdcd/types"
	"testing"
)

func TestFindRemovedNodesShouldFindRemovedNodes(t *testing.T) {
	codeNode, astNode := data.GetRemovedNodes()
	existingNodeHashMap := make(map[string]common.CodeNode)
	existingNodeHashMap[codeNode.SHA256Hash] = codeNode
	removedNodes := FindRemovedNodes(existingNodeHashMap, []types.ASTNode{astNode}, []types.ASTNode{}, []types.ASTNode{})
	removedNodeCount := len(removedNodes)
	expectedRemovedNodeCount := 1
	if removedNodeCount != expectedRemovedNodeCount {
		t.Errorf("FAILED: got %d removed nodes, want %d", removedNodeCount, expectedRemovedNodeCount)
	}
}

func TestFindRemovedNodesShouldFindNoRemovedNodes(t *testing.T) {
	codeNode, astNode := data.GetUnchangedNodes()
	existingNodeHashMap := make(map[string]common.CodeNode)
	existingNodeHashMap[codeNode.SHA256Hash] = codeNode
	removedNodes := FindRemovedNodes(existingNodeHashMap, []types.ASTNode{}, []types.ASTNode{astNode}, []types.ASTNode{})
	removedNodeCount := len(removedNodes)
	expectedRemovedNodeCount := 0
	if removedNodeCount != expectedRemovedNodeCount {
		t.Errorf("FAILED: got %d removed nodes, want %d", removedNodeCount, expectedRemovedNodeCount)
	}
}
