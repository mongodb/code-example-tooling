// Package includes provides functionality for analyzing include relationships.
//
// This package implements the "analyze includes" subcommand, which analyzes RST files
// to understand their include directive relationships. It can display results as:
//   - A hierarchical tree structure showing include relationships
//   - A flat list of all files referenced through includes
//
// This helps writers understand the impact of changes to files that are widely included
// across the documentation.
package includes

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewIncludesCommand creates the includes subcommand.
//
// This command analyzes include directive relationships in RST files.
// Supports flags for different output formats (tree or list).
//
// Flags:
//   - --tree: Display results as a hierarchical tree structure
//   - --list: Display results as a flat list of all files
//   - -v, --verbose: Show detailed processing information
func NewIncludesCommand() *cobra.Command {
	var (
		showTree bool
		showList bool
		verbose  bool
	)

	cmd := &cobra.Command{
		Use:   "includes [filepath]",
		Short: "Analyze include relationships in RST files",
		Long: `Analyze include directive relationships to understand file dependencies.

This command recursively follows .. include:: directives and shows all files
that are referenced. This helps writers understand the impact of changes to
files that are widely included across the documentation.

Output formats:
  --tree: Show hierarchical tree structure of includes
  --list: Show flat list of all included files

If neither flag is specified, shows a summary with basic statistics.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			filePath := args[0]
			return runAnalyze(filePath, showTree, showList, verbose)
		},
	}

	cmd.Flags().BoolVar(&showTree, "tree", false, "Display results as a hierarchical tree structure")
	cmd.Flags().BoolVar(&showList, "list", false, "Display results as a flat list of all files")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed processing information")

	return cmd
}

// runAnalyze executes the include analysis operation.
//
// This function analyzes the file's include relationships and displays
// the results according to the specified flags.
//
// Parameters:
//   - filePath: Path to the RST file to analyze
//   - showTree: If true, display tree structure
//   - showList: If true, display flat list
//   - verbose: If true, show detailed processing information
//
// Returns:
//   - error: Any error encountered during analysis
func runAnalyze(filePath string, showTree bool, showList bool, verbose bool) error {
	// Perform the analysis
	analysis, err := AnalyzeIncludes(filePath, verbose)
	if err != nil {
		return fmt.Errorf("failed to analyze includes: %w", err)
	}

	// Display results based on flags
	if showTree && showList {
		// Both flags specified - show both outputs
		PrintTree(analysis)
		fmt.Println()
		PrintList(analysis)
	} else if showTree {
		// Only tree
		PrintTree(analysis)
	} else if showList {
		// Only list
		PrintList(analysis)
	} else {
		// Neither flag - show summary
		PrintSummary(analysis)
	}

	return nil
}

