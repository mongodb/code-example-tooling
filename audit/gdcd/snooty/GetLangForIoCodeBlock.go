package snooty

import (
	"gdcd/add-code-examples"
	"gdcd/types"
	"gdcd/utils"
)

func GetLangForIoCodeBlock(snootyNode types.ASTNode) string {
	var language string
	// An io-code-block has both an input and output, each of which could have its own lang, and they likely have
	// different languages. We're only counting the input to determine the io-code-block language.
	inputNode := FindNodesByName([]types.ASTNode{snootyNode}, "input")
	if inputNode != nil && len(inputNode) > 0 {
		// Unlike other directive types, the language for an input directive is in the options
		maybeOptions := inputNode[0].Options
		if langValue, ok := maybeOptions["language"]; ok {
			// If the "language" key exists, get its value, confirm it's a string, and normalize it
			if langString, isString := langValue.(string); isString {
				language = add_code_examples.GetNormalizedLanguageFromString(langString)
			}
		}
		// If we don't have a valid language yet, we can try checking the code node of the input block for its language
		if language == "" || language == add_code_examples.Undefined {
			codeNode := FindNodesByName(inputNode, "code")
			if codeNode != nil && len(codeNode) > 0 {
				language = add_code_examples.GetNormalizedLanguageFromASTNode(codeNode[0])
			}
		}
		// If we don't have a valid language yet, we can try checking whether the input directive had a filepath, and use
		// that to try to figure out the language
		if language == "" || language == add_code_examples.Undefined {
			filepath := ""
			inputNodeArgs := inputNode[0].Argument
			// If the input node has at least one argument, we can assume that the first argument's value is the filepath
			if inputNodeArgs != nil && len(inputNodeArgs) > 0 {
				filepath = inputNodeArgs[0].Value
			}
			// If the filepath isn't an empty string, try to use it to figure out the language
			language = utils.GetLangFromFilepath(filepath)
		}
	}

	// If we still don't have a language at this point, there's nothing we can do to figure out the language, so return
	// undefined.
	if language == "" {
		language = add_code_examples.Undefined
	}
	return language
}
