package checker

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"dependency-manager/internal/scanner"
)

// NpmChecker handles package.json files
type NpmChecker struct{}

// NewNpmChecker creates a new npm checker
func NewNpmChecker() *NpmChecker {
	return &NpmChecker{}
}

// GetFileType returns the file type this checker handles
func (n *NpmChecker) GetFileType() scanner.FileType {
	return scanner.PackageJSON
}

// Check returns available updates for npm dependencies
func (n *NpmChecker) Check(filePath string, directOnly bool) ([]DependencyUpdate, error) {
	dir := filepath.Dir(filePath)

	// Check if npm is available
	if err := exec.Command("npm", "--version").Run(); err != nil {
		return nil, fmt.Errorf("npm is not installed or not in PATH")
	}

	// Run npm outdated to get update information
	// If directOnly is true, use --omit=dev to exclude devDependencies
	args := []string{"outdated", "--json"}
	if directOnly {
		args = append(args, "--omit=dev")
	}
	cmd := exec.Command("npm", args...)
	cmd.Dir = dir
	output, err := cmd.Output()

	// npm outdated returns exit code 1 when there are outdated packages
	// So we need to check if output is valid JSON even if there's an error
	if err != nil && len(output) == 0 {
		// No outdated packages or error running command
		return []DependencyUpdate{}, nil
	}

	var outdated map[string]struct {
		Current string `json:"current"`
		Wanted  string `json:"wanted"`
		Latest  string `json:"latest"`
		Type    string `json:"type"`
	}

	if err := json.Unmarshal(output, &outdated); err != nil {
		return nil, fmt.Errorf("failed to parse npm outdated output: %w", err)
	}

	var updates []DependencyUpdate
	for name, info := range outdated {
		// Use 'wanted' as fallback if 'current' is empty (package not installed)
		currentVersion := info.Current
		if currentVersion == "" {
			currentVersion = info.Wanted
		}

		// Skip if current version matches latest (no update needed)
		if currentVersion == info.Latest {
			continue
		}

		updateType := determineUpdateType(currentVersion, info.Latest)
		updates = append(updates, DependencyUpdate{
			Name:           name,
			CurrentVersion: currentVersion,
			LatestVersion:  info.Latest,
			UpdateType:     updateType,
		})
	}

	return updates, nil
}

// Update updates npm dependencies based on the mode
func (n *NpmChecker) Update(filePath string, mode UpdateMode, directOnly bool) error {
	dir := filepath.Dir(filePath)

	switch mode {
	case DryRun:
		// Already handled by Check
		return nil

	case UpdateFile:
		// Use npm-check-updates to update package.json
		// First check if ncu is available
		if err := exec.Command("ncu", "--version").Run(); err != nil {
			return fmt.Errorf("npm-check-updates (ncu) is not installed. Install with: npm install -g npm-check-updates")
		}

		// If directOnly is true, use --dep prod to only update production dependencies
		args := []string{"-u"}
		if directOnly {
			args = append(args, "--target", "prod")
		}
		cmd := exec.Command("ncu", args...)
		cmd.Dir = dir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to update package.json: %w", err)
		}

	case FullUpdate:
		// Update package.json
		if err := n.Update(filePath, UpdateFile, directOnly); err != nil {
			return err
		}

		// Install dependencies
		// If directOnly is true, use --omit=dev to only install production dependencies
		args := []string{"install"}
		if directOnly {
			args = append(args, "--omit=dev")
		}
		cmd := exec.Command("npm", args...)
		cmd.Dir = dir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to install dependencies: %w", err)
		}
	}

	return nil
}

// determineUpdateType determines if an update is major, minor, or patch
func determineUpdateType(current, latest string) string {
	// Remove 'v' prefix if present
	current = strings.TrimPrefix(current, "v")
	latest = strings.TrimPrefix(latest, "v")

	currentParts := strings.Split(current, ".")
	latestParts := strings.Split(latest, ".")

	if len(currentParts) < 3 || len(latestParts) < 3 {
		return "unknown"
	}

	if currentParts[0] != latestParts[0] {
		return "major"
	}
	if currentParts[1] != latestParts[1] {
		return "minor"
	}
	if currentParts[2] != latestParts[2] {
		return "patch"
	}

	return "none"
}

