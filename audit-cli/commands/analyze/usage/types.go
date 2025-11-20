package usage

// UsageAnalysis contains the results of analyzing which files use a target file.
//
// This structure holds both a flat list of files that use the target and a hierarchical
// tree structure showing the usage relationships.
type UsageAnalysis struct {
	// TargetFile is the absolute path to the file being analyzed
	TargetFile string

	// UsingFiles is a flat list of all files that use the target
	UsingFiles []FileUsage

	// UsageTree is a hierarchical tree structure of usages
	// (for future use - showing nested usages)
	UsageTree *UsageNode

	// TotalUsages is the total number of directive occurrences
	TotalUsages int

	// TotalFiles is the total number of unique files that use the target
	TotalFiles int

	// SourceDir is the source directory that was searched
	SourceDir string
}

// FileUsage represents a single file that uses the target file.
//
// This structure captures details about how and where the usage occurs.
type FileUsage struct {
	// FilePath is the absolute path to the file that uses the target
	FilePath string `json:"file_path"`

	// DirectiveType is the type of directive used to reference the file
	// Possible values: "include", "literalinclude", "io-code-block", "toctree"
	DirectiveType string `json:"directive_type"`

	// UsagePath is the path used in the directive (as written in the file)
	UsagePath string `json:"usage_path"`

	// LineNumber is the line number where the usage occurs
	LineNumber int `json:"line_number"`
}

// UsageNode represents a node in the usage tree.
//
// This structure is used to build a hierarchical view of usages,
// showing which files use the target and which files use those files.
type UsageNode struct {
	// FilePath is the absolute path to this file
	FilePath string

	// DirectiveType is the type of directive used to reference the file
	DirectiveType string

	// UsagePath is the path used in the directive
	UsagePath string

	// Children are files that include this file
	// (for building nested usage trees)
	Children []*UsageNode
}

// GroupedFileUsage represents a file with all its usages of the target.
//
// This structure groups multiple usages from the same file together,
// showing how many times a file uses the target and where.
type GroupedFileUsage struct {
	// FilePath is the absolute path to the file
	FilePath string

	// DirectiveType is the type of directive used
	DirectiveType string

	// Usages is the list of all usages from this file
	Usages []FileUsage

	// Count is the number of usages from this file
	Count int
}

