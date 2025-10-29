package filereferences

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

// AnalyzeReferences finds all files that reference the target file.
//
// This function searches through all RST files (.rst, .txt) and YAML files (.yaml, .yml)
// in the source directory to find files that reference the target file using include,
// literalinclude, or io-code-block directives. YAML files are included because extract
// and release files contain RST directives within their content blocks.
//
// By default, only content inclusion directives are searched. Set includeToctree to true
// to also search for toctree entries (navigation links).
//
// Parameters:
//   - targetFile: Absolute path to the file to analyze
//   - includeToctree: If true, include toctree entries in the search
//
// Returns:
//   - *ReferenceAnalysis: The analysis results
//   - error: Any error encountered during analysis
func AnalyzeReferences(targetFile string, includeToctree bool) (*ReferenceAnalysis, error) {
	// Get absolute path
	absTargetFile, err := filepath.Abs(targetFile)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Find the source directory
	sourceDir, err := pathresolver.FindSourceDirectory(absTargetFile)
	if err != nil {
		return nil, fmt.Errorf("failed to find source directory: %w", err)
	}

	// Initialize analysis result
	analysis := &ReferenceAnalysis{
		TargetFile:       absTargetFile,
		SourceDir:        sourceDir,
		ReferencingFiles: []FileReference{},
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

		// Search for references in this file
		refs, err := findReferencesInFile(path, absTargetFile, sourceDir, includeToctree)
		if err != nil {
			// Log error but continue processing other files
			fmt.Fprintf(os.Stderr, "Warning: failed to process %s: %v\n", path, err)
			return nil
		}

		// Add any found references
		analysis.ReferencingFiles = append(analysis.ReferencingFiles, refs...)

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk source directory: %w", err)
	}

	// Update total counts
	analysis.TotalReferences = len(analysis.ReferencingFiles)
	analysis.TotalFiles = countUniqueFiles(analysis.ReferencingFiles)

	return analysis, nil
}

// findReferencesInFile searches a single file for references to the target file.
//
// This function scans through the file line by line looking for include,
// literalinclude, and io-code-block directives that reference the target file.
// If includeToctree is true, also searches for toctree entries.
//
// Parameters:
//   - filePath: Path to the file to search
//   - targetFile: Absolute path to the target file
//   - sourceDir: Source directory (for resolving relative paths)
//   - includeToctree: If true, include toctree entries in the search
//
// Returns:
//   - []FileReference: List of references found in this file
//   - error: Any error encountered during processing
func findReferencesInFile(filePath, targetFile, sourceDir string, includeToctree bool) ([]FileReference, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var references []FileReference
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
				references = append(references, FileReference{
					FilePath:      filePath,
					DirectiveType: "include",
					ReferencePath: refPath,
					LineNumber:    lineNum,
				})
			}
			continue
		}

		// Check for literalinclude directive
		if matches := rst.LiteralIncludeDirectiveRegex.FindStringSubmatch(trimmedLine); matches != nil {
			refPath := strings.TrimSpace(matches[1])
			if referencesTarget(refPath, targetFile, sourceDir, filePath) {
				references = append(references, FileReference{
					FilePath:      filePath,
					DirectiveType: "literalinclude",
					ReferencePath: refPath,
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
					references = append(references, FileReference{
						FilePath:      filePath,
						DirectiveType: "io-code-block",
						ReferencePath: refPath,
						LineNumber:    ioCodeBlockStartLine,
					})
				}
				continue
			}

			// Check for output directive
			if matches := rst.OutputDirectiveRegex.FindStringSubmatch(trimmedLine); matches != nil {
				refPath := strings.TrimSpace(matches[1])
				if referencesTarget(refPath, targetFile, sourceDir, filePath) {
					references = append(references, FileReference{
						FilePath:      filePath,
						DirectiveType: "io-code-block",
						ReferencePath: refPath,
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
				references = append(references, FileReference{
					FilePath:      filePath,
					DirectiveType: "toctree",
					ReferencePath: docName,
					LineNumber:    toctreeStartLine,
				})
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return references, nil
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

// FilterByDirectiveType filters the analysis results to only include references
// of the specified directive type.
//
// Parameters:
//   - analysis: The original analysis results
//   - directiveType: The directive type to filter by (include, literalinclude, io-code-block)
//
// Returns:
//   - *ReferenceAnalysis: A new analysis with filtered results
func FilterByDirectiveType(analysis *ReferenceAnalysis, directiveType string) *ReferenceAnalysis {
	filtered := &ReferenceAnalysis{
		TargetFile:       analysis.TargetFile,
		SourceDir:        analysis.SourceDir,
		ReferencingFiles: []FileReference{},
		ReferenceTree:    analysis.ReferenceTree,
	}

	// Filter references
	for _, ref := range analysis.ReferencingFiles {
		if ref.DirectiveType == directiveType {
			filtered.ReferencingFiles = append(filtered.ReferencingFiles, ref)
		}
	}

	// Update counts
	filtered.TotalReferences = len(filtered.ReferencingFiles)
	filtered.TotalFiles = countUniqueFiles(filtered.ReferencingFiles)

	return filtered
}

// countUniqueFiles counts the number of unique files in the reference list.
//
// Parameters:
//   - refs: List of file references
//
// Returns:
//   - int: Number of unique files
func countUniqueFiles(refs []FileReference) int {
	uniqueFiles := make(map[string]bool)
	for _, ref := range refs {
		uniqueFiles[ref.FilePath] = true
	}
	return len(uniqueFiles)
}

// GroupReferencesByFile groups references by file path and directive type.
//
// This function takes a flat list of references and groups them by file,
// counting how many times each file references the target.
//
// Parameters:
//   - refs: List of file references
//
// Returns:
//   - []GroupedFileReference: List of grouped references, sorted by file path
func GroupReferencesByFile(refs []FileReference) []GroupedFileReference {
	// Group by file path and directive type
	type groupKey struct {
		filePath      string
		directiveType string
	}
	groups := make(map[groupKey][]FileReference)

	for _, ref := range refs {
		key := groupKey{ref.FilePath, ref.DirectiveType}
		groups[key] = append(groups[key], ref)
	}

	// Convert to slice
	var grouped []GroupedFileReference
	for key, refs := range groups {
		grouped = append(grouped, GroupedFileReference{
			FilePath:      key.filePath,
			DirectiveType: key.directiveType,
			References:    refs,
			Count:         len(refs),
		})
	}

	// Sort by file path for consistent output
	sort.Slice(grouped, func(i, j int) bool {
		return grouped[i].FilePath < grouped[j].FilePath
	})

	return grouped
}

