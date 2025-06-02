package add_code_examples

import (
	"common"
	"context"
	"fmt"
	"log"

	"github.com/tmc/langchaingo/llms/ollama"
)

func LLMAssignCategory(contents string, langCategory string, llm *ollama.LLM, ctx context.Context, isDriverProject bool) (string, error) {
	var category string
	var err error

	if langCategory == JsonLike {
		category, err = CategorizeJsonLikeSnippet(contents, llm, ctx)
	} else if langCategory == DriversMinusJs {
		category, err = CategorizeDriverLanguageSnippet(contents, llm, ctx)
	} else if langCategory == common.JavaScript || langCategory == common.Text {
		if isDriverProject {
			category, err = CategorizeDriverLanguageSnippet(contents, llm, ctx)
		} else {
			category, err = CategorizeTextSnippet(contents, llm, ctx)
		}
	} else if langCategory == common.Shell {
		category, err = CategorizeShellSnippet(contents, llm, ctx)
	} else if langCategory == common.Undefined {
		category, err = CategorizeTextSnippet(contents, llm, ctx)
	} else {
		log.Printf("Lang category is not one of the recognized ones - it's %s", langCategory)
		return "", fmt.Errorf("unrecognized language category: %s", langCategory)
	}

	if err != nil {
		return "", fmt.Errorf("failed to categorize snippet: %w", err)
	}
	return category, nil
}
