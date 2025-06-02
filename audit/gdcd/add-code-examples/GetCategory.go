package add_code_examples

import (
	"common"
	"context"
	"gdcd/add-code-examples/utils"
	"log"

	"github.com/tmc/langchaingo/llms/ollama"
)

func GetCategory(contents string, lang string, llm *ollama.LLM, ctx context.Context, isDriverProject bool) (string, bool) {
	var category string
	var err error

	validCategories := []string{common.ExampleReturnObject, common.ExampleConfigurationObject, common.NonMongoCommand, common.SyntaxExample, common.UsageExample}

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
		category, err = LLMAssignCategory(contents, langCategory, llm, ctx, isDriverProject)
		if err != nil {
			log.Printf("Error categorizing snippet with LLM: %v", err)
			return "Uncategorized", true
		}
		if utils.SliceContainsString(validCategories, category) {
			llmCategorized = true
			return category, llmCategorized
		} else {
			return "Uncategorized", true
		}
	}
}
