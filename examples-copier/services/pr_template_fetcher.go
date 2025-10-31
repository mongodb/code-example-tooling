package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/v48/github"
)

// PRTemplateFetcher defines the interface for fetching PR templates from repositories
type PRTemplateFetcher interface {
	// FetchPRTemplate fetches the PR template from a target repository
	// Returns the template content, or empty string if not found
	FetchPRTemplate(ctx context.Context, client *github.Client, repoFullName string, branch string) (string, error)
}

// DefaultPRTemplateFetcher implements PRTemplateFetcher
type DefaultPRTemplateFetcher struct{}

// NewPRTemplateFetcher creates a new PR template fetcher
func NewPRTemplateFetcher() PRTemplateFetcher {
	return &DefaultPRTemplateFetcher{}
}

// FetchPRTemplate fetches the PR template from a target repository
// It checks multiple common locations for PR templates:
// 1. .github/pull_request_template.md
// 2. .github/PULL_REQUEST_TEMPLATE.md
// 3. docs/pull_request_template.md
// 4. PULL_REQUEST_TEMPLATE.md
func (f *DefaultPRTemplateFetcher) FetchPRTemplate(ctx context.Context, client *github.Client, repoFullName string, branch string) (string, error) {
	// Parse repo owner and name
	parts := strings.Split(repoFullName, "/")
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid repo format: %s (expected owner/repo)", repoFullName)
	}
	owner := parts[0]
	repo := parts[1]

	// Common PR template locations (in order of preference)
	templatePaths := []string{
		".github/pull_request_template.md",
		".github/PULL_REQUEST_TEMPLATE.md",
		"docs/pull_request_template.md",
		"PULL_REQUEST_TEMPLATE.md",
		"pull_request_template.md",
	}

	// Try each location
	for _, path := range templatePaths {
		content, err := f.fetchFileContent(ctx, client, owner, repo, path, branch)
		if err == nil && content != "" {
			LogInfo(fmt.Sprintf("Found PR template in %s/%s at %s", owner, repo, path))
			return content, nil
		}
		// Continue to next location if not found
	}

	// No template found
	LogDebug(fmt.Sprintf("No PR template found in %s/%s (checked %d locations)", owner, repo, len(templatePaths)))
	return "", nil
}

// fetchFileContent fetches the content of a file from a repository
func (f *DefaultPRTemplateFetcher) fetchFileContent(ctx context.Context, client *github.Client, owner, repo, path, branch string) (string, error) {
	opts := &github.RepositoryContentGetOptions{
		Ref: branch,
	}

	fileContent, _, _, err := client.Repositories.GetContents(ctx, owner, repo, path, opts)
	if err != nil {
		// File not found or other error - return empty
		return "", err
	}

	if fileContent == nil {
		return "", fmt.Errorf("file content is nil")
	}

	// Decode the content
	content, err := fileContent.GetContent()
	if err != nil {
		return "", fmt.Errorf("failed to decode content: %w", err)
	}

	return content, nil
}

// MergePRBodyWithTemplate merges a configured PR body with a PR template
// The template is placed first, then the configured body is appended, separated by a horizontal rule
func MergePRBodyWithTemplate(configuredBody, template string) string {
	if template == "" {
		return configuredBody
	}

	if configuredBody == "" {
		return template
	}

	// Merge: template first, then separator, then configured body
	return fmt.Sprintf("%s\n\n---\n\n%s", template, configuredBody)
}

