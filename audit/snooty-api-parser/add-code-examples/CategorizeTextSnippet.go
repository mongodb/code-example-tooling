package add_code_examples

import (
	"context"
	"fmt"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/prompts"
	"log"
)

func CategorizeTextSnippet(contents string, llm *ollama.LLM, ctx context.Context) string {
	// To tweak the prompt for accuracy, edit this question
	const questionTemplate = `I need to sort code examples into one of these categories:
	%s
	%s
	%s
	%s
	%s
	Use these definitions for each category to help categorize the code example:
	%s: One line or only a few lines of code that demonstrate popular command-line commands, such as 'docker ', 'go run', 'jq ', 'vi ', 'mkdir ', 'npm ', 'cd ' or other common command-line command invocations. If it starts with 'atlas ' it does not belong in this category - it is an Atlas CLI Command. If it starts with 'mongosh ' it does not belong in this category - it is a 'mongosh command'.
	%s: One-line or only a few lines of code that shows the syntax of a command or a method call, but not the initialization of arguments or parameters passed into a command or method call. It demonstrates syntax but is not usable code on its own.	
	%s: Two variants: one is an example object, typically represented in JSON, enumerating fields in the return object and their types. Typically includes an '_id' field and represents one or more example documents. Many pieces of JSON that look similar or repetitive in structure. The second variant looks like text that has been logged to console, such as an error message or status information. May resemble "Backup completed." "Restore completed." or other short status messages.
	%s: Example object, typically represented in JSON or YAML, enumerating required/optional parameters and their types. If it shows an '_id' field, it is a return object, not a configuration object.
	%s: Longer code snippet that establishes parameters, performs basic set up code, and includes the larger context to demonstrate how to accomplish a task. If an example shows parameters but does not show initializing parameters, it is a syntax example, not a usage example.
	Using these definitions, which category applies to this code example? Don't list an explanation, only list the category name.`
	question := fmt.Sprintf(questionTemplate,
		NonMongoCommand,
		SyntaxExample,
		ExampleReturnObject,
		ExampleConfigurationObject,
		UsageExample,
		NonMongoCommand,
		SyntaxExample,
		ExampleReturnObject,
		ExampleConfigurationObject,
		UsageExample,
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
		log.Fatalf("failed to create a prompt from the template: %q\n, %q\n, %q\n, %q\n", template, contents, question, err)
	}
	completion, err := llms.GenerateFromSinglePrompt(ctx, llm, prompt)
	if err != nil {
		log.Fatalf("failed to generate a response from the given prompt: %q", prompt)
	}
	return completion
}
