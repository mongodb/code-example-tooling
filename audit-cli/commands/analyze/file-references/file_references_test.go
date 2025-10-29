package filereferences

import (
	"path/filepath"
	"testing"
)

// TestAnalyzeReferences tests the AnalyzeReferences function with various scenarios.
func TestAnalyzeReferences(t *testing.T) {
	// Get the testdata directory
	testDataDir := "../../../testdata/input-files/source"

	tests := []struct {
		name                  string
		targetFile            string
		expectedReferences    int
		expectedDirectiveType string
	}{
		{
			name:               "Include file with multiple references",
			targetFile:         "includes/intro.rst",
			expectedReferences: 5, // 4 RST files + 1 YAML file (no toctree by default)
			expectedDirectiveType: "include",
		},
		{
			name:               "Code example with literalinclude",
			targetFile:         "code-examples/example.py",
			expectedReferences: 2, // 1 RST file + 1 YAML file
			expectedDirectiveType: "literalinclude",
		},
		{
			name:               "Code example with multiple directive types",
			targetFile:         "code-examples/example.js",
			expectedReferences: 2, // literalinclude + io-code-block
			expectedDirectiveType: "", // mixed types
		},
		{
			name:               "File with no references",
			targetFile:         "code-block-test.rst",
			expectedReferences: 0,
			expectedDirectiveType: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Build absolute path to target file
			targetPath := filepath.Join(testDataDir, tt.targetFile)
			absTargetPath, err := filepath.Abs(targetPath)
			if err != nil {
				t.Fatalf("failed to get absolute path: %v", err)
			}

			// Run analysis (without toctree by default)
			analysis, err := AnalyzeReferences(absTargetPath, false)
			if err != nil {
				t.Fatalf("AnalyzeReferences failed: %v", err)
			}

			// Check total references
			if analysis.TotalReferences != tt.expectedReferences {
				t.Errorf("expected %d references, got %d", tt.expectedReferences, analysis.TotalReferences)
			}

			// Check that we got the expected number of referencing files
			if len(analysis.ReferencingFiles) != tt.expectedReferences {
				t.Errorf("expected %d referencing files, got %d", tt.expectedReferences, len(analysis.ReferencingFiles))
			}

			// If we expect a specific directive type, check it
			if tt.expectedDirectiveType != "" && tt.expectedReferences > 0 {
				foundExpectedType := false
				for _, ref := range analysis.ReferencingFiles {
					if ref.DirectiveType == tt.expectedDirectiveType {
						foundExpectedType = true
						break
					}
				}
				if !foundExpectedType {
					t.Errorf("expected to find directive type %s, but didn't", tt.expectedDirectiveType)
				}
			}

			// Verify source directory was found
			if analysis.SourceDir == "" {
				t.Error("source directory should not be empty")
			}

			// Verify target file matches
			if analysis.TargetFile != absTargetPath {
				t.Errorf("expected target file %s, got %s", absTargetPath, analysis.TargetFile)
			}
		})
	}
}

// TestFindReferencesInFile tests the findReferencesInFile function.
func TestFindReferencesInFile(t *testing.T) {
	testDataDir := "../../../testdata/input-files/source"
	sourceDir := testDataDir

	tests := []struct {
		name               string
		searchFile         string
		targetFile         string
		expectedReferences int
		expectedDirective  string
		includeToctree     bool
	}{
		{
			name:               "Include directive",
			searchFile:         "include-test.rst",
			targetFile:         "includes/intro.rst",
			expectedReferences: 1,
			expectedDirective:  "include",
			includeToctree:     false,
		},
		{
			name:               "Literalinclude directive",
			searchFile:         "literalinclude-test.rst",
			targetFile:         "code-examples/example.py",
			expectedReferences: 1,
			expectedDirective:  "literalinclude",
			includeToctree:     false,
		},
		{
			name:               "IO code block directive",
			searchFile:         "io-code-block-test.rst",
			targetFile:         "code-examples/example.js",
			expectedReferences: 1,
			expectedDirective:  "io-code-block",
			includeToctree:     false,
		},
		{
			name:               "Duplicate includes",
			searchFile:         "duplicate-include-test.rst",
			targetFile:         "includes/intro.rst",
			expectedReferences: 2, // Same file included twice
			expectedDirective:  "include",
			includeToctree:     false,
		},
		{
			name:               "Toctree directive",
			searchFile:         "index.rst",
			targetFile:         "include-test.rst",
			expectedReferences: 1,
			expectedDirective:  "toctree",
			includeToctree:     true, // Must enable toctree flag
		},
		{
			name:               "No references",
			searchFile:         "code-block-test.rst",
			targetFile:         "includes/intro.rst",
			expectedReferences: 0,
			expectedDirective:  "",
			includeToctree:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			searchPath := filepath.Join(testDataDir, tt.searchFile)
			targetPath := filepath.Join(testDataDir, tt.targetFile)

			// Get absolute paths
			absSearchPath, err := filepath.Abs(searchPath)
			if err != nil {
				t.Fatalf("failed to get absolute search path: %v", err)
			}
			absTargetPath, err := filepath.Abs(targetPath)
			if err != nil {
				t.Fatalf("failed to get absolute target path: %v", err)
			}
			absSourceDir, err := filepath.Abs(sourceDir)
			if err != nil {
				t.Fatalf("failed to get absolute source dir: %v", err)
			}

			refs, err := findReferencesInFile(absSearchPath, absTargetPath, absSourceDir, tt.includeToctree)
			if err != nil {
				t.Fatalf("findReferencesInFile failed: %v", err)
			}

			if len(refs) != tt.expectedReferences {
				t.Errorf("expected %d references, got %d", tt.expectedReferences, len(refs))
			}

			// Check directive type if we expect references
			if tt.expectedReferences > 0 && tt.expectedDirective != "" {
				for _, ref := range refs {
					if ref.DirectiveType != tt.expectedDirective {
						t.Errorf("expected directive type %s, got %s", tt.expectedDirective, ref.DirectiveType)
					}
				}
			}

			// Verify all references have required fields
			for _, ref := range refs {
				if ref.FilePath == "" {
					t.Error("reference should have a file path")
				}
				if ref.DirectiveType == "" {
					t.Error("reference should have a directive type")
				}
				if ref.ReferencePath == "" {
					t.Error("reference should have a reference path")
				}
				if ref.LineNumber == 0 {
					t.Error("reference should have a line number")
				}
			}
		})
	}
}

// TestReferencesTarget tests the referencesTarget function.
func TestReferencesTarget(t *testing.T) {
	testDataDir := "../../../testdata/input-files/source"

	// Get absolute paths
	absTestDataDir, err := filepath.Abs(testDataDir)
	if err != nil {
		t.Fatalf("failed to get absolute test data dir: %v", err)
	}

	tests := []struct {
		name        string
		refPath     string
		targetFile  string
		currentFile string
		expected    bool
	}{
		{
			name:        "Absolute path match",
			refPath:     "/includes/intro.rst",
			targetFile:  filepath.Join(absTestDataDir, "includes/intro.rst"),
			currentFile: filepath.Join(absTestDataDir, "include-test.rst"),
			expected:    true,
		},
		{
			name:        "Absolute path no match",
			refPath:     "/includes/intro.rst",
			targetFile:  filepath.Join(absTestDataDir, "includes/examples.rst"),
			currentFile: filepath.Join(absTestDataDir, "include-test.rst"),
			expected:    false,
		},
		{
			name:        "Relative path match",
			refPath:     "intro.rst",
			targetFile:  filepath.Join(absTestDataDir, "includes/intro.rst"),
			currentFile: filepath.Join(absTestDataDir, "includes/nested-include.rst"),
			expected:    true,
		},
		{
			name:        "Relative path no match",
			refPath:     "intro.rst",
			targetFile:  filepath.Join(absTestDataDir, "includes/examples.rst"),
			currentFile: filepath.Join(absTestDataDir, "includes/nested-include.rst"),
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := referencesTarget(tt.refPath, tt.targetFile, absTestDataDir, tt.currentFile)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestGroupByDirectiveType tests the groupByDirectiveType function.
func TestGroupByDirectiveType(t *testing.T) {
	refs := []FileReference{
		{DirectiveType: "include", FilePath: "file1.rst"},
		{DirectiveType: "include", FilePath: "file2.rst"},
		{DirectiveType: "literalinclude", FilePath: "file3.rst"},
		{DirectiveType: "io-code-block", FilePath: "file4.rst"},
		{DirectiveType: "include", FilePath: "file5.rst"},
	}

	groups := groupByDirectiveType(refs)

	// Check that we have 3 groups
	if len(groups) != 3 {
		t.Errorf("expected 3 groups, got %d", len(groups))
	}

	// Check include group
	if len(groups["include"]) != 3 {
		t.Errorf("expected 3 include references, got %d", len(groups["include"]))
	}

	// Check literalinclude group
	if len(groups["literalinclude"]) != 1 {
		t.Errorf("expected 1 literalinclude reference, got %d", len(groups["literalinclude"]))
	}

	// Check io-code-block group
	if len(groups["io-code-block"]) != 1 {
		t.Errorf("expected 1 io-code-block reference, got %d", len(groups["io-code-block"]))
	}
}

