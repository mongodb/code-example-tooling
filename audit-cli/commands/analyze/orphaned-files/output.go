package orphanedfiles

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
)

// PrintText prints the analysis results in human-readable text format.
//
// This function displays the orphaned files with helpful context and suggestions.
//
// Parameters:
//   - analysis: The analysis results to print
//   - verbose: If true, show additional details
//
// Returns:
//   - error: Any error encountered during printing
func PrintText(analysis *OrphanedFilesAnalysis, verbose bool) error {
	fmt.Printf("Orphaned Files Analysis\n")
	fmt.Printf("=======================\n\n")
	fmt.Printf("Source Directory: %s\n", analysis.SourceDir)
	fmt.Printf("Total Files: %d\n", analysis.TotalFiles)
	fmt.Printf("Orphaned Files: %d\n", analysis.TotalOrphaned)

	if analysis.IncludedToctree {
		fmt.Printf("Reference Types: include, literalinclude, io-code-block, toctree\n")
	} else {
		fmt.Printf("Reference Types: include, literalinclude, io-code-block\n")
		fmt.Printf("(Use --include-toctree to also consider toctree references)\n")
	}

	fmt.Println()

	if analysis.TotalOrphaned == 0 {
		fmt.Println("✓ No orphaned files found!")
		fmt.Println()
		fmt.Println("All files in the source directory are referenced by at least one other file.")
		return nil
	}

	fmt.Printf("The following %d file(s) have no incoming references:\n\n", analysis.TotalOrphaned)

	// Sort files for consistent output
	sortedFiles := make([]string, len(analysis.OrphanedFiles))
	copy(sortedFiles, analysis.OrphanedFiles)
	sort.Strings(sortedFiles)

	for _, file := range sortedFiles {
		fmt.Printf("  - %s\n", file)
	}

	fmt.Println()
	fmt.Println("These files might be:")
	fmt.Println("  • Unused include files that can be removed")
	fmt.Println("  • Documentation pages not linked in the navigation")
	fmt.Println("  • Entry points (like index.rst) that are referenced externally")
	fmt.Println("  • Legacy content that needs cleanup")
	fmt.Println()
	fmt.Println("Review each file to determine if it should be:")
	fmt.Println("  1. Kept (if it's an entry point or externally referenced)")
	fmt.Println("  2. Linked (if it should be in the navigation or included somewhere)")
	fmt.Println("  3. Removed (if it's truly unused)")

	return nil
}

// PrintJSON prints the analysis results in JSON format.
//
// This function outputs the complete analysis results as JSON for programmatic consumption.
//
// Parameters:
//   - analysis: The analysis results to print
//
// Returns:
//   - error: Any error encountered during printing
func PrintJSON(analysis *OrphanedFilesAnalysis) error {
	// Sort files for consistent output
	sortedFiles := make([]string, len(analysis.OrphanedFiles))
	copy(sortedFiles, analysis.OrphanedFiles)
	sort.Strings(sortedFiles)
	analysis.OrphanedFiles = sortedFiles

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(analysis); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}
	return nil
}

// PrintPathsOnly prints only the file paths, one per line.
//
// This function is useful for piping to other commands or scripts.
//
// Parameters:
//   - analysis: The analysis results to print
//
// Returns:
//   - error: Any error encountered during printing
func PrintPathsOnly(analysis *OrphanedFilesAnalysis) error {
	// Sort files for consistent output
	sortedFiles := make([]string, len(analysis.OrphanedFiles))
	copy(sortedFiles, analysis.OrphanedFiles)
	sort.Strings(sortedFiles)

	for _, file := range sortedFiles {
		fmt.Println(file)
	}
	return nil
}

