package includes

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/mongodb/code-example-tooling/audit-cli/internal/pathresolver"
)

// PrintTree prints the include tree structure.
//
// This function displays the hierarchical relationship of includes using
// tree-style formatting with box-drawing characters.
//
// Parameters:
//   - analysis: The analysis results containing the tree structure
func PrintTree(analysis *IncludeAnalysis) {
	fmt.Println("============================================================")
	fmt.Println("INCLUDE TREE")
	fmt.Println("============================================================")
	fmt.Printf("Root File: %s\n", analysis.RootFile)
	fmt.Printf("Total Files: %d\n", analysis.TotalFiles)
	fmt.Printf("Max Depth: %d\n", analysis.MaxDepth)
	fmt.Println("============================================================")
	fmt.Println()

	if analysis.Tree != nil {
		printTreeNode(analysis.Tree, "", true, true)
	}

	fmt.Println()
}

// printTreeNode recursively prints a tree node with proper formatting.
//
// This function uses box-drawing characters to create a visual tree structure.
//
// Parameters:
//   - node: The node to print
//   - prefix: Prefix string for indentation
//   - isLast: Whether this is the last child of its parent
//   - isRoot: Whether this is the root node
func printTreeNode(node *IncludeNode, prefix string, isLast bool, isRoot bool) {
	if node == nil {
		return
	}

	// Print the current node
	if isRoot {
		fmt.Printf("%s\n", formatDisplayPath(node.FilePath))
	} else {
		connector := "├── "
		if isLast {
			connector = "└── "
		}
		fmt.Printf("%s%s%s\n", prefix, connector, formatDisplayPath(node.FilePath))
	}

	// Print children
	childPrefix := prefix
	if !isRoot {
		if isLast {
			childPrefix += "    "
		} else {
			childPrefix += "│   "
		}
	}

	for i, child := range node.Children {
		isLastChild := i == len(node.Children)-1
		printTreeNode(child, childPrefix, isLastChild, false)
	}
}

// PrintList prints a flat list of all included files.
//
// This function displays all files discovered through include directives
// in the order they were discovered (depth-first traversal).
//
// Parameters:
//   - analysis: The analysis results containing the file list
func PrintList(analysis *IncludeAnalysis) {
	fmt.Println("============================================================")
	fmt.Println("INCLUDE FILE LIST")
	fmt.Println("============================================================")
	fmt.Printf("Root File: %s\n", analysis.RootFile)
	fmt.Printf("Total Files: %d\n", analysis.TotalFiles)
	fmt.Println("============================================================")
	fmt.Println()

	for i, file := range analysis.AllFiles {
		fmt.Printf("%3d. %s\n", i+1, file)
	}

	fmt.Println()
}

// PrintSummary prints a brief summary of the analysis.
//
// This function is used when neither --tree nor --list is specified,
// providing basic statistics about the include structure.
//
// Parameters:
//   - analysis: The analysis results
func PrintSummary(analysis *IncludeAnalysis) {
	fmt.Println("============================================================")
	fmt.Println("INCLUDE ANALYSIS SUMMARY")
	fmt.Println("============================================================")
	fmt.Printf("Root File: %s\n", analysis.RootFile)
	fmt.Printf("Total Files: %d\n", analysis.TotalFiles)
	fmt.Printf("Max Depth: %d\n", analysis.MaxDepth)
	fmt.Println("============================================================")
	fmt.Println()
	fmt.Println("Use --tree to see the hierarchical structure")
	fmt.Println("Use --list to see a flat list of all files")
	fmt.Println()
}

// formatDisplayPath formats a file path for display in the tree or verbose output.
//
// This function returns:
//   - If the file is in an "includes" directory: the path starting from "includes"
//     (e.g., "includes/load-sample-data.rst" or "includes/php/connection.rst")
//   - If the file is NOT in an "includes" directory: the path from the source directory
//     (e.g., "get-started/node/language-connection-steps.rst")
//
// This helps writers understand the directory structure and disambiguate files
// with the same name in different directories.
//
// Parameters:
//   - filePath: Absolute path to the file
//
// Returns:
//   - string: Formatted path for display
func formatDisplayPath(filePath string) string {
	// Try to find the source directory
	sourceDir, err := pathresolver.FindSourceDirectory(filePath)
	if err != nil {
		// If we can't find source directory, just return the base name
		return filepath.Base(filePath)
	}

	// Check if the file is in an includes directory
	// Walk up from the file to find if there's an "includes" directory
	dir := filepath.Dir(filePath)
	var includesDir string

	for {
		// Check if the current directory is named "includes"
		if filepath.Base(dir) == "includes" {
			includesDir = dir
			break
		}

		// Move up one directory
		parent := filepath.Dir(dir)

		// If we've reached the source directory or root, stop
		if parent == dir || dir == sourceDir {
			break
		}

		dir = parent
	}

	// If we found an includes directory, get the relative path from it
	if includesDir != "" {
		relPath, err := filepath.Rel(includesDir, filePath)
		if err == nil && !strings.HasPrefix(relPath, "..") {
			// Prepend "includes/" to show it's in the includes directory
			return filepath.Join("includes", relPath)
		}
	}

	// Otherwise, get the relative path from the source directory
	relPath, err := filepath.Rel(sourceDir, filePath)
	if err != nil {
		// If we can't get relative path, just return the base name
		return filepath.Base(filePath)
	}

	return relPath
}

