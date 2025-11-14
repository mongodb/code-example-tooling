# Workflow-Based Config Implementation Roadmap

## Overview

Implement Copybara-inspired workflow configuration to reduce config complexity by 70% and improve scalability.

**Timeline:** 1-2 weeks  
**Effort:** Medium  
**Risk:** Low (backward compatible)

---

## Phase 1: Core Types (Day 1-2)

### Add New Types to `types/config.go`

```go
// Workflow represents a complete copy workflow from source to one destination
type Workflow struct {
    Name            string           `yaml:"name" json:"name"`
    Destination     Destination      `yaml:"destination" json:"destination"`
    Transformations []Transformation `yaml:"transformations" json:"transformations"`
    Exclude         []string         `yaml:"exclude,omitempty" json:"exclude,omitempty"`
    CommitStrategy  *CommitStrategyConfig `yaml:"commit_strategy,omitempty" json:"commit_strategy,omitempty"`
    DeprecationCheck *DeprecationConfig   `yaml:"deprecation_check,omitempty" json:"deprecation_check,omitempty"`
}

// Destination defines where files are copied to
type Destination struct {
    Repo   string `yaml:"repo" json:"repo"`
    Branch string `yaml:"branch,omitempty" json:"branch,omitempty"` // defaults to "main"
}

// Transformation defines how to transform file paths
type Transformation struct {
    Move  *MoveTransform  `yaml:"move,omitempty" json:"move,omitempty"`
    Copy  *CopyTransform  `yaml:"copy,omitempty" json:"copy,omitempty"`
    Glob  *GlobTransform  `yaml:"glob,omitempty" json:"glob,omitempty"`
    Regex *RegexTransform `yaml:"regex,omitempty" json:"regex,omitempty"`
}

// MoveTransform moves a directory with path transformation
type MoveTransform struct {
    From string `yaml:"from" json:"from"`
    To   string `yaml:"to" json:"to"`
}

// CopyTransform copies a file or directory
type CopyTransform struct {
    From string `yaml:"from" json:"from"`
    To   string `yaml:"to" json:"to"`
}

// GlobTransform matches files with glob pattern
type GlobTransform struct {
    Pattern string `yaml:"pattern" json:"pattern"`
    To      string `yaml:"to" json:"to"`
}

// RegexTransform matches files with regex
type RegexTransform struct {
    Pattern string `yaml:"pattern" json:"pattern"`
    To      string `yaml:"to" json:"to"`
}

// Defaults holds default values for all workflows
type Defaults struct {
    CommitStrategy   *CommitStrategyConfig `yaml:"commit_strategy,omitempty" json:"commit_strategy,omitempty"`
    DeprecationCheck *DeprecationConfig    `yaml:"deprecation_check,omitempty" json:"deprecation_check,omitempty"`
    Exclude          []string              `yaml:"exclude,omitempty" json:"exclude,omitempty"`
}

// Update YAMLConfig to support both formats
type YAMLConfig struct {
    // Legacy format (backward compatible)
    SourceRepo   string     `yaml:"source_repo,omitempty" json:"source_repo,omitempty"`
    SourceBranch string     `yaml:"source_branch,omitempty" json:"source_branch,omitempty"`
    CopyRules    []CopyRule `yaml:"copy_rules,omitempty" json:"copy_rules,omitempty"`
    
    // New workflow format
    Workflows []Workflow `yaml:"workflows,omitempty" json:"workflows,omitempty"`
    Defaults  *Defaults  `yaml:"defaults,omitempty" json:"defaults,omitempty"`
}
```

### Add Validation Methods

```go
func (w *Workflow) Validate() error {
    if w.Name == "" {
        return fmt.Errorf("name is required")
    }
    if err := w.Destination.Validate(); err != nil {
        return fmt.Errorf("destination: %w", err)
    }
    if len(w.Transformations) == 0 {
        return fmt.Errorf("at least one transformation is required")
    }
    for i, t := range w.Transformations {
        if err := t.Validate(); err != nil {
            return fmt.Errorf("transformations[%d]: %w", i, err)
        }
    }
    return nil
}

func (t *Transformation) Validate() error {
    count := 0
    if t.Move != nil { count++ }
    if t.Copy != nil { count++ }
    if t.Glob != nil { count++ }
    if t.Regex != nil { count++ }
    
    if count == 0 {
        return fmt.Errorf("transformation must specify one of: move, copy, glob, regex")
    }
    if count > 1 {
        return fmt.Errorf("transformation can only specify one type")
    }
    
    // Validate specific transformation
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
```

**Deliverable:** New types in `types/config.go` with validation

---

## Phase 2: Workflow Processor (Day 3-4)

### Create `services/workflow_processor.go`

```go
package services

import (
    "context"
    "fmt"
    "strings"
    "path/filepath"
    "github.com/bmatcuk/doublestar/v4"
    "github.com/mongodb/code-example-tooling/code-copier/types"
)

// WorkflowProcessor processes workflows and matches files
type WorkflowProcessor struct {
    patternMatcher  PatternMatcher
    pathTransformer PathTransformer
}

// NewWorkflowProcessor creates a new workflow processor
func NewWorkflowProcessor() *WorkflowProcessor {
    return &WorkflowProcessor{
        patternMatcher:  NewPatternMatcher(),
        pathTransformer: NewPathTransformer(),
    }
}

// ProcessWorkflows processes all workflows for changed files
func (wp *WorkflowProcessor) ProcessWorkflows(
    ctx context.Context,
    changedFiles []types.ChangedFile,
    workflows []types.Workflow,
    defaults *types.Defaults,
) (map[string][]types.FileMapping, error) {
    
    // Map of destination repo -> file mappings
    result := make(map[string][]types.FileMapping)
    
    for _, workflow := range workflows {
        // Apply defaults
        wp.applyDefaults(&workflow, defaults)
        
        // Process each transformation
        for _, transformation := range workflow.Transformations {
            mappings := wp.processTransformation(ctx, changedFiles, transformation, workflow)
            
            // Add to result
            repoKey := workflow.Destination.Repo
            result[repoKey] = append(result[repoKey], mappings...)
        }
    }
    
    return result, nil
}

// processTransformation processes a single transformation
func (wp *WorkflowProcessor) processTransformation(
    ctx context.Context,
    changedFiles []types.ChangedFile,
    transformation types.Transformation,
    workflow types.Workflow,
) []types.FileMapping {
    
    var mappings []types.FileMapping
    
    if transformation.Move != nil {
        mappings = wp.processMove(changedFiles, transformation.Move, workflow)
    } else if transformation.Copy != nil {
        mappings = wp.processCopy(changedFiles, transformation.Copy, workflow)
    } else if transformation.Glob != nil {
        mappings = wp.processGlob(changedFiles, transformation.Glob, workflow)
    } else if transformation.Regex != nil {
        mappings = wp.processRegex(changedFiles, transformation.Regex, workflow)
    }
    
    // Apply exclusions
    return wp.applyExclusions(mappings, workflow.Exclude)
}

// processMove handles move transformations
func (wp *WorkflowProcessor) processMove(
    changedFiles []types.ChangedFile,
    move *types.MoveTransform,
    workflow types.Workflow,
) []types.FileMapping {
    
    var mappings []types.FileMapping
    prefix := strings.TrimSuffix(move.From, "/") + "/"
    
    for _, file := range changedFiles {
        if strings.HasPrefix(file.Path, prefix) {
            // Strip prefix and add new prefix
            relativePath := strings.TrimPrefix(file.Path, prefix)
            targetPath := filepath.Join(move.To, relativePath)
            
            mappings = append(mappings, types.FileMapping{
                SourcePath: file.Path,
                TargetPath: targetPath,
                TargetRepo: workflow.Destination.Repo,
                TargetBranch: workflow.Destination.Branch,
                Status: file.Status,
            })
        }
    }
    
    return mappings
}

// processCopy handles copy transformations
func (wp *WorkflowProcessor) processCopy(
    changedFiles []types.ChangedFile,
    copy *types.CopyTransform,
    workflow types.Workflow,
) []types.FileMapping {
    
    var mappings []types.FileMapping
    
    for _, file := range changedFiles {
        if file.Path == copy.From {
            mappings = append(mappings, types.FileMapping{
                SourcePath: file.Path,
                TargetPath: copy.To,
                TargetRepo: workflow.Destination.Repo,
                TargetBranch: workflow.Destination.Branch,
                Status: file.Status,
            })
        }
    }
    
    return mappings
}

// processGlob handles glob transformations
func (wp *WorkflowProcessor) processGlob(
    changedFiles []types.ChangedFile,
    glob *types.GlobTransform,
    workflow types.Workflow,
) []types.FileMapping {
    
    var mappings []types.FileMapping
    
    for _, file := range changedFiles {
        matched, _ := doublestar.Match(glob.Pattern, file.Path)
        if matched {
            // Transform path using template
            targetPath, err := wp.transformPath(file.Path, glob.To)
            if err != nil {
                continue
            }
            
            mappings = append(mappings, types.FileMapping{
                SourcePath: file.Path,
                TargetPath: targetPath,
                TargetRepo: workflow.Destination.Repo,
                TargetBranch: workflow.Destination.Branch,
                Status: file.Status,
            })
        }
    }
    
    return mappings
}

// applyExclusions filters out excluded files
func (wp *WorkflowProcessor) applyExclusions(
    mappings []types.FileMapping,
    exclusions []string,
) []types.FileMapping {
    
    if len(exclusions) == 0 {
        return mappings
    }
    
    var filtered []types.FileMapping
    for _, mapping := range mappings {
        excluded := false
        for _, pattern := range exclusions {
            matched, _ := doublestar.Match(pattern, mapping.SourcePath)
            if matched {
                excluded = true
                break
            }
        }
        if !excluded {
            filtered = append(filtered, mapping)
        }
    }
    
    return filtered
}
```

**Deliverable:** Workflow processor that converts workflows to file mappings

---

## Phase 3: Integration (Day 5)

### Update `webhook_handler_new.go`

```go
func processFilesWithWorkflows(
    ctx context.Context,
    prNumber int,
    sourceCommitSHA string,
    changedFiles []types.ChangedFile,
    yamlConfig *types.YAMLConfig,
    config *configs.Config,
    container *ServiceContainer,
) {
    
    // Check if using workflow format
    if len(yamlConfig.Workflows) > 0 {
        processor := services.NewWorkflowProcessor()
        mappings, err := processor.ProcessWorkflows(
            ctx,
            changedFiles,
            yamlConfig.Workflows,
            yamlConfig.Defaults,
        )
        if err != nil {
            LogErrorCtx(ctx, "failed to process workflows", err, nil)
            return
        }
        
        // Queue files for upload
        for repo, fileMappings := range mappings {
            for _, mapping := range fileMappings {
                queueFileForUpload(ctx, mapping, container)
            }
        }
    } else {
        // Fall back to legacy pattern matching
        processFilesWithPatternMatching(ctx, prNumber, sourceCommitSHA, changedFiles, yamlConfig, config, container)
    }
}
```

**Deliverable:** Backward-compatible integration

---

## Phase 4: Testing (Day 6-7)

### Unit Tests

```go
// services/workflow_processor_test.go
func TestWorkflowProcessor_ProcessMove(t *testing.T) {
    processor := NewWorkflowProcessor()
    
    changedFiles := []types.ChangedFile{
        {Path: "mflix/client/src/App.tsx", Status: "modified"},
        {Path: "mflix/server/java-spring/Main.java", Status: "modified"},
    }
    
    workflow := types.Workflow{
        Name: "test",
        Destination: types.Destination{
            Repo: "org/target",
            Branch: "main",
        },
        Transformations: []types.Transformation{
            {
                Move: &types.MoveTransform{
                    From: "mflix/client",
                    To: "client",
                },
            },
        },
    }
    
    mappings, err := processor.ProcessWorkflows(context.Background(), changedFiles, []types.Workflow{workflow}, nil)
    
    require.NoError(t, err)
    assert.Len(t, mappings["org/target"], 1)
    assert.Equal(t, "client/src/App.tsx", mappings["org/target"][0].TargetPath)
}
```

### Integration Tests

Test with real config files and verify backward compatibility.

**Deliverable:** Comprehensive test suite

---

## Phase 5: Config Converter Tool (Day 8)

### Create `cmd/config-converter/main.go`

```go
package main

import (
    "flag"
    "fmt"
    "os"
    "github.com/mongodb/code-example-tooling/code-copier/services"
    "github.com/mongodb/code-example-tooling/code-copier/types"
    "gopkg.in/yaml.v3"
)

func main() {
    input := flag.String("input", "", "Input config file (old format)")
    output := flag.String("output", "", "Output config file (new format)")
    flag.Parse()
    
    // Load old config
    oldConfig, err := services.LoadConfig(*input)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
        os.Exit(1)
    }
    
    // Convert to workflows
    newConfig := convertToWorkflows(oldConfig)
    
    // Write new config
    data, err := yaml.Marshal(newConfig)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error marshaling config: %v\n", err)
        os.Exit(1)
    }
    
    if err := os.WriteFile(*output, data, 0644); err != nil {
        fmt.Fprintf(os.Stderr, "Error writing config: %v\n", err)
        os.Exit(1)
    }
    
    fmt.Printf("Converted %s to %s\n", *input, *output)
}

func convertToWorkflows(oldConfig *types.YAMLConfig) *types.YAMLConfig {
    // Group rules by target repo
    workflowMap := make(map[string]*types.Workflow)
    
    for _, rule := range oldConfig.CopyRules {
        for _, target := range rule.Targets {
            workflow, exists := workflowMap[target.Repo]
            if !exists {
                workflow = &types.Workflow{
                    Name: fmt.Sprintf("workflow-%s", target.Repo),
                    Destination: types.Destination{
                        Repo: target.Repo,
                        Branch: target.Branch,
                    },
                    Transformations: []types.Transformation{},
                }
                workflowMap[target.Repo] = workflow
            }
            
            // Convert rule to transformation
            transformation := convertRuleToTransformation(rule, target)
            workflow.Transformations = append(workflow.Transformations, transformation)
        }
    }
    
    // Convert map to slice
    var workflows []types.Workflow
    for _, workflow := range workflowMap {
        workflows = append(workflows, *workflow)
    }
    
    return &types.YAMLConfig{
        SourceRepo: oldConfig.SourceRepo,
        SourceBranch: oldConfig.SourceBranch,
        Workflows: workflows,
    }
}
```

**Deliverable:** Tool to convert old configs to new format

---

## Phase 6: Documentation (Day 9)

### Update Documentation

1. **README.md** - Add workflow examples
2. **CONFIGURATION-GUIDE.md** - Document workflow format
3. **MIGRATION-GUIDE.md** - How to migrate from old format
4. **EXAMPLES.md** - Real-world workflow examples

**Deliverable:** Complete documentation

---

## Phase 7: Deployment (Day 10)

### Deployment Checklist

- [ ] All tests passing
- [ ] Convert production config
- [ ] Deploy to staging
- [ ] Test with real PRs
- [ ] Monitor logs and metrics
- [ ] Deploy to production
- [ ] Monitor for 24 hours

**Deliverable:** Production deployment

---

## Success Criteria

1. ✅ Backward compatibility - old configs still work
2. ✅ Config size reduced by 70%
3. ✅ All existing functionality preserved
4. ✅ Tests passing with >90% coverage
5. ✅ Documentation complete
6. ✅ Production deployment successful
7. ✅ No increase in error rate

---

## Risks & Mitigation

| Risk | Impact | Mitigation |
|------|--------|------------|
| Breaking changes | High | Maintain backward compatibility |
| Performance regression | Medium | Benchmark before/after |
| Config conversion errors | Medium | Extensive testing, manual review |
| Team adoption | Low | Clear documentation, examples |

---

## Timeline Summary

| Phase | Days | Deliverable |
|-------|------|-------------|
| 1. Core Types | 2 | New types with validation |
| 2. Workflow Processor | 2 | File mapping logic |
| 3. Integration | 1 | Backward-compatible handler |
| 4. Testing | 2 | Comprehensive tests |
| 5. Config Converter | 1 | Conversion tool |
| 6. Documentation | 1 | Complete docs |
| 7. Deployment | 1 | Production release |
| **Total** | **10 days** | **Simplified config system** |

---

## Next Steps

1. **Approve this roadmap** - Review and provide feedback
2. **Start implementation** - I can begin coding immediately
3. **Review PRs** - Incremental reviews as I build
4. **Test together** - Validate with your real configs
5. **Deploy** - Roll out to production

**Ready to start?** 🚀

