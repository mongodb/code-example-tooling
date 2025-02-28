package types

import "time"

// Summaries describes the document that lives in every collection, whose `_id` is `summaries`, that tracks metadata about
// the docs project at the time each audit is completed.
type Summaries struct {
	ID      string                 `bson:"_id"`
	Version map[string]VersionInfo `bson:"version"`
}

// VersionInfo tracks relevant details about the state of the docs collection when the audit is completed.
type VersionInfo struct {
	TotalPageCount int       `bson:"total_page_count"`
	TotalCodeCount int       `bson:"total_code_count"`
	LastUpdated    time.Time `bson:"last_updated_at_utc"`
}
