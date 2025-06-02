package add_code_examples

import (
	"common"
	"context"
	"fmt"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/prompts"
)

func CategorizeJsonLikeSnippet(contents string, llm *ollama.LLM, ctx context.Context) (string, error) {
	// To tweak the prompt for accuracy, edit this question
	const questionTemplate = `I need to sort code examples into one of these categories:
	%s
	%s
	Use these definitions for each category to help categorize the code example:
	%s: An example object, typically represented in JSON, enumerating fields in a return object and their types. Typically includes an '_id' field and represents one or more example documents. Many pieces of JSON that look similar or repetitive in structure.
	%s: Example configuration object, typically represented in JSON or YAML, enumerating required/optional parameters and their types. If it shows an '_id' field, it is a return object, not a configuration object.
	Using these definitions, which category applies to this code example? Don't list an explanation, only list the category name.`
	question := fmt.Sprintf(questionTemplate,
		common.ExampleReturnObject,
		common.ExampleConfigurationObject,
		common.ExampleReturnObject,
		common.ExampleConfigurationObject,
	)
	template := prompts.NewPromptTemplate(
		`Use the following pieces of context to answer the question at the end.
			Context: {{.contents}}
			Question: {{.question}}`,
		[]string{"contents", "question"},
	)
	prompt, err := template.Format(map[string]any{
		"contents": contents,
		"question": question,
	})
	if err != nil {
		return "", fmt.Errorf("failed to create a prompt from the template: %q\n, %q\n, %q\n, %q\n", template, contents, question, err)
	}
	completion, err := llms.GenerateFromSinglePrompt(ctx, llm, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to generate a response from the CategorizeJsonLikeSnippet prompt (is Ollama running locally?): %w", err)
	}
	return completion, nil
}
