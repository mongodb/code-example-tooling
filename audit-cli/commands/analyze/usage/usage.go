// Package usage provides functionality for analyzing which files reference a target file.
//
// This package implements the "analyze usage" subcommand, which finds all files
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
package usage

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewUsageCommand creates the usage subcommand.
//
// This command analyzes which files reference a given target file through
// RST directives (include, literalinclude, io-code-block, toctree).
//
// Usage:
//   analyze usage /path/to/file.rst
//   analyze usage /path/to/code-example.js
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
//   - -r, --recursive: Recursively follow usage tree until reaching only .txt files (documentation pages)
func NewUsageCommand() *cobra.Command {
	var (
		format         string
		verbose        bool
		countOnly      bool
		pathsOnly      bool
		summaryOnly    bool
		directiveType  string
		includeToctree bool
		excludePattern string
		recursive      bool
	)

	cmd := &cobra.Command{
		Use:   "usage [filepath]",
		Short: "Find all files that use a target file",
		Long: `Find all files that use a target file through RST directives.

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
  # Find what uses an include file
  analyze usage /path/to/includes/fact.rst

  # Find what uses a code example
  analyze usage /path/to/code-examples/example.js

  # Include toctree references (navigation links)
  analyze usage /path/to/file.rst --include-toctree

  # Get JSON output
  analyze usage /path/to/file.rst --format json

  # Show detailed information with line numbers
  analyze usage /path/to/file.rst --verbose

  # Just show the count
  analyze usage /path/to/file.rst --count-only

  # Just show the file paths
  analyze usage /path/to/file.rst --paths-only

  # Show summary statistics only
  analyze usage /path/to/file.rst --summary

  # Exclude certain paths from search
  analyze usage /path/to/file.rst --exclude "*/archive/*"

  # Filter by directive type
  analyze usage /path/to/file.rst --directive-type include

  # Recursively follow usage tree to find all .txt documentation pages
  analyze usage /path/to/includes/fact.rst --recursive`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUsage(args[0], format, verbose, countOnly, pathsOnly, summaryOnly, directiveType, includeToctree, excludePattern, recursive)
		},
	}

	cmd.Flags().StringVar(&format, "format", "text", "Output format (text or json)")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed information including line numbers")
	cmd.Flags().BoolVarP(&countOnly, "count-only", "c", false, "Only show the count of usages")
	cmd.Flags().BoolVar(&pathsOnly, "paths-only", false, "Only show the file paths (one per line)")
	cmd.Flags().BoolVar(&summaryOnly, "summary", false, "Only show summary statistics (total files and usages by type)")
	cmd.Flags().StringVarP(&directiveType, "directive-type", "t", "", "Filter by directive type (include, literalinclude, io-code-block, toctree)")
	cmd.Flags().BoolVar(&includeToctree, "include-toctree", false, "Include toctree entries (navigation links) in addition to content inclusion directives")
	cmd.Flags().StringVar(&excludePattern, "exclude", "", "Exclude paths matching this glob pattern (e.g., '*/archive/*' or '*/deprecated/*')")
	cmd.Flags().BoolVarP(&recursive, "recursive", "r", false, "Recursively follow usage tree until reaching only .txt files (documentation pages)")

	return cmd
}

// runUsage executes the usage analysis.
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
//   - recursive: If true, recursively follow usage tree until reaching only .txt files
//
// Returns:
//   - error: Any error encountered during analysis
func runUsage(targetFile, format string, verbose, countOnly, pathsOnly, summaryOnly bool, directiveType string, includeToctree bool, excludePattern string, recursive bool) error {
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
	var analysis *UsageAnalysis
	var err error

	if recursive {
		// Perform recursive analysis to find all .txt files
		analysis, err = AnalyzeUsageRecursive(targetFile, includeToctree, verbose, excludePattern)
	} else {
		// Perform standard single-level analysis
		analysis, err = AnalyzeUsage(targetFile, includeToctree, verbose, excludePattern)
	}

	if err != nil {
		return fmt.Errorf("failed to analyze usage: %w", err)
	}

	// Filter by directive type if specified
	if directiveType != "" {
		analysis = FilterByDirectiveType(analysis, directiveType)
	}

	// Handle count-only output
	if countOnly {
		fmt.Println(analysis.TotalUsages)
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
	return PrintAnalysis(analysis, outputFormat, verbose, recursive)
}

