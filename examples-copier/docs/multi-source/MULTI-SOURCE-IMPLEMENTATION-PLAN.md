# Multi-Source Repository Support - Implementation Plan

## Executive Summary

This document outlines the implementation plan for adding support for multiple source repositories to the examples-copier application. Currently, the application supports only a single source repository defined in the configuration. This enhancement will allow the copier to monitor and process webhooks from multiple source repositories, each with their own copy rules and configurations.

## Current Architecture Analysis

### Current Limitations

1. **Single Source Repository**: The configuration schema (`YAMLConfig`) has a single `source_repo` and `source_branch` field at the root level
2. **Hardcoded Repository Context**: Environment variables `REPO_OWNER` and `REPO_NAME` are set globally and used throughout the codebase
3. **Webhook Validation**: The webhook handler validates that incoming webhooks match the configured `source_repo` (lines 228-236 in `webhook_handler_new.go`)
4. **Config File Location**: Configuration is fetched from the single source repository defined in environment variables
5. **GitHub App Installation**: Single installation ID is configured globally

### Current Flow

```
Webhook Received → Validate Source Repo → Load Config from Source Repo → Process Files → Copy to Targets
```

## Proposed Architecture

### New Multi-Source Flow

```
Webhook Received → Identify Source Repo → Load Config for That Source → Process Files → Copy to Targets
```

### Key Design Decisions

1. **Configuration Storage**: Support both centralized (single config file) and distributed (per-repo config) approaches
2. **Backward Compatibility**: Maintain support for existing single-source configurations
3. **GitHub App Installations**: Support multiple installation IDs for different organizations
4. **Config Discovery**: Allow configs to be stored in a central location or in each source repository

## Implementation Tasks

### 1. Configuration Schema Updates

**Files to Modify:**
- `types/config.go`
- `configs/copier-config.example.yaml`

**Changes:**

#### Option A: Centralized Multi-Source Config (Recommended)
```yaml
# New schema supporting multiple sources
sources:
  - repo: "mongodb/docs-code-examples"
    branch: "main"
    installation_id: "12345678"  # Optional, falls back to default
    copy_rules:
      - name: "go-examples"
        source_pattern:
          type: "prefix"
          pattern: "examples/go/"
        targets:
          - repo: "mongodb/docs"
            branch: "main"
            path_transform: "code/${path}"
            commit_strategy:
              type: "direct"
  
  - repo: "mongodb/atlas-examples"
    branch: "main"
    installation_id: "87654321"  # Different installation for different org
    copy_rules:
      - name: "atlas-cli-examples"
        source_pattern:
          type: "glob"
          pattern: "cli/**/*.go"
        targets:
          - repo: "mongodb/atlas-cli"
            branch: "main"
            path_transform: "examples/${filename}"
            commit_strategy:
              type: "pull_request"
              pr_title: "Update examples"
              auto_merge: false

# Global defaults (optional)
defaults:
  commit_strategy:
    type: "pull_request"
    auto_merge: false
  deprecation_check:
    enabled: true
    file: "deprecated_examples.json"
```

#### Option B: Backward Compatible (Single Source at Root)
```yaml
# Backward compatible - if source_repo exists at root, treat as single source
source_repo: "mongodb/docs-code-examples"
source_branch: "main"
copy_rules:
  - name: "example"
    # ... existing structure

# OR use new multi-source structure
sources:
  - repo: "mongodb/docs-code-examples"
    # ... as above
```

**New Types:**
```go
// MultiSourceConfig represents configuration for multiple source repositories
type MultiSourceConfig struct {
    Sources  []SourceConfig  `yaml:"sources" json:"sources"`
    Defaults *DefaultsConfig `yaml:"defaults,omitempty" json:"defaults,omitempty"`
}

// SourceConfig represents a single source repository configuration
type SourceConfig struct {
    Repo           string     `yaml:"repo" json:"repo"`
    Branch         string     `yaml:"branch" json:"branch"`
    InstallationID string     `yaml:"installation_id,omitempty" json:"installation_id,omitempty"`
    ConfigFile     string     `yaml:"config_file,omitempty" json:"config_file,omitempty"` // For distributed configs
    CopyRules      []CopyRule `yaml:"copy_rules" json:"copy_rules"`
}

// DefaultsConfig provides default values for all sources
type DefaultsConfig struct {
    CommitStrategy   *CommitStrategyConfig `yaml:"commit_strategy,omitempty" json:"commit_strategy,omitempty"`
    DeprecationCheck *DeprecationConfig    `yaml:"deprecation_check,omitempty" json:"deprecation_check,omitempty"`
}

// Update YAMLConfig to support both formats
type YAMLConfig struct {
    // Legacy single-source fields (for backward compatibility)
    SourceRepo   string     `yaml:"source_repo,omitempty" json:"source_repo,omitempty"`
    SourceBranch string     `yaml:"source_branch,omitempty" json:"source_branch,omitempty"`
    CopyRules    []CopyRule `yaml:"copy_rules,omitempty" json:"copy_rules,omitempty"`
    
    // New multi-source fields
    Sources  []SourceConfig  `yaml:"sources,omitempty" json:"sources,omitempty"`
    Defaults *DefaultsConfig `yaml:"defaults,omitempty" json:"defaults,omitempty"`
}
```

### 2. Configuration Loading & Validation

**Files to Modify:**
- `services/config_loader.go`

**Changes:**

1. **Add Config Discovery Method**:
```go
// ConfigDiscovery determines where to load config from
type ConfigDiscovery interface {
    // DiscoverConfig finds the config for a given source repository
    DiscoverConfig(ctx context.Context, repoOwner, repoName string) (*SourceConfig, error)
}
```

2. **Update LoadConfig Method**:
```go
// LoadConfigForSource loads configuration for a specific source repository
func (cl *DefaultConfigLoader) LoadConfigForSource(ctx context.Context, repoOwner, repoName string, config *configs.Config) (*SourceConfig, error) {
    // Load the main config (centralized or from the source repo)
    yamlConfig, err := cl.LoadConfig(ctx, config)
    if err != nil {
        return nil, err
    }
    
    // Find the matching source configuration
    sourceRepo := fmt.Sprintf("%s/%s", repoOwner, repoName)
    sourceConfig := findSourceConfig(yamlConfig, sourceRepo)
    if sourceConfig == nil {
        return nil, fmt.Errorf("no configuration found for source repository: %s", sourceRepo)
    }
    
    return sourceConfig, nil
}

// findSourceConfig searches for a source repo in the config
func findSourceConfig(config *YAMLConfig, sourceRepo string) *SourceConfig {
    // Check if using legacy single-source format
    if config.SourceRepo != "" && config.SourceRepo == sourceRepo {
        return &SourceConfig{
            Repo:      config.SourceRepo,
            Branch:    config.SourceBranch,
            CopyRules: config.CopyRules,
        }
    }
    
    // Search in multi-source format
    for _, source := range config.Sources {
        if source.Repo == sourceRepo {
            return &source
        }
    }
    
    return nil
}
```

3. **Add Validation for Multi-Source**:
```go
func (c *YAMLConfig) Validate() error {
    // Check if using legacy or new format
    isLegacy := c.SourceRepo != ""
    isMultiSource := len(c.Sources) > 0
    
    if isLegacy && isMultiSource {
        return fmt.Errorf("cannot use both legacy (source_repo) and new (sources) format")
    }
    
    if !isLegacy && !isMultiSource {
        return fmt.Errorf("must specify either source_repo or sources")
    }
    
    if isLegacy {
        return c.validateLegacyFormat()
    }
    
    return c.validateMultiSourceFormat()
}

func (c *YAMLConfig) validateMultiSourceFormat() error {
    if len(c.Sources) == 0 {
        return fmt.Errorf("at least one source repository is required")
    }
    
    // Check for duplicate source repos
    seen := make(map[string]bool)
    for i, source := range c.Sources {
        if source.Repo == "" {
            return fmt.Errorf("sources[%d]: repo is required", i)
        }
        if seen[source.Repo] {
            return fmt.Errorf("sources[%d]: duplicate source repository: %s", i, source.Repo)
        }
        seen[source.Repo] = true
        
        if err := validateSourceConfig(&source); err != nil {
            return fmt.Errorf("sources[%d]: %w", i, err)
        }
    }
    
    return nil
}
```

### 3. Webhook Routing Logic

**Files to Modify:**
- `services/webhook_handler_new.go`
- `services/github_auth.go`

**Changes:**

1. **Update Webhook Handler**:
```go
// handleMergedPRWithContainer processes a merged PR using the new pattern matching system
func handleMergedPRWithContainer(ctx context.Context, prNumber int, sourceCommitSHA string, repoOwner string, repoName string, config *configs.Config, container *ServiceContainer) {
    startTime := time.Now()

    // Configure GitHub permissions for the source repository
    if InstallationAccessToken == "" {
        ConfigurePermissions()
    }

    // Update config with actual repository from webhook
    config.RepoOwner = repoOwner
    config.RepoName = repoName

    // Load configuration for this specific source repository
    sourceConfig, err := container.ConfigLoader.LoadConfigForSource(ctx, repoOwner, repoName, config)
    if err != nil {
        LogAndReturnError(ctx, "config_load", fmt.Sprintf("no configuration found for source repo %s/%s", repoOwner, repoName), err)
        container.MetricsCollector.RecordWebhookFailed()
        
        container.SlackNotifier.NotifyError(ctx, &ErrorEvent{
            Operation:  "config_load",
            Error:      err,
            PRNumber:   prNumber,
            SourceRepo: fmt.Sprintf("%s/%s", repoOwner, repoName),
        })
        return
    }

    // Switch GitHub installation if needed
    if sourceConfig.InstallationID != "" && sourceConfig.InstallationID != config.InstallationId {
        if err := switchGitHubInstallation(sourceConfig.InstallationID); err != nil {
            LogAndReturnError(ctx, "installation_switch", "failed to switch GitHub installation", err)
            container.MetricsCollector.RecordWebhookFailed()
            return
        }
    }

    // Continue with existing processing logic...
    // Process files with pattern matching for this source
    processFilesWithPatternMatching(ctx, prNumber, sourceCommitSHA, changedFiles, sourceConfig, config, container)
}
```

2. **Add Installation Switching**:
```go
// switchGitHubInstallation switches to a different GitHub App installation
func switchGitHubInstallation(installationID string) error {
    // Save current installation ID
    previousInstallationID := os.Getenv(configs.InstallationId)
    
    // Set new installation ID
    os.Setenv(configs.InstallationId, installationID)
    
    // Clear cached token to force re-authentication
    InstallationAccessToken = ""
    
    // Re-configure permissions with new installation
    ConfigurePermissions()
    
    LogInfo(fmt.Sprintf("Switched GitHub installation from %s to %s", previousInstallationID, installationID))
    return nil
}
```

### 4. GitHub App Installation Support

**Files to Modify:**
- `configs/environment.go`
- `services/github_auth.go`

**Changes:**

1. **Support Multiple Installation IDs**:
```go
// Config struct update
type Config struct {
    // ... existing fields
    
    // Multi-installation support
    InstallationId         string            // Default installation ID
    InstallationMapping    map[string]string // Map of repo -> installation_id
}

// Load installation mapping from environment or config
func (c *Config) GetInstallationID(repo string) string {
    if id, ok := c.InstallationMapping[repo]; ok {
        return id
    }
    return c.InstallationId // fallback to default
}
```

2. **Update Authentication**:
```go
// ConfigurePermissionsForRepo configures GitHub permissions for a specific repository
func ConfigurePermissionsForRepo(installationID string) error {
    if installationID == "" {
        return fmt.Errorf("installation ID is required")
    }
    
    // Use the provided installation ID
    token, err := generateInstallationToken(installationID)
    if err != nil {
        return fmt.Errorf("failed to generate installation token: %w", err)
    }
    
    InstallationAccessToken = token
    return nil
}
```

### 5. Metrics & Audit Logging Updates

**Files to Modify:**
- `services/health_metrics.go`
- `services/audit_logger.go`

**Changes:**

1. **Add Source Repository to Metrics**:
```go
// MetricsCollector update
type MetricsCollector struct {
    // ... existing fields
    
    // Per-source metrics
    webhooksBySource    map[string]int64
    filesBySource       map[string]int64
    uploadsBySource     map[string]int64
    mu                  sync.RWMutex
}

func (mc *MetricsCollector) RecordWebhookReceivedForSource(sourceRepo string) {
    mc.mu.Lock()
    defer mc.mu.Unlock()
    mc.webhooksReceived++
    mc.webhooksBySource[sourceRepo]++
}

func (mc *MetricsCollector) GetMetricsBySource() map[string]SourceMetrics {
    mc.mu.RLock()
    defer mc.mu.RUnlock()
    
    result := make(map[string]SourceMetrics)
    for source := range mc.webhooksBySource {
        result[source] = SourceMetrics{
            Webhooks: mc.webhooksBySource[source],
            Files:    mc.filesBySource[source],
            Uploads:  mc.uploadsBySource[source],
        }
    }
    return result
}
```

2. **Update Audit Events**:
```go
// AuditEvent already has SourceRepo field, just ensure it's populated correctly
// in all logging calls with the actual source repository
```

### 6. Documentation Updates

**Files to Create/Modify:**
- `docs/MULTI-SOURCE-GUIDE.md` (new)
- `docs/CONFIGURATION-GUIDE.md` (update)
- `README.md` (update)
- `configs/copier-config.example.yaml` (update with multi-source example)

### 7. Testing & Validation

**Files to Create:**
- `services/config_loader_multi_test.go`
- `services/webhook_handler_multi_test.go`
- `test-payloads/multi-source-webhook.json`

**Test Scenarios:**
1. Load multi-source configuration
2. Validate configuration with multiple sources
3. Route webhook to correct source configuration
4. Handle missing source repository gracefully
5. Switch between GitHub installations
6. Backward compatibility with single-source configs

### 8. Migration Guide & Backward Compatibility

**Backward Compatibility Strategy:**

1. **Auto-detect Format**: Check if `source_repo` exists at root level
2. **Convert Legacy to New**: Internally convert single-source to multi-source format
3. **Validation**: Ensure both formats validate correctly
4. **Migration Tool**: Provide CLI command to convert configs

```bash
# Convert legacy config to multi-source format
./config-validator convert-to-multi-source -input copier-config.yaml -output copier-config-multi.yaml
```

## Implementation Phases

### Phase 1: Core Infrastructure (Week 1)
- [ ] Update configuration schema
- [ ] Implement config loading for multiple sources
- [ ] Add validation for multi-source configs
- [ ] Ensure backward compatibility

### Phase 2: Webhook Routing (Week 2)
- [ ] Implement webhook routing logic
- [ ] Add GitHub installation switching
- [ ] Update authentication handling
- [ ] Test with multiple source repos

### Phase 3: Observability (Week 3)
- [ ] Update metrics collection
- [ ] Enhance audit logging
- [ ] Add per-source monitoring
- [ ] Update health endpoints

### Phase 4: Documentation & Testing (Week 4)
- [ ] Write comprehensive documentation
- [ ] Create migration guide
- [ ] Add unit and integration tests
- [ ] Perform end-to-end testing

## Risks & Mitigation

### Risk 1: Breaking Changes
**Mitigation**: Maintain full backward compatibility with legacy single-source format

### Risk 2: GitHub Rate Limits
**Mitigation**: Implement per-source rate limiting and monitoring

### Risk 3: Configuration Complexity
**Mitigation**: Provide clear examples, templates, and validation tools

### Risk 4: Installation Token Management
**Mitigation**: Implement proper token caching and refresh logic per installation

## Success Criteria

1. ✅ Support multiple source repositories in a single deployment
2. ✅ Maintain 100% backward compatibility with existing configs
3. ✅ No performance degradation for single-source use cases
4. ✅ Clear documentation and migration path
5. ✅ Comprehensive test coverage (>80%)
6. ✅ Successful deployment with 2+ source repositories

## Future Enhancements

1. **Dynamic Config Reloading**: Reload configuration without restart
2. **Per-Source Webhooks**: Different webhook endpoints for different sources
3. **Source Repository Discovery**: Auto-discover repositories with copier configs
4. **Config Validation API**: REST API for validating configurations
5. **Multi-Tenant Support**: Support multiple organizations with isolated configs

