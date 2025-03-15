package snooty

import (
	"encoding/json"
	"fmt"
	"os"
	"snooty-api-parser/types"
	"testing"
)

func LoadASTNodeTestDataFromFile(t *testing.T, filename string) []types.ASTNode {
	testFile := fmt.Sprintf("./test-data/%s", filename)
	data, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read test data file: %v", err)
	}
	// Parse the JSON data into a slice of TestData structs
	var snootyPageAST types.PageWrapper
	err = json.Unmarshal(data, &snootyPageAST)
	if err != nil {
		t.Fatalf("Failed to parse test data: %v", err)
	}

	astNodes := make([]types.ASTNode, 0)
	for _, astNode := range snootyPageAST.Data.AST.Children {
		astNodes = append(astNodes, astNode)
	}
	return astNodes
}
