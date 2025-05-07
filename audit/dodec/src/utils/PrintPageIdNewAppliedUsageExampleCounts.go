package utils

import (
	"dodec/types"
	"fmt"
	"log"
	"sort"
)

// PrintPageIdNewAppliedUsageExampleCounts prints a nicely-formatted series of tables with counts of new applied usage
// examples broken down in various ways. Each collection prints as its own table, which has a list of page IDs and counts
// for pages that have new usage examples. This makes it easy to validate the data if we want to perform manual QA. After
// the individual collection tables, we print a consolidated table breaking down the counts by sub-product and product.
func PrintPageIdNewAppliedUsageExampleCounts(productSubProductCounter types.NewAppliedUsageExampleCounterByProductSubProduct) {
	// Print one table for each product, where each row represents the count of new applied usage examples on a given page, specified here by ID
	// This information is useful for debugging/validating counts.
	pageIdTableColumnNames := []interface{}{"Product", "Sub Product", "Page ID", "Count"}
	pageIdTableColumnWidths := []int{30, 30, 70, 15}
	fmt.Printf("\nPrinting breakdown of new applied usage example counts by page ID in collections. This info is for validating and/or debugging counts.\n")
	for collectionName, pagesToPrintInCollection := range productSubProductCounter.PagesInCollections {
		collectionCount := 0
		fmt.Printf("\nTotal new Applied Usage Example Counts by Page in Collection %s\n", collectionName)
		printSeparator(pageIdTableColumnWidths...)
		printRow(pageIdTableColumnWidths, pageIdTableColumnNames...)
		printSeparator(pageIdTableColumnWidths...)
		for _, page := range pagesToPrintInCollection {
			printRow(pageIdTableColumnWidths, page.ID.Product, page.ID.SubProduct, page.ID.DocumentID, page.Count)
			collectionCount += page.Count
		}
		printSeparator(pageIdTableColumnWidths...)
		fmt.Printf("Total new applied usage example counts in collection %s: %d\n", collectionName, collectionCount)
	}

	fmt.Printf("\nPrinting new applied usage example counts aggregate table - share data below with team lead.\n")

	// Print one consolidated table that lists just the counts for each product/sub-product. This information is used to report to the leads.
	appliedUsageExampleCountsColumnNames := []interface{}{"Product", "Count", "Sub-product", "Count"}
	appliedUsageExampleCountsColumnWidths := []int{20, 18, 25, 18}

	if len(appliedUsageExampleCountsColumnNames) != len(appliedUsageExampleCountsColumnWidths) {
		log.Fatalf("Got %d column names, but %d column widths - can't print the table unless we have the same number of names and widths", len(appliedUsageExampleCountsColumnNames), len(appliedUsageExampleCountsColumnWidths))
	}

	// Gather all data together into a sortable slice
	var allEntries []struct {
		ProductName     string
		ProductCount    int
		SubProductName  string
		SubProductCount int
	}

	totalCount := 0
	for productName, subProductMap := range productSubProductCounter.ProductSubProductCounts {
		aggregateCount, exists := productSubProductCounter.ProductAggregateCount[productName]
		if !exists {
			log.Fatalf("No aggregate count found for product %s", productName)
		}
		totalCount += aggregateCount

		subProductSum := 0
		for _, subProductCount := range subProductMap {
			subProductSum += subProductCount
		}

		for subProductName, subProductCount := range subProductMap {
			allEntries = append(allEntries, struct {
				ProductName     string
				ProductCount    int
				SubProductName  string
				SubProductCount int
			}{
				ProductName:     productName,
				ProductCount:    aggregateCount,
				SubProductName:  subProductName,
				SubProductCount: subProductCount,
			})
		}

		// Add "None" sub-product if sub-product sum is less than aggregate count
		if subProductSum < aggregateCount {
			allEntries = append(allEntries, struct {
				ProductName     string
				ProductCount    int
				SubProductName  string
				SubProductCount int
			}{
				ProductName:     productName,
				ProductCount:    aggregateCount,
				SubProductName:  "None",
				SubProductCount: aggregateCount - subProductSum,
			})
		}
	}

	// Sort entries by product name
	sort.Slice(allEntries, func(i, j int) bool {
		return allEntries[i].ProductName < allEntries[j].ProductName
	})

	// Print a single consolidated table
	fmt.Printf("\nTotal New Applied Usage Example Counts in Last Week: %d\n", totalCount)
	printSeparator(appliedUsageExampleCountsColumnWidths...)
	printRow(appliedUsageExampleCountsColumnWidths, appliedUsageExampleCountsColumnNames...)
	printSeparator(appliedUsageExampleCountsColumnWidths...)

	var lastProductName string
	for _, entry := range allEntries {
		productNameToDisplay := ""
		aggregateCountToDisplay := ""

		// Only display the product and its aggregate count once for a group
		if entry.ProductName != lastProductName {
			productNameToDisplay = entry.ProductName
			aggregateCountToDisplay = fmt.Sprintf("%d", entry.ProductCount)
			lastProductName = entry.ProductName
		}

		printRow(appliedUsageExampleCountsColumnWidths, productNameToDisplay, aggregateCountToDisplay, entry.SubProductName, entry.SubProductCount)
	}

	printSeparator(appliedUsageExampleCountsColumnWidths...)
}
