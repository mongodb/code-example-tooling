// Package procedures provides functionality for extracting procedures from RST files.
//
// This package implements the "extract procedures" subcommand, which parses
// reStructuredText files and extracts procedure variations based on:
//   - Composable tutorial selections
//   - Tab selections (tabids)
//   - Ordered lists
//   - Procedure directives
//
// The extracted procedures are written to individual RST files with standardized naming:
//   {heading}-{selection}.rst
//
// Supports filtering to extract only specific variations using the --selection flag.
package procedures

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// NewProceduresCommand creates the procedures subcommand.
//
// This command extracts procedure variations from RST files and writes them to
// individual files in the output directory. Supports various flags for controlling behavior:
//   - --selection: Extract only a specific variation (by selection or tabid)
//   - -o, --output: Output directory for extracted files
//   - --dry-run: Show what would be extracted without writing files
//   - -v, --verbose: Show detailed processing information
func NewProceduresCommand() *cobra.Command {
	var (
		selection      string
		outputDir      string
		dryRun         bool
		verbose        bool
		expandIncludes bool
	)

	cmd := &cobra.Command{
		Use:   "procedures [filepath]",
		Short: "Extract procedure variations from reStructuredText files",
		Long: `Extract procedure variations from reStructuredText files.

This command parses procedures from RST files and extracts all variations based on:
  - Composable tutorial selections (.. composable-tutorial::)
  - Tab selections (.. tabs:: with :tabid:)
  - Procedure directives (.. procedure::)
  - Ordered lists

Each variation is written to a separate RST file with interpolated content,
showing the procedure as it would be rendered for that specific variation.

The output files are named using the format: {heading}-{selection}.rst
For example: "connect-to-cluster-python.rst", "create-index-drivers.rst"

By default, include directives are preserved in the output. Use --expand-includes
to inline the content of included files.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			filePath := args[0]
			return runExtract(filePath, selection, outputDir, dryRun, verbose, expandIncludes)
		},
	}

	cmd.Flags().StringVar(&selection, "selection", "", "Extract only a specific variation (by selection or tabid)")
	cmd.Flags().StringVarP(&outputDir, "output", "o", "./output", "Output directory for procedure files")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be extracted without writing files")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Provide additional information during execution")
	cmd.Flags().BoolVar(&expandIncludes, "expand-includes", false, "Expand include directives inline instead of preserving them")

	return cmd
}

// runExtract executes the extraction operation.
func runExtract(filePath string, selection string, outputDir string, dryRun bool, verbose bool, expandIncludes bool) error {
	// Verify the file exists
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("failed to access path %s: %w", filePath, err)
	}

	if fileInfo.IsDir() {
		return fmt.Errorf("path %s is a directory; please specify a file", filePath)
	}

	// Parse the file and extract procedure variations
	if verbose {
		fmt.Printf("Parsing procedures from %s\n", filePath)
		if expandIncludes {
			fmt.Println("Expanding include directives inline")
		}
	}

	variations, err := ParseFile(filePath, selection, expandIncludes)
	if err != nil {
		return err
	}

	if len(variations) == 0 {
		fmt.Println("No procedures found in the file.")
		return nil
	}

	// Report what was found
	if verbose || dryRun {
		fmt.Printf("\nFound %d unique procedure(s):\n", len(variations))
		for i, v := range variations {
			fmt.Printf("\n%d. %s\n", i+1, v.Procedure.Title)
			fmt.Printf("   Output file: %s\n", v.OutputFile)
			fmt.Printf("   Steps: %d\n", len(v.Procedure.Steps))

			if v.VariationName != "" {
				// Split the selections and format as a list
				selections := strings.Split(v.VariationName, "; ")
				fmt.Printf("   Appears in %d selection(s):\n", len(selections))
				for _, sel := range selections {
					fmt.Printf("     - %s\n", sel)
				}
			} else {
				fmt.Printf("   Appears in: (no specific selections)\n")
			}
		}
		fmt.Println()
	}

	// Write the variations
	filesWritten, err := WriteAllVariations(variations, outputDir, dryRun, verbose)
	if err != nil {
		return err
	}

	// Print summary
	if dryRun {
		fmt.Printf("Dry run complete. Would have written %d file(s) to %s\n", len(variations), outputDir)
	} else {
		fmt.Printf("Successfully extracted %d unique procedure(s) to %s\n", filesWritten, outputDir)
	}

	return nil
}

