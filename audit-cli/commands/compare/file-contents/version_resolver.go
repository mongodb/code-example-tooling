package file_contents

import (
	"fmt"
	"path/filepath"
	"strings"
)

// VersionPath represents a resolved file path for a specific version.
type VersionPath struct {
	Version  string
	FilePath string
}

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
// Given a file path under a product directory, this function extracts the
// version segment (the directory name before "source").
//
// Example:
//   Input: /path/to/manual/v8.0/source/includes/file.rst
//   Product Dir: /path/to/manual
//   Output: v8.0
//
// Parameters:
//   - filePath: The absolute path to the file
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

