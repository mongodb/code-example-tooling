package types

// NewAppliedUsageExampleCounter aggregates counts and page IDs for pages with new usage examples across collections.
// Used by the aggregations.FindNewAppliedUsageExamples function.
type NewAppliedUsageExampleCounter struct {
	ProductCounts      map[string]int
	SubProductCounts   map[string]int
	AggregateCount     int
	PagesInCollections map[string][]PageIdNewAppliedUsageExamples
}
