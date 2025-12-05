package code_examples

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLiteralIncludeDirective tests the parsing and extraction of literalinclude directives
func TestLiteralIncludeDirective(t *testing.T) {
	// Setup paths
	testDataDir := filepath.Join("..", "..", "..", "testdata")
	inputFile := filepath.Join(testDataDir, "input-files", "source", "literalinclude-test.rst")
	expectedOutputDir := filepath.Join(testDataDir, "expected-output")

	// Create temporary output directory
	tempDir, err := os.MkdirTemp("", "audit-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Run the extract command
	report, err := RunExtract(inputFile, tempDir, false, false, false, false)
	if err != nil {
		t.Fatalf("RunExtract failed: %v", err)
	}

	// Verify the report
	if report.FilesTraversed != 1 {
		t.Errorf("Expected 1 file traversed, got %d", report.FilesTraversed)
	}

	if report.OutputFilesWritten != 7 {
		t.Errorf("Expected 7 output files, got %d", report.OutputFilesWritten)
	}

	// Expected output files
	expectedFiles := []string{
		"literalinclude-test.literalinclude.1.py",
		"literalinclude-test.literalinclude.2.go",
		"literalinclude-test.literalinclude.3.js",
		"literalinclude-test.literalinclude.4.php",
		"literalinclude-test.literalinclude.5.rb",
		"literalinclude-test.literalinclude.6.ts",
		"literalinclude-test.literalinclude.7.cpp",
	}

	// Compare each output file with expected
	for _, filename := range expectedFiles {
		actualPath := filepath.Join(tempDir, filename)
		expectedPath := filepath.Join(expectedOutputDir, filename)

		// Read actual output
		actualContent, err := os.ReadFile(actualPath)
		if err != nil {
			t.Errorf("Failed to read actual output file %s: %v", filename, err)
			continue
		}

		// Read expected output
		expectedContent, err := os.ReadFile(expectedPath)
		if err != nil {
			t.Errorf("Failed to read expected output file %s: %v", filename, err)
			continue
		}

		// Compare content
		if string(actualContent) != string(expectedContent) {
			t.Errorf("Content mismatch for %s\nExpected:\n%s\n\nActual:\n%s",
				filename, string(expectedContent), string(actualContent))
		}
	}

	// Verify language counts
	expectedLanguages := map[string]int{
		"python":     1,
		"go":         1,
		"javascript": 1,
		"php":        1,
		"ruby":       1,
		"typescript": 1,
		"cpp":        1,
	}

	for lang, expectedCount := range expectedLanguages {
		if actualCount := report.LanguageCounts[lang]; actualCount != expectedCount {
			t.Errorf("Expected %d %s examples, got %d", expectedCount, lang, actualCount)
		}
	}

	// Verify directive counts
	if count := report.DirectiveCounts[LiteralInclude]; count != 7 {
		t.Errorf("Expected 7 literalinclude directives, got %d", count)
	}
}

// TestIncludeDirectiveFollowing tests that include directives are followed correctly
func TestIncludeDirectiveFollowing(t *testing.T) {
	// Setup paths
	testDataDir := filepath.Join("..", "..", "..", "testdata")
	inputFile := filepath.Join(testDataDir, "input-files", "source", "include-test.rst")
	expectedOutputDir := filepath.Join(testDataDir, "expected-output")

	// Create temporary output directory
	tempDir, err := os.MkdirTemp("", "audit-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Run the extract command with include following enabled
	report, err := RunExtract(inputFile, tempDir, false, true, false, false)
	if err != nil {
		t.Fatalf("RunExtract failed: %v", err)
	}

	// Verify that multiple files were traversed (main file + includes)
	if report.FilesTraversed < 2 {
		t.Errorf("Expected at least 2 files traversed (with includes), got %d", report.FilesTraversed)
	}

	// Verify output file was created
	if report.OutputFilesWritten != 1 {
		t.Errorf("Expected 1 output file, got %d", report.OutputFilesWritten)
	}

	// Compare output with expected
	// The literalinclude is in examples.rst (included file), so output is named after that
	actualPath := filepath.Join(tempDir, "examples.literalinclude.1.go")
	expectedPath := filepath.Join(expectedOutputDir, "examples.literalinclude.1.go")

	actualContent, err := os.ReadFile(actualPath)
	if err != nil {
		t.Fatalf("Failed to read actual output: %v", err)
	}

	expectedContent, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatalf("Failed to read expected output: %v", err)
	}

	if string(actualContent) != string(expectedContent) {
		t.Errorf("Content mismatch\nExpected:\n%s\n\nActual:\n%s",
			string(expectedContent), string(actualContent))
	}

	// Verify the language was normalized (golang -> go)
	if count := report.LanguageCounts["go"]; count != 1 {
		t.Errorf("Expected 1 go example (normalized from golang), got %d", count)
	}
}

// TestEmptyFile tests handling of files with no directives
func TestCodeBlockDirective(t *testing.T) {
	// Setup paths
	testDataDir := filepath.Join("..", "..", "..", "testdata")
	inputFile := filepath.Join(testDataDir, "input-files", "source", "code-block-test.rst")
	expectedOutputDir := filepath.Join(testDataDir, "expected-output")

	// Create temp directory for output
	tempDir, err := os.MkdirTemp("", "audit-test-code-block-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Run extract on code-block test file
	report, err := RunExtract(inputFile, tempDir, false, false, false, false)
	if err != nil {
		t.Fatalf("RunExtract failed: %v", err)
	}

	// Verify report
	if report.FilesTraversed != 1 {
		t.Errorf("Expected 1 file traversed, got %d", report.FilesTraversed)
	}

	if report.OutputFilesWritten != 7 {
		t.Errorf("Expected 7 output files, got %d", report.OutputFilesWritten)
	}

	// Expected output files
	expectedFiles := []string{
		"code-block-test.code-block.1.js",  // JavaScript with language
		"code-block-test.code-block.2.py",  // Python with options
		"code-block-test.code-block.3.js",  // JSON array example
		"code-block-test.code-block.4.txt", // No language (undefined)
		"code-block-test.code-block.5.sh",  // Shell script
		"code-block-test.code-block.6.ts",  // TypeScript normalization
		"code-block-test.code-block.7.cpp", // C++ normalization
	}

	// Compare each output file with expected
	for _, filename := range expectedFiles {
		actualPath := filepath.Join(tempDir, filename)
		expectedPath := filepath.Join(expectedOutputDir, filename)

		actualContent, err := os.ReadFile(actualPath)
		if err != nil {
			t.Errorf("Failed to read actual file %s: %v", filename, err)
			continue
		}

		expectedContent, err := os.ReadFile(expectedPath)
		if err != nil {
			t.Errorf("Failed to read expected file %s: %v", filename, err)
			continue
		}

		if string(actualContent) != string(expectedContent) {
			t.Errorf("Content mismatch for %s\nExpected:\n%s\n\nActual:\n%s",
				filename, string(expectedContent), string(actualContent))
		}
	}
}

func TestNestedCodeBlockDirective(t *testing.T) {
	// Setup paths
	testDataDir := filepath.Join("..", "..", "..", "testdata")
	inputFile := filepath.Join(testDataDir, "input-files", "source", "nested-code-block-test.rst")
	expectedOutputDir := filepath.Join(testDataDir, "expected-output")

	// Create temp directory for output
	tempDir, err := os.MkdirTemp("", "audit-test-nested-code-block-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Run extract on nested code-block test file
	report, err := RunExtract(inputFile, tempDir, false, false, false, false)
	if err != nil {
		t.Fatalf("RunExtract failed: %v", err)
	}

	// Verify we found 11 code blocks
	if report.OutputFilesWritten != 11 {
		t.Errorf("Expected 11 output files, got %d", report.OutputFilesWritten)
	}

	// Verify all are code-block directives
	if report.DirectiveCounts[CodeBlock] != 11 {
		t.Errorf("Expected 11 code-block directives, got %d", report.DirectiveCounts[CodeBlock])
	}

	// Expected files and their languages
	expectedFiles := map[string]string{
		"nested-code-block-test.code-block.1.js":   "javascript",
		"nested-code-block-test.code-block.2.js":   "javascript",
		"nested-code-block-test.code-block.3.js":   "javascript",
		"nested-code-block-test.code-block.4.py":   "python",
		"nested-code-block-test.code-block.5.go":   "go",
		"nested-code-block-test.code-block.6.ts":   "typescript",
		"nested-code-block-test.code-block.7.ts":   "typescript",
		"nested-code-block-test.code-block.8.sh":   "shell",
		"nested-code-block-test.code-block.9.rb":   "ruby",
		"nested-code-block-test.code-block.10.rb":  "ruby",
		"nested-code-block-test.code-block.11.txt": "undefined",
	}

	// Verify each expected file exists and matches
	for filename := range expectedFiles {
		actualPath := filepath.Join(tempDir, filename)
		expectedPath := filepath.Join(expectedOutputDir, filename)

		// Check file exists
		if _, err := os.Stat(actualPath); os.IsNotExist(err) {
			t.Errorf("Expected output file not created: %s", filename)
			continue
		}

		// Compare content
		actualContent, err := os.ReadFile(actualPath)
		if err != nil {
			t.Errorf("Failed to read actual file %s: %v", filename, err)
			continue
		}

		expectedContent, err := os.ReadFile(expectedPath)
		if err != nil {
			t.Errorf("Failed to read expected file %s: %v", filename, err)
			continue
		}

		if string(actualContent) != string(expectedContent) {
			t.Errorf("Content mismatch for %s\nExpected:\n%s\n\nActual:\n%s",
				filename, string(expectedContent), string(actualContent))
		}
	}
}

func TestIoCodeBlockDirective(t *testing.T) {
	// Setup paths
	testDataDir := filepath.Join("..", "..", "..", "testdata")
	inputFile := filepath.Join(testDataDir, "input-files", "source", "io-code-block-test.rst")
	expectedOutputDir := filepath.Join(testDataDir, "expected-output")

	// Create temp directory for output
	tempDir, err := os.MkdirTemp("", "audit-test-io-code-block-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Run extract on io-code-block test file
	report, err := RunExtract(inputFile, tempDir, false, false, false, false)
	if err != nil {
		t.Fatalf("RunExtract failed: %v", err)
	}

	// Verify we found 11 code examples (7 directives, but Test 2 fails, Test 7 has no output)
	// Test 1: input + output = 2
	// Test 2: fails (file not found) = 0
	// Test 3: input + output = 2
	// Test 4: input + output = 2
	// Test 5: input + output = 2
	// Test 6: input + output = 2
	// Test 7: input only = 1
	// Total: 11
	if report.OutputFilesWritten != 11 {
		t.Errorf("Expected 11 output files, got %d", report.OutputFilesWritten)
	}

	// Verify all are io-code-block directives
	if report.DirectiveCounts[IoCodeBlock] != 11 {
		t.Errorf("Expected 11 io-code-block examples, got %d", report.DirectiveCounts[IoCodeBlock])
	}

	// Expected files
	expectedFiles := []string{
		// Test 1: Inline input/output (JavaScript)
		"io-code-block-test.io-code-block.1.input.js",
		"io-code-block-test.io-code-block.1.output.js",
		// Test 2: File-based (skipped - files don't exist)
		// Test 3: Python inline
		"io-code-block-test.io-code-block.3.input.py",
		"io-code-block-test.io-code-block.3.output.py",
		// Test 4: Shell command
		"io-code-block-test.io-code-block.4.input.sh",
		"io-code-block-test.io-code-block.4.output.txt",
		// Test 5: TypeScript
		"io-code-block-test.io-code-block.5.input.ts",
		"io-code-block-test.io-code-block.5.output.txt",
		// Test 6: Nested in procedure
		"io-code-block-test.io-code-block.6.input.js",
		"io-code-block-test.io-code-block.6.output.js",
		// Test 7: Input only (Go)
		"io-code-block-test.io-code-block.7.input.go",
	}

	// Verify each expected file exists and matches
	for _, filename := range expectedFiles {
		actualPath := filepath.Join(tempDir, filename)
		expectedPath := filepath.Join(expectedOutputDir, filename)

		// Check file exists
		if _, err := os.Stat(actualPath); os.IsNotExist(err) {
			t.Errorf("Expected output file not created: %s", filename)
			continue
		}

		// Compare content
		actualContent, err := os.ReadFile(actualPath)
		if err != nil {
			t.Errorf("Failed to read actual file %s: %v", filename, err)
			continue
		}

		expectedContent, err := os.ReadFile(expectedPath)
		if err != nil {
			t.Errorf("Failed to read expected file %s: %v", filename, err)
			continue
		}

		if string(actualContent) != string(expectedContent) {
			t.Errorf("Content mismatch for %s\nExpected:\n%s\n\nActual:\n%s",
				filename, string(expectedContent), string(actualContent))
		}
	}
}

func TestEmptyFile(t *testing.T) {
	// Create a temporary file with no directives
	tempDir, err := os.MkdirTemp("", "audit-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a source directory structure
	sourceDir := filepath.Join(tempDir, "source")
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatalf("Failed to create source directory: %v", err)
	}

	emptyFile := filepath.Join(sourceDir, "empty.rst")
	if err := os.WriteFile(emptyFile, []byte("Empty File\n==========\n\nNo directives here."), 0644); err != nil {
		t.Fatalf("Failed to create empty file: %v", err)
	}

	outputDir := filepath.Join(tempDir, "output")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("Failed to create output directory: %v", err)
	}

	// Run the extract command
	report, err := RunExtract(emptyFile, outputDir, false, false, false, false)
	if err != nil {
		t.Fatalf("RunExtract failed: %v", err)
	}

	// Verify no output files were created
	if report.OutputFilesWritten != 0 {
		t.Errorf("Expected 0 output files for empty file, got %d", report.OutputFilesWritten)
	}

	// Verify the file was still traversed
	if report.FilesTraversed != 1 {
		t.Errorf("Expected 1 file traversed, got %d", report.FilesTraversed)
	}
}

// TestRecursiveDirectoryScanning tests that -r flag scans all files in subdirectories
func TestRecursiveDirectoryScanning(t *testing.T) {
	// Setup paths
	testDataDir := filepath.Join("..", "..", "..", "testdata")
	inputDir := filepath.Join(testDataDir, "input-files", "source")

	// Create temporary output directory
	tempDir, err := os.MkdirTemp("", "audit-test-recursive-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Run the extract command with recursive=true, followIncludes=false
	report, err := RunExtract(inputDir, tempDir, true, false, false, false)
	if err != nil {
		t.Fatalf("RunExtract failed: %v", err)
	}

	// Verify that multiple files were traversed
	// Should find all .rst files in source/ and source/includes/
	// Expected: code-block-test.rst, include-test.rst, io-code-block-test.rst,
	//           literalinclude-test.rst, nested-code-block-test.rst,
	//           includes/examples.rst, includes/intro.rst
	expectedMinFiles := 7
	if report.FilesTraversed < expectedMinFiles {
		t.Errorf("Expected at least %d files traversed with recursive scan, got %d",
			expectedMinFiles, report.FilesTraversed)
	}

	// Verify that code examples were extracted from multiple files
	// Without following includes, include-test.rst should have 0 examples
	// but all other files should have examples
	if report.OutputFilesWritten < 30 {
		t.Errorf("Expected at least 30 output files from recursive scan, got %d",
			report.OutputFilesWritten)
	}

	// Verify we have examples from different directive types
	if report.DirectiveCounts[CodeBlock] == 0 {
		t.Error("Expected code-block directives to be found")
	}
	if report.DirectiveCounts[LiteralInclude] == 0 {
		t.Error("Expected literalinclude directives to be found")
	}
	if report.DirectiveCounts[IoCodeBlock] == 0 {
		t.Error("Expected io-code-block directives to be found")
	}
}

// TestFollowIncludesWithoutRecursive tests that -f flag follows includes in a single file
func TestFollowIncludesWithoutRecursive(t *testing.T) {
	// Setup paths
	testDataDir := filepath.Join("..", "..", "..", "testdata")
	inputFile := filepath.Join(testDataDir, "input-files", "source", "include-test.rst")

	// Create temporary output directory
	tempDir, err := os.MkdirTemp("", "audit-test-follow-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Run the extract command with recursive=false, followIncludes=true
	report, err := RunExtract(inputFile, tempDir, false, true, false, false)
	if err != nil {
		t.Fatalf("RunExtract failed: %v", err)
	}

	// Verify that multiple files were traversed (main file + includes)
	// include-test.rst includes intro.rst and examples.rst
	expectedFiles := 3
	if report.FilesTraversed != expectedFiles {
		t.Errorf("Expected %d files traversed (main + 2 includes), got %d",
			expectedFiles, report.FilesTraversed)
	}

	// Verify that the code example from the included file was extracted
	// examples.rst has 1 literalinclude directive
	if report.OutputFilesWritten != 1 {
		t.Errorf("Expected 1 output file from included files, got %d",
			report.OutputFilesWritten)
	}

	// Verify the directive type
	if report.DirectiveCounts[LiteralInclude] != 1 {
		t.Errorf("Expected 1 literalinclude directive, got %d",
			report.DirectiveCounts[LiteralInclude])
	}
}

// TestRecursiveWithFollowIncludes tests that -r and -f together work correctly
func TestRecursiveWithFollowIncludes(t *testing.T) {
	// Setup paths
	testDataDir := filepath.Join("..", "..", "..", "testdata")
	inputDir := filepath.Join(testDataDir, "input-files", "source")

	// Create temporary output directory
	tempDir, err := os.MkdirTemp("", "audit-test-both-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Run the extract command with recursive=true, followIncludes=true
	report, err := RunExtract(inputDir, tempDir, true, true, false, false)
	if err != nil {
		t.Fatalf("RunExtract failed: %v", err)
	}

	// Verify that multiple files were traversed
	// Should find all .rst files in source/ and source/includes/
	expectedMinFiles := 7
	if report.FilesTraversed < expectedMinFiles {
		t.Errorf("Expected at least %d files traversed, got %d",
			expectedMinFiles, report.FilesTraversed)
	}

	// Verify that code examples were extracted
	// This should be the same as recursive-only since the include files
	// are already found by recursive directory scanning
	if report.OutputFilesWritten < 30 {
		t.Errorf("Expected at least 30 output files, got %d",
			report.OutputFilesWritten)
	}

	// Verify we have examples from all directive types
	if report.DirectiveCounts[CodeBlock] == 0 {
		t.Error("Expected code-block directives to be found")
	}
	if report.DirectiveCounts[LiteralInclude] == 0 {
		t.Error("Expected literalinclude directives to be found")
	}
	if report.DirectiveCounts[IoCodeBlock] == 0 {
		t.Error("Expected io-code-block directives to be found")
	}
}

// TestNoFlagsOnDirectory tests that without -r flag, directory is not scanned
func TestNoFlagsOnDirectory(t *testing.T) {
	// Setup paths
	testDataDir := filepath.Join("..", "..", "..", "testdata")
	inputDir := filepath.Join(testDataDir, "input-files", "source")

	// Create temporary output directory
	tempDir, err := os.MkdirTemp("", "audit-test-noflags-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Run the extract command with recursive=false, followIncludes=false on a directory
	report, err := RunExtract(inputDir, tempDir, false, false, false, false)
	if err != nil {
		t.Fatalf("RunExtract failed: %v", err)
	}

	// Without recursive flag, should only process files in the top-level directory
	// Should NOT include files in includes/ subdirectory
	// Expected: code-block-test.rst, duplicate-include-test.rst, include-test.rst,
	//           io-code-block-test.rst, literalinclude-test.rst, nested-code-block-test.rst,
	//           nested-include-test.rst, index.rst, procedure-test.rst, procedure-with-includes.rst (10 files)
	expectedFiles := 11
	if report.FilesTraversed != expectedFiles {
		t.Errorf("Expected %d files traversed (top-level only), got %d",
			expectedFiles, report.FilesTraversed)
	}

	// Without followIncludes, include-test.rst should have 0 examples
	// So we should have examples from the other 4 files
	if report.OutputFilesWritten < 30 {
		t.Errorf("Expected at least 30 output files, got %d",
			report.OutputFilesWritten)
	}
}
