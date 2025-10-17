package rst

import (
	"fmt"
	"os"
	"path/filepath"
)

// ParseFileWithIncludes parses a file and recursively follows include directives.
//
// This function provides a generic mechanism for processing RST files and their includes.
// It handles:
//   - Tracking visited files to prevent circular includes
//   - Calling a custom parse function for each file
//   - Recursively following .. include:: directives
//   - Resolving include paths with MongoDB-specific conventions
//
// The parseFunc is called for each file to extract content (e.g., code examples).
// It should return an error if parsing fails.
//
// Parameters:
//   - filePath: Path to the RST file to parse
//   - followIncludes: If true, recursively follow .. include:: directives
//   - visited: Map tracking already-processed files (prevents circular includes)
//   - verbose: If true, print detailed processing information
//   - parseFunc: Function to call for each file to extract content
//
// Returns:
//   - []string: List of all processed file paths (absolute paths)
//   - error: Any error encountered during parsing
func ParseFileWithIncludes(
	filePath string,
	followIncludes bool,
	visited map[string]bool,
	verbose bool,
	parseFunc func(string) error,
) ([]string, error) {
	// Prevent infinite loops from circular includes
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, err
	}

	if visited[absPath] {
		return nil, nil // Already processed this file
	}
	visited[absPath] = true

	var processedFiles []string
	processedFiles = append(processedFiles, absPath)

	// Parse the current file using the provided parse function
	if parseFunc != nil {
		if err := parseFunc(filePath); err != nil {
			return processedFiles, err
		}
	}

	// If not following includes, return just this file
	if !followIncludes {
		return processedFiles, nil
	}

	// Find and process include directives
	includeFiles, err := FindIncludeDirectives(filePath)
	if err != nil {
		return processedFiles, nil // Continue even if we can't find includes
	}

	if verbose && len(includeFiles) > 0 {
		fmt.Printf("  Found %d include(s) in %s\n", len(includeFiles), filepath.Base(filePath))
	}

	// Recursively parse included files
	for _, includeFile := range includeFiles {
		if verbose {
			fmt.Printf("  Following include: %s\n", includeFile)
		}

		includedFiles, err := ParseFileWithIncludes(includeFile, followIncludes, visited, verbose, parseFunc)
		if err != nil {
			// Log warning but continue processing other files
			fmt.Fprintf(os.Stderr, "Warning: failed to parse included file %s: %v\n", includeFile, err)
			continue
		}
		processedFiles = append(processedFiles, includedFiles...)
	}

	return processedFiles, nil
}

