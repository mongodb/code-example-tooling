package usage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// OutputFormat represents the output format for the analysis results.
type OutputFormat string

const (
	// FormatText is the default human-readable text format
	FormatText OutputFormat = "text"
	// FormatJSON is the JSON format
	FormatJSON OutputFormat = "json"
)

// PrintAnalysis prints the analysis results in the specified format.
//
// Parameters:
//   - analysis: The analysis results to print
//   - format: The output format (text or json)
//   - verbose: If true, show additional details
//   - recursive: If true, indicates recursive mode was used
func PrintAnalysis(analysis *UsageAnalysis, format OutputFormat, verbose bool, recursive bool) error {
	switch format {
	case FormatJSON:
		return printJSON(analysis)
	case FormatText:
		printText(analysis, verbose, recursive)
		return nil
	default:
		return fmt.Errorf("unknown output format: %s", format)
	}
}

// printText prints the analysis results in human-readable text format.
func printText(analysis *UsageAnalysis, verbose bool, recursive bool) {
	fmt.Println("============================================================")
	if recursive {
		fmt.Println("RECURSIVE USAGE ANALYSIS")
	} else {
		fmt.Println("USAGE ANALYSIS")
	}
	fmt.Println("============================================================")
	fmt.Printf("Target File: %s\n", analysis.TargetFile)
	if recursive {
		fmt.Printf("Total .txt Files: %d\n", analysis.TotalFiles)
		fmt.Println("(Showing only .txt documentation pages)")
	} else {
		fmt.Printf("Total Files: %d\n", analysis.TotalFiles)
		fmt.Printf("Total Usages: %d\n", analysis.TotalUsages)
	}
	fmt.Println("============================================================")
	fmt.Println()

	if analysis.TotalUsages == 0 {
		if recursive {
			fmt.Println("No .txt files ultimately use this file.")
			fmt.Println()
			fmt.Println("This could mean:")
			fmt.Println("  - The file is only used by other include files, not by any .txt pages")
			fmt.Println("  - The file might be orphaned (not used)")
			fmt.Println("  - The file is used with a different path")
		} else {
			fmt.Println("No files use this file.")
			fmt.Println()
			fmt.Println("This could mean:")
			fmt.Println("  - The file is not included in any documentation pages")
			fmt.Println("  - The file might be orphaned (not used)")
			fmt.Println("  - The file is used with a different path")
		}
		fmt.Println()
		fmt.Println("Note: By default, only content inclusion directives are searched.")
		fmt.Println("Use --include-toctree to also search for toctree navigation links.")
		fmt.Println()
		return
	}

	// In recursive mode, skip the directive type breakdown since we only show .txt files
	if !recursive {
		// Group usages by directive type
		byDirectiveType := groupByDirectiveType(analysis.UsingFiles)

		// Print breakdown by directive type with file and reference counts
		directiveTypes := []string{"include", "literalinclude", "io-code-block", "toctree"}
		for _, directiveType := range directiveTypes {
			if refs, ok := byDirectiveType[directiveType]; ok {
				uniqueFiles := countUniqueFiles(refs)
				totalRefs := len(refs)
				if uniqueFiles == totalRefs {
					// No duplicates - just show count
					fmt.Printf("%-20s: %d\n", directiveType, uniqueFiles)
				} else {
					// Has duplicates - show both counts
					if uniqueFiles == 1 {
						fmt.Printf("%-20s: %d file, %d usages\n", directiveType, uniqueFiles, totalRefs)
					} else {
						fmt.Printf("%-20s: %d files, %d usages\n", directiveType, uniqueFiles, totalRefs)
					}
				}
			}
		}
		fmt.Println()
	}

	// Group usages by file
	grouped := GroupUsagesByFile(analysis.UsingFiles)

	// Print detailed list of files using the target
	for i, group := range grouped {
		// Get relative path from source directory for cleaner output
		relPath, err := filepath.Rel(analysis.SourceDir, group.FilePath)
		if err != nil {
			relPath = group.FilePath
		}

		if recursive {
			// In recursive mode, just show the .txt file paths
			fmt.Printf("%3d. %s\n", i+1, relPath)
		} else {
			// Print file path with directive type label
			if group.Count > 1 {
				// Multiple usages from this file
				fmt.Printf("%3d. [%s] %s (%d usages)\n", i+1, group.DirectiveType, relPath, group.Count)
			} else {
				// Single usage
				fmt.Printf("%3d. [%s] %s\n", i+1, group.DirectiveType, relPath)
			}

			// Print line numbers in verbose mode
			if verbose {
				for _, usage := range group.Usages {
					fmt.Printf("     Line %d: %s\n", usage.LineNumber, usage.UsagePath)
				}
			}
		}
	}

	fmt.Println()
}

// printJSON prints the analysis results in JSON format.
func printJSON(analysis *UsageAnalysis) error {
	// Create a JSON-friendly structure
	output := struct {
		TargetFile  string      `json:"target_file"`
		SourceDir   string      `json:"source_dir"`
		TotalFiles  int         `json:"total_files"`
		TotalUsages int         `json:"total_usages"`
		UsingFiles  []FileUsage `json:"using_files"`
	}{
		TargetFile:  analysis.TargetFile,
		SourceDir:   analysis.SourceDir,
		TotalFiles:  analysis.TotalFiles,
		TotalUsages: analysis.TotalUsages,
		UsingFiles:  analysis.UsingFiles,
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

// groupByDirectiveType groups usages by their directive type.
func groupByDirectiveType(usages []FileUsage) map[string][]FileUsage {
	groups := make(map[string][]FileUsage)

	for _, usage := range usages {
		groups[usage.DirectiveType] = append(groups[usage.DirectiveType], usage)
	}

	return groups
}

// FormatReferencePath formats a reference path for display.
//
// This function shortens paths for better readability while maintaining
// enough context to identify the file.
func FormatReferencePath(path, sourceDir string) string {
	// Try to get relative path from source directory
	relPath, err := filepath.Rel(sourceDir, path)
	if err != nil {
		return path
	}

	// If the relative path is shorter, use it
	if len(relPath) < len(path) {
		return relPath
	}

	return path
}

// GetDirectiveTypeLabel returns a human-readable label for a directive type.
func GetDirectiveTypeLabel(directiveType string) string {
	labels := map[string]string{
		"include":         "Include",
		"literalinclude":  "Literal Include",
		"io-code-block":   "I/O Code Block",
	}

	if label, ok := labels[directiveType]; ok {
		return label
	}

	return strings.Title(directiveType)
}

// PrintPathsOnly prints only the file paths, one per line.
//
// This is useful for piping to other commands or for simple scripting.
//
// Parameters:
//   - analysis: The analysis results
//
// Returns:
//   - error: Any error encountered during printing
func PrintPathsOnly(analysis *UsageAnalysis) error {
	// Get unique file paths (in case there are duplicates)
	seen := make(map[string]bool)
	var paths []string

	for _, usage := range analysis.UsingFiles {
		// Get relative path from source directory for cleaner output
		relPath, err := filepath.Rel(analysis.SourceDir, usage.FilePath)
		if err != nil {
			relPath = usage.FilePath
		}

		if !seen[relPath] {
			seen[relPath] = true
			paths = append(paths, relPath)
		}
	}

	// Sort for consistent output
	sort.Strings(paths)

	// Print each path
	for _, path := range paths {
		fmt.Println(path)
	}

	return nil
}

// PrintSummary prints only summary statistics without the file list.
//
// This is useful for getting a quick overview of usage counts.
//
// Parameters:
//   - analysis: The analysis results
//
// Returns:
//   - error: Any error encountered during printing
func PrintSummary(analysis *UsageAnalysis) error {
	fmt.Printf("Total Files: %d\n", analysis.TotalFiles)
	fmt.Printf("Total Usages: %d\n", analysis.TotalUsages)

	if analysis.TotalUsages > 0 {
		// Group by directive type
		byDirectiveType := groupByDirectiveType(analysis.UsingFiles)

		// Print breakdown by type
		fmt.Println("\nBy Type:")
		directiveTypes := []string{"include", "literalinclude", "io-code-block", "toctree"}
		for _, directiveType := range directiveTypes {
			if usages, ok := byDirectiveType[directiveType]; ok {
				uniqueFiles := countUniqueFiles(usages)
				totalUsages := len(usages)
				fmt.Printf("  %-20s: %d files, %d usages\n", directiveType, uniqueFiles, totalUsages)
			}
		}
	}

	return nil
}

