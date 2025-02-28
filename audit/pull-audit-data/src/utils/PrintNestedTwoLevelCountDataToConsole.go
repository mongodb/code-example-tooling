package utils

import (
	"fmt"
	"log"
	"pull-audit-data/types"
	"sort"
)

// PrintNestedTwoLevelCountDataToConsole prints a nicely formatted series of tables with three columns. The top level map
// key is used in the table label, as well as the first level nested key. The second level key becomes the first column,
// the int counts become the second column, and a "% of Total" column is automatically calculated by dividing the key
// count from the total count across the map. This function expects columnNames to include two string column names, which
// are used as column labels. It expects columnWidths to contain two column width ints, so you can make the columns as
// wide as needed to accommodate your column names.
func PrintNestedTwoLevelCountDataToConsole(nestedTwoLevelCountMap map[string]map[string]map[string]int, tableLabel string, columnNames []interface{}, columnWidths []int) {
	if len(columnNames) != len(columnWidths) {
		log.Fatalf("Got %d column names, but %d column widths - can't print the table unless we have the same number of names and widths", len(columnNames), len(columnWidths))
	}
	columnNames = append(columnNames, "% of Total")
	columnWidths = append(columnWidths, 18)
	// Print a separate table for each nestedOneLevelMap within each top-level key
	for topLevelElement, nestedOneLevelMap := range nestedTwoLevelCountMap {
		if len(nestedOneLevelMap) == 0 {
			continue
		}
		for nestedOneLevelMapKey, secondLevelNestedMap := range nestedOneLevelMap {
			// Calculate total sum for this nestedOneLevelMapKey
			nestedOneLevelMapKeyCount := 0
			for _, count := range secondLevelNestedMap {
				nestedOneLevelMapKeyCount += count
			}

			// Sort secondLevelNestedMap by count
			var secondLevelNestedMapItems []types.TwoLevelNestedKeyCount
			for secondLevelMapKey, count := range secondLevelNestedMap {
				secondLevelNestedMapItems = append(secondLevelNestedMapItems, types.TwoLevelNestedKeyCount{TopLevelKey: topLevelElement, NestedMapKey: nestedOneLevelMapKey, SecondLevelNestedMapKey: secondLevelMapKey, Count: count})
			}
			sort.Slice(secondLevelNestedMapItems, func(i, j int) bool {
				return secondLevelNestedMapItems[i].Count > secondLevelNestedMapItems[j].Count
			})
			fmt.Printf("\n%s Counts for: %s - %s\n", tableLabel, topLevelElement, nestedOneLevelMapKey)
			fmt.Printf("Total code examples by %s: %d\n", nestedOneLevelMapKey, nestedOneLevelMapKeyCount)
			printSeparator(columnWidths...)
			printRow(columnWidths, columnNames...)
			printSeparator(columnWidths...)
			for _, item := range secondLevelNestedMapItems {
				percent := float64(item.Count) / float64(nestedOneLevelMapKeyCount) * 100
				printRow(columnWidths, item.SecondLevelNestedMapKey, item.Count, fmt.Sprintf("%.1f%%", percent))
			}
			printSeparator(columnWidths...)
		}
	}
}
