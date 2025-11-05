package types

import (
	"fmt"
	"regexp"
	"strings"
)

// PatternType defines the type of pattern matching to use
type PatternType string

const (
	PatternTypePrefix PatternType = "prefix"
	PatternTypeGlob   PatternType = "glob"
	PatternTypeRegex  PatternType = "regex"
)

// IsValid returns true if the pattern type is valid
func (p PatternType) IsValid() bool {
	return p == PatternTypePrefix || p == PatternTypeGlob || p == PatternTypeRegex
}

// String returns the string representation
func (p PatternType) String() string {
	return string(p)
}

// YAMLConfig represents the new YAML-based configuration structure
type YAMLConfig struct {
	SourceRepo    string         `yaml:"source_repo" json:"source_repo"`
	SourceBranch  string         `yaml:"source_branch" json:"source_branch"`
	BatchByRepo   bool           `yaml:"batch_by_repo,omitempty" json:"batch_by_repo,omitempty"`       // If true, batch all changes into one PR per target repo
	BatchPRConfig *BatchPRConfig `yaml:"batch_pr_config,omitempty" json:"batch_pr_config,omitempty"` // PR config used when batch_by_repo is true
	CopyRules     []CopyRule     `yaml:"copy_rules" json:"copy_rules"`
}

// BatchPRConfig defines PR metadata for batched PRs
type BatchPRConfig struct {
	PRTitle       string `yaml:"pr_title,omitempty" json:"pr_title,omitempty"`
	PRBody        string `yaml:"pr_body,omitempty" json:"pr_body,omitempty"`
	CommitMessage string `yaml:"commit_message,omitempty" json:"commit_message,omitempty"`
	UsePRTemplate bool   `yaml:"use_pr_template,omitempty" json:"use_pr_template,omitempty"`
}

// CopyRule defines a single rule for copying files with pattern matching
type CopyRule struct {
	Name          string          `yaml:"name" json:"name"`
	SourcePattern SourcePattern   `yaml:"source_pattern" json:"source_pattern"`
	Targets       []TargetConfig  `yaml:"targets" json:"targets"`
}

// SourcePattern defines how to match source files
type SourcePattern struct {
	Type            PatternType `yaml:"type" json:"type"`
	Pattern         string      `yaml:"pattern" json:"pattern"`
	ExcludePatterns []string    `yaml:"exclude_patterns,omitempty" json:"exclude_patterns,omitempty"` // Optional: regex patterns to exclude from matches
}

// TargetConfig defines where and how to copy matched files
type TargetConfig struct {
	Repo              string            `yaml:"repo" json:"repo"`
	Branch            string            `yaml:"branch" json:"branch"`
	PathTransform     string            `yaml:"path_transform" json:"path_transform"`
	CommitStrategy    CommitStrategyConfig `yaml:"commit_strategy,omitempty" json:"commit_strategy,omitempty"`
	DeprecationCheck  *DeprecationConfig   `yaml:"deprecation_check,omitempty" json:"deprecation_check,omitempty"`
}

// CommitStrategyConfig defines how to commit changes
type CommitStrategyConfig struct {
	Type          string `yaml:"type" json:"type"` // "direct", "pull_request", or "batch"
	CommitMessage string `yaml:"commit_message,omitempty" json:"commit_message,omitempty"`
	PRTitle       string `yaml:"pr_title,omitempty" json:"pr_title,omitempty"`
	PRBody        string `yaml:"pr_body,omitempty" json:"pr_body,omitempty"`
	UsePRTemplate bool   `yaml:"use_pr_template,omitempty" json:"use_pr_template,omitempty"` // If true, fetch and use PR template from target repo
	AutoMerge     bool   `yaml:"auto_merge,omitempty" json:"auto_merge,omitempty"`
	BatchSize     int    `yaml:"batch_size,omitempty" json:"batch_size,omitempty"`
}

// DeprecationConfig defines deprecation tracking settings
type DeprecationConfig struct {
	Enabled bool   `yaml:"enabled" json:"enabled"`
	File    string `yaml:"file,omitempty" json:"file,omitempty"` // defaults to deprecated_examples.json
}

// Validate validates the YAML configuration
func (c *YAMLConfig) Validate() error {
	if c.SourceRepo == "" {
		return fmt.Errorf("source_repo is required")
	}
	if c.SourceBranch == "" {
		c.SourceBranch = "main" // default
	}
	if len(c.CopyRules) == 0 {
		return fmt.Errorf("at least one copy rule is required")
	}

	for i, rule := range c.CopyRules {
		if err := rule.Validate(); err != nil {
			return fmt.Errorf("copy_rules[%d]: %w", i, err)
		}
	}

	return nil
}

// Validate validates a copy rule
func (r *CopyRule) Validate() error {
	if r.Name == "" {
		return fmt.Errorf("name is required")
	}
	if err := r.SourcePattern.Validate(); err != nil {
		return fmt.Errorf("source_pattern: %w", err)
	}
	if len(r.Targets) == 0 {
		return fmt.Errorf("at least one target is required")
	}

	for i, target := range r.Targets {
		if err := target.Validate(); err != nil {
			return fmt.Errorf("targets[%d]: %w", i, err)
		}
	}

	return nil
}

// Validate validates a source pattern
func (p *SourcePattern) Validate() error {
	if !p.Type.IsValid() {
		return fmt.Errorf("invalid pattern type: %s (must be prefix, glob, or regex)", p.Type)
	}
	if p.Pattern == "" {
		return fmt.Errorf("pattern is required")
	}

	// Validate exclude patterns if provided
	if len(p.ExcludePatterns) > 0 {
		for i, excludePattern := range p.ExcludePatterns {
			if excludePattern == "" {
				return fmt.Errorf("exclude_patterns[%d] is empty", i)
			}
			// Validate that it's a valid regex pattern
			if _, err := regexp.Compile(excludePattern); err != nil {
				return fmt.Errorf("exclude_patterns[%d] is not a valid regex: %w", i, err)
			}
		}
	}

	return nil
}

// Validate validates a target config
func (t *TargetConfig) Validate() error {
	if t.Repo == "" {
		return fmt.Errorf("repo is required")
	}
	if t.Branch == "" {
		t.Branch = "main" // default
	}
	if t.PathTransform == "" {
		return fmt.Errorf("path_transform is required")
	}

	// Validate commit strategy if provided
	if t.CommitStrategy.Type != "" {
		if err := t.CommitStrategy.Validate(); err != nil {
			return fmt.Errorf("commit_strategy: %w", err)
		}
	}

	return nil
}

// Validate validates a commit strategy config
func (c *CommitStrategyConfig) Validate() error {
	validTypes := map[string]bool{
		"direct":       true,
		"pull_request": true,
		"batch":        true,
	}

	if c.Type != "" && !validTypes[c.Type] {
		return fmt.Errorf("invalid type: %s (must be direct, pull_request, or batch)", c.Type)
	}

	if c.Type == "batch" && c.BatchSize <= 0 {
		c.BatchSize = 100 // default batch size
	}

	return nil
}

// SetDefaults sets default values for the configuration
func (c *YAMLConfig) SetDefaults() {
	if c.SourceBranch == "" {
		c.SourceBranch = "main"
	}

	for i := range c.CopyRules {
		for j := range c.CopyRules[i].Targets {
			target := &c.CopyRules[i].Targets[j]
			if target.Branch == "" {
				target.Branch = "main"
			}
			if target.CommitStrategy.Type == "" {
				target.CommitStrategy.Type = "direct"
			}
			if target.DeprecationCheck != nil && target.DeprecationCheck.File == "" {
				target.DeprecationCheck.File = "deprecated_examples.json"
			}
		}
	}
}

// MatchResult represents the result of a pattern match
type MatchResult struct {
	Matched   bool              // Whether the pattern matched
	Variables map[string]string // Extracted variables from the match
}

// NewMatchResult creates a new match result
func NewMatchResult(matched bool, variables map[string]string) MatchResult {
	if variables == nil {
		variables = make(map[string]string)
	}
	return MatchResult{
		Matched:   matched,
		Variables: variables,
	}
}

// TransformContext holds context for path transformations
type TransformContext struct {
	SourcePath string            // Original source file path
	Variables  map[string]string // Variables extracted from pattern matching
}

// NewTransformContext creates a new transformation context
func NewTransformContext(sourcePath string, variables map[string]string) *TransformContext {
	if variables == nil {
		variables = make(map[string]string)
	}
	return &TransformContext{
		SourcePath: sourcePath,
		Variables:  variables,
	}
}

// AddBuiltInVariables adds built-in variables like ${path}, ${filename}, ${dir}, ${ext}
func (tc *TransformContext) AddBuiltInVariables() {
	tc.Variables["path"] = tc.SourcePath
	
	// Extract filename
	lastSlash := strings.LastIndex(tc.SourcePath, "/")
	if lastSlash >= 0 {
		tc.Variables["filename"] = tc.SourcePath[lastSlash+1:]
		tc.Variables["dir"] = tc.SourcePath[:lastSlash]
	} else {
		tc.Variables["filename"] = tc.SourcePath
		tc.Variables["dir"] = ""
	}
	
	// Extract extension
	filename := tc.Variables["filename"]
	lastDot := strings.LastIndex(filename, ".")
	if lastDot >= 0 {
		tc.Variables["ext"] = filename[lastDot+1:]
	} else {
		tc.Variables["ext"] = ""
	}
}

// MessageContext holds context for message template rendering
type MessageContext struct {
	RuleName      string            // Name of the copy rule
	SourceRepo    string            // Source repository
	TargetRepo    string            // Target repository
	SourceBranch  string            // Source branch
	TargetBranch  string            // Target branch
	FileCount     int               // Number of files being copied
	PRNumber      int               // PR number that triggered the copy
	CommitSHA     string            // Commit SHA
	Variables     map[string]string // Variables from pattern matching
}

// NewMessageContext creates a new message context
func NewMessageContext() *MessageContext {
	return &MessageContext{
		Variables: make(map[string]string),
	}
}

