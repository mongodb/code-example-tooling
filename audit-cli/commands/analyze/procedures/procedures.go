// Package procedures provides functionality for analyzing procedures in RST files.
//
// This package implements the "analyze procedures" subcommand, which parses
// reStructuredText files and analyzes procedure variations, providing statistics
// and details about:
//   - Number of procedures and variations
//   - Implementation types (procedure directive vs ordered list)
//   - Step counts
//   - Sub-procedure detection
//   - Variation listings (composable tutorial selections and tabids)
package procedures

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// NewProceduresCommand creates the procedures subcommand for analysis.
//
// This command analyzes procedures in RST files and outputs statistics and details
// based on the specified flags:
//   - Default: Count of procedure variations
//   - --list: List all variations with their selection/tabid values
//   - --implementation: Show how each procedure is implemented
//   - --sub-procedures: Indicate if procedures contain nested sub-procedures
//   - --step-count: Show step count for each procedure
func NewProceduresCommand() *cobra.Command {
	var (
		listAll        bool
		listSummary    bool
		implementation bool
		subProcedures  bool
		stepCount      bool
	)

	cmd := &cobra.Command{
		Use:   "procedures [filepath]",
		Short: "Analyze procedure variations in reStructuredText files",
		Long: `Analyze procedure variations in reStructuredText files.

This command parses procedures from RST files and provides analysis including:
  - Total count of procedures and variations
  - Implementation types (procedure directive vs ordered list)
  - Step counts for each procedure
  - Detection of sub-procedures (ordered lists within steps)
  - Listing of all variations (composable tutorial selections and tabids)

By default, outputs a summary count. Use flags to get more detailed information.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			filePath := args[0]

			options := OutputOptions{
				ListAll:        listAll,
				ListSummary:    listSummary,
				Implementation: implementation,
				SubProcedures:  subProcedures,
				StepCount:      stepCount,
			}

			return runAnalyze(filePath, options)
		},
	}

	cmd.Flags().BoolVar(&listAll, "list-all", false, "List all procedures with full selection details")
	cmd.Flags().BoolVar(&listSummary, "list-summary", false, "List procedures grouped by heading without selection details")
	cmd.Flags().BoolVar(&implementation, "implementation", false, "Show how each procedure is implemented")
	cmd.Flags().BoolVar(&subProcedures, "sub-procedures", false, "Indicate if procedures contain nested sub-procedures")
	cmd.Flags().BoolVar(&stepCount, "step-count", false, "Show step count for each procedure")

	return cmd
}

// runAnalyze executes the analysis operation.
func runAnalyze(filePath string, options OutputOptions) error {
	// Verify the file exists
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("failed to access path %s: %w", filePath, err)
	}

	if fileInfo.IsDir() {
		return fmt.Errorf("path %s is a directory; please specify a file", filePath)
	}

	// Analyze the file
	report, err := AnalyzeFile(filePath)
	if err != nil {
		return err
	}

	if report.TotalProcedures == 0 {
		fmt.Println("No procedures found in the file.")
		return nil
	}

	// Print the report
	PrintReport(report, options)

	return nil
}

