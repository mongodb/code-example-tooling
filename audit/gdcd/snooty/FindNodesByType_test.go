package snooty

import (
	"testing"
)

func TestFindNodesByTypeShouldFindCodeNodes(t *testing.T) {
	inputNodes := LoadASTNodeTestDataFromFile(t, "page-with-code-nodes.json")
	codeNodes := FindNodesByType(inputNodes, "code")
	numberOfCodeNodesInTestData := 14
	if len(codeNodes) != numberOfCodeNodesInTestData {
		t.Errorf("FAILED: got %d nodes, want %d", len(codeNodes), numberOfCodeNodesInTestData)
	}
}

func TestFindNodesByTypeShouldFindNoCodeNodes(t *testing.T) {
	inputNodes := LoadASTNodeTestDataFromFile(t, "page-without-code-nodes.json")
	codeNodes := FindNodesByType(inputNodes, "code")
	numberOfCodeNodesInTestData := 0
	if len(codeNodes) != numberOfCodeNodesInTestData {
		t.Errorf("FAILED: got %d nodes, want %d", len(codeNodes), numberOfCodeNodesInTestData)
	}
}
