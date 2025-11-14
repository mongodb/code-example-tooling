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
	// Legacy format (for backward compatibility during transition)
	SourceRepo    string         `yaml:"source_repo,omitempty" json:"source_repo,omitempty"`
	SourceBranch  string         `yaml:"source_branch,omitempty" json:"source_branch,omitempty"`
	BatchByRepo   bool           `yaml:"batch_by_repo,omitempty" json:"batch_by_repo,omitempty"`       // If true, batch all changes into one PR per target repo
	BatchPRConfig *BatchPRConfig `yaml:"batch_pr_config,omitempty" json:"batch_pr_config,omitempty"` // PR config used when batch_by_repo is true
	CopyRules     []CopyRule     `yaml:"copy_rules,omitempty" json:"copy_rules,omitempty"`

	// New workflow format
	Workflows []Workflow `yaml:"workflows,omitempty" json:"workflows,omitempty"`
	Defaults  *Defaults  `yaml:"defaults,omitempty" json:"defaults,omitempty"`
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

// ============================================================================
// Workflow-based configuration types
// ============================================================================

// Defaults defines default settings for all workflows
type Defaults struct {
	CommitStrategy   *CommitStrategyConfig `yaml:"commit_strategy,omitempty" json:"commit_strategy,omitempty"`
	DeprecationCheck *DeprecationConfig    `yaml:"deprecation_check,omitempty" json:"deprecation_check,omitempty"`
	Exclude          []string              `yaml:"exclude,omitempty" json:"exclude,omitempty"`
}

// Workflow defines a complete source → destination mapping with transformations
type Workflow struct {
	Name             string                `yaml:"name" json:"name"`
	Source           Source                `yaml:"source" json:"source"`
	Destination      Destination           `yaml:"destination" json:"destination"`
	Transformations  []Transformation      `yaml:"transformations" json:"transformations"`
	Exclude          []string              `yaml:"exclude,omitempty" json:"exclude,omitempty"`
	CommitStrategy   *CommitStrategyConfig `yaml:"commit_strategy,omitempty" json:"commit_strategy,omitempty"`
	DeprecationCheck *DeprecationConfig    `yaml:"deprecation_check,omitempty" json:"deprecation_check,omitempty"`
}

// Source defines the source repository and branch
type Source struct {
	Repo           string `yaml:"repo" json:"repo"`
	Branch         string `yaml:"branch,omitempty" json:"branch,omitempty"`         // defaults to "main"
	InstallationID string `yaml:"installation_id,omitempty" json:"installation_id,omitempty"` // optional override
}

// Destination defines the destination repository and branch
type Destination struct {
	Repo           string `yaml:"repo" json:"repo"`
	Branch         string `yaml:"branch,omitempty" json:"branch,omitempty"`         // defaults to "main"
	InstallationID string `yaml:"installation_id,omitempty" json:"installation_id,omitempty"` // optional override
}

// TransformationType defines the type of transformation
type TransformationType string

const (
	TransformationTypeMove  TransformationType = "move"
	TransformationTypeCopy  TransformationType = "copy"
	TransformationTypeGlob  TransformationType = "glob"
	TransformationTypeRegex TransformationType = "regex"
)

// Transformation defines how to transform file paths from source to destination
type Transformation struct {
	// Type is inferred from which field is set (move, copy, glob, regex)
	Move  *MoveTransform  `yaml:"move,omitempty" json:"move,omitempty"`
	Copy  *CopyTransform  `yaml:"copy,omitempty" json:"copy,omitempty"`
	Glob  *GlobTransform  `yaml:"glob,omitempty" json:"glob,omitempty"`
	Regex *RegexTransform `yaml:"regex,omitempty" json:"regex,omitempty"`
}

// MoveTransform moves files from one directory to another
type MoveTransform struct {
	From string `yaml:"from" json:"from"` // Source path (can be directory or file)
	To   string `yaml:"to" json:"to"`     // Destination path
}

// CopyTransform copies a single file to a new location
type CopyTransform struct {
	From string `yaml:"from" json:"from"` // Source file path
	To   string `yaml:"to" json:"to"`     // Destination file path
}

// GlobTransform uses glob patterns with path transformation
type GlobTransform struct {
	Pattern   string `yaml:"pattern" json:"pattern"`       // Glob pattern (e.g., "mflix/server/**/*.js")
	Transform string `yaml:"transform" json:"transform"`   // Path transform template (e.g., "server/${relative_path}")
}

// RegexTransform uses regex patterns with named capture groups
type RegexTransform struct {
	Pattern   string `yaml:"pattern" json:"pattern"`       // Regex pattern with named groups
	Transform string `yaml:"transform" json:"transform"`   // Path transform template using captured groups
}

// Validate validates the YAML configuration
func (c *YAMLConfig) Validate() error {
	// Check if using legacy format or new workflow format
	hasLegacy := c.SourceRepo != "" || len(c.CopyRules) > 0
	hasWorkflows := len(c.Workflows) > 0

	if !hasLegacy && !hasWorkflows {
		return fmt.Errorf("either copy_rules (legacy) or workflows (new format) is required")
	}

	// Validate legacy format
	if hasLegacy {
		if c.SourceRepo == "" {
			return fmt.Errorf("source_repo is required when using copy_rules")
		}
		if c.SourceBranch == "" {
			c.SourceBranch = "main" // default
		}
		if len(c.CopyRules) == 0 {
			return fmt.Errorf("at least one copy rule is required when using legacy format")
		}

		for i, rule := range c.CopyRules {
			if err := rule.Validate(); err != nil {
				return fmt.Errorf("copy_rules[%d]: %w", i, err)
			}
		}
	}

	// Validate workflow format
	if hasWorkflows {
		for i, workflow := range c.Workflows {
			if err := workflow.Validate(); err != nil {
				return fmt.Errorf("workflows[%d]: %w", i, err)
			}
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
	// Set defaults for legacy format
	if c.SourceBranch == "" && c.SourceRepo != "" {
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

	// Set defaults for workflow format
	for i := range c.Workflows {
		workflow := &c.Workflows[i]

		// Set source defaults
		if workflow.Source.Branch == "" {
			workflow.Source.Branch = "main"
		}

		// Set destination defaults
		if workflow.Destination.Branch == "" {
			workflow.Destination.Branch = "main"
		}

		// Apply global defaults if not overridden
		if workflow.CommitStrategy == nil && c.Defaults != nil && c.Defaults.CommitStrategy != nil {
			workflow.CommitStrategy = c.Defaults.CommitStrategy
		}

		if workflow.DeprecationCheck == nil && c.Defaults != nil && c.Defaults.DeprecationCheck != nil {
			workflow.DeprecationCheck = c.Defaults.DeprecationCheck
		}

		if len(workflow.Exclude) == 0 && c.Defaults != nil && len(c.Defaults.Exclude) > 0 {
			workflow.Exclude = c.Defaults.Exclude
		}

		// Set commit strategy defaults
		if workflow.CommitStrategy != nil && workflow.CommitStrategy.Type == "" {
			workflow.CommitStrategy.Type = "pull_request"
		}

		// Set deprecation check defaults
		if workflow.DeprecationCheck != nil && workflow.DeprecationCheck.File == "" {
			workflow.DeprecationCheck.File = "deprecated_examples.json"
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

// ============================================================================
// Validation methods for workflow types
// ============================================================================

// Validate validates a workflow
func (w *Workflow) Validate() error {
	if w.Name == "" {
		return fmt.Errorf("name is required")
	}
	if err := w.Source.Validate(); err != nil {
		return fmt.Errorf("source: %w", err)
	}
	if err := w.Destination.Validate(); err != nil {
		return fmt.Errorf("destination: %w", err)
	}
	if len(w.Transformations) == 0 {
		return fmt.Errorf("at least one transformation is required")
	}

	for i, transform := range w.Transformations {
		if err := transform.Validate(); err != nil {
			return fmt.Errorf("transformations[%d]: %w", i, err)
		}
	}

	// Validate commit strategy if provided
	if w.CommitStrategy != nil {
		if err := w.CommitStrategy.Validate(); err != nil {
			return fmt.Errorf("commit_strategy: %w", err)
		}
	}

	return nil
}

// Validate validates a source
func (s *Source) Validate() error {
	if s.Repo == "" {
		return fmt.Errorf("repo is required")
	}
	if s.Branch == "" {
		s.Branch = "main" // default
	}
	return nil
}

// Validate validates a destination
func (d *Destination) Validate() error {
	if d.Repo == "" {
		return fmt.Errorf("repo is required")
	}
	if d.Branch == "" {
		d.Branch = "main" // default
	}
	return nil
}

// Validate validates a transformation
func (t *Transformation) Validate() error {
	// Count how many transformation types are set
	count := 0
	if t.Move != nil {
		count++
	}
	if t.Copy != nil {
		count++
	}
	if t.Glob != nil {
		count++
	}
	if t.Regex != nil {
		count++
	}

	if count == 0 {
		return fmt.Errorf("one of move, copy, glob, or regex must be specified")
	}
	if count > 1 {
		return fmt.Errorf("only one of move, copy, glob, or regex can be specified")
	}

	// Validate the specific transformation type
	if t.Move != nil {
		return t.Move.Validate()
	}
	if t.Copy != nil {
		return t.Copy.Validate()
	}
	if t.Glob != nil {
		return t.Glob.Validate()
	}
	if t.Regex != nil {
		return t.Regex.Validate()
	}

	return nil
}

// Validate validates a move transformation
func (m *MoveTransform) Validate() error {
	if m.From == "" {
		return fmt.Errorf("from is required")
	}
	if m.To == "" {
		return fmt.Errorf("to is required")
	}
	return nil
}

// Validate validates a copy transformation
func (c *CopyTransform) Validate() error {
	if c.From == "" {
		return fmt.Errorf("from is required")
	}
	if c.To == "" {
		return fmt.Errorf("to is required")
	}
	return nil
}

// Validate validates a glob transformation
func (g *GlobTransform) Validate() error {
	if g.Pattern == "" {
		return fmt.Errorf("pattern is required")
	}
	if g.Transform == "" {
		return fmt.Errorf("transform is required")
	}
	return nil
}

// Validate validates a regex transformation
func (r *RegexTransform) Validate() error {
	if r.Pattern == "" {
		return fmt.Errorf("pattern is required")
	}
	if r.Transform == "" {
		return fmt.Errorf("transform is required")
	}
	return nil
}

// GetType returns the type of transformation
func (t *Transformation) GetType() TransformationType {
	if t.Move != nil {
		return TransformationTypeMove
	}
	if t.Copy != nil {
		return TransformationTypeCopy
	}
	if t.Glob != nil {
		return TransformationTypeGlob
	}
	if t.Regex != nil {
		return TransformationTypeRegex
	}
	return ""
}

