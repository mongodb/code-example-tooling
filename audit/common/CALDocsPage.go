package common

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type CalDocsPage struct {
	ID                   bson.ObjectID  `bson:"_id"`              // ObjectID instead of string
	CodeNodesTotal       int            `bson:"code_nodes_total"` // Keep other fields same as `DocsPage`
	DateAdded            time.Time      `bson:"date_added"`
	DateLastUpdated      time.Time      `bson:"date_last_updated"`
	IoCodeBlocksTotal    int            `bson:"io_code_blocks_total"`
	Languages            LanguagesArray `bson:"languages"`
	LanguagesFacet       []string       `bson:"languages_facet,omitempty"`
	CategoriesFacet      []string       `bson:"categories_facet,omitempty"`
	LiteralIncludesTotal int            `bson:"literal_includes_total"`
	Nodes                *[]CodeNode    `bson:"nodes"`
	PageURL              string         `bson:"page_url"`
	ProjectName          string         `bson:"project_name"`
	Product              string         `bson:"product"`
	SubProduct           string         `bson:"sub_product,omitempty"`
	Keywords             []string       `bson:"keywords,omitempty"`
	DateRemoved          time.Time      `bson:"date_removed,omitempty"`
	IsRemoved            bool           `bson:"is_removed,omitempty"`
}
