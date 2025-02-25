package types

import (
	"time"
)

type DocsPage struct {
	Languages            LanguageCounts `json:"languages"`
	LiteralIncludesTotal int            `json:"literal_includes_total"`
	IoCodeBlocksTotal    int            `json:"io_code_blocks_total"`
	CodeNodesTotal       int            `json:"code_nodes_total"`
	Nodes                []CodeNode     `json:"nodes"`
	PageURL              string         `json:"page_url"`
	ProjectName          string         `json:"project_name"`
	ID                   string         `json:"_id"`
	DateAdded            time.Time      `json:"date_added"`
	DateLastUpdated      time.Time      `json:"date_last_updated"`
	Product              string         `json:"product"`
	SubProduct           string         `json:"sub_product"`
}
