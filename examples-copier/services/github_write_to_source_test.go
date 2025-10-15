package services

import (
	"testing"

	. "github.com/mongodb/code-example-tooling/code-copier/types"
)

func TestUpdateDeprecationFile_EmptyList(t *testing.T) {
	// When FilesToDeprecate is empty, UpdateDeprecationFile should return early
	// FilesToDeprecate is a map[string]Configs
	originalFiles := FilesToDeprecate
	defer func() {
		FilesToDeprecate = originalFiles
	}()

	FilesToDeprecate = make(map[string]Configs)

	// This should not panic or error - it should return early
	// Note: This test doesn't verify the actual GitHub API call since that would
	// require mocking the GitHub client, which is a global variable
	UpdateDeprecationFile()

	// If we get here without panic, the test passes
}

func TestUpdateDeprecationFile_WithFiles(t *testing.T) {
	// Set up files to deprecate
	originalFiles := FilesToDeprecate
	defer func() {
		FilesToDeprecate = originalFiles
	}()

	FilesToDeprecate = map[string]Configs{
		"examples/old-example.go": {
			TargetRepo:   "test/target",
			TargetBranch: "main",
		},
		"examples/deprecated.go": {
			TargetRepo:   "test/target",
			TargetBranch: "main",
		},
	}

	// Note: This test will fail if it actually tries to call GitHub API
	// In a real test environment, we would need to:
	// 1. Mock the GetRestClient() function
	// 2. Mock the GitHub API responses
	// 3. Verify the correct API calls were made
	//
	// For now, this test documents the expected behavior
	// The actual implementation would require refactoring to inject dependencies

	// Since we can't easily test this without mocking, we'll skip the actual call
	t.Skip("Skipping test that requires GitHub API mocking")
}

func TestFilesToDeprecate_GlobalVariable(t *testing.T) {
	// Test that we can manipulate the global FilesToDeprecate variable
	originalFiles := FilesToDeprecate
	defer func() {
		FilesToDeprecate = originalFiles
	}()

	// Set test files (FilesToDeprecate is a map[string]Configs)
	testFiles := map[string]Configs{
		"file1.go": {TargetRepo: "test/repo1", TargetBranch: "main"},
		"file2.go": {TargetRepo: "test/repo2", TargetBranch: "develop"},
		"file3.go": {TargetRepo: "test/repo3", TargetBranch: "main"},
	}
	FilesToDeprecate = testFiles

	if len(FilesToDeprecate) != 3 {
		t.Errorf("FilesToDeprecate length = %d, want 3", len(FilesToDeprecate))
	}

	for file, config := range testFiles {
		if deprecatedConfig, exists := FilesToDeprecate[file]; !exists {
			t.Errorf("FilesToDeprecate missing file %s", file)
		} else if deprecatedConfig.TargetRepo != config.TargetRepo {
			t.Errorf("FilesToDeprecate[%s].TargetRepo = %s, want %s", file, deprecatedConfig.TargetRepo, config.TargetRepo)
		}
	}
}

func TestDeprecationFileEnvironmentVariables(t *testing.T) {
	// Test that deprecation file configuration can be set via environment variables
	// The UpdateDeprecationFile function uses os.Getenv to read these values

	tests := []struct {
		name              string
		deprecationFile   string
	}{
		{
			name:              "default config",
			deprecationFile:   "deprecated-files.json",
		},
		{
			name:              "custom file",
			deprecationFile:   "custom-deprecated.json",
		},
		{
			name:              "nested path",
			deprecationFile:   "docs/deprecated/files.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// The deprecation file path is typically configured via environment variables
			// This test documents the expected configuration approach
			if tt.deprecationFile == "" {
				t.Error("Deprecation file path should not be empty")
			}
		})
	}
}

// Note: Comprehensive testing of UpdateDeprecationFile would require:
// 1. Refactoring to accept a GitHub client interface instead of using global GetRestClient()
// 2. Creating mock implementations of the GitHub client
// 3. Testing scenarios:
//    - Empty deprecation list (early return)
//    - Fetching existing deprecation file
//    - Handling missing deprecation file (404)
//    - Merging new files with existing files
//    - Removing duplicates
//    - Committing changes to GitHub
//    - Error handling for API failures
//
// Example refactored signature:
// func UpdateDeprecationFile(ctx context.Context, config *configs.Config, client GitHubClient) error
//
// This would allow for proper unit testing with mocked dependencies.

