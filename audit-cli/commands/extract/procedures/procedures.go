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
//
//	{heading}-{selection}.rst
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
		selection         string
		outputDir         string
		dryRun            bool
		verbose           bool
		expandIncludes    bool
		showSteps         bool
		showSubProcedures bool
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
			return runExtract(filePath, selection, outputDir, dryRun, verbose, expandIncludes, showSteps, showSubProcedures)
		},
	}

	cmd.Flags().StringVar(&selection, "selection", "", "Extract only a specific variation (by selection or tabid)")
	cmd.Flags().StringVarP(&outputDir, "output", "o", "./output", "Output directory for procedure files")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be extracted without writing files")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Provide additional information during execution")
	cmd.Flags().BoolVar(&expandIncludes, "expand-includes", false, "Expand include directives inline instead of preserving them")
	cmd.Flags().BoolVar(&showSteps, "show-steps", false, "Show detailed information about each step in the procedure")
	cmd.Flags().BoolVar(&showSubProcedures, "show-sub-procedures", false, "Show information about detected sub-procedures within steps")

	return cmd
}

// runExtract executes the extraction operation.
func runExtract(filePath string, selection string, outputDir string, dryRun bool, verbose bool, expandIncludes bool, showSteps bool, showSubProcedures bool) error {
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
		fmt.Printf("\nFound %d unique procedures:\n", len(variations))
		for i, v := range variations {
			fmt.Printf("\n%d. %s\n", i+1, v.Procedure.Title)
			fmt.Printf("   Output file: %s\n", v.OutputFile)
			fmt.Printf("   Steps: %d\n", len(v.Procedure.Steps))

			if v.VariationName != "" {
				// Split the selections and format as a list
				selections := strings.Split(v.VariationName, "; ")
				fmt.Printf("   Appears in %d selections:\n", len(selections))
				for _, sel := range selections {
					fmt.Printf("     - %s\n", sel)
				}
			} else {
				fmt.Printf("   Appears in: (no specific selections)\n")
			}

			// Show step details if requested
			if showSteps {
				fmt.Printf("\n   Step Details:\n")
				for stepIdx, step := range v.Procedure.Steps {
					// Check if the title already contains numbering
					hasNumbering := false
					title := step.Title
					if len(title) > 0 {
						// Check for numbered (1., 2., etc.) or lettered (a., b., etc.) prefix
						if (title[0] >= '0' && title[0] <= '9') || (title[0] >= 'a' && title[0] <= 'z') {
							if len(title) > 1 && title[1] == '.' {
								hasNumbering = true
							}
						}
					}

					if hasNumbering {
						fmt.Printf("   - %s\n", title)
					} else {
						fmt.Printf("   %d. %s\n", stepIdx+1, title)
					}

					if len(step.SubProcedures) > 0 {
						totalSubSteps := 0
						for _, subProc := range step.SubProcedures {
							totalSubSteps += len(subProc.Steps)
						}
						fmt.Printf("      Contains %d sub-procedures with a total of %d sub-steps\n", len(step.SubProcedures), totalSubSteps)
					}
					if len(step.Variations) > 0 {
						fmt.Printf("      Contains %d variations\n", len(step.Variations))
					}
				}
			}

			// Show sub-procedure information if requested
			if showSubProcedures && v.Procedure.HasSubSteps {
				fmt.Printf("\n   Sub-Procedures:\n")
				for stepIdx, step := range v.Procedure.Steps {
					if len(step.SubProcedures) > 0 {
						totalSubSteps := 0
						for _, subProc := range step.SubProcedures {
							totalSubSteps += len(subProc.Steps)
						}
						fmt.Printf("   Step %d (%s) contains %d sub-procedures with a total of %d sub-steps\n",
							stepIdx+1, step.Title, len(step.SubProcedures), totalSubSteps)

						for subProcIdx, subProc := range step.SubProcedures {
							fmt.Printf("\n      Sub-procedure %d (%d steps):\n", subProcIdx+1, len(subProc.Steps))
							for subStepIdx, subStep := range subProc.Steps {
								// Use the appropriate marker based on list type
								marker := ""
								if subProc.ListType == "lettered" {
									// Convert index to letter (0->a, 1->b, etc.)
									marker = string(rune('a' + subStepIdx))
								} else {
									// Default to numbered
									marker = fmt.Sprintf("%d", subStepIdx+1)
								}

								// Strip any existing marker from the title
								title := subStep.Title
								// Check if title starts with a marker (e.g., "a. ", "b. ", "1. ", "2. ")
								if len(title) > 2 && title[1] == '.' && title[2] == ' ' {
									// Check if it's a letter or number marker
									if (title[0] >= 'a' && title[0] <= 'z') || (title[0] >= '0' && title[0] <= '9') {
										title = title[3:] // Strip the marker
									}
								}

								fmt.Printf("         %s. %s\n", marker, title)
							}
						}
					}
				}
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
		fmt.Printf("Dry run complete. Would have written %d files to %s\n", len(variations), outputDir)
	} else {
		fmt.Printf("Successfully extracted %d unique procedures to %s\n", filesWritten, outputDir)
	}

	return nil
}
