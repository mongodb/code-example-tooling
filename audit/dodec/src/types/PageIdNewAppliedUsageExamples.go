package types

import "common"

type PageIdNewAppliedUsageExamples struct {
	ID                      string            `bson:"_id"`
	NewAppliedUsageExamples []common.CodeNode `bson:"new_applied_usage_examples"`
	Count                   int               `bson:"count"`
}
