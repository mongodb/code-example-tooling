// Package pages implements the pages subcommand for counting documentation pages.
package pages

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// NewPagesCommand creates the pages subcommand.
//
// This command counts documentation pages (.txt files) in the MongoDB documentation monorepo.
//
// Usage:
//   count pages /path/to/docs-monorepo
//   count pages /path/to/docs-monorepo --for-project manual
//   count pages /path/to/docs-monorepo --count-by-project
//   count pages /path/to/docs-monorepo --exclude-dirs api-reference,generated
//
// Flags:
//   - --for-project: Only count pages for a specific project
//   - --count-by-project: Display a list of projects with counts for each
//   - --exclude-dirs: Comma-separated list of directory names to exclude
func NewPagesCommand() *cobra.Command {
	var (
		forProject     string
		countByProject bool
		excludeDirs    string
		currentOnly    bool
		byVersion      bool
	)

	cmd := &cobra.Command{
		Use:   "pages [directory-path]",
		Short: "Count documentation pages in the monorepo",
		Long: `Count documentation pages (.txt files) in the MongoDB documentation monorepo.

This command navigates to the content directory and counts all .txt files recursively,
with the following exclusions:

Automatic exclusions:
  - Files in code-examples directories (at root of content or source)
  - Files in the following directories at the root of content:
    - 404
    - meta
    - table-of-contents
  - All non-.txt files

Each directory under content/ represents a different product/project.

By default, returns only a total count of all pages.

Examples:
  # Get total count of all documentation pages
  count pages /path/to/docs-monorepo

  # Count pages for a specific project
  count pages /path/to/docs-monorepo --for-project manual

  # Show counts broken down by project
  count pages /path/to/docs-monorepo --count-by-project

  # Exclude specific directories from counting
  count pages /path/to/docs-monorepo --exclude-dirs api-reference,generated

  # Combine flags: count pages for a specific project, excluding certain directories
  count pages /path/to/docs-monorepo --for-project atlas --exclude-dirs deprecated

  # Count only current versions
  count pages /path/to/docs-monorepo --current-only

  # Show counts by version
  count pages /path/to/docs-monorepo --by-version`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPages(args[0], forProject, countByProject, excludeDirs, currentOnly, byVersion)
		},
	}

	cmd.Flags().StringVar(&forProject, "for-project", "", "Only count pages for a specific project")
	cmd.Flags().BoolVar(&countByProject, "count-by-project", false, "Display counts for each project")
	cmd.Flags().StringVar(&excludeDirs, "exclude-dirs", "", "Comma-separated list of directory names to exclude")
	cmd.Flags().BoolVar(&currentOnly, "current-only", false, "Only count pages in the current version")
	cmd.Flags().BoolVar(&byVersion, "by-version", false, "Display counts grouped by project and version")

	return cmd
}

// runPages executes the pages counting operation.
func runPages(dirPath string, forProject string, countByProject bool, excludeDirs string, currentOnly bool, byVersion bool) error {
	// Validate flag combinations
	if forProject != "" && countByProject {
		return fmt.Errorf("cannot use --for-project and --count-by-project together")
	}

	if currentOnly && byVersion {
		return fmt.Errorf("cannot use --current-only and --by-version together")
	}

	// If byVersion is set, it implies countByProject
	if byVersion {
		countByProject = true
	}

	// Parse exclude-dirs flag
	var excludeDirsList []string
	if excludeDirs != "" {
		excludeDirsList = strings.Split(excludeDirs, ",")
		// Trim whitespace from each directory name
		for i, dir := range excludeDirsList {
			excludeDirsList[i] = strings.TrimSpace(dir)
		}
	}

	// Count the pages
	result, err := CountPages(dirPath, forProject, excludeDirsList, currentOnly, byVersion)
	if err != nil {
		return fmt.Errorf("failed to count pages: %w", err)
	}

	// Print the results
	PrintResults(result, countByProject, byVersion)

	return nil
}

