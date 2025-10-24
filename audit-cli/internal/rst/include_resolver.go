package rst

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/mongodb/code-example-tooling/audit-cli/internal/pathresolver"
)

// IncludeDirectiveRegex matches .. include:: directives in RST files.
var IncludeDirectiveRegex = regexp.MustCompile(`^\.\.\s+include::\s+(.+)$`)

// FindIncludeDirectives finds all include directives in a file and resolves their paths.
//
// This function scans the file for .. include:: directives and resolves each path
// using MongoDB-specific conventions (steps files, extracts, template variables, etc.).
//
// Parameters:
//   - filePath: Path to the RST file to scan
//
// Returns:
//   - []string: List of resolved absolute paths to included files
//   - error: Any error encountered during scanning
func FindIncludeDirectives(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var includePaths []string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Check if this line is an include directive
		matches := IncludeDirectiveRegex.FindStringSubmatch(line)
		if len(matches) > 1 {
			includePath := strings.TrimSpace(matches[1])

			// Resolve the include path relative to the source directory
			resolvedPath, err := ResolveIncludePath(filePath, includePath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to resolve include path %s: %v\n", includePath, err)
				continue
			}

			includePaths = append(includePaths, resolvedPath)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return includePaths, nil
}

// ResolveIncludePath resolves an include path relative to the source directory
// Handles multiple special cases:
// - Template variables ({{var_name}})
// - Steps files (/includes/steps/name.rst -> /includes/steps-name.yaml)
// - Extracts files (ref-based YAML content blocks)
// - Release files (ref-based YAML content blocks)
// - Files without extensions (auto-append .rst)
func ResolveIncludePath(currentFilePath, includePath string) (string, error) {
	// Handle template variables by looking up replacements in the current file
	if strings.HasPrefix(includePath, "{{") && strings.HasSuffix(includePath, "}}") {
		// Extract the variable name
		varName := strings.TrimSuffix(strings.TrimPrefix(includePath, "{{"), "}}")
		varName = strings.TrimSpace(varName)

		// Try to resolve the variable from the current file's replacement section
		resolvedPath, err := ResolveTemplateVariable(currentFilePath, varName)
		if err != nil {
			return "", fmt.Errorf("failed to resolve template variable %s: %w", includePath, err)
		}

		// Now resolve the replacement path as a normal include
		includePath = resolvedPath
	}

	// Find the source directory by walking up from the current file
	sourceDir, err := pathresolver.FindSourceDirectory(currentFilePath)
	if err != nil {
		return "", err
	}

	// Clean the include path (remove leading slash if present)
	cleanIncludePath := strings.TrimPrefix(includePath, "/")

	// Special handling for steps/ includes
	// Convert /includes/steps/filename.rst to /includes/steps-filename.yaml
	if strings.Contains(cleanIncludePath, "steps/") {
		fullPath, err := resolveSpecialIncludePath(sourceDir, cleanIncludePath, "steps")
		if err == nil {
			return fullPath, nil
		}
		// If steps resolution fails, continue with normal resolution
	}

	// Special handling for extracts/ includes
	// These reference content blocks in YAML files by ref ID
	// Convert /includes/extracts/ref-name.rst to the YAML file containing that ref
	if strings.Contains(cleanIncludePath, "extracts/") {
		fullPath, err := resolveRefBasedIncludePath(sourceDir, cleanIncludePath, "extracts")
		if err == nil {
			return fullPath, nil
		}
		// If extracts resolution fails, continue with normal resolution
	}

	// Special handling for release/ includes
	// These also reference content blocks in YAML files by ref ID
	if strings.Contains(cleanIncludePath, "release/") {
		fullPath, err := resolveRefBasedIncludePath(sourceDir, cleanIncludePath, "release")
		if err == nil {
			return fullPath, nil
		}
		// If release resolution fails, continue with normal resolution
	}

	// Construct the full path
	fullPath := filepath.Join(sourceDir, cleanIncludePath)

	// If the file exists as-is, return it
	if _, err := os.Stat(fullPath); err == nil {
		return fullPath, nil
	}

	// If the path doesn't have an extension, try adding .rst
	if filepath.Ext(cleanIncludePath) == "" {
		fullPathWithRst := fullPath + ".rst"
		if _, err := os.Stat(fullPathWithRst); err == nil {
			return fullPathWithRst, nil
		}
	}

	return "", fmt.Errorf("include file not found: %s", fullPath)
}

// resolveSpecialIncludePath handles special include paths (steps/)
// Converts: /includes/steps/run-mongodb-on-a-linux-distribution-systemd.rst
// To:       /includes/steps-run-mongodb-on-a-linux-distribution-systemd.yaml
func resolveSpecialIncludePath(sourceDir, includePath, dirType string) (string, error) {
	// Find the "dirType/" part in the path (e.g., "steps/")
	searchPattern := dirType + "/"
	dirIndex := strings.Index(includePath, searchPattern)
	if dirIndex == -1 {
		return "", fmt.Errorf("no %s/ found in path", dirType)
	}

	// Split the path at "dirType/"
	beforeDir := includePath[:dirIndex]
	afterDir := includePath[dirIndex+len(searchPattern):]

	// Remove the file extension from afterDir
	afterDir = strings.TrimSuffix(afterDir, filepath.Ext(afterDir))

	// Construct the new path: before + "dirType-" + after + ".yaml"
	newPath := beforeDir + dirType + "-" + afterDir + ".yaml"

	// Construct the full path
	fullPath := filepath.Join(sourceDir, newPath)

	// Verify the file exists
	if _, err := os.Stat(fullPath); err != nil {
		return "", fmt.Errorf("%s file not found: %s", dirType, fullPath)
	}

	return fullPath, nil
}

// resolveRefBasedIncludePath handles ref-based include paths (extracts/, release/)
// These reference content blocks in YAML files by ref ID
// Example: /includes/extracts/install-mongodb-community-manually-redhat.rst
// References a ref in a YAML file like /includes/extracts-install-mongodb-manually.yaml
// Example: /includes/release/pin-repo-to-version-yum.rst
// References a ref in a YAML file like /includes/release-pinning.yaml
func resolveRefBasedIncludePath(sourceDir, includePath, dirType string) (string, error) {
	// Extract the ref name from the path
	// /includes/dirType/ref-name.rst -> ref-name
	searchPattern := dirType + "/"
	dirIndex := strings.Index(includePath, searchPattern)
	if dirIndex == -1 {
		return "", fmt.Errorf("no %s/ found in path", dirType)
	}

	refName := includePath[dirIndex+len(searchPattern):]
	refName = strings.TrimSuffix(refName, filepath.Ext(refName))

	// Get the directory part before "dirType/"
	beforeDir := includePath[:dirIndex]
	searchDir := filepath.Join(sourceDir, beforeDir)

	// Find all dirType-*.yaml files in the includes directory
	pattern := filepath.Join(searchDir, dirType+"-*.yaml")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return "", fmt.Errorf("failed to search for %s files: %w", dirType, err)
	}

	// Search each YAML file for the ref
	for _, yamlFile := range matches {
		hasRef, err := YAMLFileContainsRef(yamlFile, refName)
		if err != nil {
			continue // Skip files we can't read
		}
		if hasRef {
			return yamlFile, nil
		}
	}

	return "", fmt.Errorf("no %s file found containing ref: %s", dirType, refName)
}

// YAMLFileContainsRef checks if a YAML file contains a specific ref.
//
// This function scans a YAML file for a line matching "ref: <refName>".
// Used to find the correct YAML file for ref-based includes (extracts, release).
//
// Parameters:
//   - filePath: Path to the YAML file to check
//   - refName: The ref name to search for
//
// Returns:
//   - bool: True if the file contains the ref, false otherwise
//   - error: Any error encountered during scanning
func YAMLFileContainsRef(filePath, refName string) (bool, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return false, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	searchPattern := "ref: " + refName

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == searchPattern {
			return true, nil
		}
	}

	return false, scanner.Err()
}

// ResolveTemplateVariable resolves a template variable from a YAML file's replacement section.
//
// MongoDB documentation uses template variables in include directives like:
//   .. include:: {{release_specification_default}}
//
// These are resolved by looking up the variable in the YAML file's replacement section:
//   replacement:
//     release_specification_default: "/includes/release/install-windows-default.rst"
//
// Parameters:
//   - yamlFilePath: Path to the YAML file containing the replacement section
//   - varName: The variable name to resolve (without {{ }})
//
// Returns:
//   - string: The resolved path from the replacement section
//   - error: Any error encountered during resolution
func ResolveTemplateVariable(yamlFilePath, varName string) (string, error) {
	file, err := os.Open(yamlFilePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	inReplacementSection := false
	searchPattern := varName + ":"

	for scanner.Scan() {
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)

		// Check if we're entering the replacement section
		if trimmedLine == "replacement:" {
			inReplacementSection = true
			continue
		}

		// If we're in the replacement section
		if inReplacementSection {
			// Check if we've left the replacement section (new top-level key or document separator)
			if len(line) > 0 && line[0] != ' ' && line[0] != '\t' {
				// We've left the replacement section
				break
			}
			if trimmedLine == "..." || trimmedLine == "---" {
				// Document separator - we've left the replacement section
				break
			}

			// Look for our variable
			if strings.HasPrefix(trimmedLine, searchPattern) {
				// Extract the value (everything after "varName: ")
				value := strings.TrimPrefix(trimmedLine, searchPattern)
				value = strings.TrimSpace(value)
				// Remove quotes if present
				value = strings.Trim(value, "\"'")
				return value, nil
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return "", fmt.Errorf("template variable %s not found in replacement section of %s", varName, yamlFilePath)
}



