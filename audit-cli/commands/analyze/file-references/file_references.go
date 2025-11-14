// Package filereferences provides functionality for analyzing which files reference a target file.
//
// This package implements the "analyze file-references" subcommand, which finds all files
// that reference a given file through RST directives (include, literalinclude, io-code-block, toctree).
//
// The command searches both RST files (.rst, .txt) and YAML files (.yaml, .yml) since
// extract and release YAML files contain RST directives within their content blocks.
//
// The command performs reverse dependency analysis, showing which files depend on the
// target file. This is useful for:
//   - Understanding the impact of changes to a file
//   - Finding all usages of an include file
//   - Tracking code example references
package filereferences

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewFileReferencesCommand creates the file-references subcommand.
//
// This command analyzes which files reference a given target file through
// RST directives (include, literalinclude, io-code-block, toctree).
//
// Usage:
//   analyze file-references /path/to/file.rst
//   analyze file-references /path/to/code-example.js
//
// Flags:
//   - --format: Output format (text or json)
//   - -v, --verbose: Show detailed information including line numbers
//   - -c, --count-only: Only show the count of references
//   - --paths-only: Only show the file paths
//   - --summary: Only show summary statistics (total files and references by type)
//   - -t, --directive-type: Filter by directive type (include, literalinclude, io-code-block, toctree)
//   - --include-toctree: Include toctree entries (navigation links) in addition to content inclusion directives
//   - --exclude: Exclude paths matching this glob pattern (e.g., '*/archive/*')
func NewFileReferencesCommand() *cobra.Command {
	var (
		format         string
		verbose        bool
		countOnly      bool
		pathsOnly      bool
		summaryOnly    bool
		directiveType  string
		includeToctree bool
		excludePattern string
	)

	cmd := &cobra.Command{
		Use:   "file-references [filepath]",
		Short: "Find all files that reference a target file",
		Long: `Find all files that reference a target file through RST directives.

This command performs reverse dependency analysis, showing which files reference
the target file through content inclusion directives (include, literalinclude,
io-code-block). Use --include-toctree to also search for toctree entries, which
are navigation links rather than content transclusion.

Supported directive types:
  - .. include::         RST content includes (transcluded)
  - .. literalinclude::  Code file references (transcluded)
  - .. io-code-block::   Input/output examples with file arguments (transcluded)
  - .. toctree::         Table of contents entries (navigation links, requires --include-toctree)

The command searches all RST files (.rst, .txt) and YAML files (.yaml, .yml) in
the source directory tree. YAML files are included because extract and release
files contain RST directives within their content blocks.

This is useful for:
  - Understanding the impact of changes to a file
  - Finding all usages of an include file
  - Tracking code example references

Examples:
  # Find what references an include file
  analyze file-references /path/to/includes/fact.rst

  # Find what references a code example
  analyze file-references /path/to/code-examples/example.js

  # Include toctree references (navigation links)
  analyze file-references /path/to/file.rst --include-toctree

  # Get JSON output
  analyze file-references /path/to/file.rst --format json

  # Show detailed information with line numbers
  analyze file-references /path/to/file.rst --verbose

  # Just show the count
  analyze file-references /path/to/file.rst --count-only

  # Just show the file paths
  analyze file-references /path/to/file.rst --paths-only

  # Show summary statistics only
  analyze file-references /path/to/file.rst --summary

  # Exclude certain paths from search
  analyze file-references /path/to/file.rst --exclude "*/archive/*"

  # Filter by directive type
  analyze file-references /path/to/file.rst --directive-type include`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runReferences(args[0], format, verbose, countOnly, pathsOnly, summaryOnly, directiveType, includeToctree, excludePattern)
		},
	}

	cmd.Flags().StringVar(&format, "format", "text", "Output format (text or json)")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed information including line numbers")
	cmd.Flags().BoolVarP(&countOnly, "count-only", "c", false, "Only show the count of references")
	cmd.Flags().BoolVar(&pathsOnly, "paths-only", false, "Only show the file paths (one per line)")
	cmd.Flags().BoolVar(&summaryOnly, "summary", false, "Only show summary statistics (total files and references by type)")
	cmd.Flags().StringVarP(&directiveType, "directive-type", "t", "", "Filter by directive type (include, literalinclude, io-code-block, toctree)")
	cmd.Flags().BoolVar(&includeToctree, "include-toctree", false, "Include toctree entries (navigation links) in addition to content inclusion directives")
	cmd.Flags().StringVar(&excludePattern, "exclude", "", "Exclude paths matching this glob pattern (e.g., '*/archive/*' or '*/deprecated/*')")

	return cmd
}

// runReferences executes the references analysis.
//
// This function performs the analysis and prints the results in the specified format.
//
// Parameters:
//   - targetFile: Path to the file to analyze
//   - format: Output format (text or json)
//   - verbose: If true, show detailed information
//   - countOnly: If true, only show the count
//   - pathsOnly: If true, only show the file paths
//   - summaryOnly: If true, only show summary statistics
//   - directiveType: Filter by directive type (empty string means all types)
//   - includeToctree: If true, include toctree entries in the search
//   - excludePattern: Glob pattern for paths to exclude (empty string means no exclusion)
//
// Returns:
//   - error: Any error encountered during analysis
func runReferences(targetFile, format string, verbose, countOnly, pathsOnly, summaryOnly bool, directiveType string, includeToctree bool, excludePattern string) error {
	// Validate directive type if specified
	if directiveType != "" {
		validTypes := map[string]bool{
			"include":         true,
			"literalinclude":  true,
			"io-code-block":   true,
			"toctree":         true,
		}
		if !validTypes[directiveType] {
			return fmt.Errorf("invalid directive type: %s (must be 'include', 'literalinclude', 'io-code-block', or 'toctree')", directiveType)
		}
	}

	// Validate format
	outputFormat := OutputFormat(format)
	if outputFormat != FormatText && outputFormat != FormatJSON {
		return fmt.Errorf("invalid format: %s (must be 'text' or 'json')", format)
	}

	// Validate flag combinations
	exclusiveFlags := 0
	if countOnly {
		exclusiveFlags++
	}
	if pathsOnly {
		exclusiveFlags++
	}
	if summaryOnly {
		exclusiveFlags++
	}
	if exclusiveFlags > 1 {
		return fmt.Errorf("cannot use --count-only, --paths-only, and --summary together")
	}
	if (countOnly || pathsOnly || summaryOnly) && outputFormat == FormatJSON {
		return fmt.Errorf("--count-only, --paths-only, and --summary are not compatible with --format json")
	}

	// Perform analysis
	analysis, err := AnalyzeReferences(targetFile, includeToctree, verbose, excludePattern)
	if err != nil {
		return fmt.Errorf("failed to analyze references: %w", err)
	}

	// Filter by directive type if specified
	if directiveType != "" {
		analysis = FilterByDirectiveType(analysis, directiveType)
	}

	// Handle count-only output
	if countOnly {
		fmt.Println(analysis.TotalReferences)
		return nil
	}

	// Handle paths-only output
	if pathsOnly {
		return PrintPathsOnly(analysis)
	}

	// Handle summary-only output
	if summaryOnly {
		return PrintSummary(analysis)
	}

	// Print full results
	return PrintAnalysis(analysis, outputFormat, verbose)
}

