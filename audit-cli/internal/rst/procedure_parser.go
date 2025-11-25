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
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// ProcedureType represents the type of procedure implementation.
type ProcedureType string

const (
	// ProcedureDirective represents procedures using .. procedure:: directive
	ProcedureDirective ProcedureType = "procedure-directive"
	// OrderedList represents procedures using ordered lists
	OrderedList ProcedureType = "ordered-list"
)

// Procedure represents a parsed procedure from an RST file.
type Procedure struct {
	Type               ProcedureType     // Type of procedure (directive or ordered list)
	Title              string            // Title/heading above the procedure
	Options            map[string]string // Directive options (for procedure directive)
	Steps              []Step            // Steps in the procedure
	LineNum            int               // Line number where procedure starts (1-based)
	EndLineNum         int               // Line number where procedure ends (1-based)
	HasSubSteps        bool              // Whether this procedure contains sub-procedures
	IsSubProcedure     bool              // Whether this is a sub-procedure within a step
	ComposableTutorial *ComposableTutorial // Composable tutorial wrapping this procedure (if any)
	TabSet             *TabSetInfo       // Tab set wrapping this procedure (if any)
	TabID              string            // The specific tab ID this procedure belongs to (if part of a tab set)
}

// TabSetInfo represents information about a tab set containing procedure variations.
// This is used for grouping procedures for analysis/reporting purposes.
type TabSetInfo struct {
	TabIDs     []string            // All tab IDs in the set (for grouping)
	Procedures map[string]Procedure // All procedures by tabid (for grouping)
}

// Step represents a single step in a procedure.
type Step struct {
	Title      string            // Step title (for .. step:: directive)
	Content    string            // Step content (raw RST)
	Options    map[string]string // Step options
	LineNum    int               // Line number where step starts
	Variations []Variation       // Variations within this step (tabs or selected content)
	SubSteps   []Step            // Sub-steps (ordered lists within this step)
}

// Variation represents a content variation within a step.
type Variation struct {
	Type    VariationType     // Type of variation (tab or selected-content)
	Options []string          // Available options (tabids or selections)
	Content map[string]string // Content for each option
}

// VariationType represents the type of content variation.
type VariationType string

const (
	// TabVariation represents variations using .. tabs:: directive
	TabVariation VariationType = "tabs"
	// SelectedContentVariation represents variations using .. selected-content:: directive
	SelectedContentVariation VariationType = "selected-content"
)

// ComposableTutorial represents a composable tutorial structure.
type ComposableTutorial struct {
	Title                string            // Title/heading above the composable tutorial
	Options              []string          // Available option names (e.g., ["interface", "language"])
	Defaults             []string          // Default selections (e.g., ["driver", "nodejs"])
	Selections           []string          // All unique selection combinations found
	GeneralContent       []string          // Content lines that apply to all selections
	LineNum              int               // Line number where tutorial starts
	FilePath             string            // Path to the source file (for resolving includes)
	Procedure            *Procedure        // The procedure within the composable tutorial
	SelectedContentBlocks []SelectedContent // All selected-content blocks (for extracting multiple procedures)
}

// TabSet represents a tabs directive containing procedures.
type TabSet struct {
	Title      string              // Title/heading above the tabs
	Tabs       map[string][]string // Tab content by tabid (lines of RST)
	TabIDs     []string            // Ordered list of tab IDs
	Procedures map[string]Procedure // Parsed procedures by tabid
	LineNum    int                 // Line number where tabs start
	FilePath   string              // Path to the source file (for resolving includes)
}

// SelectedContent represents a selected-content block within a composable tutorial.
type SelectedContent struct {
	Selections []string // The selections for this content (e.g., ["driver", "nodejs"])
	Content    string   // The content for this selection
	LineNum    int      // Line number where this selected-content starts
}

// Regular expressions for parsing ordered lists
var (
	// Matches numbered lists: 1. or 1)
	numberedListRegex = regexp.MustCompile(`^(\s*)(\d+)[\.\)]\s+(.*)$`)
	// Matches lettered lists: a. or a) or A. or A)
	letteredListRegex = regexp.MustCompile(`^(\s*)([a-zA-Z])[\.\)]\s+(.*)$`)
)

// ParseProcedures parses all procedures from an RST file.
//
// This function scans the file and extracts all procedures, whether they are
// implemented using .. procedure:: directives, ordered lists, or composable tutorials.
//
// Parameters:
//   - filePath: Path to the RST file to parse
//
// Returns:
//   - []Procedure: Slice of all parsed procedures
//   - error: Any error encountered during parsing
func ParseProcedures(filePath string) ([]Procedure, error) {
	return ParseProceduresWithOptions(filePath, false)
}

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
				if trimmedLine != "" && headingLower != "procedure" && headingLower != "overview" && headingLower != "steps" {
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

		// Include substeps
		for _, substep := range step.SubSteps {
			content.WriteString(substep.Title)
			content.WriteString("|")
			content.WriteString(substep.Content)
			content.WriteString("|")
		}
	}

	// Compute SHA256 hash
	hash := sha256.Sum256([]byte(content.String()))
	return hex.EncodeToString(hash[:])
}

// isOrderedListStart checks if a line starts an ordered list
func isOrderedListStart(line string) bool {
	return numberedListRegex.MatchString(line) || letteredListRegex.MatchString(line)
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

// YAMLStep represents a step in a YAML steps file
type YAMLStep struct {
	Title   string `yaml:"title"`
	StepNum int    `yaml:"stepnum"`
	Level   int    `yaml:"level"`
	Ref     string `yaml:"ref"`
	Pre     string `yaml:"pre"`
	Action  interface{} `yaml:"action"`
	Post    string `yaml:"post"`
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

	// Check for sub-steps
	for _, step := range procedure.Steps {
		if len(step.SubSteps) > 0 {
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
		SubSteps:   []Step{},
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
			subSteps, endLine := parseOrderedListSteps(lines, i)
			step.SubSteps = append(step.SubSteps, subSteps...)
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

	steps, endLine := parseOrderedListSteps(lines, startIdx)
	procedure.Steps = steps

	return procedure, endLine
}

// parseOrderedListSteps parses ordered list items as steps
func parseOrderedListSteps(lines []string, startIdx int) ([]Step, int) {
	var steps []Step
	i := startIdx
	baseIndent := getIndentLevel(lines[i])

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
			step, endLine := parseOrderedListItem(lines, i)
			steps = append(steps, step)
			i = endLine + 1
			continue
		}

		// If we've dedented or hit a non-list line at base level, we're done
		if indent <= baseIndent && !isOrderedListStart(trimmedLine) {
			break
		}

		i++
	}

	return steps, i - 1
}

// parseOrderedListItem parses a single ordered list item
func parseOrderedListItem(lines []string, startIdx int) (Step, int) {
	line := lines[startIdx]
	var title string
	var contentLines []string

	// Extract the title from the list marker (don't add the line itself to content)
	if matches := numberedListRegex.FindStringSubmatch(line); len(matches) > 3 {
		title = strings.TrimSpace(matches[3])
	} else if matches := letteredListRegex.FindStringSubmatch(line); len(matches) > 3 {
		title = strings.TrimSpace(matches[3])
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

// GetProcedureVariations returns all variations of a procedure based on tabs and selected content.
//
// Parameters:
//   - procedure: The procedure to analyze
//
// Returns:
//   - []string: List of variation identifiers (e.g., "python", "nodejs", "drivers-tab")
func GetProcedureVariations(procedure Procedure) []string {
	// If this procedure has a composable tutorial, return those selections
	if procedure.ComposableTutorial != nil {
		return procedure.ComposableTutorial.Selections
	}

	// If this procedure has a specific tab ID, return only that tab ID
	// This is for individual procedures extracted from a tab set
	if procedure.TabID != "" {
		return []string{procedure.TabID}
	}

	// If this procedure is part of a tab set (for grouping/analysis),
	// return all tab IDs in the set
	if procedure.TabSet != nil {
		return procedure.TabSet.TabIDs
	}

	// Otherwise, collect variations from tabs within steps
	variations := []string{}
	variationSet := make(map[string]bool)

	for _, step := range procedure.Steps {
		for _, variation := range step.Variations {
			if variation.Type == TabVariation {
				for _, option := range variation.Options {
					variationSet[option] = true
				}
			}
		}
	}

	// Convert set to slice and sort for deterministic order
	for variation := range variationSet {
		variations = append(variations, variation)
	}
	sort.Strings(variations)

	// If no variations found, return a single empty variation
	if len(variations) == 0 {
		return []string{""}
	}

	return variations
}

// parseComposableTutorial parses a .. composable-tutorial:: directive
func parseComposableTutorial(lines []string, startIdx int, title string, filePath string) (*ComposableTutorial, int) {
	tutorial := &ComposableTutorial{
		Title:          title,
		Options:        []string{},
		Defaults:       []string{},
		Selections:     []string{},
		GeneralContent: []string{},
		LineNum:        startIdx + 1,
		FilePath:       filePath,
	}

	i := startIdx + 1 // Skip the .. composable-tutorial:: line
	baseIndent := -1

	// Track selected-content blocks to check for procedures inside them
	var selectedContentBlocks []SelectedContent

	// Parse options and procedure
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

		// Check for options
		if matches := optionRegex.FindStringSubmatch(line); len(matches) > 1 {
			if matches[1] == "options" {
				tutorial.Options = strings.Split(strings.TrimSpace(matches[2]), ",")
				for j := range tutorial.Options {
					tutorial.Options[j] = strings.TrimSpace(tutorial.Options[j])
				}
			} else if matches[1] == "defaults" {
				tutorial.Defaults = strings.Split(strings.TrimSpace(matches[2]), ",")
				for j := range tutorial.Defaults {
					tutorial.Defaults[j] = strings.TrimSpace(tutorial.Defaults[j])
				}
			}
			i++
			continue
		}

		// Check for selected-content directive
		if SelectedContentDirectiveRegex.MatchString(trimmedLine) {
			selectedContent, endLine := parseSelectedContent(lines, i)
			selectedContentBlocks = append(selectedContentBlocks, selectedContent)
			i = endLine + 1
			continue
		}

		// Check for procedure directive within composable tutorial
		// NOTE: We only capture tutorial-level procedures if we haven't found any selected-content blocks yet.
		// If we have selected-content blocks, all procedures should be extracted from those blocks,
		// not from the tutorial level (which may contain expanded includes from multiple selections).
		if ProcedureDirectiveRegex.MatchString(trimmedLine) && len(selectedContentBlocks) == 0 {
			procedure, endLine := parseProcedureDirectiveFromLines(lines, i, title, tutorial.FilePath)
			tutorial.Procedure = &procedure

			// Extract all unique selections from the procedure's steps
			selectionsMap := make(map[string]bool)
			for _, step := range procedure.Steps {
				for _, variation := range step.Variations {
					if variation.Type == SelectedContentVariation {
						for _, option := range variation.Options {
							selectionsMap[option] = true
						}
					}
				}
			}

			// Convert to slice and sort for deterministic order
			var selections []string
			for selection := range selectionsMap {
				selections = append(selections, selection)
			}
			sort.Strings(selections)
			tutorial.Selections = selections

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

		// This is general content
		tutorial.GeneralContent = append(tutorial.GeneralContent, line)
		i++
	}

	// Store the selected-content blocks in the tutorial
	tutorial.SelectedContentBlocks = selectedContentBlocks

	// NOTE: We do NOT set tutorial.Procedure here even if we find procedures in selected-content blocks.
	// The extractProceduresFromComposableTutorial function will extract ALL procedures from all
	// selected-content blocks and return them as separate Procedure objects.

	return tutorial, i - 1
}

// extractProceduresFromComposableTutorial extracts ALL procedures from a composable tutorial.
// This function finds all procedures across all selected-content blocks and returns them
// as separate Procedure objects, each with their own list of variations (selections).
//
// Key insight: A composable tutorial can contain MULTIPLE procedures, and each selected-content
// block can have DIFFERENT procedures. We need to track which procedures appear in which selections.
func extractProceduresFromComposableTutorial(tutorial *ComposableTutorial, startLine int) []Procedure {
	var procedures []Procedure

	// If there's a procedure at the tutorial level, use it
	if tutorial.Procedure != nil {
		tutorial.Procedure.LineNum = startLine + 1
		tutorial.Procedure.ComposableTutorial = tutorial
		procedures = append(procedures, *tutorial.Procedure)
		return procedures
	}

	// Map to track procedures by a unique identifier
	// Key: unique procedure identifier (based on first step title), Value: procedure info
	type ProcedureInfo struct {
		Procedure  *Procedure
		Selections []string
	}
	proceduresMap := make(map[string]*ProcedureInfo)

	// Extract procedures from each selected-content block
	for _, sc := range tutorial.SelectedContentBlocks {
		selectionKey := strings.Join(sc.Selections, ", ")
		contentLines := strings.Split(sc.Content, "\n")

		// Expand includes in the selected-content block
		// This is necessary because the selected-content block may contain include directives
		// that reference files with procedure directives
		expandedLines, err := expandIncludesInLines(tutorial.FilePath, contentLines)
		if err != nil {
			// Fall back to unexpanded lines if expansion fails
			expandedLines = contentLines
		}

		contentLines = expandedLines

		// Find ALL procedures in this selected-content block
		// Track the most recent heading to use as the procedure title
		currentHeading := tutorial.Title // Start with the composable tutorial's title
		j := 0
		for j < len(contentLines) {
			trimmedLine := strings.TrimSpace(contentLines[j])

			// Check for headings (look ahead for underline)
			if j+1 < len(contentLines) {
				nextLine := strings.TrimSpace(contentLines[j+1])
				if isHeadingUnderline(nextLine) && len(nextLine) >= len(trimmedLine) {
					// Skip empty headings and generic headings that don't provide meaningful context
					headingLower := strings.ToLower(trimmedLine)
					if trimmedLine != "" && headingLower != "procedure" && headingLower != "overview" && headingLower != "steps" {
						currentHeading = trimmedLine
					}
					j += 2 // Skip heading and underline
					continue
				}
			}

			if ProcedureDirectiveRegex.MatchString(trimmedLine) {
				// Parse the procedure and set its title to the most recent heading
				procedure, endLine := parseProcedureDirectiveFromLines(contentLines, j, currentHeading, tutorial.FilePath)
				procedure.LineNum = startLine + 1

				// Create a unique identifier based on the procedure's actual content
				// This allows us to detect when the same procedure appears in multiple selections
				contentHash := computeProcedureContentHash(&procedure)

				// Use heading + content hash as the key
				// This groups procedures with identical content but keeps them separate if content differs
				procedureID := currentHeading + "::" + contentHash

				// Track this procedure and which selection it appears in
				if proceduresMap[procedureID] == nil {
					proceduresMap[procedureID] = &ProcedureInfo{
						Procedure:  &procedure,
						Selections: []string{},
					}
				}
				proceduresMap[procedureID].Selections = append(proceduresMap[procedureID].Selections, selectionKey)

				j = endLine + 1
			} else {
				j++
			}
		}
	}

	// Convert the map to a list of procedures with their selections
	// Sort the keys to ensure deterministic order
	var procedureIDs []string
	for id := range proceduresMap {
		procedureIDs = append(procedureIDs, id)
	}
	sort.Strings(procedureIDs)

	for _, id := range procedureIDs {
		info := proceduresMap[id]
		// Create a new composable tutorial for this procedure with its specific selections
		procTutorial := &ComposableTutorial{
			Options:        tutorial.Options,
			Defaults:       tutorial.Defaults,
			Selections:     info.Selections,
			GeneralContent: tutorial.GeneralContent,
			LineNum:        tutorial.LineNum,
		}

		info.Procedure.ComposableTutorial = procTutorial
		procedures = append(procedures, *info.Procedure)
	}

	return procedures
}

// parseSelectedContent parses a .. selected-content:: directive
func parseSelectedContent(lines []string, startIdx int) (SelectedContent, int) {
	selectedContent := SelectedContent{
		Selections: []string{},
		LineNum:    startIdx + 1,
	}

	i := startIdx + 1 // Skip the .. selected-content:: line
	baseIndent := -1
	var contentLines []string

	for i < len(lines) {
		line := lines[i]
		trimmedLine := strings.TrimSpace(line)

		if trimmedLine == "" {
			contentLines = append(contentLines, "")
			i++
			continue
		}

		indent := getIndentLevel(line)
		if baseIndent == -1 && indent > 0 {
			baseIndent = indent
		}

		// Check for :selections: option
		if matches := optionRegex.FindStringSubmatch(line); len(matches) > 1 {
			if matches[1] == "selections" {
				selectedContent.Selections = strings.Split(strings.TrimSpace(matches[2]), ",")
				for j := range selectedContent.Selections {
					selectedContent.Selections[j] = strings.TrimSpace(selectedContent.Selections[j])
				}
			}
			i++
			continue
		}

		// Check for next selected-content or step directive
		if SelectedContentDirectiveRegex.MatchString(trimmedLine) || StepDirectiveRegex.MatchString(trimmedLine) {
			break
		}

		// If we've dedented, we're done
		if baseIndent > 0 && indent < baseIndent && trimmedLine != "" {
			break
		}

		if indent == 0 && trimmedLine != "" {
			break
		}

		// Add content line
		contentLines = append(contentLines, line)
		i++
	}

	// Normalize indentation before storing
	rawContent := strings.Join(contentLines, "\n")
	selectedContent.Content = normalizeIndentation(rawContent)

	return selectedContent, i - 1
}

// FormatProcedureForVariation formats a procedure for a specific variation.
//
// This function interpolates the general content with the selection-specific content
// to produce the complete procedure as it would be rendered for that variation.
//
// Parameters:
//   - procedure: The procedure to format
//   - variation: The variation identifier (e.g., "python", "nodejs", "driver, nodejs")
//
// Returns:
//   - string: The formatted procedure content in RST format
//   - error: Any error encountered during formatting
func FormatProcedureForVariation(procedure Procedure, variation string) (string, error) {
	var output strings.Builder

	// Write procedure header if it's a directive
	if procedure.Type == ProcedureDirective {
		output.WriteString(".. procedure::\n")
		for key, value := range procedure.Options {
			output.WriteString(fmt.Sprintf("   :%s: %s\n", key, value))
		}
		output.WriteString("\n")
	}

	// Write each step
	for i, step := range procedure.Steps {
		if procedure.Type == ProcedureDirective {
			output.WriteString(fmt.Sprintf("   .. step:: %s\n\n", step.Title))
		} else if procedure.Type == OrderedList {
			// For ordered lists, preserve the list marker
			output.WriteString(fmt.Sprintf("%d. %s\n\n", i+1, step.Title))
		}

		// Write step content, filtering for the specific variation
		content := filterContentForVariation(step, variation)

		// Indent content for procedure directive
		if procedure.Type == ProcedureDirective {
			lines := strings.Split(content, "\n")
			for _, line := range lines {
				if line != "" {
					output.WriteString("      " + line + "\n")
				} else {
					output.WriteString("\n")
				}
			}
		} else {
			output.WriteString(content)
		}

		output.WriteString("\n")
	}

	return output.String(), nil
}

// filterContentForVariation filters step content to only include the specified variation
func filterContentForVariation(step Step, variation string) string {
	var result strings.Builder

	// Start with general content (content that's not in variations)
	if step.Content != "" {
		result.WriteString(step.Content)
	}

	// If no variations, return all content
	if len(step.Variations) == 0 {
		return result.String()
	}

	// Add variation-specific content
	for _, v := range step.Variations {
		if content, ok := v.Content[variation]; ok {
			if result.Len() > 0 {
				result.WriteString("\n\n")
			}
			result.WriteString(content)
		}
	}

	return result.String()
}

