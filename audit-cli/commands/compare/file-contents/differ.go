package file_contents

import (
	"github.com/aymanbagabas/go-udiff"
)

// GenerateDiff generates a unified diff between two file contents.
//
// This function uses the Myers diff algorithm to compute the differences
// between two strings and formats the output as a unified diff.
//
// Parameters:
//   - fromName: Name/label for the "from" file (e.g., "manual/source/file.rst")
//   - fromContent: Content of the "from" file
//   - toName: Name/label for the "to" file (e.g., "v8.0/source/file.rst")
//   - toContent: Content of the "to" file
//
// Returns:
//   - string: The unified diff output, or empty string if files are identical
//   - error: Any error encountered during diff generation
func GenerateDiff(fromName, fromContent, toName, toContent string) (string, error) {
	// If contents are identical, return empty string
	if fromContent == toContent {
		return "", nil
	}

	// Generate unified diff using go-udiff
	// This uses the default number of context lines (3)
	diff := udiff.Unified(fromName, toName, fromContent, toContent)

	return diff, nil
}

// GenerateDiffWithContext generates a unified diff with custom context lines.
//
// This function is similar to GenerateDiff but allows specifying the number
// of context lines to include around changes.
//
// Parameters:
//   - fromName: Name/label for the "from" file
//   - fromContent: Content of the "from" file
//   - toName: Name/label for the "to" file
//   - toContent: Content of the "to" file
//   - contextLines: Number of context lines to show around changes (typically 3)
//
// Returns:
//   - string: The unified diff output, or empty string if files are identical
//   - error: Any error encountered during diff generation
func GenerateDiffWithContext(fromName, fromContent, toName, toContent string, contextLines int) (string, error) {
	// If contents are identical, return empty string
	if fromContent == toContent {
		return "", nil
	}

	// Compute edits
	edits := udiff.Strings(fromContent, toContent)

	// Generate unified diff with custom context lines
	// ToUnified returns a string directly
	diff, err := udiff.ToUnified(fromName, toName, fromContent, edits, contextLines)
	if err != nil {
		return "", err
	}

	return diff, nil
}

// AreFilesIdentical checks if two file contents are identical.
//
// This is a simple byte-by-byte comparison.
//
// Parameters:
//   - content1: First file content
//   - content2: Second file content
//
// Returns:
//   - bool: true if contents are identical, false otherwise
func AreFilesIdentical(content1, content2 string) bool {
	return content1 == content2
}

