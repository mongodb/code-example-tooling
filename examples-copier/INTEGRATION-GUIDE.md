# Integration Guide

This guide explains how to integrate the new refactored features into the existing examples-copier application.

## Step 1: Update ServiceContainer

The `ServiceContainer` needs to include the new services. Add these fields:

```go
// In services/service_container.go or similar

type ServiceContainer struct {
    Config            *configs.Config
    WebServer         *WebServer
    WebhookService    *WebhookService
    FileStateService  FileStateService
    
    // New services
    ConfigLoader      ConfigLoader
    PatternMatcher    PatternMatcher
    PathTransformer   PathTransformer
    MessageTemplater  MessageTemplater
    AuditLogger       AuditLogger
    MetricsCollector  *MetricsCollector
}

func NewServiceContainer(config *configs.Config) (*ServiceContainer, error) {
    // Initialize new services
    configLoader := NewConfigLoader()
    patternMatcher := NewPatternMatcher()
    pathTransformer := NewPathTransformer()
    messageTemplater := NewMessageTemplater()
    metricsCollector := NewMetricsCollector()
    
    // Initialize audit logger
    ctx := context.Background()
    auditLogger, err := NewMongoAuditLogger(
        ctx,
        config.MongoURI,
        config.AuditDatabase,
        config.AuditCollection,
        config.AuditEnabled,
    )
    if err != nil {
        return nil, fmt.Errorf("failed to initialize audit logger: %w", err)
    }
    
    // ... rest of initialization
    
    return &ServiceContainer{
        Config:           config,
        ConfigLoader:     configLoader,
        PatternMatcher:   patternMatcher,
        PathTransformer:  pathTransformer,
        MessageTemplater: messageTemplater,
        AuditLogger:      auditLogger,
        MetricsCollector: metricsCollector,
        // ... other services
    }, nil
}
```

## Step 2: Update Web Server Routes

Add the new health and metrics endpoints:

```go
// In services/web_server.go

func NewHTTPHandlerWithConfig(config *configs.Config, webhookService *WebhookService, 
    metricsCollector *MetricsCollector, fileStateService FileStateService) http.Handler {
    
    mux := http.NewServeMux()
    
    // Existing webhook endpoint
    mux.HandleFunc(config.WebserverPath, func(w http.ResponseWriter, r *http.Request) {
        // Record webhook received
        metricsCollector.RecordWebhookReceived()
        startTime := time.Now()
        
        baseCtx, rid := WithRequestID(r)
        ctx, cancel := context.WithTimeout(baseCtx, time.Duration(config.RequestTimeoutSeconds)*time.Second)
        defer cancel()
        
        r = r.WithContext(ctx)
        webhookService.HandleWebhook(w, r)
        
        // Record processing time
        metricsCollector.RecordWebhookProcessed(time.Since(startTime))
    })
    
    // New health endpoint
    mux.HandleFunc("/health", HealthHandler(fileStateService, time.Now()))
    
    // New metrics endpoint
    if config.MetricsEnabled {
        mux.HandleFunc("/metrics", MetricsHandler(metricsCollector, fileStateService))
    }
    
    return mux
}
```

## Step 3: Update Config Loading

Replace the old config loading with the new loader:

```go
// In services/webhook_handler.go or wherever config is loaded

func handlePrClosedEventWithConfig(ctx context.Context, prNumber int, sourceCommitSHA string, 
    config *configs.Config, container *ServiceContainer) {
    
    // Use new config loader
    yamlConfig, err := container.ConfigLoader.LoadConfig(ctx, config)
    if err != nil {
        LogAndReturnError(ctx, "config_load", "failed to load config", err)
        return
    }
    
    // Get changed files
    changedFiles, err := GetFilesChangedInPRWithConfig(ctx, prNumber, config)
    if err != nil {
        return
    }
    
    // Process files with new pattern matching
    processFilesWithNewMatching(ctx, config, changedFiles, yamlConfig, container)
}
```

## Step 4: Update File Matching Logic

Replace the old `computeTargetPath` function with new pattern matching:

```go
// New function using pattern matching and path transformation

func processFilesWithNewMatching(ctx context.Context, config *configs.Config, 
    changedFiles []ChangedFile, yamlConfig *types.YAMLConfig, container *ServiceContainer) {
    
    for _, file := range changedFiles {
        for _, rule := range yamlConfig.CopyRules {
            // Match file against pattern
            matchResult := container.PatternMatcher.Match(file.Path, rule.SourcePattern)
            if !matchResult.Matched {
                continue
            }
            
            // Record matched file
            container.MetricsCollector.RecordFileMatched()
            
            // Process each target
            for _, target := range rule.Targets {
                // Transform path
                targetPath, err := container.PathTransformer.Transform(
                    file.Path, 
                    target.PathTransform, 
                    matchResult.Variables,
                )
                if err != nil {
                    LogError(ctx, "path_transform", "failed to transform path", err)
                    continue
                }
                
                if file.Status == "DELETED" {
                    // Handle deprecation
                    handleDeprecation(ctx, file, rule, target, targetPath, container)
                } else {
                    // Handle copy
                    handleFileCopy(ctx, file, rule, target, targetPath, matchResult.Variables, container)
                }
            }
        }
    }
}
```

## Step 5: Add Audit Logging

Integrate audit logging into file operations:

```go
func handleFileCopy(ctx context.Context, file ChangedFile, rule types.CopyRule, 
    target types.TargetConfig, targetPath string, variables map[string]string, 
    container *ServiceContainer) {
    
    startTime := time.Now()
    
    // Retrieve file content
    fc, err := RetrieveFileContentsWithConfigAndBranch(ctx, file.Path, target.Branch, container.Config)
    if err != nil {
        // Log error event
        container.AuditLogger.LogErrorEvent(ctx, &AuditEvent{
            RuleName:     rule.Name,
            SourceRepo:   container.Config.RepoOwner + "/" + container.Config.RepoName,
            SourcePath:   file.Path,
            TargetRepo:   target.Repo,
            TargetPath:   targetPath,
            Success:      false,
            ErrorMessage: err.Error(),
            DurationMs:   time.Since(startTime).Milliseconds(),
        })
        container.MetricsCollector.RecordFileUploadFailed()
        return
    }
    
    // Queue for upload
    fc.Name = github.String(targetPath)
    queueFileForUpload(target, *fc, rule, container)
    
    // Log successful copy event
    container.AuditLogger.LogCopyEvent(ctx, &AuditEvent{
        RuleName:   rule.Name,
        SourceRepo: container.Config.RepoOwner + "/" + container.Config.RepoName,
        SourcePath: file.Path,
        TargetRepo: target.Repo,
        TargetPath: targetPath,
        Success:    true,
        DurationMs: time.Since(startTime).Milliseconds(),
        FileSize:   int64(len(*fc.Content)),
    })
    
    container.MetricsCollector.RecordFileUploaded(time.Since(startTime))
}

func handleDeprecation(ctx context.Context, file ChangedFile, rule types.CopyRule, 
    target types.TargetConfig, targetPath string, container *ServiceContainer) {
    
    // Add to deprecation queue
    addToDeprecationMap(targetPath, target, container.FileStateService)
    
    // Log deprecation event
    container.AuditLogger.LogDeprecationEvent(ctx, &AuditEvent{
        RuleName:   rule.Name,
        SourceRepo: container.Config.RepoOwner + "/" + container.Config.RepoName,
        SourcePath: file.Path,
        TargetRepo: target.Repo,
        TargetPath: targetPath,
        Success:    true,
    })
    
    container.MetricsCollector.RecordFileDeprecated()
}
```

## Step 6: Add Message Templating

Use the message templater for commit messages and PR titles:

```go
func createCommitOrPR(ctx context.Context, rule types.CopyRule, target types.TargetConfig, 
    files []github.RepositoryContent, variables map[string]string, container *ServiceContainer) {
    
    // Create message context
    msgCtx := types.NewMessageContext()
    msgCtx.RuleName = rule.Name
    msgCtx.SourceRepo = container.Config.RepoOwner + "/" + container.Config.RepoName
    msgCtx.TargetRepo = target.Repo
    msgCtx.TargetBranch = target.Branch
    msgCtx.FileCount = len(files)
    msgCtx.Variables = variables
    
    // Render messages
    commitMsg := container.MessageTemplater.RenderCommitMessage(
        target.CommitStrategy.CommitMessage, 
        msgCtx,
    )
    
    if target.CommitStrategy.Type == "pull_request" {
        prTitle := container.MessageTemplater.RenderPRTitle(
            target.CommitStrategy.PRTitle,
            msgCtx,
        )
        prBody := container.MessageTemplater.RenderPRBody(
            target.CommitStrategy.PRBody,
            msgCtx,
        )
        
        // Create PR with rendered messages
        createPullRequest(ctx, target.Repo, target.Branch, prTitle, prBody, files)
    } else {
        // Create direct commit with rendered message
        createDirectCommit(ctx, target.Repo, target.Branch, commitMsg, files)
    }
}
```

## Step 7: Update go.mod

Run these commands to download the new dependencies:

```bash
cd examples-copier
go get gopkg.in/yaml.v3
go get go.mongodb.org/mongo-driver
go mod tidy
```

## Step 8: Build and Test

```bash
# Build the main application
go build -o examples-copier .

# Build the CLI tool
go build -o config-validator ./cmd/config-validator

# Test with dry-run mode
DRY_RUN=true ./examples-copier -env ./configs/.env.test

# Validate a config file
./config-validator validate -config configs/config.example.yaml -v

# Test pattern matching
./config-validator test-pattern \
  -type regex \
  -pattern "^examples/(?P<lang>[^/]+)/(?P<file>.+)$" \
  -file "examples/go/main.go"
```

## Step 9: Environment Configuration

Update your `.env` file with new options:

```bash
# Enable new features
DRY_RUN=false
AUDIT_ENABLED=true
MONGO_URI=mongodb+srv://user:pass@cluster.mongodb.net
METRICS_ENABLED=true

# Use YAML config
CONFIG_FILE=config.yaml
```

## Testing Checklist

- [ ] Config validation works for both YAML and JSON
- [ ] Pattern matching works (prefix, glob, regex)
- [ ] Path transformations work with variables
- [ ] Message templates render correctly
- [ ] Audit events are logged to MongoDB
- [ ] /health endpoint returns correct status
- [ ] /metrics endpoint returns metrics
- [ ] Dry-run mode prevents actual changes
- [ ] CLI tool validates configs correctly

## Rollback Plan

If issues arise, you can rollback by:

1. Reverting to JSON config files
2. Setting `AUDIT_ENABLED=false`
3. Using the old config loading logic
4. The new code maintains backward compatibility

## Support

For questions or issues:
1. Check `REFACTORING-SUMMARY.md` for feature documentation
2. Review example configs in `configs/config.example.yaml`
3. Use CLI tool to test patterns and transformations
4. Check audit logs in MongoDB for debugging

