// Package compare provides the parent command for comparing files across versions.
//
// This package serves as the parent command for various comparison operations.
// Currently supports:
//   - file-contents: Compare file contents across different versions
//
// Future subcommands could include comparing metadata, structure, or other aspects.
package compare

import (
	"github.com/mongodb/code-example-tooling/audit-cli/commands/compare/file-contents"
	"github.com/spf13/cobra"
)

// NewCompareCommand creates the compare parent command.
//
// This command serves as a parent for various comparison operations on documentation files.
// It doesn't perform any operations itself but provides a namespace for subcommands.
func NewCompareCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "compare",
		Short: "Compare files across different versions",
		Long: `Compare files across different versions of MongoDB documentation.

Currently supports comparing file contents to identify differences between
the same file across multiple documentation versions. This helps writers
understand how content has diverged across versions and identify maintenance work.`,
	}

	// Add subcommands
	cmd.AddCommand(file_contents.NewFileContentsCommand())

	return cmd
}

