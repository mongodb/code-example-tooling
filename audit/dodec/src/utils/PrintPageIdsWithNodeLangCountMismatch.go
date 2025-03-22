package utils

import "fmt"

func PrintPageIdsWithNodeLangCountMismatch(mapToPrint map[string][]string) {
	columnNames := []interface{}{"Page Ids"}
	// Print a separate table for each top-level element
	columnWidths := []int{70}
	for collectionName, pagesToPrintInCollection := range mapToPrint {
		fmt.Printf("\nPages with Node/Lang array count mismatch in Collection %s\n", collectionName)
		printSeparator(columnWidths...)
		printRow(columnWidths, columnNames...)
		printSeparator(columnWidths...)
		for _, page := range pagesToPrintInCollection {
			printRow(columnWidths, page)
		}
		printSeparator(columnWidths...)
	}
}
