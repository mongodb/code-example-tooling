// Package references provides functionality for analyzing which files reference a target file.
//
// This package implements the "analyze references" subcommand, which finds all files
// that reference a given file through RST directives (include, literalinclude, io-code-block).
//
// The command performs reverse dependency analysis, showing which files depend on the
// target file. This is useful for:
//   - Understanding the impact of changes to a file
//   - Finding all usages of an include file
//   - Tracking code example references
//   - Identifying orphaned files (files with no references)
package references

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewReferencesCommand creates the references subcommand.
//
// This command analyzes which files reference a given target file through
// RST directives (include, literalinclude, io-code-block).
//
// Usage:
//   analyze references /path/to/file.rst
//   analyze references /path/to/code-example.js
//
// Flags:
//   - --format: Output format (text or json)
//   - -v, --verbose: Show detailed information including line numbers
//   - -c, --count-only: Only show the count of references
//   - --paths-only: Only show the file paths
//   - -t, --directive-type: Filter by directive type (include, literalinclude, io-code-block)
func NewReferencesCommand() *cobra.Command {
	var (
		format        string
		verbose       bool
		countOnly     bool
		pathsOnly     bool
		directiveType string
	)

	cmd := &cobra.Command{
		Use:   "references [filepath]",
		Short: "Find all files that reference a target file",
		Long: `Find all files that reference a target file through RST directives.

This command performs reverse dependency analysis, showing which files reference
the target file through include, literalinclude, or io-code-block directives.

Supported directive types:
  - .. include::         RST content includes
  - .. literalinclude::  Code file references
  - .. io-code-block::   Input/output examples with file arguments

The command searches all RST files in the source directory tree and identifies
files that reference the target file. This is useful for:
  - Understanding the impact of changes to a file
  - Finding all usages of an include file
  - Tracking code example references
  - Identifying orphaned files (files with no references)

Examples:
  # Find what references an include file
  analyze references /path/to/includes/fact.rst

  # Find what references a code example
  analyze references /path/to/code-examples/example.js

  # Get JSON output
  analyze references /path/to/file.rst --format json

  # Show detailed information with line numbers
  analyze references /path/to/file.rst --verbose

  # Just show the count
  analyze references /path/to/file.rst --count-only

  # Just show the file paths
  analyze references /path/to/file.rst --paths-only

  # Filter by directive type
  analyze references /path/to/file.rst --directive-type include`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runReferences(args[0], format, verbose, countOnly, pathsOnly, directiveType)
		},
	}

	cmd.Flags().StringVar(&format, "format", "text", "Output format (text or json)")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed information including line numbers")
	cmd.Flags().BoolVarP(&countOnly, "count-only", "c", false, "Only show the count of references")
	cmd.Flags().BoolVar(&pathsOnly, "paths-only", false, "Only show the file paths (one per line)")
	cmd.Flags().StringVarP(&directiveType, "directive-type", "t", "", "Filter by directive type (include, literalinclude, io-code-block)")

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
//   - directiveType: Filter by directive type (empty string means all types)
//
// Returns:
//   - error: Any error encountered during analysis
func runReferences(targetFile, format string, verbose, countOnly, pathsOnly bool, directiveType string) error {
	// Validate directive type if specified
	if directiveType != "" {
		validTypes := map[string]bool{
			"include":         true,
			"literalinclude":  true,
			"io-code-block":   true,
		}
		if !validTypes[directiveType] {
			return fmt.Errorf("invalid directive type: %s (must be 'include', 'literalinclude', or 'io-code-block')", directiveType)
		}
	}

	// Validate format
	outputFormat := OutputFormat(format)
	if outputFormat != FormatText && outputFormat != FormatJSON {
		return fmt.Errorf("invalid format: %s (must be 'text' or 'json')", format)
	}

	// Validate flag combinations
	if countOnly && pathsOnly {
		return fmt.Errorf("cannot use --count-only and --paths-only together")
	}
	if (countOnly || pathsOnly) && outputFormat == FormatJSON {
		return fmt.Errorf("--count-only and --paths-only are not compatible with --format json")
	}

	// Perform analysis
	analysis, err := AnalyzeReferences(targetFile)
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

	// Print full results
	return PrintAnalysis(analysis, outputFormat, verbose)
}

