package common

type CollectionReport struct {
	ID      string                        `bson:"_id" json:"_id"`
	Version map[string]CollectionInfoView `bson:"version" json:"version"`
}
