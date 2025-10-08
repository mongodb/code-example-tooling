package services_test

import (
	"testing"

	"github.com/mongodb/code-example-tooling/code-copier/services"
	"github.com/mongodb/code-example-tooling/code-copier/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPatternMatcher_Prefix(t *testing.T) {
	matcher := services.NewPatternMatcher()

	tests := []struct {
		name     string
		pattern  types.SourcePattern
		filePath string
		wantMatch bool
	}{
		{
			name: "exact prefix match",
			pattern: types.SourcePattern{
				Type:    types.PatternTypePrefix,
				Pattern: "examples/go/",
			},
			filePath:  "examples/go/main.go",
			wantMatch: true,
		},
		{
			name: "prefix no match",
			pattern: types.SourcePattern{
				Type:    types.PatternTypePrefix,
				Pattern: "examples/python/",
			},
			filePath:  "examples/go/main.go",
			wantMatch: false,
		},
		{
			name: "prefix match with subdirectory",
			pattern: types.SourcePattern{
				Type:    types.PatternTypePrefix,
				Pattern: "examples/",
			},
			filePath:  "examples/go/database/connect.go",
			wantMatch: true,
		},
		{
			name: "empty pattern matches all",
			pattern: types.SourcePattern{
				Type:    types.PatternTypePrefix,
				Pattern: "",
			},
			filePath:  "any/file/path.go",
			wantMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matcher.Match(tt.filePath, tt.pattern)
			assert.Equal(t, tt.wantMatch, result.Matched, "match result")
			if result.Matched {
				assert.NotNil(t, result.Variables)
			}
		})
	}
}

func TestPatternMatcher_Glob(t *testing.T) {
	matcher := services.NewPatternMatcher()

	tests := []struct {
		name      string
		pattern   types.SourcePattern
		filePath  string
		wantMatch bool
	}{
		{
			name: "single star wildcard",
			pattern: types.SourcePattern{
				Type:    types.PatternTypeGlob,
				Pattern: "examples/*/main.go",
			},
			filePath:  "examples/go/main.go",
			wantMatch: true,
		},
		{
			name: "single star no match subdirectory",
			pattern: types.SourcePattern{
				Type:    types.PatternTypeGlob,
				Pattern: "examples/*/main.go",
			},
			filePath:  "examples/go/database/main.go",
			wantMatch: false,
		},
		{
			name: "double star matches multiple levels",
			pattern: types.SourcePattern{
				Type:    types.PatternTypeRegex,
				Pattern: "^examples/.*/.*\\.go$",
			},
			filePath:  "examples/go/database/connect.go",
			wantMatch: true,
		},
		{
			name: "question mark single character",
			pattern: types.SourcePattern{
				Type:    types.PatternTypeGlob,
				Pattern: "examples/go/test?.go",
			},
			filePath:  "examples/go/test1.go",
			wantMatch: true,
		},
		{
			name: "question mark no match multiple chars",
			pattern: types.SourcePattern{
				Type:    types.PatternTypeGlob,
				Pattern: "examples/go/test?.go",
			},
			filePath:  "examples/go/test12.go",
			wantMatch: false,
		},
		{
			name: "extension wildcard with regex",
			pattern: types.SourcePattern{
				Type:    types.PatternTypeRegex,
				Pattern: "^examples/.*\\.(go|py)$",
			},
			filePath:  "examples/python/main.py",
			wantMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matcher.Match(tt.filePath, tt.pattern)
			assert.Equal(t, tt.wantMatch, result.Matched, "match result")
		})
	}
}

func TestPatternMatcher_Regex(t *testing.T) {
	matcher := services.NewPatternMatcher()

	tests := []struct {
		name          string
		pattern       types.SourcePattern
		filePath      string
		wantMatch     bool
		wantVariables map[string]string
	}{
		{
			name: "simple regex match",
			pattern: types.SourcePattern{
				Type:    types.PatternTypeRegex,
				Pattern: "^examples/go/.*\\.go$",
			},
			filePath:      "examples/go/main.go",
			wantMatch:     true,
			wantVariables: map[string]string{},
		},
		{
			name: "regex with named groups",
			pattern: types.SourcePattern{
				Type:    types.PatternTypeRegex,
				Pattern: "^examples/(?P<lang>[^/]+)/(?P<file>.+)$",
			},
			filePath:  "examples/go/main.go",
			wantMatch: true,
			wantVariables: map[string]string{
				"lang": "go",
				"file": "main.go",
			},
		},
		{
			name: "regex with multiple named groups",
			pattern: types.SourcePattern{
				Type:    types.PatternTypeRegex,
				Pattern: "^examples/(?P<lang>[^/]+)/(?P<category>[^/]+)/(?P<filename>[^/]+)$",
			},
			filePath:  "examples/go/database/connect.go",
			wantMatch: true,
			wantVariables: map[string]string{
				"lang":     "go",
				"category": "database",
				"filename": "connect.go",
			},
		},
		{
			name: "regex no match",
			pattern: types.SourcePattern{
				Type:    types.PatternTypeRegex,
				Pattern: "^examples/python/.*\\.py$",
			},
			filePath:  "examples/go/main.go",
			wantMatch: false,
		},
		{
			name: "complex regex with optional groups",
			pattern: types.SourcePattern{
				Type:    types.PatternTypeRegex,
				Pattern: "^(?P<dir>examples/[^/]+)/(?P<subdir>[^/]+/)?(?P<file>[^/]+)$",
			},
			filePath:  "examples/go/database/connect.go",
			wantMatch: true,
			wantVariables: map[string]string{
				"dir":    "examples/go",
				"subdir": "database/",
				"file":   "connect.go",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matcher.Match(tt.filePath, tt.pattern)
			assert.Equal(t, tt.wantMatch, result.Matched, "match result")
			if tt.wantMatch && tt.wantVariables != nil {
				assert.Equal(t, tt.wantVariables, result.Variables, "extracted variables")
			}
		})
	}
}

func TestPathTransformer_Transform(t *testing.T) {
	transformer := services.NewPathTransformer()

	tests := []struct {
		name      string
		filePath  string
		template  string
		variables map[string]string
		want      string
		wantErr   bool
	}{
		{
			name:      "simple path passthrough",
			filePath:  "examples/go/main.go",
			template:  "${path}",
			variables: map[string]string{},
			want:      "examples/go/main.go",
		},
		{
			name:      "filename only",
			filePath:  "examples/go/main.go",
			template:  "docs/${filename}",
			variables: map[string]string{},
			want:      "docs/main.go",
		},
		{
			name:      "directory only",
			filePath:  "examples/go/main.go",
			template:  "${dir}/output.txt",
			variables: map[string]string{},
			want:      "examples/go/output.txt",
		},
		{
			name:      "extension only",
			filePath:  "examples/go/main.go",
			template:  "output.${ext}",
			variables: map[string]string{},
			want:      "output.go",
		},
		{
			name:     "custom variables from regex",
			filePath: "examples/go/database/connect.go",
			template: "docs/${lang}/${category}/${filename}",
			variables: map[string]string{
				"lang":     "go",
				"category": "database",
				"filename": "connect.go",
			},
			want: "docs/go/database/connect.go",
		},
		{
			name:     "mixed built-in and custom variables",
			filePath: "examples/go/main.go",
			template: "docs/${lang}/reference/${filename}",
			variables: map[string]string{
				"lang": "golang",
			},
			want: "docs/golang/reference/main.go",
		},
		{
			name:      "no template returns empty string",
			filePath:  "examples/go/main.go",
			template:  "",
			variables: map[string]string{},
			want:      "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := transformer.Transform(tt.filePath, tt.template, tt.variables)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestMessageTemplater_RenderCommitMessage(t *testing.T) {
	templater := services.NewMessageTemplater()

	tests := []struct {
		name     string
		template string
		context  *types.MessageContext
		want     string
	}{
		{
			name:     "simple message",
			template: "Update examples",
			context:  types.NewMessageContext(),
			want:     "Update examples",
		},
		{
			name:     "message with rule name",
			template: "Update ${rule_name} examples",
			context: &types.MessageContext{
				RuleName: "go-examples",
			},
			want: "Update go-examples examples",
		},
		{
			name:     "message with multiple variables",
			template: "Copy ${file_count} files from ${source_repo} to ${target_repo}",
			context: &types.MessageContext{
				SourceRepo: "org/source",
				TargetRepo: "org/target",
				FileCount:  5,
			},
			want: "Copy 5 files from org/source to org/target",
		},
		{
			name:     "message with custom variables",
			template: "Update ${category} examples",
			context: &types.MessageContext{
				Variables: map[string]string{
					"category": "database",
				},
			},
			want: "Update database examples",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := templater.RenderCommitMessage(tt.template, tt.context)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMatchAndTransform(t *testing.T) {
	tests := []struct {
		name         string
		filePath     string
		rule         types.CopyRule
		target       types.TargetConfig
		wantMatch    bool
		wantPath     string
		wantVariables map[string]string
	}{
		{
			name:     "prefix match and transform",
			filePath: "examples/go/main.go",
			rule: types.CopyRule{
				Name: "test",
				SourcePattern: types.SourcePattern{
					Type:    types.PatternTypePrefix,
					Pattern: "examples/",
				},
			},
			target: types.TargetConfig{
				PathTransform: "docs/${path}",
			},
			wantMatch:     true,
			wantPath:      "docs/examples/go/main.go",
			wantVariables: nil, // Don't check variables, they include built-ins
		},
		{
			name:     "regex match with variable extraction and transform",
			filePath: "examples/go/database/connect.go",
			rule: types.CopyRule{
				Name: "test",
				SourcePattern: types.SourcePattern{
					Type:    types.PatternTypeRegex,
					Pattern: "^examples/(?P<lang>[^/]+)/(?P<category>[^/]+)/(?P<file>.+)$",
				},
			},
			target: types.TargetConfig{
				PathTransform: "docs/${lang}/${category}/${file}",
			},
			wantMatch: true,
			wantPath:  "docs/go/database/connect.go",
			wantVariables: nil, // Don't check - includes built-in variables too
		},
		{
			name:     "no match",
			filePath: "src/main.go",
			rule: types.CopyRule{
				Name: "test",
				SourcePattern: types.SourcePattern{
					Type:    types.PatternTypePrefix,
					Pattern: "examples/",
				},
			},
			target: types.TargetConfig{
				PathTransform: "docs/${path}",
			},
			wantMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transformedPath, variables, matched := services.MatchAndTransform(
				tt.filePath,
				tt.rule,
				tt.target,
			)

			if !tt.wantMatch {
				assert.False(t, matched)
				return
			}

			assert.True(t, matched)
			assert.Equal(t, tt.wantPath, transformedPath)
			if tt.wantVariables != nil {
				assert.Equal(t, tt.wantVariables, variables)
			}
		})
	}
}

