package main

import (
	"gdcd/add-code-examples"
	"gdcd/snooty"
	"gdcd/types"
)

func MakeLanguagesArray(codeNodes []types.CodeNode, literalIncludeNodes []types.ASTNode, ioCodeBlockNodes []types.ASTNode) types.LanguagesArray {
	languages := make(map[string]types.LanguageCounts)
	canonicalLanguages := add_code_examples.CanonicalLanguages
	for _, language := range canonicalLanguages {
		languages[language] = types.LanguageCounts{}
	}
	for _, node := range codeNodes {
		if languageCounts, exists := languages[node.Language]; exists {
			languageCounts.Total += 1
			languages[node.Language] = languageCounts
		} else {
			countsForLang := languages[add_code_examples.Undefined]
			countsForLang.LiteralIncludes += 1
			languages[add_code_examples.Undefined] = countsForLang
		}
	}
	for _, node := range literalIncludeNodes {
		lang := snooty.GetLangForLiteralInclude(node)
		if languageCounts, exists := languages[lang]; exists {
			languageCounts.LiteralIncludes += 1
			languages[lang] = languageCounts
		} else {
			countsForLang := languages[add_code_examples.Undefined]
			countsForLang.LiteralIncludes += 1
			languages[add_code_examples.Undefined] = countsForLang
		}
	}
	for _, node := range ioCodeBlockNodes {
		lang := snooty.GetLangForIoCodeBlock(node)
		if languageCounts, exists := languages[lang]; exists {
			languageCounts.IOCodeBlock += 1
			languages[lang] = languageCounts
		} else {
			countsForLang := languages[add_code_examples.Undefined]
			countsForLang.IOCodeBlock += 1
			languages[add_code_examples.Undefined] = countsForLang
		}
	}

	// Convert languages map to LanguagesArray
	var languagesArray types.LanguagesArray
	for lang, counts := range languages {
		languagesArray = append(languagesArray, map[string]types.LanguageCounts{lang: counts})
	}

	return languagesArray
}
