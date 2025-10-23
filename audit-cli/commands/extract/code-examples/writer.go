package code_examples

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// WriteCodeExample writes a code example to a file in the output directory.
//
// Generates a standardized filename and writes the code content to that file.
// If dryRun is true, returns the filename without actually writing the file.
//
// Parameters:
//   - example: The code example to write
//   - outputDir: Directory where the file should be written
//   - dryRun: If true, skip writing and only return the filename
//
// Returns:
//   - string: The full path to the output file
//   - error: Any error encountered during writing
func WriteCodeExample(example CodeExample, outputDir string, dryRun bool) (string, error) {
	filename := GenerateOutputFilename(example)
	outputPath := filepath.Join(outputDir, filename)

	if dryRun {
		return outputPath, nil
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	if err := os.WriteFile(outputPath, []byte(example.Content), 0644); err != nil {
		return "", fmt.Errorf("failed to write file %s: %w", outputPath, err)
	}

	return outputPath, nil
}

// GenerateOutputFilename generates a standardized filename for a code example.
//
// The filename format is: {source-base}.{directive-type}.{index}.{ext}
// For io-code-block directives: {source-base}.{directive-type}.{index}.{subtype}.{ext}
//
// Examples:
//   - my-doc.code-block.1.js
//   - my-doc.literalinclude.2.py
//   - my-doc.io-code-block.1.input.js
//   - my-doc.io-code-block.1.output.json
//
// Parameters:
//   - example: The code example to generate a filename for
//
// Returns:
//   - string: The generated filename (without directory path)
func GenerateOutputFilename(example CodeExample) string {
	sourceBase := filepath.Base(example.SourceFile)
	sourceBase = strings.TrimSuffix(sourceBase, filepath.Ext(sourceBase))

	extension := GetFileExtensionFromLanguage(example.Language)

	// For io-code-block, include the subtype (input/output) in the filename
	if example.DirectiveName == IoCodeBlock && example.SubType != "" {
		filename := fmt.Sprintf("%s.%s.%d.%s%s",
			sourceBase,
			example.DirectiveName,
			example.Index,
			example.SubType,
			extension,
		)
		return filename
	}

	filename := fmt.Sprintf("%s.%s.%d%s",
		sourceBase,
		example.DirectiveName,
		example.Index,
		extension,
	)

	return filename
}

// EnsureOutputDirectory ensures the output directory exists.
//
// Creates the directory and any necessary parent directories with permissions 0755.
//
// Parameters:
//   - outputDir: Path to the directory to create
//
// Returns:
//   - error: Any error encountered during directory creation
func EnsureOutputDirectory(outputDir string) error {
	return os.MkdirAll(outputDir, 0755)
}
