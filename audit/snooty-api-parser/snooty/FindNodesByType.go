package snooty

import "snooty-api-parser/types"

// FindNodesByType recursively finds ASTNodes with a specific type (used to find `code` nodes)
func FindNodesByType(nodes []types.ASTNode, nodeType string) []types.ASTNode {
	var result []types.ASTNode
	for _, node := range nodes {
		if node.Type == nodeType {
			result = append(result, node)
		}
		// Recursively search in the children of the current node
		result = append(result, FindNodesByType(node.Children, nodeType)...)
	}
	return result
}
