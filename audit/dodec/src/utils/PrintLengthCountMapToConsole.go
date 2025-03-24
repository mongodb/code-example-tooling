package utils

import (
	"dodec/types"
	"fmt"
)

// PrintCodeLengthMapToConsole prints a nicely-formatted table with columns for the docs project name, columns for the
// minimum, median, and maximum character counts in the collection, and a column that counts the number of one-line code
// examples in the collection.
func PrintCodeLengthMapToConsole(lengthCountMap map[string]types.CodeLengthStats) {
	tableLable := "Minimum, median, and maximum code character count, and one-line count, by collection."
	columnNames := []interface{}{"Project", "Min", "Med", "Max", "One line"}
	columnWidths := []int{25, 10, 10, 10, 10}

	minAccumulator := 0
	medianAccumulator := 0
	maxAccumulator := 0
	collectionCount := 0
	shortCodeCount := 0

	fmt.Printf("\n%s\n", tableLable)
	printSeparator(columnWidths...)
	printRow(columnWidths, columnNames...)
	printSeparator(columnWidths...)
	for name, stats := range lengthCountMap {
		printRow(columnWidths, name, stats.Min, stats.Median, stats.Max, stats.ShortCodeCount)
		minAccumulator += stats.Min
		medianAccumulator += stats.Median
		maxAccumulator += stats.Max
		shortCodeCount += stats.ShortCodeCount
		collectionCount++
	}
	printSeparator(columnWidths...)

	fmt.Printf("Aggregate min: %d\n", minAccumulator/collectionCount)
	fmt.Printf("Aggregate median: %d\n", medianAccumulator/collectionCount)
	fmt.Printf("Aggregate max: %d\n", maxAccumulator/collectionCount)
	fmt.Printf("Total one-line count across collections: %d\n", shortCodeCount)
}
