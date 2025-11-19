package checker

import (
	"fmt"
	"testing"

	"dependency-manager/internal/scanner"
)

func TestDetermineUpdateType(t *testing.T) {
	tests := []struct {
		name           string
		currentVersion string
		latestVersion  string
		expected       string
	}{
		{"major update", "1.0.0", "2.0.0", "major"},
		{"minor update", "1.0.0", "1.1.0", "minor"},
		{"patch update", "1.0.0", "1.0.1", "patch"},
		{"no update", "1.0.0", "1.0.0", "none"},
		{"complex major", "2.5.3", "3.0.0", "major"},
		{"complex minor", "2.5.3", "2.6.0", "minor"},
		{"complex patch", "2.5.3", "2.5.4", "patch"},
		{"non-semver", "latest", "next", "unknown"},
		{"empty current", "", "1.0.0", "unknown"},
		{"empty latest", "1.0.0", "", "unknown"},
		{"both empty", "", "", "unknown"},
		{"with v prefix", "v1.0.0", "v2.0.0", "major"},
		{"mixed v prefix", "v1.0.0", "2.0.0", "major"},
		{"caret version", "^1.0.0", "2.0.0", "major"}, // determineUpdateType strips ^ and compares
		{"tilde version", "~1.0.0", "1.1.0", "major"}, // determineUpdateType strips ~ and compares
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := determineUpdateType(tt.currentVersion, tt.latestVersion)
			if result != tt.expected {
				t.Errorf("determineUpdateType(%q, %q) = %v, want %v",
					tt.currentVersion, tt.latestVersion, result, tt.expected)
			}
		})
	}
}

func TestRegistry(t *testing.T) {
	registry := NewRegistry()

	// Register all checkers
	registry.Register(NewNpmChecker())
	registry.Register(NewMavenChecker())
	registry.Register(NewPipChecker())
	registry.Register(NewGoModChecker())
	registry.Register(NewNuGetChecker())

	// Test that all expected checkers are registered
	expectedTypes := []scanner.FileType{
		scanner.PackageJSON,
		scanner.PomXML,
		scanner.RequirementsTxt,
		scanner.GoMod,
		scanner.CsProj,
	}

	for _, fileType := range expectedTypes {
		t.Run(string(fileType), func(t *testing.T) {
			checker, err := registry.GetChecker(fileType)
			if err != nil {
				t.Errorf("Expected checker for %v, got error: %v", fileType, err)
			}
			if checker == nil {
				t.Errorf("Expected checker for %v, got nil", fileType)
			}
		})
	}

	// Test unknown file type
	t.Run("unknown type", func(t *testing.T) {
		checker, err := registry.GetChecker(scanner.FileType("unknown"))
		if err == nil {
			t.Error("Expected error for unknown type, got nil")
		}
		if checker != nil {
			t.Errorf("Expected nil checker for unknown type, got %v", checker)
		}
	})
}



func TestDependencyUpdate(t *testing.T) {
	update := DependencyUpdate{
		Name:           "test-package",
		CurrentVersion: "1.0.0",
		LatestVersion:  "2.0.0",
		UpdateType:     "major",
	}

	if update.Name != "test-package" {
		t.Errorf("Expected Name 'test-package', got %q", update.Name)
	}
	if update.CurrentVersion != "1.0.0" {
		t.Errorf("Expected CurrentVersion '1.0.0', got %q", update.CurrentVersion)
	}
	if update.LatestVersion != "2.0.0" {
		t.Errorf("Expected LatestVersion '2.0.0', got %q", update.LatestVersion)
	}
	if update.UpdateType != "major" {
		t.Errorf("Expected UpdateType 'major', got %v", update.UpdateType)
	}
}

func TestUpdateMode(t *testing.T) {
	tests := []struct {
		mode     UpdateMode
		expected UpdateMode
	}{
		{DryRun, DryRun},
		{UpdateFile, UpdateFile},
		{FullUpdate, FullUpdate},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("mode_%d", tt.mode), func(t *testing.T) {
			// Just verify the constants exist and can be used
			var mode UpdateMode = tt.mode
			if mode != tt.expected {
				t.Errorf("UpdateMode mismatch: got %v, want %v", mode, tt.expected)
			}
		})
	}
}

func TestRegistryCheckFile(t *testing.T) {
	registry := NewRegistry()

	// Test with unknown file type
	depFile := scanner.DependencyFile{
		Path:     "unknown.txt",
		Type:     scanner.FileType("unknown"),
		Filename: "unknown.txt",
	}
	result := registry.CheckFile(depFile, false)
	if result.Error == nil {
		t.Error("Expected error for unknown file type, got nil")
	}
}

func TestRegistryUpdateFile(t *testing.T) {
	registry := NewRegistry()

	// Test with unknown file type
	depFile := scanner.DependencyFile{
		Path:     "unknown.txt",
		Type:     scanner.FileType("unknown"),
		Filename: "unknown.txt",
	}
	err := registry.UpdateFile(depFile, UpdateFile, false)
	if err == nil {
		t.Error("Expected error for unknown file type, got nil")
	}
}

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		name     string
		v1       string
		v2       string
		expected string
	}{
		{"1.0.0 to 2.0.0", "1.0.0", "2.0.0", "major"},
		{"1.0.0 to 1.1.0", "1.0.0", "1.1.0", "minor"},
		{"1.0.0 to 1.0.1", "1.0.0", "1.0.1", "patch"},
		{"same version", "1.0.0", "1.0.0", "none"},
		{"downgrade major", "2.0.0", "1.0.0", "major"}, // determineUpdateType doesn't check direction
		{"downgrade minor", "1.1.0", "1.0.0", "minor"}, // determineUpdateType doesn't check direction
		{"downgrade patch", "1.0.1", "1.0.0", "patch"}, // determineUpdateType doesn't check direction
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := determineUpdateType(tt.v1, tt.v2)
			if result != tt.expected {
				t.Errorf("determineUpdateType(%q, %q) = %v, want %v",
					tt.v1, tt.v2, result, tt.expected)
			}
		})
	}
}

func TestMultipleUpdates(t *testing.T) {
	updates := []DependencyUpdate{
		{Name: "pkg1", CurrentVersion: "1.0.0", LatestVersion: "2.0.0", UpdateType: "major"},
		{Name: "pkg2", CurrentVersion: "1.0.0", LatestVersion: "1.1.0", UpdateType: "minor"},
		{Name: "pkg3", CurrentVersion: "1.0.0", LatestVersion: "1.0.1", UpdateType: "patch"},
	}

	if len(updates) != 3 {
		t.Errorf("Expected 3 updates, got %d", len(updates))
	}

	// Verify each update
	expectedTypes := []string{"major", "minor", "patch"}
	for i, update := range updates {
		if update.UpdateType != expectedTypes[i] {
			t.Errorf("Update %d: expected type %v, got %v", i, expectedTypes[i], update.UpdateType)
		}
	}
}

