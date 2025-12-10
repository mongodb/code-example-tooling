package file_contents

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mongodb/code-example-tooling/audit-cli/internal/projectinfo"
)

// CompareFiles performs a direct comparison between two files.
//
// This function compares two files directly without version resolution.
//
// Parameters:
//   - file1Path: Path to the first file
//   - file2Path: Path to the second file
//   - generateDiff: If true, generate unified diff for differences
//   - verbose: If true, show detailed processing information
//
// Returns:
//   - *ComparisonResult: The comparison result
//   - error: Any error encountered during comparison
func CompareFiles(file1Path, file2Path string, generateDiff bool, verbose bool) (*ComparisonResult, error) {
	if verbose {
		fmt.Printf("Comparing files:\n")
		fmt.Printf("  File 1: %s\n", file1Path)
		fmt.Printf("  File 2: %s\n", file2Path)
	}

	// Read the reference file
	content1, err := os.ReadFile(file1Path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", file1Path, err)
	}

	// Read the comparison file
	content2, err := os.ReadFile(file2Path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", file2Path, err)
	}

	// Compare contents
	result := &ComparisonResult{
		ReferenceFile: file1Path,
		TotalFiles:    1,
	}

	comparison := FileComparison{
		Version:  filepath.Base(filepath.Dir(file2Path)),
		FilePath: file2Path,
	}

	if AreFilesIdentical(string(content1), string(content2)) {
		comparison.Status = FileMatches
		result.MatchingFiles = 1
	} else {
		comparison.Status = FileDiffers
		result.DifferingFiles = 1

		if generateDiff {
			diff, err := GenerateDiff(file1Path, string(content1), file2Path, string(content2))
			if err != nil {
				return nil, fmt.Errorf("failed to generate diff: %w", err)
			}
			comparison.Diff = diff
		}
	}

	result.Comparisons = []FileComparison{comparison}

	return result, nil
}

// CompareVersions performs a version-based comparison.
//
// This function compares a reference file against the same file across
// multiple versions of the documentation.
//
// Parameters:
//   - referenceFile: Path to the reference file
//   - productDir: Path to the product directory
//   - versions: List of version identifiers to compare
//   - generateDiff: If true, generate unified diff for differences
//   - verbose: If true, show detailed processing information
//
// Returns:
//   - *ComparisonResult: The comparison result
//   - error: Any error encountered during comparison
func CompareVersions(referenceFile, productDir string, versions []string, generateDiff bool, verbose bool) (*ComparisonResult, error) {
	if verbose {
		fmt.Printf("Comparing file across %d versions...\n", len(versions))
		fmt.Printf("  Reference file: %s\n", referenceFile)
		fmt.Printf("  Product directory: %s\n", productDir)
		fmt.Printf("  Versions: %v\n", versions)
	}

	// Extract the reference version from the path
	referenceVersion, err := ExtractVersionFromPath(referenceFile, productDir)
	if err != nil {
		return nil, fmt.Errorf("failed to extract version from reference file: %w", err)
	}

	if verbose {
		fmt.Printf("  Reference version: %s\n", referenceVersion)
	}

	// Read the reference file
	referenceContent, err := os.ReadFile(referenceFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read reference file %s: %w", referenceFile, err)
	}

	// Resolve version paths
	versionPaths, err := ResolveVersionPaths(referenceFile, productDir, versions)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve version paths: %w", err)
	}

	// Initialize result
	result := &ComparisonResult{
		ReferenceFile:    referenceFile,
		ReferenceVersion: referenceVersion,
		TotalFiles:       len(versionPaths),
	}

	// Compare each version
	for _, vp := range versionPaths {
		if verbose {
			fmt.Printf("  Checking %s: %s\n", vp.Version, vp.FilePath)
		}

		comparison := compareFile(referenceFile, string(referenceContent), vp, generateDiff, verbose)
		result.Comparisons = append(result.Comparisons, comparison)

		// Update counters
		switch comparison.Status {
		case FileMatches:
			result.MatchingFiles++
		case FileDiffers:
			result.DifferingFiles++
		case FileNotFound:
			result.NotFoundFiles++
		case FileError:
			result.ErrorFiles++
		}
	}

	return result, nil
}

// compareFile compares a single version file against the reference content.
//
// This is an internal helper function used by CompareVersions.
//
// Parameters:
//   - referencePath: Path to the reference file (for diff labels)
//   - referenceContent: Content of the reference file
//   - versionPath: The version path to compare
//   - generateDiff: If true, generate unified diff for differences
//   - verbose: If true, show detailed processing information
//
// Returns:
//   - FileComparison: The comparison result for this file
func compareFile(referencePath, referenceContent string, versionPath projectinfo.VersionPath, generateDiff bool, verbose bool) FileComparison {
	comparison := FileComparison{
		Version:  versionPath.Version,
		FilePath: versionPath.FilePath,
	}

	// Check if file exists
	if _, err := os.Stat(versionPath.FilePath); os.IsNotExist(err) {
		comparison.Status = FileNotFound
		if verbose {
			fmt.Printf("    → File not found\n")
		}
		return comparison
	}

	// Read the file
	content, err := os.ReadFile(versionPath.FilePath)
	if err != nil {
		comparison.Status = FileError
		comparison.Error = fmt.Errorf("failed to read file: %w", err)
		if verbose {
			fmt.Printf("    → Error reading file: %v\n", err)
		}
		return comparison
	}

	// Compare contents
	if AreFilesIdentical(referenceContent, string(content)) {
		comparison.Status = FileMatches
		if verbose {
			fmt.Printf("    → Matches\n")
		}
	} else {
		comparison.Status = FileDiffers
		if verbose {
			fmt.Printf("    → Differs\n")
		}

		if generateDiff {
			diff, err := GenerateDiff(referencePath, referenceContent, versionPath.FilePath, string(content))
			if err != nil {
				comparison.Status = FileError
				comparison.Error = fmt.Errorf("failed to generate diff: %w", err)
				if verbose {
					fmt.Printf("    → Error generating diff: %v\n", err)
				}
			} else {
				comparison.Diff = diff
			}
		}
	}

	return comparison
}

