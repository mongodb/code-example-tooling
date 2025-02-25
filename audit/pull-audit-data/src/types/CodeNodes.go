package types

import "time"

type CodeNode struct {
	Code           string    `bson:"code,omitempty" json:"code"`
	Language       string    `bson:"language,omitempty" json:"language"`
	FileExtension  string    `bson:"file_extension,omitempty" json:"file_extension"`
	Category       string    `bson:"category,omitempty" json:"category"`
	SHA256Hash     string    `bson:"sha_256_hash,omitempty" json:"sha_256_hash"`
	LLMCategorized bool      `bson:"llm_categorized" bson:"llm_categorized"`
	DateAdded      time.Time `bson:"date_added,omitempty" bson:"date_added"`
}
