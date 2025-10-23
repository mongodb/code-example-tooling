package find_string

import (
	"path/filepath"
	"testing"
)

// TestDefaultBehaviorCaseInsensitive tests that search is case-insensitive by default
func TestDefaultBehaviorCaseInsensitive(t *testing.T) {
	testDataDir := filepath.Join("..", "..", "..", "testdata", "search-test-files")
	mixedCaseFile := filepath.Join(testDataDir, "mixed-case.txt")

	// Search for lowercase "curl" with default settings (case-insensitive)
	report, err := RunSearch(mixedCaseFile, "curl", false, false, false, false, false)
	if err != nil {
		t.Fatalf("RunSearch failed: %v", err)
	}

	// Should match because it's case-insensitive
	if report.FilesContaining != 1 {
		t.Errorf("Expected 1 file containing 'curl' (case-insensitive), got %d", report.FilesContaining)
	}
}

// TestCaseSensitiveFlag tests that --case-sensitive flag works correctly
func TestCaseSensitiveFlag(t *testing.T) {
	testDataDir := filepath.Join("..", "..", "..", "testdata", "search-test-files")
	mixedCaseFile := filepath.Join(testDataDir, "mixed-case.txt")

	// Search for uppercase "CURL" with case-sensitive flag
	report, err := RunSearch(mixedCaseFile, "CURL", false, false, false, true, false)
	if err != nil {
		t.Fatalf("RunSearch failed: %v", err)
	}

	// Should match only the uppercase version
	if report.FilesContaining != 1 {
		t.Errorf("Expected 1 file containing 'CURL' (case-sensitive), got %d", report.FilesContaining)
	}

	// Search for lowercase "curl" with case-sensitive flag
	report2, err := RunSearch(mixedCaseFile, "curl", false, false, false, true, false)
	if err != nil {
		t.Fatalf("RunSearch failed: %v", err)
	}

	// Should match only the lowercase version
	if report2.FilesContaining != 1 {
		t.Errorf("Expected 1 file containing 'curl' (case-sensitive), got %d", report2.FilesContaining)
	}
}

// TestDefaultBehaviorExactWordMatch tests that exact word matching is the default
func TestDefaultBehaviorExactWordMatch(t *testing.T) {
	testDataDir := filepath.Join("..", "..", "..", "testdata", "search-test-files")
	
	// Search for "curl" in a file that only has "curl" as a standalone word
	curlFile := filepath.Join(testDataDir, "curl-examples.txt")
	report1, err := RunSearch(curlFile, "curl", false, false, false, false, false)
	if err != nil {
		t.Fatalf("RunSearch failed: %v", err)
	}
	if report1.FilesContaining != 1 {
		t.Errorf("Expected 1 file containing 'curl' as exact word, got %d", report1.FilesContaining)
	}

	// Search for "curl" in a file that only has "libcurl" (should NOT match with exact word matching)
	libcurlFile := filepath.Join(testDataDir, "libcurl-examples.txt")
	report2, err := RunSearch(libcurlFile, "curl", false, false, false, false, false)
	if err != nil {
		t.Fatalf("RunSearch failed: %v", err)
	}
	if report2.FilesContaining != 0 {
		t.Errorf("Expected 0 files containing 'curl' as exact word in libcurl file, got %d", report2.FilesContaining)
	}
}

// TestPartialMatchFlag tests that --partial-match flag allows substring matching
func TestPartialMatchFlag(t *testing.T) {
	testDataDir := filepath.Join("..", "..", "..", "testdata", "search-test-files")
	libcurlFile := filepath.Join(testDataDir, "libcurl-examples.txt")

	// Search for "curl" with partial match enabled (should match "libcurl")
	report, err := RunSearch(libcurlFile, "curl", false, false, false, false, true)
	if err != nil {
		t.Fatalf("RunSearch failed: %v", err)
	}

	if report.FilesContaining != 1 {
		t.Errorf("Expected 1 file containing 'curl' with partial match, got %d", report.FilesContaining)
	}
}

// TestWordBoundaries tests various word boundary scenarios
func TestWordBoundaries(t *testing.T) {
	testDataDir := filepath.Join("..", "..", "..", "testdata", "search-test-files")
	boundariesFile := filepath.Join(testDataDir, "word-boundaries.txt")

	// Test exact word match (should match "curl" but not "libcurl", "curlopt", etc.)
	report, err := RunSearch(boundariesFile, "curl", false, false, false, false, false)
	if err != nil {
		t.Fatalf("RunSearch failed: %v", err)
	}

	// The file contains "curl" as a standalone word, so should match
	if report.FilesContaining != 1 {
		t.Errorf("Expected 1 file containing 'curl' as exact word, got %d", report.FilesContaining)
	}

	// Test partial match (should match all occurrences)
	report2, err := RunSearch(boundariesFile, "curl", false, false, false, false, true)
	if err != nil {
		t.Fatalf("RunSearch failed: %v", err)
	}

	// Should match because partial matching is enabled
	if report2.FilesContaining != 1 {
		t.Errorf("Expected 1 file containing 'curl' with partial match, got %d", report2.FilesContaining)
	}
}

// TestDirectorySearch tests searching across multiple files in a directory
func TestDirectorySearch(t *testing.T) {
	testDataDir := filepath.Join("..", "..", "..", "testdata", "search-test-files")

	// Search for "curl" in the directory (exact word match, case-insensitive)
	report, err := RunSearch(testDataDir, "curl", false, false, false, false, false)
	if err != nil {
		t.Fatalf("RunSearch failed: %v", err)
	}

	// Should find "curl" in:
	// - curl-examples.txt (has "curl" as standalone word)
	// - mixed-case.txt (has "curl", "CURL", "Curl" - case insensitive)
	// - word-boundaries.txt (has "curl" as standalone word)
	// - python-code.py (has "curl" as standalone word)
	// Should NOT find in:
	// - libcurl-examples.txt (only has "libcurl", not standalone "curl")
	// - no-match.txt (doesn't contain "curl" at all)
	expectedMatches := 4
	if report.FilesContaining != expectedMatches {
		t.Errorf("Expected %d files containing 'curl', got %d", expectedMatches, report.FilesContaining)
	}

	// Verify total files scanned
	if report.FilesScanned != 6 {
		t.Errorf("Expected 6 files scanned, got %d", report.FilesScanned)
	}
}

// TestDirectorySearchWithPartialMatch tests directory search with partial matching
func TestDirectorySearchWithPartialMatch(t *testing.T) {
	testDataDir := filepath.Join("..", "..", "..", "testdata", "search-test-files")

	// Search for "curl" with partial match enabled
	report, err := RunSearch(testDataDir, "curl", false, false, false, false, true)
	if err != nil {
		t.Fatalf("RunSearch failed: %v", err)
	}

	// Should find "curl" in all files except no-match.txt:
	// - curl-examples.txt
	// - libcurl-examples.txt (now matches because of partial match)
	// - mixed-case.txt
	// - word-boundaries.txt
	// - python-code.py
	expectedMatches := 5
	if report.FilesContaining != expectedMatches {
		t.Errorf("Expected %d files containing 'curl' with partial match, got %d", expectedMatches, report.FilesContaining)
	}
}

// TestCombinedFlags tests using both case-sensitive and partial-match flags together
func TestCombinedFlags(t *testing.T) {
	testDataDir := filepath.Join("..", "..", "..", "testdata", "search-test-files")
	mixedCaseFile := filepath.Join(testDataDir, "mixed-case.txt")

	// Search for lowercase "curl" with both case-sensitive and partial match
	report, err := RunSearch(mixedCaseFile, "curl", false, false, false, true, true)
	if err != nil {
		t.Fatalf("RunSearch failed: %v", err)
	}

	// Should match only lowercase "curl"
	if report.FilesContaining != 1 {
		t.Errorf("Expected 1 file containing 'curl' (case-sensitive + partial), got %d", report.FilesContaining)
	}

	// Search for uppercase "CURL" with both flags
	report2, err := RunSearch(mixedCaseFile, "CURL", false, false, false, true, true)
	if err != nil {
		t.Fatalf("RunSearch failed: %v", err)
	}

	// Should match only uppercase "CURL"
	if report2.FilesContaining != 1 {
		t.Errorf("Expected 1 file containing 'CURL' (case-sensitive + partial), got %d", report2.FilesContaining)
	}
}

// TestLanguageDetection tests that language is correctly detected from file extensions
func TestLanguageDetection(t *testing.T) {
	testDataDir := filepath.Join("..", "..", "..", "testdata", "search-test-files")

	// Search in directory and check language counts
	report, err := RunSearch(testDataDir, "curl", false, false, false, false, false)
	if err != nil {
		t.Fatalf("RunSearch failed: %v", err)
	}

	// Should have detected .txt and .py files
	if _, hasTxt := report.LanguageCounts["txt"]; !hasTxt {
		t.Error("Expected to find 'txt' in language counts")
	}

	if _, hasPy := report.LanguageCounts["py"]; !hasPy {
		t.Error("Expected to find 'py' in language counts")
	}

	// Check that txt count is correct (3 txt files should match)
	if report.LanguageCounts["txt"] != 3 {
		t.Errorf("Expected 3 txt files, got %d", report.LanguageCounts["txt"])
	}

	// Check that py count is correct (1 py file should match)
	if report.LanguageCounts["py"] != 1 {
		t.Errorf("Expected 1 py file, got %d", report.LanguageCounts["py"])
	}
}

// TestNoMatches tests searching for a string that doesn't exist
func TestNoMatches(t *testing.T) {
	testDataDir := filepath.Join("..", "..", "..", "testdata", "search-test-files")
	noMatchFile := filepath.Join(testDataDir, "no-match.txt")

	// Search for "curl" in a file that doesn't contain it
	report, err := RunSearch(noMatchFile, "curl", false, false, false, false, false)
	if err != nil {
		t.Fatalf("RunSearch failed: %v", err)
	}

	if report.FilesContaining != 0 {
		t.Errorf("Expected 0 files containing 'curl', got %d", report.FilesContaining)
	}

	if report.FilesScanned != 1 {
		t.Errorf("Expected 1 file scanned, got %d", report.FilesScanned)
	}
}

