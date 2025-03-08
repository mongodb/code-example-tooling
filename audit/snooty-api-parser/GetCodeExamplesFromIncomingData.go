package main

import "snooty-api-parser/types"

func GetCodeExamplesFromIncomingData(incomingData types.AST) ([]types.ASTNode, []types.ASTNode, []types.ASTNode) {
	incomingCodeNodes := findNodesByType(incomingData.Children, "code")
	incomingLiteralIncludeNodes := findNodesByName(incomingData.Children, "literalinclude")
	incomingIoCodeBlockNodes := findNodesByName(incomingData.Children, "io-code-block")
	return incomingCodeNodes, incomingLiteralIncludeNodes, incomingIoCodeBlockNodes
}
