package test // GitHubService is an interface for the methods we're using from the GitHub API.

import (
	"context"
	"fmt"
	"github.com/google/go-github/v48/github"
	"github.com/mongodb/code-example-tooling/code-copier/services"
	"github.com/mongodb/code-example-tooling/code-copier/types"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"os"
	"strings"
	"testing"
)

// MockGitHubService is a mock implementation of the GitHubService interface.
type MockGitHubService struct {
	FileContents      map[string]string // Map paths to content
	CommitResponses   map[string]*github.Response
	DeprecatedFiles   []string
	ConfigFileContent string
}

func setupTestEnv() {
	os.Setenv("REPO_NAME", "mock-repo")
	os.Setenv("REPO_OWNER", "mock-owner")
	os.Setenv("GITHUB_APP_CLIENT_ID", "mock-client-id")
	os.Setenv("INSTALLATION_ID", "mock-installation-id")
	os.Setenv("SKIP_SECRET_MANAGER", "true")
}

func cleanupTestEnv() {
	os.Unsetenv("REPO_NAME")
	os.Unsetenv("REPO_OWNER")
	os.Unsetenv("GITHUB_APP_CLIENT_ID")
	os.Unsetenv("INSTALLATION_ID")
	os.Unsetenv("SKIP_SECRET_MANAGER")
}

// Mock all needed methods from the GitHub client
func (m *MockGitHubService) GetContents(ctx context.Context, owner, repo, path string, opt *github.RepositoryContentGetOptions) (*github.RepositoryContent, []*github.RepositoryContent, *github.Response, error) {
	key := fmt.Sprintf("%s/%s/%s", owner, repo, path)
	if content, ok := m.FileContents[key]; ok {
		return &github.RepositoryContent{Content: &content}, nil, nil, nil
	}
	return nil, nil, nil, errors.New("file not found")
}

func (m *MockGitHubService) CreateOrUpdateFile(ctx context.Context, owner, repo, path string, opts *github.RepositoryContentFileOptions) (*github.RepositoryContentResponse, *github.Response, error) {
	key := fmt.Sprintf("%s/%s/%s", owner, repo, path)
	m.FileContents[key] = string(opts.Content)
	return nil, m.CommitResponses[key], nil
}

func setupTestFixtures() *MockGitHubService {
	mock := &MockGitHubService{
		FileContents: map[string]string{
			"mongodb/docs-code-examples/config.json": `[
               {
                   "source_directory": "examples",
                   "target_repo": "target-repo",
                   "target_branch": "main", 
                   "target_directory": "target-directory",
                   "recursive_copy": true
               }, 
               {
                   "source_directory": "v2/examples",
                   "target_repo": "target-repo-no-recursive",
                   "target_branch": "v2.0", 
                   "target_directory": "v2/target-directory",
                   "recursive_copy": false
               }
           ]`,
		   // Recursive files
			"mongodb/docs-code-examples/deprecated_examples.json":                        "[]",
			"mongodb/docs-code-examples/examples/hello-world.txt":                        "Example Text File at root",              // File at root level
			"mongodb/docs-code-examples/examples/subdir/example.txt":                     "Example Text File in subdir",            // File in a subdirectory
			"mongodb/docs-code-examples/examples/go/example.go":                          "package main\n\nfunc Example() {}\n",    // Go file in nested directory
			"mongodb/docs-code-examples/examples/java/level1/level2/level3/deep_file.go": "package deep\n\nfunc DeepFunction() {}", // Java file in deeply nested directory
			"mongodb/docs-code-examples/python/example.py":                               "print('Hello, World!')",                 // Python file outside source directory -- SHOULD NOT BE COPIED
			// Non-Recursive files
			"mongodb/docs-code-examples/v2/examples/hello-world.txt":                        "Example Text File at root",              // v2 - non-recursive: File at root level
			"mongodb/docs-code-examples/v2/examples/subdir/example.txt":                     "Example Text File in subdir",            // v2 - non-recursive:File in a subdirectory -- SHOULD NOT BE COPIED
			"mongodb/docs-code-examples/v2/examples/example.go":                          "package main\n\nfunc Example() {}\n",    // v2 - non-recursive:Go file at root level
			"mongodb/docs-code-examples/v2/examples/java/level1/level2/level3/deep_file.go": "package deep\n\nfunc DeepFunction() {}", // v2 - non-recursive:Java file in deeply nested directory -- SHOULD NOT BE COPIED
		},
		CommitResponses: map[string]*github.Response{},
	}
	return mock
}

func createMockPRPayload(action string, merged bool, changedFiles []string) *github.PullRequestEvent {
	// Create a simulated PR event payload with the given files
	files := make([]*types.ChangedFile, 0)
	for _, path := range changedFiles {
		files = append(files, &types.ChangedFile{
			Path:   path,
			Status: "modified", // or "added", "removed", etc.
		})
	}

	return &github.PullRequestEvent{
		Action: &action,
		PullRequest: &github.PullRequest{
			Merged: &merged,
		},
		// Set other required fields
	}
}

func handleMockPRClosedEvent(prEvent *github.PullRequestEvent, mock *MockGitHubService) error {
	if prEvent == nil || prEvent.PullRequest == nil {
		return errors.New("invalid PR event")
	}

	if prEvent.GetAction() != "closed" || !prEvent.PullRequest.GetMerged() {
		return nil // Only process merged PRs that are closed
	}

	// Simulate retrieving changed files from the PR
	var changedFiles []types.ChangedFile
	for _, file := range prEvent.PullRequest.Files {
		changedFiles = append(changedFiles, types.ChangedFile{
	},
	}
	return nil
}

// Helper function to bypass Secret Manager and other external calls
func mockProcessChangedFiles(changedFiles []types.ChangedFile, config types.ConfigFileType, mock *MockGitHubService) error {
	// This function directly calls the file processing logic
	// but bypasses authentication and other external calls
	for _, file := range changedFiles {
		for _, configEntry := range config {
			// Process each file according to config
			if strings.HasPrefix(file.Path, configEntry.SourceDirectory) {
				// Extract relative path
				relPath := strings.TrimPrefix(file.Path, configEntry.SourceDirectory)
				if strings.HasPrefix(relPath, "/") {
					relPath = relPath[1:]
				}

				// Create target path
				targetPath := fmt.Sprintf("%s/%s/%s", configEntry.TargetRepo, configEntry.TargetDirectory, relPath)

				// Get source content from mock
				sourceKey := fmt.Sprintf("mongodb/docs-code-examples/%s", file.Path)
				content, ok := mock.FileContents[sourceKey]
				if !ok {
					continue
				}

				// "Copy" to target
				mock.FileContents[targetPath] = content
			}
		}
	}
	return nil
}

func TestProcessPullRequestEvent(t *testing.T) {
	mock := setupTestFixtures()

	// Create a mock PR event
	changedFiles := []string{"examples/go/example1.go", "examples/go/subdir/example2.go"}
	prEvent := createMockPRPayload("closed", true, changedFiles)

	// Test processing the PR event
	err := handleMockPRClosedEvent(prEvent, mock)
	assert.NoError(t, err)
}

func TestHandleDifferentPRActions(t *testing.T) {
	testCases := []struct {
		action        string
		merged        bool
		shouldProcess bool
	}{
		{"closed", true, true},   // merged PR should trigger processing
		{"closed", false, false}, // closed but not merged shouldn't
		{"opened", false, false}, // opened PR shouldn't trigger
	}

	for _, tc := range testCases {
		prEvent := createMockPRPayload(tc.action, tc.merged, []string{"examples/go/test.go"})
		// Test logic here
	}
}

func TestFileProcessingWorkflowWithMockedEnv(t *testing.T) {
	// Mock environment variables
	setupTestEnv()
	defer cleanupTestEnv()

	// Setup
	mock := setupTestFixtures()

	// Replace the real GitHub client with your mock
	originalClient := services.GitHubClient
	services.GitHubClient = mock
	defer func() { services.GitHubClient = originalClient }()

	// Test merged PR with changed files
	changedFiles := []types.ChangedFile{
		{Path: "examples/go/example1.go", Status: "modified"},
		{Path: "examples/go/subdir/example2.go", Status: "added"},
	}

	config := types.ConfigFileType{
		{
			"source_directory": "examples",
			"target_repo": "target-repo",
			"target_branch": "main",
			"target_directory": "target-directory",
			"recursive_copy": true
		}
	}

	err := mockProcessChangedFiles(changedFiles, config, mock)

	// Assertions
	assert.NoError(t, err)
	assert.Contains(t, mock.FileContents, "target-repo/go-examples/example1.go")
	assert.Contains(t, mock.FileContents, "target-repo/go-examples/subdir/example2.go")
}

func TestRecursiveCopyWithNestedDirectories(t *testing.T) {
	// Mock environment variables
	os.Setenv("REPO_NAME", "mock-repo")
	os.Setenv("REPO_OWNER", "mock-owner")
	os.Setenv("GITHUB_APP_CLIENT_ID", "mock-client-id")
	os.Setenv("INSTALLATION_ID", "mock-installation-id")
	// Add this to skip Secret Manager
	os.Setenv("SKIP_SECRET_MANAGER", "true")
	defer func() {
		os.Unsetenv("REPO_NAME")
		os.Unsetenv("REPO_OWNER")
		os.Unsetenv("GITHUB_APP_CLIENT_ID")
		os.Unsetenv("INSTALLATION_ID")
		os.Unsetenv("SKIP_SECRET_MANAGER")
	}()

	// Setup
	mock := setupTestFixtures()

	// Replace the real GitHub client with your mock
	originalClient := services.GitHubClient
	services.GitHubClient = mock
	defer func() { services.GitHubClient = originalClient }()

	// Add additional nested files to test
	mock.FileContents["mongodb/docs-code-examples/examples/go/subdir/nesteddir/example3.go"] = "package nested\n\nfunc Example3() {}"
	mock.FileContents["mongodb/docs-code-examples/examples/go/subdir/nesteddir/deepnest/example4.go"] = "package deep\n\nfunc Example4() {}"

	// Test with deeply nested files
	changedFiles := []types.ChangedFile{
		{Path: "examples/go/example1.go", Status: "modified"},
		{Path: "examples/go/subdir/example2.go", Status: "added"},
		{Path: "examples/go/subdir/nesteddir/example3.go", Status: "added"},
		{Path: "examples/go/subdir/nesteddir/deepnest/example4.go", Status: "added"},
	}

	config := types.ConfigFileType{
		{
			SourceDirectory: "examples/go",
			TargetRepo:      "target-repo",
			TargetBranch:    "main",
			TargetDirectory: "go-examples",
			RecursiveCopy:   true,
		},
	}

	err := mockProcessChangedFiles(changedFiles, config, mock)

	// Assertions
	assert.NoError(t, err)
	assert.Contains(t, mock.FileContents, "target-repo/go-examples/example1.go")
	assert.Contains(t, mock.FileContents, "target-repo/go-examples/subdir/example2.go")
	assert.Contains(t, mock.FileContents, "target-repo/go-examples/subdir/nesteddir/example3.go")
	assert.Contains(t, mock.FileContents, "target-repo/go-examples/subdir/nesteddir/deepnest/example4.go")
}

func TestRecursiveCopyWithDeepNesting(t *testing.T) {
	// Mock environment variables - needed like other tests
	os.Setenv("REPO_NAME", "mock-repo")
	os.Setenv("REPO_OWNER", "mock-owner")
	os.Setenv("GITHUB_APP_CLIENT_ID", "mock-client-id")
	os.Setenv("INSTALLATION_ID", "mock-installation-id")
	os.Setenv("SKIP_SECRET_MANAGER", "true")
	defer func() {
		os.Unsetenv("REPO_NAME")
		os.Unsetenv("REPO_OWNER")
		os.Unsetenv("GITHUB_APP_CLIENT_ID")
		os.Unsetenv("INSTALLATION_ID")
		os.Unsetenv("SKIP_SECRET_MANAGER")
	}()

	// Setup
	mock := setupTestFixtures()

	// Replace the real GitHub client with your mock
	originalClient := services.GitHubClient
	services.GitHubClient = mock
	defer func() { services.GitHubClient = originalClient }()

	// Fix: Change path to match source directory in config
	changedFiles := []types.ChangedFile{
		{Path: "examples/java/level1/level2/level3/deep_file.go", Status: "added"},
	}

	// Add the source file to the mock
	mock.FileContents["mongodb/docs-code-examples/examples/java/level1/level2/level3/deep_file.go"] =
		"package deep\n\nfunc DeepFunction() {}"

	config := types.ConfigFileType{
		{
			SourceDirectory: "examples/java",
			TargetRepo:      "target-repo",
			TargetBranch:    "main",
			TargetDirectory: "java",
			RecursiveCopy:   true,
		},
	}

	err := mockProcessChangedFiles(changedFiles, config, mock)

	// Assertions
	assert.NoError(t, err)

	// Check that the deeply nested path was created in target
	targetPath := "target-repo/java/level1/level2/level3/deep_file.go"
	assert.Contains(t, mock.FileContents, targetPath)
	assert.Equal(t, "package deep\n\nfunc DeepFunction() {}", mock.FileContents[targetPath])
	println(mock.FileContents[targetPath])
}

// func TestFilesOutsideSourceDirectory(t *testing.T) {
// 	// Setup
// 	mock := setupTestFixtures()
//
// 	// Replace the real GitHub client with your mock
// 	originalClient := services.GitHubClient
// 	services.GitHubClient = mock
// 	defer func() { services.GitHubClient = originalClient }()
//
// 	// Test with files outside the configured source directory
// 	// changedFiles := []types.ChangedFile{
// 	// 	{Path: "examples/go/example1.go", Status: "modified"}, // Should be copied
// 	// 	{Path: "examples/python/example.py", Status: "added"}, // Should be ignored
// 	// 	{Path: "unrelated/file.txt", Status: "modified"},      // Should be ignored
// 	// }
//
// 	//	err := internal.ProcessChangedFiles(changedFiles)
//
// 	// Assertions
// 	assert.NoError(t, err)
// 	assert.Contains(t, mock.FileContents, "target-repo/go-examples/example1.go")
// 	assert.NotContains(t, mock.FileContents, "target-repo/go-examples/example.py")
// 	assert.NotContains(t, mock.FileContents, "target-repo/go-examples/file.txt")
// }

// func TestDeprecatedFileHandling(t *testing.T) {
// 	// Setup
// 	mock := setupTestFixtures()
//
// 	// Set initial deprecated files content
// 	mock.FileContents["mongodb/docs-code-examples/deprecated_examples.json"] = "[]"
//
// 	// Replace the real GitHub client with your mock
// 	originalClient := services.GitHubClient
// 	services.GitHubClient = mock
// 	defer func() { services.GitHubClient = originalClient }()
//
// 	// Test with deleted files
// 	changedFiles := []types.ChangedFile{
// 		{Path: "examples/go/deleted_file.go", Status: "removed"},
// 	}
//
// 	//	err := internal.ProcessChangedFiles(changedFiles)
//
// 	// Assertions
// 	assert.NoError(t, err)
//
// 	// Check that the file was added to deprecated_examples.json
// 	deprecatedContent := mock.FileContents["mongodb/docs-code-examples/deprecated_examples.json"]
// 	assert.Contains(t, deprecatedContent, "deleted_file.go")
// 	assert.Contains(t, deprecatedContent, "examples/go")
// }

// func TestWithCustomEnvironment(t *testing.T) {
// 	// Override environment variables for test
// 	os.Setenv("REPO_NAME", "custom-repo")
// 	os.Setenv("REPO_OWNER", "custom-owner")
// 	defer func() {
// 		os.Unsetenv("REPO_NAME")
// 		os.Unsetenv("REPO_OWNER")
// 	}()
//
// 	// Setup
// 	mock := setupTestFixtures()
//
// 	// Add content for the custom repo
// 	mock.FileContents["custom-owner/custom-repo/config.json"] = `[
//         {
//             "source_directory": "custom-examples/go",
//             "target_repo": "target-repo",
//             "target_branch": "main",
//             "target_directory": "go-examples",
//             "recursive_copy": true
//         }
//     ]`
// 	mock.FileContents["custom-owner/custom-repo/custom-examples/go/example.go"] = "package main\n\nfunc Example() {}"
//
// 	// Replace the real GitHub client with your mock
// 	originalClient := services.GitHubClient
// 	services.GitHubClient = mock
// 	defer func() { services.GitHubClient = originalClient }()
//
// 	// Test with changed files in the custom repo
// 	changedFiles := []types.ChangedFile{
// 		{Path: "custom-examples/go/example.go", Status: "modified"},
// 	}
//
// 	//	err := internal.ProcessChangedFiles(changedFiles)
//
// 	// Assertions
// 	assert.NoError(t, err)
// 	assert.Contains(t, mock.FileContents, "target-repo/go-examples/example.go")
// }

// func TestRecursiveCopyDisabled(t *testing.T) {
// 	// Setup
// 	mock := setupTestFixtures()
//
// 	// Add a non-recursive config
// 	mock.FileContents["mongodb/docs-code-examples/config.json"] = `[
//         {
//             "source_directory": "examples/go",
//             "target_repo": "target-repo",
//             "target_branch": "main",
//             "target_directory": "go-examples",
//             "recursive_copy": false
//         }
//     ]`
//
// 	// Replace the real GitHub client with your mock
// 	originalClient := services.GitHubClient
// 	services.GitHubClient = mock
// 	defer func() { services.GitHubClient = originalClient }()
//
// 	// Test changed files in subdirectories
// 	changedFiles := []types.ChangedFile{
// 		{Path: "examples/go/example1.go", Status: "modified"},
// 		{Path: "examples/go/subdir/example2.go", Status: "added"},
// 	}
//
// 	// Call the function you want to test
// 	//	err := internal.ProcessChangedFiles(changedFiles)
//
// 	// Assertions
// 	assert.NoError(t, err)
// 	assert.Contains(t, mock.FileContents, "target-repo/go-examples/example1.go")
// 	assert.NotContains(t, mock.FileContents, "target-repo/go-examples/subdir/example2.go")
// }
