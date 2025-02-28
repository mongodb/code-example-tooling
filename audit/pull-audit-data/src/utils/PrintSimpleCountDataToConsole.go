package utils

import (
	"fmt"
	"log"
	"pull-audit-data/types"
	"sort"
)

// PrintSimpleCountDataToConsole prints a nicely formatted table with three columns; the key names as a column,
// the int counts for each key as a column, and a "% of Total" column that is automatically calculated by dividing the key
// count from the total count across the map. This function expects columnNames to include two string column names, which
// are used as column labels. It expects columnWidths to contain two column width ints, so you can make the columns as
// wide as needed to accommodate your column names.
func PrintSimpleCountDataToConsole(simpleMap map[string]int, tableLable string, columnNames []interface{}, columnWidths []int) {
	if len(columnNames) != len(columnWidths) {
		log.Fatalf("Got %d column names, but %d column widths - can't print the table unless we have the same number of names and widths", len(columnNames), len(columnWidths))
	}
	totalValue, totalExists := simpleMap["total"]
	var totalCount int
	if totalExists {
		totalCount = totalValue
	} else {
		totalCount = 0
		for _, count := range simpleMap {
			totalCount += count
		}
	}
	var elementCounts []types.KeyCount
	// Sort keys by count
	for key, value := range simpleMap {
		if key != "total" {
			elementCounts = append(elementCounts, types.KeyCount{Key: key, Count: value})
		}
		if key == "" {
			fmt.Println("Found an empty string key whose count is ", value)
		}
	}
	sort.Slice(elementCounts, func(i, j int) bool {
		return elementCounts[i].Count > elementCounts[j].Count
	})
	fmt.Printf("\n%s Counts\n", tableLable)
	fmt.Printf("Total code examples by %s: %d\n", tableLable, totalCount)
	columnNames = append(columnNames, "% of Total")
	columnWidths = append(columnWidths, 18)
	printSeparator(columnWidths...)
	printRow(columnWidths, columnNames...)
	printSeparator(columnWidths...)
	for _, item := range elementCounts {
		percent := float64(item.Count) / float64(totalCount) * 100
		printRow(columnWidths, item.Key, item.Count, fmt.Sprintf("%.1f%%", percent))
	}
	printSeparator(columnWidths...)
}
