package main

func GetCategorySums(counts map[string]map[string]int) map[string]map[string]int {
	for category, languageCounts := range counts {
		// Initialize a sum variable for the current category
		sum := 0
		// Iterate over each language in the inner map
		for _, count := range languageCounts {
			// Accumulate the total count
			sum += count
		}
		counts[category]["totals"] = sum
	}
	return counts
}
