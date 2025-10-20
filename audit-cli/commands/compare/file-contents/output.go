package file_contents

import (
	"fmt"
	"strings"
)

// PrintComparisonResult prints the comparison result with progressive detail levels.
//
// The output format depends on the flags:
//   - Default: Summary only
//   - showPaths: Summary + file paths
//   - showDiff: Summary + paths + diffs
//
// Parameters:
//   - result: The comparison result to print
//   - showPaths: If true, show file paths
//   - showDiff: If true, show diffs (implies showPaths)
func PrintComparisonResult(result *ComparisonResult, showPaths bool, showDiff bool) {
	// If showDiff is true, we also need to show paths
	if showDiff {
		showPaths = true
	}

	// Print summary
	printSummary(result)

	// Print paths if requested
	if showPaths {
		fmt.Println()
		printPaths(result)
	}

	// Print diffs if requested
	if showDiff {
		fmt.Println()
		printDiffs(result)
	}
}

// printSummary prints a summary of the comparison results.
func printSummary(result *ComparisonResult) {
	if result.ReferenceVersion != "" {
		// Version comparison mode
		fmt.Printf("Comparing file across %d versions...\n", result.TotalFiles)
	} else {
		// Direct comparison mode
		fmt.Println("Comparing files...")
	}

	if result.AllMatch() {
		// All files match
		fmt.Printf("✓ All versions match (%d/%d files identical)\n", result.MatchingFiles, result.TotalFiles)
	} else if result.HasDifferences() {
		// Some files differ
		fmt.Printf("⚠ Differences found: %d of %d versions differ", result.DifferingFiles, result.TotalFiles)
		if result.ReferenceVersion != "" {
			fmt.Printf(" from %s\n", result.ReferenceVersion)
		} else {
			fmt.Println()
		}

		// Show breakdown
		if result.MatchingFiles > 0 {
			fmt.Printf("  - %d version(s) match\n", result.MatchingFiles)
		}
		if result.DifferingFiles > 0 {
			fmt.Printf("  - %d version(s) differ\n", result.DifferingFiles)
		}
		if result.NotFoundFiles > 0 {
			fmt.Printf("  - %d version(s) not found (file does not exist)\n", result.NotFoundFiles)
		}
		if result.ErrorFiles > 0 {
			fmt.Printf("  - %d version(s) had errors\n", result.ErrorFiles)
		}

		// Show hints
		fmt.Println()
		fmt.Println("Use --show-paths to see which files differ")
		fmt.Println("Use --show-diff to see the differences")
	} else if result.NotFoundFiles > 0 || result.ErrorFiles > 0 {
		// No differences, but some files not found or had errors
		fmt.Printf("✓ No differences found among existing files\n")
		if result.NotFoundFiles > 0 {
			fmt.Printf("  - %d version(s) not found (file does not exist)\n", result.NotFoundFiles)
		}
		if result.ErrorFiles > 0 {
			fmt.Printf("  - %d version(s) had errors\n", result.ErrorFiles)
		}
	}
}

// printPaths prints the file paths grouped by status.
func printPaths(result *ComparisonResult) {
	// Group comparisons by status
	var matching, differing, notFound, errors []FileComparison
	for _, comp := range result.Comparisons {
		switch comp.Status {
		case FileMatches:
			matching = append(matching, comp)
		case FileDiffers:
			differing = append(differing, comp)
		case FileNotFound:
			notFound = append(notFound, comp)
		case FileError:
			errors = append(errors, comp)
		}
	}

	// Print matching files
	if len(matching) > 0 {
		fmt.Println("Files that match:")
		for _, comp := range matching {
			if comp.Version == result.ReferenceVersion {
				fmt.Printf("  ✓ %s (reference)\n", comp.FilePath)
			} else {
				fmt.Printf("  ✓ %s\n", comp.FilePath)
			}
		}
	}

	// Print differing files
	if len(differing) > 0 {
		if len(matching) > 0 {
			fmt.Println()
		}
		fmt.Println("Files that differ:")
		for _, comp := range differing {
			fmt.Printf("  ✗ %s\n", comp.FilePath)
		}
	}

	// Print not found files
	if len(notFound) > 0 {
		if len(matching) > 0 || len(differing) > 0 {
			fmt.Println()
		}
		fmt.Println("Files not found:")
		for _, comp := range notFound {
			fmt.Printf("  - %s\n", comp.FilePath)
		}
	}

	// Print error files
	if len(errors) > 0 {
		if len(matching) > 0 || len(differing) > 0 || len(notFound) > 0 {
			fmt.Println()
		}
		fmt.Println("Files with errors:")
		for _, comp := range errors {
			fmt.Printf("  ⚠ %s: %v\n", comp.FilePath, comp.Error)
		}
	}
}

// printDiffs prints the unified diffs for files that differ.
func printDiffs(result *ComparisonResult) {
	// Find files with diffs
	var diffsToShow []FileComparison
	for _, comp := range result.Comparisons {
		if comp.Status == FileDiffers && comp.Diff != "" {
			diffsToShow = append(diffsToShow, comp)
		}
	}

	if len(diffsToShow) == 0 {
		return
	}

	fmt.Println("Diffs:")
	fmt.Println(strings.Repeat("=", 80))

	for i, comp := range diffsToShow {
		if i > 0 {
			fmt.Println()
		}

		// Print header
		if result.ReferenceVersion != "" {
			fmt.Printf("Diff: %s vs %s\n", result.ReferenceVersion, comp.Version)
		} else {
			fmt.Printf("Diff: %s\n", comp.Version)
		}
		fmt.Println(strings.Repeat("-", 80))

		// Print the diff
		fmt.Print(comp.Diff)

		// Ensure there's a newline at the end
		if !strings.HasSuffix(comp.Diff, "\n") {
			fmt.Println()
		}
	}

	fmt.Println(strings.Repeat("=", 80))
}

