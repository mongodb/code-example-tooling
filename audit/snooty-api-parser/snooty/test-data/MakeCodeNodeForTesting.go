package test_data

import (
	"snooty-api-parser/add-code-examples"
	"snooty-api-parser/snooty"
	"snooty-api-parser/types"
	"time"
)

func MakeCodeNodeForTesting(language string, category string) types.CodeNode {
	code := "Some code goes here"
	fileExtension := add_code_examples.GetFileExtensionFromStringLang(language)
	sha256Hash := snooty.MakeSha256HashForCode(code)
	return types.CodeNode{
		Code:           code,
		Language:       language,
		FileExtension:  fileExtension,
		Category:       category,
		SHA256Hash:     sha256Hash,
		LLMCategorized: false,
		DateAdded:      time.Now(),
		DateUpdated:    time.Time{},
		DateRemoved:    time.Time{},
		IsRemoved:      false,
	}
}
