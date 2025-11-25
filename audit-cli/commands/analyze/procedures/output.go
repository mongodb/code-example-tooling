package procedures

import (
	"fmt"
	"strings"
)

// OutputOptions controls what information is displayed in the output.
type OutputOptions struct {
	ListAll        bool // List all variations with their selection/tabid values
	ListSummary    bool // List procedures grouped by heading without selection details
	Implementation bool // Show how each procedure is implemented
	SubProcedures  bool // Indicate if procedures contain nested sub-procedures
	StepCount      bool // Show step count for each procedure
}

// PrintReport prints the analysis report to stdout based on the output options.
func PrintReport(report *AnalysisReport, options OutputOptions) {
	// If no special options are set, just print the count
	if !options.ListAll && !options.ListSummary && !options.Implementation && !options.SubProcedures && !options.StepCount {
		printSummary(report)
		return
	}

	// Print detailed report
	printDetailedReport(report, options)
}

// groupProceduresByHeading groups procedures by their heading and returns the groups and order.
func groupProceduresByHeading(procedures []ProcedureAnalysis) (map[string][]ProcedureAnalysis, []string) {
	headingGroups := make(map[string][]ProcedureAnalysis)
	headingOrder := []string{}

	for _, analysis := range procedures {
		heading := analysis.Procedure.Title
		if heading == "" {
			heading = "(Untitled)"
		}

		if _, exists := headingGroups[heading]; !exists {
			headingOrder = append(headingOrder, heading)
		}
		headingGroups[heading] = append(headingGroups[heading], analysis)
	}

	return headingGroups, headingOrder
}

// calculateTotals calculates total unique procedures and appearances from grouped data.
func calculateTotals(headingGroups map[string][]ProcedureAnalysis) (int, int) {
	totalUniqueProcedures := 0
	totalAppearances := 0

	for _, procedures := range headingGroups {
		totalUniqueProcedures += len(procedures)
		for _, proc := range procedures {
			totalAppearances += proc.VariationCount
		}
	}

	return totalUniqueProcedures, totalAppearances
}

// printSummary prints a summary of the analysis.
func printSummary(report *AnalysisReport) {
	fmt.Printf("File: %s\n", report.FilePath)
	fmt.Printf("Total unique procedures: %d\n", len(report.Procedures))
	fmt.Printf("Total procedure appearances: %d\n", report.TotalVariations)
}

// printDetailedReport prints a detailed analysis report.
func printDetailedReport(report *AnalysisReport, options OutputOptions) {
	fmt.Printf("Procedure Analysis for: %s\n", report.FilePath)
	fmt.Println(strings.Repeat("=", 80))

	// Group procedures by heading first to get accurate counts
	headingGroups, headingOrder := groupProceduresByHeading(report.Procedures)
	totalUniqueProcedures, totalAppearances := calculateTotals(headingGroups)

	fmt.Printf("\nTotal unique procedures: %d\n", totalUniqueProcedures)
	fmt.Printf("Total procedure appearances: %d\n\n", totalAppearances)

	// Print implementation type summary if requested
	if options.Implementation {
		fmt.Println("Procedures by implementation type:")
		for implType, count := range report.ProceduresByType {
			fmt.Printf("  - %s: %d\n", implType, count)
		}
		fmt.Println()
	}

	// Print details grouped by heading (headingGroups already created above)
	fmt.Println("Procedures by Heading:")
	fmt.Println(strings.Repeat("-", 80))

	headingNum := 1
	for _, heading := range headingOrder {
		procedures := headingGroups[heading]

		fmt.Printf("\n%d. %s\n", headingNum, heading)
		fmt.Printf("   Unique procedures: %d\n", len(procedures))

		// Calculate total appearances for this heading
		totalAppearances := 0
		for _, proc := range procedures {
			totalAppearances += proc.VariationCount
		}
		fmt.Printf("   Total appearances: %d\n", totalAppearances)

		// If only showing summary, skip the individual procedure details
		if options.ListSummary && !options.ListAll {
			headingNum++
			continue
		}

		// Determine if we need sub-numbering (only when there are multiple unique procedures)
		useSubNumbering := len(procedures) > 1

		// Show each unique procedure under this heading
		for i, analysis := range procedures {
			fmt.Printf("\n   ")

			// Only show sub-numbering if there are multiple unique procedures
			if useSubNumbering {
				fmt.Printf("%d.%d. ", headingNum, i+1)
			}

			// Show the first step to distinguish procedures (only if there are multiple)
			if useSubNumbering {
				if len(analysis.Procedure.Steps) > 0 && analysis.Procedure.Steps[0].Title != "" {
					fmt.Printf("%s\n", analysis.Procedure.Steps[0].Title)
				} else if len(analysis.Procedure.Steps) > 0 {
					fmt.Printf("(Untitled first step)\n")
				} else {
					fmt.Printf("(No steps)\n")
				}
			} else {
				// For single procedures, just show the step count
				fmt.Printf("Steps: %d\n", len(analysis.Procedure.Steps))
			}

			// Indent based on whether we're using sub-numbering
			indent := "        "
			if !useSubNumbering {
				indent = "   "
			}

			// Only show step count if we already showed the first step title
			if useSubNumbering {
				fmt.Printf("%sSteps: %d\n", indent, len(analysis.Procedure.Steps))
			}

			// Print implementation type if requested
			if options.Implementation {
				fmt.Printf("%sImplementation: %s\n", indent, analysis.Implementation)
			}

			// Print sub-procedures flag if requested
			if options.SubProcedures {
				if analysis.HasSubSteps {
					fmt.Printf("%sContains sub-procedures: yes\n", indent)
				} else {
					fmt.Printf("%sContains sub-procedures: no\n", indent)
				}
			}

			// Print selections if requested
			if options.ListAll {
				if analysis.VariationCount == 1 {
					fmt.Printf("%sAppears in 1 selection:\n", indent)
				} else {
					fmt.Printf("%sAppears in %d selections:\n", indent, analysis.VariationCount)
				}

				if len(analysis.Variations) > 0 && analysis.Variations[0] != "(no variations)" {
					for _, variation := range analysis.Variations {
						fmt.Printf("%s  - %s\n", indent, variation)
					}
				} else {
					fmt.Printf("%s  (single variation, no tabs or selections)\n", indent)
				}
			} else if options.ListSummary {
				// For summary, just show the count without listing all selections
				if analysis.VariationCount == 1 {
					fmt.Printf("%sAppears in 1 selection\n", indent)
				} else {
					fmt.Printf("%sAppears in %d selections\n", indent, analysis.VariationCount)
				}
			}
		}

		headingNum++
	}

	fmt.Println()
}

