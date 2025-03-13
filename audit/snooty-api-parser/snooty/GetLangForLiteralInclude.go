package snooty

import (
	add_code_examples "snooty-api-parser/add-code-examples"
	"snooty-api-parser/types"
	"snooty-api-parser/utils"
)

func GetLangForLiteralInclude(snootyNode types.ASTNode) string {
	// If the literalinclude node has a language value, just use the normalized version of it
	language := add_code_examples.GetNormalizedLanguage(snootyNode)
	// If the language is undefined, try to get it from the filepath
	if language == add_code_examples.Undefined {
		filepath := ""
		nodeArgs := snootyNode.Argument
		// If the literalinclude has at least one argument, we can assume that the first argument's value is the filepath
		if nodeArgs != nil && len(nodeArgs) > 0 {
			filepath = nodeArgs[0].Value
		}
		// If the filepath isn't an empty string, try to use it to figure out the language
		language = utils.GetLangFromFilepath(filepath)
	}
	// If the language is still undefined after trying to get it from the filepath, check for a child code node
	// and try to read its lang
	if language == add_code_examples.Undefined {
		if snootyNode.Children != nil {
			for _, child := range snootyNode.Children {
				if child.Type == "code" {
					if child.Lang != "" {
						language = add_code_examples.GetNormalizedLanguage(child)
					}
				}
			}
		}
	}
	return language
}
