package types

// NewAppliedUsageExampleCounterByProductSubProduct aggregates counts and page IDs for pages with new usage examples across collections.
// Used by the aggregations.FindNewAppliedUsageExamples function.
type NewAppliedUsageExampleCounterByProductSubProduct struct {
	ProductSubProductCounts map[string]map[string]int
	ProductAggregateCount   map[string]int
	PagesInCollections      map[string][]PageIdNewAppliedUsageExamples
}
