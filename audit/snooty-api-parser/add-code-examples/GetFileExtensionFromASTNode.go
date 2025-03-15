package add_code_examples

import "snooty-api-parser/types"

func GetFileExtensionFromASTNode(snootyNode types.ASTNode) string {
	return GetFileExtensionFromStringLang(snootyNode.Lang)
}
