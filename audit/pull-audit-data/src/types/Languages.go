package types

type LanguageCounts struct {
	LiteralIncludes int `bson:"literal_includes" json:"literal_includes"`
	IOCodeBlock     int `bson:"io_code_block" json:"io_code_block"`
	Total           int `bson:"total" json:"total"`
}

type Language map[string]LanguageCounts
