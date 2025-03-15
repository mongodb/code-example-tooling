package add_code_examples

import (
	"context"
	"github.com/tmc/langchaingo/llms/ollama"
	"snooty-api-parser/add-code-examples/utils"
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
	}
	return category
}
