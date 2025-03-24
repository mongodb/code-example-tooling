package utils

import (
	"dodec/types"
	"fmt"
	"log"
	"sort"
)

// PrintNestedOneLevelCountDataToConsole prints a nicely formatted series of tables with three columns. The first key
// name is used in the table label, and the second key name is used as the first column. The int counts for each key
// become a column, and a "% of Total" column is automatically calculated by dividing the key count from the total count
// across the map. This function expects columnNames to include two string column names, which are used as column labels.
// It expects columnWidths to contain two column width ints, so you can make the columns as wide as needed to
// accommodate your column names.
func PrintNestedOneLevelCountDataToConsole(nestedOneLevelMap map[string]map[string]int, tableLabel string, columnNames []interface{}, columnWidths []int) {
	if len(columnNames) != len(columnWidths) {
		log.Fatalf("Got %d column names, but %d column widths - can't print the table unless we have the same number of names and widths", len(columnNames), len(columnWidths))
	}
	// Print a separate table for each top-level element
	columnNames = append(columnNames, "% of Total")
	columnWidths = append(columnWidths, 18)
	for topLevelElement, nestedMap := range nestedOneLevelMap {
		// Calculate total sum for this top-level element
		totalTopLevelElementCount := 0
		for _, count := range nestedMap {
			totalTopLevelElementCount += count
		}

		// Sort nested map by count
		var nestedMapElements []types.KeyCount
		for key, count := range nestedMap {
			nestedMapElements = append(nestedMapElements, types.KeyCount{Key: key, Count: count})
		}
		sort.Slice(nestedMapElements, func(i, j int) bool {
			return nestedMapElements[i].Count > nestedMapElements[j].Count
		})
		fmt.Printf("\n%s Counts for: %s\n", tableLabel, topLevelElement)
		fmt.Printf("Total code examples by %s: %d\n", topLevelElement, totalTopLevelElementCount)
		printSeparator(columnWidths...)
		printRow(columnWidths, columnNames...)
		printSeparator(columnWidths...)
		for _, item := range nestedMapElements {
			percent := float64(item.Count) / float64(totalTopLevelElementCount) * 100
			printRow(columnWidths, item.Key, item.Count, fmt.Sprintf("%.1f%%", percent))
		}
		printSeparator(columnWidths...)
	}
}
