package includes

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mongodb/code-example-tooling/audit-cli/internal/rst"
)

// AnalyzeIncludes analyzes a file and builds a tree of include relationships.
//
// This function recursively follows include directives and builds both a tree structure
// and a flat list of all files discovered. It tracks the maximum depth of nesting.
//
// Parameters:
//   - filePath: Path to the RST file to analyze
//   - verbose: If true, print detailed processing information
//
// Returns:
//   - *IncludeAnalysis: Analysis results including tree and file list
//   - error: Any error encountered during analysis
func AnalyzeIncludes(filePath string, verbose bool) (*IncludeAnalysis, error) {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Verify the file exists
	if _, err := os.Stat(absPath); err != nil {
		return nil, fmt.Errorf("file not found: %s", absPath)
	}

	if verbose {
		fmt.Printf("Analyzing includes for: %s\n\n", absPath)
	}

	// Build the tree structure
	visited := make(map[string]bool)
	tree, err := buildIncludeTree(absPath, visited, verbose, 0)
	if err != nil {
		return nil, err
	}

	// Collect all unique files from the visited map
	// The visited map contains all unique files that were processed
	allFiles := make([]string, 0, len(visited))
	for file := range visited {
		allFiles = append(allFiles, file)
	}

	// Calculate max depth
	maxDepth := calculateMaxDepth(tree, 0)

	analysis := &IncludeAnalysis{
		RootFile:   absPath,
		Tree:       tree,
		AllFiles:   allFiles,
		TotalFiles: len(allFiles),
		MaxDepth:   maxDepth,
	}

	return analysis, nil
}

// buildIncludeTree recursively builds a tree of include relationships.
//
// This function creates an IncludeNode for the given file and recursively
// processes all files it includes, preventing circular includes.
//
// Parameters:
//   - filePath: Path to the file to process
//   - visited: Map tracking already-processed files (prevents circular includes)
//   - verbose: If true, print detailed processing information
//   - depth: Current depth in the tree (for verbose output)
//
// Returns:
//   - *IncludeNode: Tree node representing this file and its includes
//   - error: Any error encountered during processing
func buildIncludeTree(filePath string, visited map[string]bool, verbose bool, depth int) (*IncludeNode, error) {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, err
	}

	// Create the node for this file
	node := &IncludeNode{
		FilePath: absPath,
		Children: []*IncludeNode{},
	}

	// Check if we've already visited this file (circular include)
	if visited[absPath] {
		if verbose {
			indent := getIndent(depth)
			fmt.Printf("%s⚠ Circular include detected: %s\n", indent, filepath.Base(absPath))
		}
		return node, nil
	}
	visited[absPath] = true

	// Find include directives in this file
	includeFiles, err := rst.FindIncludeDirectives(absPath)
	if err != nil {
		// Not a fatal error - file might not have includes
		includeFiles = []string{}
	}

	if verbose && len(includeFiles) > 0 {
		indent := getIndent(depth)
		fmt.Printf("%s📄 %s (%d includes)\n", indent, filepath.Base(absPath), len(includeFiles))
	}

	// Recursively process each included file
	for _, includeFile := range includeFiles {
		childNode, err := buildIncludeTree(includeFile, visited, verbose, depth+1)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to process file %s: %v\n", includeFile, err)
			continue
		}
		node.Children = append(node.Children, childNode)
	}

	return node, nil
}

// calculateMaxDepth calculates the maximum depth of the include tree.
//
// This function recursively traverses the tree to find the deepest nesting level.
//
// Parameters:
//   - node: Current node being processed
//   - currentDepth: Depth of the current node
//
// Returns:
//   - int: Maximum depth found in the tree
func calculateMaxDepth(node *IncludeNode, currentDepth int) int {
	if node == nil || len(node.Children) == 0 {
		return currentDepth
	}

	maxChildDepth := currentDepth
	for _, child := range node.Children {
		childDepth := calculateMaxDepth(child, currentDepth+1)
		if childDepth > maxChildDepth {
			maxChildDepth = childDepth
		}
	}

	return maxChildDepth
}

// getIndent returns an indentation string for the given depth level.
//
// This is used for verbose output to show the tree structure.
//
// Parameters:
//   - depth: Nesting depth level
//
// Returns:
//   - string: Indentation string (2 spaces per level)
func getIndent(depth int) string {
	indent := ""
	for i := 0; i < depth; i++ {
		indent += "  "
	}
	return indent
}

