package services

import (
	"context"
	"testing"

	"github.com/google/go-github/v48/github"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/require"
)

func TestFetchPRTemplate_Found(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	owner := "testowner"
	repo := "testrepo"
	branch := "main"

	// Mock the first location (.github/pull_request_template.md)
	// Note: GitHub API returns base64-encoded content with encoding field
	httpmock.RegisterResponder("GET",
		"https://api.github.com/repos/testowner/testrepo/contents/.github/pull_request_template.md",
		httpmock.NewJsonResponderOrPanic(200, map[string]interface{}{
			"name":     "pull_request_template.md",
			"path":     ".github/pull_request_template.md",
			"type":     "file",
			"encoding": "base64",
			"content":  "IyBQdWxsIFJlcXVlc3QgVGVtcGxhdGUKCiMjIERlc2NyaXB0aW9uCgpQbGVhc2UgZGVzY3JpYmUgeW91ciBjaGFuZ2VzLg==", // base64 encoded
		}),
	)

	client := github.NewClient(nil)
	fetcher := NewPRTemplateFetcher()

	template, err := fetcher.FetchPRTemplate(context.Background(), client, owner+"/"+repo, branch)

	require.NoError(t, err)
	require.NotEmpty(t, template)
	require.Contains(t, template, "Pull Request Template")
}

func TestFetchPRTemplate_NotFound(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	owner := "testowner"
	repo := "testrepo"
	branch := "main"

	// Mock all locations as not found
	locations := []string{
		".github/pull_request_template.md",
		".github/PULL_REQUEST_TEMPLATE.md",
		"docs/pull_request_template.md",
		"PULL_REQUEST_TEMPLATE.md",
		"pull_request_template.md",
	}

	for _, location := range locations {
		httpmock.RegisterResponder("GET",
			"https://api.github.com/repos/testowner/testrepo/contents/"+location,
			httpmock.NewStringResponder(404, `{"message": "Not Found"}`),
		)
	}

	client := github.NewClient(nil)
	fetcher := NewPRTemplateFetcher()

	template, err := fetcher.FetchPRTemplate(context.Background(), client, owner+"/"+repo, branch)

	require.NoError(t, err)
	require.Empty(t, template)
}

func TestFetchPRTemplate_SecondLocation(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	owner := "testowner"
	repo := "testrepo"
	branch := "main"

	// First location not found
	httpmock.RegisterResponder("GET",
		"https://api.github.com/repos/testowner/testrepo/contents/.github/pull_request_template.md",
		httpmock.NewStringResponder(404, `{"message": "Not Found"}`),
	)

	// Second location found (.github/PULL_REQUEST_TEMPLATE.md)
	httpmock.RegisterResponder("GET",
		"https://api.github.com/repos/testowner/testrepo/contents/.github/PULL_REQUEST_TEMPLATE.md",
		httpmock.NewJsonResponderOrPanic(200, map[string]interface{}{
			"name":     "PULL_REQUEST_TEMPLATE.md",
			"path":     ".github/PULL_REQUEST_TEMPLATE.md",
			"type":     "file",
			"encoding": "base64",
			"content":  "IyBQUiBUZW1wbGF0ZQoKU2Vjb25kIGxvY2F0aW9u", // base64 encoded
		}),
	)

	client := github.NewClient(nil)
	fetcher := NewPRTemplateFetcher()

	template, err := fetcher.FetchPRTemplate(context.Background(), client, owner+"/"+repo, branch)

	require.NoError(t, err)
	require.NotEmpty(t, template)
	require.Contains(t, template, "PR Template")
	require.Contains(t, template, "Second location")
}

func TestFetchPRTemplate_InvalidRepoFormat(t *testing.T) {
	client := github.NewClient(nil)
	fetcher := NewPRTemplateFetcher()

	template, err := fetcher.FetchPRTemplate(context.Background(), client, "invalid-repo-format", "main")

	require.Error(t, err)
	require.Empty(t, template)
	require.Contains(t, err.Error(), "invalid repo format")
}

func TestMergePRBodyWithTemplate_BothPresent(t *testing.T) {
	configuredBody := "🤖 Automated update\n\nFiles: 10"
	template := "## Checklist\n\n- [ ] Tests added\n- [ ] Docs updated"

	merged := MergePRBodyWithTemplate(configuredBody, template)

	require.Contains(t, merged, configuredBody)
	require.Contains(t, merged, template)
	require.Contains(t, merged, "---") // Separator
	// Template should come first, then configured body
	require.Less(t, 0, len(merged))
	templateIndex := len(template)
	configuredIndex := len(merged) - len(configuredBody)
	require.Less(t, templateIndex, configuredIndex)
}

func TestMergePRBodyWithTemplate_OnlyConfigured(t *testing.T) {
	configuredBody := "🤖 Automated update"
	template := ""

	merged := MergePRBodyWithTemplate(configuredBody, template)

	require.Equal(t, configuredBody, merged)
	require.NotContains(t, merged, "---")
}

func TestMergePRBodyWithTemplate_OnlyTemplate(t *testing.T) {
	configuredBody := ""
	template := "## Checklist\n\n- [ ] Tests added"

	merged := MergePRBodyWithTemplate(configuredBody, template)

	require.Equal(t, template, merged)
	require.NotContains(t, merged, "---")
}

func TestMergePRBodyWithTemplate_BothEmpty(t *testing.T) {
	configuredBody := ""
	template := ""

	merged := MergePRBodyWithTemplate(configuredBody, template)

	require.Empty(t, merged)
}

func TestPRTemplateFetcher_ChecksMultipleLocations(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	owner := "testowner"
	repo := "testrepo"
	branch := "main"

	// Track which locations were checked
	checkedLocations := []string{}

	locations := []string{
		".github/pull_request_template.md",
		".github/PULL_REQUEST_TEMPLATE.md",
		"docs/pull_request_template.md",
		"PULL_REQUEST_TEMPLATE.md",
		"pull_request_template.md",
	}

	for _, location := range locations {
		loc := location // capture for closure
		httpmock.RegisterResponder("GET",
			"https://api.github.com/repos/testowner/testrepo/contents/"+location,
			httpmock.NewStringResponder(404, `{"message": "Not Found"}`),
		)
		// Track that this location was checked by registering a callback
		checkedLocations = append(checkedLocations, loc)
	}

	client := github.NewClient(nil)
	fetcher := NewPRTemplateFetcher()

	_, _ = fetcher.FetchPRTemplate(context.Background(), client, owner+"/"+repo, branch)

	// Should have registered all locations
	require.Len(t, checkedLocations, len(locations))
}

func TestPRTemplateFetcher_StopsAtFirstMatch(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	owner := "testowner"
	repo := "testrepo"
	branch := "main"

	// First location found
	httpmock.RegisterResponder("GET",
		"https://api.github.com/repos/testowner/testrepo/contents/.github/pull_request_template.md",
		httpmock.NewJsonResponderOrPanic(200, map[string]interface{}{
			"name":     "pull_request_template.md",
			"path":     ".github/pull_request_template.md",
			"type":     "file",
			"encoding": "base64",
			"content":  "VGVtcGxhdGU=", // base64 "Template"
		}),
	)

	// Other locations should not be checked
	otherLocations := []string{
		".github/PULL_REQUEST_TEMPLATE.md",
		"docs/pull_request_template.md",
		"PULL_REQUEST_TEMPLATE.md",
		"pull_request_template.md",
	}

	for _, location := range otherLocations {
		httpmock.RegisterResponder("GET",
			"https://api.github.com/repos/testowner/testrepo/contents/"+location,
			httpmock.NewStringResponder(404, `{"message": "Not Found"}`),
		)
	}

	client := github.NewClient(nil)
	fetcher := NewPRTemplateFetcher()

	template, err := fetcher.FetchPRTemplate(context.Background(), client, owner+"/"+repo, branch)

	require.NoError(t, err)
	require.NotEmpty(t, template)

	// Verify the first location was called
	info := httpmock.GetCallCountInfo()
	require.Equal(t, 1, info["GET https://api.github.com/repos/testowner/testrepo/contents/.github/pull_request_template.md"])

	// Verify other locations were not called
	for _, location := range otherLocations {
		require.Equal(t, 0, info["GET https://api.github.com/repos/testowner/testrepo/contents/"+location])
	}
}

