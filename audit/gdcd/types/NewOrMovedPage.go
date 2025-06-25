package types

import "common"

type NewOrMovedPage struct {
	PageId              string
	CodeNodeCount       int
	LiteralIncludeCount int
	IoCodeBlockCount    int
	PageData            common.DocsPage
	OldPageId           string
	NewPageId           string
}
