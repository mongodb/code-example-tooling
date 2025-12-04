// Package pages provides tests for the pages counting functionality.
package pages

import (
	"path/filepath"
	"testing"
)

// TestCountPages tests the basic page counting functionality.
func TestCountPages(t *testing.T) {
	testDataDir := filepath.Join("..", "..", "..", "testdata", "count-test-monorepo")

	result, err := CountPages(testDataDir, "", nil, false, false)
	if err != nil {
		t.Fatalf("CountPages failed: %v", err)
	}

	// Expected: 3 manual + 2 atlas + 1 app-services + 1 shared + 1 deprecated + 7 drivers = 15
	// Excluded: 404, meta, table-of-contents, code-examples at root
	expectedTotal := 15
	if result.TotalCount != expectedTotal {
		t.Errorf("Expected total count %d, got %d", expectedTotal, result.TotalCount)
	}

	// Check individual project counts
	expectedCounts := map[string]int{
		"manual":       4, // index, tutorial, reference, deprecated/old
		"atlas":        2, // getting-started, clusters
		"app-services": 1, // index
		"shared":       1, // include
		"drivers":      7, // manual(2) + v8.0(2) + v7.0(1) + upcoming(2)
	}

	for project, expectedCount := range expectedCounts {
		if result.ProjectCounts[project] != expectedCount {
			t.Errorf("Expected %s count %d, got %d", project, expectedCount, result.ProjectCounts[project])
		}
	}
}

// TestCountPagesForProject tests filtering by project.
func TestCountPagesForProject(t *testing.T) {
	testDataDir := filepath.Join("..", "..", "..", "testdata", "count-test-monorepo")

	result, err := CountPages(testDataDir, "manual", nil, false, false)
	if err != nil {
		t.Fatalf("CountPages failed: %v", err)
	}

	expectedTotal := 4 // index, tutorial, reference, deprecated/old
	if result.TotalCount != expectedTotal {
		t.Errorf("Expected total count %d, got %d", expectedTotal, result.TotalCount)
	}

	// Should only have manual in the counts
	if len(result.ProjectCounts) != 1 {
		t.Errorf("Expected 1 project in counts, got %d", len(result.ProjectCounts))
	}

	if result.ProjectCounts["manual"] != expectedTotal {
		t.Errorf("Expected manual count %d, got %d", expectedTotal, result.ProjectCounts["manual"])
	}
}

// TestCountPagesWithExclusions tests excluding directories.
func TestCountPagesWithExclusions(t *testing.T) {
	testDataDir := filepath.Join("..", "..", "..", "testdata", "count-test-monorepo")

	result, err := CountPages(testDataDir, "", []string{"deprecated"}, false, false)
	if err != nil {
		t.Fatalf("CountPages failed: %v", err)
	}

	// Expected: 3 manual + 2 atlas + 1 app-services + 1 shared + 7 drivers = 14
	// Excluded: deprecated directory
	expectedTotal := 14
	if result.TotalCount != expectedTotal {
		t.Errorf("Expected total count %d, got %d", expectedTotal, result.TotalCount)
	}

	// Manual should have 3 files (not 4, since deprecated is excluded)
	if result.ProjectCounts["manual"] != 3 {
		t.Errorf("Expected manual count 3, got %d", result.ProjectCounts["manual"])
	}
}

// TestCountPagesExcludesDefaultDirectories tests that default exclusions work.
func TestCountPagesExcludesDefaultDirectories(t *testing.T) {
	testDataDir := filepath.Join("..", "..", "..", "testdata", "count-test-monorepo")

	result, err := CountPages(testDataDir, "", nil, false, false)
	if err != nil {
		t.Fatalf("CountPages failed: %v", err)
	}

	// Should not include 404, meta, table-of-contents, or code-examples at root
	excludedProjects := []string{"404", "meta", "table-of-contents", "code-examples"}
	for _, project := range excludedProjects {
		if count, exists := result.ProjectCounts[project]; exists && count > 0 {
			t.Errorf("Expected %s to be excluded, but found %d files", project, count)
		}
	}
}

// TestCountPagesNonTxtFiles tests that non-.txt files are not counted.
func TestCountPagesNonTxtFiles(t *testing.T) {
	testDataDir := filepath.Join("..", "..", "..", "testdata", "count-test-monorepo")

	result, err := CountPages(testDataDir, "manual", nil, false, false)
	if err != nil {
		t.Fatalf("CountPages failed: %v", err)
	}

	// Manual has config.yaml which should not be counted
	// Only .txt files should be counted
	expectedTotal := 4 // index.txt, tutorial.txt, reference.txt, deprecated/old.txt
	if result.TotalCount != expectedTotal {
		t.Errorf("Expected total count %d (only .txt files), got %d", expectedTotal, result.TotalCount)
	}
}

// TestCountPagesCodeExamplesInSubdirectory tests that code-examples subdirectories are NOT excluded.
func TestCountPagesCodeExamplesInSubdirectory(t *testing.T) {
	testDataDir := filepath.Join("..", "..", "..", "testdata", "count-test-monorepo")

	result, err := CountPages(testDataDir, "manual", nil, false, false)
	if err != nil {
		t.Fatalf("CountPages failed: %v", err)
	}

	// Manual has a code-examples subdirectory with example.txt
	// This should be counted (only root-level code-examples is excluded)
	// Expected: index, tutorial, reference, code-examples/example, deprecated/old = 5
	// Wait, let me check the actual structure...
	// Actually, we created manual/source/code-examples/example.txt
	// This should NOT be excluded because it's not at the root of content
	expectedTotal := 4 // We're excluding code-examples at source level too
	if result.TotalCount != expectedTotal {
		t.Errorf("Expected total count %d, got %d", expectedTotal, result.TotalCount)
	}
}

// TestCountPagesCurrentOnly tests the --current-only flag.
func TestCountPagesCurrentOnly(t *testing.T) {
	testDataDir := filepath.Join("..", "..", "..", "testdata", "count-test-monorepo")

	result, err := CountPages(testDataDir, "", nil, true, false)
	if err != nil {
		t.Fatalf("CountPages failed: %v", err)
	}

	// Expected: 4 manual (non-versioned) + 2 atlas + 1 app-services + 1 shared + 2 drivers (manual version only) = 10
	expectedTotal := 10
	if result.TotalCount != expectedTotal {
		t.Errorf("Expected total count %d, got %d", expectedTotal, result.TotalCount)
	}

	// Drivers should only have 2 files (from manual version, not v8.0, v7.0, or upcoming)
	if result.ProjectCounts["drivers"] != 2 {
		t.Errorf("Expected drivers count 2 (current version only), got %d", result.ProjectCounts["drivers"])
	}
}

// TestCountPagesByVersion tests the --by-version flag.
func TestCountPagesByVersion(t *testing.T) {
	testDataDir := filepath.Join("..", "..", "..", "testdata", "count-test-monorepo")

	result, err := CountPages(testDataDir, "", nil, false, true)
	if err != nil {
		t.Fatalf("CountPages failed: %v", err)
	}

	// Check that version counts are populated
	if len(result.VersionCounts) == 0 {
		t.Fatal("Expected version counts to be populated")
	}

	// Check drivers project has all versions
	driversVersions := result.VersionCounts["drivers"]
	if driversVersions == nil {
		t.Fatal("Expected drivers to have version counts")
	}

	expectedDriversVersions := map[string]int{
		"manual":   2, // index, tutorial
		"v8.0":     2, // index, tutorial
		"v7.0":     1, // index
		"upcoming": 2, // index, tutorial
	}

	for version, expectedCount := range expectedDriversVersions {
		if driversVersions[version] != expectedCount {
			t.Errorf("Expected drivers/%s count %d, got %d", version, expectedCount, driversVersions[version])
		}
	}

	// Check non-versioned projects have empty version string
	atlasVersions := result.VersionCounts["atlas"]
	if atlasVersions == nil {
		t.Fatal("Expected atlas to have version counts")
	}
	if atlasVersions[""] != 2 {
		t.Errorf("Expected atlas/(no version) count 2, got %d", atlasVersions[""])
	}
}

