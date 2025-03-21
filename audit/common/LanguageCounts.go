package common

// LanguageCounts captures the counts of literalincludes and io-code-block directive instances for a specific language
// on a DocsPage. The `Total` count is a sum of both directive types, plus `code-block` and `code` directives on the page.
type LanguageCounts struct {
	LiteralIncludes int `bson:"literal_includes" json:"literal_includes"`
	IOCodeBlock     int `bson:"io_code_block" json:"io_code_block"`
	Total           int `bson:"total" json:"total"`
}
