package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"dependency-manager/internal/checker"
	"dependency-manager/internal/scanner"

	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check for available dependency updates (dry run)",
	Long: `Scans for dependency management files and checks for available updates
without making any changes to the files or installing dependencies.`,
	RunE: runCheck,
}

func init() {
	rootCmd.AddCommand(checkCmd)
}

func runCheck(cmd *cobra.Command, args []string) error {
	// Initialize scanner
	var s *scanner.Scanner
	if len(ignorePaths) > 0 {
		s = scanner.NewWithIgnorePaths(startPath, ignorePaths)
	} else {
		s = scanner.New(startPath)
	}

	// Scan for dependency files
	depFiles, err := s.Scan()
	if err != nil {
		return fmt.Errorf("failed to scan for dependency files: %w", err)
	}

	if len(depFiles) == 0 {
		if !quiet {
			fmt.Println("No dependency management files found.")
		}
		return nil
	}

	if !quiet {
		fmt.Printf("Found %d dependency management file(s):\n\n", len(depFiles))
	}

	// Initialize checker registry
	registry := initializeRegistry()

	// Check each file
	for _, depFile := range depFiles {
		if !quiet {
			if directOnly {
				fmt.Printf("Checking %s (%s) [direct dependencies only]...\n", depFile.Path, depFile.Type)
			} else {
				fmt.Printf("Checking %s (%s)...\n", depFile.Path, depFile.Type)
			}
		}

		result := registry.CheckFile(depFile, directOnly)
		if result.Error != nil {
			fmt.Fprintf(os.Stderr, "  Error: %v\n\n", result.Error)
			continue
		}

		if len(result.Updates) == 0 {
			if !quiet {
				fmt.Println("  All dependencies are up to date!")
			}
			continue
		}

		// In quiet mode, only show files with updates
		if quiet {
			fmt.Printf("%s:\n", depFile.Path)
		} else {
			fmt.Printf("  Found %d update(s):\n", len(result.Updates))
		}
		printUpdates(result.Updates)
		fmt.Println()
	}

	return nil
}

func printUpdates(updates []checker.DependencyUpdate) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "  Package\tCurrent\tLatest\tType")
	fmt.Fprintln(w, "  -------\t-------\t------\t----")

	for _, update := range updates {
		fmt.Fprintf(w, "  %s\t%s\t%s\t%s\n",
			update.Name,
			update.CurrentVersion,
			update.LatestVersion,
			update.UpdateType,
		)
	}

	w.Flush()
}

func initializeRegistry() *checker.Registry {
	registry := checker.NewRegistry()
	registry.Register(checker.NewNpmChecker())
	registry.Register(checker.NewMavenChecker())
	registry.Register(checker.NewPipChecker())
	registry.Register(checker.NewGoModChecker())
	registry.Register(checker.NewNuGetChecker())
	return registry
}

