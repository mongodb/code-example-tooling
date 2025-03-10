package snooty

import (
	"snooty-api-parser/types"
	"snooty-api-parser/utils"
	"strings"
	"time"
)

func MakeCodeNodeFromSnootyAST(snootyNode types.ASTNode) types.CodeNode {
	whiteSpaceTrimmedNode := strings.TrimSpace(snootyNode.Value)
	hashString := MakeSha256HashForCode(whiteSpaceTrimmedNode)
	language := utils.GetLanguage(snootyNode)
	fileExtension := utils.GetFileExtension(language)
	category, llmCategorized := utils.GetCategory(whiteSpaceTrimmedNode)
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
