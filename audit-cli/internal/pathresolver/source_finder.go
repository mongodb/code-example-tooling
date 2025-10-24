package pathresolver

import (
	"fmt"
	"os"
	"path/filepath"
)

// FindSourceDirectory walks up the directory tree to find the "source" directory.
//
// MongoDB documentation is typically organized with a "source" directory at the root.
// This function walks up from the current file to find that directory, which is used
// as the base for resolving include paths.
//
// Parameters:
//   - filePath: Path to a file within the documentation tree
//
// Returns:
//   - string: Absolute path to the source directory
//   - error: Error if source directory cannot be found
func FindSourceDirectory(filePath string) (string, error) {
	// Get absolute path first
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Get the directory containing the file
	dir := filepath.Dir(absPath)

	// Walk up the directory tree
	for {
		// Check if the current directory is named "source"
		if filepath.Base(dir) == "source" {
			return dir, nil
		}

		// Check if there's a "source" subdirectory
		sourceSubdir := filepath.Join(dir, "source")
		if info, err := os.Stat(sourceSubdir); err == nil && info.IsDir() {
			return sourceSubdir, nil
		}

		// Move up one directory
		parent := filepath.Dir(dir)

		// If we've reached the root, stop
		if parent == dir {
			return "", fmt.Errorf("could not find source directory for %s", filePath)
		}

		dir = parent
	}
}

