package procedures

import (
	"fmt"

	"github.com/mongodb/code-example-tooling/audit-cli/internal/rst"
)

// AnalyzeFile analyzes procedures in a file and returns a report.
//
// This function parses all procedures from the file and generates analysis
// information including variation counts, step counts, implementation types,
// and sub-procedure detection.
//
// This function expands include directives to properly detect variations that
// may be defined in included files.
//
// Parameters:
//   - filePath: Path to the RST file to analyze
//
// Returns:
//   - *AnalysisReport: Analysis report containing all procedure information
//   - error: Any error encountered during analysis
func AnalyzeFile(filePath string) (*AnalysisReport, error) {
	return AnalyzeFileWithOptions(filePath, true)
}

// AnalyzeFileWithOptions analyzes procedures in a file with options and returns a report.
//
// This function parses all procedures from the file and generates analysis
// information including variation counts, step counts, implementation types,
// and sub-procedure detection.
//
// Parameters:
//   - filePath: Path to the RST file to analyze
//   - expandIncludes: If true, expands include directives inline
//
// Returns:
//   - *AnalysisReport: Analysis report containing all procedure information
//   - error: Any error encountered during analysis
func AnalyzeFileWithOptions(filePath string, expandIncludes bool) (*AnalysisReport, error) {
	// Parse all procedures from the file
	procedures, err := rst.ParseProceduresWithOptions(filePath, expandIncludes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse procedures from %s: %w", filePath, err)
	}

	// Create the report
	report := NewAnalysisReport(filePath)

	// Group procedures from the same tab set
	// Track which tab sets we've already processed
	processedTabSets := make(map[*rst.TabSetInfo]bool)

	for _, procedure := range procedures {
		// If this procedure is part of a tab set and we haven't processed it yet
		if procedure.TabSet != nil && !processedTabSets[procedure.TabSet] {
			// Mark this tab set as processed
			processedTabSets[procedure.TabSet] = true

			// Create a grouped analysis for all procedures in this tab set
			analysis := analyzeTabSet(procedure.TabSet)
			report.AddProcedure(analysis)
		} else if procedure.TabSet == nil {
			// Regular procedure (not part of a tab set)
			analysis := analyzeProcedure(procedure)
			report.AddProcedure(analysis)
		}
		// Skip procedures that are part of an already-processed tab set
	}

	return report, nil
}

// analyzeProcedure analyzes a single procedure and returns analysis results.
func analyzeProcedure(procedure rst.Procedure) ProcedureAnalysis {
	// Get variations
	variations := rst.GetProcedureVariations(procedure)

	// If no variations, count as 1 (single variation)
	variationCount := len(variations)
	if variationCount == 0 {
		variationCount = 1
		variations = []string{"(no variations)"}
	}

	// Count steps
	stepCount := len(procedure.Steps)

	// Determine implementation type
	implementation := string(procedure.Type)

	// Check for sub-steps
	hasSubSteps := procedure.HasSubSteps

	return ProcedureAnalysis{
		Procedure:      procedure,
		Variations:     variations,
		VariationCount: variationCount,
		StepCount:      stepCount,
		HasSubSteps:    hasSubSteps,
		Implementation: implementation,
	}
}

// analyzeTabSet analyzes a tab set containing multiple procedure variations.
// This groups all procedures from the same tab set for reporting purposes.
func analyzeTabSet(tabSet *rst.TabSetInfo) ProcedureAnalysis {
	// Use the first procedure as the representative
	// (they all have the same title/heading)
	var firstProc rst.Procedure
	for _, tabID := range tabSet.TabIDs {
		if proc, ok := tabSet.Procedures[tabID]; ok {
			firstProc = proc
			break
		}
	}

	// Get all tab IDs as variations
	variations := tabSet.TabIDs

	// Count total variations
	variationCount := len(variations)

	// Use the step count from the first procedure
	// (each tab may have different step counts, but we report the first one)
	stepCount := len(firstProc.Steps)

	// Determine implementation type
	implementation := string(firstProc.Type)

	// Check for sub-steps
	hasSubSteps := firstProc.HasSubSteps

	return ProcedureAnalysis{
		Procedure:      firstProc,
		Variations:     variations,
		VariationCount: variationCount,
		StepCount:      stepCount,
		HasSubSteps:    hasSubSteps,
		Implementation: implementation,
	}
}

