package pathresolver

import (
	"fmt"
	"path/filepath"
)

// DetectProjectInfo analyzes a file path and determines the project structure.
//
// This function detects whether the file is part of a versioned or non-versioned
// project and extracts relevant information about the project structure.
//
// Versioned project structure:
//   {product}/{version}/source/...
//   Example: /path/to/manual/v8.0/source/includes/file.rst
//
// Non-versioned project structure:
//   {product}/source/...
//   Example: /path/to/atlas/source/includes/file.rst
//
// Parameters:
//   - filePath: Path to a file within the documentation tree
//
// Returns:
//   - *ProjectInfo: Information about the project structure
//   - error: Any error encountered during detection
func DetectProjectInfo(filePath string) (*ProjectInfo, error) {
	// Get absolute path
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Find the source directory
	sourceDir, err := FindSourceDirectory(absPath)
	if err != nil {
		return nil, err
	}

	// Get the parent directory of source (could be version or product)
	parent := filepath.Dir(sourceDir)
	parentName := filepath.Base(parent)

	// Check if this is a versioned project
	isVersioned, err := IsVersionedProject(sourceDir)
	if err != nil {
		return nil, err
	}

	var productDir string
	var version string

	if isVersioned {
		// Versioned project: parent is the version directory
		version = parentName
		productDir = filepath.Dir(parent)
	} else {
		// Non-versioned project: parent is the product directory
		version = ""
		productDir = parent
	}

	return &ProjectInfo{
		SourceDir:   sourceDir,
		ProductDir:  productDir,
		Version:     version,
		IsVersioned: isVersioned,
	}, nil
}

// ResolveRelativeToSource resolves a path relative to the source directory.
//
// This function takes a relative path (like "/includes/file.rst") and resolves
// it to an absolute path based on the source directory.
//
// Parameters:
//   - sourceDir: The absolute path to the source directory
//   - relativePath: The relative path to resolve (can start with / or not)
//
// Returns:
//   - string: The absolute path
//   - error: Any error encountered during resolution
func ResolveRelativeToSource(sourceDir, relativePath string) (string, error) {
	// Clean the paths
	sourceDir = filepath.Clean(sourceDir)
	relativePath = filepath.Clean(relativePath)

	// Remove leading slash if present (it's relative to source, not filesystem root)
	if len(relativePath) > 0 && relativePath[0] == '/' {
		relativePath = relativePath[1:]
	}

	// Join with source directory
	fullPath := filepath.Join(sourceDir, relativePath)

	return fullPath, nil
}

// FindProductDirectory walks up the directory tree to find the product root directory.
//
// The product directory is the parent of either:
// - The version directory (for versioned projects)
// - The source directory (for non-versioned projects)
//
// Parameters:
//   - filePath: Path to a file within the documentation tree
//
// Returns:
//   - string: Absolute path to the product directory
//   - error: Error if product directory cannot be found
func FindProductDirectory(filePath string) (string, error) {
	projectInfo, err := DetectProjectInfo(filePath)
	if err != nil {
		return "", err
	}

	return projectInfo.ProductDir, nil
}

