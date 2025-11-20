package usage

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/mongodb/code-example-tooling/audit-cli/internal/pathresolver"
	"github.com/mongodb/code-example-tooling/audit-cli/internal/rst"
)

// AnalyzeUsage finds all files that use the target file.
//
// This function searches through all RST files (.rst, .txt) and YAML files (.yaml, .yml)
// in the source directory to find files that use the target file through include,
// literalinclude, or io-code-block directives. YAML files are included because extract
// and release files contain RST directives within their content blocks.
//
// By default, only content inclusion directives are searched. Set includeToctree to true
// to also search for toctree entries (navigation links).
//
// Parameters:
//   - targetFile: Absolute path to the file to analyze
//   - includeToctree: If true, include toctree entries in the search
//   - verbose: If true, show progress information
//   - excludePattern: Glob pattern for paths to exclude (empty string means no exclusion)
//
// Returns:
//   - *UsageAnalysis: The analysis results
//   - error: Any error encountered during analysis
func AnalyzeUsage(targetFile string, includeToctree bool, verbose bool, excludePattern string) (*UsageAnalysis, error) {
	// Check if target file exists
	if _, err := os.Stat(targetFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("target file does not exist: %s\n\nPlease check:\n  - The file path is correct\n  - The file hasn't been moved or deleted\n  - You have permission to access the file", targetFile)
	}

	// Get absolute path
	absTargetFile, err := filepath.Abs(targetFile)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Find the source directory
	sourceDir, err := pathresolver.FindSourceDirectory(absTargetFile)
	if err != nil {
		return nil, fmt.Errorf("failed to find source directory: %w\n\nThe source directory is detected by looking for a 'source' directory in the file's path.\nMake sure the target file is within a documentation repository with a 'source' directory.", err)
	}

	// Initialize analysis result
	analysis := &UsageAnalysis{
		TargetFile: absTargetFile,
		SourceDir:  sourceDir,
		UsingFiles: []FileUsage{},
	}

	// Track if we found any RST/YAML files
	foundAnyFiles := false
	filesProcessed := 0

	// Show progress message if verbose
	if verbose {
		fmt.Fprintf(os.Stderr, "Scanning for usages in %s...\n", sourceDir)
	}

	// Walk through all RST and YAML files in the source directory
	err = filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Only process RST files (.rst, .txt) and YAML files (.yaml, .yml)
		// YAML files may contain RST directives in extract/release content blocks
		ext := filepath.Ext(path)
		if ext != ".rst" && ext != ".txt" && ext != ".yaml" && ext != ".yml" {
			return nil
		}

		// Check if path should be excluded
		if excludePattern != "" {
			matched, err := filepath.Match(excludePattern, path)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: invalid exclude pattern: %v\n", err)
			} else if matched {
				// Skip this file
				return nil
			}
		}

		// Mark that we found at least one file
		foundAnyFiles = true
		filesProcessed++

		// Show progress every 100 files if verbose
		if verbose && filesProcessed%100 == 0 {
			fmt.Fprintf(os.Stderr, "Processed %d files...\n", filesProcessed)
		}

		// Search for usages in this file
		usages, err := findUsagesInFile(path, absTargetFile, sourceDir, includeToctree)
		if err != nil {
			// Log error but continue processing other files
			fmt.Fprintf(os.Stderr, "Warning: failed to process %s: %v\n", path, err)
			return nil
		}

		// Add any found usages
		analysis.UsingFiles = append(analysis.UsingFiles, usages...)

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk source directory: %w", err)
	}

	// Check if we found any RST/YAML files
	if !foundAnyFiles {
		return nil, fmt.Errorf("no RST or YAML files found in source directory: %s\n\nThis might not be a documentation repository.\nExpected to find files with extensions: .rst, .txt, .yaml, .yml", sourceDir)
	}

	// Show completion message if verbose
	if verbose {
		fmt.Fprintf(os.Stderr, "Scan complete. Processed %d files.\n", filesProcessed)
	}

	// Update total counts
	analysis.TotalUsages = len(analysis.UsingFiles)
	analysis.TotalFiles = countUniqueFiles(analysis.UsingFiles)

	return analysis, nil
}

// findUsagesInFile searches a single file for usages of the target file.
//
// This function scans through the file line by line looking for include,
// literalinclude, and io-code-block directives that use the target file.
// If includeToctree is true, also searches for toctree entries.
//
// Parameters:
//   - filePath: Path to the file to search
//   - targetFile: Absolute path to the target file
//   - sourceDir: Source directory (for resolving relative paths)
//   - includeToctree: If true, include toctree entries in the search
//
// Returns:
//   - []FileUsage: List of usages found in this file
//   - error: Any error encountered during processing
func findUsagesInFile(filePath, targetFile, sourceDir string, includeToctree bool) ([]FileUsage, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var usages []FileUsage
	scanner := bufio.NewScanner(file)
	lineNum := 0
	inIOCodeBlock := false
	ioCodeBlockStartLine := 0
	inToctree := false
	toctreeStartLine := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)

		// Check for toctree start (only if includeToctree is enabled)
		if includeToctree && rst.ToctreeDirectiveRegex.MatchString(trimmedLine) {
			inToctree = true
			toctreeStartLine = lineNum
			continue
		}

		// Check for io-code-block start
		if rst.IOCodeBlockDirectiveRegex.MatchString(trimmedLine) {
			inIOCodeBlock = true
			ioCodeBlockStartLine = lineNum
			continue
		}

		// Check if we're exiting toctree (unindented line that's not empty and not an option)
		if inToctree && len(line) > 0 && line[0] != ' ' && line[0] != '\t' {
			inToctree = false
		}

		// Check if we're exiting io-code-block (unindented line that's not empty)
		if inIOCodeBlock && len(line) > 0 && line[0] != ' ' && line[0] != '\t' {
			inIOCodeBlock = false
		}

		// Check for include directive
		if matches := rst.IncludeDirectiveRegex.FindStringSubmatch(trimmedLine); matches != nil {
			refPath := strings.TrimSpace(matches[1])
			if referencesTarget(refPath, targetFile, sourceDir, filePath) {
				usages = append(usages, FileUsage{
					FilePath:      filePath,
					DirectiveType: "include",
					UsagePath:     refPath,
					LineNumber:    lineNum,
				})
			}
			continue
		}

		// Check for literalinclude directive
		if matches := rst.LiteralIncludeDirectiveRegex.FindStringSubmatch(trimmedLine); matches != nil {
			refPath := strings.TrimSpace(matches[1])
			if referencesTarget(refPath, targetFile, sourceDir, filePath) {
				usages = append(usages, FileUsage{
					FilePath:      filePath,
					DirectiveType: "literalinclude",
					UsagePath:     refPath,
					LineNumber:    lineNum,
				})
			}
			continue
		}

		// Check for input/output directives within io-code-block
		if inIOCodeBlock {
			// Check for input directive
			if matches := rst.InputDirectiveRegex.FindStringSubmatch(trimmedLine); matches != nil {
				refPath := strings.TrimSpace(matches[1])
				if referencesTarget(refPath, targetFile, sourceDir, filePath) {
					usages = append(usages, FileUsage{
						FilePath:      filePath,
						DirectiveType: "io-code-block",
						UsagePath:     refPath,
						LineNumber:    ioCodeBlockStartLine,
					})
				}
				continue
			}

			// Check for output directive
			if matches := rst.OutputDirectiveRegex.FindStringSubmatch(trimmedLine); matches != nil {
				refPath := strings.TrimSpace(matches[1])
				if referencesTarget(refPath, targetFile, sourceDir, filePath) {
					usages = append(usages, FileUsage{
						FilePath:      filePath,
						DirectiveType: "io-code-block",
						UsagePath:     refPath,
						LineNumber:    ioCodeBlockStartLine,
					})
				}
				continue
			}
		}

		// Check for toctree entries (indented document names)
		if inToctree {
			// Skip empty lines and option lines (starting with :)
			if trimmedLine == "" || strings.HasPrefix(trimmedLine, ":") {
				continue
			}

			// This is a document name in the toctree
			// Document names can be relative or absolute (starting with /)
			docName := trimmedLine
			if referencesToctreeTarget(docName, targetFile, sourceDir, filePath) {
				usages = append(usages, FileUsage{
					FilePath:      filePath,
					DirectiveType: "toctree",
					UsagePath:     docName,
					LineNumber:    toctreeStartLine,
				})
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return usages, nil
}

// referencesTarget checks if a reference path points to the target file.
//
// This function resolves the reference path and compares it to the target file.
//
// Parameters:
//   - refPath: The path from the directive (e.g., "/includes/file.rst")
//   - targetFile: Absolute path to the target file
//   - sourceDir: Source directory (for resolving relative paths)
//   - currentFile: Path to the file containing the reference
//
// Returns:
//   - bool: true if the reference points to the target file
func referencesTarget(refPath, targetFile, sourceDir, currentFile string) bool {
	// Resolve the reference path
	var resolvedPath string

	if strings.HasPrefix(refPath, "/") {
		// Absolute path (relative to source directory)
		resolvedPath = filepath.Join(sourceDir, refPath)
	} else {
		// Relative path (relative to current file)
		currentDir := filepath.Dir(currentFile)
		resolvedPath = filepath.Join(currentDir, refPath)
	}

	// Clean and get absolute path
	resolvedPath = filepath.Clean(resolvedPath)
	absResolvedPath, err := filepath.Abs(resolvedPath)
	if err != nil {
		return false
	}

	// Compare with target file
	return absResolvedPath == targetFile
}

// referencesToctreeTarget checks if a toctree document name points to the target file.
//
// This function uses the shared rst.ResolveToctreePath to resolve the document name
// and then compares it to the target file.
//
// Parameters:
//   - docName: The document name from the toctree (e.g., "intro" or "/includes/intro")
//   - targetFile: Absolute path to the target file
//   - sourceDir: Source directory (for resolving relative paths)
//   - currentFile: Path to the file containing the toctree
//
// Returns:
//   - bool: true if the document name points to the target file
func referencesToctreeTarget(docName, targetFile, sourceDir, currentFile string) bool {
	// Use the shared toctree path resolution from rst package
	resolvedPath, err := rst.ResolveToctreePath(currentFile, docName)
	if err != nil {
		// If we can't resolve it, it doesn't match
		return false
	}

	// Compare with target file
	return resolvedPath == targetFile
}

// FilterByDirectiveType filters the analysis results to only include usages
// of the specified directive type.
//
// Parameters:
//   - analysis: The original analysis results
//   - directiveType: The directive type to filter by (include, literalinclude, io-code-block)
//
// Returns:
//   - *UsageAnalysis: A new analysis with filtered results
func FilterByDirectiveType(analysis *UsageAnalysis, directiveType string) *UsageAnalysis {
	filtered := &UsageAnalysis{
		TargetFile: analysis.TargetFile,
		SourceDir:  analysis.SourceDir,
		UsingFiles: []FileUsage{},
		UsageTree:  analysis.UsageTree,
	}

	// Filter usages
	for _, usage := range analysis.UsingFiles {
		if usage.DirectiveType == directiveType {
			filtered.UsingFiles = append(filtered.UsingFiles, usage)
		}
	}

	// Update counts
	filtered.TotalUsages = len(filtered.UsingFiles)
	filtered.TotalFiles = countUniqueFiles(filtered.UsingFiles)

	return filtered
}

// countUniqueFiles counts the number of unique files in the usage list.
//
// Parameters:
//   - usages: List of file usages
//
// Returns:
//   - int: Number of unique files
func countUniqueFiles(usages []FileUsage) int {
	uniqueFiles := make(map[string]bool)
	for _, usage := range usages {
		uniqueFiles[usage.FilePath] = true
	}
	return len(uniqueFiles)
}

// GroupUsagesByFile groups usages by file path and directive type.
//
// This function takes a flat list of usages and groups them by file,
// counting how many times each file uses the target.
//
// Parameters:
//   - usages: List of file usages
//
// Returns:
//   - []GroupedFileUsage: List of grouped usages, sorted by file path
func GroupUsagesByFile(usages []FileUsage) []GroupedFileUsage {
	// Group by file path and directive type
	type groupKey struct {
		filePath      string
		directiveType string
	}
	groups := make(map[groupKey][]FileUsage)

	for _, usage := range usages {
		key := groupKey{usage.FilePath, usage.DirectiveType}
		groups[key] = append(groups[key], usage)
	}

	// Convert to slice
	var grouped []GroupedFileUsage
	for key, usages := range groups {
		grouped = append(grouped, GroupedFileUsage{
			FilePath:      key.filePath,
			DirectiveType: key.directiveType,
			Usages:        usages,
			Count:         len(usages),
		})
	}

	// Sort by file path for consistent output
	sort.Slice(grouped, func(i, j int) bool {
		return grouped[i].FilePath < grouped[j].FilePath
	})

	return grouped
}

