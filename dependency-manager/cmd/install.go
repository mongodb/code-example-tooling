package cmd

import (
	"fmt"

	"dependency-manager/internal/checker"
	"dependency-manager/internal/scanner"

	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Update and install dependencies",
	Long: `Performs a full update: updates dependency management files to the latest versions
and then installs/syncs the new dependencies.`,
	RunE: runInstall,
}

func init() {
	rootCmd.AddCommand(installCmd)
}

func runInstall(cmd *cobra.Command, args []string) error {
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

	// Update and install for each file
	for _, depFile := range depFiles {
		if !quiet {
			if directOnly {
				fmt.Printf("Updating and installing dependencies for %s (%s) [direct dependencies only]...\n", depFile.Path, depFile.Type)
			} else {
				fmt.Printf("Updating and installing dependencies for %s (%s)...\n", depFile.Path, depFile.Type)
			}
		}

		err := registry.UpdateFile(depFile, checker.FullUpdate, directOnly)
		if err != nil {
			// Always show errors, even in quiet mode
			fmt.Printf("Error updating and installing %s: %v\n", depFile.Path, err)
			continue
		}

		if !quiet {
			fmt.Println("  Successfully updated and installed!")
		}
	}

	return nil
}

