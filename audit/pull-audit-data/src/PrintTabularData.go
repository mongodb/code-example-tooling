package main

import (
	"fmt"
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

func PrintProductCategoryData(productCategoryMap map[string]map[string]int) {
	// Column widths for table formatting
	columnWidths := []int{30, 15, 20}
	// Print a separate table for each Product
	for product, categories := range productCategoryMap {
		// Calculate total sum for this Product
		totalProductCount := 0
		for _, count := range categories {
			totalProductCount += count
		}

		// Sort categories by count
		var productCategories []types.ProductCategoryCount
		for category, count := range categories {
			productCategories = append(productCategories, types.ProductCategoryCount{Category: category, Count: count})
		}
		sort.Slice(productCategories, func(i, j int) bool {
			return productCategories[i].Count > productCategories[j].Count
		})
		fmt.Printf("\nProduct Category Counts for: %s\n", product)
		fmt.Printf("Total code examples by product: %d\n", totalProductCount)
		printSeparator(columnWidths...)
		printRow(columnWidths, "Category", "Counts", "% of Total")
		printSeparator(columnWidths...)
		for _, pc := range productCategories {
			percent := float64(pc.Count) / float64(totalProductCount) * 100
			printRow(columnWidths, pc.Category, pc.Count, fmt.Sprintf("%.1f%%", percent))
		}
		printSeparator(columnWidths...)
	}
}

func PrintSubProductCategoryData(subProductCategoryMap map[string]map[string]map[string]int) {
	columnWidths := []int{30, 15, 20}
	// Print a separate table for each SubProduct within each Product
	// Print a separate table for each SubProduct within each Product
	for product, subProducts := range subProductCategoryMap {
		for subProduct, categories := range subProducts {
			// Calculate total sum for this sub_product
			totalSubProductCount := 0
			for _, count := range categories {
				totalSubProductCount += count
			}

			// Sort categories by count
			var subProductCategories []types.SubProductCategoryCount
			for category, count := range categories {
				subProductCategories = append(subProductCategories, types.SubProductCategoryCount{SubProduct: subProduct, Category: category, Count: count})
			}
			sort.Slice(subProductCategories, func(i, j int) bool {
				return subProductCategories[i].Count > subProductCategories[j].Count
			})
			fmt.Printf("\nSubProduct Category Counts for: %s - %s\n", product, subProduct)
			fmt.Printf("Total code examples by sub-product: %d\n", totalSubProductCount)
			printSeparator(columnWidths...)
			printRow(columnWidths, "Category", "Counts", "% of Total")
			printSeparator(columnWidths...)
			for _, spc := range subProductCategories {
				percent := float64(spc.Count) / float64(totalSubProductCount) * 100
				printRow(columnWidths, spc.Category, spc.Count, fmt.Sprintf("%.1f%%", percent))
			}
			printSeparator(columnWidths...)
		}
	}
}

func PrintProductLanguageData(productLanguageMap map[string]map[string]int) {
	// Column widths for table formatting
	columnWidths := []int{20, 15, 18}
	// Print a separate table for each Product
	for product, languages := range productLanguageMap {
		// Calculate total sum for this Product
		totalProductCount := 0
		for _, count := range languages {
			totalProductCount += count
		}

		// Sort categories by count
		var productLanguages []types.LanguageCount
		for language, count := range languages {
			productLanguages = append(productLanguages, types.LanguageCount{Language: language, Count: count})
		}
		sort.Slice(productLanguages, func(i, j int) bool {
			return productLanguages[i].Count > productLanguages[j].Count
		})
		fmt.Printf("\nProduct Language Counts for: %s\n", product)
		fmt.Printf("Total code examples by product: %d\n", totalProductCount)
		printSeparator(columnWidths...)
		printRow(columnWidths, "Language", "Counts", "% of Total")
		printSeparator(columnWidths...)
		for _, lc := range productLanguages {
			percent := float64(lc.Count) / float64(totalProductCount) * 100
			printRow(columnWidths, lc.Language, lc.Count, fmt.Sprintf("%.1f%%", percent))
		}
		printSeparator(columnWidths...)
	}
}

func PrintSubProductLanguageData(subProductLanguageMap map[string]map[string]map[string]int) {
	columnWidths := []int{20, 15, 18}
	// Print a separate table for each SubProduct within each Product
	for product, subProducts := range subProductLanguageMap {
		if len(subProducts) == 0 {
			continue
		}
		for subProduct, languages := range subProducts {
			// Calculate total sum for this sub_product
			totalSubProductCount := 0
			for _, count := range languages {
				totalSubProductCount += count
			}

			// Sort languages by count
			var subProductLanguages []types.SubProductLanguageCount
			for language, count := range languages {
				subProductLanguages = append(subProductLanguages, types.SubProductLanguageCount{Product: product, SubProduct: subProduct, Language: language, Count: count})
			}
			sort.Slice(subProductLanguages, func(i, j int) bool {
				return subProductLanguages[i].Count > subProductLanguages[j].Count
			})
			fmt.Printf("\nSubProduct Language Counts for: %s - %s\n", product, subProduct)
			fmt.Printf("Total code examples by sub-product: %d\n", totalSubProductCount)
			printSeparator(columnWidths...)
			printRow(columnWidths, "Language", "Counts", "% of Total")
			printSeparator(columnWidths...)
			for _, spl := range subProductLanguages {
				percent := float64(spl.Count) / float64(totalSubProductCount) * 100
				printRow(columnWidths, spl.Language, spl.Count, fmt.Sprintf("%.1f%%", percent))
			}
			printSeparator(columnWidths...)
		}
	}
}
