package add_code_examples

import (
	"context"
	"gdcd/add-code-examples/utils"
	"github.com/tmc/langchaingo/llms/ollama"
)

func GetCategory(contents string, lang string, llm *ollama.LLM, ctx context.Context, isDriverProject bool) (string, bool) {
	var category string
	validCategories := []string{ExampleReturnObject, ExampleConfigurationObject, NonMongoCommand, SyntaxExample, UsageExample}

	/* If the start characters of the code example match a pattern we have defined for a given category,
	 * return the category - no need to get the LLM involved.
	 */
	langCategory := utils.GetLanguageCategory(lang)
	category, stringMatchSuccessful := utils.CheckForStringMatch(contents, langCategory)
	llmCategorized := false
	if stringMatchSuccessful {
		/* The bool we are returning from this func represents whether the LLM categorized the snippet
		 * If we have successfully used string matching to categorize the snippet, the LLM does not process it, so we
		 * return false here
		 */
		return category, llmCategorized
	} else {
		category = LLMAssignCategory(contents, langCategory, llm, ctx, isDriverProject)
		if utils.SliceContainsString(validCategories, category) {
			llmCategorized = true
			return category, llmCategorized
		} else {
			return "Uncategorized", true
		}
	}
}
