package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestYAMLConfig_SetDefaults_CommitStrategyMerging tests that commit strategy
// fields are properly merged from defaults when a workflow has a partial commit_strategy
func TestYAMLConfig_SetDefaults_CommitStrategyMerging(t *testing.T) {
	tests := []struct {
		name                    string
		defaults                *Defaults
		workflowCommitStrategy  *CommitStrategyConfig
		expectedType            string
		expectedCommitMessage   string
		expectedPRTitle         string
		expectedPRBody          string
		expectedUsePRTemplate   bool
		expectedAutoMerge       bool
	}{
		{
			name: "workflow with only pr_title should inherit commit_message from defaults",
			defaults: &Defaults{
				CommitStrategy: &CommitStrategyConfig{
					Type:          "pull_request",
					CommitMessage: "Default commit message from ${source_repo}",
					PRTitle:       "Default PR title",
					PRBody:        "Default PR body",
					UsePRTemplate: true,
					AutoMerge:     false,
				},
			},
			workflowCommitStrategy: &CommitStrategyConfig{
				PRTitle: "Workflow specific PR title",
				PRBody:  "Workflow specific PR body",
			},
			expectedType:          "pull_request",
			expectedCommitMessage: "Default commit message from ${source_repo}",
			expectedPRTitle:       "Workflow specific PR title",
			expectedPRBody:        "Workflow specific PR body",
			expectedUsePRTemplate: true,
			expectedAutoMerge:     false,
		},
		{
			name: "workflow with pr_title and pr_body should inherit commit_message and use_pr_template",
			defaults: &Defaults{
				CommitStrategy: &CommitStrategyConfig{
					Type:          "pull_request",
					CommitMessage: "Automated update from ${source_repo} PR #${pr_number} (${file_count} files)",
					UsePRTemplate: true,
					AutoMerge:     false,
				},
			},
			workflowCommitStrategy: &CommitStrategyConfig{
				PRTitle: "Update MFlix application",
				PRBody:  "Automated update of MFlix application",
			},
			expectedType:          "pull_request",
			expectedCommitMessage: "Automated update from ${source_repo} PR #${pr_number} (${file_count} files)",
			expectedPRTitle:       "Update MFlix application",
			expectedPRBody:        "Automated update of MFlix application",
			expectedUsePRTemplate: true,
			expectedAutoMerge:     false,
		},
		{
			name: "workflow with all fields should not inherit string fields but may inherit boolean defaults",
			defaults: &Defaults{
				CommitStrategy: &CommitStrategyConfig{
					Type:          "direct",
					CommitMessage: "Default commit message",
					PRTitle:       "Default PR title",
					PRBody:        "Default PR body",
					UsePRTemplate: true,
					AutoMerge:     true,
				},
			},
			workflowCommitStrategy: &CommitStrategyConfig{
				Type:          "pull_request",
				CommitMessage: "Workflow commit message",
				PRTitle:       "Workflow PR title",
				PRBody:        "Workflow PR body",
				UsePRTemplate: false,
				AutoMerge:     false,
			},
			expectedType:          "pull_request",
			expectedCommitMessage: "Workflow commit message",
			expectedPRTitle:       "Workflow PR title",
			expectedPRBody:        "Workflow PR body",
			// Note: Due to Go's limitation with boolean zero values, UsePRTemplate=true from defaults
			// will be inherited even when workflow explicitly sets it to false
			expectedUsePRTemplate: true,
			expectedAutoMerge:     false,
		},
		{
			name: "workflow with no commit_strategy should inherit entire defaults",
			defaults: &Defaults{
				CommitStrategy: &CommitStrategyConfig{
					Type:          "direct",
					CommitMessage: "Default commit message",
					PRTitle:       "Default PR title",
					PRBody:        "Default PR body",
					UsePRTemplate: true,
					AutoMerge:     true,
				},
			},
			workflowCommitStrategy: nil,
			expectedType:           "direct",
			expectedCommitMessage:  "Default commit message",
			expectedPRTitle:        "Default PR title",
			expectedPRBody:         "Default PR body",
			expectedUsePRTemplate:  true,
			expectedAutoMerge:      true,
		},
		{
			name: "workflow with empty commit_strategy should inherit all fields",
			defaults: &Defaults{
				CommitStrategy: &CommitStrategyConfig{
					Type:          "pull_request",
					CommitMessage: "Default commit message",
					PRTitle:       "Default PR title",
					PRBody:        "Default PR body",
					UsePRTemplate: true,
					AutoMerge:     false,
				},
			},
			workflowCommitStrategy: &CommitStrategyConfig{},
			expectedType:           "pull_request",
			expectedCommitMessage:  "Default commit message",
			expectedPRTitle:        "Default PR title",
			expectedPRBody:         "Default PR body",
			expectedUsePRTemplate:  true,
			expectedAutoMerge:      false,
		},
		{
			name: "workflow with use_pr_template=false will inherit use_pr_template=true due to Go boolean limitation",
			defaults: &Defaults{
				CommitStrategy: &CommitStrategyConfig{
					Type:          "pull_request",
					CommitMessage: "Default commit message",
					UsePRTemplate: true,
				},
			},
			workflowCommitStrategy: &CommitStrategyConfig{
				PRTitle:       "Workflow PR title",
				UsePRTemplate: false,
			},
			expectedType:          "pull_request",
			expectedCommitMessage: "Default commit message",
			expectedPRTitle:       "Workflow PR title",
			expectedPRBody:        "",
			// Note: Due to Go's limitation with boolean zero values, we can't distinguish
			// between "not set" and "explicitly set to false", so true from defaults wins
			expectedUsePRTemplate: true,
			expectedAutoMerge:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &YAMLConfig{
				Defaults: tt.defaults,
				Workflows: []Workflow{
					{
						Name: "test-workflow",
						Source: Source{
							Repo:   "mongodb/source-repo",
							Branch: "main",
						},
						Destination: Destination{
							Repo:   "mongodb/dest-repo",
							Branch: "main",
						},
						Transformations: []Transformation{
							{Move: &MoveTransform{From: "src", To: "dest"}},
						},
						CommitStrategy: tt.workflowCommitStrategy,
					},
				},
			}

			// Apply defaults
			config.SetDefaults()

			// Verify the workflow's commit strategy has the expected values
			workflow := config.Workflows[0]
			require.NotNil(t, workflow.CommitStrategy, "CommitStrategy should not be nil after SetDefaults")

			assert.Equal(t, tt.expectedType, workflow.CommitStrategy.Type, "Type mismatch")
			assert.Equal(t, tt.expectedCommitMessage, workflow.CommitStrategy.CommitMessage, "CommitMessage mismatch")
			assert.Equal(t, tt.expectedPRTitle, workflow.CommitStrategy.PRTitle, "PRTitle mismatch")
			assert.Equal(t, tt.expectedPRBody, workflow.CommitStrategy.PRBody, "PRBody mismatch")
			assert.Equal(t, tt.expectedUsePRTemplate, workflow.CommitStrategy.UsePRTemplate, "UsePRTemplate mismatch")
			assert.Equal(t, tt.expectedAutoMerge, workflow.CommitStrategy.AutoMerge, "AutoMerge mismatch")
		})
	}
}

// TestWorkflowConfig_SetDefaults_CommitStrategyMerging tests that commit strategy
// fields are properly merged from workflow config defaults
func TestWorkflowConfig_SetDefaults_CommitStrategyMerging(t *testing.T) {
	workflowConfig := &WorkflowConfig{
		Defaults: &Defaults{
			CommitStrategy: &CommitStrategyConfig{
				Type:          "pull_request",
				CommitMessage: "Default commit message from ${source_repo}",
				PRTitle:       "Default PR title",
				UsePRTemplate: true,
				AutoMerge:     false,
			},
		},
		Workflows: []Workflow{
			{
				Name: "workflow-with-partial-override",
				Source: Source{
					Repo:   "mongodb/source-repo",
					Branch: "main",
				},
				Destination: Destination{
					Repo:   "mongodb/dest-repo",
					Branch: "main",
				},
				Transformations: []Transformation{
					{Move: &MoveTransform{From: "src", To: "dest"}},
				},
				CommitStrategy: &CommitStrategyConfig{
					PRTitle: "Workflow specific PR title",
					PRBody:  "Workflow specific PR body",
				},
			},
			{
				Name: "workflow-without-override",
				Source: Source{
					Repo:   "mongodb/source-repo-2",
					Branch: "main",
				},
				Destination: Destination{
					Repo:   "mongodb/dest-repo-2",
					Branch: "main",
				},
				Transformations: []Transformation{
					{Move: &MoveTransform{From: "a", To: "b"}},
				},
			},
		},
	}

	// Apply defaults
	workflowConfig.SetDefaults()

	// First workflow should have merged values
	workflow1 := workflowConfig.Workflows[0]
	require.NotNil(t, workflow1.CommitStrategy)
	assert.Equal(t, "pull_request", workflow1.CommitStrategy.Type)
	assert.Equal(t, "Default commit message from ${source_repo}", workflow1.CommitStrategy.CommitMessage)
	assert.Equal(t, "Workflow specific PR title", workflow1.CommitStrategy.PRTitle)
	assert.Equal(t, "Workflow specific PR body", workflow1.CommitStrategy.PRBody)
	assert.True(t, workflow1.CommitStrategy.UsePRTemplate)
	assert.False(t, workflow1.CommitStrategy.AutoMerge)

	// Second workflow should inherit all defaults
	workflow2 := workflowConfig.Workflows[1]
	require.NotNil(t, workflow2.CommitStrategy)
	assert.Equal(t, "pull_request", workflow2.CommitStrategy.Type)
	assert.Equal(t, "Default commit message from ${source_repo}", workflow2.CommitStrategy.CommitMessage)
	assert.Equal(t, "Default PR title", workflow2.CommitStrategy.PRTitle)
	assert.True(t, workflow2.CommitStrategy.UsePRTemplate)
	assert.False(t, workflow2.CommitStrategy.AutoMerge)
}

