package services

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/mongodb/code-example-tooling/code-copier/types"
)

// PatternMatcher handles pattern matching for file paths
type PatternMatcher interface {
	Match(filePath string, pattern types.SourcePattern) types.MatchResult
}

// DefaultPatternMatcher implements the PatternMatcher interface
type DefaultPatternMatcher struct{}

// NewPatternMatcher creates a new pattern matcher
func NewPatternMatcher() PatternMatcher {
	return &DefaultPatternMatcher{}
}

// Match matches a file path against a source pattern
func (pm *DefaultPatternMatcher) Match(filePath string, pattern types.SourcePattern) types.MatchResult {
	// First, check if the main pattern matches
	var result types.MatchResult
	switch pattern.Type {
	case types.PatternTypePrefix:
		result = pm.matchPrefix(filePath, pattern.Pattern)
	case types.PatternTypeGlob:
		result = pm.matchGlob(filePath, pattern.Pattern)
	case types.PatternTypeRegex:
		result = pm.matchRegex(filePath, pattern.Pattern)
	default:
		return types.NewMatchResult(false, nil)
	}

	// If the main pattern didn't match, return early
	if !result.Matched {
		return result
	}

	// Check if the file should be excluded
	if pm.shouldExclude(filePath, pattern.ExcludePatterns) {
		return types.NewMatchResult(false, nil)
	}

	// Main pattern matched and file is not excluded
	return result
}

// matchPrefix matches using simple prefix matching
func (pm *DefaultPatternMatcher) matchPrefix(filePath, pattern string) types.MatchResult {
	// Normalize paths (remove trailing slashes)
	pattern = strings.TrimSuffix(pattern, "/")
	
	if strings.HasPrefix(filePath, pattern) {
		// Extract the relative path after the prefix
		relPath := strings.TrimPrefix(filePath, pattern)
		relPath = strings.TrimPrefix(relPath, "/")
		
		variables := map[string]string{
			"matched_prefix": pattern,
			"relative_path":  relPath,
		}
		
		return types.NewMatchResult(true, variables)
	}
	
	return types.NewMatchResult(false, nil)
}

// matchGlob matches using glob patterns
func (pm *DefaultPatternMatcher) matchGlob(filePath, pattern string) types.MatchResult {
	// Use doublestar library which properly supports ** patterns
	matched, err := doublestar.Match(pattern, filePath)
	if err != nil {
		// Fall back to filepath.Match for simple patterns
		matched, err = filepath.Match(pattern, filePath)
		if err != nil {
			return types.NewMatchResult(false, nil)
		}
	}

	if matched {
		variables := map[string]string{
			"matched_pattern": pattern,
		}
		return types.NewMatchResult(true, variables)
	}

	return types.NewMatchResult(false, nil)
}



// matchRegex matches using regular expressions with named capture groups
func (pm *DefaultPatternMatcher) matchRegex(filePath, pattern string) types.MatchResult {
	re, err := regexp.Compile(pattern)
	if err != nil {
		LogInfo(fmt.Sprintf("REGEX COMPILE ERROR: pattern=%s, error=%v", pattern, err))
		return types.NewMatchResult(false, nil)
	}

	match := re.FindStringSubmatch(filePath)
	if match == nil {
		// Log server file pattern attempts for debugging
		if strings.Contains(pattern, "server/") && strings.Contains(filePath, "server/") {
			LogInfo(fmt.Sprintf("REGEX NO MATCH: file=%s, pattern=%s", filePath, pattern))
		}
		return types.NewMatchResult(false, nil)
	}

	// Extract named capture groups
	variables := make(map[string]string)
	for i, name := range re.SubexpNames() {
		if i > 0 && i <= len(match) && name != "" {
			variables[name] = match[i]
		}
	}

	// Log server file matches for debugging
	if strings.Contains(pattern, "server/") {
		LogInfo(fmt.Sprintf("REGEX MATCH SUCCESS: file=%s, pattern=%s, variables=%v", filePath, pattern, variables))
	}

	return types.NewMatchResult(true, variables)
}

// PathTransformer handles path transformations
type PathTransformer interface {
	Transform(sourcePath string, template string, variables map[string]string) (string, error)
}

// DefaultPathTransformer implements the PathTransformer interface
type DefaultPathTransformer struct{}

// NewPathTransformer creates a new path transformer
func NewPathTransformer() PathTransformer {
	return &DefaultPathTransformer{}
}

// Transform transforms a source path using a template and variables
func (pt *DefaultPathTransformer) Transform(sourcePath string, template string, variables map[string]string) (string, error) {
	// Create transformation context
	ctx := types.NewTransformContext(sourcePath, variables)
	ctx.AddBuiltInVariables()
	
	// Replace variables in template
	result := template
	for key, value := range ctx.Variables {
		placeholder := fmt.Sprintf("${%s}", key)
		result = strings.ReplaceAll(result, placeholder, value)
	}
	
	// Check for unreplaced variables
	if strings.Contains(result, "${") {
		// Extract unreplaced variable names for better error message
		unreplaced := extractUnreplacedVars(result)
		if len(unreplaced) > 0 {
			return "", fmt.Errorf("unreplaced variables in template: %v", unreplaced)
		}
	}
	
	return result, nil
}

// extractUnreplacedVars extracts variable names that weren't replaced
func extractUnreplacedVars(s string) []string {
	var unreplaced []string
	re := regexp.MustCompile(`\$\{([^}]+)\}`)
	matches := re.FindAllStringSubmatch(s, -1)
	for _, match := range matches {
		if len(match) > 1 {
			unreplaced = append(unreplaced, match[1])
		}
	}
	return unreplaced
}

// MessageTemplater handles message template rendering
type MessageTemplater interface {
	RenderCommitMessage(template string, ctx *types.MessageContext) string
	RenderPRTitle(template string, ctx *types.MessageContext) string
	RenderPRBody(template string, ctx *types.MessageContext) string
}

// DefaultMessageTemplater implements the MessageTemplater interface
type DefaultMessageTemplater struct{}

// NewMessageTemplater creates a new message templater
func NewMessageTemplater() MessageTemplater {
	return &DefaultMessageTemplater{}
}

// RenderCommitMessage renders a commit message template
func (mt *DefaultMessageTemplater) RenderCommitMessage(template string, ctx *types.MessageContext) string {
	if template == "" {
		return fmt.Sprintf("Update code examples from %s", ctx.SourceRepo)
	}
	return mt.render(template, ctx)
}

// RenderPRTitle renders a PR title template
func (mt *DefaultMessageTemplater) RenderPRTitle(template string, ctx *types.MessageContext) string {
	if template == "" {
		return fmt.Sprintf("Update code examples from %s", ctx.SourceRepo)
	}
	return mt.render(template, ctx)
}

// RenderPRBody renders a PR body template
func (mt *DefaultMessageTemplater) RenderPRBody(template string, ctx *types.MessageContext) string {
	if template == "" {
		return fmt.Sprintf("Automated update of %d file(s) from %s (PR #%d)", 
			ctx.FileCount, ctx.SourceRepo, ctx.PRNumber)
	}
	return mt.render(template, ctx)
}

// render performs the actual template rendering
func (mt *DefaultMessageTemplater) render(template string, ctx *types.MessageContext) string {
	result := template
	
	// Built-in context variables
	replacements := map[string]string{
		"${rule_name}":     ctx.RuleName,
		"${source_repo}":   ctx.SourceRepo,
		"${target_repo}":   ctx.TargetRepo,
		"${source_branch}": ctx.SourceBranch,
		"${target_branch}": ctx.TargetBranch,
		"${file_count}":    fmt.Sprintf("%d", ctx.FileCount),
		"${pr_number}":     fmt.Sprintf("%d", ctx.PRNumber),
		"${commit_sha}":    ctx.CommitSHA,
	}
	
	// Apply built-in replacements
	for placeholder, value := range replacements {
		result = strings.ReplaceAll(result, placeholder, value)
	}
	
	// Apply custom variables from pattern matching
	for key, value := range ctx.Variables {
		placeholder := fmt.Sprintf("${%s}", key)
		result = strings.ReplaceAll(result, placeholder, value)
	}
	
	return result
}

// shouldExclude checks if a file path matches any of the exclude patterns
func (pm *DefaultPatternMatcher) shouldExclude(filePath string, excludePatterns []string) bool {
	if len(excludePatterns) == 0 {
		return false
	}

	for _, excludePattern := range excludePatterns {
		// Compile and match the exclude pattern
		re, err := regexp.Compile(excludePattern)
		if err != nil {
			// If the pattern is invalid, log and skip it
			// (validation should have caught this earlier)
			continue
		}

		if re.MatchString(filePath) {
			return true
		}
	}

	return false
}

