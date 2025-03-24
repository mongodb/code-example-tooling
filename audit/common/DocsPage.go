package common

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type DocsPage struct {
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
	d.Languages = aux.Languages
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
