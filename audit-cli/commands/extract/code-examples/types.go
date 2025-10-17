package code_examples

// DirectiveType represents the type of reStructuredText directive.
type DirectiveType string

const (
	// CodeBlock represents inline code blocks (.. code-block::)
	CodeBlock DirectiveType = "code-block"
	// LiteralInclude represents external file references (.. literalinclude::)
	LiteralInclude DirectiveType = "literalinclude"
	// IoCodeBlock represents input/output examples (.. io-code-block::)
	IoCodeBlock DirectiveType = "io-code-block"
)

// CodeExample represents a single code example extracted from an RST file.
//
// Each code example corresponds to one directive occurrence in the source file
// and will be written to a separate output file.
type CodeExample struct {
	SourceFile    string        // Path to the source RST file
	DirectiveName DirectiveType // Type of directive (code-block, literalinclude, io-code-block)
	Language      string        // Programming language (normalized)
	Content       string        // The actual code content
	Index         int           // The occurrence index of this directive in the source file (1-based)
	SubType       string        // For io-code-block: "input" or "output"
}

// Report contains statistics about the extraction operation.
//
// Tracks overall statistics as well as per-source-file statistics for detailed reporting.
type Report struct {
	FilesTraversed     int                       // Total number of RST files processed
	TraversedFilepaths []string                  // List of all processed file paths
	OutputFilesWritten int                       // Total number of code example files written
	LanguageCounts     map[string]int            // Count of examples by language
	DirectiveCounts    map[DirectiveType]int     // Count of examples by directive type
	SourcePathStats    map[string]*SourceStats   // Per-file statistics
}

// SourceStats contains statistics for a single source file.
//
// Used for verbose reporting to show detailed breakdown per source file.
type SourceStats struct {
	DirectiveCounts map[DirectiveType]int // Count of directives by type in this file
	LanguageCounts  map[string]int        // Count of examples by language in this file
	OutputFiles     []string              // List of output files generated from this source
}

// NewReport creates a new initialized Report with empty maps and slices.
func NewReport() *Report {
	return &Report{
		TraversedFilepaths: make([]string, 0),
		LanguageCounts:     make(map[string]int),
		DirectiveCounts:    make(map[DirectiveType]int),
		SourcePathStats:    make(map[string]*SourceStats),
	}
}

// NewSourceStats creates a new initialized SourceStats with empty maps and slices.
func NewSourceStats() *SourceStats {
	return &SourceStats{
		DirectiveCounts: make(map[DirectiveType]int),
		LanguageCounts:  make(map[string]int),
		OutputFiles:     make([]string, 0),
	}
}

// AddCodeExample updates the report with a new code example.
//
// This method updates both global statistics and per-source-file statistics.
// It should be called once for each code example that is successfully extracted.
func (r *Report) AddCodeExample(example CodeExample, outputPath string) {
	// Update global counts
	r.LanguageCounts[example.Language]++
	r.DirectiveCounts[example.DirectiveName]++

	// Update source-specific stats
	if _, exists := r.SourcePathStats[example.SourceFile]; !exists {
		r.SourcePathStats[example.SourceFile] = NewSourceStats()
	}
	stats := r.SourcePathStats[example.SourceFile]
	stats.DirectiveCounts[example.DirectiveName]++
	stats.LanguageCounts[example.Language]++
	stats.OutputFiles = append(stats.OutputFiles, outputPath)
}

// AddTraversedFile adds a file to the list of traversed files.
//
// This method should be called once for each RST file that is processed,
// including files discovered through include directives.
func (r *Report) AddTraversedFile(filepath string) {
	r.FilesTraversed++
	r.TraversedFilepaths = append(r.TraversedFilepaths, filepath)
}
