package main

import (
	"snooty-api-parser/types"
	"strings"
	"time"
)

func MakeCodeNodeFromSnootyAST(snootyNode types.ASTNode) types.CodeNode {
	whiteSpaceTrimmedNode := strings.TrimSpace(snootyNode.Value)
	hashString := MakeSha256HashForCode(whiteSpaceTrimmedNode)
	language := GetLanguage(snootyNode)
	fileExtension := GetFileExtension(language)
	category, llmCategorized := GetCategory(whiteSpaceTrimmedNode)
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
