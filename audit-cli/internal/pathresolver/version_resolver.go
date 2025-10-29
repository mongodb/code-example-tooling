package pathresolver

import (
	"fmt"
	"path/filepath"
	"strings"
)

// ResolveVersionPaths resolves file paths for all specified versions.
//
// Given a reference file path and a list of versions, this function constructs
// the corresponding file paths for each version by replacing the version segment
// in the path.
//
// Example:
//   Input: /path/to/manual/manual/source/includes/file.rst
//   Versions: [manual, upcoming, v8.1, v8.0]
//   Output:
//     - manual: /path/to/manual/manual/source/includes/file.rst
//     - upcoming: /path/to/manual/upcoming/source/includes/file.rst
//     - v8.1: /path/to/manual/v8.1/source/includes/file.rst
//     - v8.0: /path/to/manual/v8.0/source/includes/file.rst
//
// Parameters:
//   - referenceFile: The absolute path to the reference file
//   - productDir: The absolute path to the product directory (e.g., /path/to/manual)
//   - versions: List of version identifiers
//
// Returns:
//   - []VersionPath: List of resolved version paths
//   - error: Any error encountered during resolution
func ResolveVersionPaths(referenceFile string, productDir string, versions []string) ([]VersionPath, error) {
	// Clean the paths
	referenceFile = filepath.Clean(referenceFile)
	productDir = filepath.Clean(productDir)

	// Ensure productDir ends with a separator for proper prefix matching
	if !strings.HasSuffix(productDir, string(filepath.Separator)) {
		productDir += string(filepath.Separator)
	}

	// Check if referenceFile is under productDir
	if !strings.HasPrefix(referenceFile, productDir) {
		return nil, fmt.Errorf("reference file %s is not under product directory %s", referenceFile, productDir)
	}

	// Extract the relative path from productDir
	relativePath := strings.TrimPrefix(referenceFile, productDir)

	// Find the version segment and the path after it
	// Expected format: {version}/source/{rest-of-path}
	parts := strings.Split(relativePath, string(filepath.Separator))
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid file path structure: expected {version}/source/... format, got %s", relativePath)
	}

	// Find the "source" directory
	sourceIndex := -1
	for i, part := range parts {
		if part == "source" {
			sourceIndex = i
			break
		}
	}

	if sourceIndex == -1 {
		return nil, fmt.Errorf("could not find 'source' directory in path: %s", relativePath)
	}

	if sourceIndex == 0 {
		return nil, fmt.Errorf("invalid path structure: 'source' cannot be the first segment in %s", relativePath)
	}

	// The version is the segment before "source"
	// Everything from "source" onwards is the path we want to preserve
	pathFromSource := strings.Join(parts[sourceIndex:], string(filepath.Separator))

	// Build version paths
	var versionPaths []VersionPath
	for _, version := range versions {
		versionPath := filepath.Join(productDir, version, pathFromSource)
		versionPaths = append(versionPaths, VersionPath{
			Version:  version,
			FilePath: versionPath,
		})
	}

	return versionPaths, nil
}

// ExtractVersionFromPath extracts the version identifier from a file path.
//
// Given a file path within a versioned project, this function extracts the
// version segment (the directory before "source").
//
// Example:
//   Input: /path/to/manual/v8.0/source/includes/file.rst
//   Output: "v8.0"
//
// Parameters:
//   - filePath: The absolute path to a file
//   - productDir: The absolute path to the product directory
//
// Returns:
//   - string: The version identifier
//   - error: Any error encountered during extraction
func ExtractVersionFromPath(filePath string, productDir string) (string, error) {
	// Clean the paths
	filePath = filepath.Clean(filePath)
	productDir = filepath.Clean(productDir)

	// Ensure productDir ends with a separator for proper prefix matching
	if !strings.HasSuffix(productDir, string(filepath.Separator)) {
		productDir += string(filepath.Separator)
	}

	// Check if filePath is under productDir
	if !strings.HasPrefix(filePath, productDir) {
		return "", fmt.Errorf("file path %s is not under product directory %s", filePath, productDir)
	}

	// Extract the relative path from productDir
	relativePath := strings.TrimPrefix(filePath, productDir)

	// Split into parts
	parts := strings.Split(relativePath, string(filepath.Separator))
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid file path structure: expected {version}/source/... format, got %s", relativePath)
	}

	// Find the "source" directory
	sourceIndex := -1
	for i, part := range parts {
		if part == "source" {
			sourceIndex = i
			break
		}
	}

	if sourceIndex == -1 {
		return "", fmt.Errorf("could not find 'source' directory in path: %s", relativePath)
	}

	if sourceIndex == 0 {
		return "", fmt.Errorf("invalid path structure: 'source' cannot be the first segment in %s", relativePath)
	}

	// The version is the segment before "source"
	version := parts[sourceIndex-1]

	return version, nil
}

// IsVersionedProject determines if a path is part of a versioned project.
//
// A versioned project has the structure: {product}/{version}/source/...
// A non-versioned project has the structure: {product}/source/...
//
// This function checks if there's a directory between the product root and "source".
//
// Parameters:
//   - sourceDir: The absolute path to the source directory
//
// Returns:
//   - bool: True if this is a versioned project
//   - error: Any error encountered during detection
func IsVersionedProject(sourceDir string) (bool, error) {
	// Get the parent directory of source
	parent := filepath.Dir(sourceDir)
	
	// Check if the parent directory name looks like a version
	// Common patterns: v8.0, v7.0, manual, upcoming, master, current
	parentName := filepath.Base(parent)
	
	// If parent is named after common version patterns, it's versioned
	// This is a heuristic - we check if there's a directory between product root and source
	grandparent := filepath.Dir(parent)
	
	// If grandparent is the root or doesn't exist, it's likely non-versioned
	if grandparent == parent || grandparent == "/" || grandparent == "." {
		return false, nil
	}
	
	// Check if there's another "source" directory at the grandparent level
	// If not, then parent is likely a version directory
	grandparentSource := filepath.Join(grandparent, "source")
	if grandparentSource == sourceDir {
		// This means parent is not a version directory
		return false, nil
	}
	
	// If we have: grandparent/parent/source, then parent is likely a version
	// and this is a versioned project
	return parentName != "", nil
}

