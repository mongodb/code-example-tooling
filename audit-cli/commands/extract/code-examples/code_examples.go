// Package code_examples provides functionality for extracting code examples from RST files.
//
// This package implements the "extract code-examples" subcommand, which parses
// reStructuredText files and extracts code examples from various directives:
//   - literalinclude: External file references with optional partial extraction
//   - code-block: Inline code blocks with automatic dedenting
//   - io-code-block: Input/output examples with nested directives
//
// The extracted code examples are written to individual files with standardized naming:
//   {source-base}.{directive-type}.{index}.{ext}
//
// Supports recursive directory scanning and following include directives to process
// entire documentation trees.
package code_examples

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// NewCodeExamplesCommand creates the code-examples subcommand.
//
// This command extracts code examples from RST files and writes them to individual
// files in the output directory. Supports various flags for controlling behavior:
//   - -r, --recursive: Recursively scan directories for RST files
//   - -f, --follow-includes: Follow .. include:: directives
//   - -o, --output: Output directory for extracted files
//   - --dry-run: Show what would be extracted without writing files
//   - -v, --verbose: Show detailed processing information
//   - --preserve-dirs: Preserve directory structure when used with --recursive
func NewCodeExamplesCommand() *cobra.Command {
	var (
		recursive      bool
		followIncludes bool
		outputDir      string
		dryRun         bool
		verbose        bool
		preserveDirs   bool
	)

	cmd := &cobra.Command{
		Use:   "code-examples [filepath]",
		Short: "Extract code examples from reStructuredText files",
		Long: `Extract code examples from reStructuredText directives (code-block, literalinclude, io-code-block)
and output them as individual files.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			filePath := args[0]
			return runExtract(filePath, recursive, followIncludes, outputDir, dryRun, verbose, preserveDirs)
		},
	}

	cmd.Flags().BoolVarP(&recursive, "recursive", "r", false, "Recursively scan directories for files to process")
	cmd.Flags().BoolVarP(&followIncludes, "follow-includes", "f", false, "Follow .. include:: directives in RST files")
	cmd.Flags().StringVarP(&outputDir, "output", "o", "./output", "Output directory for code example files")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be outputted without writing files")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Provide additional information during execution")
	cmd.Flags().BoolVar(&preserveDirs, "preserve-dirs", false, "Preserve directory structure in output (use with --recursive)")

	return cmd
}

// RunExtract executes the extraction operation and returns the report.
//
// This function is exported for use in tests. It extracts code examples from the
// specified file or directory and writes them to the output directory.
//
// Parameters:
//   - filePath: Path to RST file or directory to process
//   - outputDir: Directory where extracted files will be written
//   - recursive: If true, recursively scan directories for RST files
//   - followIncludes: If true, follow .. include:: directives
//   - dryRun: If true, show what would be extracted without writing files
//   - verbose: If true, show detailed processing information
//   - preserveDirs: If true, preserve directory structure in output (use with recursive)
//
// Returns:
//   - *Report: Statistics about the extraction operation
//   - error: Any error encountered during extraction
func RunExtract(filePath string, outputDir string, recursive bool, followIncludes bool, dryRun bool, verbose bool, preserveDirs bool) (*Report, error) {
	report, err := runExtractInternal(filePath, recursive, followIncludes, outputDir, dryRun, verbose, preserveDirs)
	return report, err
}

// runExtract executes the extraction operation (internal wrapper for CLI).
//
// This is a thin wrapper around runExtractInternal that discards the report
// and only returns errors, suitable for use in the CLI command handler.
func runExtract(filePath string, recursive bool, followIncludes bool, outputDir string, dryRun bool, verbose bool, preserveDirs bool) error {
	_, err := runExtractInternal(filePath, recursive, followIncludes, outputDir, dryRun, verbose, preserveDirs)
	return err
}

// runExtractInternal executes the extraction operation
func runExtractInternal(filePath string, recursive bool, followIncludes bool, outputDir string, dryRun bool, verbose bool, preserveDirs bool) (*Report, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to access path %s: %w", filePath, err)
	}

	report := NewReport()

	var filesToProcess []string
	var rootPath string

	if fileInfo.IsDir() {
		if verbose {
			fmt.Printf("Scanning directory: %s (recursive: %v)\n", filePath, recursive)
		}
		filesToProcess, err = TraverseDirectory(filePath, recursive)
		if err != nil {
			return nil, fmt.Errorf("failed to traverse directory: %w", err)
		}
		rootPath = filePath
	} else {
		filesToProcess = []string{filePath}
		rootPath = ""
	}

	var filteredFiles []string
	for _, file := range filesToProcess {
		if ShouldProcessFile(file) {
			filteredFiles = append(filteredFiles, file)
		}
	}
	filesToProcess = filteredFiles

	if verbose {
		fmt.Printf("Found %d files to process\n", len(filesToProcess))
	}

	if !dryRun {
		if err := EnsureOutputDirectory(outputDir); err != nil {
			return nil, fmt.Errorf("failed to create output directory: %w", err)
		}
	}

	// Track visited files to prevent circular includes
	visited := make(map[string]bool)

	for _, file := range filesToProcess {
		if verbose {
			fmt.Printf("Processing: %s\n", file)
		}

		// Use ParseFileWithIncludes to follow include directives when followIncludes flag is set
		examples, processedFiles, err := ParseFileWithIncludes(file, followIncludes, visited, verbose)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to parse %s: %v\n", file, err)
			continue
		}

		// Add all processed files (including includes) to the report
		for _, processedFile := range processedFiles {
			report.AddTraversedFile(processedFile)
		}

		for _, example := range examples {
			outputPath, err := WriteCodeExample(example, outputDir, rootPath, dryRun, preserveDirs)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to write code example: %v\n", err)
				continue
			}

			if verbose {
				if dryRun {
					fmt.Printf("  [DRY RUN] Would write: %s\n", outputPath)
				} else {
					fmt.Printf("  Wrote: %s\n", outputPath)
				}
			}

			report.AddCodeExample(example, outputPath)
			if !dryRun {
				report.OutputFilesWritten++
			}
		}
	}

	if dryRun {
		fmt.Println("\n[DRY RUN MODE - No files were written]")
	}
	PrintReport(report, verbose)

	return report, nil
}
