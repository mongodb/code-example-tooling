package cmd

import (
	"github.com/spf13/cobra"
)

var (
	startPath   string
	directOnly  bool
	ignorePaths []string
	quiet       bool
)

var rootCmd = &cobra.Command{
	Use:   "depman",
	Short: "Dependency Manager - Scan and update dependencies across multiple package managers",
	Long: `A CLI tool to scan directories for dependency management files 
(package.json, pom.xml, requirements.txt, go.mod, .csproj) and check or update dependencies.`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&startPath, "path", "p", ".", "Starting filepath or directory to scan")
	rootCmd.PersistentFlags().BoolVar(&directOnly, "direct-only", false, "Only check direct dependencies (excludes indirect/dev dependencies)")
	rootCmd.PersistentFlags().StringSliceVar(&ignorePaths, "ignore", []string{}, "Additional directory names to ignore (node_modules, .git, vendor, target, dist, build are always ignored)")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "Minimal output (only show updates/errors)")
}

