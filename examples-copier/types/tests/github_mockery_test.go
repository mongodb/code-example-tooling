package main // GitHubService is an interface for the methods we're using from the GitHub API.
import (
	"context"
	"fmt"
	"github.com/google/go-github/v48/github"
	"github.com/pkg/errors"
	"github.com/thompsch/app-tester/services"
	"github.com/thompsch/app-tester/types"

	"testing"
)

// MockGitHubService is a mock implementation of the GitHubService interface.
type MockGitHubService struct{}

// GetContents is the mock implementation of the GetContents method.
func (m *MockGitHubService) GetContents(ctx context.Context, owner, repo, path string, opt *github.RepositoryContentGetOptions) (*github.RepositoryContent, []*github.RepositoryContent, *github.Response, error) {
	if owner == "test-owner" && repo == "test-repo" && path == "test/file.txt" {
		content := "Hello, World!"

		fileContent := &github.RepositoryContent{
			Content: &content,
		}
		return fileContent, nil, nil, nil
	}
	return nil, nil, nil, errors.New("file not found")
}

// fetchFileContent uses the GitHubService interface to get file content.
func fetchFileContent(service types.GitHubService, owner, repo, filePath string) (string, error) {
	ctx := context.Background()

	fileContent, _, _, err := service.GetContents(ctx, owner, repo, filePath, nil)
	if err != nil {
		return "", err
	}

	fmt.Println(fileContent)

	content, err := fileContent.GetContent()
	if err != nil {
		return "", err
	}

	fmt.Println("******", content)

	return string(content), nil
}

// Main test function
func TestFetchFileContent(t *testing.T) {
	mockService := &MockGitHubService{}
	content := services.RetrieveFileContents(mockService, "test/file.txt")

	expectedContent := "Hello, World!"
	if content != expectedContent {
		t.Errorf("Expected %v, got %v", expectedContent, content)
	}
}

func main() {
	fmt.Println("Run tests with: go test")
}
