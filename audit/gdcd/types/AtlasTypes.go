package types

import (
	"go.mongodb.org/mongo-driver/v2/bson"
	"time"
)

type DocsPage struct {
	ID                string         `bson:"_id"`
	CodeNodesTotal    int            `bson:"code_nodes_total"`
	DateAdded         time.Time      `bson:"date_added"`
	DateLastUpdated   time.Time      `bson:"date_last_updated"`
	IoCodeBlocksTotal int            `bson:"io_code_blocks_total"`
	Languages         LanguagesArray `bson:"languages"`
	//Languages            map[string]types.LanguageCounts
	LiteralIncludesTotal int         `bson:"literal_includes_total"`
	Nodes                *[]CodeNode `bson:"nodes"`
	PageURL              string      `bson:"page_url"`
	ProjectName          string      `bson:"project_name"`
	Product              string      `bson:"product"`
	SubProduct           string      `bson:"sub_product,omitempty"`
	Keywords             []string    `bson:"keywords,omitempty"`
	DateRemoved          time.Time   `bson:"date_removed,omitempty"`
	IsRemoved            bool        `bson:"is_removed,omitempty"`
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
	DateUpdated    time.Time `bson:"date_updated,omitempty"`
	DateRemoved    time.Time `bson:"date_removed,omitempty"`
	IsRemoved      bool      `bson:"is_removed,omitempty"`
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
		Keywords             []string       `bson:"keywords,omitempty"`
		DateRemoved          time.Time      `bson:"date_removed,omitempty"`
		IsRemoved            bool           `bson:"is_removed,omitempty"`
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
	//d.Languages = aux.Languages.ToMap()
	d.LiteralIncludesTotal = aux.LiteralIncludesTotal
	d.Nodes = aux.Nodes
	d.PageURL = aux.PageURL
	d.ProjectName = aux.ProjectName
	d.Product = aux.Product
	d.SubProduct = aux.SubProduct
	d.Keywords = aux.Keywords
	d.DateRemoved = aux.DateRemoved
	d.IsRemoved = aux.IsRemoved
	return nil
}

type CollectionReport struct {
	ID      string                        `bson:"_id" json:"_id"`
	Version map[string]CollectionInfoView `bson:"version" json:"version"`
}
type CollectionInfoView struct {
	TotalPageCount   int       `bson:"total_page_count" json:"total_page_count"`
	TotalCodeCount   int       `bson:"total_code_count" json:"total_code_count"`
	LastUpdatedAtUTC time.Time `bson:"last_updated_at_utc" json:"last_updated_at_utc"`
}
