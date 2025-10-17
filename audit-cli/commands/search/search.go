// Package search provides the parent command for searching through extracted content.
//
// This package serves as the parent command for various search operations.
// Currently supports:
//   - find-string: Search for substrings in extracted code example files
//
// Future subcommands could include pattern matching, regex search, or semantic search.
package search

import (
	"github.com/mongodb/code-example-tooling/audit-cli/commands/search/find-string"
	"github.com/spf13/cobra"
)

// NewSearchCommand creates the search parent command.
//
// This command serves as a parent for various search operations on extracted content.
// It doesn't perform any operations itself but provides a namespace for subcommands.
func NewSearchCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search through extracted content",
		Long: `Search through extracted content such as code examples.

Currently supports searching for substrings in extracted code example files.
Future subcommands may support pattern matching, regex search, or semantic search.`,
	}

	// Add subcommands
	cmd.AddCommand(find_string.NewFindStringCommand())

	return cmd
}
