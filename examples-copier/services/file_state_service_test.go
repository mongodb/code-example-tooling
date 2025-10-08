package services_test

import (
	"sync"
	"testing"

	"github.com/google/go-github/v48/github"
	"github.com/mongodb/code-example-tooling/code-copier/services"
	"github.com/mongodb/code-example-tooling/code-copier/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileStateService_AddAndGetFilesToUpload(t *testing.T) {
	service := services.NewFileStateService()

	key := types.UploadKey{
		RepoName:   "org/repo",
		BranchPath: "refs/heads/main",
	}

	content := types.UploadFileContent{
		TargetBranch:   "main",
		CommitStrategy: types.CommitStrategyDirect,
		CommitMessage:  "Test commit",
		Content: []github.RepositoryContent{
			{Path: github.String("test.go")},
		},
	}

	// Add file
	service.AddFileToUpload(key, content)

	// Get files
	files := service.GetFilesToUpload()
	require.Len(t, files, 1)

	retrieved, exists := files[key]
	require.True(t, exists)
	assert.Equal(t, "main", retrieved.TargetBranch)
	assert.Equal(t, types.CommitStrategyDirect, retrieved.CommitStrategy)
	assert.Equal(t, "Test commit", retrieved.CommitMessage)
	assert.Len(t, retrieved.Content, 1)
}

func TestFileStateService_AddAndGetFilesToDeprecate(t *testing.T) {
	service := services.NewFileStateService()

	entry := types.DeprecatedFileEntry{
		FileName: "old_example.go",
		Repo:     "org/repo",
		Branch:   "main",
	}

	// Add file
	service.AddFileToDeprecate("deprecated.json", entry)

	// Get files
	files := service.GetFilesToDeprecate()
	require.Len(t, files, 1)

	retrieved, exists := files["deprecated.json"]
	require.True(t, exists)
	assert.Equal(t, "old_example.go", retrieved.FileName)
	assert.Equal(t, "org/repo", retrieved.Repo)
	assert.Equal(t, "main", retrieved.Branch)
}

func TestFileStateService_ClearFilesToUpload(t *testing.T) {
	service := services.NewFileStateService()

	key := types.UploadKey{
		RepoName:   "org/repo",
		BranchPath: "refs/heads/main",
	}

	content := types.UploadFileContent{
		TargetBranch: "main",
	}

	service.AddFileToUpload(key, content)
	assert.Len(t, service.GetFilesToUpload(), 1)

	service.ClearFilesToUpload()
	assert.Len(t, service.GetFilesToUpload(), 0)
}

func TestFileStateService_ClearFilesToDeprecate(t *testing.T) {
	service := services.NewFileStateService()

	entry := types.DeprecatedFileEntry{
		FileName: "test.go",
		Repo:     "org/repo",
		Branch:   "main",
	}

	service.AddFileToDeprecate("deprecated.json", entry)
	assert.Len(t, service.GetFilesToDeprecate(), 1)

	service.ClearFilesToDeprecate()
	assert.Len(t, service.GetFilesToDeprecate(), 0)
}

func TestFileStateService_UpdateExistingFile(t *testing.T) {
	service := services.NewFileStateService()

	key := types.UploadKey{
		RepoName:   "org/repo",
		BranchPath: "refs/heads/main",
	}

	// Add first file
	content1 := types.UploadFileContent{
		TargetBranch: "main",
		Content: []github.RepositoryContent{
			{Path: github.String("file1.go")},
		},
	}
	service.AddFileToUpload(key, content1)

	// Update with second file
	content2 := types.UploadFileContent{
		TargetBranch: "main",
		Content: []github.RepositoryContent{
			{Path: github.String("file1.go")},
			{Path: github.String("file2.go")},
		},
	}
	service.AddFileToUpload(key, content2)

	// Should have replaced, not appended
	files := service.GetFilesToUpload()
	require.Len(t, files, 1)
	assert.Len(t, files[key].Content, 2)
}

func TestFileStateService_ThreadSafety(t *testing.T) {
	service := services.NewFileStateService()

	var wg sync.WaitGroup
	numGoroutines := 100

	// Concurrent writes
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			key := types.UploadKey{
				RepoName:   "org/repo",
				BranchPath: "refs/heads/main",
			}

			content := types.UploadFileContent{
				TargetBranch: "main",
			}

			service.AddFileToUpload(key, content)
		}(i)
	}

	// Concurrent reads
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = service.GetFilesToUpload()
		}()
	}

	wg.Wait()

	// Should have one entry (all goroutines wrote to same key)
	files := service.GetFilesToUpload()
	assert.Len(t, files, 1)
}

func TestFileStateService_MultipleRepos(t *testing.T) {
	service := services.NewFileStateService()

	key1 := types.UploadKey{
		RepoName:   "org/repo1",
		BranchPath: "refs/heads/main",
	}

	key2 := types.UploadKey{
		RepoName:   "org/repo2",
		BranchPath: "refs/heads/develop",
	}

	content1 := types.UploadFileContent{
		TargetBranch: "main",
		Content: []github.RepositoryContent{
			{Path: github.String("file1.go")},
		},
	}

	content2 := types.UploadFileContent{
		TargetBranch: "develop",
		Content: []github.RepositoryContent{
			{Path: github.String("file2.go")},
		},
	}

	service.AddFileToUpload(key1, content1)
	service.AddFileToUpload(key2, content2)

	files := service.GetFilesToUpload()
	require.Len(t, files, 2)

	assert.Equal(t, "main", files[key1].TargetBranch)
	assert.Equal(t, "develop", files[key2].TargetBranch)
}

func TestFileStateService_IsolatedCopies(t *testing.T) {
	service := services.NewFileStateService()

	key := types.UploadKey{
		RepoName:   "org/repo",
		BranchPath: "refs/heads/main",
	}

	content := types.UploadFileContent{
		TargetBranch: "main",
		Content: []github.RepositoryContent{
			{Path: github.String("file1.go")},
		},
	}

	service.AddFileToUpload(key, content)

	// Get first copy
	files1 := service.GetFilesToUpload()

	// Get second copy
	files2 := service.GetFilesToUpload()

	// Modify first copy (should not affect second)
	for k := range files1 {
		delete(files1, k)
	}

	// Second copy should still have the data
	assert.Len(t, files2, 1)

	// Original service should still have the data
	assert.Len(t, service.GetFilesToUpload(), 1)
}

func TestFileStateService_CommitStrategyTypes(t *testing.T) {
	service := services.NewFileStateService()

	tests := []struct {
		name     string
		strategy types.CommitStrategy
	}{
		{
			name:     "direct commit",
			strategy: types.CommitStrategyDirect,
		},
		{
			name:     "pull request",
			strategy: types.CommitStrategyPR,
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := types.UploadKey{
				RepoName:   "org/repo",
				BranchPath: "refs/heads/main",
			}

			content := types.UploadFileContent{
				TargetBranch:   "main",
				CommitStrategy: tt.strategy,
				CommitMessage:  "Test",
				PRTitle:        "Test PR",
				AutoMergePR:    i%2 == 0,
			}

			service.AddFileToUpload(key, content)

			files := service.GetFilesToUpload()
			retrieved := files[key]

			assert.Equal(t, tt.strategy, retrieved.CommitStrategy)
			assert.Equal(t, "Test", retrieved.CommitMessage)
			assert.Equal(t, "Test PR", retrieved.PRTitle)
			assert.Equal(t, i%2 == 0, retrieved.AutoMergePR)

			service.ClearFilesToUpload()
		})
	}
}

