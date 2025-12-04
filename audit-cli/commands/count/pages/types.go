// Package pages provides functionality for counting documentation pages.
package pages

// CountResult represents the result of counting pages.
type CountResult struct {
	// TotalCount is the total number of .txt files counted
	TotalCount int
	// ProjectCounts maps project directory names to their page counts
	ProjectCounts map[string]int
	// VersionCounts maps project names to version names to counts
	// For versioned projects: {"manual": {"manual": 100, "v8.0": 95}}
	// For non-versioned projects: {"atlas": {"": 200}}
	VersionCounts map[string]map[string]int
	// ContentDir is the path to the content directory
	ContentDir string
}

// VersionInfo contains information about a version directory.
type VersionInfo struct {
	// Name is the version directory name (e.g., "manual", "v8.0", "current")
	Name string
	// IsCurrent indicates if this is the current version
	IsCurrent bool
}

