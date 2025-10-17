// Package rst provides utilities for parsing reStructuredText (RST) files.
//
// This package contains the core RST parsing logic used by the extract commands.
// It handles:
//   - Parsing RST directives (literalinclude, code-block, io-code-block)
//   - Following include directives recursively
//   - Resolving include paths with MongoDB-specific conventions
//   - Traversing directories for RST files
//
// The package is designed to be reusable across different extraction operations.
package rst

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// DirectiveType represents the type of reStructuredText directive.
type DirectiveType string

const (
	// CodeBlock represents inline code blocks (.. code-block::)
	CodeBlock DirectiveType = "code-block"
	// LiteralInclude represents external file references (.. literalinclude::)
	LiteralInclude DirectiveType = "literalinclude"
	// IoCodeBlock represents input/output examples (.. io-code-block::)
	IoCodeBlock DirectiveType = "io-code-block"
)

// Directive represents a parsed reStructuredText directive.
//
// Contains all information needed to extract content from the directive,
// including the directive type, arguments, options, and content.
type Directive struct {
	Type     DirectiveType     // Type of directive (code-block, literalinclude, io-code-block)
	Argument string            // Main argument (e.g., language for code-block, filepath for literalinclude)
	Options  map[string]string // Directive options (e.g., :language:, :start-after:, etc.)
	Content  string            // Content of the directive (for code-block and inline io-code-block)
	LineNum  int               // Line number where directive starts (1-based)

	// For io-code-block directives
	InputDirective  *SubDirective // The .. input:: nested directive
	OutputDirective *SubDirective // The .. output:: nested directive
}

// SubDirective represents a nested directive within io-code-block.
//
// Can contain either a filepath argument (for external file reference)
// or inline content (for embedded code).
type SubDirective struct {
	Argument string            // Filepath argument (if provided)
	Options  map[string]string // Directive options (e.g., :language:)
	Content  string            // Inline content (if no filepath)
}

// Regular expressions for directive parsing
var (
	// Matches: .. literalinclude:: /path/to/file.php
	literalIncludeRegex = regexp.MustCompile(`^\.\.\s+literalinclude::\s+(.+)$`)

	// Matches: .. code-block:: python (language is optional)
	codeBlockRegex = regexp.MustCompile(`^\.\.\s+code-block::\s*(.*)$`)

	// Matches: .. io-code-block::
	ioCodeBlockRegex = regexp.MustCompile(`^\.\.\s+io-code-block::\s*$`)

	// Matches: .. input:: /path/to/file.cs (filepath is optional)
	inputDirectiveRegex = regexp.MustCompile(`^\.\.\s+input::\s*(.*)$`)

	// Matches: .. output:: /path/to/file.txt (filepath is optional)
	outputDirectiveRegex = regexp.MustCompile(`^\.\.\s+output::\s*(.*)$`)

	// Matches directive options like:   :language: python
	optionRegex = regexp.MustCompile(`^\s+:([^:]+):\s*(.*)$`)
)

// ParseDirectives parses all directives from an RST file.
//
// This function scans the file line-by-line and extracts all supported directives
// (literalinclude, code-block, io-code-block). For each directive, it parses:
//   - The directive type and argument
//   - All directive options (e.g., :language:, :start-after:)
//   - The directive content (for code-block and io-code-block)
//   - Nested directives (for io-code-block)
//
// Parameters:
//   - filePath: Path to the RST file to parse
//
// Returns:
//   - []Directive: Slice of all parsed directives in order of appearance
//   - error: Any error encountered during parsing
func ParseDirectives(filePath string) ([]Directive, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var directives []Directive
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)

		// Check for literalinclude directive
		if matches := literalIncludeRegex.FindStringSubmatch(trimmedLine); len(matches) > 1 {
			directive := Directive{
				Type:     LiteralInclude,
				Argument: strings.TrimSpace(matches[1]),
				Options:  make(map[string]string),
				LineNum:  lineNum,
			}

			// Parse options on following lines
			parseDirectiveOptions(scanner, &directive, &lineNum)
			directives = append(directives, directive)
			continue
		}

		// Check for code-block directive
		if matches := codeBlockRegex.FindStringSubmatch(trimmedLine); len(matches) > 1 {
			directive := Directive{
				Type:     CodeBlock,
				Argument: strings.TrimSpace(matches[1]),
				Options:  make(map[string]string),
				LineNum:  lineNum,
			}

			// Parse options and content on following lines
			firstContentLine := parseDirectiveOptions(scanner, &directive, &lineNum)
			parseDirectiveContent(scanner, &directive, &lineNum, firstContentLine)
			directives = append(directives, directive)
			continue
		}

		// Check for io-code-block directive
		if ioCodeBlockRegex.MatchString(trimmedLine) {
			directive := Directive{
				Type:    IoCodeBlock,
				Options: make(map[string]string),
				LineNum: lineNum,
			}

			// Parse io-code-block with its nested input/output directives
			parseIoCodeBlock(scanner, &directive, &lineNum)
			directives = append(directives, directive)
			continue
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return directives, nil
}

// parseDirectiveOptions parses the options following a directive
// Returns the first content line if encountered, or empty string if not
func parseDirectiveOptions(scanner *bufio.Scanner, directive *Directive, lineNum *int) string {
	for scanner.Scan() {
		*lineNum++
		line := scanner.Text()

		// Check if this is an option line
		if matches := optionRegex.FindStringSubmatch(line); len(matches) > 1 {
			optionName := strings.TrimSpace(matches[1])
			optionValue := strings.TrimSpace(matches[2])
			directive.Options[optionName] = optionValue
			continue
		}

		// If we hit a blank line or non-indented line, we're done with options
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "" {
			continue // Skip blank lines between options and content
		}

		// If the line is not indented and not an option, we're done
		if len(line) > 0 && line[0] != ' ' && line[0] != '\t' {
			// Non-indented line means end of directive
			return ""
		}

		// If we have indented content (not an option), this is the start of content
		if len(line) > 0 && (line[0] == ' ' || line[0] == '\t') && !optionRegex.MatchString(line) {
			return line
		}
	}
	return ""
}

// parseDirectiveContent parses the content block of a directive (for code-block, io-code-block)
// firstContentLine is the first line of content (if already consumed by parseDirectiveOptions)
func parseDirectiveContent(scanner *bufio.Scanner, directive *Directive, lineNum *int, firstContentLine string) {
	var contentLines []string
	var baseIndent int = -1

	// Process the first content line if provided
	if firstContentLine != "" {
		// Calculate indentation
		indent := len(firstContentLine) - len(strings.TrimLeft(firstContentLine, " \t"))
		baseIndent = indent

		// Add the first line, removing the base indentation
		contentLines = append(contentLines, firstContentLine[baseIndent:])
	}

	for scanner.Scan() {
		*lineNum++
		line := scanner.Text()

		// Empty lines are part of the content
		if strings.TrimSpace(line) == "" {
			contentLines = append(contentLines, "")
			continue
		}

		// Calculate indentation
		indent := len(line) - len(strings.TrimLeft(line, " \t"))

		// If this is the first content line, establish the base indentation
		if baseIndent == -1 {
			baseIndent = indent
		}

		// If the line is less indented than the base, we're done with content
		if indent < baseIndent {
			break
		}

		// Add the line to content, removing the base indentation
		if indent >= baseIndent {
			contentLines = append(contentLines, line[baseIndent:])
		}
	}

	directive.Content = strings.TrimSpace(strings.Join(contentLines, "\n"))
}

// ExtractLiteralIncludeContent extracts content from a literalinclude directive
// Handles start-after and end-before options
func ExtractLiteralIncludeContent(currentFilePath string, directive Directive) (string, error) {
	if directive.Type != LiteralInclude {
		return "", fmt.Errorf("directive is not a literalinclude")
	}

	// Resolve the file path
	resolvedPath, err := ResolveIncludePath(currentFilePath, directive.Argument)
	if err != nil {
		return "", fmt.Errorf("failed to resolve literalinclude path %s: %w", directive.Argument, err)
	}

	// Read the file content
	content, err := os.ReadFile(resolvedPath)
	if err != nil {
		return "", fmt.Errorf("failed to read literalinclude file %s: %w", resolvedPath, err)
	}

	contentStr := string(content)

	// Handle start-after option
	if startAfter, hasStartAfter := directive.Options["start-after"]; hasStartAfter {
		startIdx := strings.Index(contentStr, startAfter)
		if startIdx == -1 {
			return "", fmt.Errorf("start-after tag '%s' not found in %s", startAfter, resolvedPath)
		}
		// Find the end of the line containing the start-after tag
		lineEnd := strings.Index(contentStr[startIdx:], "\n")
		if lineEnd == -1 {
			// Tag is on the last line, take everything after it
			contentStr = ""
		} else {
			// Skip past the newline to start at the next line
			contentStr = contentStr[startIdx+lineEnd+1:]
		}
	}

	// Handle end-before option
	if endBefore, hasEndBefore := directive.Options["end-before"]; hasEndBefore {
		endIdx := strings.Index(contentStr, endBefore)
		if endIdx == -1 {
			return "", fmt.Errorf("end-before tag '%s' not found in %s", endBefore, resolvedPath)
		}
		// Find the start of the line containing the end-before tag
		lineStart := strings.LastIndex(contentStr[:endIdx], "\n")
		if lineStart == -1 {
			lineStart = 0
		} else {
			lineStart++ // Move past the newline
		}
		// Cut before the line containing the tag, but keep the newline before it
		if lineStart > 0 {
			contentStr = contentStr[:lineStart-1]
		} else {
			contentStr = ""
		}
	}

	// Handle dedent option
	if _, hasDedent := directive.Options["dedent"]; hasDedent {
		contentStr = dedentContent(contentStr)
	}

	return strings.TrimSpace(contentStr), nil
}

// dedentContent removes common leading whitespace from all lines
func dedentContent(content string) string {
	lines := strings.Split(content, "\n")
	if len(lines) == 0 {
		return content
	}

	// Find the minimum indentation (ignoring empty lines)
	minIndent := -1
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		indent := len(line) - len(strings.TrimLeft(line, " \t"))
		if minIndent == -1 || indent < minIndent {
			minIndent = indent
		}
	}

	if minIndent <= 0 {
		return content
	}

	// Remove the common indentation from all lines
	var dedentedLines []string
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			dedentedLines = append(dedentedLines, "")
		} else if len(line) >= minIndent {
			dedentedLines = append(dedentedLines, line[minIndent:])
		} else {
			dedentedLines = append(dedentedLines, line)
		}
	}

	return strings.Join(dedentedLines, "\n")
}

// parseIoCodeBlock parses an io-code-block directive with its nested input/output directives
func parseIoCodeBlock(scanner *bufio.Scanner, directive *Directive, lineNum *int) {
	// First, parse any options for the io-code-block itself
	// This might return the first input/output directive line
	firstLine := parseDirectiveOptions(scanner, directive, lineNum)

	// Now parse the nested input and output directives
	var pendingLine string = firstLine
	for {
		var line string
		var trimmedLine string

		// Use pending line if we have one, otherwise scan for next line
		if pendingLine != "" {
			line = pendingLine
			trimmedLine = strings.TrimSpace(line)
			pendingLine = ""
		} else {
			if !scanner.Scan() {
				break
			}
			*lineNum++
			line = scanner.Text()
			trimmedLine = strings.TrimSpace(line)
		}

		// Stop if we hit a blank line followed by dedent to base level
		if trimmedLine == "" {
			// Peek ahead to see if next line is dedented
			if !scanner.Scan() {
				break
			}
			*lineNum++
			nextLine := scanner.Text()
			if len(nextLine) > 0 && nextLine[0] != ' ' && nextLine[0] != '\t' {
				// We've reached the end of the io-code-block
				break
			}
			// Not dedented, continue parsing
			line = nextLine
			trimmedLine = strings.TrimSpace(line)
		}

		// Check for input directive
		if matches := inputDirectiveRegex.FindStringSubmatch(trimmedLine); len(matches) > 0 {
			subDir := &SubDirective{
				Argument: strings.TrimSpace(matches[1]),
				Options:  make(map[string]string),
			}
			pendingLine = parseSubDirective(scanner, subDir, lineNum)
			directive.InputDirective = subDir
			continue
		}

		// Check for output directive
		if matches := outputDirectiveRegex.FindStringSubmatch(trimmedLine); len(matches) > 0 {
			subDir := &SubDirective{
				Argument: strings.TrimSpace(matches[1]),
				Options:  make(map[string]string),
			}
			pendingLine = parseSubDirective(scanner, subDir, lineNum)
			directive.OutputDirective = subDir
			continue
		}

		// If we get here, the line is neither input nor output directive
		// This means we've reached the end of the io-code-block
		break
	}
}

// parseSubDirective parses a nested directive (input or output) within io-code-block
// Returns the last line read (which might be the start of the next directive)
func parseSubDirective(scanner *bufio.Scanner, subDir *SubDirective, lineNum *int) string {
	var contentLines []string
	var baseIndent int = -1
	var lastLine string

	// Parse options and content
	for scanner.Scan() {
		*lineNum++
		line := scanner.Text()
		lastLine = line
		trimmedLine := strings.TrimSpace(line)

		// Empty line - might be part of content or end of directive
		if trimmedLine == "" {
			if len(contentLines) > 0 {
				contentLines = append(contentLines, "")
			}
			continue
		}

		// Check if this is an option line
		if matches := optionRegex.FindStringSubmatch(line); len(matches) > 2 {
			subDir.Options[matches[1]] = strings.TrimSpace(matches[2])
			continue
		}

		// Check if this is the start of another directive (input/output)
		if inputDirectiveRegex.MatchString(trimmedLine) || outputDirectiveRegex.MatchString(trimmedLine) {
			// Return this line so the caller can process it
			break
		}

		// Check if line is indented (content)
		if len(line) > 0 && (line[0] == ' ' || line[0] == '\t') {
			indent := len(line) - len(strings.TrimLeft(line, " \t"))

			// Set base indent from first content line
			if baseIndent == -1 {
				baseIndent = indent
			}

			// If we've dedented back to or past the base level, we're done
			if len(contentLines) > 0 && indent < baseIndent {
				break
			}

			// Add content line (remove base indentation)
			if baseIndent >= 0 && len(line) >= baseIndent {
				contentLines = append(contentLines, line[baseIndent:])
			} else {
				contentLines = append(contentLines, strings.TrimLeft(line, " \t"))
			}
		} else {
			// Non-indented, non-empty line means we're done with this directive
			break
		}
	}

	// Set the content
	if len(contentLines) > 0 {
		subDir.Content = strings.TrimSpace(strings.Join(contentLines, "\n"))
	}

	return lastLine
}

