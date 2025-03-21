package add_code_examples

import "gdcd/types"

func GetFileExtensionFromASTNode(snootyNode types.ASTNode) string {
	return GetFileExtensionFromStringLang(snootyNode.Lang)
}
