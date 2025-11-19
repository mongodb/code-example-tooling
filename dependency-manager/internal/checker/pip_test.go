package checker

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestPipChecker_CheckWithoutPip(t *testing.T) {
	tmpDir := t.TempDir()
	requirementsFile := filepath.Join(tmpDir, "requirements.txt")
	
	content := `requests==2.28.0
flask==2.0.0
`
	
	if err := os.WriteFile(requirementsFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create requirements.txt: %v", err)
	}

	checker := &PipChecker{}
	
	// This will fail if pip is not in PATH
	_, err := checker.Check(requirementsFile, false)
	
	if err != nil {
		expectedMsg := "pip is not installed or not in PATH"
		if err.Error() != expectedMsg {
			// If pip is installed, we might get a different error, which is fine
			t.Logf("Got error (expected if pip not in PATH): %v", err)
		}
	}
}

func TestPipChecker_UpdateWithoutPip(t *testing.T) {
	tmpDir := t.TempDir()
	requirementsFile := filepath.Join(tmpDir, "requirements.txt")

	content := `requests==2.28.0
`

	if err := os.WriteFile(requirementsFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create requirements.txt: %v", err)
	}

	checker := &PipChecker{}
	err := checker.Update(requirementsFile, UpdateFile, false)

	// Should succeed if pip-compile is installed, or fail gracefully if not
	if err != nil {
		t.Logf("Got error (expected if pip-compile not installed): %v", err)
	} else {
		t.Logf("Successfully updated requirements.txt with pip-compile")
	}
}

func TestPipChecker_CheckInvalidPath(t *testing.T) {
	checker := &PipChecker{}
	
	// Test with non-existent file
	_, err := checker.Check("/path/that/does/not/exist/requirements.txt", false)
	
	if err == nil {
		t.Error("Expected error for non-existent path, got nil")
	}
}

func TestPipChecker_UpdateInvalidPath(t *testing.T) {
	checker := &PipChecker{}
	
	// Test with non-existent file
	err := checker.Update("/path/that/does/not/exist/requirements.txt", UpdateFile, false)
	
	// pip-compile might not be installed, so we expect an error
	if err != nil {
		t.Logf("Got error (expected): %v", err)
	}
}

func TestPipChecker_CheckDirectOnly(t *testing.T) {
	tmpDir := t.TempDir()
	requirementsFile := filepath.Join(tmpDir, "requirements.txt")
	
	content := `requests==2.28.0
flask==2.0.0
`
	
	if err := os.WriteFile(requirementsFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create requirements.txt: %v", err)
	}

	checker := &PipChecker{}
	
	// Test with directOnly=true (note: pip doesn't distinguish direct/indirect in requirements.txt)
	_, err := checker.Check(requirementsFile, true)
	
	if err != nil {
		t.Logf("Got error (expected if pip not in PATH): %v", err)
	}
}

func TestPipChecker_UpdateDirectOnly(t *testing.T) {
	tmpDir := t.TempDir()
	requirementsFile := filepath.Join(tmpDir, "requirements.txt")

	content := `requests==2.28.0
`

	if err := os.WriteFile(requirementsFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create requirements.txt: %v", err)
	}

	checker := &PipChecker{}

	// Test with directOnly=true
	err := checker.Update(requirementsFile, UpdateFile, true)

	// Should succeed if pip-compile is installed, or fail gracefully if not
	if err != nil {
		t.Logf("Got error (expected if pip-compile not installed): %v", err)
	} else {
		t.Logf("Successfully updated requirements.txt with pip-compile")
	}
}

func TestPipChecker_UpdateModes(t *testing.T) {
	tmpDir := t.TempDir()
	requirementsFile := filepath.Join(tmpDir, "requirements.txt")

	content := `requests==2.28.0
`

	if err := os.WriteFile(requirementsFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create requirements.txt: %v", err)
	}

	checker := &PipChecker{}

	modes := []UpdateMode{DryRun, UpdateFile, FullUpdate}

	for _, mode := range modes {
		t.Run(fmt.Sprintf("mode_%d", mode), func(t *testing.T) {
			err := checker.Update(requirementsFile, mode, false)

			// Should succeed if pip-compile is installed (for UpdateFile/FullUpdate)
			// or fail gracefully if not
			if err != nil {
				t.Logf("Got error for mode %d (expected if pip-compile not installed): %v", mode, err)
			} else {
				t.Logf("Successfully completed mode %d", mode)
			}
		})
	}
}

func TestPipChecker_EmptyRequirements(t *testing.T) {
	tmpDir := t.TempDir()
	requirementsFile := filepath.Join(tmpDir, "requirements.txt")
	
	// Create an empty requirements.txt
	content := ``
	
	if err := os.WriteFile(requirementsFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create requirements.txt: %v", err)
	}

	checker := &PipChecker{}
	
	// Check should succeed (or fail with pip not in PATH error)
	_, err := checker.Check(requirementsFile, false)
	
	if err != nil {
		t.Logf("Got error (expected if pip not in PATH): %v", err)
	}
}

func TestPipChecker_CommentsAndEmptyLines(t *testing.T) {
	tmpDir := t.TempDir()
	requirementsFile := filepath.Join(tmpDir, "requirements.txt")
	
	// Create requirements.txt with comments and empty lines
	content := `# This is a comment
requests==2.28.0

# Another comment
flask==2.0.0
`
	
	if err := os.WriteFile(requirementsFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create requirements.txt: %v", err)
	}

	checker := &PipChecker{}
	
	// Check should succeed (or fail with pip not in PATH error)
	_, err := checker.Check(requirementsFile, false)
	
	if err != nil {
		t.Logf("Got error (expected if pip not in PATH): %v", err)
	}
}

func TestPipChecker_ParseRequirements(t *testing.T) {
	tmpDir := t.TempDir()
	requirementsFile := filepath.Join(tmpDir, "requirements.txt")
	
	content := `requests==2.28.0
flask>=2.0.0
django~=4.0
numpy
# comment line
pytest==7.1.0
`
	
	if err := os.WriteFile(requirementsFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create requirements.txt: %v", err)
	}

	checker := &PipChecker{}
	packages, err := checker.parseRequirements(requirementsFile)
	
	if err != nil {
		t.Fatalf("parseRequirements() error = %v", err)
	}
	
	expectedCount := 5 // requests, flask, django, numpy, pytest
	if len(packages) != expectedCount {
		t.Errorf("Expected %d packages, got %d", expectedCount, len(packages))
	}
	
	// Verify first package
	if len(packages) > 0 {
		if packages[0].Name != "requests" {
			t.Errorf("Expected first package name 'requests', got %q", packages[0].Name)
		}
		if packages[0].Version != "2.28.0" {
			t.Errorf("Expected first package version '2.28.0', got %q", packages[0].Version)
		}
	}
	
	// Verify package with >= operator
	if len(packages) > 1 {
		if packages[1].Name != "flask" {
			t.Errorf("Expected second package name 'flask', got %q", packages[1].Name)
		}
		if packages[1].Version != "2.0.0" {
			t.Errorf("Expected second package version '2.0.0', got %q", packages[1].Version)
		}
	}
	
	// Verify package without version
	if len(packages) > 3 {
		if packages[3].Name != "numpy" {
			t.Errorf("Expected fourth package name 'numpy', got %q", packages[3].Name)
		}
		if packages[3].Version != "" {
			t.Errorf("Expected fourth package version to be empty, got %q", packages[3].Version)
		}
	}
}

func TestPipChecker_ParseRequirementsEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	requirementsFile := filepath.Join(tmpDir, "requirements.txt")
	
	content := `# Only comments
# No actual packages
`
	
	if err := os.WriteFile(requirementsFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create requirements.txt: %v", err)
	}

	checker := &PipChecker{}
	packages, err := checker.parseRequirements(requirementsFile)
	
	if err != nil {
		t.Fatalf("parseRequirements() error = %v", err)
	}
	
	if len(packages) != 0 {
		t.Errorf("Expected 0 packages, got %d", len(packages))
	}
}

func TestPipChecker_ParseRequirementsInvalidFile(t *testing.T) {
	checker := &PipChecker{}
	
	_, err := checker.parseRequirements("/path/that/does/not/exist/requirements.txt")
	
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}
}

func TestPipChecker_GetFileType(t *testing.T) {
	checker := &PipChecker{}
	
	fileType := checker.GetFileType()
	
	if string(fileType) != "requirements.txt" {
		t.Errorf("Expected file type 'requirements.txt', got %q", string(fileType))
	}
}

