package add_code_examples

import (
	"context"
	"gdcd/add-code-examples/utils"
	"github.com/tmc/langchaingo/llms/ollama"
	"log"
)

func LLMAssignCategory(contents string, langCategory string, llm *ollama.LLM, ctx context.Context, isDriverProject bool) string {
	var category string
	if langCategory == utils.JSON_LIKE {
		category = CategorizeJsonLikeSnippet(contents, llm, ctx)
	} else if langCategory == utils.DRIVERS_MINUS_JS {
		category = CategorizeDriverLanguageSnippet(contents, llm, ctx)
	} else if langCategory == utils.JavaScript || langCategory == utils.Text {
		if isDriverProject {
			category = CategorizeDriverLanguageSnippet(contents, llm, ctx)
		} else {
			category = CategorizeTextSnippet(contents, llm, ctx)
		}
	} else if langCategory == utils.Shell {
		category = CategorizeShellSnippet(contents, llm, ctx)
	} else if langCategory == Undefined {
		category = CategorizeTextSnippet(contents, llm, ctx)
	} else {
		log.Printf("Lang category is not one of the recognized ones - it's %s", langCategory)
	}
	return category
}
