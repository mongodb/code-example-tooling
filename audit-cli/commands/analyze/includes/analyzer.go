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
	// Use a recursion path to detect true circular includes
	recursionPath := make(map[string]bool)
	// Track which files we've seen for verbose output (to show duplicates with different bullet)
	seenFiles := make(map[string]bool)
	tree, err := buildIncludeTree(absPath, recursionPath, seenFiles, verbose, 0)
	if err != nil {
		return nil, err
	}

	// Collect all unique files from the tree
	allFiles := collectUniqueFiles(tree)

	// Calculate max depth
	maxDepth := calculateMaxDepth(tree, 0)

	// Count total include directives
	totalDirectives := countIncludeDirectives(tree)

	analysis := &IncludeAnalysis{
		RootFile:               absPath,
		Tree:                   tree,
		AllFiles:               allFiles,
		TotalFiles:             len(allFiles),
		TotalIncludeDirectives: totalDirectives,
		MaxDepth:               maxDepth,
	}

	return analysis, nil
}

// buildIncludeTree recursively builds a tree of include relationships.
//
// This function creates an IncludeNode for the given file and recursively
// processes all files it includes, preventing true circular includes.
//
// Parameters:
//   - filePath: Path to the file to process
//   - recursionPath: Map tracking files in the current recursion path (prevents circular includes)
//   - seenFiles: Map tracking files we've already printed (for duplicate indicators in verbose mode)
//   - verbose: If true, print detailed processing information
//   - depth: Current depth in the tree (for verbose output)
//
// Returns:
//   - *IncludeNode: Tree node representing this file and its includes
//   - error: Any error encountered during processing
func buildIncludeTree(filePath string, recursionPath map[string]bool, seenFiles map[string]bool, verbose bool, depth int) (*IncludeNode, error) {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, err
	}

	// Create the node for this file
	node := &IncludeNode{
		FilePath: absPath,
		Children: []*IncludeNode{},
	}

	// Check if this file is already in the current recursion path (true circular include)
	if recursionPath[absPath] {
		if verbose {
			indent := getIndent(depth)
			fmt.Printf("%s⚠ Circular include detected: %s\n", indent, formatDisplayPath(absPath))
		}
		return node, nil
	}

	// Add this file to the recursion path
	recursionPath[absPath] = true
	// Ensure we remove it when we're done processing this branch
	defer delete(recursionPath, absPath)

	// Find include directives in this file
	includeFiles, err := rst.FindIncludeDirectives(absPath)
	if err != nil {
		// Not a fatal error - file might not have includes
		includeFiles = []string{}
	}

	// Print verbose output for this file
	if verbose {
		indent := getIndent(depth)
		// Use hollow bullet (◦) for files we've seen before, filled bullet (•) for first occurrence
		bullet := "•"
		if seenFiles[absPath] {
			bullet = "◦"
		} else {
			seenFiles[absPath] = true
		}

		if len(includeFiles) > 0 {
			directiveWord := "include directives"
			if len(includeFiles) == 1 {
				directiveWord = "include directive"
			}
			fmt.Printf("%s%s %s (%d %s)\n", indent, bullet, formatDisplayPath(absPath), len(includeFiles), directiveWord)
		} else {
			fmt.Printf("%s%s %s\n", indent, bullet, formatDisplayPath(absPath))
		}
	}

	// Recursively process each included file
	for _, includeFile := range includeFiles {
		childNode, err := buildIncludeTree(includeFile, recursionPath, seenFiles, verbose, depth+1)
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

// collectUniqueFiles traverses the tree and collects all unique file paths.
//
// This function recursively walks the tree and builds a list of all unique
// files that appear in the tree, even if they appear multiple times.
//
// Parameters:
//   - node: The root node of the tree to traverse
//
// Returns:
//   - []string: List of unique file paths
func collectUniqueFiles(node *IncludeNode) []string {
	if node == nil {
		return []string{}
	}

	visited := make(map[string]bool)
	var files []string

	var traverse func(*IncludeNode)
	traverse = func(n *IncludeNode) {
		if n == nil {
			return
		}

		// Add this file if we haven't seen it before
		if !visited[n.FilePath] {
			visited[n.FilePath] = true
			files = append(files, n.FilePath)
		}

		// Traverse children
		for _, child := range n.Children {
			traverse(child)
		}
	}

	traverse(node)
	return files
}

// countIncludeDirectives counts the total number of include directive instances in the tree.
//
// This function counts every include directive in every file, including duplicates.
// For example, if file A includes file B, and file C also includes file B,
// that counts as 2 include directives (even though B is only one unique file).
//
// Parameters:
//   - node: The root node of the tree to traverse
//
// Returns:
//   - int: Total number of include directive instances
func countIncludeDirectives(node *IncludeNode) int {
	if node == nil {
		return 0
	}

	// Count the children of this node (these are the include directives in this file)
	count := len(node.Children)

	// Recursively count include directives in all children
	for _, child := range node.Children {
		count += countIncludeDirectives(child)
	}

	return count
}

