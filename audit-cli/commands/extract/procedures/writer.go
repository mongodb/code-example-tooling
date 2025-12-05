package procedures

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mongodb/code-example-tooling/audit-cli/internal/rst"
)

// WriteVariation writes a procedure variation to a file.
//
// This function formats the procedure for the specific variation and writes
// it to the output file in RST format.
//
// Parameters:
//   - variation: The procedure variation to write
//   - outputDir: Directory where the file should be written
//   - dryRun: If true, don't actually write the file
//
// Returns:
//   - error: Any error encountered during writing
func WriteVariation(variation ProcedureVariation, outputDir string, dryRun bool) error {
	// Format the procedure for this variation
	content, err := rst.FormatProcedureForVariation(variation.Procedure, variation.VariationName)
	if err != nil {
		return fmt.Errorf("failed to format procedure variation: %w", err)
	}

	// Generate output path
	outputPath := filepath.Join(outputDir, variation.OutputFile)

	if dryRun {
		fmt.Printf("Would write: %s\n", outputPath)
		return nil
	}

	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Write the file
	if err := os.WriteFile(outputPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", outputPath, err)
	}

	return nil
}

// WriteAllVariations writes all procedure variations to files.
//
// Parameters:
//   - variations: Slice of procedure variations to write
//   - outputDir: Directory where files should be written
//   - dryRun: If true, don't actually write files
//   - verbose: If true, print detailed information
//
// Returns:
//   - int: Number of files written (or would be written in dry run mode)
//   - error: Any error encountered during writing
func WriteAllVariations(variations []ProcedureVariation, outputDir string, dryRun bool, verbose bool) (int, error) {
	filesWritten := 0

	for _, variation := range variations {
		if err := WriteVariation(variation, outputDir, dryRun); err != nil {
			return filesWritten, err
		}

		filesWritten++
	}

	return filesWritten, nil
}

