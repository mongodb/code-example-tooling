package services

import (
	"testing"

	"github.com/mongodb/code-example-tooling/code-copier/types"
)

func TestExcludePatterns_PrefixPattern(t *testing.T) {
	matcher := NewPatternMatcher()

	tests := []struct {
		name            string
		filePath        string
		pattern         string
		excludePatterns []string
		shouldMatch     bool
	}{
		{
			name:            "No exclusions - should match",
			filePath:        "examples/test.js",
			pattern:         "examples/",
			excludePatterns: nil,
			shouldMatch:     true,
		},
		{
			name:            "Exclude .gitignore - should not match",
			filePath:        "examples/.gitignore",
			pattern:         "examples/",
			excludePatterns: []string{`\.gitignore$`},
			shouldMatch:     false,
		},
		{
			name:            "Exclude .gitignore - other file should match",
			filePath:        "examples/test.js",
			pattern:         "examples/",
			excludePatterns: []string{`\.gitignore$`},
			shouldMatch:     true,
		},
		{
			name:            "Exclude multiple patterns - .env excluded",
			filePath:        "examples/.env",
			pattern:         "examples/",
			excludePatterns: []string{`\.gitignore$`, `\.env$`},
			shouldMatch:     false,
		},
		{
			name:            "Exclude multiple patterns - .gitignore excluded",
			filePath:        "examples/.gitignore",
			pattern:         "examples/",
			excludePatterns: []string{`\.gitignore$`, `\.env$`},
			shouldMatch:     false,
		},
		{
			name:            "Exclude multiple patterns - normal file matches",
			filePath:        "examples/test.js",
			pattern:         "examples/",
			excludePatterns: []string{`\.gitignore$`, `\.env$`},
			shouldMatch:     true,
		},
		{
			name:            "Exclude all .md files",
			filePath:        "examples/README.md",
			pattern:         "examples/",
			excludePatterns: []string{`\.md$`},
			shouldMatch:     false,
		},
		{
			name:            "Exclude node_modules directory",
			filePath:        "examples/node_modules/package.json",
			pattern:         "examples/",
			excludePatterns: []string{`node_modules/`},
			shouldMatch:     false,
		},
		{
			name:            "Exclude build artifacts",
			filePath:        "examples/dist/bundle.js",
			pattern:         "examples/",
			excludePatterns: []string{`/dist/`, `/build/`, `\.min\.js$`},
			shouldMatch:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sourcePattern := types.SourcePattern{
				Type:            types.PatternTypePrefix,
				Pattern:         tt.pattern,
				ExcludePatterns: tt.excludePatterns,
			}

			result := matcher.Match(tt.filePath, sourcePattern)

			if result.Matched != tt.shouldMatch {
				t.Errorf("Expected match=%v, got match=%v for file=%s with excludes=%v",
					tt.shouldMatch, result.Matched, tt.filePath, tt.excludePatterns)
			}
		})
	}
}

func TestExcludePatterns_RegexPattern(t *testing.T) {
	matcher := NewPatternMatcher()

	tests := []struct {
		name            string
		filePath        string
		pattern         string
		excludePatterns []string
		shouldMatch     bool
	}{
		{
			name:            "Regex match with no exclusions",
			filePath:        "mflix/server/java-spring/src/Main.java",
			pattern:         `^mflix/server/java-spring/(?P<file>.+)$`,
			excludePatterns: nil,
			shouldMatch:     true,
		},
		{
			name:            "Regex match - exclude .gitignore",
			filePath:        "mflix/server/java-spring/.gitignore",
			pattern:         `^mflix/server/java-spring/(?P<file>.+)$`,
			excludePatterns: []string{`\.gitignore$`},
			shouldMatch:     false,
		},
		{
			name:            "Regex match - exclude test files",
			filePath:        "mflix/server/java-spring/src/test/TestMain.java",
			pattern:         `^mflix/server/java-spring/(?P<file>.+)$`,
			excludePatterns: []string{`/test/`},
			shouldMatch:     false,
		},
		{
			name:            "Regex match - normal file passes exclusion",
			filePath:        "mflix/server/java-spring/src/Main.java",
			pattern:         `^mflix/server/java-spring/(?P<file>.+)$`,
			excludePatterns: []string{`/test/`, `\.gitignore$`},
			shouldMatch:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sourcePattern := types.SourcePattern{
				Type:            types.PatternTypeRegex,
				Pattern:         tt.pattern,
				ExcludePatterns: tt.excludePatterns,
			}

			result := matcher.Match(tt.filePath, sourcePattern)

			if result.Matched != tt.shouldMatch {
				t.Errorf("Expected match=%v, got match=%v for file=%s with excludes=%v",
					tt.shouldMatch, result.Matched, tt.filePath, tt.excludePatterns)
			}
		})
	}
}

func TestExcludePatterns_GlobPattern(t *testing.T) {
	matcher := NewPatternMatcher()

	tests := []struct {
		name            string
		filePath        string
		pattern         string
		excludePatterns []string
		shouldMatch     bool
	}{
		{
			name:            "Glob match with no exclusions",
			filePath:        "examples/test.js",
			pattern:         "examples/*.js",
			excludePatterns: nil,
			shouldMatch:     true,
		},
		{
			name:            "Glob match - exclude .min.js files",
			filePath:        "examples/bundle.min.js",
			pattern:         "examples/*.js",
			excludePatterns: []string{`\.min\.js$`},
			shouldMatch:     false,
		},
		{
			name:            "Glob match - normal file passes exclusion",
			filePath:        "examples/app.js",
			pattern:         "examples/*.js",
			excludePatterns: []string{`\.min\.js$`},
			shouldMatch:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sourcePattern := types.SourcePattern{
				Type:            types.PatternTypeGlob,
				Pattern:         tt.pattern,
				ExcludePatterns: tt.excludePatterns,
			}

			result := matcher.Match(tt.filePath, sourcePattern)

			if result.Matched != tt.shouldMatch {
				t.Errorf("Expected match=%v, got match=%v for file=%s with excludes=%v",
					tt.shouldMatch, result.Matched, tt.filePath, tt.excludePatterns)
			}
		})
	}
}

func TestExcludePatterns_Validation(t *testing.T) {
	tests := []struct {
		name            string
		excludePatterns []string
		shouldError     bool
	}{
		{
			name:            "Valid regex patterns",
			excludePatterns: []string{`\.gitignore$`, `\.env$`, `/node_modules/`},
			shouldError:     false,
		},
		{
			name:            "Empty pattern - should error",
			excludePatterns: []string{""},
			shouldError:     true,
		},
		{
			name:            "Invalid regex - should error",
			excludePatterns: []string{`[invalid`},
			shouldError:     true,
		},
		{
			name:            "Mix of valid and invalid - should error",
			excludePatterns: []string{`\.gitignore$`, `[invalid`},
			shouldError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sourcePattern := types.SourcePattern{
				Type:            types.PatternTypePrefix,
				Pattern:         "examples/",
				ExcludePatterns: tt.excludePatterns,
			}

			err := sourcePattern.Validate()

			if tt.shouldError && err == nil {
				t.Errorf("Expected validation error for patterns=%v, but got none", tt.excludePatterns)
			}
			if !tt.shouldError && err != nil {
				t.Errorf("Expected no validation error for patterns=%v, but got: %v", tt.excludePatterns, err)
			}
		})
	}
}

func TestExcludePatterns_ComplexScenarios(t *testing.T) {
	matcher := NewPatternMatcher()

	tests := []struct {
		name            string
		filePath        string
		pattern         string
		excludePatterns []string
		shouldMatch     bool
		description     string
	}{
		{
			name:            "Exclude all hidden files",
			filePath:        "examples/.hidden",
			pattern:         "examples/",
			excludePatterns: []string{`/\.[^/]+$`},
			shouldMatch:     false,
			description:     "Files starting with . should be excluded",
		},
		{
			name:            "Exclude all hidden files - normal file matches",
			filePath:        "examples/visible.txt",
			pattern:         "examples/",
			excludePatterns: []string{`/\.[^/]+$`},
			shouldMatch:     true,
			description:     "Normal files should match",
		},
		{
			name:            "Exclude build artifacts and dependencies",
			filePath:        "examples/node_modules/package.json",
			pattern:         "examples/",
			excludePatterns: []string{`node_modules/`, `dist/`, `build/`, `\.min\.(js|css)$`},
			shouldMatch:     false,
			description:     "node_modules should be excluded",
		},
		{
			name:            "Exclude build artifacts - source file matches",
			filePath:        "examples/src/app.js",
			pattern:         "examples/",
			excludePatterns: []string{`node_modules/`, `dist/`, `build/`, `\.min\.(js|css)$`},
			shouldMatch:     true,
			description:     "Source files should match",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sourcePattern := types.SourcePattern{
				Type:            types.PatternTypePrefix,
				Pattern:         tt.pattern,
				ExcludePatterns: tt.excludePatterns,
			}

			result := matcher.Match(tt.filePath, sourcePattern)

			if result.Matched != tt.shouldMatch {
				t.Errorf("%s: Expected match=%v, got match=%v for file=%s",
					tt.description, tt.shouldMatch, result.Matched, tt.filePath)
			}
		})
	}
}

