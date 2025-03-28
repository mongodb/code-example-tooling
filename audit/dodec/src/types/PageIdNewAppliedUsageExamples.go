package types

import "common"

type PageIdNewAppliedUsageExamples struct {
	ID                      ProductSubProductDocumentID `bson:"_id"`
	NewAppliedUsageExamples []common.CodeNode           `bson:"new_applied_usage_examples"`
	Count                   int                         `bson:"count"`
}

// ProductSubProductDocumentID represents the structure for the grouped _id field.
type ProductSubProductDocumentID struct {
	Product    string `bson:"product"`
	SubProduct string `bson:"subProduct"`
	DocumentID string `bson:"documentId"`
}
