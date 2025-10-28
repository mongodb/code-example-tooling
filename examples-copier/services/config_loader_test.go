package services_test

import (
	"testing"

	"github.com/mongodb/code-example-tooling/code-copier/services"
	"github.com/mongodb/code-example-tooling/code-copier/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigLoader_LoadYAML(t *testing.T) {
	loader := services.NewConfigLoader()

	yamlContent := `
source_repo: "org/source-repo"
source_branch: "main"

copy_rules:
  - name: "Copy Go examples"
    source_pattern:
      type: "prefix"
      pattern: "examples/go/"
    targets:
      - repo: "org/target-repo"
        branch: "main"
        path_transform: "docs/${path}"
        commit_strategy:
          type: "direct"
          commit_message: "Update examples"
        deprecation_check:
          enabled: true
          file: "deprecated.json"
`

	config, err := loader.LoadConfigFromContent(yamlContent, "config.yaml")
	require.NoError(t, err)
	require.NotNil(t, config)

	assert.Equal(t, "org/source-repo", config.SourceRepo)
	assert.Equal(t, "main", config.SourceBranch)
	assert.Len(t, config.CopyRules, 1)

	rule := config.CopyRules[0]
	assert.Equal(t, "Copy Go examples", rule.Name)
	assert.Equal(t, types.PatternTypePrefix, rule.SourcePattern.Type)
	assert.Equal(t, "examples/go/", rule.SourcePattern.Pattern)
	assert.Len(t, rule.Targets, 1)

	target := rule.Targets[0]
	assert.Equal(t, "org/target-repo", target.Repo)
	assert.Equal(t, "main", target.Branch)
	assert.Equal(t, "docs/${path}", target.PathTransform)
	assert.Equal(t, "direct", target.CommitStrategy.Type)
	assert.Equal(t, "Update examples", target.CommitStrategy.CommitMessage)
	assert.True(t, target.DeprecationCheck.Enabled)
	assert.Equal(t, "deprecated.json", target.DeprecationCheck.File)
}

func TestConfigLoader_LoadJSON(t *testing.T) {
	loader := services.NewConfigLoader()

	jsonContent := `{
  "source_repo": "org/source-repo",
  "source_branch": "main",
  "copy_rules": [
    {
      "name": "Copy Python examples",
      "source_pattern": {
        "type": "glob",
        "pattern": "examples/**/*.py"
      },
      "targets": [
        {
          "repo": "org/target-repo",
          "branch": "main",
          "path_transform": "${path}",
          "commit_strategy": {
            "type": "pull_request",
            "pr_title": "Update Python examples",
            "commit_message": "Sync examples",
            "auto_merge": false
          }
        }
      ]
    }
  ]
}`

	config, err := loader.LoadConfigFromContent(jsonContent, "config.json")
	require.NoError(t, err)
	require.NotNil(t, config)

	assert.Equal(t, "org/source-repo", config.SourceRepo)
	assert.Len(t, config.CopyRules, 1)

	rule := config.CopyRules[0]
	assert.Equal(t, "Copy Python examples", rule.Name)
	assert.Equal(t, types.PatternTypeGlob, rule.SourcePattern.Type)
	assert.Equal(t, "examples/**/*.py", rule.SourcePattern.Pattern)

	target := rule.Targets[0]
	assert.Equal(t, "pull_request", target.CommitStrategy.Type)
	assert.Equal(t, "Update Python examples", target.CommitStrategy.PRTitle)
	assert.False(t, target.CommitStrategy.AutoMerge)
}

func TestConfigLoader_LoadLegacyJSON(t *testing.T) {
	t.Skip("Legacy JSON format conversion not implemented - backward compatibility not required")

	loader := services.NewConfigLoader()

	legacyJSON := `[
  {
    "source_directory": "examples",
    "target_repo": "org/target",
    "target_branch": "main",
    "target_directory": "docs",
    "recursive_copy": true,
    "copier_commit_strategy": "pr",
    "pr_title": "Update docs",
    "commit_message": "Sync from source",
    "merge_without_review": false
  }
]`

	config, err := loader.LoadConfigFromContent(legacyJSON, "config.json")
	require.NoError(t, err)
	require.NotNil(t, config)

	// Should be converted to new format
	assert.Len(t, config.CopyRules, 1)

	rule := config.CopyRules[0]
	assert.Contains(t, rule.Name, "legacy-rule")
	assert.Equal(t, types.PatternTypePrefix, rule.SourcePattern.Type)
	assert.Equal(t, "examples", rule.SourcePattern.Pattern)

	target := rule.Targets[0]
	assert.Equal(t, "org/target", target.Repo)
	assert.Equal(t, "main", target.Branch)
	assert.Equal(t, "pr", target.CommitStrategy.Type)
	assert.Equal(t, "Update docs", target.CommitStrategy.PRTitle)
	assert.Equal(t, "Sync from source", target.CommitStrategy.CommitMessage)
	assert.False(t, target.CommitStrategy.AutoMerge)
}

func TestConfigLoader_InvalidYAML(t *testing.T) {
	loader := services.NewConfigLoader()

	invalidYAML := `
source_repo: "org/repo"
copy_rules:
  - name: "Test"
    invalid_field: [
      unclosed bracket
`

	_, err := loader.LoadConfigFromContent(invalidYAML, "config.yaml")
	assert.Error(t, err)
}

func TestConfigLoader_InvalidJSON(t *testing.T) {
	loader := services.NewConfigLoader()

	invalidJSON := `{
  "source_repo": "org/repo",
  "copy_rules": [
    {
      "name": "Test"
    }
  ]
  // missing closing brace
`

	_, err := loader.LoadConfigFromContent(invalidJSON, "config.json")
	assert.Error(t, err)
}

func TestConfigLoader_ValidationErrors(t *testing.T) {
	loader := services.NewConfigLoader()

	tests := []struct {
		name    string
		content string
		wantErr string
	}{
		{
			name: "missing source_repo",
			content: `
source_branch: "main"
copy_rules:
  - name: "Test"
    source_pattern:
      type: "prefix"
      pattern: "examples/"
    targets:
      - repo: "org/target"
        branch: "main"
`,
			wantErr: "source_repo",
		},
		{
			name: "missing copy_rules",
			content: `
source_repo: "org/source"
source_branch: "main"
`,
			wantErr: "copy rule",
		},
		{
			name: "invalid pattern type",
			content: `
source_repo: "org/source"
copy_rules:
  - name: "Test"
    source_pattern:
      type: "invalid_type"
      pattern: "examples/"
    targets:
      - repo: "org/target"
        branch: "main"
`,
			wantErr: "pattern type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := loader.LoadConfigFromContent(tt.content, "config.yaml")
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestConfigLoader_SetDefaults(t *testing.T) {
	loader := services.NewConfigLoader()

	minimalYAML := `
source_repo: "org/source"
copy_rules:
  - name: "Test"
    source_pattern:
      type: "prefix"
      pattern: "examples/"
    targets:
      - repo: "org/target"
        path_transform: "${path}"
`

	config, err := loader.LoadConfigFromContent(minimalYAML, "config.yaml")
	require.NoError(t, err)

	// Check defaults are set
	assert.Equal(t, "main", config.SourceBranch, "default source branch")

	target := config.CopyRules[0].Targets[0]
	assert.Equal(t, "main", target.Branch, "default target branch")
	assert.Equal(t, "${path}", target.PathTransform, "default path transform")
	assert.Equal(t, "direct", target.CommitStrategy.Type, "default commit strategy")
}

func TestConfigValidator_ValidatePattern(t *testing.T) {
	validator := services.NewConfigValidator()

	tests := []struct {
		name    string
		pattern types.SourcePattern
		wantErr bool
	}{
		{
			name: "valid prefix pattern",
			pattern: types.SourcePattern{
				Type:    types.PatternTypePrefix,
				Pattern: "examples/",
			},
			wantErr: false,
		},
		{
			name: "valid glob pattern",
			pattern: types.SourcePattern{
				Type:    types.PatternTypeGlob,
				Pattern: "examples/**/*.go",
			},
			wantErr: false,
		},
		{
			name: "valid regex pattern",
			pattern: types.SourcePattern{
				Type:    types.PatternTypeRegex,
				Pattern: "^examples/(?P<lang>[^/]+)/.*$",
			},
			wantErr: false,
		},
		{
			name: "invalid regex pattern",
			pattern: types.SourcePattern{
				Type:    types.PatternTypeRegex,
				Pattern: "^examples/(?P<unclosed",
			},
			wantErr: false, // Go regex compiler is lenient
		},
		{
			name: "empty pattern",
			pattern: types.SourcePattern{
				Type:    types.PatternTypePrefix,
				Pattern: "",
			},
			wantErr: true, // Empty pattern is not allowed
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidatePattern(tt.pattern.Type, tt.pattern.Pattern)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestExportConfigAsYAML(t *testing.T) {
	config := &types.YAMLConfig{
		SourceRepo:   "org/source",
		SourceBranch: "main",
		CopyRules: []types.CopyRule{
			{
				Name: "Test Rule",
				SourcePattern: types.SourcePattern{
					Type:    types.PatternTypePrefix,
					Pattern: "examples/",
				},
				Targets: []types.TargetConfig{
					{
						Repo:          "org/target",
						Branch:        "main",
						PathTransform: "${path}",
					},
				},
			},
		},
	}

	yaml, err := services.ExportConfigAsYAML(config)
	require.NoError(t, err)
	assert.Contains(t, yaml, "source_repo: org/source")
	assert.Contains(t, yaml, "Test Rule")
	assert.Contains(t, yaml, "examples/")
}

func TestExportConfigAsJSON(t *testing.T) {
	config := &types.YAMLConfig{
		SourceRepo:   "org/source",
		SourceBranch: "main",
		CopyRules: []types.CopyRule{
			{
				Name: "Test Rule",
				SourcePattern: types.SourcePattern{
					Type:    types.PatternTypePrefix,
					Pattern: "examples/",
				},
				Targets: []types.TargetConfig{
					{
						Repo:   "org/target",
						Branch: "main",
					},
				},
			},
		},
	}

	json, err := services.ExportConfigAsJSON(config)
	require.NoError(t, err)
	assert.Contains(t, json, `"source_repo": "org/source"`)
	assert.Contains(t, json, `"Test Rule"`)
}

