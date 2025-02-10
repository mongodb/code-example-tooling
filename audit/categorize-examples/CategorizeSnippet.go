package main

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/prompts"
)

func HasStringMatchPrefix(contents string, langCategory string) (string, bool) {
	// These prefixes related to syntax examples
	atlasCli := "atlas "
	mongosh := "mongosh "

	// These prefixes relate to usage examples
	importPrefix := "import "
	fromPrefix := "from "
	namespacePrefix := "namespace "
	packagePrefix := "package "
	usingPrefix := "using "
	mongoConnectionStringPrefix := "mongodb://"
	alternoConnectionStringPrefix := "mongodb+srv://"

	// These prefixes relate to command-line commands that *aren't* MongoDB specific, such as other tools, package managers, etc.
	mkdir := "mkdir "
	cd := "cd "
	docker := "docker "
	dockerCompose := "docker-compose "
	brew := "brew "
	yum := "yum "
	apt := "apt-"
	npm := "npm "
	pip := "pip "
	goRun := "go run "
	node := "node "
	dotnet := "dotnet "
	export := "export "
	sudo := "sudo "
	copyPrefix := "cp "
	tar := "tar "
	jq := "jq "
	vi := "vi "
	cmake := "cmake "
	syft := "syft "
	choco := "choco "

	syntaxExamplePrefixes := []string{atlasCli, mongosh}
	usageExamplePrefixes := []string{importPrefix, fromPrefix, namespacePrefix, packagePrefix, usingPrefix, mongoConnectionStringPrefix, alternoConnectionStringPrefix}
	nonMongoPrefixes := []string{mkdir, cd, docker, dockerCompose, dockerCompose, brew, yum, apt, npm, pip, goRun, node, dotnet, export, sudo, copyPrefix, tar, jq, vi, cmake, syft, choco}

	if langCategory == SHELL {
		for _, prefix := range syntaxExamplePrefixes {
			if strings.HasPrefix(contents, prefix) {
				return SyntaxExample, true
			}
		}
		for _, prefix := range nonMongoPrefixes {
			if strings.HasPrefix(contents, prefix) {
				return NonMongoCommand, true
			}
		}
		return "Uncategorized", false
	} else if langCategory == TEXT {
		for _, prefix := range syntaxExamplePrefixes {
			if strings.HasPrefix(contents, prefix) {
				return SyntaxExample, true
			}
		}
		for _, prefix := range nonMongoPrefixes {
			if strings.HasPrefix(contents, prefix) {
				return NonMongoCommand, true
			}
		}
		for _, prefix := range usageExamplePrefixes {
			if strings.HasPrefix(contents, prefix) {
				return UsageExample, true
			}
		}
		return "Uncategorized", false
	} else {
		for _, prefix := range usageExamplePrefixes {
			if strings.HasPrefix(contents, prefix) {
				return UsageExample, true
			}
		}
		return "Uncategorized", false
	}
}

func ContainsString(contents string) (string, bool) {
	// These strings are typically included in usage examples
	aggregationExample := ".aggregate"
	mongoConnectionStringPrefix := "mongodb://"
	alternoConnectionStringPrefix := "mongodb+srv://"

	// These strings are typically included in return objects
	warningString := "warning"
	deprecatedString := "deprecated"
	idString := "_id"

	// These strings are typically included in non-MongoDB commands
	cmake := "cmake "

	// Some of the examples can be quite long. For the current case, we only care if `.aggregate` appears near the beginning of the example
	substringLengthToCheck := 50
	usageExampleSubstringsToEvaluate := []string{aggregationExample, mongoConnectionStringPrefix, alternoConnectionStringPrefix}
	returnObjectStringsToEvaluate := []string{warningString, deprecatedString, idString}
	nonMongoDBStringsToEvaluate := []string{cmake}

	if substringLengthToCheck < len(contents) {
		substring := contents[:substringLengthToCheck]
		for _, exampleString := range usageExampleSubstringsToEvaluate {
			if strings.Contains(substring, exampleString) {
				return UsageExample, true
			}
		}
		for _, exampleString := range returnObjectStringsToEvaluate {
			if strings.Contains(substring, exampleString) {
				return ExampleReturnObject, true
			}
		}
		for _, exampleString := range nonMongoDBStringsToEvaluate {
			if strings.Contains(substring, exampleString) {
				return NonMongoCommand, true
			}
		}
	} else {
		for _, exampleString := range usageExampleSubstringsToEvaluate {
			if strings.Contains(contents, exampleString) {
				return UsageExample, true
			}
		}
		for _, exampleString := range returnObjectStringsToEvaluate {
			if strings.Contains(contents, exampleString) {
				return ExampleReturnObject, true
			}
		}
		for _, exampleString := range nonMongoDBStringsToEvaluate {
			if strings.Contains(contents, exampleString) {
				return NonMongoCommand, true
			}
		}
	}

	/* 	This Regexp checks for '$' followed by 2 or more characters, followed by a colon
	i.e. '$gte:' or '$project:'
	AND the capture group (the part in parentheses) checks for a pair of angle brackets, which may span
	multiple lines. If the regexp matches, it's an aggregation example. If it contains one or more capture
	groups, it's an aggregation example containing something like '<placeholder>'. According to our definitions,
	we would consider an agg example with placeholders a "syntax example" - not a "usage example" - so
	the number of matches determines whether the example contains placeholders and is a syntax example. A single match
	means it does not have any capture groups and therefore does not contain placeholders, so it's a usage example.
	More than one match means it has one or more capture groups in addition to the single match, and that makes it
	a syntax example.
	*/
	aggPipeline := `(?s)\$[a-zA-Z]{2,}: ?(.*?<.+?>)?`
	re, err := regexp.Compile(aggPipeline)
	if err != nil {
		log.Fatal("Error compiling the regexp for the agg pipeline: ", err)
	}
	regExpMatches := re.FindStringSubmatch(contents)
	matchLength := len(regExpMatches)
	if matchLength > 1 {
		if regExpMatches[1] != "" {
			return SyntaxExample, true
		} else {
			return UsageExample, true
		}
	} else {
		return "Uncategorized", false
	}
}

// CheckForStringMatch The bool we return from this func represents whether the string matching was successful.
// If the string match was successful, we don't need to move on to LLM matching.
func CheckForStringMatch(contents string, langCategory string) (string, bool) {
	// Prefix matching should be fastest as it only has to search the first N characters of a string to determine whether it's
	// a match. So first, try to match prefixes.
	category, hasPrefix := HasStringMatchPrefix(contents, langCategory)
	if hasPrefix {
		return category, hasPrefix
	} else {
		// If the prefix matching doesn't work, try the slower string matching.
		thisCategory, containsExampleString := ContainsString(contents)
		if containsExampleString {
			return thisCategory, containsExampleString
		} else {
			return "Uncategorized", false
		}
	}
}

func ProcessSnippet(contents string, lang string, llm *ollama.LLM, ctx context.Context, isDriverProject bool) (string, bool) {
	var category string
	validCategories := []string{ExampleReturnObject, ExampleConfigurationObject, NonMongoCommand, SyntaxExample, UsageExample}

	/* If the start characters of the code example match a pattern we have defined for a given category,
	 * return the category - no need to get the LLM involved.
	 */
	langCategory := GetLanguageCategory(lang)
	category, stringMatchSuccessful := CheckForStringMatch(contents, langCategory)
	if stringMatchSuccessful {
		/* The bool we are returning from this func represents whether the LLM categorized the snippet
		 * If we have successfully used string matching to categorize the snippet, the LLM does not process it, so we
		 * return false here
		 */
		return category, false
	} else {
		category = LLMAssignCategory(contents, langCategory, llm, ctx, isDriverProject)

		if containsString(validCategories, category) {
			return category, true
		} else {
			return "Uncategorized", true
		}
	}
}

func GetLanguageCategory(lang string) string {
	jsonLike := []string{JSON, XML, YAML}
	driversLanguagesMinusJS := []string{C, CPP, CSHARP, GO, JAVA, KOTLIN, PHP, PYTHON, RUBY, RUST, SCALA, SWIFT, TYPESCRIPT}
	if containsString([]string{SHELL}, lang) {
		return SHELL
	} else if containsString(jsonLike, lang) {
		return JSON_LIKE
	} else if containsString(driversLanguagesMinusJS, lang) {
		return DRIVERS_MINUS_JS
	} else if lang == JAVASCRIPT {
		return JAVASCRIPT
	} else if lang == TEXT {
		return TEXT
	} else {
		return "Unknown language"
	}
}

func LLMAssignCategory(contents string, langCategory string, llm *ollama.LLM, ctx context.Context, isDriverProject bool) string {
	var category string
	if langCategory == JSON_LIKE {
		category = CategorizeJsonLikeSnippet(contents, llm, ctx)
	} else if langCategory == DRIVERS_MINUS_JS {
		category = CategorizeDriverLanguageSnippet(contents, llm, ctx)
	} else if langCategory == JAVASCRIPT || langCategory == TEXT {
		if isDriverProject {
			category = CategorizeDriverLanguageSnippet(contents, llm, ctx)
		} else {
			category = CategorizeTextSnippet(contents, llm, ctx)
		}
	} else if langCategory == SHELL {
		category = CategorizeShellSnippet(contents, llm, ctx)
	}
	return category
}

func CategorizeJsonLikeSnippet(contents string, llm *ollama.LLM, ctx context.Context) string {
	// To tweak the prompt for accuracy, edit this question
	const questionTemplate = `I need to sort code examples into one of these categories:
	%s
	%s
	Use these definitions for each category to help categorize the code example:
	%s: An example object, typically represented in JSON, enumerating fields in a return object and their types. Typically includes an '_id' field and represents one or more example documents. Many pieces of JSON that look similar or repetitive in structure.
	%s: Example configuration object, typically represented in JSON or YAML, enumerating required/optional parameters and their types. If it shows an '_id' field, it is a return object, not a configuration object.
	Using these definitions, which category applies to this code example? Don't list an explanation, only list the category name.`
	question := fmt.Sprintf(questionTemplate,
		ExampleReturnObject,
		ExampleConfigurationObject,
		ExampleReturnObject,
		ExampleConfigurationObject,
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

func CategorizeShellSnippet(contents string, llm *ollama.LLM, ctx context.Context) string {
	// To tweak the prompt for accuracy, edit this question
	const questionTemplate = `I need to sort code examples into one of these categories:
	%s
	%s
	%s
	%s
	Use these definitions for each category to help categorize the code example:
	%s: One line or only a few lines of code that demonstrate popular command-line commands, such as 'docker ', 'go run', 'jq ', 'vi ', 'mkdir ', 'npm ', 'cd ' or other common command-line command invocations. If it starts with 'atlas ' it does not belong in this category - it is an Atlas CLI Command. If it starts with 'mongosh ' it does not belong in this category - it is a 'mongosh command'.
	%s: One-line or only a few lines of code that shows the syntax of a command or a method call, but not the initialization of arguments or parameters passed into a command or method call. It demonstrates syntax but is not usable code on its own.
	%s: Two variants: one is an example object, typically represented in JSON, enumerating fields in the return object and their types. Typically includes an '_id' field and represents one or more example documents. Many pieces of JSON that look similar or repetitive in structure. The second variant looks like text that has been logged to console, such as an error message or status information. May resemble "Backup completed." "Restore completed." or other short status messages.
	%s: Example object, typically represented in JSON or YAML, enumerating required/optional parameters and their types. If it shows an '_id' field, it is a return object, not a configuration object.
	Using these definitions, which category applies to this code example? Don't list an explanation, only list the category name.`
	question := fmt.Sprintf(questionTemplate,
		NonMongoCommand,
		SyntaxExample,
		ExampleReturnObject,
		ExampleConfigurationObject,
		NonMongoCommand,
		SyntaxExample,
		ExampleReturnObject,
		ExampleConfigurationObject,
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

func CategorizeDriverLanguageSnippet(contents string, llm *ollama.LLM, ctx context.Context) string {
	// To tweak the prompt for accuracy, edit this question
	const questionTemplate = `I need to sort code examples into one of these categories:
		%s
		%s
		Use these definitions for each category to help categorize the code example:
		%s: One-line or only a few lines of code that shows the syntax of a command or a method call, but not the initialization of arguments or parameters passed into a command or method call. It demonstrates syntax but is not usable code on its own.
		%s: Longer code snippet that establishes parameters, performs basic set up code, and includes the larger context to demonstrate how to accomplish a task. If an example shows parameters but does not show initializing parameters, it is a syntax example, not a usage example.
		Using these definitions, which category applies to this code example? Don't list an explanation, only list the category name.`
	question := fmt.Sprintf(questionTemplate,
		SyntaxExample,
		UsageExample,
		SyntaxExample,
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
