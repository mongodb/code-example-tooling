// Package count provides the parent command for counting code examples and documentation pages.
//
// This package serves as the parent command for various counting operations.
// Currently supports:
//   - tested-examples: Count tested code examples in the MongoDB documentation monorepo
//   - pages: Count documentation pages (.txt files) in the MongoDB documentation monorepo
package count

import (
	"github.com/mongodb/code-example-tooling/audit-cli/commands/count/pages"
	"github.com/mongodb/code-example-tooling/audit-cli/commands/count/tested-examples"
	"github.com/spf13/cobra"
)

// NewCountCommand creates the count parent command.
//
// This command serves as a parent for various counting operations on code examples and documentation pages.
// It doesn't perform any operations itself but provides a namespace for subcommands.
func NewCountCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "count",
		Short: "Count code examples and documentation pages",
		Long: `Count various types of content in the MongoDB documentation.

Currently supports:
  - tested-examples: Count tested code examples in the documentation monorepo
  - pages: Count documentation pages (.txt files) in the documentation monorepo`,
	}

	// Add subcommands
	cmd.AddCommand(tested_examples.NewTestedExamplesCommand())
	cmd.AddCommand(pages.NewPagesCommand())

	return cmd
}

