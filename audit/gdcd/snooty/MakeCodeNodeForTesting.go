package snooty

import (
	"common"
	add_code_examples "gdcd/add-code-examples"
	"time"
)

func MakeCodeNodeForTesting(language string, category string) common.CodeNode {
	code := "Some code goes here"
	fileExtension := add_code_examples.GetFileExtensionFromStringLang(language)
	sha256Hash := MakeSha256HashForCode(code)
	return common.CodeNode{
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
