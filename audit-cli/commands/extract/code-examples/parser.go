package code_examples

import (
	"fmt"
	"os"

	"github.com/mongodb/code-example-tooling/audit-cli/internal/rst"
)

// ParseFile parses a file and extracts code examples from reStructuredText directives.
//
// This function parses all supported RST directives (literalinclude, code-block, io-code-block)
// and converts them into CodeExample structs ready for writing to files.
//
// Parameters:
//   - filePath: Path to the RST file to parse
//
// Returns:
//   - []CodeExample: Slice of extracted code examples
//   - error: Any error encountered during parsing
func ParseFile(filePath string) ([]CodeExample, error) {
	// Parse all directives from the file
	directives, err := rst.ParseDirectives(filePath)
	if err != nil {
		return nil, err
	}

	var examples []CodeExample
	directiveCounts := make(map[rst.DirectiveType]int)

	for _, directive := range directives {
		// Track directive index for this type
		directiveCounts[directive.Type]++
		index := directiveCounts[directive.Type]

		switch directive.Type {
		case rst.LiteralInclude:
			example, err := parseLiteralInclude(filePath, directive, index)
			if err != nil {
				// Log warning but continue processing
				fmt.Fprintf(os.Stderr, "Warning: failed to parse literalinclude at line %d in %s: %v\n",
					directive.LineNum, filePath, err)
				continue
			}
			examples = append(examples, example)

		case rst.CodeBlock:
			example, err := parseCodeBlock(filePath, directive, index)
			if err != nil {
				// Log warning but continue processing
				fmt.Fprintf(os.Stderr, "Warning: failed to parse code-block at line %d in %s: %v\n",
					directive.LineNum, filePath, err)
				continue
			}
			examples = append(examples, example)

		case rst.IoCodeBlock:
			examples = append(examples, parseIoCodeBlock(filePath, directive, index)...)
			continue
		}
	}

	return examples, nil
}

// parseLiteralInclude parses a literalinclude directive and extracts the code content
func parseLiteralInclude(sourceFile string, directive rst.Directive, index int) (CodeExample, error) {
	// Extract the content from the referenced file
	content, err := rst.ExtractLiteralIncludeContent(sourceFile, directive)
	if err != nil {
		return CodeExample{}, err
	}

	// Get the language from the :language: option
	language := directive.Options["language"]
	if language == "" {
		language = Undefined
	}

	// Normalize the language
	language = NormalizeLanguage(language)

	return CodeExample{
		SourceFile:    sourceFile,
		DirectiveName: DirectiveType(directive.Type),
		Language:      language,
		Content:       content,
		Index:         index,
	}, nil
}

// parseCodeBlock parses a code-block directive and extracts the inline code content.
//
// The content is already dedented by the directive parser based on the first line's indentation.
// Language can be specified either as an argument (.. code-block:: javascript) or as an option (:language: javascript).
func parseCodeBlock(sourceFile string, directive rst.Directive, index int) (CodeExample, error) {
	// The content is already parsed and dedented by the directive parser
	content := directive.Content
	if content == "" {
		return CodeExample{}, fmt.Errorf("code-block has no content")
	}

	// Get the language from the directive argument (e.g., .. code-block:: javascript)
	// or from the :language: option
	language := directive.Argument
	if language == "" {
		language = directive.Options["language"]
	}
	if language == "" {
		language = Undefined
	}

	// Normalize the language
	language = NormalizeLanguage(language)

	return CodeExample{
		SourceFile:    sourceFile,
		DirectiveName: DirectiveType(directive.Type),
		Language:      language,
		Content:       content,
		Index:         index,
	}, nil
}

// ParseFileWithIncludes parses a file and recursively follows include directives.
//
// This function wraps the internal RST package's ParseFileWithIncludes to extract
// code examples from the main file and all included files.
//
// Parameters:
//   - filePath: Path to the RST file to parse
//   - followIncludes: If true, recursively follow .. include:: directives
//   - visited: Map tracking already-processed files to prevent circular includes
//   - verbose: If true, print detailed processing information
//
// Returns:
//   - []CodeExample: All code examples from the file and its includes
//   - []string: List of all processed file paths
//   - error: Any error encountered during parsing
func ParseFileWithIncludes(filePath string, followIncludes bool, visited map[string]bool, verbose bool) ([]CodeExample, []string, error) {
	var examples []CodeExample

	// Define the parse function that will be called for each file
	parseFunc := func(path string) error {
		fileExamples, err := ParseFile(path)
		if err != nil {
			return err
		}
		examples = append(examples, fileExamples...)
		return nil
	}

	// Use the internal RST package to handle include following
	processedFiles, err := rst.ParseFileWithIncludes(filePath, followIncludes, visited, verbose, parseFunc)
	if err != nil {
		return nil, processedFiles, err
	}

	return examples, processedFiles, nil
}

// TraverseDirectory recursively traverses a directory and returns all file paths.
//
// This is a wrapper around the internal RST package's TraverseDirectory function.
//
// Parameters:
//   - rootPath: Root directory to traverse
//   - recursive: If true, recursively scan subdirectories
//
// Returns:
//   - []string: List of all file paths found
//   - error: Any error encountered during traversal
func TraverseDirectory(rootPath string, recursive bool) ([]string, error) {
	return rst.TraverseDirectory(rootPath, recursive)
}

// ShouldProcessFile determines if a file should be processed based on its extension.
//
// This is a wrapper around the internal RST package's ShouldProcessFile function.
// Returns true for files with .rst, .txt, or .md extensions.
func ShouldProcessFile(filePath string) bool {
	return rst.ShouldProcessFile(filePath)
}

// parseIoCodeBlock parses an io-code-block directive and extracts input/output code examples
// Returns a slice of CodeExample (one for input, one for output if present)
func parseIoCodeBlock(sourceFile string, directive rst.Directive, index int) []CodeExample {
	var examples []CodeExample

	// Process input directive
	if directive.InputDirective != nil {
		inputExample, err := parseSubDirective(sourceFile, directive.InputDirective, "input", index)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to parse input directive at line %d in %s: %v\n",
				directive.LineNum, sourceFile, err)
		} else {
			examples = append(examples, inputExample)
		}
	}

	// Process output directive
	if directive.OutputDirective != nil {
		outputExample, err := parseSubDirective(sourceFile, directive.OutputDirective, "output", index)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to parse output directive at line %d in %s: %v\n",
				directive.LineNum, sourceFile, err)
		} else {
			examples = append(examples, outputExample)
		}
	}

	return examples
}

// parseSubDirective parses an input or output sub-directive within an io-code-block
func parseSubDirective(sourceFile string, subDir *rst.SubDirective, dirType string, index int) (CodeExample, error) {
	var content string
	var err error

	// If there's a filepath argument, read from the file
	if subDir.Argument != "" {
		content, err = rst.ExtractLiteralIncludeContent(sourceFile, rst.Directive{
			Argument: subDir.Argument,
			Options:  subDir.Options,
		})
		if err != nil {
			return CodeExample{}, fmt.Errorf("failed to read file %s: %w", subDir.Argument, err)
		}
	} else {
		// Use inline content
		content = subDir.Content
		if content == "" {
			return CodeExample{}, fmt.Errorf("%s directive has no content or filepath", dirType)
		}
	}

	// Get language from options
	language := subDir.Options["language"]
	if language == "" {
		language = Undefined
	}

	language = NormalizeLanguage(language)

	return CodeExample{
		SourceFile:    sourceFile,
		DirectiveName: DirectiveType(rst.IoCodeBlock),
		Language:      language,
		Content:       content,
		Index:         index,
		SubType:       dirType, // "input" or "output"
	}, nil
}
