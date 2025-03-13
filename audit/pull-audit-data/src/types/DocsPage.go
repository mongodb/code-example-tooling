package types

import (
	"go.mongodb.org/mongo-driver/v2/bson"
	"time"
)

// A DocsPage is the primary document structure in all collections in the database. It contains metadata about the page,
// and the `Nodes` array holds details about the code examples on a docs page. Every page in a docs project has a DocsPage,
// even if it has no code examples. On pages where there are no code examples, the `Nodes` array value is null.
type DocsPage struct {
	ID                   string    `bson:"_id"`
	CodeNodesTotal       int       `bson:"code_nodes_total"`
	DateAdded            time.Time `bson:"date_added"`
	DateLastUpdated      time.Time `bson:"date_last_updated"`
	IoCodeBlocksTotal    int       `bson:"io_code_blocks_total"`
	Languages            map[string]LanguageCounts
	LiteralIncludesTotal int         `bson:"literal_includes_total"`
	Nodes                *[]CodeNode `bson:"nodes"`
	PageURL              string      `bson:"page_url"`
	ProjectName          string      `bson:"project_name"`
	Product              string      `bson:"product"`
	SubProduct           string      `bson:"sub_product,omitempty"`
}

// CodeNode captures metadata about a specific code example. The `Code` field contains the example itself.
type CodeNode struct {
	Code           string    `bson:"code"`
	Language       string    `bson:"language"`
	FileExtension  string    `bson:"file_extension"`
	Category       string    `bson:"category"`
	SHA256Hash     string    `bson:"sha_256_hash"`
	LLMCategorized bool      `bson:"llm_categorized"`
	DateAdded      time.Time `bson:"date_added"`
}

// LanguageCounts captures the counts of literalincludes and io-code-block directive instances for a specific language
// on a DocsPage. The `Total` count is a sum of both directive types, plus `code-block` and `code` directives on the page.
type LanguageCounts struct {
	LiteralIncludes int `bson:"literal_includes" json:"literal_includes"`
	IOCodeBlock     int `bson:"io_code_block" json:"io_code_block"`
	Total           int `bson:"total" json:"total"`
}

// LanguagesArray is a custom type to handle unmarshalling languages.
type LanguagesArray []map[string]LanguageCounts

// ToMap converts the LanguagesArray to a map[string]LanguageMetrics.
func (languages LanguagesArray) ToMap() map[string]LanguageCounts {
	result := make(map[string]LanguageCounts)
	for _, languageEntry := range languages {
		for lang, metrics := range languageEntry {
			result[lang] = metrics
		}
	}
	return result
}

// UnmarshalBSON handles the custom unmarshalling of the languages field.
func (d *DocsPage) UnmarshalBSON(data []byte) error {
	aux := struct {
		ID                   string         `bson:"_id"`
		CodeNodesTotal       int            `bson:"code_nodes_total"`
		DateAdded            time.Time      `bson:"date_added"`
		DateLastUpdated      time.Time      `bson:"date_last_updated"`
		IoCodeBlocksTotal    int            `bson:"io_code_blocks_total"`
		Languages            LanguagesArray `bson:"languages"`
		LiteralIncludesTotal int            `bson:"literal_includes_total"`
		Nodes                *[]CodeNode    `bson:"nodes"`
		PageURL              string         `bson:"page_url"`
		ProjectName          string         `bson:"project_name"`
		Product              string         `bson:"product"`
		SubProduct           string         `bson:"sub_product,omitempty"`
	}{}
	if err := bson.Unmarshal(data, &aux); err != nil {
		return err
	}
	// Copy fields
	d.ID = aux.ID
	d.CodeNodesTotal = aux.CodeNodesTotal
	d.DateAdded = aux.DateAdded
	d.DateLastUpdated = aux.DateLastUpdated
	d.IoCodeBlocksTotal = aux.IoCodeBlocksTotal
	d.Languages = aux.Languages.ToMap()
	d.LiteralIncludesTotal = aux.LiteralIncludesTotal
	d.Nodes = aux.Nodes
	d.PageURL = aux.PageURL
	d.ProjectName = aux.ProjectName
	d.Product = aux.Product
	d.SubProduct = aux.SubProduct
	return nil
}
