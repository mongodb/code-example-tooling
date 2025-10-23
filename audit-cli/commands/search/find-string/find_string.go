// Package find_string provides functionality for searching code example files for substrings.
//
// This package implements the "search find-string" subcommand, which searches through
// extracted code example files to find occurrences of a specific substring.
//
// By default, the search is case-insensitive and matches exact words only (not partial matches
// within larger words). These behaviors can be changed with the --case-sensitive and
// --partial-match flags. Each file is counted only once, even if the substring appears
// multiple times in the same file.
//
// Supports:
//   - Recursive directory scanning
//   - Following include directives in RST files
//   - Verbose output with file paths and language breakdown
//   - Language detection based on file extension
//   - Case-insensitive search (default) or case-sensitive search (--case-sensitive flag)
//   - Exact word matching (default) or partial matching (--partial-match flag)
package find_string

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mongodb/code-example-tooling/audit-cli/internal/rst"
	"github.com/spf13/cobra"
)

// NewFindStringCommand creates the find-string subcommand.
//
// This command searches through extracted code example files for a specific substring.
// Supports flags for recursive search, following includes, and verbose output.
//
// Flags:
//   - -r, --recursive: Recursively search all files in subdirectories
//   - -f, --follow-includes: Follow .. include:: directives in RST files
//   - -v, --verbose: Show file paths and language breakdown
//   - --case-sensitive: Make search case-sensitive (default: case-insensitive)
//   - --partial-match: Allow partial matches within words (default: exact word matching)
func NewFindStringCommand() *cobra.Command {
	var (
		recursive      bool
		followIncludes bool
		verbose        bool
		caseSensitive  bool
		partialMatch   bool
	)

	cmd := &cobra.Command{
		Use:   "find-string [filepath] [substring]",
		Short: "Search for a substring in extracted code example files",
		Long: `Search through extracted code example files to find occurrences of a specific substring.
Reports the number of code examples containing the substring.

By default, the search is case-insensitive and matches exact words only. Use --case-sensitive
to make the search case-sensitive, or --partial-match to allow matching the substring as part
of larger words (e.g., "curl" matching "libcurl").`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			filePath := args[0]
			substring := args[1]
			return runSearch(filePath, substring, recursive, followIncludes, verbose, caseSensitive, partialMatch)
		},
	}

	cmd.Flags().BoolVarP(&recursive, "recursive", "r", false, "Recursively search all files in subdirectories")
	cmd.Flags().BoolVarP(&followIncludes, "follow-includes", "f", false, "Follow .. include:: directives in RST files")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Provide additional information during execution")
	cmd.Flags().BoolVar(&caseSensitive, "case-sensitive", false, "Make search case-sensitive (default: case-insensitive)")
	cmd.Flags().BoolVar(&partialMatch, "partial-match", false, "Allow partial matches within words (default: exact word matching)")

	return cmd
}

// RunSearch executes the search operation and returns the report.
//
// This function is exported for use in tests. It searches for the substring in the
// specified file or directory and returns statistics about the search.
//
// Parameters:
//   - filePath: Path to file or directory to search
//   - substring: The substring to search for
//   - recursive: If true, recursively search subdirectories
//   - followIncludes: If true, follow .. include:: directives
//   - verbose: If true, show detailed information during search
//   - caseSensitive: If true, search is case-sensitive; if false, case-insensitive
//   - partialMatch: If true, allow partial matches within words; if false, match exact words only
//
// Returns:
//   - *SearchReport: Statistics about the search operation
//   - error: Any error encountered during search
func RunSearch(filePath string, substring string, recursive bool, followIncludes bool, verbose bool, caseSensitive bool, partialMatch bool) (*SearchReport, error) {
	return runSearchInternal(filePath, substring, recursive, followIncludes, verbose, caseSensitive, partialMatch)
}

// runSearch executes the search operation (internal wrapper for CLI).
//
// This is a thin wrapper around runSearchInternal that discards the report
// and only returns errors, suitable for use in the CLI command handler.
func runSearch(filePath string, substring string, recursive bool, followIncludes bool, verbose bool, caseSensitive bool, partialMatch bool) error {
	_, err := runSearchInternal(filePath, substring, recursive, followIncludes, verbose, caseSensitive, partialMatch)
	return err
}

// runSearchInternal contains the core logic for the search-code-examples command
func runSearchInternal(filePath string, substring string, recursive bool, followIncludes bool, verbose bool, caseSensitive bool, partialMatch bool) (*SearchReport, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to access path %s: %w", filePath, err)
	}

	report := NewSearchReport()

	var filesToSearch []string

	if fileInfo.IsDir() {
		if verbose {
			fmt.Printf("Scanning directory: %s (recursive: %v)\n", filePath, recursive)
		}
		filesToSearch, err = collectFiles(filePath, recursive)
		if err != nil {
			return nil, fmt.Errorf("failed to traverse directory: %w", err)
		}
	} else {
		filesToSearch = []string{filePath}
	}

	if verbose {
		fmt.Printf("Found %d files to search\n", len(filesToSearch))
		fmt.Printf("Searching for substring: %q\n", substring)
		fmt.Printf("Case sensitive: %v\n", caseSensitive)
		fmt.Printf("Partial match: %v\n", partialMatch)
		fmt.Printf("Follow includes: %v\n\n", followIncludes)
	}

	// Track visited files to prevent circular includes
	visited := make(map[string]bool)

	for _, file := range filesToSearch {
		if verbose {
			fmt.Printf("Searching: %s\n", file)
		}

		// If followIncludes is enabled, collect all files including those referenced by includes
		var filesToSearchWithIncludes []string
		if followIncludes {
			// Use ParseFileWithIncludes to get all files (main + includes)
			processedFiles, err := collectFilesWithIncludes(file, visited, verbose)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to follow includes for %s: %v\n", file, err)
				filesToSearchWithIncludes = []string{file}
			} else {
				filesToSearchWithIncludes = processedFiles
			}
		} else {
			filesToSearchWithIncludes = []string{file}
		}

		// Search all collected files
		for _, fileToSearch := range filesToSearchWithIncludes {
			result, err := searchFile(fileToSearch, substring, caseSensitive, partialMatch)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to search %s: %v\n", fileToSearch, err)
				continue
			}

			report.AddResult(result)

			if verbose && result.Contains {
				fmt.Printf("  ✓ Found substring in %s\n", fileToSearch)
			}
		}
	}

	PrintReport(report, verbose)

	return report, nil
}

// collectFiles gathers all files to search
func collectFiles(dirPath string, recursive bool) ([]string, error) {
	var files []string

	if recursive {
		err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
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
		entries, err := os.ReadDir(dirPath)
		if err != nil {
			return nil, err
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				files = append(files, filepath.Join(dirPath, entry.Name()))
			}
		}
	}

	return files, nil
}

// collectFilesWithIncludes collects a file and all files it includes via .. include:: directives
func collectFilesWithIncludes(filePath string, visited map[string]bool, verbose bool) ([]string, error) {
	// Use the RST package's ParseFileWithIncludes to get all files
	// We pass a no-op parseFunc since we just want the list of files
	processedFiles, err := rst.ParseFileWithIncludes(
		filePath,
		true, // followIncludes = true
		visited,
		verbose,
		nil, // no-op parseFunc
	)
	if err != nil {
		return nil, err
	}

	return processedFiles, nil
}

// searchFile searches a single file for the substring
func searchFile(filePath string, substring string, caseSensitive bool, partialMatch bool) (SearchResult, error) {
	result := SearchResult{
		FilePath: filePath,
		Language: extractLanguageFromFilename(filePath),
		Contains: false,
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return result, err
	}

	contentStr := string(content)
	searchStr := substring

	// Handle case sensitivity
	if !caseSensitive {
		contentStr = strings.ToLower(contentStr)
		searchStr = strings.ToLower(searchStr)
	}

	// Check if substring exists in content
	if !strings.Contains(contentStr, searchStr) {
		return result, nil
	}

	// If partial match is allowed, we're done
	if partialMatch {
		result.Contains = true
		return result, nil
	}

	// For exact word matching, check if the match is a whole word
	result.Contains = isExactWordMatch(contentStr, searchStr)

	return result, nil
}

// isExactWordMatch checks if the substring appears as a complete word in the content.
// A word boundary is defined as the start/end of the string or a non-alphanumeric character.
func isExactWordMatch(content string, substring string) bool {
	// Find all occurrences of the substring
	index := 0
	for {
		pos := strings.Index(content[index:], substring)
		if pos == -1 {
			break
		}

		actualPos := index + pos

		// Check if this is a whole word match
		// Check character before (if not at start)
		beforeOK := actualPos == 0 || !isWordChar(rune(content[actualPos-1]))

		// Check character after (if not at end)
		afterPos := actualPos + len(substring)
		afterOK := afterPos >= len(content) || !isWordChar(rune(content[afterPos]))

		if beforeOK && afterOK {
			return true
		}

		// Move to next potential match
		index = actualPos + 1
	}

	return false
}

// isWordChar returns true if the character is alphanumeric or underscore.
// These characters are considered part of a word.
func isWordChar(c rune) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_'
}

// extractLanguageFromFilename extracts the language from the file extension
func extractLanguageFromFilename(filePath string) string {
	ext := filepath.Ext(filePath)
	if ext == "" {
		return "unknown"
	}
	// Remove the leading dot
	return strings.TrimPrefix(ext, ".")
}
