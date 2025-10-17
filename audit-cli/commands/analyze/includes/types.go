package includes

// IncludeNode represents a file and its included files in a tree structure.
//
// This type is used to build a hierarchical representation of include relationships,
// where each node represents a file and its children are the files it includes.
type IncludeNode struct {
	FilePath string         // Absolute path to the file
	Children []*IncludeNode // Files included by this file
}

// IncludeAnalysis contains the results of analyzing include directives.
//
// This type holds both the tree structure and the flat list of all files
// discovered through include directives.
type IncludeAnalysis struct {
	RootFile   string       // The original file that was analyzed
	Tree       *IncludeNode // Tree structure of include relationships
	AllFiles   []string     // Flat list of all files (in order discovered)
	TotalFiles int          // Total number of unique files
	MaxDepth   int          // Maximum depth of include nesting
}

