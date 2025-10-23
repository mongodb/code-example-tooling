// Package file_contents provides functionality for comparing file contents across versions.
package file_contents

// FileStatus represents the status of a file in a comparison.
type FileStatus int

const (
	// FileMatches indicates the file content matches the reference file
	FileMatches FileStatus = iota
	// FileDiffers indicates the file content differs from the reference file
	FileDiffers
	// FileNotFound indicates the file does not exist at the expected path
	FileNotFound
	// FileError indicates an error occurred while reading the file
	FileError
)

// String returns a string representation of the FileStatus.
func (s FileStatus) String() string {
	switch s {
	case FileMatches:
		return "matches"
	case FileDiffers:
		return "differs"
	case FileNotFound:
		return "not found"
	case FileError:
		return "error"
	default:
		return "unknown"
	}
}

// FileComparison represents the comparison result for a single file.
type FileComparison struct {
	// Version is the version identifier (e.g., "v8.0", "upcoming")
	Version string
	// FilePath is the absolute path to the file
	FilePath string
	// Status is the comparison status
	Status FileStatus
	// Error is any error encountered (only set if Status == FileError)
	Error error
	// Diff is the unified diff output (only set if Status == FileDiffers and diff was requested)
	Diff string
}

// ComparisonResult represents the overall comparison result.
type ComparisonResult struct {
	// ReferenceFile is the path to the reference file being compared against
	ReferenceFile string
	// ReferenceVersion is the version of the reference file (empty for direct comparison)
	ReferenceVersion string
	// Comparisons is the list of file comparisons
	Comparisons []FileComparison
	// TotalFiles is the total number of files compared
	TotalFiles int
	// MatchingFiles is the number of files that match
	MatchingFiles int
	// DifferingFiles is the number of files that differ
	DifferingFiles int
	// NotFoundFiles is the number of files not found
	NotFoundFiles int
	// ErrorFiles is the number of files with errors
	ErrorFiles int
}

// HasDifferences returns true if any files differ from the reference.
func (r *ComparisonResult) HasDifferences() bool {
	return r.DifferingFiles > 0
}

// AllMatch returns true if all files match the reference (excluding not found files).
func (r *ComparisonResult) AllMatch() bool {
	return r.DifferingFiles == 0 && r.ErrorFiles == 0 && r.MatchingFiles > 0
}

