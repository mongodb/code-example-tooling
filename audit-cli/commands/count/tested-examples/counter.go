// Package tested_examples provides counting functionality for tested code examples.
package tested_examples

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// CountTestedExamples counts tested code examples in the monorepo.
//
// This function navigates to content/code-examples/tested from the monorepo root
// and counts files based on the specified filters.
//
// Parameters:
//   - monorepoPath: Path to the documentation monorepo root
//   - forProduct: If non-empty, only count files for this product
//   - excludeOutput: If true, exclude output files (.txt, .sh)
//
// Returns:
//   - *CountResult: The counting results
//   - error: Any error encountered during counting
func CountTestedExamples(monorepoPath string, forProduct string, excludeOutput bool) (*CountResult, error) {
	// Get absolute path to monorepo
	absMonorepoPath, err := filepath.Abs(monorepoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Check if monorepo path exists
	if _, err := os.Stat(absMonorepoPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("monorepo path does not exist: %s", absMonorepoPath)
	}

	// Navigate to tested directory
	testedDir := filepath.Join(absMonorepoPath, "content", "code-examples", "tested")
	if _, err := os.Stat(testedDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("tested directory does not exist: %s\n\nPlease ensure you provided the path to the monorepo root", testedDir)
	}

	result := &CountResult{
		TotalCount:    0,
		ProductCounts: make(map[string]int),
		TestedDir:     testedDir,
	}

	// Walk through the tested directory
	err = filepath.Walk(testedDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Get the file extension
		ext := filepath.Ext(path)

		// If excluding output files, skip .txt and .sh files
		if excludeOutput && isOutputFile(ext) {
			return nil
		}

		// Get relative path from tested directory
		relPath, err := filepath.Rel(testedDir, path)
		if err != nil {
			return err
		}

		// Extract product key from path
		// Path structure: <language>/<product>/<files...>
		productKey := extractProductKey(relPath)
		if productKey == "" {
			// Skip files not in a product directory
			return nil
		}

		// If filtering by product, check if this file matches
		if forProduct != "" && productKey != forProduct {
			return nil
		}

		// If excluding output, verify this is a source file for the product
		if excludeOutput && !isSourceFileForProduct(ext, productKey) {
			return nil
		}

		// Count this file
		result.TotalCount++
		result.ProductCounts[productKey]++

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk tested directory: %w", err)
	}

	return result, nil
}

// extractProductKey extracts the product key from a relative path.
//
// Path structure: <language>/<product>/<files...>
// Returns: "<language>/<product>" or just "<product>" for special cases
func extractProductKey(relPath string) string {
	parts := strings.Split(relPath, string(filepath.Separator))
	if len(parts) < 2 {
		return ""
	}

	language := parts[0]
	product := parts[1]

	// Special case for mongosh (command-line/mongosh)
	if language == "command-line" && product == "mongosh" {
		return "mongosh"
	}

	// Special case for pymongo (python/pymongo)
	if language == "python" && product == "pymongo" {
		return "pymongo"
	}

	// For all other products, use language/product format
	productKey := language + "/" + product

	// Verify this is a known product
	if IsValidProduct(productKey) {
		return productKey
	}

	return ""
}

// isOutputFile checks if a file extension represents an output file.
func isOutputFile(ext string) bool {
	for _, outputExt := range OutputExtensions {
		if ext == outputExt {
			return true
		}
	}
	return false
}

// isSourceFileForProduct checks if a file extension is a source file for the given product.
func isSourceFileForProduct(ext string, productKey string) bool {
	productInfo, exists := ProductMap[productKey]
	if !exists {
		return false
	}

	for _, sourceExt := range productInfo.SourceExtensions {
		if ext == sourceExt {
			return true
		}
	}
	return false
}

