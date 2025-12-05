package procedures

import "github.com/mongodb/code-example-tooling/audit-cli/internal/rst"

// ProcedureAnalysis contains analysis results for a single procedure.
type ProcedureAnalysis struct {
	Procedure       rst.Procedure // The procedure being analyzed
	Variations      []string      // List of variation identifiers
	VariationCount  int           // Number of variations
	StepCount       int           // Number of steps
	HasSubSteps     bool          // Whether procedure contains sub-steps
	Implementation  string        // How the procedure is implemented (directive or ordered-list)
}

// AnalysisReport contains the complete analysis results for a file.
type AnalysisReport struct {
	FilePath           string              // Path to the analyzed file
	Procedures         []ProcedureAnalysis // Analysis for each procedure
	TotalProcedures    int                 // Total number of procedures
	TotalVariations    int                 // Total number of variations across all procedures
	ProceduresByType   map[string]int      // Count of procedures by implementation type
}

// NewAnalysisReport creates a new analysis report.
func NewAnalysisReport(filePath string) *AnalysisReport {
	return &AnalysisReport{
		FilePath:         filePath,
		Procedures:       []ProcedureAnalysis{},
		ProceduresByType: make(map[string]int),
	}
}

// AddProcedure adds a procedure analysis to the report.
func (r *AnalysisReport) AddProcedure(analysis ProcedureAnalysis) {
	r.Procedures = append(r.Procedures, analysis)
	r.TotalProcedures++
	r.TotalVariations += analysis.VariationCount
	r.ProceduresByType[analysis.Implementation]++
}

