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

// GoModChecker handles go.mod files
type GoModChecker struct{}

// NewGoModChecker creates a new Go modules checker
func NewGoModChecker() *GoModChecker {
	return &GoModChecker{}
}

// GetFileType returns the file type this checker handles
func (g *GoModChecker) GetFileType() scanner.FileType {
	return scanner.GoMod
}

// Check returns available updates for Go module dependencies
func (g *GoModChecker) Check(filePath string, directOnly bool) ([]DependencyUpdate, error) {
	dir := filepath.Dir(filePath)

	// Check if go is available
	if err := exec.Command("go", "version").Run(); err != nil {
		return nil, fmt.Errorf("go is not installed or not in PATH")
	}

	// Run go list -u -m all to get update information
	cmd := exec.Command("go", "list", "-u", "-m", "all")
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to check for updates: %w", err)
	}

	updates, err := g.parseGoListOutput(string(output), filePath, directOnly)
	if err != nil {
		return nil, err
	}

	return updates, nil
}

// parseGoListOutput parses the output of 'go list -u -m all'
func (g *GoModChecker) parseGoListOutput(output string, filePath string, directOnly bool) ([]DependencyUpdate, error) {
	var updates []DependencyUpdate

	// If directOnly is true, we need to read go.mod to get direct dependencies
	var directDeps map[string]bool
	if directOnly {
		var err error
		directDeps, err = g.getDirectDependencies(filePath)
		if err != nil {
			return nil, err
		}
	}

	scanner := bufio.NewScanner(strings.NewReader(output))

	// Regex to match module lines with updates
	// Example: "github.com/spf13/cobra v1.7.0 [v1.8.0]"
	updateRegex := regexp.MustCompile(`^([^\s]+)\s+v([^\s]+)\s+\[v([^\]]+)\]`)

	for scanner.Scan() {
		line := scanner.Text()
		matches := updateRegex.FindStringSubmatch(line)
		if len(matches) == 4 {
			moduleName := matches[1]
			currentVersion := matches[2]
			latestVersion := matches[3]

			// If directOnly is true, skip indirect dependencies
			if directOnly && !directDeps[moduleName] {
				continue
			}

			updateType := determineUpdateType(currentVersion, latestVersion)
			updates = append(updates, DependencyUpdate{
				Name:           moduleName,
				CurrentVersion: "v" + currentVersion,
				LatestVersion:  "v" + latestVersion,
				UpdateType:     updateType,
			})
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error parsing go list output: %w", err)
	}

	return updates, nil
}

// getDirectDependencies reads go.mod and returns a map of direct dependencies
func (g *GoModChecker) getDirectDependencies(filePath string) (map[string]bool, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open go.mod: %w", err)
	}
	defer file.Close()

	directDeps := make(map[string]bool)
	scanner := bufio.NewScanner(file)
	inRequireBlock := false

	// Regex to match require lines
	// Example: "github.com/spf13/cobra v1.7.0" or "require github.com/spf13/cobra v1.7.0"
	// Note: lines with "// indirect" are indirect dependencies
	requireRegex := regexp.MustCompile(`(?:require\s+)?([^\s]+)\s+v[^\s]+(?:\s+//\s*indirect)?`)
	indirectRegex := regexp.MustCompile(`//\s*indirect`)

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Check for require block
		if strings.HasPrefix(trimmed, "require (") {
			inRequireBlock = true
			continue
		}
		if inRequireBlock && trimmed == ")" {
			inRequireBlock = false
			continue
		}

		// Parse require lines
		if strings.HasPrefix(trimmed, "require ") || inRequireBlock {
			matches := requireRegex.FindStringSubmatch(line)
			if len(matches) >= 2 {
				moduleName := matches[1]
				// Only add if it's NOT marked as indirect
				if !indirectRegex.MatchString(line) {
					directDeps[moduleName] = true
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading go.mod: %w", err)
	}

	return directDeps, nil
}

// Update updates Go module dependencies based on the mode
func (g *GoModChecker) Update(filePath string, mode UpdateMode, directOnly bool) error {
	dir := filepath.Dir(filePath)

	switch mode {
	case DryRun:
		// Already handled by Check
		return nil

	case UpdateFile:
		if directOnly {
			// Get direct dependencies and update them individually
			directDeps, err := g.getDirectDependencies(filePath)
			if err != nil {
				return err
			}

			// Update each direct dependency
			for dep := range directDeps {
				cmd := exec.Command("go", "get", "-u", dep)
				cmd.Dir = dir
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				if err := cmd.Run(); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to update %s: %v\n", dep, err)
					continue
				}
			}
		} else {
			// Use go get -u to update all dependencies
			cmd := exec.Command("go", "get", "-u", "./...")
			cmd.Dir = dir
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("failed to update go.mod: %w", err)
			}
		}

		// Run go mod tidy to clean up
		cmd := exec.Command("go", "mod", "tidy")
		cmd.Dir = dir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to tidy go.mod: %w", err)
		}

	case FullUpdate:
		// Update go.mod
		if err := g.Update(filePath, UpdateFile, directOnly); err != nil {
			return err
		}

		// Download dependencies
		cmd := exec.Command("go", "mod", "download")
		cmd.Dir = dir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to download dependencies: %w", err)
		}

		// Verify dependencies
		cmd = exec.Command("go", "mod", "verify")
		cmd.Dir = dir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to verify dependencies: %w", err)
		}
	}

	return nil
}

