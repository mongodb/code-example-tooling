package orphanedfiles

// OutputFormat represents the output format for the command.
type OutputFormat string

const (
	// FormatText outputs results in human-readable text format
	FormatText OutputFormat = "text"
	// FormatJSON outputs results in JSON format
	FormatJSON OutputFormat = "json"
)

// OrphanedFilesAnalysis represents the results of an orphaned files analysis.
//
// This structure contains information about all files scanned and which ones
// were found to be orphaned (having no incoming references).
type OrphanedFilesAnalysis struct {
	// SourceDir is the absolute path to the source directory that was scanned
	SourceDir string `json:"source_dir"`

	// TotalFiles is the total number of RST/YAML files found in the source directory
	TotalFiles int `json:"total_files"`

	// TotalOrphaned is the number of files with no incoming references
	TotalOrphaned int `json:"total_orphaned"`

	// OrphanedFiles is the list of files that have no incoming references
	// Each path is relative to the source directory
	OrphanedFiles []string `json:"orphaned_files"`

	// IncludedToctree indicates whether toctree references were considered
	IncludedToctree bool `json:"included_toctree"`
}

