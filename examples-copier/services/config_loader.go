package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/google/go-github/v48/github"
	"gopkg.in/yaml.v3"

	"github.com/mongodb/code-example-tooling/code-copier/configs"
	"github.com/mongodb/code-example-tooling/code-copier/types"
)

// ConfigLoader handles loading and parsing configuration files
type ConfigLoader interface {
	LoadConfig(ctx context.Context, config *configs.Config) (*types.YAMLConfig, error)
	LoadConfigFromContent(content string, filename string) (*types.YAMLConfig, error)
}

// DefaultConfigLoader implements the ConfigLoader interface
type DefaultConfigLoader struct{}

// NewConfigLoader creates a new config loader
func NewConfigLoader() ConfigLoader {
	return &DefaultConfigLoader{}
}

// LoadConfig loads configuration from the repository or local file
func (cl *DefaultConfigLoader) LoadConfig(ctx context.Context, config *configs.Config) (*types.YAMLConfig, error) {
	var content string
	var err error

	// Try to load from local file first (for testing)
	content, err = loadLocalConfigFile(config.ConfigFile)
	if err == nil {
		LogInfoCtx(ctx, "loaded config from local file", map[string]interface{}{
			"file": config.ConfigFile,
		})
	} else {
		// Fall back to fetching from repository
		content, err = retrieveConfigFileContent(ctx, config.ConfigFile, config)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve config file: %w", err)
		}
	}

	return cl.LoadConfigFromContent(content, config.ConfigFile)
}

// LoadConfigFromContent loads configuration from a string
func (cl *DefaultConfigLoader) LoadConfigFromContent(content string, filename string) (*types.YAMLConfig, error) {
	if content == "" {
		return nil, fmt.Errorf("config file is empty")
	}

	// Parse as YAML (supports both YAML and JSON since YAML is a superset of JSON)
	var yamlConfig types.YAMLConfig
	err := yaml.Unmarshal([]byte(content), &yamlConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Set defaults
	yamlConfig.SetDefaults()

	// Validate
	if err := yamlConfig.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &yamlConfig, nil
}

// retrieveConfigFileContent fetches the config file content from the repository
func retrieveConfigFileContent(ctx context.Context, filePath string, config *configs.Config) (string, error) {
	// Get GitHub client for the config repo's org (auto-discovers installation ID)
	client, err := GetRestClientForOrg(config.ConfigRepoOwner)
	if err != nil {
		return "", fmt.Errorf("failed to get GitHub client for org %s: %w", config.ConfigRepoOwner, err)
	}

	// Fetch file content
	fileContent, _, _, err := client.Repositories.GetContents(
		ctx,
		config.ConfigRepoOwner,
		config.ConfigRepoName,
		filePath,
		&github.RepositoryContentGetOptions{
			Ref: config.ConfigRepoBranch,
		},
	)
	if err != nil {
		return "", fmt.Errorf("failed to get config file: %w", err)
	}

	// Decode content
	content, err := fileContent.GetContent()
	if err != nil {
		return "", fmt.Errorf("failed to decode config file: %w", err)
	}

	return content, nil
}

// ValidateConfig validates a YAML configuration
func ValidateConfig(config *types.YAMLConfig) error {
	return config.Validate()
}

// ConfigValidator provides validation utilities
type ConfigValidator struct{}

// NewConfigValidator creates a new config validator
func NewConfigValidator() *ConfigValidator {
	return &ConfigValidator{}
}

// ValidatePattern validates a pattern and returns any errors
func (cv *ConfigValidator) ValidatePattern(patternType types.PatternType, pattern string) error {
	sp := types.SourcePattern{
		Type:    patternType,
		Pattern: pattern,
	}
	return sp.Validate()
}

// TestPattern tests a pattern against a file path
func (cv *ConfigValidator) TestPattern(patternType types.PatternType, pattern string, filePath string) (types.MatchResult, error) {
	sp := types.SourcePattern{
		Type:    patternType,
		Pattern: pattern,
	}

	if err := sp.Validate(); err != nil {
		return types.NewMatchResult(false, nil), err
	}

	matcher := NewPatternMatcher()
	return matcher.Match(filePath, sp), nil
}

// TestTransform tests a path transformation
func (cv *ConfigValidator) TestTransform(sourcePath string, template string, variables map[string]string) (string, error) {
	transformer := NewPathTransformer()
	return transformer.Transform(sourcePath, template, variables)
}

// ExportConfigAsYAML exports a config as YAML string
func ExportConfigAsYAML(config *types.YAMLConfig) (string, error) {
	data, err := yaml.Marshal(config)
	if err != nil {
		return "", fmt.Errorf("failed to marshal config to YAML: %w", err)
	}
	return string(data), nil
}

// ExportConfigAsJSON exports a config as JSON string
func ExportConfigAsJSON(config *types.YAMLConfig) (string, error) {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal config to JSON: %w", err)
	}
	return string(data), nil
}

// ConfigTemplate represents a configuration template
type ConfigTemplate struct {
	Name        string
	Description string
	Content     string
}

// GetConfigTemplates returns available configuration templates
func GetConfigTemplates() []ConfigTemplate {
	return []ConfigTemplate{
		{
			Name:        "basic",
			Description: "Basic configuration with prefix pattern matching",
			Content: `source_repo: "owner/source-repo"
source_branch: "main"

copy_rules:
  - name: "example-rule"
    source_pattern:
      type: "prefix"
      pattern: "examples/"
    targets:
      - repo: "owner/target-repo"
        branch: "main"
        path_transform: "code-examples/${relative_path}"
        commit_strategy:
          type: "direct"
          commit_message: "Update code examples"
`,
		},
		{
			Name:        "glob",
			Description: "Configuration with glob pattern matching",
			Content: `source_repo: "owner/source-repo"
source_branch: "main"

copy_rules:
  - name: "go-examples"
    source_pattern:
      type: "glob"
      pattern: "examples/**/*.go"
    targets:
      - repo: "owner/target-repo"
        branch: "main"
        path_transform: "go-examples/${filename}"
        commit_strategy:
          type: "pull_request"
          pr_title: "Update Go examples"
          pr_body: "Automated update from source repository"
          auto_merge: false
`,
		},
		{
			Name:        "regex",
			Description: "Advanced configuration with regex pattern matching",
			Content: `source_repo: "owner/source-repo"
source_branch: "main"

copy_rules:
  - name: "language-examples"
    source_pattern:
      type: "regex"
      pattern: "^examples/(?P<lang>[^/]+)/(?P<category>[^/]+)/(?P<file>.+)$"
    targets:
      - repo: "owner/docs-repo"
        branch: "main"
        path_transform: "source/code-examples/${lang}/${category}/${file}"
        commit_strategy:
          type: "pull_request"
          pr_title: "Update ${lang} examples"
          pr_body: "Automated update of ${lang} examples from source repository"
        deprecation_check:
          enabled: true
          file: "deprecated_examples.json"
`,
		},
	}
}

// loadLocalConfigFile attempts to load config from a local file
// This is useful for local testing and development
func loadLocalConfigFile(filename string) (string, error) {
	// Try to read from current directory
	data, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
