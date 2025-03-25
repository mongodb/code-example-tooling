package utils

import (
	"dodec/types"
	"fmt"
)

func PrintPageIdChangesCountMap(mapToPrint map[string][]types.PageIdChangedCounts) {
	columnNameStrings := []string{"Page ID", "Added", "Updated", "Removed"}
	columnNames := []interface{}{columnNameStrings[0], columnNameStrings[1], columnNameStrings[2], columnNameStrings[3]}
	// Print a separate table for each top-level element
	columnWidths := []int{70, 15, 15, 15}
	for collectionName, pagesToPrintInCollection := range mapToPrint {
		fmt.Printf("\nRecently Updated Pages in Collection %s\n", collectionName)
		printSeparator(columnWidths...)
		printRow(columnWidths, columnNames...)
		printSeparator(columnWidths...)
		for _, page := range pagesToPrintInCollection {
			printRow(columnWidths, page.ID, page.AddedCount, page.UpdatedCount, page.RemovedCount)
		}
		printSeparator(columnWidths...)
	}
}
