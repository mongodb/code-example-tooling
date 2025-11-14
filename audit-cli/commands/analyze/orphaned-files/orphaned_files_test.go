package orphanedfiles

import (
	"path/filepath"
	"testing"
)

// TestFindOrphanedFiles tests the FindOrphanedFiles function with various scenarios.
func TestFindOrphanedFiles(t *testing.T) {
	// Get the testdata directory
	testDataDir := "../../../testdata/input-files/source"

	tests := []struct {
		name                  string
		includeToctree        bool
		expectedOrphanedCount int
		shouldContain         []string // Files that should be in orphaned list
		shouldNotContain      []string // Files that should NOT be in orphaned list
	}{
		{
			name:                  "Without toctree",
			includeToctree:        false,
			expectedOrphanedCount: 9,
			shouldContain: []string{
				"index.rst",
				"include-test.rst",
				"literalinclude-test.rst",
			},
			shouldNotContain: []string{
				"includes/fact.rst", // Referenced by include-test.rst
			},
		},
		{
			name:                  "With toctree",
			includeToctree:        true,
			expectedOrphanedCount: 7,
			shouldContain: []string{
				"index.rst", // Entry point, not referenced by anything
			},
			shouldNotContain: []string{
				"include-test.rst",      // Referenced in toctree
				"literalinclude-test.rst", // Referenced in toctree
				"includes/fact.rst",     // Referenced by include-test.rst
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Get absolute path to test data
			absTestDataDir, err := filepath.Abs(testDataDir)
			if err != nil {
				t.Fatalf("failed to get absolute path: %v", err)
			}

			// Run analysis (not verbose, no exclude pattern)
			analysis, err := FindOrphanedFiles(absTestDataDir, tt.includeToctree, false, "")
			if err != nil {
				t.Fatalf("FindOrphanedFiles failed: %v", err)
			}

			// Check total orphaned count
			if analysis.TotalOrphaned != tt.expectedOrphanedCount {
				t.Errorf("expected %d orphaned files, got %d", tt.expectedOrphanedCount, analysis.TotalOrphaned)
			}

			// Create a map for easy lookup
			orphanedMap := make(map[string]bool)
			for _, file := range analysis.OrphanedFiles {
				orphanedMap[file] = true
			}

			// Check that expected files are in the orphaned list
			for _, file := range tt.shouldContain {
				if !orphanedMap[file] {
					t.Errorf("expected %s to be in orphaned list, but it wasn't", file)
				}
			}

			// Check that unexpected files are NOT in the orphaned list
			for _, file := range tt.shouldNotContain {
				if orphanedMap[file] {
					t.Errorf("expected %s to NOT be in orphaned list, but it was", file)
				}
			}

			// Verify IncludedToctree flag
			if analysis.IncludedToctree != tt.includeToctree {
				t.Errorf("expected IncludedToctree to be %v, got %v", tt.includeToctree, analysis.IncludedToctree)
			}
		})
	}
}

// TestFindOrphanedFilesInvalidDirectory tests error handling for invalid directories.
func TestFindOrphanedFilesInvalidDirectory(t *testing.T) {
	tests := []struct {
		name      string
		sourceDir string
		wantError bool
	}{
		{
			name:      "Non-existent directory",
			sourceDir: "/path/that/does/not/exist",
			wantError: true,
		},
		{
			name:      "File instead of directory",
			sourceDir: "../../../testdata/input-files/source/index.rst",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := FindOrphanedFiles(tt.sourceDir, false, false, "")
			if (err != nil) != tt.wantError {
				t.Errorf("FindOrphanedFiles() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

