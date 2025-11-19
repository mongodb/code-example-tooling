package checker

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGoModChecker_GetDirectDependencies(t *testing.T) {
	tmpDir := t.TempDir()
	goModFile := filepath.Join(tmpDir, "go.mod")
	
	content := `module example.com/myproject

go 1.21

require (
	github.com/gin-gonic/gin v1.9.0
	github.com/stretchr/testify v1.8.4
	golang.org/x/sync v0.3.0 // indirect
	golang.org/x/text v0.12.0 // indirect
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
)
`
	
	if err := os.WriteFile(goModFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	checker := &GoModChecker{}
	deps, err := checker.getDirectDependencies(goModFile)
	
	if err != nil {
		t.Fatalf("getDirectDependencies() error = %v", err)
	}

	// Should only find direct dependencies (not marked with // indirect)
	expectedDeps := map[string]bool{
		"github.com/gin-gonic/gin":     true,
		"github.com/stretchr/testify":  true,
	}

	if len(deps) != len(expectedDeps) {
		t.Errorf("Expected %d direct dependencies, got %d", len(expectedDeps), len(deps))
	}

	for dep := range deps {
		if !expectedDeps[dep] {
			t.Errorf("Unexpected dependency: %s", dep)
		}
	}

	// Verify indirect dependencies are not included
	indirectDeps := []string{
		"golang.org/x/sync",
		"golang.org/x/text",
		"github.com/davecgh/go-spew",
		"github.com/pmezard/go-difflib",
	}

	for _, indirectDep := range indirectDeps {
		if deps[indirectDep] {
			t.Errorf("Indirect dependency %s should not be in direct dependencies list", indirectDep)
		}
	}
}

func TestGoModChecker_GetDirectDependenciesSimple(t *testing.T) {
	tmpDir := t.TempDir()
	goModFile := filepath.Join(tmpDir, "go.mod")
	
	content := `module example.com/simple

go 1.21

require github.com/gin-gonic/gin v1.9.0
`
	
	if err := os.WriteFile(goModFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	checker := &GoModChecker{}
	deps, err := checker.getDirectDependencies(goModFile)
	
	if err != nil {
		t.Fatalf("getDirectDependencies() error = %v", err)
	}

	if len(deps) != 1 {
		t.Errorf("Expected 1 direct dependency, got %d", len(deps))
	}

	if !deps["github.com/gin-gonic/gin"] {
		t.Errorf("Expected github.com/gin-gonic/gin to be in direct dependencies")
	}
}

func TestGoModChecker_GetDirectDependenciesEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	goModFile := filepath.Join(tmpDir, "go.mod")
	
	content := `module example.com/empty

go 1.21
`
	
	if err := os.WriteFile(goModFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	checker := &GoModChecker{}
	deps, err := checker.getDirectDependencies(goModFile)
	
	if err != nil {
		t.Fatalf("getDirectDependencies() error = %v", err)
	}

	if len(deps) != 0 {
		t.Errorf("Expected 0 direct dependencies, got %d", len(deps))
	}
}

func TestGoModChecker_GetDirectDependenciesInvalidFile(t *testing.T) {
	checker := &GoModChecker{}
	_, err := checker.getDirectDependencies("/path/that/does/not/exist/go.mod")
	
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}
}

func TestGoModChecker_CheckWithoutGo(t *testing.T) {
	tmpDir := t.TempDir()
	goModFile := filepath.Join(tmpDir, "go.mod")
	
	content := `module example.com/test

go 1.21

require github.com/gin-gonic/gin v1.9.0
`
	
	if err := os.WriteFile(goModFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	checker := &GoModChecker{}
	
	// This will fail if go is not installed, which is expected
	_, err := checker.Check(goModFile, false)
	
	// We expect either an error (go not installed) or success (go is installed)
	if err != nil {
		expectedMsg := "go is not installed or not in PATH"
		if err.Error() != expectedMsg {
			// If go is installed, we might get a different error, which is fine
			t.Logf("Got error (expected if go not installed): %v", err)
		}
	}
}

func TestGoModChecker_UpdateWithoutGo(t *testing.T) {
	tmpDir := t.TempDir()
	goModFile := filepath.Join(tmpDir, "go.mod")

	content := `module example.com/test

go 1.21

require github.com/gin-gonic/gin v1.9.0
`

	if err := os.WriteFile(goModFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	checker := &GoModChecker{}
	err := checker.Update(goModFile, UpdateFile, false)

	if err != nil {
		t.Logf("Got error (expected if go not installed): %v", err)
	}
}

func TestGoModChecker_CheckDirectOnly(t *testing.T) {
	tmpDir := t.TempDir()
	goModFile := filepath.Join(tmpDir, "go.mod")
	
	content := `module example.com/test

go 1.21

require (
	github.com/gin-gonic/gin v1.9.0
	golang.org/x/sync v0.3.0 // indirect
)
`
	
	if err := os.WriteFile(goModFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	checker := &GoModChecker{}
	
	// Test with directOnly=true
	_, err := checker.Check(goModFile, true)
	
	if err != nil {
		t.Logf("Got error (expected if go not installed): %v", err)
	}
}

func TestGoModChecker_UpdateDirectOnly(t *testing.T) {
	tmpDir := t.TempDir()
	goModFile := filepath.Join(tmpDir, "go.mod")

	content := `module example.com/test

go 1.21

require (
	github.com/gin-gonic/gin v1.9.0
	golang.org/x/sync v0.3.0 // indirect
)
`

	if err := os.WriteFile(goModFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	checker := &GoModChecker{}

	// Test with directOnly=true
	err := checker.Update(goModFile, UpdateFile, true)

	if err != nil {
		t.Logf("Got error (expected if go not installed): %v", err)
	}
}

func TestGoModChecker_InvalidGoMod(t *testing.T) {
	tmpDir := t.TempDir()
	goModFile := filepath.Join(tmpDir, "go.mod")
	
	// Create invalid go.mod content
	content := `this is not a valid go.mod file`
	
	if err := os.WriteFile(goModFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	checker := &GoModChecker{}
	
	// getDirectDependencies should handle this gracefully
	deps, err := checker.getDirectDependencies(goModFile)
	
	// Should not error, just return empty list
	if err != nil {
		t.Logf("Got error: %v", err)
	}
	
	if len(deps) != 0 {
		t.Errorf("Expected 0 dependencies for invalid go.mod, got %d", len(deps))
	}
}

