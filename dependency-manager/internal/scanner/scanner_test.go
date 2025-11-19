package scanner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsDependencyFile(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{"package.json", "/path/to/package.json", true},
		{"pom.xml", "/path/to/pom.xml", true},
		{"requirements.txt", "/path/to/requirements.txt", true},
		{"go.mod", "/path/to/go.mod", true},
		{".csproj file", "/path/to/MyProject.csproj", true},
		{"random file", "/path/to/random.txt", false},
		{"go.sum", "/path/to/go.sum", false},
		{"package-lock.json", "/path/to/package-lock.json", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsDependencyFile(tt.path)
			if result != tt.expected {
				t.Errorf("IsDependencyFile(%q) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestScanSingleFile(t *testing.T) {
	// Create a temporary directory
	tmpDir := t.TempDir()
	
	// Create a test package.json file
	packageJSON := filepath.Join(tmpDir, "package.json")
	if err := os.WriteFile(packageJSON, []byte(`{"name": "test"}`), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	scanner := New(packageJSON)
	files, err := scanner.Scan()
	
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}
	
	if len(files) != 1 {
		t.Fatalf("Expected 1 file, got %d", len(files))
	}
	
	if files[0].Type != PackageJSON {
		t.Errorf("Expected type %v, got %v", PackageJSON, files[0].Type)
	}
	
	if files[0].Filename != "package.json" {
		t.Errorf("Expected filename 'package.json', got %q", files[0].Filename)
	}
}

func TestScanDirectory(t *testing.T) {
	// Create a temporary directory structure
	tmpDir := t.TempDir()
	
	// Create test files
	files := map[string]string{
		"package.json":     `{"name": "test"}`,
		"pom.xml":          `<project></project>`,
		"requirements.txt": `requests==2.28.0`,
		"go.mod":           `module test`,
		"random.txt":       `not a dependency file`,
	}
	
	for name, content := range files {
		path := filepath.Join(tmpDir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", name, err)
		}
	}
	
	scanner := New(tmpDir)
	depFiles, err := scanner.Scan()
	
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}
	
	// Should find 4 dependency files (excluding random.txt)
	if len(depFiles) != 4 {
		t.Errorf("Expected 4 dependency files, got %d", len(depFiles))
	}
	
	// Verify all expected types are found
	foundTypes := make(map[FileType]bool)
	for _, f := range depFiles {
		foundTypes[f.Type] = true
	}
	
	expectedTypes := []FileType{PackageJSON, PomXML, RequirementsTxt, GoMod}
	for _, expectedType := range expectedTypes {
		if !foundTypes[expectedType] {
			t.Errorf("Expected to find %v, but didn't", expectedType)
		}
	}
}

func TestScanWithIgnoredDirectories(t *testing.T) {
	// Create a temporary directory structure
	tmpDir := t.TempDir()
	
	// Create package.json in root
	rootPackage := filepath.Join(tmpDir, "package.json")
	if err := os.WriteFile(rootPackage, []byte(`{"name": "root"}`), 0644); err != nil {
		t.Fatalf("Failed to create root package.json: %v", err)
	}
	
	// Create node_modules directory with package.json (should be ignored)
	nodeModules := filepath.Join(tmpDir, "node_modules")
	if err := os.Mkdir(nodeModules, 0755); err != nil {
		t.Fatalf("Failed to create node_modules: %v", err)
	}
	nodeModulesPackage := filepath.Join(nodeModules, "package.json")
	if err := os.WriteFile(nodeModulesPackage, []byte(`{"name": "ignored"}`), 0644); err != nil {
		t.Fatalf("Failed to create node_modules package.json: %v", err)
	}
	
	// Create .git directory with go.mod (should be ignored)
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.Mkdir(gitDir, 0755); err != nil {
		t.Fatalf("Failed to create .git: %v", err)
	}
	gitGoMod := filepath.Join(gitDir, "go.mod")
	if err := os.WriteFile(gitGoMod, []byte(`module test`), 0644); err != nil {
		t.Fatalf("Failed to create .git go.mod: %v", err)
	}
	
	scanner := New(tmpDir)
	depFiles, err := scanner.Scan()
	
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}
	
	// Should only find the root package.json, not the ones in ignored directories
	if len(depFiles) != 1 {
		t.Errorf("Expected 1 dependency file, got %d", len(depFiles))
		for _, f := range depFiles {
			t.Logf("Found: %s", f.Path)
		}
	}
	
	if len(depFiles) > 0 && depFiles[0].Filename != "package.json" {
		t.Errorf("Expected to find root package.json, got %s", depFiles[0].Path)
	}
}

func TestScanWithCustomIgnorePaths(t *testing.T) {
	// Create a temporary directory structure
	tmpDir := t.TempDir()
	
	// Create package.json in root
	rootPackage := filepath.Join(tmpDir, "package.json")
	if err := os.WriteFile(rootPackage, []byte(`{"name": "root"}`), 0644); err != nil {
		t.Fatalf("Failed to create root package.json: %v", err)
	}
	
	// Create custom-ignore directory with package.json
	customDir := filepath.Join(tmpDir, "custom-ignore")
	if err := os.Mkdir(customDir, 0755); err != nil {
		t.Fatalf("Failed to create custom-ignore: %v", err)
	}
	customPackage := filepath.Join(customDir, "package.json")
	if err := os.WriteFile(customPackage, []byte(`{"name": "custom"}`), 0644); err != nil {
		t.Fatalf("Failed to create custom package.json: %v", err)
	}
	
	// Scan with custom ignore paths
	scanner := NewWithIgnorePaths(tmpDir, []string{"custom-ignore"})
	depFiles, err := scanner.Scan()
	
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}
	
	// Should only find the root package.json
	if len(depFiles) != 1 {
		t.Errorf("Expected 1 dependency file, got %d", len(depFiles))
	}
}

func TestScanNonExistentPath(t *testing.T) {
	scanner := New("/path/that/does/not/exist")
	_, err := scanner.Scan()
	
	if err == nil {
		t.Error("Expected error for non-existent path, got nil")
	}
}

func TestScanCsProjFile(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create a .csproj file
	csprojFile := filepath.Join(tmpDir, "MyProject.csproj")
	if err := os.WriteFile(csprojFile, []byte(`<Project Sdk="Microsoft.NET.Sdk"></Project>`), 0644); err != nil {
		t.Fatalf("Failed to create .csproj file: %v", err)
	}
	
	scanner := New(tmpDir)
	depFiles, err := scanner.Scan()
	
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}
	
	if len(depFiles) != 1 {
		t.Fatalf("Expected 1 file, got %d", len(depFiles))
	}
	
	if depFiles[0].Type != CsProj {
		t.Errorf("Expected type %v, got %v", CsProj, depFiles[0].Type)
	}
}

func TestScanRecursiveDirectories(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create nested directory structure
	subDir1 := filepath.Join(tmpDir, "frontend")
	subDir2 := filepath.Join(tmpDir, "backend")
	subDir3 := filepath.Join(tmpDir, "backend", "api")
	
	for _, dir := range []string{subDir1, subDir2, subDir3} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}
	
	// Create dependency files in different directories
	files := map[string]string{
		filepath.Join(tmpDir, "package.json"):        `{"name": "root"}`,
		filepath.Join(subDir1, "package.json"):       `{"name": "frontend"}`,
		filepath.Join(subDir2, "requirements.txt"):   `requests==2.28.0`,
		filepath.Join(subDir3, "go.mod"):             `module api`,
	}
	
	for path, content := range files {
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", path, err)
		}
	}
	
	scanner := New(tmpDir)
	depFiles, err := scanner.Scan()
	
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}
	
	if len(depFiles) != 4 {
		t.Errorf("Expected 4 dependency files, got %d", len(depFiles))
	}
}

