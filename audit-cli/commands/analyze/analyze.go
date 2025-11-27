// Package analyze provides the parent command for analyzing RST file structures.
//
// This package serves as the parent command for various analysis operations.
// Currently supports:
//   - includes: Analyze include directive relationships in RST files
//   - usage: Find all files that use a target file
//   - procedures: Analyze procedure variations and statistics
//
// Future subcommands could include analyzing cross-references, broken links, or content metrics.
package analyze

import (
	"github.com/mongodb/code-example-tooling/audit-cli/commands/analyze/includes"
	"github.com/mongodb/code-example-tooling/audit-cli/commands/analyze/procedures"
	"github.com/mongodb/code-example-tooling/audit-cli/commands/analyze/usage"
	"github.com/spf13/cobra"
)

// NewAnalyzeCommand creates the analyze parent command.
//
// This command serves as a parent for various analysis operations on RST files.
// It doesn't perform any operations itself but provides a namespace for subcommands.
func NewAnalyzeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "analyze",
		Short: "Analyze reStructuredText file structures",
		Long: `Analyze various aspects of reStructuredText files and their relationships.

Currently supports:
  - includes: Analyze include directive relationships (forward dependencies)
  - usage: Find all files that use a target file (reverse dependencies)
  - procedures: Analyze procedure variations and statistics

Future subcommands may support analyzing cross-references, broken links, or content metrics.`,
	}

	// Add subcommands
	cmd.AddCommand(includes.NewIncludesCommand())
	cmd.AddCommand(usage.NewUsageCommand())
	cmd.AddCommand(procedures.NewProceduresCommand())

	return cmd
}

