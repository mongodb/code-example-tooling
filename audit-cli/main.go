// Package main provides the entry point for the audit-cli tool.
//
// audit-cli is a command-line tool for extracting and analyzing code examples
// from MongoDB documentation written in reStructuredText (RST).
//
// The CLI is organized into parent commands with subcommands:
//   - extract: Extract content from RST files
//     - code-examples: Extract code examples from RST directives
//   - search: Search through extracted content
//     - find-string: Search for substrings in extracted files
//   - analyze: Analyze RST file structures
//     - includes: Analyze include directive relationships
//   - compare: Compare files across different versions
//     - file-contents: Compare file contents across versions
package main

import (
	"github.com/mongodb/code-example-tooling/audit-cli/commands/analyze"
	"github.com/mongodb/code-example-tooling/audit-cli/commands/compare"
	"github.com/mongodb/code-example-tooling/audit-cli/commands/extract"
	"github.com/mongodb/code-example-tooling/audit-cli/commands/search"
	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "audit-cli",
		Short: "A CLI tool for extracting and analyzing code examples from MongoDB documentation",
		Long: `audit-cli extracts code examples from reStructuredText files and provides
tools for searching and analyzing the extracted content.

Supports extraction from literalinclude, code-block, and io-code-block directives,
with special handling for MongoDB documentation conventions.`,
	}

	// Add parent commands
	rootCmd.AddCommand(extract.NewExtractCommand())
	rootCmd.AddCommand(search.NewSearchCommand())
	rootCmd.AddCommand(analyze.NewAnalyzeCommand())
	rootCmd.AddCommand(compare.NewCompareCommand())

	rootCmd.Execute()
}
