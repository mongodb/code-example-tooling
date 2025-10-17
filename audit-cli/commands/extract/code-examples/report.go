package code_examples

import (
	"fmt"
	"sort"
	"strings"
)

// PrintReport prints the extraction report to stdout.
//
// Displays statistics about the extraction operation including:
//   - Number of files traversed
//   - Number of output files written
//   - Code examples by language (summary or detailed based on verbose flag)
//   - Code examples by directive type
//   - Per-source-file statistics (if verbose is true)
//
// Parameters:
//   - report: The report to print
//   - verbose: If true, show detailed breakdown including file paths and per-source stats
func PrintReport(report *Report, verbose bool) {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("CODE EXTRACTION REPORT")
	fmt.Println(strings.Repeat("=", 60))

	fmt.Printf("\nFiles Traversed: %d\n", report.FilesTraversed)
	if verbose && len(report.TraversedFilepaths) > 0 {
		fmt.Println("\nTraversed Filepaths:")
		for _, path := range report.TraversedFilepaths {
			fmt.Printf("  - %s\n", path)
		}
	}

	fmt.Printf("\nOutput Files Written: %d\n", report.OutputFilesWritten)

	if len(report.LanguageCounts) > 0 {
		fmt.Println("\nCode Examples by Language:")

		languages := make([]string, 0, len(report.LanguageCounts))
		for lang := range report.LanguageCounts {
			languages = append(languages, lang)
		}
		sort.Strings(languages)

		if verbose {
			for _, lang := range languages {
				count := report.LanguageCounts[lang]
				fmt.Printf("  %-15s: %d\n", lang, count)
			}
		} else {
			total := 0
			for _, count := range report.LanguageCounts {
				total += count
			}
			fmt.Printf("  Total: %d (use --verbose for breakdown)\n", total)
		}
	}

	if len(report.DirectiveCounts) > 0 {
		fmt.Println("\nCode Examples by Directive Type:")

		directives := []DirectiveType{CodeBlock, LiteralInclude, IoCodeBlock}
		for _, directive := range directives {
			if count, exists := report.DirectiveCounts[directive]; exists {
				fmt.Printf("  %-20s: %d\n", directive, count)
			}
		}
	}

	if verbose && len(report.SourcePathStats) > 0 {
		fmt.Println("\nStatistics by Source File:")

		sourcePaths := make([]string, 0, len(report.SourcePathStats))
		for path := range report.SourcePathStats {
			sourcePaths = append(sourcePaths, path)
		}
		sort.Strings(sourcePaths)

		for _, sourcePath := range sourcePaths {
			stats := report.SourcePathStats[sourcePath]
			fmt.Printf("\n  %s:\n", sourcePath)

			if len(stats.DirectiveCounts) > 0 {
				fmt.Println("    Directives:")
				directives := []DirectiveType{CodeBlock, LiteralInclude, IoCodeBlock}
				for _, directive := range directives {
					if count, exists := stats.DirectiveCounts[directive]; exists {
						fmt.Printf("      %-20s: %d\n", directive, count)
					}
				}
			}

			if len(stats.LanguageCounts) > 0 {
				fmt.Println("    Languages:")
				languages := make([]string, 0, len(stats.LanguageCounts))
				for lang := range stats.LanguageCounts {
					languages = append(languages, lang)
				}
				sort.Strings(languages)

				for _, lang := range languages {
					count := stats.LanguageCounts[lang]
					fmt.Printf("      %-15s: %d\n", lang, count)
				}
			}

			if len(stats.OutputFiles) > 0 {
				fmt.Printf("    Output Files: %d\n", len(stats.OutputFiles))
			}
		}
	}

	fmt.Println("\n" + strings.Repeat("=", 60))
}
