// Package tested_examples provides functionality for counting tested code examples.
package tested_examples

// ProductInfo represents information about a MongoDB product.
type ProductInfo struct {
	// Key is the product identifier used in commands (e.g., "pymongo", "go/driver")
	Key string
	// Name is the human-readable product name
	Name string
	// SourceExtensions are the file extensions for source code files
	SourceExtensions []string
}

// CountResult represents the result of counting files.
type CountResult struct {
	// TotalCount is the total number of files counted
	TotalCount int
	// ProductCounts maps product keys to their file counts
	ProductCounts map[string]int
	// TestedDir is the path to the tested directory
	TestedDir string
}

// ProductMap defines the mapping of product keys to product information.
var ProductMap = map[string]ProductInfo{
	"mongosh": {
		Key:              "mongosh",
		Name:             "MongoDB Shell",
		SourceExtensions: []string{".js"},
	},
	"csharp/driver": {
		Key:              "csharp/driver",
		Name:             "C#/.NET Driver",
		SourceExtensions: []string{".cs"},
	},
	"go/driver": {
		Key:              "go/driver",
		Name:             "Go Driver",
		SourceExtensions: []string{".go"},
	},
	"go/atlas-sdk": {
		Key:              "go/atlas-sdk",
		Name:             "Atlas Go SDK",
		SourceExtensions: []string{".go"},
	},
	"java/driver-sync": {
		Key:              "java/driver-sync",
		Name:             "Java Sync Driver",
		SourceExtensions: []string{".java"},
	},
	"javascript/driver": {
		Key:              "javascript/driver",
		Name:             "Node.js Driver",
		SourceExtensions: []string{".js"},
	},
	"pymongo": {
		Key:              "pymongo",
		Name:             "PyMongo Driver",
		SourceExtensions: []string{".py"},
	},
}

// OutputExtensions are file extensions that represent output files (not source code).
var OutputExtensions = []string{".txt", ".sh"}

// IsValidProduct checks if a product key is valid.
func IsValidProduct(productKey string) bool {
	_, exists := ProductMap[productKey]
	return exists
}

// GetProductList returns a formatted list of valid products for help text.
func GetProductList() string {
	return `Valid products:
  - mongosh              (MongoDB Shell)
  - csharp/driver        (C#/.NET Driver)
  - go/driver            (Go Driver)
  - go/atlas-sdk         (Atlas Go SDK)
  - java/driver-sync     (Java Sync Driver)
  - javascript/driver    (Node.js Driver)
  - pymongo              (PyMongo Driver)`
}

