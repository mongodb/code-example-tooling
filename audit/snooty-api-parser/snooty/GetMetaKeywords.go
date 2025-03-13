package snooty

import (
	"snooty-api-parser/types"
	"strings"
)

// GetMetaKeywords gets a directive named "meta", looks for a "keywords" key in the options, and returns its value as a
// slice of strings. The "meta" directive may or may not exist on the page, and its options may or may not contain a
// "keywords" key. If it does contain a "keywords" key, the data is a string containing a comma-separated list of
// keywords. We separate them and return each keyword as an individual element in a []string.
func GetMetaKeywords(nodes []types.ASTNode) []string {
	var metaKeywords []string
	incomingMetaNodes := FindNodesByName(nodes, "meta")
	if len(incomingMetaNodes) == 0 {
		// Meta node is not present
		return metaKeywords
	}
	if keywordsValue, exists := incomingMetaNodes[0].Options["keywords"]; exists {
		if keywordsStr, ok := keywordsValue.(string); ok {
			rawKeywords := strings.Split(keywordsStr, ",")
			for _, rawKeyword := range rawKeywords {
				whiteSpaceTrimmedKeyword := strings.TrimSpace(rawKeyword)
				metaKeywords = append(metaKeywords, whiteSpaceTrimmedKeyword)
			}
			return metaKeywords
		} else {
			// Keywords key is present, but it's not a string
			return metaKeywords
		}
	} else {
		// Keywords key is not present
		return metaKeywords
	}
}
