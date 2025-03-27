package utils

import (
	"dodec/types"
	"fmt"
)

func PrintPageIdNewAppliedUsageExampleCounts(mapToPrint map[string][]types.PageIdNewAppliedUsageExamples) {
	columnNames := []interface{}{"Page ID", "Count"}
	// Print a separate table for each top-level element
	columnWidths := []int{70, 15}
	aggregateCount := 0
	for collectionName, pagesToPrintInCollection := range mapToPrint {
		collectionCount := 0
		fmt.Printf("\nNew Applied Usage Example Counts by Page in Collection %s\n", collectionName)
		printSeparator(columnWidths...)
		printRow(columnWidths, columnNames...)
		printSeparator(columnWidths...)
		// This type also stores the code nodes directly - do we want to print any details about the specific nodes that
		// match our conditions?
		for _, page := range pagesToPrintInCollection {
			printRow(columnWidths, page.ID, page.Count)
			aggregateCount += page.Count
			collectionCount += page.Count
		}
		printSeparator(columnWidths...)
		fmt.Printf("\nTotal new applied usage example counts in %s: %d\n", collectionName, collectionCount)
	}
	fmt.Printf("\nTotal New Applied Usage Example Counts in Last Week: %d\n", aggregateCount)
}
