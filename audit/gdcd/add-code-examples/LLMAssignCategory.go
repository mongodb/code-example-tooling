package add_code_examples

import (
	"common"
	"context"
	"log"

	"github.com/tmc/langchaingo/llms/ollama"
)

func LLMAssignCategory(contents string, langCategory string, llm *ollama.LLM, ctx context.Context, isDriverProject bool) string {
	var category string
	if langCategory == JsonLike {
		category = CategorizeJsonLikeSnippet(contents, llm, ctx)
	} else if langCategory == DriversMinusJs {
		category = CategorizeDriverLanguageSnippet(contents, llm, ctx)
	} else if langCategory == common.JavaScript || langCategory == common.Text {
		if isDriverProject {
			category = CategorizeDriverLanguageSnippet(contents, llm, ctx)
		} else {
			category = CategorizeTextSnippet(contents, llm, ctx)
		}
	} else if langCategory == common.Shell {
		category = CategorizeShellSnippet(contents, llm, ctx)
	} else if langCategory == common.Undefined {
		category = CategorizeTextSnippet(contents, llm, ctx)
	} else {
		log.Printf("Lang category is not one of the recognized ones - it's %s", langCategory)
	}
	return category
}
