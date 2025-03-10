package main

import "snooty-api-parser/types"
import "snooty-api-parser/snooty"

func GetCodeExamplesFromIncomingData(incomingData types.AST) ([]types.ASTNode, []types.ASTNode, []types.ASTNode) {
	incomingCodeNodes := snooty.FindNodesByType(incomingData.Children, "code")
	incomingLiteralIncludeNodes := snooty.FindNodesByName(incomingData.Children, "literalinclude")
	incomingIoCodeBlockNodes := snooty.FindNodesByName(incomingData.Children, "io-code-block")
	return incomingCodeNodes, incomingLiteralIncludeNodes, incomingIoCodeBlockNodes
}
