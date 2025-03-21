package add_code_examples

import (
	"gdcd/types"
)

func GetNormalizedLanguageFromASTNode(snootyNode types.ASTNode) string {
	return GetNormalizedLanguageFromString(snootyNode.Lang)
}
