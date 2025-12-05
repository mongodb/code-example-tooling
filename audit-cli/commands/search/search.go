// Package search provides the parent command for searching through documentation files.
//
// This package serves as the parent command for various search operations.
// Currently supports:
//   - find-string: Search for substrings in documentation files or extracted content
//
// Future subcommands could include pattern matching, regex search, or semantic search.
package search

import (
	"github.com/mongodb/code-example-tooling/audit-cli/commands/search/find-string"
	"github.com/spf13/cobra"
)

// NewSearchCommand creates the search parent command.
//
// This command serves as a parent for various search operations on documentation files.
// It doesn't perform any operations itself but provides a namespace for subcommands.
func NewSearchCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search through documentation files",
		Long: `Search through documentation files or extracted content.

Currently supports searching for substrings in RST source files or extracted content.
Helps writers identify files that need updates and scope maintenance work.

Future subcommands may support pattern matching, regex search, or semantic search.`,
	}

	// Add subcommands
	cmd.AddCommand(find_string.NewFindStringCommand())

	return cmd
}
