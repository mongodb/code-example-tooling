package file_contents

import (
	"strings"
	"testing"
)

// TestCompareFiles tests direct file comparison
func TestCompareFiles(t *testing.T) {
	testDataDir := "../../../testdata/compare"

	tests := []struct {
		name           string
		file1          string
		file2          string
		generateDiff   bool
		expectError    bool
		expectDiff     bool
		expectMatching bool
	}{
		{
			name:           "different files without diff",
			file1:          testDataDir + "/file1.txt",
			file2:          testDataDir + "/file2.txt",
			generateDiff:   false,
			expectError:    false,
			expectDiff:     true,
			expectMatching: false,
		},
		{
			name:           "different files with diff",
			file1:          testDataDir + "/file1.txt",
			file2:          testDataDir + "/file2.txt",
			generateDiff:   true,
			expectError:    false,
			expectDiff:     true,
			expectMatching: false,
		},
		{
			name:           "identical files",
			file1:          testDataDir + "/identical1.txt",
			file2:          testDataDir + "/identical2.txt",
			generateDiff:   false,
			expectError:    false,
			expectDiff:     false,
			expectMatching: true,
		},
		{
			name:        "nonexistent file",
			file1:       testDataDir + "/file1.txt",
			file2:       testDataDir + "/nonexistent.txt",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := CompareFiles(tt.file1, tt.file2, tt.generateDiff, false)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result == nil {
				t.Fatal("expected result but got nil")
			}

			if tt.expectMatching && result.MatchingFiles != 1 {
				t.Errorf("expected 1 matching file, got %d", result.MatchingFiles)
			}

			if tt.expectDiff && result.DifferingFiles != 1 {
				t.Errorf("expected 1 differing file, got %d", result.DifferingFiles)
			}

			if tt.generateDiff && tt.expectDiff {
				if len(result.Comparisons) == 0 {
					t.Fatal("expected comparisons but got none")
				}
				if result.Comparisons[0].Diff == "" {
					t.Error("expected diff output but got empty string")
				}
			}
		})
	}
}

// TestCompareVersions tests version-based comparison
func TestCompareVersions(t *testing.T) {
	testDataDir := "../../../testdata/compare"

	tests := []struct {
		name            string
		referenceFile   string
		productDir      string
		versions        []string
		generateDiff    bool
		expectError     bool
		expectMatching  int
		expectDiffering int
		expectNotFound  int
	}{
		{
			name:            "compare across three versions",
			referenceFile:   testDataDir + "/product/manual/source/includes/example.rst",
			productDir:      testDataDir + "/product",
			versions:        []string{"manual", "upcoming", "v8.0"},
			generateDiff:    false,
			expectError:     false,
			expectMatching:  1, // manual matches itself
			expectDiffering: 2, // upcoming and v8.0 differ
			expectNotFound:  0,
		},
		{
			name:            "compare with diff generation",
			referenceFile:   testDataDir + "/product/manual/source/includes/example.rst",
			productDir:      testDataDir + "/product",
			versions:        []string{"manual", "upcoming"},
			generateDiff:    true,
			expectError:     false,
			expectMatching:  1,
			expectDiffering: 1,
			expectNotFound:  0,
		},
		{
			name:            "file not found in some versions",
			referenceFile:   testDataDir + "/product/manual/source/includes/new-feature.rst",
			productDir:      testDataDir + "/product",
			versions:        []string{"manual", "upcoming", "v8.0"},
			generateDiff:    false,
			expectError:     false,
			expectMatching:  2, // manual and upcoming match
			expectDiffering: 0,
			expectNotFound:  1, // v8.0 doesn't have this file
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := CompareVersions(tt.referenceFile, tt.productDir, tt.versions, tt.generateDiff, false)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result == nil {
				t.Fatal("expected result but got nil")
			}

			if result.MatchingFiles != tt.expectMatching {
				t.Errorf("expected %d matching files, got %d", tt.expectMatching, result.MatchingFiles)
			}

			if result.DifferingFiles != tt.expectDiffering {
				t.Errorf("expected %d differing files, got %d", tt.expectDiffering, result.DifferingFiles)
			}

			if result.NotFoundFiles != tt.expectNotFound {
				t.Errorf("expected %d not found files, got %d", tt.expectNotFound, result.NotFoundFiles)
			}

			if result.TotalFiles != len(tt.versions) {
				t.Errorf("expected %d total files, got %d", len(tt.versions), result.TotalFiles)
			}

			// Verify diff generation if requested
			if tt.generateDiff && tt.expectDiffering > 0 {
				foundDiff := false
				for _, comp := range result.Comparisons {
					if comp.Status == FileDiffers && comp.Diff != "" {
						foundDiff = true
						break
					}
				}
				if !foundDiff {
					t.Error("expected diff output but none found")
				}
			}
		})
	}
}

// TestResolveVersionPaths tests version path resolution
func TestResolveVersionPaths(t *testing.T) {
	testDataDir := "../../../testdata/compare"

	tests := []struct {
		name          string
		referenceFile string
		productDir    string
		versions      []string
		expectError   bool
		expectedPaths map[string]string // version -> expected path suffix
	}{
		{
			name:          "resolve paths for multiple versions",
			referenceFile: testDataDir + "/product/manual/source/includes/example.rst",
			productDir:    testDataDir + "/product",
			versions:      []string{"manual", "upcoming", "v8.0"},
			expectError:   false,
			expectedPaths: map[string]string{
				"manual":   "manual/source/includes/example.rst",
				"upcoming": "upcoming/source/includes/example.rst",
				"v8.0":     "v8.0/source/includes/example.rst",
			},
		},
		{
			name:          "file not under product dir",
			referenceFile: "/some/other/path/file.rst",
			productDir:    testDataDir + "/product",
			versions:      []string{"manual"},
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			paths, err := ResolveVersionPaths(tt.referenceFile, tt.productDir, tt.versions)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(paths) != len(tt.versions) {
				t.Fatalf("expected %d paths, got %d", len(tt.versions), len(paths))
			}

			for _, vp := range paths {
				expectedSuffix, ok := tt.expectedPaths[vp.Version]
				if !ok {
					t.Errorf("unexpected version: %s", vp.Version)
					continue
				}

				if !strings.HasSuffix(vp.FilePath, expectedSuffix) {
					t.Errorf("expected path to end with %s, got %s", expectedSuffix, vp.FilePath)
				}
			}
		})
	}
}

// TestExtractVersionFromPath tests version extraction from file paths
func TestExtractVersionFromPath(t *testing.T) {
	testDataDir := "../../../testdata/compare"

	tests := []struct {
		name            string
		filePath        string
		productDir      string
		expectedVersion string
		expectError     bool
	}{
		{
			name:            "extract manual version",
			filePath:        testDataDir + "/product/manual/source/includes/example.rst",
			productDir:      testDataDir + "/product",
			expectedVersion: "manual",
			expectError:     false,
		},
		{
			name:            "extract v8.0 version",
			filePath:        testDataDir + "/product/v8.0/source/includes/example.rst",
			productDir:      testDataDir + "/product",
			expectedVersion: "v8.0",
			expectError:     false,
		},
		{
			name:        "file not under product dir",
			filePath:    "/some/other/path/file.rst",
			productDir:  testDataDir + "/product",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version, err := ExtractVersionFromPath(tt.filePath, tt.productDir)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if version != tt.expectedVersion {
				t.Errorf("expected version %s, got %s", tt.expectedVersion, version)
			}
		})
	}
}

// TestGenerateDiff tests unified diff generation
func TestGenerateDiff(t *testing.T) {
	tests := []struct {
		name        string
		fromName    string
		fromContent string
		toName      string
		toContent   string
		expectEmpty bool
	}{
		{
			name:        "identical content",
			fromName:    "file1.txt",
			fromContent: "Line 1\nLine 2\n",
			toName:      "file2.txt",
			toContent:   "Line 1\nLine 2\n",
			expectEmpty: true,
		},
		{
			name:        "different content",
			fromName:    "file1.txt",
			fromContent: "Line 1\nLine 2\n",
			toName:      "file2.txt",
			toContent:   "Line 1\nLine 2 modified\n",
			expectEmpty: false,
		},
		{
			name:        "added lines",
			fromName:    "file1.txt",
			fromContent: "Line 1\n",
			toName:      "file2.txt",
			toContent:   "Line 1\nLine 2\n",
			expectEmpty: false,
		},
		{
			name:        "removed lines",
			fromName:    "file1.txt",
			fromContent: "Line 1\nLine 2\n",
			toName:      "file2.txt",
			toContent:   "Line 1\n",
			expectEmpty: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diff, err := GenerateDiff(tt.fromName, tt.fromContent, tt.toName, tt.toContent)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.expectEmpty {
				if diff != "" {
					t.Errorf("expected empty diff but got: %s", diff)
				}
			} else {
				if diff == "" {
					t.Error("expected non-empty diff but got empty string")
				}
				// Verify it's a unified diff format
				if !strings.Contains(diff, "---") || !strings.Contains(diff, "+++") {
					t.Errorf("expected unified diff format but got: %s", diff)
				}
			}
		})
	}
}

// TestAreFilesIdentical tests file identity checking
func TestAreFilesIdentical(t *testing.T) {
	tests := []struct {
		name      string
		content1  string
		content2  string
		identical bool
	}{
		{
			name:      "identical content",
			content1:  "Hello, world!\n",
			content2:  "Hello, world!\n",
			identical: true,
		},
		{
			name:      "different content",
			content1:  "Hello, world!\n",
			content2:  "Hello, Go!\n",
			identical: false,
		},
		{
			name:      "empty strings",
			content1:  "",
			content2:  "",
			identical: true,
		},
		{
			name:      "whitespace difference",
			content1:  "Hello\n",
			content2:  "Hello \n",
			identical: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AreFilesIdentical(tt.content1, tt.content2)
			if result != tt.identical {
				t.Errorf("expected %v but got %v", tt.identical, result)
			}
		})
	}
}

// TestComparisonResultMethods tests ComparisonResult helper methods
func TestComparisonResultMethods(t *testing.T) {
	t.Run("HasDifferences", func(t *testing.T) {
		result := &ComparisonResult{
			DifferingFiles: 1,
		}
		if !result.HasDifferences() {
			t.Error("expected HasDifferences to return true")
		}

		result.DifferingFiles = 0
		if result.HasDifferences() {
			t.Error("expected HasDifferences to return false")
		}
	})

	t.Run("AllMatch", func(t *testing.T) {
		result := &ComparisonResult{
			MatchingFiles:  3,
			DifferingFiles: 0,
			ErrorFiles:     0,
		}
		if !result.AllMatch() {
			t.Error("expected AllMatch to return true")
		}

		result.DifferingFiles = 1
		if result.AllMatch() {
			t.Error("expected AllMatch to return false when files differ")
		}

		result.DifferingFiles = 0
		result.ErrorFiles = 1
		if result.AllMatch() {
			t.Error("expected AllMatch to return false when errors exist")
		}

		result.ErrorFiles = 0
		result.MatchingFiles = 0
		if result.AllMatch() {
			t.Error("expected AllMatch to return false when no matching files")
		}
	})
}

// TestParseVersions tests version string parsing
func TestParseVersions(t *testing.T) {
	tests := []struct {
		name            string
		versionsStr     string
		expectedCount   int
		expectedVersion []string
	}{
		{
			name:            "single version",
			versionsStr:     "manual",
			expectedCount:   1,
			expectedVersion: []string{"manual"},
		},
		{
			name:            "multiple versions",
			versionsStr:     "manual,upcoming,v8.0",
			expectedCount:   3,
			expectedVersion: []string{"manual", "upcoming", "v8.0"},
		},
		{
			name:            "versions with spaces",
			versionsStr:     "manual, upcoming, v8.0",
			expectedCount:   3,
			expectedVersion: []string{"manual", "upcoming", "v8.0"},
		},
		{
			name:            "empty string",
			versionsStr:     "",
			expectedCount:   0,
			expectedVersion: []string{},
		},
		{
			name:            "trailing comma",
			versionsStr:     "manual,upcoming,",
			expectedCount:   2,
			expectedVersion: []string{"manual", "upcoming"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			versions := parseVersions(tt.versionsStr)

			if len(versions) != tt.expectedCount {
				t.Errorf("expected %d versions, got %d", tt.expectedCount, len(versions))
			}

			for i, expected := range tt.expectedVersion {
				if i >= len(versions) {
					t.Errorf("missing expected version: %s", expected)
					continue
				}
				if versions[i] != expected {
					t.Errorf("expected version %s at index %d, got %s", expected, i, versions[i])
				}
			}
		})
	}
}
