package rst

import (
	"os"
	"path/filepath"
	"strings"
)

// TraverseDirectory traverses a directory and returns all file paths.
//
// If recursive is true, walks the entire directory tree. If false, only
// returns files in the immediate directory (no subdirectories).
//
// Parameters:
//   - rootPath: Root directory to traverse
//   - recursive: If true, recursively scan subdirectories
//
// Returns:
//   - []string: List of all file paths found
//   - error: Any error encountered during traversal
func TraverseDirectory(rootPath string, recursive bool) ([]string, error) {
	var files []string

	if recursive {
		err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				files = append(files, path)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	} else {
		entries, err := os.ReadDir(rootPath)
		if err != nil {
			return nil, err
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				files = append(files, filepath.Join(rootPath, entry.Name()))
			}
		}
	}

	return files, nil
}

// ShouldProcessFile determines if a file should be processed based on its extension.
//
// Returns true for files with .rst, .txt, or .md extensions (case-insensitive).
// This is used to filter files during directory traversal.
//
// Parameters:
//   - filePath: Path to the file to check
//
// Returns:
//   - bool: True if the file should be processed, false otherwise
func ShouldProcessFile(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	validExtensions := []string{".rst", ".txt", ".md"}
	for _, validExt := range validExtensions {
		if ext == validExt {
			return true
		}
	}
	return false
}

