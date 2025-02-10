package main

type SnippetInfo struct {
	Page           string `json:"page"`
	Category       string `json:"category"`
	Language       string `json:"language"`
	LLMCategorized bool   `json:"llm_categorized"`
}
