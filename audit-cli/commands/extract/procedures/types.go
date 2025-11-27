package procedures

import "github.com/mongodb/code-example-tooling/audit-cli/internal/rst"

// ProcedureVariation represents a single variation of a procedure to be extracted.
type ProcedureVariation struct {
	Procedure     rst.Procedure // The procedure
	VariationName string        // The variation identifier (e.g., "python", "nodejs", "driver, nodejs")
	SourceFile    string        // Path to the source RST file
	OutputFile    string        // Path to the output file for this variation
}

// ExtractionReport contains statistics about the extraction operation.
type ExtractionReport struct {
	TotalProcedures  int      // Total number of procedures found
	TotalVariations  int      // Total number of variations extracted
	FilesProcessed   int      // Number of files processed
	FilesWritten     int      // Number of output files written
	Errors           []string // Any errors encountered
}

// NewExtractionReport creates a new extraction report.
func NewExtractionReport() *ExtractionReport {
	return &ExtractionReport{
		Errors: []string{},
	}
}

// AddError adds an error to the report.
func (r *ExtractionReport) AddError(err string) {
	r.Errors = append(r.Errors, err)
}

