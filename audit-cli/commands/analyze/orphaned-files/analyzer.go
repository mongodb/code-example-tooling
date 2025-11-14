package orphanedfiles

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mongodb/code-example-tooling/audit-cli/internal/rst"
)

// FindOrphanedFiles finds all files in the source directory that have no incoming references.
//
// This function scans all RST files (.rst, .txt) and YAML files (.yaml, .yml) in the source
// directory to build a complete reference map, then identifies files with zero incoming references.
//
// Parameters:
//   - sourceDir: Path to the source directory to scan
//   - includeToctree: If true, include toctree references when determining orphaned status
//   - verbose: If true, show progress information
//   - excludePattern: Glob pattern for paths to exclude (empty string means no exclusion)
//
// Returns:
//   - *OrphanedFilesAnalysis: The analysis results
//   - error: Any error encountered during analysis
func FindOrphanedFiles(sourceDir string, includeToctree bool, verbose bool, excludePattern string) (*OrphanedFilesAnalysis, error) {
	// Get absolute path
	absSourceDir, err := filepath.Abs(sourceDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Verify the directory exists
	if info, err := os.Stat(absSourceDir); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("source directory does not exist: %s\n\nPlease check:\n  - The directory path is correct\n  - The directory hasn't been moved or deleted\n  - You have permission to access the directory", absSourceDir)
		}
		return nil, fmt.Errorf("failed to access source directory: %w", err)
	} else if !info.IsDir() {
		return nil, fmt.Errorf("path is not a directory: %s", absSourceDir)
	}

	// Initialize tracking structures
	allFiles := make(map[string]bool)       // All RST/YAML files found
	referencedFiles := make(map[string]bool) // Files that have incoming references
	filesProcessed := 0

	// Show progress message if verbose
	if verbose {
		fmt.Fprintf(os.Stderr, "Scanning for files and references in %s...\n", absSourceDir)
	}

	// Walk through all RST and YAML files in the source directory
	err = filepath.Walk(absSourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Only process RST files (.rst, .txt) and YAML files (.yaml, .yml)
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

		// Add this file to the list of all files
		allFiles[path] = true
		filesProcessed++

		// Show progress every 100 files if verbose
		if verbose && filesProcessed%100 == 0 {
			fmt.Fprintf(os.Stderr, "Processed %d files...\n", filesProcessed)
		}

		// Scan this file for references to other files
		refs, err := findAllReferencesInFile(path, absSourceDir, includeToctree)
		if err != nil {
			// Log error but continue processing other files
			fmt.Fprintf(os.Stderr, "Warning: failed to process %s: %v\n", path, err)
			return nil
		}

		// Mark all referenced files
		for _, ref := range refs {
			referencedFiles[ref] = true
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk source directory: %w", err)
	}

	// Check if we found any RST/YAML files
	if len(allFiles) == 0 {
		return nil, fmt.Errorf("no RST or YAML files found in source directory: %s\n\nThis might not be a documentation repository.\nExpected to find files with extensions: .rst, .txt, .yaml, .yml", absSourceDir)
	}

	// Show completion message if verbose
	if verbose {
		fmt.Fprintf(os.Stderr, "Scan complete. Processed %d files.\n", filesProcessed)
		fmt.Fprintf(os.Stderr, "Found %d referenced files.\n", len(referencedFiles))
	}

	// Find orphaned files (files in allFiles but not in referencedFiles)
	orphanedFiles := []string{}
	for file := range allFiles {
		if !referencedFiles[file] {
			// Convert to relative path for cleaner output
			relPath, err := filepath.Rel(absSourceDir, file)
			if err != nil {
				relPath = file // Fall back to absolute path if relative fails
			}
			orphanedFiles = append(orphanedFiles, relPath)
		}
	}

	// Create analysis result
	analysis := &OrphanedFilesAnalysis{
		SourceDir:       absSourceDir,
		TotalFiles:      len(allFiles),
		TotalOrphaned:   len(orphanedFiles),
		OrphanedFiles:   orphanedFiles,
		IncludedToctree: includeToctree,
	}

	return analysis, nil
}

// findAllReferencesInFile finds all file references in a given file.
//
// This function scans the file for include, literalinclude, io-code-block (input/output),
// and optionally toctree directives, and returns the absolute paths of all referenced files.
//
// Parameters:
//   - filePath: Path to the file to scan
//   - sourceDir: Absolute path to the source directory
//   - includeToctree: If true, include toctree references
//
// Returns:
//   - []string: Slice of absolute paths to referenced files
//   - error: Any error encountered during scanning
func findAllReferencesInFile(filePath string, sourceDir string, includeToctree bool) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var references []string
	scanner := bufio.NewScanner(file)
	lineNum := 0
	inToctree := false
	inIOCodeBlock := false

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)

		// Check for toctree start (only if includeToctree is enabled)
		if includeToctree && rst.ToctreeDirectiveRegex.MatchString(trimmedLine) {
			inToctree = true
			continue
		}

		// Check for io-code-block start
		if rst.IOCodeBlockDirectiveRegex.MatchString(trimmedLine) {
			inIOCodeBlock = true
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
			if absPath := resolveReferencePath(refPath, sourceDir, filePath); absPath != "" {
				references = append(references, absPath)
			}
			continue
		}

		// Check for literalinclude directive
		if matches := rst.LiteralIncludeDirectiveRegex.FindStringSubmatch(trimmedLine); matches != nil {
			refPath := strings.TrimSpace(matches[1])
			if absPath := resolveReferencePath(refPath, sourceDir, filePath); absPath != "" {
				references = append(references, absPath)
			}
			continue
		}

		// Check for input/output directives within io-code-block
		if inIOCodeBlock {
			// Check for input directive
			if matches := rst.InputDirectiveRegex.FindStringSubmatch(trimmedLine); matches != nil {
				refPath := strings.TrimSpace(matches[1])
				if absPath := resolveReferencePath(refPath, sourceDir, filePath); absPath != "" {
					references = append(references, absPath)
				}
				continue
			}

			// Check for output directive
			if matches := rst.OutputDirectiveRegex.FindStringSubmatch(trimmedLine); matches != nil {
				refPath := strings.TrimSpace(matches[1])
				if absPath := resolveReferencePath(refPath, sourceDir, filePath); absPath != "" {
					references = append(references, absPath)
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
			docName := trimmedLine
			if absPath := resolveToctreePath(docName, sourceDir, filePath); absPath != "" {
				references = append(references, absPath)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return references, nil
}

// resolveReferencePath resolves a reference path to an absolute path.
//
// This function handles both absolute paths (starting with /) and relative paths.
//
// Parameters:
//   - refPath: The reference path from the directive
//   - sourceDir: Absolute path to the source directory
//   - currentFile: Absolute path to the file containing the reference
//
// Returns:
//   - string: Absolute path to the referenced file, or empty string if resolution fails
func resolveReferencePath(refPath string, sourceDir string, currentFile string) string {
	var absPath string

	if strings.HasPrefix(refPath, "/") {
		// Absolute path from source root
		absPath = filepath.Join(sourceDir, refPath[1:])
	} else {
		// Relative path from current file
		currentDir := filepath.Dir(currentFile)
		absPath = filepath.Join(currentDir, refPath)
	}

	// Clean the path
	absPath = filepath.Clean(absPath)

	// Verify the file exists (optional - we still want to track references even if file doesn't exist)
	// But we'll return empty string if the path is clearly invalid
	if _, err := os.Stat(absPath); err == nil {
		return absPath
	}

	// File doesn't exist, but we'll still return the path as it's a reference
	return absPath
}

// resolveToctreePath resolves a toctree document name to an absolute path.
//
// Toctree entries are document names without the .rst extension.
//
// Parameters:
//   - docName: The document name from the toctree
//   - sourceDir: Absolute path to the source directory
//   - currentFile: Absolute path to the file containing the toctree
//
// Returns:
//   - string: Absolute path to the referenced file, or empty string if resolution fails
func resolveToctreePath(docName string, sourceDir string, currentFile string) string {
	// Add .rst extension if not present
	if !strings.HasSuffix(docName, ".rst") && !strings.HasSuffix(docName, ".txt") {
		docName = docName + ".rst"
	}

	return resolveReferencePath(docName, sourceDir, currentFile)
}

