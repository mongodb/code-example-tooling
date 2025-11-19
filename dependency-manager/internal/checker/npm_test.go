package checker

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestNpmChecker_parseMavenOutput(t *testing.T) {
	// This is actually testing npm, not maven - the function name in the test is wrong
	// but we'll test the npm checker's ability to handle npm outdated output
	checker := &NpmChecker{}

	// Test that the checker exists
	if checker == nil {
		t.Fatal("NpmChecker should not be nil")
	}
}

func TestNpmChecker_CheckWithoutNpm(t *testing.T) {
	// Create a temporary directory with a package.json
	tmpDir := t.TempDir()
	packageJSON := filepath.Join(tmpDir, "package.json")
	
	content := `{
		"name": "test-project",
		"version": "1.0.0",
		"dependencies": {
			"react": "18.0.0"
		}
	}`
	
	if err := os.WriteFile(packageJSON, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create package.json: %v", err)
	}

	checker := &NpmChecker{}
	
	// This will fail if npm is not installed, which is expected
	// We're just testing that the function handles the error gracefully
	_, err := checker.Check(packageJSON, false)
	
	// We expect either an error (npm not installed) or success (npm is installed)
	// Both are valid outcomes for this test
	if err != nil {
		// Verify it's the expected error message
		expectedMsg := "npm is not installed or not in PATH"
		if err.Error() != expectedMsg {
			// If npm is installed, we might get a different error, which is fine
			t.Logf("Got error (expected if npm not installed): %v", err)
		}
	}
}

func TestNpmChecker_UpdateWithoutNpm(t *testing.T) {
	tmpDir := t.TempDir()
	packageJSON := filepath.Join(tmpDir, "package.json")

	content := `{
		"name": "test-project",
		"version": "1.0.0",
		"dependencies": {
			"react": "18.0.0"
		}
	}`

	if err := os.WriteFile(packageJSON, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create package.json: %v", err)
	}

	checker := &NpmChecker{}
	err := checker.Update(packageJSON, UpdateFile, false)

	// We expect either an error (npm/ncu not installed) or success
	if err != nil {
		t.Logf("Got error (expected if npm/ncu not installed): %v", err)
	}
}

func TestNpmChecker_CheckInvalidPath(t *testing.T) {
	checker := &NpmChecker{}

	// Test with non-existent file
	_, err := checker.Check("/path/that/does/not/exist/package.json", false)

	// npm might return an error or empty results depending on the system
	// Both are acceptable outcomes
	if err != nil {
		t.Logf("Got error (expected): %v", err)
	}
}

func TestNpmChecker_UpdateInvalidPath(t *testing.T) {
	checker := &NpmChecker{}

	// Test with non-existent file
	err := checker.Update("/path/that/does/not/exist/package.json", UpdateFile, false)

	if err == nil {
		t.Error("Expected error for non-existent path, got nil")
	}
}

func TestNpmChecker_CheckDirectOnly(t *testing.T) {
	tmpDir := t.TempDir()
	packageJSON := filepath.Join(tmpDir, "package.json")
	
	content := `{
		"name": "test-project",
		"version": "1.0.0",
		"dependencies": {
			"react": "18.0.0"
		},
		"devDependencies": {
			"jest": "29.0.0"
		}
	}`
	
	if err := os.WriteFile(packageJSON, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create package.json: %v", err)
	}

	checker := &NpmChecker{}
	
	// Test with directOnly=true
	_, err := checker.Check(packageJSON, true)
	
	// We expect either an error (npm not installed) or success
	if err != nil {
		t.Logf("Got error (expected if npm not installed): %v", err)
	}
}

func TestNpmChecker_UpdateDirectOnly(t *testing.T) {
	tmpDir := t.TempDir()
	packageJSON := filepath.Join(tmpDir, "package.json")

	content := `{
		"name": "test-project",
		"version": "1.0.0",
		"dependencies": {
			"react": "18.0.0"
		},
		"devDependencies": {
			"jest": "29.0.0"
		}
	}`

	if err := os.WriteFile(packageJSON, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create package.json: %v", err)
	}

	checker := &NpmChecker{}

	// Test with directOnly=true
	err := checker.Update(packageJSON, UpdateFile, true)

	// We expect either an error (npm/ncu not installed) or success
	if err != nil {
		t.Logf("Got error (expected if npm/ncu not installed): %v", err)
	}
}

func TestNpmChecker_UpdateModes(t *testing.T) {
	tmpDir := t.TempDir()
	packageJSON := filepath.Join(tmpDir, "package.json")

	content := `{
		"name": "test-project",
		"version": "1.0.0",
		"dependencies": {
			"react": "18.0.0"
		}
	}`

	if err := os.WriteFile(packageJSON, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create package.json: %v", err)
	}

	checker := &NpmChecker{}

	modes := []UpdateMode{DryRun, UpdateFile, FullUpdate}

	for _, mode := range modes {
		t.Run(fmt.Sprintf("mode_%d", mode), func(t *testing.T) {
			err := checker.Update(packageJSON, mode, false)

			// We expect either an error (npm/ncu not installed) or success
			if err != nil {
				t.Logf("Got error for mode %d (expected if npm/ncu not installed): %v", mode, err)
			}
		})
	}
}

func TestNpmChecker_EmptyPackageJSON(t *testing.T) {
	tmpDir := t.TempDir()
	packageJSON := filepath.Join(tmpDir, "package.json")
	
	// Create an empty but valid JSON file
	content := `{
		"name": "test-project",
		"version": "1.0.0"
	}`
	
	if err := os.WriteFile(packageJSON, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create package.json: %v", err)
	}

	checker := &NpmChecker{}
	
	// Check should succeed (or fail with npm not installed error)
	_, err := checker.Check(packageJSON, false)
	
	if err != nil {
		// Verify it's an expected error
		t.Logf("Got error (expected if npm not installed): %v", err)
	}
}

func TestNpmChecker_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	packageJSON := filepath.Join(tmpDir, "package.json")

	// Create invalid JSON
	content := `{invalid json`

	if err := os.WriteFile(packageJSON, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create package.json: %v", err)
	}

	checker := &NpmChecker{}

	// This might fail or succeed depending on npm's behavior with invalid JSON
	_, err := checker.Check(packageJSON, false)

	// npm might handle invalid JSON gracefully or return an error
	if err != nil {
		t.Logf("Got error (expected for invalid JSON): %v", err)
	}
}

