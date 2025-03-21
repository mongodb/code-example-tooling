package main

import "common"

func MakeEmptyLanguagesArray() common.LanguagesArray {
	languages := make(map[string]common.LanguageCounts)
	canonicalLanguages := common.CanonicalLanguages
	for _, language := range canonicalLanguages {
		languages[language] = common.LanguageCounts{
			Total:           0,
			LiteralIncludes: 0,
			IOCodeBlock:     0,
		}
	}
	// Convert languages map to LanguagesArray
	var languagesArray common.LanguagesArray
	for lang, counts := range languages {
		languagesArray = append(languagesArray, map[string]common.LanguageCounts{lang: counts})
	}
	return languagesArray
}
