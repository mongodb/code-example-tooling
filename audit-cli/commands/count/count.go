// Package count provides the parent command for counting code examples.
//
// This package serves as the parent command for various counting operations.
// Currently supports:
//   - tested-examples: Count tested code examples in the MongoDB documentation monorepo
//
// Future subcommands could include counting other types of content.
package count

import (
	"github.com/mongodb/code-example-tooling/audit-cli/commands/count/tested-examples"
	"github.com/spf13/cobra"
)

// NewCountCommand creates the count parent command.
//
// This command serves as a parent for various counting operations on code examples.
// It doesn't perform any operations itself but provides a namespace for subcommands.
func NewCountCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "count",
		Short: "Count code examples",
		Long: `Count various types of code examples in the MongoDB documentation.

Currently supports counting tested code examples in the documentation monorepo.
Future subcommands may support counting other types of content.`,
	}

	// Add subcommands
	cmd.AddCommand(tested_examples.NewTestedExamplesCommand())

	return cmd
}

