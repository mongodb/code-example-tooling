package main

import (
	"common"
	"gdcd/snooty"
	"gdcd/types"
)

func MakeLanguagesArray(codeNodes []common.CodeNode, literalIncludeNodes []types.ASTNode, ioCodeBlockNodes []types.ASTNode) common.LanguagesArray {
	languages := make(map[string]common.LanguageCounts)
	canonicalLanguages := common.CanonicalLanguages
	for _, language := range canonicalLanguages {
		languages[language] = common.LanguageCounts{}
	}
	for _, node := range codeNodes {
		if node.IsRemoved {
			// If the node is removed, we don't want to count it in the languages array, so just continue the loop
			continue
		} else {
			if languageCounts, exists := languages[node.Language]; exists {
				languageCounts.Total += 1
				languages[node.Language] = languageCounts
			} else {
				countsForLang := languages[common.Undefined]
				countsForLang.LiteralIncludes += 1
				languages[common.Undefined] = countsForLang
			}
		}
	}
	for _, node := range literalIncludeNodes {
		lang := snooty.GetLangForLiteralInclude(node)
		if languageCounts, exists := languages[lang]; exists {
			languageCounts.LiteralIncludes += 1
			languages[lang] = languageCounts
		} else {
			countsForLang := languages[common.Undefined]
			countsForLang.LiteralIncludes += 1
			languages[common.Undefined] = countsForLang
		}
	}
	for _, node := range ioCodeBlockNodes {
		lang := snooty.GetLangForIoCodeBlock(node)
		if languageCounts, exists := languages[lang]; exists {
			languageCounts.IOCodeBlock += 1
			languages[lang] = languageCounts
		} else {
			countsForLang := languages[common.Undefined]
			countsForLang.IOCodeBlock += 1
			languages[common.Undefined] = countsForLang
		}
	}

	// Convert languages map to LanguagesArray
	var languagesArray common.LanguagesArray
	for lang, counts := range languages {
		languagesArray = append(languagesArray, map[string]common.LanguageCounts{lang: counts})
	}

	return languagesArray
}
