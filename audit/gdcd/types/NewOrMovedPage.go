package types

import "time"

type NewOrMovedPage struct {
	PageId              string
	CodeNodeCount       int
	LiteralIncludeCount int
	IoCodeBlockCount    int
	PageData            PageMetadata
	OldPageId           string
	NewPageId           string
	DateAdded           time.Time
}
