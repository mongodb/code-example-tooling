// Package rst provides parsing and analysis of reStructuredText (RST) procedures
// from MongoDB documentation.
//
// # What is a Procedure?
//
// A procedure is a set of sequential steps that guide users through a task. In MongoDB
// documentation, procedures can be implemented in several formats:
//
//  1. Procedure Directive: Using .. procedure:: and .. step:: directives
//  2. Ordered Lists: Using numbered or lettered lists (1., 2., 3. or a., b., c.)
//  3. YAML Steps Files: Using .yaml files with a steps: array (converted to procedures during build)
//
// # Procedure Variations
//
// MongoDB documentation uses procedures inconsistently across different contexts (drivers,
// deployment methods, platforms, etc.). This parser handles three mechanisms for representing
// procedure variations:
//
// 1. Composable Tutorials with Selected Content Blocks
//
// A composable tutorial wraps a procedure and defines variations using selected-content blocks:
//
//	.. composable-tutorial::
//	   :options: driver, atlas-cli
//	   :defaults: driver=nodejs; atlas-cli=none
//
//	   .. procedure::
//	      .. step:: Install dependencies
//	         .. selected-content::
//	            :selections: driver=nodejs
//	            npm install mongodb
//	         .. selected-content::
//	            :selections: driver=python
//	            pip install pymongo
//
// This creates variations like "driver=nodejs" and "driver=python" with different content
// for the same logical step.
//
// 2. Tabs Within Steps
//
// Tabs can appear within procedure steps to show different ways to accomplish the same task:
//
//	.. procedure::
//	   .. step:: Connect to MongoDB
//	      .. tabs::
//	         .. tab:: Node.js
//	            :tabid: nodejs
//	            const client = new MongoClient(uri);
//	         .. tab:: Python
//	            :tabid: python
//	            client = MongoClient(uri)
//
// This creates variations "nodejs" and "python" for the same procedure.
//
// 3. Tabs Containing Procedures
//
// Tabs can contain entirely different procedures for different platforms/contexts:
//
//	Installation Instructions
//	--------------------------
//	.. tabs::
//	   .. tab:: macOS
//	      :tabid: macos
//	      .. procedure::
//	         .. step:: Install Homebrew
//	   .. tab:: Windows
//	      :tabid: windows
//	      .. procedure::
//	         .. step:: Download the installer
//
// This creates separate procedures that are grouped for analysis but extracted separately.
//
// # Include Directive Expansion
//
// The parser handles .. include:: directives with special logic:
//
//   - If a file has NO composable tutorial: Expands all includes globally before parsing
//   - If a file HAS a composable tutorial: Expands includes within selected-content blocks
//     and within procedure steps to detect selected-content blocks in included files
//
// This ensures that variations defined in included files are properly detected.
//
// # Uniqueness and Grouping
//
// Procedures are identified by their heading (title) and content hash. The content hash
// includes step titles, content, and variations to detect when procedures are identical
// vs. different.
//
// For Analysis/Reporting:
//   - Procedures with the same TabSet are grouped as one logical procedure
//   - Procedures with the same ComposableTutorial selections are grouped together
//   - Shows "1 unique procedure with N variations"
//
// For Extraction:
//   - Each unique procedure (by content hash) is extracted to a separate file
//   - Tabs containing procedures: Each tab's procedure is extracted separately
//   - Composable tutorials: One file per unique procedure, listing all selections
//   - Tabs within steps: One file listing all tab variations
//
// # Key Design Decisions
//
//  1. Deterministic Ordering: All map iterations are sorted to ensure consistent output
//  2. Content Hashing: SHA256 hash of step content to detect identical procedures
//  3. Grouping Semantics: Same procedures grouped differently for analysis vs. extraction
//  4. Include Expansion: Context-aware expansion to detect variations in included files
//
// For detailed examples and edge cases, see docs/PROCEDURE_PARSING.md
package rst

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// ParseProceduresWithOptions parses all procedures from an RST file with options.
//
// This function scans the file and extracts all procedures, whether they are
// implemented using .. procedure:: directives, ordered lists, or composable tutorials.
//
// Parameters:
//   - filePath: Path to the RST file to parse
//   - expandIncludes: If true, expands .. include:: directives inline to detect variations
//
// Returns:
//   - []Procedure: Slice of all parsed procedures
//   - error: Any error encountered during parsing
func ParseProceduresWithOptions(filePath string, expandIncludes bool) ([]Procedure, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(content), "\n")

	// Check if the file contains composable tutorials
	hasComposableTutorial := false
	for _, line := range lines {
		if ComposableTutorialDirectiveRegex.MatchString(strings.TrimSpace(line)) {
			hasComposableTutorial = true
			break
		}
	}

	// If expandIncludes is true AND there are no composable tutorials, expand all include directives inline
	// If there ARE composable tutorials, we DON'T expand includes globally because each selected-content
	// block will expand its own includes to preserve the block boundaries
	if expandIncludes && !hasComposableTutorial {
		lines, err = expandIncludesInLines(filePath, lines)
		if err != nil {
			return nil, fmt.Errorf("failed to expand includes: %w", err)
		}
	}

	return parseProceduresFromLines(lines, filePath)
}

// parseProceduresFromLines parses procedures from a slice of lines.
func parseProceduresFromLines(lines []string, filePath string) ([]Procedure, error) {

	var procedures []Procedure
	var currentHeading string

	i := 0
	for i < len(lines) {
		line := lines[i]
		trimmedLine := strings.TrimSpace(line)

		// Track headings for procedure titles
		if i+1 < len(lines) {
			nextLine := strings.TrimSpace(lines[i+1])
			if isHeadingUnderline(nextLine) && len(nextLine) >= len(trimmedLine) {
				// Skip empty headings and generic headings that don't provide meaningful context
				headingLower := strings.ToLower(trimmedLine)

				// Check if this is a "Procedure" heading
				if headingLower == "procedure" || headingLower == "steps" {
					currentHeading = trimmedLine
					i += 2 // Skip heading and underline

					// Look ahead to see if the next heading is numbered
					// If so, parse as hierarchical procedure
					j := i
					for j < len(lines) && strings.TrimSpace(lines[j]) == "" {
						j++
					}
					if j+1 < len(lines) {
						nextHeading := strings.TrimSpace(lines[j])
						nextUnderline := strings.TrimSpace(lines[j+1])
						if isHeadingUnderline(nextUnderline) && isNumberedHeading(nextHeading) {
							// Parse hierarchical procedure
							procedure, endLine := parseHierarchicalProcedure(lines, j, currentHeading)
							if len(procedure.Steps) > 0 {
								procedure.LineNum = i - 1 // Line where "Procedure" heading starts
								procedure.EndLineNum = endLine + 1
								procedures = append(procedures, procedure)
							}
							i = endLine + 1
							continue
						}
					}
					continue
				}

				if trimmedLine != "" && headingLower != "overview" {
					currentHeading = trimmedLine
				}
				i += 2 // Skip heading and underline
				continue
			}
		}

		// Check for composable tutorial directive
		if ComposableTutorialDirectiveRegex.MatchString(trimmedLine) {
			tutorial, endLine := parseComposableTutorial(lines, i, currentHeading, filePath)
			if tutorial != nil {
				// Extract ALL procedures from this composable tutorial
				tutorialProcs := extractProceduresFromComposableTutorial(tutorial, i)
				procedures = append(procedures, tutorialProcs...)
			}
			i = endLine + 1
			continue
		}

		// Check for tabs directive at top level (tabs containing procedures)
		if TabsDirectiveRegex.MatchString(trimmedLine) {
			tabSet, endLine := parseTabSetWithProcedures(lines, i, currentHeading, filePath)
			if tabSet != nil && len(tabSet.Procedures) > 0 {
				// Extract procedures from the tab set
				tabProcs := extractProceduresFromTabSet(tabSet)
				procedures = append(procedures, tabProcs...)
			}
			i = endLine + 1
			continue
		}

		// Check for procedure directive
		if ProcedureDirectiveRegex.MatchString(trimmedLine) {
			procedure, endLine := parseProcedureDirectiveFromLines(lines, i, currentHeading, filePath)
			procedure.LineNum = i + 1
			procedure.EndLineNum = endLine + 1
			procedures = append(procedures, procedure)
			i = endLine + 1
			continue
		}

		// Check for ordered list (potential procedure)
		if isOrderedListStart(trimmedLine) {
			procedure, endLine := parseOrderedListProcedure(lines, i, currentHeading)
			if len(procedure.Steps) > 0 {
				procedure.LineNum = i + 1
				procedure.EndLineNum = endLine + 1
				procedures = append(procedures, procedure)
			}
			i = endLine + 1
			continue
		}

		i++
	}

	// Sort procedures by line number for deterministic order
	sort.Slice(procedures, func(i, j int) bool {
		return procedures[i].LineNum < procedures[j].LineNum
	})

	return procedures, nil
}

// isHeading checks if the current line is part of a heading (checks next line for underline)
func isHeading(lines []string, idx int) bool {
	if idx+1 >= len(lines) {
		return false
	}
	nextLine := strings.TrimSpace(lines[idx+1])
	return isHeadingUnderline(nextLine)
}

// isHeadingUnderline checks if a line is a heading underline
func isHeadingUnderline(line string) bool {
	if len(line) == 0 {
		return false
	}
	// RST headings are underlined with =, -, ~, ^, ", `, +, etc.
	firstChar := line[0]
	underlineChars := "=-~`^\"'+*#"
	if !strings.ContainsRune(underlineChars, rune(firstChar)) {
		return false
	}
	// Check if entire line is the same character
	for _, ch := range line {
		if ch != rune(firstChar) {
			return false
		}
	}
	return true
}

// isNumberedHeading checks if a heading starts with a number followed by a period
func isNumberedHeading(heading string) bool {
	trimmed := strings.TrimSpace(heading)
	if len(trimmed) < 3 {
		return false
	}
	// Check if it starts with a digit followed by a period
	if trimmed[0] >= '0' && trimmed[0] <= '9' {
		// Find the period
		for i := 1; i < len(trimmed); i++ {
			if trimmed[i] == '.' {
				return true
			}
			if trimmed[i] < '0' || trimmed[i] > '9' {
				return false
			}
		}
	}
	return false
}

// parseHierarchicalProcedure parses a procedure with numbered headings as steps
// This handles the pattern where a "Procedure" heading is followed by numbered headings
// like "1. First Step", "2. Second Step", etc.
func parseHierarchicalProcedure(lines []string, startIdx int, title string) (Procedure, int) {
	procedure := Procedure{
		Type:  OrderedList,
		Title: title,
		Steps: []Step{},
	}

	i := startIdx

	// Parse each numbered heading as a step
	for i < len(lines) {
		line := lines[i]
		trimmedLine := strings.TrimSpace(line)

		// Empty line
		if trimmedLine == "" {
			i++
			continue
		}

		// Check if this is a numbered heading
		if i+1 < len(lines) {
			nextLine := strings.TrimSpace(lines[i+1])
			if isHeadingUnderline(nextLine) && len(nextLine) >= len(trimmedLine) {
				if isNumberedHeading(trimmedLine) {
					// Parse this numbered heading as a step
					step, endLine := parseNumberedHeadingStep(lines, i)
					procedure.Steps = append(procedure.Steps, step)
					i = endLine + 1
					continue
				} else {
					// Non-numbered heading - end of this procedure
					break
				}
			}
		}

		// Check for directive or other content that signals end of procedure
		if strings.HasPrefix(trimmedLine, "..") {
			break
		}

		i++
	}

	// Check for sub-procedures
	for _, step := range procedure.Steps {
		if len(step.SubProcedures) > 0 {
			procedure.HasSubSteps = true
			break
		}
	}

	return procedure, i - 1
}

// parseNumberedHeadingStep parses a numbered heading and its content as a procedure step
func parseNumberedHeadingStep(lines []string, startIdx int) (Step, int) {
	heading := strings.TrimSpace(lines[startIdx])
	_ = strings.TrimSpace(lines[startIdx+1]) // underline (not used but needed to skip)

	step := Step{
		Title:   heading,
		LineNum: startIdx + 1,
	}

	i := startIdx + 2 // Skip heading and underline
	var contentLines []string
	var subProcedures []SubProcedure

	// Parse the content under this heading
	for i < len(lines) {
		line := lines[i]
		trimmedLine := strings.TrimSpace(line)

		// Empty line
		if trimmedLine == "" {
			contentLines = append(contentLines, "")
			i++
			continue
		}

		// Check if we've hit the next numbered heading
		if i+1 < len(lines) {
			nextLine := strings.TrimSpace(lines[i+1])
			if isHeadingUnderline(nextLine) && len(nextLine) >= len(trimmedLine) {
				// This is a heading - check if it's numbered (next step) or a subheading
				if isNumberedHeading(trimmedLine) {
					// Next numbered step - we're done with this step
					break
				}
				// Non-numbered heading - could be a subheading, include it in content
			}
		}

		// Check for ordered list (sub-steps)
		if isOrderedListStart(trimmedLine) {
			subProcedureSteps, listType, endLine := parseOrderedListSteps(lines, i)
			// Add as a separate sub-procedure with its list type
			subProcedures = append(subProcedures, SubProcedure{
				Steps:    subProcedureSteps,
				ListType: listType,
			})
			// Add the sub-steps to content as well
			for j := i; j <= endLine; j++ {
				contentLines = append(contentLines, lines[j])
			}
			i = endLine + 1
			continue
		}

		// Check for directive
		if strings.HasPrefix(trimmedLine, "..") {
			// Include directives in content
			contentLines = append(contentLines, line)
			i++
			continue
		}

		// Regular content line
		contentLines = append(contentLines, line)
		i++
	}

	step.Content = strings.Join(contentLines, "\n")
	step.SubProcedures = subProcedures

	return step, i - 1
}


// computeProcedureContentHash generates a hash of the procedure's content
// to detect when procedures are identical across different selections
func computeProcedureContentHash(proc *Procedure) string {
	var content strings.Builder

	// Include all step titles and content
	for _, step := range proc.Steps {
		content.WriteString(step.Title)
		content.WriteString("|")
		content.WriteString(step.Content)
		content.WriteString("|")

		// Include variations
		for _, variation := range step.Variations {
			content.WriteString(string(variation.Type))
			content.WriteString("|")
			for _, opt := range variation.Options {
				content.WriteString(opt)
				content.WriteString("|")
			}
			// Sort keys for deterministic hash
			var keys []string
			for key := range variation.Content {
				keys = append(keys, key)
			}
			sort.Strings(keys)
			for _, key := range keys {
				content.WriteString(key)
				content.WriteString(":")
				content.WriteString(variation.Content[key])
				content.WriteString("|")
			}
		}

		// Include sub-procedures
		for _, subProc := range step.SubProcedures {
			content.WriteString(subProc.ListType)
			content.WriteString("|")
			for _, substep := range subProc.Steps {
				content.WriteString(substep.Title)
				content.WriteString("|")
				content.WriteString(substep.Content)
				content.WriteString("|")
			}
		}
	}

	// Compute SHA256 hash
	hash := sha256.Sum256([]byte(content.String()))
	return hex.EncodeToString(hash[:])
}

// isOrderedListStart checks if a line starts an ordered list
func isOrderedListStart(line string) bool {
	return numberedListRegex.MatchString(line) || letteredListRegex.MatchString(line) || continuationMarkerRegex.MatchString(line)
}

// getIndentLevel returns the indentation level of a line
func getIndentLevel(line string) int {
	count := 0
	for _, ch := range line {
		if ch == ' ' {
			count++
		} else if ch == '\t' {
			count += 4 // Treat tab as 4 spaces
		} else {
			break
		}
	}
	return count
}

// expandIncludesInLines expands all .. include:: directives in the lines.
//
// This function recursively processes include directives, replacing them with
// the content of the included files. The included content is indented to match
// the indentation of the include directive.
//
// Special handling for YAML steps files: When a .yaml steps file is encountered,
// it's converted to a placeholder procedure directive so it can be detected as a procedure.
//
// Parameters:
//   - filePath: Path to the file being parsed (for resolving relative includes)
//   - lines: The lines to process
//
// Returns:
//   - []string: Lines with includes expanded
//   - error: Any error encountered during expansion
func expandIncludesInLines(filePath string, lines []string) ([]string, error) {
	var result []string
	visited := make(map[string]bool) // Track visited files to prevent circular includes

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		trimmedLine := strings.TrimSpace(line)

		// Check if this is an include directive
		if matches := IncludeDirectiveRegex.FindStringSubmatch(trimmedLine); len(matches) > 1 {
			includePath := strings.TrimSpace(matches[1])

			// Resolve the include path
			resolvedPath, err := ResolveIncludePath(filePath, includePath)
			if err != nil {
				// If we can't resolve the include, keep the directive as-is
				result = append(result, line)
				continue
			}

			// Check for circular includes
			if visited[resolvedPath] {
				result = append(result, line)
				continue
			}
			visited[resolvedPath] = true

			// Get the indentation of the include directive
			indent := getIndentLevel(line)

			// Special handling for YAML steps files
			// These are procedures defined in YAML format
			if strings.HasSuffix(resolvedPath, ".yaml") && strings.Contains(resolvedPath, "steps-") {
				// Parse the YAML steps file and convert to RST procedure format
				yamlLines, err := parseYAMLStepsFile(resolvedPath, indent)
				if err != nil {
					// If parsing fails, skip this include
					delete(visited, resolvedPath)
					continue
				}
				result = append(result, yamlLines...)
				delete(visited, resolvedPath)
				continue
			}

			// Read the included file
			includeContent, err := os.ReadFile(resolvedPath)
			if err != nil {
				// If we can't read the file, keep the directive as-is
				result = append(result, line)
				continue
			}

			// Split included content into lines
			includeLines := strings.Split(string(includeContent), "\n")

			// Recursively expand includes in the included file
			expandedLines, err := expandIncludesInLines(resolvedPath, includeLines)
			if err != nil {
				// If expansion fails, use the original lines
				expandedLines = includeLines
			}

			// Add the included content with proper indentation
			for _, includeLine := range expandedLines {
				if strings.TrimSpace(includeLine) == "" {
					result = append(result, "")
				} else {
					// Add the include directive's indentation to each line
					result = append(result, strings.Repeat(" ", indent)+includeLine)
				}
			}

			delete(visited, resolvedPath)
		} else {
			result = append(result, line)
		}
	}

	return result, nil
}

// parseYAMLStepsFile parses a YAML steps file and converts it to RST procedure format
func parseYAMLStepsFile(yamlPath string, indent int) ([]string, error) {
	content, err := os.ReadFile(yamlPath)
	if err != nil {
		return nil, err
	}

	// Split by YAML document separator (---)
	docs := strings.Split(string(content), "\n---\n")

	var steps []YAMLStep
	for _, doc := range docs {
		if strings.TrimSpace(doc) == "" {
			continue
		}

		var step YAMLStep
		if err := yaml.Unmarshal([]byte(doc), &step); err != nil {
			// Skip malformed steps
			continue
		}
		steps = append(steps, step)
	}

	// Convert to RST format
	var result []string
	indentStr := strings.Repeat(" ", indent)

	result = append(result, indentStr+".. procedure::")
	result = append(result, indentStr+"   :style: normal")
	result = append(result, "")

	for _, step := range steps {
		result = append(result, indentStr+"   .. step:: "+step.Title)
		result = append(result, "")

		// Add pre-action content if present
		if step.Pre != "" {
			for _, line := range strings.Split(strings.TrimSpace(step.Pre), "\n") {
				result = append(result, indentStr+"      "+line)
			}
			result = append(result, "")
		}

		// Add action content if present
		if step.Action != nil {
			// Action can be a map or a slice of maps
			result = append(result, indentStr+"      (Action content from YAML)")
			result = append(result, "")
		}

		// Add post-action content if present
		if step.Post != "" {
			for _, line := range strings.Split(strings.TrimSpace(step.Post), "\n") {
				result = append(result, indentStr+"      "+line)
			}
			result = append(result, "")
		}
	}

	return result, nil
}

// extractStepsTitle extracts a title from a YAML steps filename
// Example: steps-run-mongodb-on-a-linux-distribution-systemd.yaml -> "Run MongoDB"
func extractStepsTitle(yamlPath string) string {
	basename := filepath.Base(yamlPath)
	// Remove "steps-" prefix and ".yaml" suffix
	basename = strings.TrimPrefix(basename, "steps-")
	basename = strings.TrimSuffix(basename, ".yaml")

	// Convert hyphens to spaces and title case the first word
	parts := strings.Split(basename, "-")
	if len(parts) > 0 {
		parts[0] = strings.Title(parts[0])
		return strings.Join(parts, " ")
	}
	return "Steps"
}

// normalizeIndentation removes the base indentation from all lines
// This is used to normalize content before re-indenting it for output
func normalizeIndentation(content string) string {
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
		indent := getIndentLevel(line)
		if minIndent == -1 || indent < minIndent {
			minIndent = indent
		}
	}

	// If no indentation found, return as-is
	if minIndent <= 0 {
		return content
	}

	// Remove the base indentation from all lines
	var result []string
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			result = append(result, "")
		} else if len(line) >= minIndent {
			result = append(result, line[minIndent:])
		} else {
			result = append(result, line)
		}
	}

	return strings.Join(result, "\n")
}

// containsIncludeDirective checks if any line contains an include directive
func containsIncludeDirective(lines []string) bool {
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, ".. include::") {
			return true
		}
	}
	return false
}

// parseProcedureDirectiveFromLines parses a .. procedure:: directive and its steps
func parseProcedureDirectiveFromLines(lines []string, startIdx int, title string, filePath string) (Procedure, int) {
	procedure := Procedure{
		Type:    ProcedureDirective,
		Title:   title,
		Options: make(map[string]string),
		Steps:   []Step{},
	}

	i := startIdx + 1 // Skip the .. procedure:: line
	baseIndent := -1

	// Parse options and steps
	for i < len(lines) {
		line := lines[i]
		trimmedLine := strings.TrimSpace(line)

		// Empty line
		if trimmedLine == "" {
			i++
			continue
		}

		indent := getIndentLevel(line)

		// Set base indent from first non-empty line
		if baseIndent == -1 && indent > 0 {
			baseIndent = indent
		}

		// Check for option
		if matches := optionRegex.FindStringSubmatch(line); len(matches) > 1 {
			procedure.Options[matches[1]] = strings.TrimSpace(matches[2])
			i++
			continue
		}

		// Check for step directive
		if matches := StepDirectiveRegex.FindStringSubmatch(trimmedLine); len(matches) > 0 {
			step, endLine := parseStepDirectiveFromLines(lines, i, matches[1], filePath)
			procedure.Steps = append(procedure.Steps, step)
			i = endLine + 1
			continue
		}

		// If we've dedented back to base level or beyond, procedure is done
		if baseIndent > 0 && indent < baseIndent && trimmedLine != "" {
			break
		}

		// Check if line is not indented - end of procedure
		if indent == 0 && trimmedLine != "" {
			break
		}

		i++
	}

	// Check for sub-procedures
	for _, step := range procedure.Steps {
		if len(step.SubProcedures) > 0 {
			procedure.HasSubSteps = true
			break
		}
	}

	return procedure, i - 1
}

// parseStepDirectiveFromLines parses a .. step:: directive
func parseStepDirectiveFromLines(lines []string, startIdx int, title string, filePath string) (Step, int) {
	step := Step{
		Title:      strings.TrimSpace(title),
		Options:    make(map[string]string),
		LineNum:    startIdx + 1,
		Variations: []Variation{},
	}

	i := startIdx + 1 // Skip the .. step:: line
	baseIndent := -1
	var contentLines []string
	var selectedContents []SelectedContent

	for i < len(lines) {
		line := lines[i]
		trimmedLine := strings.TrimSpace(line)

		// Empty line
		if trimmedLine == "" {
			contentLines = append(contentLines, "")
			i++
			continue
		}

		indent := getIndentLevel(line)

		// Set base indent from first non-empty line
		if baseIndent == -1 && indent > 0 {
			baseIndent = indent
		}

		// Check for option
		if matches := optionRegex.FindStringSubmatch(line); len(matches) > 1 {
			step.Options[matches[1]] = strings.TrimSpace(matches[2])
			i++
			continue
		}

		// Check for next step directive - we're done
		if StepDirectiveRegex.MatchString(trimmedLine) {
			break
		}

		// Check for tabs directive
		if TabsDirectiveRegex.MatchString(trimmedLine) {
			variation, endLine := parseTabsVariation(lines, i)
			step.Variations = append(step.Variations, variation)
			// Don't add tabs content to contentLines - it's in the variation
			i = endLine + 1
			continue
		}

		// Check for selected-content directive
		if SelectedContentDirectiveRegex.MatchString(trimmedLine) {
			selectedContent, endLine := parseSelectedContent(lines, i)
			selectedContents = append(selectedContents, selectedContent)
			// Don't add selected-content to contentLines - it's tracked separately
			i = endLine + 1
			continue
		}

		// Check for ordered list (sub-steps)
		if isOrderedListStart(trimmedLine) {
			subProcedureSteps, listType, endLine := parseOrderedListSteps(lines, i)
			// Add as a separate sub-procedure with its list type
			step.SubProcedures = append(step.SubProcedures, SubProcedure{
				Steps:    subProcedureSteps,
				ListType: listType,
			})
			// Add the sub-steps to content as well
			for j := i; j <= endLine; j++ {
				contentLines = append(contentLines, lines[j])
			}
			i = endLine + 1
			continue
		}

		// If we've dedented significantly, we're done with this step
		if baseIndent > 0 && indent < baseIndent && trimmedLine != "" {
			break
		}

		// Check if line is not indented - end of step
		if indent == 0 && trimmedLine != "" {
			break
		}

		// Add content line (this is general content, not variation-specific)
		contentLines = append(contentLines, line)
		i++
	}

	// IMPORTANT: If we haven't found any selected-content blocks yet, but the content
	// contains include directives, we need to expand them to check for selected-content
	// blocks in the included files. This handles the case where composable tutorials
	// have procedures with steps that include files containing selected-content blocks.
	if len(selectedContents) == 0 && containsIncludeDirective(contentLines) {
		expandedLines, err := expandIncludesInLines(filePath, contentLines)
		if err == nil {
			// Re-parse the expanded content to find selected-content blocks
			j := 0
			for j < len(expandedLines) {
				trimmedLine := strings.TrimSpace(expandedLines[j])

				if SelectedContentDirectiveRegex.MatchString(trimmedLine) {
					selectedContent, endLine := parseSelectedContent(expandedLines, j)
					selectedContents = append(selectedContents, selectedContent)
					j = endLine + 1
					continue
				}
				j++
			}
		}
	}

	// If we have selected-content blocks, create a variation from them
	if len(selectedContents) > 0 {
		variation := Variation{
			Type:    SelectedContentVariation,
			Options: []string{},
			Content: make(map[string]string),
		}

		for _, sc := range selectedContents {
			selectionKey := strings.Join(sc.Selections, ", ")
			variation.Options = append(variation.Options, selectionKey)
			variation.Content[selectionKey] = sc.Content
		}

		step.Variations = append(step.Variations, variation)
	}

	// Normalize indentation for general content
	rawContent := strings.Join(contentLines, "\n")
	step.Content = normalizeIndentation(rawContent)
	return step, i - 1
}

// parseOrderedListProcedure parses an ordered list as a procedure
func parseOrderedListProcedure(lines []string, startIdx int, title string) (Procedure, int) {
	procedure := Procedure{
		Type:  OrderedList,
		Title: title,
		Steps: []Step{},
	}

	steps, _, endLine := parseOrderedListSteps(lines, startIdx)
	procedure.Steps = steps

	return procedure, endLine
}

// parseOrderedListSteps parses ordered list items as steps and returns the list type
func parseOrderedListSteps(lines []string, startIdx int) ([]Step, string, int) {
	var steps []Step
	i := startIdx
	baseIndent := getIndentLevel(lines[i])

	// Track the list type (numbered or lettered) and the last marker
	var listType string // "numbered" or "lettered"
	var lastMarker string // last number or letter used

	for i < len(lines) {
		line := lines[i]
		trimmedLine := strings.TrimSpace(line)

		// Empty line - might be between list items
		if trimmedLine == "" {
			i++
			continue
		}

		indent := getIndentLevel(line)

		// Check if this is a list item at the same level
		if indent == baseIndent && isOrderedListStart(trimmedLine) {
			// Determine list type from first item if not set
			if listType == "" {
				if numberedListRegex.MatchString(trimmedLine) {
					listType = "numbered"
				} else if letteredListRegex.MatchString(trimmedLine) {
					listType = "lettered"
				}
			}

			step, endLine := parseOrderedListItem(lines, i, listType, lastMarker)
			steps = append(steps, step)

			// Update last marker based on the step we just parsed
			marker := getListMarker(lines[i], listType)
			if marker != "" {
				// Regular marker - use it
				lastMarker = marker
			} else {
				// Continuation marker - compute the next marker
				lastMarker = getNextMarker(lastMarker, listType)
			}

			i = endLine + 1
			continue
		}

		// If we've dedented or hit a non-list line at base level, we're done
		if indent <= baseIndent && !isOrderedListStart(trimmedLine) {
			break
		}

		i++
	}

	return steps, listType, i - 1
}

// parseOrderedListItem parses a single ordered list item
func parseOrderedListItem(lines []string, startIdx int, listType string, lastMarker string) (Step, int) {
	line := lines[startIdx]
	var title string
	var contentLines []string

	// Extract the title from the list marker (don't add the line itself to content)
	if matches := numberedListRegex.FindStringSubmatch(line); len(matches) > 3 {
		title = strings.TrimSpace(matches[3])
	} else if matches := letteredListRegex.FindStringSubmatch(line); len(matches) > 3 {
		title = strings.TrimSpace(matches[3])
	} else if matches := continuationMarkerRegex.FindStringSubmatch(line); len(matches) > 2 {
		// Handle continuation marker (#.) - convert to next number/letter
		nextMarker := getNextMarker(lastMarker, listType)
		title = strings.TrimSpace(matches[2])
		// Prepend the computed marker to the title for display purposes
		if nextMarker != "" {
			title = nextMarker + ". " + title
		}
	}

	baseIndent := getIndentLevel(line)
	i := startIdx + 1

	// The content indent should be greater than the list marker indent
	contentIndent := -1

	// Parse the content of this list item
	for i < len(lines) {
		currentLine := lines[i]
		trimmedLine := strings.TrimSpace(currentLine)

		// Empty line - could be within the list item
		if trimmedLine == "" {
			contentLines = append(contentLines, "")
			i++
			continue
		}

		indent := getIndentLevel(currentLine)

		// Set content indent from first non-empty line after list marker
		if contentIndent == -1 && indent > baseIndent {
			contentIndent = indent
		}

		// Check if this is the next list item at the same level
		if indent == baseIndent && isOrderedListStart(trimmedLine) {
			break
		}

		// Check if we've dedented to or past the base indent (end of list item)
		if indent <= baseIndent && trimmedLine != "" {
			break
		}

		// Check for RST directives at base level (end of list)
		if indent == 0 && (strings.HasPrefix(trimmedLine, "..") || isHeading(lines, i)) {
			break
		}

		// Add content line
		contentLines = append(contentLines, currentLine)
		i++
	}

	step := Step{
		Title:   title,
		Content: strings.Join(contentLines, "\n"),
		LineNum: startIdx + 1,
	}

	return step, i - 1
}

// getListMarker extracts the marker (number or letter) from a list item line
func getListMarker(line string, listType string) string {
	trimmedLine := strings.TrimSpace(line)

	// Check for continuation marker - return empty string as we'll compute it
	if continuationMarkerRegex.MatchString(trimmedLine) {
		return ""
	}

	if listType == "numbered" {
		if matches := numberedListRegex.FindStringSubmatch(trimmedLine); len(matches) > 2 {
			return matches[2]
		}
	} else if listType == "lettered" {
		if matches := letteredListRegex.FindStringSubmatch(trimmedLine); len(matches) > 2 {
			return matches[2]
		}
	}

	return ""
}

// getNextMarker computes the next marker in a sequence
func getNextMarker(lastMarker string, listType string) string {
	if lastMarker == "" {
		// If no last marker, start from 1 or 'a'
		if listType == "numbered" {
			return "1"
		} else if listType == "lettered" {
			return "a"
		}
		return ""
	}

	if listType == "numbered" {
		// Parse the number and increment
		if num, err := strconv.Atoi(lastMarker); err == nil {
			return strconv.Itoa(num + 1)
		}
	} else if listType == "lettered" {
		// Increment the letter
		if len(lastMarker) == 1 {
			char := lastMarker[0]
			if char >= 'a' && char < 'z' {
				return string(char + 1)
			} else if char >= 'A' && char < 'Z' {
				return string(char + 1)
			} else if char == 'z' {
				return "aa" // Handle overflow (rare case)
			} else if char == 'Z' {
				return "AA"
			}
		}
	}

	return lastMarker
}


// parseTabsVariation parses a .. tabs:: directive and its tab content
func parseTabsVariation(lines []string, startIdx int) (Variation, int) {
	variation := Variation{
		Type:    TabVariation,
		Options: []string{},
		Content: make(map[string]string),
	}

	i := startIdx + 1 // Skip the .. tabs:: line
	baseIndent := -1

	// Parse tabs options first
	for i < len(lines) {
		line := lines[i]
		trimmedLine := strings.TrimSpace(line)

		if trimmedLine == "" {
			i++
			continue
		}

		indent := getIndentLevel(line)
		if baseIndent == -1 && indent > 0 {
			baseIndent = indent
		}

		// Check for tabs options
		if matches := optionRegex.FindStringSubmatch(line); len(matches) > 1 {
			i++
			continue
		}

		// Check for tab directive
		if TabDirectiveRegex.MatchString(trimmedLine) {
			tabid, content, endLine := parseTabContent(lines, i)
			if tabid != "" {
				variation.Options = append(variation.Options, tabid)
				variation.Content[tabid] = content
			}
			i = endLine + 1
			continue
		}

		// If we've dedented, we're done
		if baseIndent > 0 && indent < baseIndent && trimmedLine != "" {
			break
		}

		if indent == 0 && trimmedLine != "" {
			break
		}

		i++
	}

	return variation, i - 1
}

// parseTabContent parses a single .. tab:: directive
func parseTabContent(lines []string, startIdx int) (string, string, int) {
	var tabid string
	var contentLines []string

	// Extract tabid from options
	i := startIdx + 1
	baseIndent := -1

	for i < len(lines) {
		currentLine := lines[i]
		trimmedCurrentLine := strings.TrimSpace(currentLine)

		if trimmedCurrentLine == "" {
			contentLines = append(contentLines, "")
			i++
			continue
		}

		indent := getIndentLevel(currentLine)
		if baseIndent == -1 && indent > 0 {
			baseIndent = indent
		}

		// Check for :tabid: option
		if matches := optionRegex.FindStringSubmatch(currentLine); len(matches) > 1 {
			if matches[1] == "tabid" {
				tabid = strings.TrimSpace(matches[2])
			}
			i++
			continue
		}

		// Check for next tab directive
		if TabDirectiveRegex.MatchString(trimmedCurrentLine) {
			break
		}

		// If we've dedented significantly, we're done
		if baseIndent > 0 && indent < baseIndent && trimmedCurrentLine != "" {
			break
		}

		if indent == 0 && trimmedCurrentLine != "" {
			break
		}

		// Add content line
		contentLines = append(contentLines, currentLine)
		i++
	}

	// Normalize indentation before storing
	rawContent := strings.Join(contentLines, "\n")
	content := normalizeIndentation(rawContent)
	return tabid, content, i - 1
}

// parseTabSetWithProcedures parses a top-level .. tabs:: directive that contains procedures.
// This is different from parseTabsVariation which handles tabs within steps.
func parseTabSetWithProcedures(lines []string, startIdx int, title string, filePath string) (*TabSet, int) {
	tabSet := &TabSet{
		Title:      title,
		Tabs:       make(map[string][]string),
		TabIDs:     []string{},
		Procedures: make(map[string]Procedure),
		LineNum:    startIdx + 1,
		FilePath:   filePath,
	}

	i := startIdx + 1 // Skip the .. tabs:: line
	baseIndent := -1

	// Parse each tab
	for i < len(lines) {
		line := lines[i]
		trimmedLine := strings.TrimSpace(line)

		if trimmedLine == "" {
			i++
			continue
		}

		indent := getIndentLevel(line)
		if baseIndent == -1 && indent > 0 {
			baseIndent = indent
		}

		// Check for tabs options (skip them)
		if matches := optionRegex.FindStringSubmatch(line); len(matches) > 1 {
			i++
			continue
		}

		// Check for tab directive
		if TabDirectiveRegex.MatchString(trimmedLine) {
			tabid, contentLines, endLine := parseTabContentLines(lines, i)
			if tabid != "" {
				tabSet.TabIDs = append(tabSet.TabIDs, tabid)
				tabSet.Tabs[tabid] = contentLines
			}
			i = endLine + 1
			continue
		}

		// If we've dedented, we're done
		if baseIndent > 0 && indent < baseIndent && trimmedLine != "" {
			break
		}

		if indent == 0 && trimmedLine != "" {
			break
		}

		i++
	}

	// Now parse procedures from each tab's content
	for _, tabid := range tabSet.TabIDs {
		contentLines := tabSet.Tabs[tabid]
		// Parse procedures from this tab's content
		procedures, err := parseProceduresFromLines(contentLines, filePath)
		if err == nil && len(procedures) > 0 {
			// Take the first procedure found in this tab
			// (typically there should only be one procedure per tab)
			procedure := procedures[0]
			procedure.Title = title // Use the heading as the title
			tabSet.Procedures[tabid] = procedure
		}
	}

	return tabSet, i - 1
}

// parseTabContentLines parses a single .. tab:: directive and returns the content as lines.
// This is similar to parseTabContent but returns lines instead of normalized content.
func parseTabContentLines(lines []string, startIdx int) (string, []string, int) {
	var tabid string
	var contentLines []string

	// Extract tabid from options
	i := startIdx + 1
	baseIndent := -1

	for i < len(lines) {
		currentLine := lines[i]
		trimmedCurrentLine := strings.TrimSpace(currentLine)

		if trimmedCurrentLine == "" {
			contentLines = append(contentLines, "")
			i++
			continue
		}

		indent := getIndentLevel(currentLine)
		if baseIndent == -1 && indent > 0 {
			baseIndent = indent
		}

		// Check for :tabid: option
		if matches := optionRegex.FindStringSubmatch(currentLine); len(matches) > 1 {
			if matches[1] == "tabid" {
				tabid = strings.TrimSpace(matches[2])
			}
			i++
			continue
		}

		// Check for next tab directive
		if TabDirectiveRegex.MatchString(trimmedCurrentLine) {
			break
		}

		// If we've dedented significantly, we're done
		if baseIndent > 0 && indent < baseIndent && trimmedCurrentLine != "" {
			break
		}

		if indent == 0 && trimmedCurrentLine != "" {
			break
		}

		// Add content line (preserve original indentation)
		contentLines = append(contentLines, currentLine)
		i++
	}

	return tabid, contentLines, i - 1
}

// extractProceduresFromTabSet extracts procedures from a tab set.
// Each tab's procedure is returned as a separate procedure, but they all share
// the same TabSet reference so they can be grouped for analysis/reporting.
func extractProceduresFromTabSet(tabSet *TabSet) []Procedure {
	if len(tabSet.Procedures) == 0 {
		return []Procedure{}
	}

	// Sort tab IDs for deterministic order
	sortedTabIDs := make([]string, len(tabSet.TabIDs))
	copy(sortedTabIDs, tabSet.TabIDs)
	sort.Strings(sortedTabIDs)

	// Create a shared TabSetInfo that all procedures will reference
	sharedTabSetInfo := &TabSetInfo{
		TabIDs:     sortedTabIDs,
		Procedures: tabSet.Procedures,
	}

	// Extract each tab's procedure as a separate procedure
	var procedures []Procedure
	for _, tabid := range sortedTabIDs {
		if proc, ok := tabSet.Procedures[tabid]; ok {
			// Attach the shared TabSet reference for grouping
			proc.TabSet = sharedTabSetInfo
			// Set the specific tab ID for this procedure
			proc.TabID = tabid
			procedures = append(procedures, proc)
		}
	}

	return procedures
}
