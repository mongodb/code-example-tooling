package snooty

import "snooty-api-parser/types"

func GetCodeExamplesFromIncomingData(incomingData types.AST) ([]types.ASTNode, []types.ASTNode, []types.ASTNode) {
	incomingCodeNodes := FindNodesByType(incomingData.Children, "code")
	incomingLiteralIncludeNodes := FindNodesByName(incomingData.Children, "literalinclude")
	incomingIoCodeBlockNodes := FindNodesByName(incomingData.Children, "io-code-block")
	return incomingCodeNodes, incomingLiteralIncludeNodes, incomingIoCodeBlockNodes
}
