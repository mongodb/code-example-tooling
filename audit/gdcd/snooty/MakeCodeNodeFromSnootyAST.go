package snooty

import (
	"common"
	"context"
	add_code_examples "gdcd/add-code-examples"
	"gdcd/types"
	"strings"
	"time"

	"github.com/tmc/langchaingo/llms/ollama"
)

func MakeCodeNodeFromSnootyAST(snootyNode types.ASTNode, llm *ollama.LLM, ctx context.Context, isDriverProject bool) common.CodeNode {
	whiteSpaceTrimmedCode := strings.TrimSpace(snootyNode.Value)
	hashString := MakeSha256HashForCode(whiteSpaceTrimmedCode)
	language := add_code_examples.GetNormalizedLanguageFromASTNode(snootyNode)
	fileExtension := add_code_examples.GetFileExtensionFromASTNode(snootyNode)
	category, llmCategorized := add_code_examples.GetCategory(whiteSpaceTrimmedCode, language, llm, ctx, isDriverProject)
	return common.CodeNode{
		Code:           whiteSpaceTrimmedCode,
		Language:       language,
		FileExtension:  fileExtension,
		Category:       category,
		SHA256Hash:     hashString,
		LLMCategorized: llmCategorized,
		DateAdded:      time.Now(),
	}
}
