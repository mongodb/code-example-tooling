package snooty

import "snooty-api-parser/types"

// FindNodesByName recursively finds ASTNodes with a specific name (used to find `literalinclude`, `io-code-block`, and `meta` nodes)
func FindNodesByName(nodes []types.ASTNode, name string) []types.ASTNode {
	var result []types.ASTNode
	for _, node := range nodes {
		if node.Name == name {
			result = append(result, node)
		}
		// Recursively search in the children of the current node
		result = append(result, FindNodesByName(node.Children, name)...)
	}
	return result
}
