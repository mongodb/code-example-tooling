// Package tested_examples implements the tested-examples subcommand for counting code examples.
package tested_examples

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewTestedExamplesCommand creates the tested-examples subcommand.
//
// This command counts tested code examples in the MongoDB documentation monorepo.
//
// Usage:
//   count tested-examples /path/to/docs-monorepo
//   count tested-examples /path/to/docs-monorepo --for-product pymongo
//   count tested-examples /path/to/docs-monorepo --count-by-product
//   count tested-examples /path/to/docs-monorepo --exclude-output
//
// Flags:
//   - --for-product: Only count code examples for a specific product
//   - --count-by-product: Display a list of products with counts for each
//   - --exclude-output: Only count source files, excluding output files (.txt, .sh)
func NewTestedExamplesCommand() *cobra.Command {
	var (
		forProduct     string
		countByProduct bool
		excludeOutput  bool
	)

	cmd := &cobra.Command{
		Use:   "tested-examples [monorepo-path]",
		Short: "Count tested code examples in the documentation monorepo",
		Long: `Count tested code examples in the MongoDB documentation monorepo.

This command navigates to the content/code-examples/tested directory from the
monorepo root and counts all files recursively.

The tested directory structure has two levels:
  L1: Language directories (command-line, csharp, go, java, javascript, python)
  L2: Product directories (mongosh, driver, atlas-sdk, driver-sync, pymongo, etc.)

By default, returns only a total count of all files.

` + GetProductList() + `

Examples:
  # Get total count of all tested code examples
  count tested-examples /path/to/docs-monorepo

  # Count examples for a specific product
  count tested-examples /path/to/docs-monorepo --for-product pymongo

  # Show counts broken down by product
  count tested-examples /path/to/docs-monorepo --count-by-product

  # Count only source files (exclude .txt and .sh output files)
  count tested-examples /path/to/docs-monorepo --exclude-output

  # Combine flags: count source files for a specific product
  count tested-examples /path/to/docs-monorepo --for-product pymongo --exclude-output`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTestedExamples(args[0], forProduct, countByProduct, excludeOutput)
		},
	}

	cmd.Flags().StringVar(&forProduct, "for-product", "", "Only count code examples for a specific product")
	cmd.Flags().BoolVar(&countByProduct, "count-by-product", false, "Display counts for each product")
	cmd.Flags().BoolVar(&excludeOutput, "exclude-output", false, "Only count source files (exclude .txt and .sh files)")

	return cmd
}

// runTestedExamples executes the tested-examples counting operation.
func runTestedExamples(monorepoPath string, forProduct string, countByProduct bool, excludeOutput bool) error {
	// Validate product if specified
	if forProduct != "" && !IsValidProduct(forProduct) {
		return fmt.Errorf("invalid product: %s\n\n%s", forProduct, GetProductList())
	}

	// Validate flag combinations
	if forProduct != "" && countByProduct {
		return fmt.Errorf("cannot use --for-product and --count-by-product together")
	}

	// Count the files
	result, err := CountTestedExamples(monorepoPath, forProduct, excludeOutput)
	if err != nil {
		return fmt.Errorf("failed to count tested examples: %w", err)
	}

	// Print the results
	PrintResults(result, countByProduct)

	return nil
}

