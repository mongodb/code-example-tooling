# Multi-Source Repository Support - Technical Specification

## Document Information

- **Version**: 1.0
- **Status**: Draft
- **Last Updated**: 2025-10-15
- **Author**: Examples Copier Team

## 1. Overview

### 1.1 Purpose

This document provides detailed technical specifications for implementing multi-source repository support in the examples-copier application.

### 1.2 Scope

The implementation will enable the copier to:
- Monitor multiple source repositories simultaneously
- Route webhooks to appropriate source configurations
- Manage multiple GitHub App installations
- Maintain backward compatibility with existing single-source configurations

### 1.3 Goals

- **Primary**: Support multiple source repositories in a single deployment
- **Secondary**: Improve observability with per-source metrics
- **Tertiary**: Simplify deployment and reduce infrastructure costs

## 2. System Architecture

### 2.1 Current Architecture Limitations

```
Current Flow (Single Source):
┌─────────────────┐
│  Source Repo    │
│  (hardcoded)    │
└────────┬────────┘
         │ Webhook
         ▼
┌─────────────────┐
│ Webhook Handler │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Load Config    │
│  (from source)  │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Process Files   │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Target Repos    │
└─────────────────┘
```

### 2.2 Proposed Architecture

```
New Flow (Multi-Source):
┌──────────┐  ┌──────────┐  ┌──────────┐
│ Source 1 │  │ Source 2 │  │ Source 3 │
└────┬─────┘  └────┬─────┘  └────┬─────┘
     │ Webhook     │ Webhook     │ Webhook
     └─────────────┴─────────────┘
                   │
                   ▼
         ┌─────────────────┐
         │ Webhook Router  │
         │ (new component) │
         └────────┬────────┘
                  │
                  ▼
         ┌─────────────────┐
         │  Config Loader  │
         │  (enhanced)     │
         └────────┬────────┘
                  │
         ┌────────┴────────┐
         │                 │
         ▼                 ▼
    ┌─────────┐      ┌─────────┐
    │Config 1 │      │Config 2 │
    └────┬────┘      └────┬────┘
         │                │
         └────────┬───────┘
                  │
                  ▼
         ┌─────────────────┐
         │ Process Files   │
         └────────┬────────┘
                  │
         ┌────────┴────────┐
         │                 │
         ▼                 ▼
    ┌─────────┐      ┌─────────┐
    │Target 1 │      │Target 2 │
    └─────────┘      └─────────┘
```

## 3. Data Models

### 3.1 Configuration Schema

#### 3.1.1 MultiSourceConfig

```go
// MultiSourceConfig represents the root configuration
type MultiSourceConfig struct {
    // New multi-source format
    Sources  []SourceConfig  `yaml:"sources,omitempty" json:"sources,omitempty"`
    Defaults *DefaultsConfig `yaml:"defaults,omitempty" json:"defaults,omitempty"`
    
    // Legacy single-source format (for backward compatibility)
    SourceRepo   string     `yaml:"source_repo,omitempty" json:"source_repo,omitempty"`
    SourceBranch string     `yaml:"source_branch,omitempty" json:"source_branch,omitempty"`
    CopyRules    []CopyRule `yaml:"copy_rules,omitempty" json:"copy_rules,omitempty"`
}
```

#### 3.1.2 SourceConfig

```go
// SourceConfig represents a single source repository
type SourceConfig struct {
    // Repository identifier (owner/repo format)
    Repo string `yaml:"repo" json:"repo"`
    
    // Branch to monitor (default: "main")
    Branch string `yaml:"branch" json:"branch"`
    
    // GitHub App installation ID for this repository
    // Optional: falls back to default INSTALLATION_ID
    InstallationID string `yaml:"installation_id,omitempty" json:"installation_id,omitempty"`
    
    // Path to config file in the repository
    // Optional: for distributed config approach
    ConfigFile string `yaml:"config_file,omitempty" json:"config_file,omitempty"`
    
    // Copy rules for this source
    CopyRules []CopyRule `yaml:"copy_rules" json:"copy_rules"`
    
    // Source-specific settings
    Settings *SourceSettings `yaml:"settings,omitempty" json:"settings,omitempty"`
}
```

#### 3.1.3 SourceSettings

```go
// SourceSettings contains source-specific configuration
type SourceSettings struct {
    // Enable/disable this source
    Enabled bool `yaml:"enabled" json:"enabled"`
    
    // Timeout for processing webhooks from this source
    TimeoutSeconds int `yaml:"timeout_seconds,omitempty" json:"timeout_seconds,omitempty"`
    
    // Rate limiting settings
    RateLimit *RateLimitConfig `yaml:"rate_limit,omitempty" json:"rate_limit,omitempty"`
}

// RateLimitConfig defines rate limiting per source
type RateLimitConfig struct {
    // Maximum webhooks per minute
    MaxWebhooksPerMinute int `yaml:"max_webhooks_per_minute" json:"max_webhooks_per_minute"`
    
    // Maximum concurrent processing
    MaxConcurrent int `yaml:"max_concurrent" json:"max_concurrent"`
}
```

#### 3.1.4 DefaultsConfig

```go
// DefaultsConfig provides default values for all sources
type DefaultsConfig struct {
    CommitStrategy   *CommitStrategyConfig `yaml:"commit_strategy,omitempty" json:"commit_strategy,omitempty"`
    DeprecationCheck *DeprecationConfig    `yaml:"deprecation_check,omitempty" json:"deprecation_check,omitempty"`
    Settings         *SourceSettings       `yaml:"settings,omitempty" json:"settings,omitempty"`
}
```

### 3.2 Runtime Data Structures

#### 3.2.1 SourceContext

```go
// SourceContext holds runtime context for a source repository
type SourceContext struct {
    // Source configuration
    Config *SourceConfig
    
    // GitHub client for this source
    GitHubClient *github.Client
    
    // Installation token
    InstallationToken string
    
    // Token expiration time
    TokenExpiry time.Time
    
    // Metrics for this source
    Metrics *SourceMetrics
    
    // Last processed webhook timestamp
    LastWebhook time.Time
}
```

#### 3.2.2 SourceMetrics

```go
// SourceMetrics tracks metrics per source repository
type SourceMetrics struct {
    SourceRepo string
    
    // Webhook metrics
    WebhooksReceived  int64
    WebhooksProcessed int64
    WebhooksFailed    int64
    
    // File metrics
    FilesMatched      int64
    FilesUploaded     int64
    FilesUploadFailed int64
    FilesDeprecated   int64
    
    // Timing metrics
    AvgProcessingTime time.Duration
    MaxProcessingTime time.Duration
    MinProcessingTime time.Duration
    
    // Last update
    LastUpdated time.Time
}
```

## 4. Component Specifications

### 4.1 Webhook Router

**Purpose**: Route incoming webhooks to the correct source configuration

**Interface**:
```go
type WebhookRouter interface {
    // RouteWebhook routes a webhook to the appropriate source handler
    RouteWebhook(ctx context.Context, event *github.PullRequestEvent) (*SourceConfig, error)
    
    // RegisterSource registers a source configuration
    RegisterSource(config *SourceConfig) error
    
    // UnregisterSource removes a source configuration
    UnregisterSource(repo string) error
    
    // GetSource retrieves a source configuration
    GetSource(repo string) (*SourceConfig, error)
    
    // ListSources returns all registered sources
    ListSources() []*SourceConfig
}
```

**Implementation**:
```go
type DefaultWebhookRouter struct {
    sources map[string]*SourceConfig
    mu      sync.RWMutex
}

func (r *DefaultWebhookRouter) RouteWebhook(ctx context.Context, event *github.PullRequestEvent) (*SourceConfig, error) {
    repo := event.GetRepo()
    if repo == nil {
        return nil, fmt.Errorf("webhook missing repository info")
    }
    
    repoFullName := repo.GetFullName()
    
    r.mu.RLock()
    defer r.mu.RUnlock()
    
    source, ok := r.sources[repoFullName]
    if !ok {
        return nil, fmt.Errorf("no configuration found for repository: %s", repoFullName)
    }
    
    // Check if source is enabled
    if source.Settings != nil && !source.Settings.Enabled {
        return nil, fmt.Errorf("source repository is disabled: %s", repoFullName)
    }
    
    return source, nil
}
```

### 4.2 Config Loader (Enhanced)

**Purpose**: Load and manage multi-source configurations

**New Methods**:
```go
type ConfigLoader interface {
    // Existing method
    LoadConfig(ctx context.Context, config *configs.Config) (*types.YAMLConfig, error)
    
    // New methods for multi-source
    LoadMultiSourceConfig(ctx context.Context, config *configs.Config) (*types.MultiSourceConfig, error)
    LoadSourceConfig(ctx context.Context, repo string, config *configs.Config) (*types.SourceConfig, error)
    ValidateMultiSourceConfig(config *types.MultiSourceConfig) error
    ConvertLegacyToMultiSource(legacy *types.YAMLConfig) (*types.MultiSourceConfig, error)
}
```

**Implementation**:
```go
func (cl *DefaultConfigLoader) LoadMultiSourceConfig(ctx context.Context, config *configs.Config) (*types.MultiSourceConfig, error) {
    // Load raw config
    yamlConfig, err := cl.LoadConfig(ctx, config)
    if err != nil {
        return nil, err
    }
    
    // Detect format
    if yamlConfig.SourceRepo != "" {
        // Legacy format - convert to multi-source
        return cl.ConvertLegacyToMultiSource(yamlConfig)
    }
    
    // Already multi-source format
    multiConfig := &types.MultiSourceConfig{
        Sources:  yamlConfig.Sources,
        Defaults: yamlConfig.Defaults,
    }
    
    // Validate
    if err := cl.ValidateMultiSourceConfig(multiConfig); err != nil {
        return nil, err
    }
    
    return multiConfig, nil
}

func (cl *DefaultConfigLoader) ConvertLegacyToMultiSource(legacy *types.YAMLConfig) (*types.MultiSourceConfig, error) {
    source := types.SourceConfig{
        Repo:      legacy.SourceRepo,
        Branch:    legacy.SourceBranch,
        CopyRules: legacy.CopyRules,
    }
    
    return &types.MultiSourceConfig{
        Sources: []types.SourceConfig{source},
    }, nil
}
```

### 4.3 Installation Manager

**Purpose**: Manage multiple GitHub App installations

**Interface**:
```go
type InstallationManager interface {
    // GetInstallationToken gets or refreshes token for an installation
    GetInstallationToken(ctx context.Context, installationID string) (string, error)
    
    // GetClientForInstallation gets a GitHub client for an installation
    GetClientForInstallation(ctx context.Context, installationID string) (*github.Client, error)
    
    // RefreshToken refreshes an installation token
    RefreshToken(ctx context.Context, installationID string) error
    
    // ClearCache clears cached tokens
    ClearCache()
}
```

**Implementation**:
```go
type DefaultInstallationManager struct {
    tokens map[string]*InstallationToken
    mu     sync.RWMutex
}

type InstallationToken struct {
    Token     string
    ExpiresAt time.Time
}

func (im *DefaultInstallationManager) GetInstallationToken(ctx context.Context, installationID string) (string, error) {
    im.mu.RLock()
    token, ok := im.tokens[installationID]
    im.mu.RUnlock()
    
    // Check if token exists and is not expired
    if ok && time.Now().Before(token.ExpiresAt.Add(-5*time.Minute)) {
        return token.Token, nil
    }
    
    // Generate new token
    newToken, err := generateInstallationToken(installationID)
    if err != nil {
        return "", err
    }
    
    // Cache token
    im.mu.Lock()
    im.tokens[installationID] = &InstallationToken{
        Token:     newToken,
        ExpiresAt: time.Now().Add(1 * time.Hour),
    }
    im.mu.Unlock()
    
    return newToken, nil
}
```

### 4.4 Metrics Collector (Enhanced)

**Purpose**: Track metrics per source repository

**New Methods**:
```go
type MetricsCollector interface {
    // Existing methods...
    
    // New methods for multi-source
    RecordWebhookReceivedForSource(sourceRepo string)
    RecordWebhookProcessedForSource(sourceRepo string, duration time.Duration)
    RecordWebhookFailedForSource(sourceRepo string)
    RecordFileMatchedForSource(sourceRepo string)
    RecordFileUploadedForSource(sourceRepo string)
    RecordFileUploadFailedForSource(sourceRepo string)
    
    GetMetricsBySource(sourceRepo string) *SourceMetrics
    GetAllSourceMetrics() map[string]*SourceMetrics
}
```

## 5. API Specifications

### 5.1 Enhanced Health Endpoint

**Endpoint**: `GET /health`

**Response**:
```json
{
  "status": "healthy",
  "started": true,
  "github": {
    "status": "healthy",
    "authenticated": true
  },
  "sources": {
    "mongodb/docs-code-examples": {
      "status": "healthy",
      "last_webhook": "2025-10-15T10:30:00Z",
      "installation_id": "12345678"
    },
    "mongodb/atlas-examples": {
      "status": "healthy",
      "last_webhook": "2025-10-15T10:25:00Z",
      "installation_id": "87654321"
    }
  },
  "queues": {
    "upload_count": 0,
    "deprecation_count": 0
  },
  "uptime": "2h15m30s"
}
```

### 5.2 Enhanced Metrics Endpoint

**Endpoint**: `GET /metrics`

**Response**:
```json
{
  "global": {
    "webhooks": {
      "received": 150,
      "processed": 145,
      "failed": 5,
      "success_rate": 96.67
    },
    "files": {
      "matched": 320,
      "uploaded": 310,
      "upload_failed": 5,
      "deprecated": 5
    }
  },
  "by_source": {
    "mongodb/docs-code-examples": {
      "webhooks": {
        "received": 100,
        "processed": 98,
        "failed": 2
      },
      "files": {
        "matched": 200,
        "uploaded": 195,
        "upload_failed": 3
      },
      "last_webhook": "2025-10-15T10:30:00Z"
    },
    "mongodb/atlas-examples": {
      "webhooks": {
        "received": 50,
        "processed": 47,
        "failed": 3
      },
      "files": {
        "matched": 120,
        "uploaded": 115,
        "upload_failed": 2
      },
      "last_webhook": "2025-10-15T10:25:00Z"
    }
  }
}
```

## 6. Error Handling

### 6.1 Error Scenarios

| Scenario | HTTP Status | Response | Action |
|----------|-------------|----------|--------|
| Unknown source repo | 204 No Content | Empty | Log warning, ignore webhook |
| Disabled source | 204 No Content | Empty | Log info, ignore webhook |
| Config load failure | 500 Internal Server Error | Error message | Alert, retry |
| Installation auth failure | 500 Internal Server Error | Error message | Alert, retry |
| Pattern match failure | 200 OK | Success (no files matched) | Log info |
| Upload failure | 200 OK | Success (logged as failed) | Log error, alert |

### 6.2 Error Response Format

```json
{
  "error": "configuration error",
  "message": "no configuration found for repository: mongodb/unknown-repo",
  "source_repo": "mongodb/unknown-repo",
  "timestamp": "2025-10-15T10:30:00Z",
  "request_id": "abc123"
}
```

## 7. Performance Considerations

### 7.1 Scalability

- **Concurrent Processing**: Support up to 10 concurrent webhook processing
- **Config Caching**: Cache loaded configurations for 5 minutes
- **Token Caching**: Cache installation tokens until 5 minutes before expiry
- **Rate Limiting**: Per-source rate limiting to prevent abuse

### 7.2 Resource Limits

- **Max Sources**: 50 source repositories per deployment
- **Max Copy Rules**: 100 copy rules per source
- **Max Targets**: 20 targets per copy rule
- **Config Size**: 1 MB maximum config file size

## 8. Security Considerations

### 8.1 Authentication

- Each source repository requires valid GitHub App installation
- Installation tokens are cached securely in memory
- Tokens are refreshed automatically before expiry

### 8.2 Authorization

- Verify webhook signatures for all incoming requests
- Validate source repository against configured sources
- Ensure installation has required permissions

### 8.3 Data Protection

- No sensitive data in logs
- Installation tokens never logged
- Audit logs contain only necessary information

## 9. Testing Strategy

### 9.1 Unit Tests

- Config loading and validation
- Webhook routing logic
- Installation token management
- Metrics collection

### 9.2 Integration Tests

- Multi-source webhook processing
- Installation switching
- Config format conversion
- Error handling

### 9.3 End-to-End Tests

- Complete workflow with multiple sources
- Cross-organization copying
- Failure recovery
- Performance under load

## 10. Deployment Strategy

### 10.1 Rollout Plan

1. **Phase 1**: Deploy with backward compatibility (Week 1)
2. **Phase 2**: Enable multi-source for staging (Week 2)
3. **Phase 3**: Gradual production rollout (Week 3)
4. **Phase 4**: Full production deployment (Week 4)

### 10.2 Monitoring

- Track metrics per source repository
- Alert on failures
- Monitor GitHub API rate limits
- Track installation token refresh

## 11. Appendix

### 11.1 Configuration Examples

See `configs/copier-config.multi-source.example.yaml`

### 11.2 Migration Guide

See `docs/MULTI-SOURCE-MIGRATION-GUIDE.md`

### 11.3 Implementation Plan

See `docs/MULTI-SOURCE-IMPLEMENTATION-PLAN.md`

