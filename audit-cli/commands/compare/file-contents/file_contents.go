// Package file_contents provides functionality for comparing file contents across versions.
//
// This package implements the "compare file-contents" subcommand, which compares
// file contents either directly between two files or across multiple versions of
// MongoDB documentation.
//
// The command supports two modes:
//   1. Direct comparison: Compare two specific files
//   2. Version comparison: Compare the same file across multiple versions
//
// Output can be progressively detailed:
//   - Default: Summary of differences
//   - --show-paths: Include file paths
//   - --show-diff: Include unified diffs
package file_contents

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// NewFileContentsCommand creates the file-contents subcommand.
//
// This command compares file contents either directly between two files
// or across multiple versions of documentation.
//
// Usage modes:
//   1. Direct comparison:
//      compare file-contents file1.rst file2.rst
//
//   2. Version comparison:
//      compare file-contents file.rst --product-dir /path/to/product --versions v1,v2,v3
//
// Flags:
//   - -p, --product-dir: Product directory path (required for version comparison)
//   - -V, --versions: Comma-separated list of versions (required for version comparison)
//   - --show-paths: Display file paths of files that differ
//   - -d, --show-diff: Display unified diff output
//   - -v, --verbose: Show detailed processing information
func NewFileContentsCommand() *cobra.Command {
	var (
		productDir string
		versions   string
		showPaths  bool
		showDiff   bool
		verbose    bool
	)

	cmd := &cobra.Command{
		Use:   "file-contents [file1] [file2]",
		Short: "Compare file contents across versions or between two files",
		Long: `Compare file contents to identify differences.

This command supports two modes:

1. Direct comparison (two file arguments):
   Compare two specific files directly.
   Example: compare file-contents file1.rst file2.rst

2. Version comparison (one file argument + flags):
   Compare the same file across multiple documentation versions.
   Example: compare file-contents /path/to/manual/manual/source/file.rst \
            --product-dir /path/to/manual \
            --versions manual,upcoming,v8.1,v8.0

The command provides progressive output detail:
  - Default: Summary of differences
  - --show-paths: Include file paths grouped by status
  - --show-diff: Include unified diffs (implies --show-paths)

Files that don't exist in certain versions are reported separately and
do not cause errors.`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCompare(args, productDir, versions, showPaths, showDiff, verbose)
		},
	}

	cmd.Flags().StringVarP(&productDir, "product-dir", "p", "", "Product directory path (e.g., /path/to/manual)")
	cmd.Flags().StringVarP(&versions, "versions", "V", "", "Comma-separated list of versions (e.g., manual,upcoming,v8.1)")
	cmd.Flags().BoolVar(&showPaths, "show-paths", false, "Display file paths of files that differ")
	cmd.Flags().BoolVarP(&showDiff, "show-diff", "d", false, "Display unified diff output")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed processing information")

	return cmd
}

// runCompare executes the comparison operation.
//
// This function validates arguments and delegates to the appropriate
// comparison function based on the mode (direct or version comparison).
//
// Parameters:
//   - args: Command line arguments (1 or 2 file paths)
//   - productDir: Product directory path (for version comparison)
//   - versions: Comma-separated version list (for version comparison)
//   - showPaths: If true, show file paths
//   - showDiff: If true, show diffs
//   - verbose: If true, show detailed processing information
//
// Returns:
//   - error: Any error encountered during comparison
func runCompare(args []string, productDir, versions string, showPaths, showDiff, verbose bool) error {
	// Validate arguments based on mode
	if len(args) == 2 {
		// Direct comparison mode
		if productDir != "" || versions != "" {
			return fmt.Errorf("--product-dir and --versions cannot be used with two file arguments")
		}
		return runDirectComparison(args[0], args[1], showPaths, showDiff, verbose)
	} else if len(args) == 1 {
		// Version comparison mode
		if productDir == "" {
			return fmt.Errorf("--product-dir is required when comparing versions (use -p or --product-dir)")
		}
		if versions == "" {
			return fmt.Errorf("--versions is required when comparing versions (use -V or --versions)")
		}
		return runVersionComparison(args[0], productDir, versions, showPaths, showDiff, verbose)
	}

	return fmt.Errorf("expected 1 or 2 file arguments")
}

// runDirectComparison performs a direct comparison between two files.
//
// Parameters:
//   - file1: Path to the first file
//   - file2: Path to the second file
//   - showPaths: If true, show file paths
//   - showDiff: If true, show diffs
//   - verbose: If true, show detailed processing information
//
// Returns:
//   - error: Any error encountered during comparison
func runDirectComparison(file1, file2 string, showPaths, showDiff, verbose bool) error {
	result, err := CompareFiles(file1, file2, showDiff, verbose)
	if err != nil {
		return fmt.Errorf("comparison failed: %w", err)
	}

	PrintComparisonResult(result, showPaths, showDiff)
	return nil
}

// runVersionComparison performs a version-based comparison.
//
// Parameters:
//   - referenceFile: Path to the reference file
//   - productDir: Product directory path
//   - versionsStr: Comma-separated version list
//   - showPaths: If true, show file paths
//   - showDiff: If true, show diffs
//   - verbose: If true, show detailed processing information
//
// Returns:
//   - error: Any error encountered during comparison
func runVersionComparison(referenceFile, productDir, versionsStr string, showPaths, showDiff, verbose bool) error {
	// Parse versions
	versionList := parseVersions(versionsStr)
	if len(versionList) == 0 {
		return fmt.Errorf("no versions specified")
	}

	result, err := CompareVersions(referenceFile, productDir, versionList, showDiff, verbose)
	if err != nil {
		return fmt.Errorf("comparison failed: %w", err)
	}

	PrintComparisonResult(result, showPaths, showDiff)
	return nil
}

// parseVersions parses a comma-separated version string into a slice.
//
// This function splits the version string by commas and trims whitespace
// from each version identifier.
//
// Parameters:
//   - versionsStr: Comma-separated version string (e.g., "manual, upcoming, v8.1")
//
// Returns:
//   - []string: List of version identifiers
func parseVersions(versionsStr string) []string {
	parts := strings.Split(versionsStr, ",")
	var versions []string
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			versions = append(versions, trimmed)
		}
	}
	return versions
}

