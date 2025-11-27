package procedures

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/mongodb/code-example-tooling/audit-cli/internal/rst"
)

// ParseFile parses a file and extracts all procedure variations.
//
// This function parses all procedures from the file and generates variations
// based on tabs, composable tutorials, or returns a single variation if no
// variations are present.
//
// Parameters:
//   - filePath: Path to the RST file to parse
//   - selectionFilter: Optional filter to extract only a specific variation
//   - expandIncludes: If true, expands .. include:: directives inline
//
// Returns:
//   - []ProcedureVariation: Slice of procedure variations to extract
//   - error: Any error encountered during parsing
func ParseFile(filePath string, selectionFilter string, expandIncludes bool) ([]ProcedureVariation, error) {
	// Parse all procedures from the file
	procedures, err := rst.ParseProceduresWithOptions(filePath, expandIncludes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse procedures from %s: %w", filePath, err)
	}

	var variations []ProcedureVariation

	// Create one variation per unique procedure
	// Each procedure represents a unique piece of content (grouped by heading + content hash)
	for _, procedure := range procedures {
		// Get all selections this procedure appears in
		variationNames := rst.GetProcedureVariations(procedure)

		// If a selection filter is specified, check if this procedure matches
		if selectionFilter != "" {
			matches := false
			for _, varName := range variationNames {
				if varName == selectionFilter {
					matches = true
					break
				}
			}
			if !matches {
				continue
			}
		}

		// Create a single variation representing this unique procedure
		// The VariationName will contain all selections this procedure appears in
		variationName := ""
		if len(variationNames) > 0 {
			variationName = strings.Join(variationNames, "; ")
		}

		variation := ProcedureVariation{
			Procedure:     procedure,
			VariationName: variationName,
			SourceFile:    filePath,
			OutputFile:    generateOutputFilename(filePath, procedure, ""),
		}

		variations = append(variations, variation)
	}

	return variations, nil
}

// generateVariations generates all variations for a procedure.
func generateVariations(procedure rst.Procedure, sourceFile string, selectionFilter string) []ProcedureVariation {
	var variations []ProcedureVariation

	// Get all variation identifiers for this procedure
	variationNames := rst.GetProcedureVariations(procedure)

	// If no variations, create a single variation with empty name
	if len(variationNames) == 0 {
		variationNames = []string{""}
	}

	// Generate a variation for each identifier
	for _, variationName := range variationNames {
		// If a selection filter is specified, only include matching variations
		if selectionFilter != "" && variationName != selectionFilter {
			continue
		}

		variation := ProcedureVariation{
			Procedure:     procedure,
			VariationName: variationName,
			SourceFile:    sourceFile,
			OutputFile:    generateOutputFilename(sourceFile, procedure, variationName),
		}

		variations = append(variations, variation)
	}

	return variations
}

// generateOutputFilename generates the output filename for a procedure.
//
// Format: {heading}-{first-step-title}-{hash}.rst
// Example: "before-you-begin-pull-the-mongodb-docker-image-a1b2c3.rst"
//
// The hash is a short (6 character) hash of the procedure content to ensure uniqueness.
func generateOutputFilename(sourceFile string, procedure rst.Procedure, variationName string) string {
	// Get the base name from the source file
	baseName := filepath.Base(sourceFile)
	baseName = strings.TrimSuffix(baseName, filepath.Ext(baseName))

	// Sanitize the procedure title (heading) for use in filename
	title := sanitizeFilename(procedure.Title)
	if title == "" {
		title = baseName
	}

	// Generate a short hash of the procedure content for uniqueness
	contentHash := computeContentHash(procedure)
	shortHash := contentHash[:6]

	// If the procedure has steps, use the first step title to make the filename descriptive
	if len(procedure.Steps) > 0 && procedure.Steps[0].Title != "" {
		firstStepTitle := sanitizeFilename(procedure.Steps[0].Title)
		return fmt.Sprintf("%s-%s-%s.rst", title, firstStepTitle, shortHash)
	}

	return fmt.Sprintf("%s-%s.rst", title, shortHash)
}

// computeContentHash generates a hash of the procedure's content for uniqueness.
func computeContentHash(proc rst.Procedure) string {
	var content strings.Builder

	// Include title
	content.WriteString(proc.Title)
	content.WriteString("|")

	// Include all step titles and content
	for _, step := range proc.Steps {
		content.WriteString(step.Title)
		content.WriteString("|")
		content.WriteString(step.Content)
		content.WriteString("|")
	}

	// Compute SHA256 hash
	hash := sha256.Sum256([]byte(content.String()))
	return hex.EncodeToString(hash[:])
}

// sanitizeFilename sanitizes a string for use in a filename.
func sanitizeFilename(s string) string {
	// Convert to lowercase
	s = strings.ToLower(s)

	// Replace spaces and special characters with hyphens
	s = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			return r
		}
		if r == ' ' || r == '_' || r == ',' {
			return '-'
		}
		return -1 // Remove character
	}, s)

	// Remove multiple consecutive hyphens
	for strings.Contains(s, "--") {
		s = strings.ReplaceAll(s, "--", "-")
	}

	// Trim hyphens from start and end
	s = strings.Trim(s, "-")

	return s
}

