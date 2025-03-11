package snooty

import (
	"snooty-api-parser/add-code-examples"
	"snooty-api-parser/types"
	"strings"
	"time"
)

func MakeCodeNodeFromSnootyAST(snootyNode types.ASTNode) types.CodeNode {
	whiteSpaceTrimmedNode := strings.TrimSpace(snootyNode.Value)
	hashString := MakeSha256HashForCode(whiteSpaceTrimmedNode)
	language := add_code_examples.GetLanguage(snootyNode)
	fileExtension := add_code_examples.GetFileExtension(language)
	category, llmCategorized := add_code_examples.GetCategory(whiteSpaceTrimmedNode)
	return types.CodeNode{
		Code:           whiteSpaceTrimmedNode,
		Language:       language,
		FileExtension:  fileExtension,
		Category:       category,
		SHA256Hash:     hashString,
		LLMCategorized: llmCategorized,
		DateAdded:      time.Now(),
	}
}
