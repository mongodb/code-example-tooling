package utils

import "fmt"

// Helper function to print a separator line for tables
func printSeparator(columns ...int) {
	for _, width := range columns {
		fmt.Print("+")
		for i := 0; i < width+2; i++ {
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
