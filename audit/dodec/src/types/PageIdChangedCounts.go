package types

type PageIdChangedCounts struct {
	ID           string `bson:"_id"`
	AddedCount   int    `bson:"added_count"`
	UpdatedCount int    `bson:"updated_count"`
	RemovedCount int    `bson:"removed_count"`
}
