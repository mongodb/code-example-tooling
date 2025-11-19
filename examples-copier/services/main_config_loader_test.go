package services_test

import (
	"context"
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mongodb/code-example-tooling/code-copier/configs"
	"github.com/mongodb/code-example-tooling/code-copier/services"
	test "github.com/mongodb/code-example-tooling/code-copier/tests"
)

// Helper to encode YAML content to base64 for main config tests
func b64MainConfig(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}

func TestMainConfigLoader_LoadMainConfigFromContent_InlineWorkflows(t *testing.T) {
	loader := services.NewMainConfigLoader().(*services.DefaultMainConfigLoader)
	ctx := context.Background()

	mainConfigYAML := `
defaults:
  commit_strategy:
    type: "pull_request"
    auto_merge: false
  exclude:
    - "**/.env"
    - "**/node_modules/**"

workflow_configs:
  - source: "inline"
    workflows:
      - name: "test-workflow-1"
        source:
          repo: "mongodb/source-repo"
          branch: "main"
        destination:
          repo: "mongodb/dest-repo"
          branch: "main"
        transformations:
          - move:
              from: "examples"
              to: "code-examples"
      - name: "test-workflow-2"
        source:
          repo: "mongodb/source-repo-2"
          branch: "main"
        destination:
          repo: "mongodb/dest-repo-2"
          branch: "main"
        transformations:
          - move:
              from: "src"
              to: "dest"
`

	config := &configs.Config{
		ConfigRepoOwner:  "mongodb",
		ConfigRepoName:   "config-repo",
		ConfigRepoBranch: "main",
		MainConfigFile:   "main-config.yaml",
	}

	yamlConfig, err := loader.LoadMainConfigFromContent(ctx, mainConfigYAML, config)
	require.NoError(t, err)
	require.NotNil(t, yamlConfig)

	// Check that workflows were loaded
	assert.Len(t, yamlConfig.Workflows, 2)
	assert.Equal(t, "test-workflow-1", yamlConfig.Workflows[0].Name)
	assert.Equal(t, "test-workflow-2", yamlConfig.Workflows[1].Name)

	// Check that defaults were applied
	assert.NotNil(t, yamlConfig.Defaults)
	assert.NotNil(t, yamlConfig.Defaults.CommitStrategy)
	assert.Equal(t, "pull_request", yamlConfig.Defaults.CommitStrategy.Type)
	assert.False(t, yamlConfig.Defaults.CommitStrategy.AutoMerge)

	// Check exclude patterns
	assert.Len(t, yamlConfig.Defaults.Exclude, 2)
	assert.Contains(t, yamlConfig.Defaults.Exclude, "**/.env")
	assert.Contains(t, yamlConfig.Defaults.Exclude, "**/node_modules/**")
}

func TestMainConfigLoader_LoadMainConfigFromContent_LocalWorkflows(t *testing.T) {
	_ = test.WithHTTPMock(t)
	loader := services.NewMainConfigLoader().(*services.DefaultMainConfigLoader)
	ctx := context.Background()

	// Setup org token to bypass GitHub App authentication
	test.SetupOrgToken("mongodb", "test-token")

	// Mock the local workflow config file
	workflowConfigYAML := `
workflows:
  - name: "local-workflow"
    source:
      repo: "mongodb/source-repo"
      branch: "main"
    destination:
      repo: "mongodb/dest-repo"
      branch: "main"
    transformations:
      - move:
          from: "examples"
          to: "code-examples"
`
	test.MockContentsEndpoint("mongodb", "config-repo", "workflows/test-workflows.yaml", b64MainConfig(workflowConfigYAML))

	mainConfigYAML := `
defaults:
  commit_strategy:
    type: "direct"

workflow_configs:
  - source: "local"
    path: "workflows/test-workflows.yaml"
`

	config := &configs.Config{
		ConfigRepoOwner:  "mongodb",
		ConfigRepoName:   "config-repo",
		ConfigRepoBranch: "main",
		MainConfigFile:   "main-config.yaml",
	}

	yamlConfig, err := loader.LoadMainConfigFromContent(ctx, mainConfigYAML, config)
	require.NoError(t, err)
	require.NotNil(t, yamlConfig)

	// Check that workflow was loaded
	assert.Len(t, yamlConfig.Workflows, 1)
	assert.Equal(t, "local-workflow", yamlConfig.Workflows[0].Name)

	// Check that defaults were applied
	assert.NotNil(t, yamlConfig.Defaults)
	assert.NotNil(t, yamlConfig.Defaults.CommitStrategy)
	assert.Equal(t, "direct", yamlConfig.Defaults.CommitStrategy.Type)
}

func TestMainConfigLoader_LoadMainConfigFromContent_RemoteWorkflows(t *testing.T) {
	_ = test.WithHTTPMock(t)
	loader := services.NewMainConfigLoader().(*services.DefaultMainConfigLoader)
	ctx := context.Background()

	// Setup org token to bypass GitHub App authentication
	test.SetupOrgToken("mongodb", "test-token")

	// Mock the remote workflow config file
	workflowConfigYAML := `
defaults:
  commit_strategy:
    type: "pull_request"
    auto_merge: true

workflows:
  - name: "remote-workflow"
    source:
      repo: "mongodb/source-repo"
      branch: "main"
    destination:
      repo: "mongodb/dest-repo"
      branch: "main"
    transformations:
      - move:
          from: "src"
          to: "dest"
`
	test.MockContentsEndpoint("mongodb", "source-repo", ".copier/workflows.yaml", b64MainConfig(workflowConfigYAML))

	mainConfigYAML := `
defaults:
  commit_strategy:
    type: "direct"
  exclude:
    - "**/.env"

workflow_configs:
  - source: "repo"
    repo: "mongodb/source-repo"
    branch: "main"
    path: ".copier/workflows.yaml"
`

	config := &configs.Config{
		ConfigRepoOwner:  "mongodb",
		ConfigRepoName:   "config-repo",
		ConfigRepoBranch: "main",
		MainConfigFile:   "main-config.yaml",
	}

	yamlConfig, err := loader.LoadMainConfigFromContent(ctx, mainConfigYAML, config)
	require.NoError(t, err)
	require.NotNil(t, yamlConfig)

	// Check that workflow was loaded
	assert.Len(t, yamlConfig.Workflows, 1)
	assert.Equal(t, "remote-workflow", yamlConfig.Workflows[0].Name)

	// Check that main config defaults were applied
	assert.NotNil(t, yamlConfig.Defaults)
	assert.Len(t, yamlConfig.Defaults.Exclude, 1)
	assert.Contains(t, yamlConfig.Defaults.Exclude, "**/.env")

	// Check that workflow config defaults override main config defaults
	// The workflow should inherit the workflow config's auto_merge: true
	assert.NotNil(t, yamlConfig.Workflows[0].CommitStrategy)
	assert.Equal(t, "pull_request", yamlConfig.Workflows[0].CommitStrategy.Type)
	assert.True(t, yamlConfig.Workflows[0].CommitStrategy.AutoMerge)
}

func TestMainConfigLoader_LoadMainConfigFromContent_MixedSources(t *testing.T) {
	_ = test.WithHTTPMock(t)
	loader := services.NewMainConfigLoader().(*services.DefaultMainConfigLoader)
	ctx := context.Background()

	// Setup org token to bypass GitHub App authentication
	test.SetupOrgToken("mongodb", "test-token")

	// Mock local workflow config
	localWorkflowYAML := `
workflows:
  - name: "local-workflow"
    source:
      repo: "mongodb/source-1"
      branch: "main"
    destination:
      repo: "mongodb/dest-1"
      branch: "main"
    transformations:
      - move: { from: "a", to: "b" }
`
	test.MockContentsEndpoint("mongodb", "config-repo", "workflows/local.yaml", b64MainConfig(localWorkflowYAML))

	// Mock remote workflow config
	remoteWorkflowYAML := `
workflows:
  - name: "remote-workflow"
    source:
      repo: "mongodb/source-2"
      branch: "main"
    destination:
      repo: "mongodb/dest-2"
      branch: "main"
    transformations:
      - move: { from: "c", to: "d" }
`
	test.MockContentsEndpoint("mongodb", "source-repo", ".copier/workflows.yaml", b64MainConfig(remoteWorkflowYAML))

	mainConfigYAML := `
defaults:
  commit_strategy:
    type: "pull_request"

workflow_configs:
  - source: "inline"
    workflows:
      - name: "inline-workflow"
        source:
          repo: "mongodb/source-0"
          branch: "main"
        destination:
          repo: "mongodb/dest-0"
          branch: "main"
        transformations:
          - move: { from: "x", to: "y" }

  - source: "local"
    path: "workflows/local.yaml"

  - source: "repo"
    repo: "mongodb/source-repo"
    branch: "main"
    path: ".copier/workflows.yaml"
`

	config := &configs.Config{
		ConfigRepoOwner:  "mongodb",
		ConfigRepoName:   "config-repo",
		ConfigRepoBranch: "main",
		MainConfigFile:   "main-config.yaml",
	}

	yamlConfig, err := loader.LoadMainConfigFromContent(ctx, mainConfigYAML, config)
	require.NoError(t, err)
	require.NotNil(t, yamlConfig)

	// Check that all workflows were loaded and merged
	assert.Len(t, yamlConfig.Workflows, 3)

	// Verify workflow names in order
	workflowNames := []string{
		yamlConfig.Workflows[0].Name,
		yamlConfig.Workflows[1].Name,
		yamlConfig.Workflows[2].Name,
	}
	assert.Contains(t, workflowNames, "inline-workflow")
	assert.Contains(t, workflowNames, "local-workflow")
	assert.Contains(t, workflowNames, "remote-workflow")
}

func TestMainConfigLoader_LoadMainConfigFromContent_EmptyContent(t *testing.T) {
	loader := services.NewMainConfigLoader().(*services.DefaultMainConfigLoader)
	ctx := context.Background()

	config := &configs.Config{
		ConfigRepoOwner:  "mongodb",
		ConfigRepoName:   "config-repo",
		ConfigRepoBranch: "main",
		MainConfigFile:   "main-config.yaml",
	}

	_, err := loader.LoadMainConfigFromContent(ctx, "", config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty")
}

func TestMainConfigLoader_LoadMainConfigFromContent_InvalidYAML(t *testing.T) {
	loader := services.NewMainConfigLoader().(*services.DefaultMainConfigLoader)
	ctx := context.Background()

	invalidYAML := `
defaults:
  commit_strategy:
    type: "pull_request"
    auto_merge: false
  invalid_indent
workflow_configs:
`

	config := &configs.Config{
		ConfigRepoOwner:  "mongodb",
		ConfigRepoName:   "config-repo",
		ConfigRepoBranch: "main",
		MainConfigFile:   "main-config.yaml",
	}

	_, err := loader.LoadMainConfigFromContent(ctx, invalidYAML, config)
	assert.Error(t, err)
}

func TestMainConfigLoader_LoadMainConfigFromContent_NoWorkflowConfigs(t *testing.T) {
	loader := services.NewMainConfigLoader().(*services.DefaultMainConfigLoader)
	ctx := context.Background()

	// This should be treated as legacy format
	legacyYAML := `
workflows:
  - name: "legacy-workflow"
    source:
      repo: "mongodb/source-repo"
      branch: "main"
    destination:
      repo: "mongodb/dest-repo"
      branch: "main"
    transformations:
      - move:
          from: "examples"
          to: "code-examples"
`

	config := &configs.Config{
		ConfigRepoOwner:  "mongodb",
		ConfigRepoName:   "config-repo",
		ConfigRepoBranch: "main",
		ConfigFile:       "copier-config.yaml",
	}

	yamlConfig, err := loader.LoadMainConfigFromContent(ctx, legacyYAML, config)
	require.NoError(t, err)
	require.NotNil(t, yamlConfig)

	// Should fall back to legacy format
	assert.Len(t, yamlConfig.Workflows, 1)
	assert.Equal(t, "legacy-workflow", yamlConfig.Workflows[0].Name)
}

func TestMainConfigLoader_LoadMainConfigFromContent_InvalidWorkflowConfigRef(t *testing.T) {
	loader := services.NewMainConfigLoader().(*services.DefaultMainConfigLoader)
	ctx := context.Background()

	mainConfigYAML := `
workflow_configs:
  - source: "invalid-source"
    path: "workflows/test.yaml"
`

	config := &configs.Config{
		ConfigRepoOwner:  "mongodb",
		ConfigRepoName:   "config-repo",
		ConfigRepoBranch: "main",
		MainConfigFile:   "main-config.yaml",
	}

	_, err := loader.LoadMainConfigFromContent(ctx, mainConfigYAML, config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid source")
}

func TestMainConfigLoader_LoadMainConfigFromContent_DefaultPrecedence(t *testing.T) {
	_ = test.WithHTTPMock(t)
	loader := services.NewMainConfigLoader().(*services.DefaultMainConfigLoader)
	ctx := context.Background()

	// Setup org token to bypass GitHub App authentication
	test.SetupOrgToken("mongodb", "test-token")

	// Workflow config with its own defaults
	workflowConfigYAML := `
defaults:
  commit_strategy:
    type: "pull_request"
    auto_merge: true
    pr_title: "Workflow Config Default Title"
  exclude:
    - "**/workflow-exclude/**"

workflows:
  - name: "workflow-with-override"
    source:
      repo: "mongodb/source-repo"
      branch: "main"
    destination:
      repo: "mongodb/dest-repo"
      branch: "main"
    transformations:
      - move: { from: "src", to: "dest" }
    commit_strategy:
      pr_title: "Workflow Specific Title"

  - name: "workflow-without-override"
    source:
      repo: "mongodb/source-repo-2"
      branch: "main"
    destination:
      repo: "mongodb/dest-repo-2"
      branch: "main"
    transformations:
      - move: { from: "a", to: "b" }
`
	test.MockContentsEndpoint("mongodb", "source-repo", ".copier/workflows.yaml", b64MainConfig(workflowConfigYAML))

	mainConfigYAML := `
defaults:
  commit_strategy:
    type: "direct"
    commit_message: "Main Config Default Message"
  exclude:
    - "**/.env"
    - "**/main-exclude/**"

workflow_configs:
  - source: "repo"
    repo: "mongodb/source-repo"
    branch: "main"
    path: ".copier/workflows.yaml"
`

	config := &configs.Config{
		ConfigRepoOwner:  "mongodb",
		ConfigRepoName:   "config-repo",
		ConfigRepoBranch: "main",
		MainConfigFile:   "main-config.yaml",
	}

	yamlConfig, err := loader.LoadMainConfigFromContent(ctx, mainConfigYAML, config)
	require.NoError(t, err)
	require.NotNil(t, yamlConfig)

	assert.Len(t, yamlConfig.Workflows, 2)

	// First workflow should have workflow-specific title
	workflow1 := yamlConfig.Workflows[0]
	assert.Equal(t, "workflow-with-override", workflow1.Name)
	assert.NotNil(t, workflow1.CommitStrategy)
	assert.Equal(t, "Workflow Specific Title", workflow1.CommitStrategy.PRTitle)
	// When commit_strategy is specified at workflow level, it replaces the entire object
	// So AutoMerge will be false (default) since it wasn't specified in the workflow-level override
	assert.False(t, workflow1.CommitStrategy.AutoMerge)

	// Second workflow should inherit workflow config defaults
	workflow2 := yamlConfig.Workflows[1]
	assert.Equal(t, "workflow-without-override", workflow2.Name)
	assert.NotNil(t, workflow2.CommitStrategy)
	assert.Equal(t, "Workflow Config Default Title", workflow2.CommitStrategy.PRTitle)
	assert.True(t, workflow2.CommitStrategy.AutoMerge)

	// Main config defaults should be present
	assert.NotNil(t, yamlConfig.Defaults)
	assert.Len(t, yamlConfig.Defaults.Exclude, 2)
	assert.Contains(t, yamlConfig.Defaults.Exclude, "**/.env")
	assert.Contains(t, yamlConfig.Defaults.Exclude, "**/main-exclude/**")
}

func TestMainConfigLoader_LoadMainConfigFromContent_MultipleRemoteRepos(t *testing.T) {
	_ = test.WithHTTPMock(t)
	loader := services.NewMainConfigLoader().(*services.DefaultMainConfigLoader)
	ctx := context.Background()

	// Setup org tokens to bypass GitHub App authentication
	test.SetupOrgToken("mongodb", "test-token")
	test.SetupOrgToken("10gen", "test-token")

	// Mock workflow configs from different repos
	workflow1YAML := `
workflows:
  - name: "repo1-workflow"
    source:
      repo: "mongodb/source-1"
      branch: "main"
    destination:
      repo: "mongodb/dest-1"
      branch: "main"
    transformations:
      - move: { from: "a", to: "b" }
`
	test.MockContentsEndpoint("mongodb", "repo-1", ".copier/workflows.yaml", b64MainConfig(workflow1YAML))

	workflow2YAML := `
workflows:
  - name: "repo2-workflow"
    source:
      repo: "mongodb/source-2"
      branch: "main"
    destination:
      repo: "mongodb/dest-2"
      branch: "main"
    transformations:
      - move: { from: "c", to: "d" }
`
	test.MockContentsEndpoint("10gen", "repo-2", ".copier/workflows.yaml", b64MainConfig(workflow2YAML))

	mainConfigYAML := `
workflow_configs:
  - source: "repo"
    repo: "mongodb/repo-1"
    branch: "main"
    path: ".copier/workflows.yaml"

  - source: "repo"
    repo: "10gen/repo-2"
    branch: "main"
    path: ".copier/workflows.yaml"
`

	config := &configs.Config{
		ConfigRepoOwner:  "mongodb",
		ConfigRepoName:   "config-repo",
		ConfigRepoBranch: "main",
		MainConfigFile:   "main-config.yaml",
	}

	yamlConfig, err := loader.LoadMainConfigFromContent(ctx, mainConfigYAML, config)
	require.NoError(t, err)
	require.NotNil(t, yamlConfig)

	// Check that workflows from both repos were loaded
	assert.Len(t, yamlConfig.Workflows, 2)

	workflowNames := []string{
		yamlConfig.Workflows[0].Name,
		yamlConfig.Workflows[1].Name,
	}
	assert.Contains(t, workflowNames, "repo1-workflow")
	assert.Contains(t, workflowNames, "repo2-workflow")
}

func TestMainConfigLoader_LoadMainConfigFromContent_DisabledWorkflowConfig(t *testing.T) {
	_ = test.WithHTTPMock(t)
	loader := services.NewMainConfigLoader().(*services.DefaultMainConfigLoader)
	ctx := context.Background()

	// Setup org token to bypass GitHub App authentication
	test.SetupOrgToken("mongodb", "test-token")

	// Mock the workflow config file for enabled repo
	test.MockContentsEndpoint("mongodb", "enabled-repo", ".copier/workflows.yaml", b64MainConfig(`
defaults:
  commit_strategy:
    type: "pull_request"

workflows:
  - name: "enabled-workflow"
    source:
      repo: "mongodb/enabled-repo"
      branch: "main"
      path: "examples/"
    destination:
      repo: "mongodb/dest-repo"
      branch: "main"
      path: "examples/"
    transformations:
      - copy:
          from: "examples/"
          to: "examples/"
`))

	// Mock the workflow config file for disabled repo (should not be fetched)
	test.MockContentsEndpoint("mongodb", "disabled-repo", ".copier/workflows.yaml", b64MainConfig(`
defaults:
  commit_strategy:
    type: "pull_request"

workflows:
  - name: "disabled-workflow"
    source:
      repo: "mongodb/disabled-repo"
      branch: "main"
      path: "examples/"
    destination:
      repo: "mongodb/dest-repo"
      branch: "main"
      path: "examples/"
    transformations:
      - copy:
          from: "examples/"
          to: "examples/"
`))

	mainConfigYAML := `
defaults:
  commit_strategy:
    type: "pull_request"

workflow_configs:
  - source: "repo"
    repo: "mongodb/enabled-repo"
    branch: "main"
    path: ".copier/workflows.yaml"
    enabled: true

  - source: "repo"
    repo: "mongodb/disabled-repo"
    branch: "main"
    path: ".copier/workflows.yaml"
    enabled: false

  - source: "inline"
    enabled: false
    workflows:
      - name: "disabled-inline-workflow"
        source:
          repo: "mongodb/source"
          branch: "main"
          path: "examples/"
        destination:
          repo: "mongodb/dest"
          branch: "main"
          path: "examples/"
        transformations:
          - copy:
              from: "examples/"
              to: "examples/"
`

	config := &configs.Config{
		ConfigRepoOwner:  "mongodb",
		ConfigRepoName:   "config-repo",
		ConfigRepoBranch: "main",
		ConfigFile:       "main-config.yaml",
	}

	yamlConfig, err := loader.LoadMainConfigFromContent(ctx, mainConfigYAML, config)
	require.NoError(t, err)
	require.NotNil(t, yamlConfig)

	// Should only have 1 workflow from the enabled repo
	assert.Len(t, yamlConfig.Workflows, 1)
	assert.Equal(t, "enabled-workflow", yamlConfig.Workflows[0].Name)
	assert.Equal(t, "mongodb/enabled-repo", yamlConfig.Workflows[0].Source.Repo)
}

