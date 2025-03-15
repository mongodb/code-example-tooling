package add_code_examples

import (
	"snooty-api-parser/types"
)

func GetNormalizedLanguageFromASTNode(snootyNode types.ASTNode) string {
	return GetNormalizedLanguageFromString(snootyNode.Lang)
}
