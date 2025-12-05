// Package rst provides parsing and analysis of reStructuredText (RST) procedures
// from MongoDB documentation.
package rst

import (
	"fmt"
	"sort"
	"strings"
)

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
