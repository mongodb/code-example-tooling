package services

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/google/go-github/v48/github"
	"gopkg.in/yaml.v3"

	"github.com/mongodb/code-example-tooling/code-copier/configs"
	"github.com/mongodb/code-example-tooling/code-copier/types"
)

// DefaultMainConfigLoader implements the ConfigLoader interface with main config support
type DefaultMainConfigLoader struct {
	configLoader ConfigLoader
}

// NewMainConfigLoader creates a new main config loader
func NewMainConfigLoader() ConfigLoader {
	return &DefaultMainConfigLoader{
		configLoader: NewConfigLoader(),
	}
}

// LoadConfig implements the ConfigLoader interface
// It delegates to LoadMainConfig for the main config format
func (mcl *DefaultMainConfigLoader) LoadConfig(ctx context.Context, config *configs.Config) (*types.YAMLConfig, error) {
	return mcl.LoadMainConfig(ctx, config)
}

// LoadConfigFromContent implements the ConfigLoader interface
// It delegates to LoadMainConfigFromContent for the main config format
func (mcl *DefaultMainConfigLoader) LoadConfigFromContent(content string, filename string) (*types.YAMLConfig, error) {
	// Create a minimal config for parsing
	config := &configs.Config{
		ConfigFile: filename,
	}
	return mcl.LoadMainConfigFromContent(context.Background(), content, config)
}

// LoadMainConfig loads the main configuration and resolves all workflow references
func (mcl *DefaultMainConfigLoader) LoadMainConfig(ctx context.Context, config *configs.Config) (*types.YAMLConfig, error) {
	var content string
	var err error

	// Determine which config file to load
	configFile := config.ConfigFile
	if config.MainConfigFile != "" {
		configFile = config.MainConfigFile
	}

	// Try to load from local file first (for testing)
	content, err = loadLocalConfigFile(configFile)
	if err == nil {
		LogInfoCtx(ctx, "loaded main config from local file", map[string]interface{}{
			"file": configFile,
		})
	} else {
		// Fall back to fetching from repository
		content, err = retrieveConfigFileContent(ctx, configFile, config)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve main config file: %w", err)
		}
	}

	return mcl.LoadMainConfigFromContent(ctx, content, config)
}

// LoadMainConfigFromContent loads main configuration from a string and resolves references
func (mcl *DefaultMainConfigLoader) LoadMainConfigFromContent(ctx context.Context, content string, config *configs.Config) (*types.YAMLConfig, error) {
	if content == "" {
		return nil, fmt.Errorf("main config file is empty")
	}

	// Parse as MainConfig
	var mainConfig types.MainConfig
	err := yaml.Unmarshal([]byte(content), &mainConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to parse main config: %w", err)
	}

	// Validate that workflow_configs is present
	if len(mainConfig.WorkflowConfigs) == 0 {
		return nil, fmt.Errorf("main config must have at least one workflow_config entry")
	}

	// Set defaults for main config
	mainConfig.SetDefaults()

	// Validate main config
	if err := mainConfig.Validate(); err != nil {
		return nil, fmt.Errorf("main config validation failed: %w", err)
	}

	LogInfoCtx(ctx, "loaded main config with workflow references", map[string]interface{}{
		"workflow_config_count": len(mainConfig.WorkflowConfigs),
	})

	// Resolve all workflow references and merge into a single YAMLConfig
	return mcl.resolveWorkflowReferences(ctx, &mainConfig, config)
}

// resolveWorkflowReferences resolves all workflow config references and merges them
func (mcl *DefaultMainConfigLoader) resolveWorkflowReferences(ctx context.Context, mainConfig *types.MainConfig, config *configs.Config) (*types.YAMLConfig, error) {
	mergedConfig := &types.YAMLConfig{
		Defaults:  mainConfig.Defaults,
		Workflows: []types.Workflow{},
	}

	// Process each workflow config reference
	for i, ref := range mainConfig.WorkflowConfigs {
		// Skip disabled workflow configs
		if ref.Enabled != nil && !*ref.Enabled {
			LogInfoCtx(ctx, "skipping disabled workflow config reference", map[string]interface{}{
				"index":  i,
				"source": ref.Source,
				"path":   ref.Path,
				"repo":   ref.Repo,
			})
			continue
		}

		LogInfoCtx(ctx, "resolving workflow config reference", map[string]interface{}{
			"index":  i,
			"source": ref.Source,
			"path":   ref.Path,
			"repo":   ref.Repo,
		})

		workflowConfig, err := mcl.loadWorkflowConfig(ctx, &ref, config)
		if err != nil {
			// Log warning and continue instead of failing completely
			// This allows the app to process other workflow configs even if one is missing
			LogWarningCtx(ctx, "failed to load workflow config, skipping", map[string]interface{}{
				"index":  i,
				"source": ref.Source,
				"path":   ref.Path,
				"repo":   ref.Repo,
				"error":  err.Error(),
			})
			continue
		}

		// Apply source context (allows workflows to omit source.repo/branch)
		workflowConfig.ApplySourceContext()

		// Apply global defaults to workflow config
		workflowConfig.ApplyGlobalDefaults(mainConfig.Defaults)

		// Set defaults for workflow config
		workflowConfig.SetDefaults()

		// Validate workflow config
		if err := workflowConfig.Validate(); err != nil {
			// Log warning and continue instead of failing completely
			LogWarningCtx(ctx, "workflow config validation failed, skipping", map[string]interface{}{
				"index": i,
				"error": err.Error(),
			})
			continue
		}

		// Merge workflows into the main config
		mergedConfig.Workflows = append(mergedConfig.Workflows, workflowConfig.Workflows...)

		LogInfoCtx(ctx, "resolved workflow config reference", map[string]interface{}{
			"index":          i,
			"workflow_count": len(workflowConfig.Workflows),
		})
	}

	// Validate merged config
	if err := mergedConfig.Validate(); err != nil {
		return nil, fmt.Errorf("merged config validation failed: %w", err)
	}

	LogInfoCtx(ctx, "successfully resolved all workflow references", map[string]interface{}{
		"total_workflows": len(mergedConfig.Workflows),
	})

	return mergedConfig, nil
}

// loadWorkflowConfig loads a workflow config based on the reference type
func (mcl *DefaultMainConfigLoader) loadWorkflowConfig(ctx context.Context, ref *types.WorkflowConfigRef, config *configs.Config) (*types.WorkflowConfig, error) {
	switch ref.Source {
	case "inline":
		// Inline workflows - already in the reference
		return &types.WorkflowConfig{
			Workflows: ref.Workflows,
		}, nil

	case "local":
		// Local file in the same repo as main config
		return mcl.loadLocalWorkflowConfig(ctx, ref, config)

	case "repo":
		// Remote file in a different repo
		return mcl.loadRemoteWorkflowConfig(ctx, ref)

	default:
		return nil, fmt.Errorf("unsupported workflow config source: %s", ref.Source)
	}
}

// loadLocalWorkflowConfig loads a workflow config from the same repo as main config
func (mcl *DefaultMainConfigLoader) loadLocalWorkflowConfig(ctx context.Context, ref *types.WorkflowConfigRef, config *configs.Config) (*types.WorkflowConfig, error) {
	// Try local file first
	content, err := loadLocalConfigFile(ref.Path)
	if err == nil {
		LogInfoCtx(ctx, "loaded workflow config from local file", map[string]interface{}{
			"path": ref.Path,
		})
		workflowConfig, err := mcl.parseWorkflowConfig(content, ref.Path)
		if err != nil {
			return nil, err
		}

		// Resolve $ref references
		baseRepo := fmt.Sprintf("%s/%s", config.ConfigRepoOwner, config.ConfigRepoName)
		if err := mcl.resolveWorkflowFieldReferences(ctx, workflowConfig, baseRepo, config.ConfigRepoBranch, ref.Path); err != nil {
			return nil, err
		}

		return workflowConfig, nil
	}

	// Fall back to fetching from config repo
	client, err := GetRestClientForOrg(config.ConfigRepoOwner)
	if err != nil {
		return nil, fmt.Errorf("failed to get GitHub client for org %s: %w", config.ConfigRepoOwner, err)
	}

	fileContent, _, _, err := client.Repositories.GetContents(
		ctx,
		config.ConfigRepoOwner,
		config.ConfigRepoName,
		ref.Path,
		&github.RepositoryContentGetOptions{
			Ref: config.ConfigRepoBranch,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow config file: %w", err)
	}

	content, err = fileContent.GetContent()
	if err != nil {
		return nil, fmt.Errorf("failed to decode workflow config file: %w", err)
	}

	workflowConfig, err := mcl.parseWorkflowConfig(content, ref.Path)
	if err != nil {
		return nil, err
	}

	// Resolve $ref references
	baseRepo := fmt.Sprintf("%s/%s", config.ConfigRepoOwner, config.ConfigRepoName)
	if err := mcl.resolveWorkflowFieldReferences(ctx, workflowConfig, baseRepo, config.ConfigRepoBranch, ref.Path); err != nil {
		return nil, err
	}

	return workflowConfig, nil
}

// loadRemoteWorkflowConfig loads a workflow config from a different repo
func (mcl *DefaultMainConfigLoader) loadRemoteWorkflowConfig(ctx context.Context, ref *types.WorkflowConfigRef) (*types.WorkflowConfig, error) {
	// Parse repo owner and name
	parts := strings.Split(ref.Repo, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid repo format: %s (expected owner/repo)", ref.Repo)
	}
	owner := parts[0]
	repo := parts[1]

	// Get GitHub client for the repo's org
	client, err := GetRestClientForOrg(owner)
	if err != nil {
		return nil, fmt.Errorf("failed to get GitHub client for org %s: %w", owner, err)
	}

	// Fetch file content
	fileContent, _, _, err := client.Repositories.GetContents(
		ctx,
		owner,
		repo,
		ref.Path,
		&github.RepositoryContentGetOptions{
			Ref: ref.Branch,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow config file from %s: %w", ref.Repo, err)
	}

	content, err := fileContent.GetContent()
	if err != nil {
		return nil, fmt.Errorf("failed to decode workflow config file: %w", err)
	}

	LogInfoCtx(ctx, "loaded workflow config from remote repo", map[string]interface{}{
		"repo":   ref.Repo,
		"branch": ref.Branch,
		"path":   ref.Path,
	})

	workflowConfig, err := mcl.parseWorkflowConfig(content, ref.Path)
	if err != nil {
		return nil, err
	}

	// Set source context so workflows can omit source.repo and source.branch
	workflowConfig.SourceRepo = ref.Repo
	workflowConfig.SourceBranch = ref.Branch

	// Resolve $ref references
	if err := mcl.resolveWorkflowFieldReferences(ctx, workflowConfig, ref.Repo, ref.Branch, ref.Path); err != nil {
		return nil, err
	}

	return workflowConfig, nil
}

// parseWorkflowConfig parses a workflow config from content
func (mcl *DefaultMainConfigLoader) parseWorkflowConfig(content string, filename string) (*types.WorkflowConfig, error) {
	if content == "" {
		return nil, fmt.Errorf("workflow config file is empty")
	}

	var workflowConfig types.WorkflowConfig
	err := yaml.Unmarshal([]byte(content), &workflowConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to parse workflow config file %s: %w", filename, err)
	}

	return &workflowConfig, nil
}

// resolveWorkflowFieldReferences resolves all $ref references in workflow fields (transformations, exclude, commit_strategy)
func (mcl *DefaultMainConfigLoader) resolveWorkflowFieldReferences(ctx context.Context, workflowConfig *types.WorkflowConfig, baseRepo string, baseBranch string, basePath string) error {
	for i := range workflowConfig.Workflows {
		workflow := &workflowConfig.Workflows[i]

		// Resolve transformations $ref
		if workflow.TransformationsRef != "" {
			LogInfoCtx(ctx, "resolving transformations $ref", map[string]interface{}{
				"workflow": workflow.Name,
				"ref":      workflow.TransformationsRef,
			})

			content, err := mcl.resolveReference(ctx, workflow.TransformationsRef, baseRepo, baseBranch, basePath)
			if err != nil {
				return fmt.Errorf("failed to resolve transformations $ref for workflow %s: %w", workflow.Name, err)
			}

			var transformations []types.Transformation
			if err := yaml.Unmarshal([]byte(content), &transformations); err != nil {
				return fmt.Errorf("failed to parse transformations from $ref for workflow %s: %w", workflow.Name, err)
			}
			workflow.Transformations = transformations
			workflow.TransformationsRef = "" // Clear the ref after resolution
		}

		// Resolve exclude $ref
		if workflow.ExcludeRef != "" {
			LogInfoCtx(ctx, "resolving exclude $ref", map[string]interface{}{
				"workflow": workflow.Name,
				"ref":      workflow.ExcludeRef,
			})

			content, err := mcl.resolveReference(ctx, workflow.ExcludeRef, baseRepo, baseBranch, basePath)
			if err != nil {
				return fmt.Errorf("failed to resolve exclude $ref for workflow %s: %w", workflow.Name, err)
			}

			var exclude []string
			if err := yaml.Unmarshal([]byte(content), &exclude); err != nil {
				return fmt.Errorf("failed to parse exclude from $ref for workflow %s: %w", workflow.Name, err)
			}
			workflow.Exclude = exclude
			workflow.ExcludeRef = "" // Clear the ref after resolution
		}

		// Resolve commit_strategy $ref
		if workflow.CommitStrategyRef != "" {
			LogInfoCtx(ctx, "resolving commit_strategy $ref", map[string]interface{}{
				"workflow": workflow.Name,
				"ref":      workflow.CommitStrategyRef,
			})

			content, err := mcl.resolveReference(ctx, workflow.CommitStrategyRef, baseRepo, baseBranch, basePath)
			if err != nil {
				return fmt.Errorf("failed to resolve commit_strategy $ref for workflow %s: %w", workflow.Name, err)
			}

			var strategy types.CommitStrategyConfig
			if err := yaml.Unmarshal([]byte(content), &strategy); err != nil {
				return fmt.Errorf("failed to parse commit_strategy from $ref for workflow %s: %w", workflow.Name, err)
			}
			workflow.CommitStrategy = &strategy
			workflow.CommitStrategyRef = "" // Clear the ref after resolution
		}
	}

	return nil
}

// resolveReference resolves a $ref reference to actual content
// This supports references in transformations, commit strategies, etc.
func (mcl *DefaultMainConfigLoader) resolveReference(ctx context.Context, ref string, baseRepo string, baseBranch string, basePath string) (string, error) {
	// Parse reference format
	// Supports:
	// - Relative paths: "strategies/pr-strategy.yaml"
	// - Repo references: "repo://owner/repo/path/to/file.yaml@branch"
	
	if strings.HasPrefix(ref, "repo://") {
		// Remote repo reference
		return mcl.resolveRemoteReference(ctx, ref)
	}

	// Relative path reference
	return mcl.resolveRelativeReference(ctx, ref, baseRepo, baseBranch, basePath)
}

// resolveRemoteReference resolves a repo:// reference
func (mcl *DefaultMainConfigLoader) resolveRemoteReference(ctx context.Context, ref string) (string, error) {
	// Parse: repo://owner/repo/path/to/file.yaml@branch
	ref = strings.TrimPrefix(ref, "repo://")
	
	// Split by @ to get branch
	parts := strings.Split(ref, "@")
	branch := "main"
	if len(parts) == 2 {
		ref = parts[0]
		branch = parts[1]
	}

	// Split path to get owner/repo and file path
	pathParts := strings.SplitN(ref, "/", 3)
	if len(pathParts) < 3 {
		return "", fmt.Errorf("invalid repo reference format: %s", ref)
	}

	owner := pathParts[0]
	repo := pathParts[1]
	filePath := pathParts[2]

	// Fetch file content
	client, err := GetRestClientForOrg(owner)
	if err != nil {
		return "", fmt.Errorf("failed to get GitHub client for org %s: %w", owner, err)
	}

	fileContent, _, _, err := client.Repositories.GetContents(
		ctx,
		owner,
		repo,
		filePath,
		&github.RepositoryContentGetOptions{
			Ref: branch,
		},
	)
	if err != nil {
		return "", fmt.Errorf("failed to get referenced file: %w", err)
	}

	return fileContent.GetContent()
}

// resolveRelativeReference resolves a relative path reference
func (mcl *DefaultMainConfigLoader) resolveRelativeReference(ctx context.Context, ref string, baseRepo string, baseBranch string, basePath string) (string, error) {
	// Resolve relative to base path
	baseDir := filepath.Dir(basePath)
	resolvedPath := filepath.Join(baseDir, ref)

	// Parse repo
	parts := strings.Split(baseRepo, "/")
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid base repo format: %s", baseRepo)
	}
	owner := parts[0]
	repo := parts[1]

	// Fetch file content
	client, err := GetRestClientForOrg(owner)
	if err != nil {
		return "", fmt.Errorf("failed to get GitHub client for org %s: %w", owner, err)
	}

	fileContent, _, _, err := client.Repositories.GetContents(
		ctx,
		owner,
		repo,
		resolvedPath,
		&github.RepositoryContentGetOptions{
			Ref: baseBranch,
		},
	)
	if err != nil {
		return "", fmt.Errorf("failed to get referenced file: %w", err)
	}

	return fileContent.GetContent()
}

