package references

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
func PrintAnalysis(analysis *ReferenceAnalysis, format OutputFormat, verbose bool) error {
	switch format {
	case FormatJSON:
		return printJSON(analysis)
	case FormatText:
		printText(analysis, verbose)
		return nil
	default:
		return fmt.Errorf("unknown output format: %s", format)
	}
}

// printText prints the analysis results in human-readable text format.
func printText(analysis *ReferenceAnalysis, verbose bool) {
	fmt.Println("============================================================")
	fmt.Println("REFERENCE ANALYSIS")
	fmt.Println("============================================================")
	fmt.Printf("Target File: %s\n", analysis.TargetFile)
	fmt.Printf("Total Files: %d\n", analysis.TotalFiles)
	fmt.Printf("Total References: %d\n", analysis.TotalReferences)
	fmt.Println("============================================================")
	fmt.Println()

	if analysis.TotalReferences == 0 {
		fmt.Println("No files reference this file.")
		fmt.Println()
		return
	}

	// Group references by directive type
	byDirectiveType := groupByDirectiveType(analysis.ReferencingFiles)

	// Print breakdown by directive type with file and reference counts
	for _, directiveType := range []string{"include", "literalinclude", "io-code-block"} {
		if refs, ok := byDirectiveType[directiveType]; ok {
			uniqueFiles := countUniqueFiles(refs)
			totalRefs := len(refs)
			if uniqueFiles == totalRefs {
				// No duplicates - just show count
				fmt.Printf("%-20s: %d\n", directiveType, uniqueFiles)
			} else {
				// Has duplicates - show both counts
				if uniqueFiles == 1 {
					fmt.Printf("%-20s: %d file, %d references\n", directiveType, uniqueFiles, totalRefs)
				} else {
					fmt.Printf("%-20s: %d files, %d references\n", directiveType, uniqueFiles, totalRefs)
				}
			}
		}
	}
	fmt.Println()

	// Group references by file
	grouped := GroupReferencesByFile(analysis.ReferencingFiles)

	// Print detailed list of referencing files
	for i, group := range grouped {
		// Get relative path from source directory for cleaner output
		relPath, err := filepath.Rel(analysis.SourceDir, group.FilePath)
		if err != nil {
			relPath = group.FilePath
		}

		// Print file path with directive type label
		if group.Count > 1 {
			// Multiple references from this file
			fmt.Printf("%3d. [%s] %s (%d references)\n", i+1, group.DirectiveType, relPath, group.Count)
		} else {
			// Single reference
			fmt.Printf("%3d. [%s] %s\n", i+1, group.DirectiveType, relPath)
		}

		// Print line numbers in verbose mode
		if verbose {
			for _, ref := range group.References {
				fmt.Printf("     Line %d: %s\n", ref.LineNumber, ref.ReferencePath)
			}
		}
	}

	fmt.Println()
}

// printJSON prints the analysis results in JSON format.
func printJSON(analysis *ReferenceAnalysis) error {
	// Create a JSON-friendly structure
	output := struct {
		TargetFile       string          `json:"target_file"`
		SourceDir        string          `json:"source_dir"`
		TotalFiles       int             `json:"total_files"`
		TotalReferences  int             `json:"total_references"`
		ReferencingFiles []FileReference `json:"referencing_files"`
	}{
		TargetFile:       analysis.TargetFile,
		SourceDir:        analysis.SourceDir,
		TotalFiles:       analysis.TotalFiles,
		TotalReferences:  analysis.TotalReferences,
		ReferencingFiles: analysis.ReferencingFiles,
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

// groupByDirectiveType groups references by their directive type.
func groupByDirectiveType(refs []FileReference) map[string][]FileReference {
	groups := make(map[string][]FileReference)

	for _, ref := range refs {
		groups[ref.DirectiveType] = append(groups[ref.DirectiveType], ref)
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
func PrintPathsOnly(analysis *ReferenceAnalysis) error {
	// Get unique file paths (in case there are duplicates)
	seen := make(map[string]bool)
	var paths []string

	for _, ref := range analysis.ReferencingFiles {
		// Get relative path from source directory for cleaner output
		relPath, err := filepath.Rel(analysis.SourceDir, ref.FilePath)
		if err != nil {
			relPath = ref.FilePath
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

