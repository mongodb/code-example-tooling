// Package orphanedfiles provides functionality for finding orphaned files in documentation.
//
// This package implements the "analyze orphaned-files" subcommand, which identifies files
// that have no incoming references from other files. An orphaned file is one that is not
// referenced by any include, literalinclude, io-code-block, or toctree directive.
//
// The command searches through all RST files (.rst, .txt) and YAML files (.yaml, .yml)
// in a source directory to build a complete reference map, then identifies files with
// zero incoming references.
//
// This is useful for:
//   - Finding unused include files that can be removed
//   - Identifying documentation pages not linked in the navigation
//   - Cleaning up legacy content
//   - Maintaining documentation hygiene
package orphanedfiles

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewOrphanedFilesCommand creates the orphaned-files subcommand.
//
// This command finds files that have no incoming references from other files.
// It scans all RST and YAML files in the source directory to build a complete
// reference map, then identifies files with zero references.
//
// Usage:
//   analyze orphaned-files /path/to/source/directory
//
// Flags:
//   - --format: Output format (text or json)
//   - -v, --verbose: Show detailed information during scanning
//   - --paths-only: Only show the file paths (one per line)
//   - --count-only: Only show the count of orphaned files
//   - --include-toctree: Include toctree references when determining orphaned status
//   - --exclude: Exclude paths matching this glob pattern
func NewOrphanedFilesCommand() *cobra.Command {
	var (
		format         string
		verbose        bool
		pathsOnly      bool
		countOnly      bool
		includeToctree bool
		excludePattern string
	)

	cmd := &cobra.Command{
		Use:   "orphaned-files [source-directory]",
		Short: "Find files with no incoming references",
		Long: `Find files that have no incoming references from other files.

This command scans all RST and YAML files in the source directory to identify
files that are not referenced by any include, literalinclude, io-code-block,
or toctree directive.

By default, only content inclusion directives are considered. Use --include-toctree
to also consider toctree entries (navigation links) when determining orphaned status.

An orphaned file might be:
  - An unused include file that can be removed
  - A documentation page not linked in the navigation
  - Legacy content that needs cleanup
  - A file that should be referenced but isn't

Examples:
  # Find orphaned files in a source directory
  analyze orphaned-files /path/to/source

  # Include toctree references (consider navigation links)
  analyze orphaned-files /path/to/source --include-toctree

  # Get JSON output
  analyze orphaned-files /path/to/source --format json

  # Show detailed scanning progress
  analyze orphaned-files /path/to/source --verbose

  # Just show the count
  analyze orphaned-files /path/to/source --count-only

  # Just show the file paths
  analyze orphaned-files /path/to/source --paths-only

  # Exclude certain paths from analysis
  analyze orphaned-files /path/to/source --exclude "*/archive/*"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runOrphanedFiles(args[0], format, verbose, pathsOnly, countOnly, includeToctree, excludePattern)
		},
	}

	cmd.Flags().StringVar(&format, "format", "text", "Output format (text or json)")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed information during scanning")
	cmd.Flags().BoolVar(&pathsOnly, "paths-only", false, "Only show the file paths (one per line)")
	cmd.Flags().BoolVarP(&countOnly, "count-only", "c", false, "Only show the count of orphaned files")
	cmd.Flags().BoolVar(&includeToctree, "include-toctree", false, "Include toctree references when determining orphaned status")
	cmd.Flags().StringVar(&excludePattern, "exclude", "", "Exclude paths matching this glob pattern (e.g., '*/archive/*')")

	return cmd
}

// runOrphanedFiles executes the orphaned files analysis.
//
// This function performs the analysis and prints the results in the specified format.
//
// Parameters:
//   - sourceDir: Path to the source directory to analyze
//   - format: Output format (text or json)
//   - verbose: If true, show detailed information during scanning
//   - pathsOnly: If true, only show the file paths
//   - countOnly: If true, only show the count
//   - includeToctree: If true, include toctree references
//   - excludePattern: Glob pattern for paths to exclude (empty string means no exclusion)
//
// Returns:
//   - error: Any error encountered during analysis
func runOrphanedFiles(sourceDir, format string, verbose, pathsOnly, countOnly, includeToctree bool, excludePattern string) error {
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
	analysis, err := FindOrphanedFiles(sourceDir, includeToctree, verbose, excludePattern)
	if err != nil {
		return fmt.Errorf("failed to find orphaned files: %w", err)
	}

	// Handle count-only output
	if countOnly {
		fmt.Println(analysis.TotalOrphaned)
		return nil
	}

	// Handle paths-only output
	if pathsOnly {
		return PrintPathsOnly(analysis)
	}

	// Handle regular output
	if outputFormat == FormatJSON {
		return PrintJSON(analysis)
	}

	return PrintText(analysis, verbose)
}

