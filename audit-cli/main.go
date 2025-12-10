// Package main provides the entry point for the audit-cli tool.
//
// audit-cli is a command-line tool for performing audit-related tasks in the
// MongoDB documentation monorepo. It helps technical writers with maintenance
// tasks, scoping work, and reporting information to stakeholders.
//
// The CLI is organized into parent commands with subcommands:
//   - extract: Extract content from RST files (code examples, procedures)
//   - search: Search through documentation files
//   - analyze: Analyze RST file structures and relationships
//   - compare: Compare files across different versions
//   - count: Count documentation content (code examples, pages)
package main

import (
	"github.com/mongodb/code-example-tooling/audit-cli/commands/analyze"
	"github.com/mongodb/code-example-tooling/audit-cli/commands/compare"
	"github.com/mongodb/code-example-tooling/audit-cli/commands/count"
	"github.com/mongodb/code-example-tooling/audit-cli/commands/extract"
	"github.com/mongodb/code-example-tooling/audit-cli/commands/search"
	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "audit-cli",
		Short: "A CLI tool for auditing and analyzing MongoDB documentation",
		Long: `audit-cli helps MongoDB technical writers perform audit-related tasks in the
documentation monorepo, including:

  - Extracting content (code examples, procedures) for testing and migration
  - Searching documentation files for specific strings or patterns
  - Analyzing file dependencies and relationships
  - Comparing files across documentation versions
  - Counting documentation content for reporting and metrics

Designed for maintenance tasks, scoping work, and reporting to stakeholders.`,
	}

	// Add parent commands
	rootCmd.AddCommand(extract.NewExtractCommand())
	rootCmd.AddCommand(search.NewSearchCommand())
	rootCmd.AddCommand(analyze.NewAnalyzeCommand())
	rootCmd.AddCommand(compare.NewCompareCommand())
	rootCmd.AddCommand(count.NewCountCommand())

	err := rootCmd.Execute()
	if err != nil {
		return
	}
}
