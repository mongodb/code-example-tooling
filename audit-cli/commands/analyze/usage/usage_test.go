package usage

import (
	"path/filepath"
	"testing"
)

// TestAnalyzeUsage tests the AnalyzeUsage function with various scenarios.
func TestAnalyzeUsage(t *testing.T) {
	// Get the testdata directory
	testDataDir := "../../../testdata/input-files/source"

	tests := []struct {
		name                  string
		targetFile            string
		expectedUsages        int
		expectedDirectiveType string
	}{
		{
			name:               "Include file with multiple usages",
			targetFile:         "includes/intro.rst",
			expectedUsages:     5, // 4 RST files + 1 YAML file (no toctree by default)
			expectedDirectiveType: "include",
		},
		{
			name:               "Code example with literalinclude",
			targetFile:         "code-examples/example.py",
			expectedUsages:     2, // 1 RST file + 1 YAML file
			expectedDirectiveType: "literalinclude",
		},
		{
			name:               "Code example with multiple directive types",
			targetFile:         "code-examples/example.js",
			expectedUsages:     2, // literalinclude + io-code-block
			expectedDirectiveType: "", // mixed types
		},
		{
			name:               "File with no usages",
			targetFile:         "code-block-test.rst",
			expectedUsages:     0,
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

			// Run analysis (without toctree by default, not verbose, no exclude pattern)
			analysis, err := AnalyzeUsage(absTargetPath, false, false, "")
			if err != nil {
				t.Fatalf("AnalyzeUsage failed: %v", err)
			}

			// Check total usages
			if analysis.TotalUsages != tt.expectedUsages {
				t.Errorf("expected %d usages, got %d", tt.expectedUsages, analysis.TotalUsages)
			}

			// Check that we got the expected number of files using the target
			if len(analysis.UsingFiles) != tt.expectedUsages {
				t.Errorf("expected %d using files, got %d", tt.expectedUsages, len(analysis.UsingFiles))
			}

			// If we expect a specific directive type, check it
			if tt.expectedDirectiveType != "" && tt.expectedUsages > 0 {
				foundExpectedType := false
				for _, usage := range analysis.UsingFiles {
					if usage.DirectiveType == tt.expectedDirectiveType {
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

// TestFindUsagesInFile tests the findUsagesInFile function.
func TestFindUsagesInFile(t *testing.T) {
	testDataDir := "../../../testdata/input-files/source"
	sourceDir := testDataDir

	tests := []struct {
		name              string
		searchFile        string
		targetFile        string
		expectedUsages    int
		expectedDirective string
		includeToctree    bool
	}{
		{
			name:              "Include directive",
			searchFile:        "include-test.rst",
			targetFile:        "includes/intro.rst",
			expectedUsages:    1,
			expectedDirective: "include",
			includeToctree:    false,
		},
		{
			name:              "Literalinclude directive",
			searchFile:        "literalinclude-test.rst",
			targetFile:        "code-examples/example.py",
			expectedUsages:    1,
			expectedDirective: "literalinclude",
			includeToctree:    false,
		},
		{
			name:              "IO code block directive",
			searchFile:        "io-code-block-test.rst",
			targetFile:        "code-examples/example.js",
			expectedUsages:    1,
			expectedDirective: "io-code-block",
			includeToctree:    false,
		},
		{
			name:              "Duplicate includes",
			searchFile:        "duplicate-include-test.rst",
			targetFile:        "includes/intro.rst",
			expectedUsages:    2, // Same file included twice
			expectedDirective: "include",
			includeToctree:    false,
		},
		{
			name:              "Toctree directive",
			searchFile:        "index.rst",
			targetFile:        "include-test.rst",
			expectedUsages:    1,
			expectedDirective: "toctree",
			includeToctree:    true, // Must enable toctree flag
		},
		{
			name:              "No usages",
			searchFile:        "code-block-test.rst",
			targetFile:        "includes/intro.rst",
			expectedUsages:    0,
			expectedDirective: "",
			includeToctree:    false,
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

			usages, err := findUsagesInFile(absSearchPath, absTargetPath, absSourceDir, tt.includeToctree)
			if err != nil {
				t.Fatalf("findUsagesInFile failed: %v", err)
			}

			if len(usages) != tt.expectedUsages {
				t.Errorf("expected %d usages, got %d", tt.expectedUsages, len(usages))
			}

			// Check directive type if we expect usages
			if tt.expectedUsages > 0 && tt.expectedDirective != "" {
				for _, usage := range usages {
					if usage.DirectiveType != tt.expectedDirective {
						t.Errorf("expected directive type %s, got %s", tt.expectedDirective, usage.DirectiveType)
					}
				}
			}

			// Verify all usages have required fields
			for _, usage := range usages {
				if usage.FilePath == "" {
					t.Error("usage should have a file path")
				}
				if usage.DirectiveType == "" {
					t.Error("usage should have a directive type")
				}
				if usage.UsagePath == "" {
					t.Error("usage should have a usage path")
				}
				if usage.LineNumber == 0 {
					t.Error("usage should have a line number")
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
	usages := []FileUsage{
		{DirectiveType: "include", FilePath: "file1.rst"},
		{DirectiveType: "include", FilePath: "file2.rst"},
		{DirectiveType: "literalinclude", FilePath: "file3.rst"},
		{DirectiveType: "io-code-block", FilePath: "file4.rst"},
		{DirectiveType: "include", FilePath: "file5.rst"},
	}

	groups := groupByDirectiveType(usages)

	// Check that we have 3 groups
	if len(groups) != 3 {
		t.Errorf("expected 3 groups, got %d", len(groups))
	}

	// Check include group
	if len(groups["include"]) != 3 {
		t.Errorf("expected 3 include usages, got %d", len(groups["include"]))
	}

	// Check literalinclude group
	if len(groups["literalinclude"]) != 1 {
		t.Errorf("expected 1 literalinclude usage, got %d", len(groups["literalinclude"]))
	}

	// Check io-code-block group
	if len(groups["io-code-block"]) != 1 {
		t.Errorf("expected 1 io-code-block usage, got %d", len(groups["io-code-block"]))
	}
}

