package services_test

import (
	"testing"

	"github.com/mongodb/code-example-tooling/code-copier/services"
	"github.com/mongodb/code-example-tooling/code-copier/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBatchPRConfig_LoadsCorrectly(t *testing.T) {
	loader := services.NewConfigLoader()

	yamlContent := `
source_repo: "org/source-repo"
source_branch: "main"
batch_by_repo: true

batch_pr_config:
  pr_title: "Custom batch PR title"
  pr_body: "Custom batch PR body with ${file_count} files"
  commit_message: "Batch commit from ${source_repo}"

copy_rules:
  - name: "test-rule"
    source_pattern:
      type: "prefix"
      pattern: "examples/"
    targets:
      - repo: "org/target-repo"
        branch: "main"
        path_transform: "docs/${relative_path}"
        commit_strategy:
          type: "pull_request"
`

	config, err := loader.LoadConfigFromContent(yamlContent, "config.yaml")
	require.NoError(t, err)
	require.NotNil(t, config)

	assert.True(t, config.BatchByRepo)
	require.NotNil(t, config.BatchPRConfig)
	assert.Equal(t, "Custom batch PR title", config.BatchPRConfig.PRTitle)
	assert.Equal(t, "Custom batch PR body with ${file_count} files", config.BatchPRConfig.PRBody)
	assert.Equal(t, "Batch commit from ${source_repo}", config.BatchPRConfig.CommitMessage)
}

func TestBatchPRConfig_OptionalField(t *testing.T) {
	loader := services.NewConfigLoader()

	yamlContent := `
source_repo: "org/source-repo"
source_branch: "main"
batch_by_repo: true

copy_rules:
  - name: "test-rule"
    source_pattern:
      type: "prefix"
      pattern: "examples/"
    targets:
      - repo: "org/target-repo"
        branch: "main"
        path_transform: "docs/${relative_path}"
        commit_strategy:
          type: "pull_request"
`

	config, err := loader.LoadConfigFromContent(yamlContent, "config.yaml")
	require.NoError(t, err)
	require.NotNil(t, config)

	assert.True(t, config.BatchByRepo)
	assert.Nil(t, config.BatchPRConfig) // Should be nil when not specified
}

func TestBatchPRConfig_StructureValidation(t *testing.T) {
	// Test that the BatchPRConfig struct is properly defined
	yamlConfig := &types.YAMLConfig{
		SourceRepo:   "owner/source-repo",
		SourceBranch: "main",
		BatchByRepo:  true,
		BatchPRConfig: &types.BatchPRConfig{
			PRTitle:       "Batch update from ${source_repo}",
			PRBody:        "Updated ${file_count} files from PR #${pr_number}",
			CommitMessage: "Batch commit",
		},
	}

	// Verify the config structure
	assert.NotNil(t, yamlConfig.BatchPRConfig)
	assert.Equal(t, "Batch update from ${source_repo}", yamlConfig.BatchPRConfig.PRTitle)
	assert.Equal(t, "Updated ${file_count} files from PR #${pr_number}", yamlConfig.BatchPRConfig.PRBody)
	assert.Equal(t, "Batch commit", yamlConfig.BatchPRConfig.CommitMessage)
}

func TestMessageTemplater_RendersFileCount(t *testing.T) {
	templater := services.NewMessageTemplater()

	ctx := types.NewMessageContext()
	ctx.SourceRepo = "owner/source-repo"
	ctx.FileCount = 42
	ctx.PRNumber = 123

	template := "Updated ${file_count} files from ${source_repo} PR #${pr_number}"
	result := templater.RenderPRBody(template, ctx)

	assert.Equal(t, "Updated 42 files from owner/source-repo PR #123", result)
}

func TestMessageTemplater_DefaultBatchPRBody(t *testing.T) {
	templater := services.NewMessageTemplater()

	ctx := types.NewMessageContext()
	ctx.SourceRepo = "owner/source-repo"
	ctx.FileCount = 15
	ctx.PRNumber = 456

	// Empty template should use default
	result := templater.RenderPRBody("", ctx)

	assert.Contains(t, result, "15 file(s)")
	assert.Contains(t, result, "owner/source-repo")
	assert.Contains(t, result, "#456")
}

