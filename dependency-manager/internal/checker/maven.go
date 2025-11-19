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

// MavenChecker handles pom.xml files
type MavenChecker struct{}

// NewMavenChecker creates a new Maven checker
func NewMavenChecker() *MavenChecker {
	return &MavenChecker{}
}

// GetFileType returns the file type this checker handles
func (m *MavenChecker) GetFileType() scanner.FileType {
	return scanner.PomXML
}

// Check returns available updates for Maven dependencies
func (m *MavenChecker) Check(filePath string, directOnly bool) ([]DependencyUpdate, error) {
	dir := filepath.Dir(filePath)

	// Check if mvn is available
	if err := exec.Command("mvn", "--version").Run(); err != nil {
		return nil, fmt.Errorf("maven is not installed or not in PATH")
	}

	// Run mvn versions:display-dependency-updates and capture output
	cmd := exec.Command("mvn", "versions:display-dependency-updates", "-q")
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to check for updates: %w", err)
	}

	// Parse the output
	updates, err := m.parseMavenOutput(string(output))
	if err != nil {
		return nil, err
	}

	return updates, nil
}

// parseMavenOutput parses Maven dependency updates from command output
func (m *MavenChecker) parseMavenOutput(output string) ([]DependencyUpdate, error) {
	var updates []DependencyUpdate
	scanner := bufio.NewScanner(strings.NewReader(output))

	// Regex to match dependency update lines
	// Example: "  org.springframework:spring-core .................... 5.3.0 -> 5.3.10"
	updateRegex := regexp.MustCompile(`^\s+([^:]+):([^\s]+)\s+.*\s+([^\s]+)\s+->\s+([^\s]+)`)

	for scanner.Scan() {
		line := scanner.Text()
		matches := updateRegex.FindStringSubmatch(line)
		if len(matches) == 5 {
			groupID := matches[1]
			artifactID := matches[2]
			currentVersion := matches[3]
			latestVersion := matches[4]

			updateType := determineUpdateType(currentVersion, latestVersion)
			updates = append(updates, DependencyUpdate{
				Name:           fmt.Sprintf("%s:%s", groupID, artifactID),
				CurrentVersion: currentVersion,
				LatestVersion:  latestVersion,
				UpdateType:     updateType,
			})
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error parsing maven output: %w", err)
	}

	return updates, nil
}

// Update updates Maven dependencies based on the mode
func (m *MavenChecker) Update(filePath string, mode UpdateMode, directOnly bool) error {
	dir := filepath.Dir(filePath)

	switch mode {
	case DryRun:
		// Already handled by Check
		return nil

	case UpdateFile:
		// Use mvn versions:use-latest-releases to update pom.xml
		// Note: Maven doesn't have a built-in way to distinguish direct vs transitive dependencies
		// in the update command, so directOnly flag doesn't affect Maven updates
		cmd := exec.Command("mvn", "versions:use-latest-releases")
		cmd.Dir = dir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to update pom.xml: %w", err)
		}

		// Clean up backup files
		backupFile := filepath.Join(dir, "pom.xml.versionsBackup")
		os.Remove(backupFile)

	case FullUpdate:
		// Update pom.xml
		if err := m.Update(filePath, UpdateFile, directOnly); err != nil {
			return err
		}

		// Install/update dependencies
		cmd := exec.Command("mvn", "clean", "install")
		cmd.Dir = dir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to install dependencies: %w", err)
		}
	}

	return nil
}

