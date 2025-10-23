// Package extract provides the parent command for extracting content from RST files.
//
// This package serves as the parent command for various extraction operations.
// Currently supports:
//   - code-examples: Extract code examples from RST directives
//
// Future subcommands could include extracting tables, images, or other structured content.
package extract

import (
	"github.com/mongodb/code-example-tooling/audit-cli/commands/extract/code-examples"
	"github.com/spf13/cobra"
)

// NewExtractCommand creates the extract parent command.
//
// This command serves as a parent for various extraction operations on RST files.
// It doesn't perform any operations itself but provides a namespace for subcommands.
func NewExtractCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "extract",
		Short: "Extract content from reStructuredText files",
		Long: `Extract various types of content from reStructuredText files.

Currently supports extracting code examples from directives like literalinclude,
code-block, and io-code-block. Future subcommands may support extracting other
types of structured content such as tables, images, or metadata.`,
	}

	// Add subcommands
	cmd.AddCommand(code_examples.NewCodeExamplesCommand())

	return cmd
}
