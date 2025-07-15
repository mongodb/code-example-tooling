package add_code_examples

import (
	"common"
	"gdcd/types"
	"log"
	"strings"
)

func GetCategoryFromASTNode(node types.ASTNode) string {
	if node.Category != "" {
		return normalizeCategoryValue(node.Category)
	} else {
		return ""
	}
}

func normalizeCategoryValue(category string) string {
	lowercaseCategory := strings.ToLower(category)

	switch {
	case strings.Contains(lowercaseCategory, "syntax"):
		return common.SyntaxExample
	case strings.Contains(lowercaseCategory, "usage"):
		return common.UsageExample
	case strings.Contains(lowercaseCategory, "return"):
		return common.ExampleReturnObject
	case strings.Contains(lowercaseCategory, "configuration"):
		return common.ExampleConfigurationObject
	case strings.Contains(lowercaseCategory, "command"):
		return common.NonMongoCommand
	default:
		log.Println("ISSUE: Handle the following non-normalizable category value: ", lowercaseCategory)
		return ""
	}
}
