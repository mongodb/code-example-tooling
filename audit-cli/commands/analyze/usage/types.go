package usage

// ReferenceAnalysis contains the results of analyzing which files reference a target file.
//
// This structure holds both a flat list of referencing files and a hierarchical
// tree structure showing the reference relationships.
type ReferenceAnalysis struct {
	// TargetFile is the absolute path to the file being analyzed
	TargetFile string

	// ReferencingFiles is a flat list of all files that reference the target
	ReferencingFiles []FileReference

	// ReferenceTree is a hierarchical tree structure of references
	// (for future use - showing nested references)
	ReferenceTree *ReferenceNode

	// TotalReferences is the total number of directive occurrences
	TotalReferences int

	// TotalFiles is the total number of unique files that reference the target
	TotalFiles int

	// SourceDir is the source directory that was searched
	SourceDir string
}

// FileReference represents a single file that references the target file.
//
// This structure captures details about how and where the reference occurs.
type FileReference struct {
	// FilePath is the absolute path to the file that references the target
	FilePath string `json:"file_path"`

	// DirectiveType is the type of directive used to reference the file
	// Possible values: "include", "literalinclude", "io-code-block", "toctree"
	DirectiveType string `json:"directive_type"`

	// ReferencePath is the path used in the directive (as written in the file)
	ReferencePath string `json:"reference_path"`

	// LineNumber is the line number where the reference occurs
	LineNumber int `json:"line_number"`
}

// ReferenceNode represents a node in the reference tree.
//
// This structure is used to build a hierarchical view of references,
// showing which files reference the target and which files reference those files.
type ReferenceNode struct {
	// FilePath is the absolute path to this file
	FilePath string

	// DirectiveType is the type of directive used to reference the file
	DirectiveType string

	// ReferencePath is the path used in the directive
	ReferencePath string

	// Children are files that include this file
	// (for building nested reference trees)
	Children []*ReferenceNode
}

// GroupedFileReference represents a file with all its references to the target.
//
// This structure groups multiple references from the same file together,
// showing how many times a file references the target and where.
type GroupedFileReference struct {
	// FilePath is the absolute path to the file
	FilePath string

	// DirectiveType is the type of directive used
	DirectiveType string

	// References is the list of all references from this file
	References []FileReference

	// Count is the number of references from this file
	Count int
}

