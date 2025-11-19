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

// PipChecker handles requirements.txt files
type PipChecker struct{}

// NewPipChecker creates a new pip checker
func NewPipChecker() *PipChecker {
	return &PipChecker{}
}

// GetFileType returns the file type this checker handles
func (p *PipChecker) GetFileType() scanner.FileType {
	return scanner.RequirementsTxt
}

// Check returns available updates for pip dependencies
func (p *PipChecker) Check(filePath string, directOnly bool) ([]DependencyUpdate, error) {
	// Check if pip is available
	if err := exec.Command("pip", "--version").Run(); err != nil {
		return nil, fmt.Errorf("pip is not installed or not in PATH")
	}

	// Read requirements.txt to get package names
	packages, err := p.parseRequirements(filePath)
	if err != nil {
		return nil, err
	}

	var updates []DependencyUpdate

	// Check each package for updates using pip list --outdated
	for _, pkg := range packages {
		cmd := exec.Command("pip", "list", "--outdated", "--format=json")
		output, err := cmd.Output()
		if err != nil {
			continue
		}

		// Parse JSON output to find this package
		// Simple string matching for now
		if strings.Contains(string(output), pkg.Name) {
			// Use pip-review or pip list --outdated to get version info
			currentVersion, latestVersion, err := p.getPackageVersions(pkg.Name)
			if err == nil && currentVersion != latestVersion {
				updateType := determineUpdateType(currentVersion, latestVersion)
				updates = append(updates, DependencyUpdate{
					Name:           pkg.Name,
					CurrentVersion: currentVersion,
					LatestVersion:  latestVersion,
					UpdateType:     updateType,
				})
			}
		}
	}

	return updates, nil
}

// Package represents a Python package from requirements.txt
type Package struct {
	Name    string
	Version string
}

// parseRequirements parses requirements.txt file
func (p *PipChecker) parseRequirements(filePath string) ([]Package, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open requirements.txt: %w", err)
	}
	defer file.Close()

	var packages []Package
	scanner := bufio.NewScanner(file)

	// Regex to match package specifications
	// Examples: "package==1.0.0", "package>=1.0.0", "package"
	pkgRegex := regexp.MustCompile(`^([a-zA-Z0-9_-]+)(?:==|>=|<=|>|<|~=)?([0-9.]*)?`)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		matches := pkgRegex.FindStringSubmatch(line)
		if len(matches) >= 2 {
			pkg := Package{
				Name: matches[1],
			}
			if len(matches) >= 3 {
				pkg.Version = matches[2]
			}
			packages = append(packages, pkg)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading requirements.txt: %w", err)
	}

	return packages, nil
}

// getPackageVersions gets current and latest versions for a package
func (p *PipChecker) getPackageVersions(packageName string) (string, string, error) {
	// Get current version
	cmd := exec.Command("pip", "show", packageName)
	output, err := cmd.Output()
	if err != nil {
		return "", "", err
	}

	currentVersion := ""
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "Version:") {
			currentVersion = strings.TrimSpace(strings.TrimPrefix(line, "Version:"))
			break
		}
	}

	// Get latest version using pip index
	cmd = exec.Command("pip", "index", "versions", packageName)
	output, err = cmd.Output()
	if err != nil {
		// Fallback: assume current is latest if we can't check
		return currentVersion, currentVersion, nil
	}

	latestVersion := ""
	scanner = bufio.NewScanner(strings.NewReader(string(output)))
	versionRegex := regexp.MustCompile(`Available versions: ([0-9.]+)`)
	for scanner.Scan() {
		line := scanner.Text()
		matches := versionRegex.FindStringSubmatch(line)
		if len(matches) >= 2 {
			latestVersion = matches[1]
			break
		}
	}

	if latestVersion == "" {
		latestVersion = currentVersion
	}

	return currentVersion, latestVersion, nil
}

// Update updates pip dependencies based on the mode
func (p *PipChecker) Update(filePath string, mode UpdateMode, directOnly bool) error {
	dir := filepath.Dir(filePath)

	switch mode {
	case DryRun:
		// Already handled by Check
		return nil

	case UpdateFile:
		// Use pip-upgrader or manually update requirements.txt
		// For simplicity, we'll use pip-compile if available
		// Note: pip doesn't have a built-in way to distinguish direct vs transitive dependencies
		// in requirements.txt, so directOnly flag doesn't affect pip updates
		if err := exec.Command("pip-compile", "--version").Run(); err == nil {
			// pip-compile doesn't allow same input/output file
			// Create a temporary .in file, compile it, then replace original
			inFile := filepath.Join(dir, "requirements.in")

			// Copy requirements.txt to requirements.in
			content, err := os.ReadFile(filePath)
			if err != nil {
				return fmt.Errorf("failed to read requirements.txt: %w", err)
			}
			if err := os.WriteFile(inFile, content, 0644); err != nil {
				return fmt.Errorf("failed to create requirements.in: %w", err)
			}
			defer os.Remove(inFile)

			// Run pip-compile on the .in file
			cmd := exec.Command("pip-compile", "--upgrade", inFile)
			cmd.Dir = dir
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("failed to update requirements.txt: %w", err)
			}
		} else {
			return fmt.Errorf("pip-compile (pip-tools) is not installed. Install with: pip install pip-tools")
		}

	case FullUpdate:
		// Update requirements.txt
		if err := p.Update(filePath, UpdateFile, directOnly); err != nil {
			return err
		}

		// Install dependencies
		cmd := exec.Command("pip", "install", "-r", filePath, "--upgrade")
		cmd.Dir = dir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to install dependencies: %w", err)
		}
	}

	return nil
}

