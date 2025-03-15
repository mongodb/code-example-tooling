package snooty

import "testing"

func TestFindNodesByNameShouldFindLiteralIncludeNodes(t *testing.T) {
	inputNodes := LoadASTNodeTestDataFromFile(t, "page-with-code-nodes.json")
	literalIncludeNodes := FindNodesByName(inputNodes, "literalinclude")
	literalIncludeNodeCount := len(literalIncludeNodes)
	numberOfLiteralIncludeNodesInTestData := 1
	if literalIncludeNodeCount != numberOfLiteralIncludeNodesInTestData {
		t.Errorf("FAILED: got %d nodes, want %d", literalIncludeNodeCount, numberOfLiteralIncludeNodesInTestData)
	}
}

func TestFindNodesByTypeShouldFindNoLiteralIncludeNodes(t *testing.T) {
	inputNodes := LoadASTNodeTestDataFromFile(t, "page-without-code-nodes.json")
	literalIncludeNodes := FindNodesByName(inputNodes, "literalinclude")
	literalIncludeNodeCount := len(literalIncludeNodes)
	numberOfLiteralIncludeNodesInTestData := 0
	if literalIncludeNodeCount != numberOfLiteralIncludeNodesInTestData {
		t.Errorf("FAILED: got %d nodes, want %d", literalIncludeNodeCount, numberOfLiteralIncludeNodesInTestData)
	}
}

func TestFindNodesByNameShouldFindIoCodeBlockNodes(t *testing.T) {
	inputNodes := LoadASTNodeTestDataFromFile(t, "page-with-io-code-block-nodes.json")
	ioCodeBlockNodes := FindNodesByName(inputNodes, "io-code-block")
	ioCodeBlockNodeCount := len(ioCodeBlockNodes)
	numberOfIoCodeBlockNodesInTestData := 2
	if ioCodeBlockNodeCount != numberOfIoCodeBlockNodesInTestData {
		t.Errorf("FAILED: got %d nodes, want %d", ioCodeBlockNodeCount, numberOfIoCodeBlockNodesInTestData)
	}
}

func TestFindNodesByNameShouldFindNoIoCodeBlockNodes(t *testing.T) {
	inputNodes := LoadASTNodeTestDataFromFile(t, "page-without-code-nodes.json")
	ioCodeBlockNodes := FindNodesByName(inputNodes, "io-code-block")
	ioCodeBlockNodeCount := len(ioCodeBlockNodes)
	numberOfIoCodeBlockNodesInTestData := 0
	if ioCodeBlockNodeCount != numberOfIoCodeBlockNodesInTestData {
		t.Errorf("FAILED: got %d nodes, want %d", ioCodeBlockNodeCount, numberOfIoCodeBlockNodesInTestData)
	}
}
