package utils

import (
	"dodec/types"
	"strings"
)

// GetFocusAreaAsSubProduct takes info about pages that have new applied usage examples, parses the document `_id` string
// to determine if it contains a substring related to one of the focus areas the docs org cares about, and returns a
// version of the types.PageIdNewAppliedUsageExamples struct with the focus area as the sub-product, even if it's not
// "really" a sub-product. This does not modify the document in the database - only gives us a field to tally results
// after we pull data from Atlas.
func GetFocusAreaAsSubProduct(newAppliedUsageExampleResult types.PageIdNewAppliedUsageExamples) types.PageIdNewAppliedUsageExamples {
	maybeModifiedResult := newAppliedUsageExampleResult
	if strings.Contains(newAppliedUsageExampleResult.ID.DocumentID, "vector-search") {
		maybeModifiedResult.ID.SubProduct = "Vector Search"
	} else if strings.Contains(newAppliedUsageExampleResult.ID.DocumentID, "atlas-search") {
		maybeModifiedResult.ID.SubProduct = "Atlas Search"
	} else if strings.Contains(newAppliedUsageExampleResult.ID.DocumentID, "time-series") || strings.Contains(newAppliedUsageExampleResult.ID.DocumentID, "timeseries") {
		maybeModifiedResult.ID.SubProduct = "Time Series"
	}
	return maybeModifiedResult
}
