package common

import "time"

// CodeNode captures metadata about a specific code example. The `Code` field contains the example itself.
type CodeNode struct {
	Code           string    `bson:"code"`
	Language       string    `bson:"language"`
	FileExtension  string    `bson:"file_extension"`
	Category       string    `bson:"category"`
	SHA256Hash     string    `bson:"sha_256_hash"`
	LLMCategorized bool      `bson:"llm_categorized"`
	DateAdded      time.Time `bson:"date_added"`
	DateUpdated    time.Time `bson:"date_updated,omitempty"`
	DateRemoved    time.Time `bson:"date_removed,omitempty"`
	IsRemoved      bool      `bson:"is_removed,omitempty"`
}
