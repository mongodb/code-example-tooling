package utils

import (
	"dodec/types"
	"fmt"
)

// PrintPageIdNewAppliedUsageExampleCounts prints a nicely-formatted series of tables with counts of new applied usage
// examples broken down in various ways. Each collection prints as its own table, which has a list of page IDs and counts
// for pages that have new usage examples. This makes it easy to validate the data if we want to perform manual QA. After
// the individual collection tables, we print tables breaking down the counts by sub-product and product.
func PrintPageIdNewAppliedUsageExampleCounts(appliedUsageExampleCounter types.NewAppliedUsageExampleCounter) {
	columnNames := []interface{}{"Product", "Sub Product", "Page ID", "Count"}
	// Print a separate table for each top-level element
	columnWidths := []int{30, 30, 70, 15}
	for collectionName, pagesToPrintInCollection := range appliedUsageExampleCounter.PagesInCollections {
		collectionCount := 0
		fmt.Printf("\nNew Applied Usage Example Counts by Page in Collection %s\n", collectionName)
		printSeparator(columnWidths...)
		printRow(columnWidths, columnNames...)
		printSeparator(columnWidths...)
		// This type also stores the code nodes directly - do we want to print any details about the specific nodes that
		// match our conditions?
		for _, page := range pagesToPrintInCollection {
			printRow(columnWidths, page.ID.Product, page.ID.SubProduct, page.ID.DocumentID, page.Count)
			collectionCount += page.Count
		}
		printSeparator(columnWidths...)
		fmt.Printf("\nTotal new applied usage example counts in %s: %d\n", collectionName, collectionCount)
	}
	fmt.Printf("\nTotal New Applied Usage Example Counts in Last Week: %d\n", appliedUsageExampleCounter.AggregateCount)

	// This prints a table showing counts of new applied usage examples broken down by sub-product. To simplify reporting
	// up the docs chain, the function that performs the aggregation, aggregations.FindNewAppliedUsageExamples, assigns
	// an arbitrary sub-product related to the Page ID for key focus areas. This print function does not distinguish
	// between "real" sub-products and sub-product focus areas.
	PrintSimpleCountDataToConsole(appliedUsageExampleCounter.SubProductCounts, "SubProduct", []interface{}{"SubProduct", "Count"}, []int{20, 15})

	// This prints a table showing counts of new applied usage examples broken down by product. Every docs page has a
	// Product, and some but not all docs pages have a Sub-Product. The Sub-Product counts are a subset of Product counts.
	// For example, Vector Search sub-product code example counts are *also* reported separately in the Atlas product
	// counts - i.e. 46 Atlas code examples would *include* 43 Vector Search sub-product counts.
	PrintSimpleCountDataToConsole(appliedUsageExampleCounter.ProductCounts, "Product", []interface{}{"Product", "Count"}, []int{40, 15})
}
