package common

import "time"

type CollectionInfoView struct {
	TotalPageCount   int       `bson:"total_page_count" json:"total_page_count"`
	TotalCodeCount   int       `bson:"total_code_count" json:"total_code_count"`
	LastUpdatedAtUTC time.Time `bson:"last_updated_at_utc" json:"last_updated_at_utc"`
}
