// Package pages provides counting functionality for documentation pages.
package pages

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mongodb/code-example-tooling/audit-cli/internal/projectinfo"
)

// CountPages counts .txt files in the content directory.
//
// This function navigates to the content directory from the monorepo root
// and counts .txt files based on the specified filters.
//
// Parameters:
//   - dirPath: Path to the directory to count (can be monorepo root or content dir)
//   - forProject: If non-empty, only count files for this project
//   - excludeDirs: List of directory names to exclude from counting
//   - currentOnly: If true, only count files in the current version
//   - byVersion: If true, track counts by version
//
// Returns:
//   - *CountResult: The counting results
//   - error: Any error encountered during counting
func CountPages(dirPath string, forProject string, excludeDirs []string, currentOnly bool, byVersion bool) (*CountResult, error) {
	// Get absolute path
	absDirPath, err := filepath.Abs(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Check if path exists
	if _, err := os.Stat(absDirPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("directory does not exist: %s", absDirPath)
	}

	// Find the content directory
	contentDir, err := findContentDirectory(absDirPath)
	if err != nil {
		return nil, err
	}

	result := &CountResult{
		TotalCount:    0,
		ProjectCounts: make(map[string]int),
		VersionCounts: make(map[string]map[string]int),
		ContentDir:    contentDir,
	}

	// Default exclusions at the root of content or source
	defaultExclusions := map[string]bool{
		"404":               true,
		"docs-platform":     true,
		"meta":              true,
		"table-of-contents": true,
	}

	// Add user-specified exclusions
	userExclusions := make(map[string]bool)
	for _, dir := range excludeDirs {
		userExclusions[dir] = true
	}

	// Walk through the content directory
	err = filepath.Walk(contentDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get relative path from content directory
		relPath, err := filepath.Rel(contentDir, path)
		if err != nil {
			return err
		}

		// Skip the content directory itself
		if relPath == "." {
			return nil
		}

		// Check if this is a directory we should skip
		if info.IsDir() {
			dirName := info.Name()

			// Check if this is a code-examples directory at root of content or source
			if dirName == "code-examples" {
				parentDir := filepath.Dir(path)
				// Skip if at root of content
				if parentDir == contentDir {
					return filepath.SkipDir
				}
				// Skip if at root of source (content/project/source/code-examples)
				if filepath.Base(parentDir) == "source" {
					grandparentDir := filepath.Dir(parentDir)
					// Check if grandparent is a direct child of content
					if filepath.Dir(grandparentDir) == contentDir {
						return filepath.SkipDir
					}
				}
			}

			// Check default exclusions (only at root of content)
			if filepath.Dir(path) == contentDir && defaultExclusions[dirName] {
				return filepath.SkipDir
			}

			// Check user exclusions (anywhere in the tree)
			if userExclusions[dirName] {
				return filepath.SkipDir
			}

			return nil
		}

		// Only count .txt files
		if filepath.Ext(path) != ".txt" {
			return nil
		}

		// Extract project name from path (first directory under content)
		projectName := extractProjectName(relPath)
		if projectName == "" {
			// File is directly in content directory, not in a project
			return nil
		}

		// If filtering by project, check if this file matches
		if forProject != "" && projectName != forProject {
			return nil
		}

		// Extract version information if needed
		var versionName string
		if byVersion || currentOnly {
			projectDir := filepath.Join(contentDir, projectName)
			versionName = extractVersionFromPath(relPath, projectName)

			// If currentOnly is set, check if this file is in the current version
			if currentOnly {
				// Check if project is versioned
				versions, err := findVersionDirectories(projectDir)
				if err != nil {
					return err
				}

				// For non-versioned projects, versionName will be empty, which is fine
				if len(versions) > 0 {
					// This is a versioned project - only count if in current version
					if !projectinfo.IsCurrentVersion(versionName) {
						return nil
					}
				}
			}
		}

		// Count this file
		result.TotalCount++
		result.ProjectCounts[projectName]++

		// Track by version if requested
		if byVersion {
			if result.VersionCounts[projectName] == nil {
				result.VersionCounts[projectName] = make(map[string]int)
			}
			result.VersionCounts[projectName][versionName]++
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk content directory: %w", err)
	}

	return result, nil
}

// findContentDirectory finds the content directory from the given path.
// It checks if the path is already a content directory, or if it contains one.
func findContentDirectory(dirPath string) (string, error) {
	// Check if this is already a content directory
	if filepath.Base(dirPath) == "content" {
		return dirPath, nil
	}

	// Check if there's a content subdirectory
	contentDir := filepath.Join(dirPath, "content")
	if _, err := os.Stat(contentDir); err == nil {
		return contentDir, nil
	}

	return "", fmt.Errorf("content directory not found in: %s\n\nPlease provide the path to the monorepo root or content directory", dirPath)
}

// extractProjectName extracts the project name from a relative path.
// Returns the first directory component, which represents the project.
func extractProjectName(relPath string) string {
	parts := strings.Split(relPath, string(filepath.Separator))
	if len(parts) < 1 {
		return ""
	}
	return parts[0]
}

// extractVersionFromPath extracts the version name from a relative path.
// For versioned projects: content/project/version/source/file.txt -> "version"
// For non-versioned projects: content/project/source/file.txt -> ""
// Parameters:
//   - relPath: Relative path from content directory
//   - projectName: Name of the project (first directory component)
//
// Returns the version name, or empty string if non-versioned
func extractVersionFromPath(relPath string, projectName string) string {
	parts := strings.Split(relPath, string(filepath.Separator))

	// Need at least: project/version/source/file or project/source/file
	if len(parts) < 3 {
		return ""
	}

	// parts[0] is the project name
	// parts[1] could be either "source" (non-versioned) or version name (versioned)
	if parts[1] == "source" {
		// Non-versioned project
		return ""
	}

	// Check if parts[1] looks like a version directory
	if projectinfo.IsVersionDirectory(parts[1]) {
		return parts[1]
	}

	return ""
}

// findVersionDirectories finds all version directories within a project directory.
// Returns a list of VersionInfo structs with version names and whether they're current.
// If the project has no versions (source is directly under project), returns empty slice.
func findVersionDirectories(projectDir string) ([]VersionInfo, error) {
	// Check if there's a direct "source" directory (non-versioned project)
	sourceDir := filepath.Join(projectDir, "source")
	if _, err := os.Stat(sourceDir); err == nil {
		// Non-versioned project
		return []VersionInfo{}, nil
	}

	// Use projectinfo to discover all versions
	versionNames, err := projectinfo.DiscoverAllVersions(projectDir)
	if err != nil {
		// If no versions found, treat as non-versioned
		return []VersionInfo{}, nil
	}

	// Convert to VersionInfo structs with IsCurrent flag
	var versions []VersionInfo
	for _, name := range versionNames {
		versions = append(versions, VersionInfo{
			Name:      name,
			IsCurrent: projectinfo.IsCurrentVersion(name),
		})
	}

	return versions, nil
}

// getCurrentVersion finds the current version directory within a project.
// Returns the version name if found, empty string if not found or non-versioned.
func getCurrentVersion(projectDir string) (string, error) {
	versions, err := findVersionDirectories(projectDir)
	if err != nil {
		return "", err
	}

	// Non-versioned project
	if len(versions) == 0 {
		return "", nil
	}

	// Find the current version
	for _, v := range versions {
		if v.IsCurrent {
			return v.Name, nil
		}
	}

	return "", fmt.Errorf("no current version found in project directory: %s", projectDir)
}
