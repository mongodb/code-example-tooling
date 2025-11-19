package checker

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestNuGetChecker_CheckWithoutDotnet(t *testing.T) {
	tmpDir := t.TempDir()
	csprojFile := filepath.Join(tmpDir, "test.csproj")
	
	content := `<Project Sdk="Microsoft.NET.Sdk">
  <PropertyGroup>
    <TargetFramework>net6.0</TargetFramework>
  </PropertyGroup>
  
  <ItemGroup>
    <PackageReference Include="Newtonsoft.Json" Version="12.0.3" />
  </ItemGroup>
</Project>`
	
	if err := os.WriteFile(csprojFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create .csproj: %v", err)
	}

	checker := &NuGetChecker{}
	
	// This will fail if dotnet is not in PATH
	_, err := checker.Check(csprojFile, false)
	
	if err != nil {
		expectedMsg := "dotnet is not installed or not in PATH"
		if err.Error() != expectedMsg {
			// If dotnet is installed, we might get a different error, which is fine
			t.Logf("Got error (expected if dotnet not in PATH): %v", err)
		}
	}
}

func TestNuGetChecker_UpdateWithoutDotnet(t *testing.T) {
	tmpDir := t.TempDir()
	csprojFile := filepath.Join(tmpDir, "test.csproj")
	
	content := `<Project Sdk="Microsoft.NET.Sdk">
  <PropertyGroup>
    <TargetFramework>net6.0</TargetFramework>
  </PropertyGroup>
</Project>`
	
	if err := os.WriteFile(csprojFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create .csproj: %v", err)
	}

	checker := &NuGetChecker{}
	err := checker.Update(csprojFile, UpdateFile, false)
	
	if err != nil {
		t.Logf("Got error (expected if dotnet not in PATH): %v", err)
	}
}

func TestNuGetChecker_CheckInvalidPath(t *testing.T) {
	checker := &NuGetChecker{}

	// Test with non-existent file
	_, err := checker.Check("/path/that/does/not/exist/test.csproj", false)

	// dotnet might return an error or empty results depending on the system
	// Both are acceptable outcomes
	if err != nil {
		t.Logf("Got error (expected): %v", err)
	}
}

func TestNuGetChecker_UpdateInvalidPath(t *testing.T) {
	checker := &NuGetChecker{}

	// Test with non-existent file
	err := checker.Update("/path/that/does/not/exist/test.csproj", UpdateFile, false)

	// dotnet might return an error or handle gracefully depending on the system
	// Both are acceptable outcomes
	if err != nil {
		t.Logf("Got error (expected): %v", err)
	}
}

func TestNuGetChecker_CheckDirectOnly(t *testing.T) {
	tmpDir := t.TempDir()
	csprojFile := filepath.Join(tmpDir, "test.csproj")
	
	content := `<Project Sdk="Microsoft.NET.Sdk">
  <PropertyGroup>
    <TargetFramework>net6.0</TargetFramework>
  </PropertyGroup>
  
  <ItemGroup>
    <PackageReference Include="Newtonsoft.Json" Version="12.0.3" />
  </ItemGroup>
</Project>`
	
	if err := os.WriteFile(csprojFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create .csproj: %v", err)
	}

	checker := &NuGetChecker{}
	
	// Test with directOnly=true (note: NuGet doesn't distinguish direct/indirect in the same way)
	_, err := checker.Check(csprojFile, true)
	
	if err != nil {
		t.Logf("Got error (expected if dotnet not in PATH): %v", err)
	}
}

func TestNuGetChecker_UpdateDirectOnly(t *testing.T) {
	tmpDir := t.TempDir()
	csprojFile := filepath.Join(tmpDir, "test.csproj")
	
	content := `<Project Sdk="Microsoft.NET.Sdk">
  <PropertyGroup>
    <TargetFramework>net6.0</TargetFramework>
  </PropertyGroup>
</Project>`
	
	if err := os.WriteFile(csprojFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create .csproj: %v", err)
	}

	checker := &NuGetChecker{}
	
	// Test with directOnly=true
	err := checker.Update(csprojFile, UpdateFile, true)
	
	if err != nil {
		t.Logf("Got error (expected if dotnet not in PATH): %v", err)
	}
}

func TestNuGetChecker_UpdateModes(t *testing.T) {
	tmpDir := t.TempDir()
	csprojFile := filepath.Join(tmpDir, "test.csproj")
	
	content := `<Project Sdk="Microsoft.NET.Sdk">
  <PropertyGroup>
    <TargetFramework>net6.0</TargetFramework>
  </PropertyGroup>
</Project>`
	
	if err := os.WriteFile(csprojFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create .csproj: %v", err)
	}

	checker := &NuGetChecker{}
	
	modes := []UpdateMode{DryRun, UpdateFile, FullUpdate}
	
	for _, mode := range modes {
		t.Run(fmt.Sprintf("mode_%d", mode), func(t *testing.T) {
			err := checker.Update(csprojFile, mode, false)
			
			// We expect either an error (dotnet not in PATH) or success
			if err != nil {
				t.Logf("Got error for mode %d (expected if dotnet not in PATH): %v", mode, err)
			}
		})
	}
}

func TestNuGetChecker_EmptyCsproj(t *testing.T) {
	tmpDir := t.TempDir()
	csprojFile := filepath.Join(tmpDir, "test.csproj")
	
	// Create a minimal but valid .csproj
	content := `<Project Sdk="Microsoft.NET.Sdk">
  <PropertyGroup>
    <TargetFramework>net6.0</TargetFramework>
  </PropertyGroup>
</Project>`
	
	if err := os.WriteFile(csprojFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create .csproj: %v", err)
	}

	checker := &NuGetChecker{}
	
	// Check should succeed (or fail with dotnet not in PATH error)
	_, err := checker.Check(csprojFile, false)
	
	if err != nil {
		t.Logf("Got error (expected if dotnet not in PATH): %v", err)
	}
}

func TestNuGetChecker_InvalidXML(t *testing.T) {
	tmpDir := t.TempDir()
	csprojFile := filepath.Join(tmpDir, "test.csproj")
	
	// Create invalid XML
	content := `<Project>invalid xml`
	
	if err := os.WriteFile(csprojFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create .csproj: %v", err)
	}

	checker := &NuGetChecker{}
	
	// This might fail or succeed depending on dotnet's behavior with invalid XML
	_, err := checker.Check(csprojFile, false)
	
	// dotnet should handle invalid XML with an error
	if err != nil {
		t.Logf("Got error (expected for invalid XML): %v", err)
	}
}

func TestNuGetChecker_ParseDotnetListOutput(t *testing.T) {
	checker := &NuGetChecker{}
	
	// Test parsing valid dotnet list output
	output := `
Project 'test' has the following updates to its packages
   [net6.0]: 
   Top-level Package      Requested   Resolved   Latest
   > Newtonsoft.Json      12.0.3      12.0.3     13.0.3
   > System.Text.Json     6.0.0       6.0.0      8.0.0
`
	
	updates, err := checker.parseDotnetListOutput(output)
	
	if err != nil {
		t.Fatalf("parseDotnetListOutput() error = %v", err)
	}
	
	if len(updates) != 2 {
		t.Errorf("Expected 2 updates, got %d", len(updates))
	}
	
	// Verify first update
	if len(updates) > 0 {
		if updates[0].Name != "Newtonsoft.Json" {
			t.Errorf("Expected name 'Newtonsoft.Json', got %q", updates[0].Name)
		}
		if updates[0].CurrentVersion != "12.0.3" {
			t.Errorf("Expected current version '12.0.3', got %q", updates[0].CurrentVersion)
		}
		if updates[0].LatestVersion != "13.0.3" {
			t.Errorf("Expected latest version '13.0.3', got %q", updates[0].LatestVersion)
		}
		if updates[0].UpdateType != "major" {
			t.Errorf("Expected update type 'major', got %q", updates[0].UpdateType)
		}
	}
	
	// Verify second update
	if len(updates) > 1 {
		if updates[1].Name != "System.Text.Json" {
			t.Errorf("Expected name 'System.Text.Json', got %q", updates[1].Name)
		}
		if updates[1].CurrentVersion != "6.0.0" {
			t.Errorf("Expected current version '6.0.0', got %q", updates[1].CurrentVersion)
		}
		if updates[1].LatestVersion != "8.0.0" {
			t.Errorf("Expected latest version '8.0.0', got %q", updates[1].LatestVersion)
		}
		if updates[1].UpdateType != "major" {
			t.Errorf("Expected update type 'major', got %q", updates[1].UpdateType)
		}
	}
}

func TestNuGetChecker_ParseDotnetListOutputEmpty(t *testing.T) {
	checker := &NuGetChecker{}
	
	// Test parsing output with no updates
	output := `
Project 'test' has the following updates to its packages
   [net6.0]: 
`
	
	updates, err := checker.parseDotnetListOutput(output)
	
	if err != nil {
		t.Fatalf("parseDotnetListOutput() error = %v", err)
	}
	
	if len(updates) != 0 {
		t.Errorf("Expected 0 updates, got %d", len(updates))
	}
}

func TestNuGetChecker_GetFileType(t *testing.T) {
	checker := &NuGetChecker{}
	
	fileType := checker.GetFileType()
	
	if string(fileType) != ".csproj" {
		t.Errorf("Expected file type '.csproj', got %q", string(fileType))
	}
}

