package find_string

import (
	"fmt"
	"sort"
	"strings"
)

// PrintReport prints the search report to stdout.
//
// Displays statistics about the search operation including:
//   - Number of files scanned
//   - Number of files containing the substring
//   - Files containing substring by language (if verbose is true)
//   - List of file paths containing the substring (if verbose is true)
//
// Parameters:
//   - report: The report to print
//   - verbose: If true, show detailed breakdown including file paths and language counts
func PrintReport(report *SearchReport, verbose bool) {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("SEARCH REPORT")
	fmt.Println(strings.Repeat("=", 60))

	fmt.Printf("\nFiles Scanned: %d\n", report.FilesScanned)
	fmt.Printf("Files Containing Substring: %d\n", report.FilesContaining)

	if verbose && len(report.LanguageCounts) > 0 {
		fmt.Println("\nFiles Containing Substring by Language:")

		languages := make([]string, 0, len(report.LanguageCounts))
		for lang := range report.LanguageCounts {
			languages = append(languages, lang)
		}
		sort.Strings(languages)

		for _, lang := range languages {
			count := report.LanguageCounts[lang]
			fmt.Printf("  %-15s: %d\n", lang, count)
		}
	}

	if verbose && len(report.FilesWithSubstring) > 0 {
		fmt.Println("\nFiles Containing Substring:")
		for _, path := range report.FilesWithSubstring {
			fmt.Printf("  - %s\n", path)
		}
	}

	fmt.Println(strings.Repeat("=", 60))
}
