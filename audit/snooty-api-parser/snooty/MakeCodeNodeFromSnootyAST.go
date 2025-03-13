package snooty

import (
	"context"
	"github.com/tmc/langchaingo/llms/ollama"
	"snooty-api-parser/add-code-examples"
	"snooty-api-parser/types"
	"strings"
	"time"
)

func MakeCodeNodeFromSnootyAST(snootyNode types.ASTNode, llm *ollama.LLM, ctx context.Context, isDriverProject bool) types.CodeNode {
	whiteSpaceTrimmedNode := strings.TrimSpace(snootyNode.Value)
	hashString := MakeSha256HashForCode(whiteSpaceTrimmedNode)
	language := add_code_examples.GetNormalizedLanguage(snootyNode)
	fileExtension := add_code_examples.GetFileExtension(snootyNode)
	category, llmCategorized := add_code_examples.GetCategory(whiteSpaceTrimmedNode, language, llm, ctx, isDriverProject)
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
