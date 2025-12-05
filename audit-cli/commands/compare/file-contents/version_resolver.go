package file_contents

import (
	"github.com/mongodb/code-example-tooling/audit-cli/internal/projectinfo"
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
//   - []projectinfo.VersionPath: List of resolved version paths
//   - error: Any error encountered during resolution
func ResolveVersionPaths(referenceFile string, productDir string, versions []string) ([]projectinfo.VersionPath, error) {
	return projectinfo.ResolveVersionPaths(referenceFile, productDir, versions)
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
	return projectinfo.ExtractVersionFromPath(filePath, productDir)
}

