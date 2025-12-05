// Package tested_examples provides output formatting for count results.
package tested_examples

import (
	"fmt"
	"sort"
)

// PrintResults prints the counting results.
//
// If countByProduct is true, prints a breakdown by product.
// Otherwise, prints only the total count.
//
// Parameters:
//   - result: The counting results
//   - countByProduct: If true, show breakdown by product
func PrintResults(result *CountResult, countByProduct bool) {
	if countByProduct {
		printByProduct(result)
	} else {
		printTotal(result)
	}
}

// printTotal prints only the total count as a single integer.
func printTotal(result *CountResult) {
	fmt.Println(result.TotalCount)
}

// printByProduct prints a breakdown of counts by product.
func printByProduct(result *CountResult) {
	if len(result.ProductCounts) == 0 {
		fmt.Println("No files found")
		return
	}

	// Get sorted list of product keys
	var productKeys []string
	for key := range result.ProductCounts {
		productKeys = append(productKeys, key)
	}
	sort.Strings(productKeys)

	// Print header
	fmt.Println("Product Counts:")
	fmt.Println()

	// Print each product with its count
	for _, key := range productKeys {
		count := result.ProductCounts[key]
		productInfo := ProductMap[key]
		fmt.Printf("  %-25s %5d  (%s)\n", key, count, productInfo.Name)
	}

	// Print total
	fmt.Println()
	fmt.Printf("Total: %d\n", result.TotalCount)
}

