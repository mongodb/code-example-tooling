package main

import (
	"fmt"
	"log"
	"pull-audit-data/types"
	"sort"
)

// Helper function to print a separator line for tables
func printSeparator(columns ...int) {
	for _, width := range columns {
		fmt.Print("+")
		for i := 0; i < width; i++ {
			fmt.Print("-")
		}
	}
	fmt.Println("+")
}

// Helper function to print formatted table rows
func printRow(columnWidth []int, columns ...interface{}) {
	for i, col := range columns {
		fmt.Printf("| %-*v ", columnWidth[i], col)
	}
	fmt.Println("|")
}

// PrintSimpleCountDataToConsole expects two string column names and two column width ints
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
	}
	var elementCounts []types.KeyCount
	// Sort keys by count
	for key, value := range simpleMap {
		if key != "total" {
			elementCounts = append(elementCounts, types.KeyCount{Key: key, Count: value})
		}
	}
	sort.Slice(elementCounts, func(i, j int) bool {
		return elementCounts[i].Count > elementCounts[j].Count
	})
	fmt.Printf("\n%s Counts\n", tableLable)
	fmt.Printf("Total code examples by %s: %d\n", tableLable, totalCount)
	// If there is a "total" key in the map, we want to present a column that displays the count as a percent of the total
	// Otherwise, we omit the percent column
	if totalCount > 0 {
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
	} else {
		printSeparator(columnWidths...)
		printRow(columnWidths, columnNames...)
		printSeparator(columnWidths...)
		for _, item := range elementCounts {
			printRow(columnWidths, item.Key, item.Count)
		}
		printSeparator(columnWidths...)
	}
}

// PrintNestedOneLevelCountDataToConsole expects two string column names and two column width ints
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

// PrintNestedTwoLevelCountDataToConsole expects two string column names and two column width ints
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
