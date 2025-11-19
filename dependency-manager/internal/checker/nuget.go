package checker

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"dependency-manager/internal/scanner"
)

// NuGetChecker handles .csproj files
type NuGetChecker struct{}

// NewNuGetChecker creates a new NuGet checker
func NewNuGetChecker() *NuGetChecker {
	return &NuGetChecker{}
}

// GetFileType returns the file type this checker handles
func (n *NuGetChecker) GetFileType() scanner.FileType {
	return scanner.CsProj
}

// Check returns available updates for NuGet dependencies
func (n *NuGetChecker) Check(filePath string, directOnly bool) ([]DependencyUpdate, error) {
	dir := filepath.Dir(filePath)

	// Check if dotnet is available
	if err := exec.Command("dotnet", "--version").Run(); err != nil {
		return nil, fmt.Errorf("dotnet is not installed or not in PATH")
	}

	// Run dotnet list package --outdated
	cmd := exec.Command("dotnet", "list", "package", "--outdated")
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		// Command might fail if no outdated packages, check output
		if len(output) == 0 {
			return []DependencyUpdate{}, nil
		}
	}

	updates, err := n.parseDotnetListOutput(string(output))
	if err != nil {
		return nil, err
	}

	return updates, nil
}

// parseDotnetListOutput parses the output of 'dotnet list package --outdated'
func (n *NuGetChecker) parseDotnetListOutput(output string) ([]DependencyUpdate, error) {
	var updates []DependencyUpdate
	scanner := bufio.NewScanner(strings.NewReader(output))

	// Regex to match package lines
	// Example: "   > PackageName    1.0.0    1.0.1    1.2.0"
	// Format: package name, requested version, resolved version, latest version
	updateRegex := regexp.MustCompile(`^\s*>\s+([^\s]+)\s+([^\s]+)\s+([^\s]+)\s+([^\s]+)`)

	for scanner.Scan() {
		line := scanner.Text()
		matches := updateRegex.FindStringSubmatch(line)
		if len(matches) == 5 {
			packageName := matches[1]
			currentVersion := matches[3] // resolved version
			latestVersion := matches[4]

			updateType := determineUpdateType(currentVersion, latestVersion)
			updates = append(updates, DependencyUpdate{
				Name:           packageName,
				CurrentVersion: currentVersion,
				LatestVersion:  latestVersion,
				UpdateType:     updateType,
			})
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error parsing dotnet list output: %w", err)
	}

	return updates, nil
}

// Update updates NuGet dependencies based on the mode
func (n *NuGetChecker) Update(filePath string, mode UpdateMode, directOnly bool) error {
	dir := filepath.Dir(filePath)
	projectFile := filepath.Base(filePath)

	switch mode {
	case DryRun:
		// Already handled by Check
		return nil

	case UpdateFile:
		// Get list of outdated packages
		updates, err := n.Check(filePath, directOnly)
		if err != nil {
			return err
		}

		// Update each package
		for _, update := range updates {
			cmd := exec.Command("dotnet", "add", projectFile, "package", update.Name)
			cmd.Dir = dir
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to update %s: %v\n", update.Name, err)
				continue
			}
		}

	case FullUpdate:
		// Update project file
		if err := n.Update(filePath, UpdateFile, directOnly); err != nil {
			return err
		}

		// Restore packages
		cmd := exec.Command("dotnet", "restore")
		cmd.Dir = dir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to restore packages: %w", err)
		}

		// Build to verify
		cmd = exec.Command("dotnet", "build")
		cmd.Dir = dir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to build project: %w", err)
		}
	}

	return nil
}

