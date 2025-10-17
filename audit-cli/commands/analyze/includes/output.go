package includes

import (
	"fmt"
	"path/filepath"
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
		fmt.Printf("%s\n", filepath.Base(node.FilePath))
	} else {
		connector := "├── "
		if isLast {
			connector = "└── "
		}
		fmt.Printf("%s%s%s\n", prefix, connector, filepath.Base(node.FilePath))
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

