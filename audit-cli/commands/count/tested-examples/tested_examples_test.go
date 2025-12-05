package tested_examples

import (
	"path/filepath"
	"testing"
)

// TestCountAllFiles tests counting all files in the tested directory
func TestCountAllFiles(t *testing.T) {
	testDataDir := filepath.Join("..", "..", "..", "testdata", "count-test-monorepo")

	result, err := CountTestedExamples(testDataDir, "", false)
	if err != nil {
		t.Fatalf("CountTestedExamples failed: %v", err)
	}

	// Total files: 
	// pymongo: 4 files (2 .py, 1 .txt, 1 .sh)
	// mongosh: 3 files (2 .js, 1 .txt)
	// go/driver: 1 file (.go)
	// go/atlas-sdk: 1 file (.go)
	// javascript/driver: 2 files (1 .js, 1 .txt)
	// java/driver-sync: 1 file (.java)
	// csharp/driver: 1 file (.cs)
	// Total: 13 files
	expectedTotal := 13
	if result.TotalCount != expectedTotal {
		t.Errorf("Expected total count %d, got %d", expectedTotal, result.TotalCount)
	}
}

// TestCountByProduct tests counting files for a specific product
func TestCountByProduct(t *testing.T) {
	testDataDir := filepath.Join("..", "..", "..", "testdata", "count-test-monorepo")

	tests := []struct {
		name          string
		product       string
		expectedCount int
	}{
		{
			name:          "count pymongo files",
			product:       "pymongo",
			expectedCount: 4, // 2 .py + 1 .txt + 1 .sh
		},
		{
			name:          "count mongosh files",
			product:       "mongosh",
			expectedCount: 3, // 2 .js + 1 .txt
		},
		{
			name:          "count go/driver files",
			product:       "go/driver",
			expectedCount: 1, // 1 .go
		},
		{
			name:          "count go/atlas-sdk files",
			product:       "go/atlas-sdk",
			expectedCount: 1, // 1 .go
		},
		{
			name:          "count javascript/driver files",
			product:       "javascript/driver",
			expectedCount: 2, // 1 .js + 1 .txt
		},
		{
			name:          "count java/driver-sync files",
			product:       "java/driver-sync",
			expectedCount: 1, // 1 .java
		},
		{
			name:          "count csharp/driver files",
			product:       "csharp/driver",
			expectedCount: 1, // 1 .cs
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := CountTestedExamples(testDataDir, tt.product, false)
			if err != nil {
				t.Fatalf("CountTestedExamples failed: %v", err)
			}

			if result.TotalCount != tt.expectedCount {
				t.Errorf("Expected count %d for product %s, got %d", tt.expectedCount, tt.product, result.TotalCount)
			}

			// Verify product counts map
			if result.ProductCounts[tt.product] != tt.expectedCount {
				t.Errorf("Expected product count %d for %s, got %d", tt.expectedCount, tt.product, result.ProductCounts[tt.product])
			}
		})
	}
}

// TestExcludeOutput tests excluding output files (.txt, .sh)
func TestExcludeOutput(t *testing.T) {
	testDataDir := filepath.Join("..", "..", "..", "testdata", "count-test-monorepo")

	result, err := CountTestedExamples(testDataDir, "", true)
	if err != nil {
		t.Fatalf("CountTestedExamples failed: %v", err)
	}

	// Total source files only:
	// pymongo: 2 .py files
	// mongosh: 2 .js files
	// go/driver: 1 .go file
	// go/atlas-sdk: 1 .go file
	// javascript/driver: 1 .js file
	// java/driver-sync: 1 .java file
	// csharp/driver: 1 .cs file
	// Total: 9 files
	expectedTotal := 9
	if result.TotalCount != expectedTotal {
		t.Errorf("Expected total count %d (excluding output), got %d", expectedTotal, result.TotalCount)
	}
}

// TestExcludeOutputForProduct tests excluding output files for a specific product
func TestExcludeOutputForProduct(t *testing.T) {
	testDataDir := filepath.Join("..", "..", "..", "testdata", "count-test-monorepo")

	tests := []struct {
		name          string
		product       string
		expectedCount int
	}{
		{
			name:          "pymongo source only",
			product:       "pymongo",
			expectedCount: 2, // 2 .py files (excluding .txt and .sh)
		},
		{
			name:          "mongosh source only",
			product:       "mongosh",
			expectedCount: 2, // 2 .js files (excluding .txt)
		},
		{
			name:          "javascript/driver source only",
			product:       "javascript/driver",
			expectedCount: 1, // 1 .js file (excluding .txt)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := CountTestedExamples(testDataDir, tt.product, true)
			if err != nil {
				t.Fatalf("CountTestedExamples failed: %v", err)
			}

			if result.TotalCount != tt.expectedCount {
				t.Errorf("Expected count %d for product %s (excluding output), got %d", tt.expectedCount, tt.product, result.TotalCount)
			}
		})
	}
}

// TestProductCounts tests the product counts map
func TestProductCounts(t *testing.T) {
	testDataDir := filepath.Join("..", "..", "..", "testdata", "count-test-monorepo")

	result, err := CountTestedExamples(testDataDir, "", false)
	if err != nil {
		t.Fatalf("CountTestedExamples failed: %v", err)
	}

	expectedCounts := map[string]int{
		"pymongo":            4,
		"mongosh":            3,
		"go/driver":          1,
		"go/atlas-sdk":       1,
		"javascript/driver":  2,
		"java/driver-sync":   1,
		"csharp/driver":      1,
	}

	for product, expectedCount := range expectedCounts {
		actualCount := result.ProductCounts[product]
		if actualCount != expectedCount {
			t.Errorf("Expected %d files for product %s, got %d", expectedCount, product, actualCount)
		}
	}

	// Verify we have exactly the expected number of products
	if len(result.ProductCounts) != len(expectedCounts) {
		t.Errorf("Expected %d products, got %d", len(expectedCounts), len(result.ProductCounts))
	}
}

// TestInvalidMonorepoPath tests error handling for invalid monorepo path
func TestInvalidMonorepoPath(t *testing.T) {
	_, err := CountTestedExamples("/nonexistent/path", "", false)
	if err == nil {
		t.Error("Expected error for nonexistent monorepo path, got nil")
	}
}

// TestMissingTestedDirectory tests error handling when tested directory doesn't exist
func TestMissingTestedDirectory(t *testing.T) {
	testDataDir := filepath.Join("..", "..", "..", "testdata")

	_, err := CountTestedExamples(testDataDir, "", false)
	if err == nil {
		t.Error("Expected error for missing tested directory, got nil")
	}
}

// TestIsValidProduct tests product validation
func TestIsValidProduct(t *testing.T) {
	tests := []struct {
		product string
		valid   bool
	}{
		{"pymongo", true},
		{"mongosh", true},
		{"go/driver", true},
		{"go/atlas-sdk", true},
		{"javascript/driver", true},
		{"java/driver-sync", true},
		{"csharp/driver", true},
		{"invalid-product", false},
		{"python/driver", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.product, func(t *testing.T) {
			result := IsValidProduct(tt.product)
			if result != tt.valid {
				t.Errorf("IsValidProduct(%s) = %v, expected %v", tt.product, result, tt.valid)
			}
		})
	}
}

// TestExtractProductKey tests product key extraction from paths
func TestExtractProductKey(t *testing.T) {
	tests := []struct {
		name        string
		relPath     string
		expectedKey string
	}{
		{
			name:        "pymongo path",
			relPath:     "python/pymongo/example.py",
			expectedKey: "pymongo",
		},
		{
			name:        "mongosh path",
			relPath:     "command-line/mongosh/example.js",
			expectedKey: "mongosh",
		},
		{
			name:        "go driver path",
			relPath:     "go/driver/example.go",
			expectedKey: "go/driver",
		},
		{
			name:        "go atlas-sdk path",
			relPath:     "go/atlas-sdk/example.go",
			expectedKey: "go/atlas-sdk",
		},
		{
			name:        "javascript driver path",
			relPath:     "javascript/driver/example.js",
			expectedKey: "javascript/driver",
		},
		{
			name:        "java driver-sync path",
			relPath:     "java/driver-sync/Example.java",
			expectedKey: "java/driver-sync",
		},
		{
			name:        "csharp driver path",
			relPath:     "csharp/driver/Example.cs",
			expectedKey: "csharp/driver",
		},
		{
			name:        "invalid path - too short",
			relPath:     "python",
			expectedKey: "",
		},
		{
			name:        "invalid path - unknown product",
			relPath:     "python/unknown/example.py",
			expectedKey: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractProductKey(tt.relPath)
			if result != tt.expectedKey {
				t.Errorf("extractProductKey(%s) = %s, expected %s", tt.relPath, result, tt.expectedKey)
			}
		})
	}
}

// TestIsOutputFile tests output file detection
func TestIsOutputFile(t *testing.T) {
	tests := []struct {
		ext      string
		isOutput bool
	}{
		{".txt", true},
		{".sh", true},
		{".py", false},
		{".js", false},
		{".go", false},
		{".java", false},
		{".cs", false},
		{".md", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.ext, func(t *testing.T) {
			result := isOutputFile(tt.ext)
			if result != tt.isOutput {
				t.Errorf("isOutputFile(%s) = %v, expected %v", tt.ext, result, tt.isOutput)
			}
		})
	}
}

// TestIsSourceFileForProduct tests source file detection for products
func TestIsSourceFileForProduct(t *testing.T) {
	tests := []struct {
		name     string
		ext      string
		product  string
		isSource bool
	}{
		{"python source for pymongo", ".py", "pymongo", true},
		{"txt not source for pymongo", ".txt", "pymongo", false},
		{"js source for mongosh", ".js", "mongosh", true},
		{"js source for javascript/driver", ".js", "javascript/driver", true},
		{"go source for go/driver", ".go", "go/driver", true},
		{"go source for go/atlas-sdk", ".go", "go/atlas-sdk", true},
		{"java source for java/driver-sync", ".java", "java/driver-sync", true},
		{"cs source for csharp/driver", ".cs", "csharp/driver", true},
		{"py not source for go/driver", ".py", "go/driver", false},
		{"invalid product", ".py", "invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isSourceFileForProduct(tt.ext, tt.product)
			if result != tt.isSource {
				t.Errorf("isSourceFileForProduct(%s, %s) = %v, expected %v", tt.ext, tt.product, result, tt.isSource)
			}
		})
	}
}

