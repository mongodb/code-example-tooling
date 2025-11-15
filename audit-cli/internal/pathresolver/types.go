package pathresolver

// ProjectInfo contains information about a documentation project's structure.
//
// MongoDB documentation projects can be either versioned or non-versioned:
// - Versioned: {product}/{version}/source/... (e.g., manual/v8.0/source/...)
// - Non-versioned: {product}/source/... (e.g., atlas/source/...)
type ProjectInfo struct {
	// SourceDir is the absolute path to the source directory
	SourceDir string

	// ProductDir is the absolute path to the product directory
	ProductDir string

	// Version is the version identifier (e.g., "v8.0", "manual", "upcoming")
	// Empty string for non-versioned projects
	Version string

	// IsVersioned indicates whether this is a versioned project
	IsVersioned bool
}

// VersionPath represents a resolved file path for a specific version.
//
// Used when resolving the same file across multiple versions of a product.
type VersionPath struct {
	// Version is the version identifier (e.g., "v8.0", "manual", "upcoming")
	Version string

	// FilePath is the absolute path to the file in this version
	FilePath string
}

