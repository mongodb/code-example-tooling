# Multi-Source Repository Support - Implementation Summary

## Executive Summary

This document provides a comprehensive overview of the multi-source repository support implementation plan for the examples-copier application.

## What's Being Built

The multi-source feature enables the examples-copier to monitor and process webhooks from **multiple source repositories** in a single deployment, eliminating the need for separate copier instances.

### Current State
- ✅ Single source repository per deployment
- ✅ Hardcoded repository configuration
- ✅ One GitHub App installation per instance
- ✅ Manual deployment for each source

### Future State
- 🎯 Multiple source repositories per deployment
- 🎯 Dynamic webhook routing
- 🎯 Multiple GitHub App installations
- 🎯 Centralized configuration management
- 🎯 Per-source metrics and monitoring

## Key Benefits

1. **Simplified Operations**: One deployment handles all source repositories
2. **Cost Reduction**: Shared infrastructure reduces hosting costs
3. **Easier Maintenance**: Single codebase and configuration to manage
4. **Better Observability**: Unified metrics and audit logging
5. **Scalability**: Easy to add new source repositories

## Documentation Deliverables

### 1. Implementation Plan
**File**: `docs/MULTI-SOURCE-IMPLEMENTATION-PLAN.md`

Comprehensive plan covering:
- Current architecture analysis
- Proposed architecture design
- Detailed implementation tasks (8 phases)
- Risk assessment and mitigation
- Success criteria
- Timeline (4 weeks)

**Key Sections**:
- Configuration schema updates
- Webhook routing logic
- GitHub App installation support
- Metrics and audit logging
- Testing strategy
- Deployment phases

### 2. Technical Specification
**File**: `docs/MULTI-SOURCE-TECHNICAL-SPEC.md`

Detailed technical specifications including:
- Data models and schemas
- Component interfaces
- API specifications
- Error handling
- Performance considerations
- Security requirements

**Key Components**:
- `WebhookRouter`: Routes webhooks to correct source config
- `InstallationManager`: Manages multiple GitHub App installations
- `ConfigLoader`: Enhanced to support multi-source configs
- `MetricsCollector`: Tracks per-source metrics

### 3. Migration Guide
**File**: `docs/MULTI-SOURCE-MIGRATION-GUIDE.md`

Step-by-step guide for migrating from single to multi-source:
- Backward compatibility assurance
- Manual and automated conversion options
- Consolidation of multiple deployments
- Testing and validation procedures
- Rollback plan
- FAQ section

**Migration Steps**:
1. Assess current setup
2. Backup configuration
3. Convert format (manual or automated)
4. Consolidate deployments
5. Update environment variables
6. Validate configuration
7. Deploy to staging
8. Test thoroughly
9. Production deployment
10. Decommission old deployments

### 4. Quick Reference Guide
**File**: `docs/MULTI-SOURCE-QUICK-REFERENCE.md`

Quick reference for daily operations:
- Configuration format examples
- Common tasks and patterns
- Validation commands
- Monitoring and troubleshooting
- Best practices
- Quick command reference

### 5. Example Configurations
**File**: `configs/copier-config.multi-source.example.yaml`

Complete example showing:
- Multiple source repositories
- Different organizations (mongodb, 10gen)
- Various pattern types (prefix, glob, regex)
- Multiple targets per source
- Cross-organization copying
- Global defaults

## Architecture Overview

### High-Level Flow

```
Multiple Source Repos → Webhooks → Router → Config Loader → Pattern Matcher → Target Repos
                                      ↓
                              Installation Manager
                                      ↓
                              Metrics & Audit Logging
```

### Key Components

1. **Webhook Router** (New)
   - Routes incoming webhooks to correct source configuration
   - Validates source repository against configured sources
   - Returns 204 for unknown sources

2. **Config Loader** (Enhanced)
   - Supports both legacy and multi-source formats
   - Auto-detects configuration format
   - Validates multi-source configurations
   - Converts legacy to multi-source format

3. **Installation Manager** (New)
   - Manages multiple GitHub App installations
   - Caches installation tokens
   - Handles token refresh automatically
   - Switches between installations per source

4. **Metrics Collector** (Enhanced)
   - Tracks metrics per source repository
   - Provides global and per-source statistics
   - Monitors webhook processing times
   - Tracks success/failure rates

5. **Audit Logger** (Enhanced)
   - Logs events with source repository context
   - Enables per-source audit queries
   - Tracks cross-organization operations

## Configuration Schema

### Multi-Source Format

```yaml
sources:
  - repo: "mongodb/docs-code-examples"
    branch: "main"
    installation_id: "12345678"  # Optional
    copy_rules:
      - name: "go-examples"
        source_pattern:
          type: "prefix"
          pattern: "examples/go/"
        targets:
          - repo: "mongodb/docs"
            branch: "main"
            path_transform: "code/go/${path}"
            commit_strategy:
              type: "pull_request"
              pr_title: "Update Go examples"
              auto_merge: false

  - repo: "mongodb/atlas-examples"
    branch: "main"
    installation_id: "87654321"
    copy_rules:
      # ... additional rules

defaults:
  commit_strategy:
    type: "pull_request"
    auto_merge: false
  deprecation_check:
    enabled: true
```

### Backward Compatibility

The system automatically detects and supports the legacy single-source format:

```yaml
# Legacy format - still works!
source_repo: "mongodb/docs-code-examples"
source_branch: "main"
copy_rules:
  - name: "example"
    # ... rules
```

## Implementation Phases

### Phase 1: Core Infrastructure (Week 1)
- Update configuration schema
- Implement config loading for multiple sources
- Add validation for multi-source configs
- Ensure backward compatibility

### Phase 2: Webhook Routing (Week 2)
- Implement webhook routing logic
- Add GitHub installation switching
- Update authentication handling
- Test with multiple source repos

### Phase 3: Observability (Week 3)
- Update metrics collection
- Enhance audit logging
- Add per-source monitoring
- Update health endpoints

### Phase 4: Documentation & Testing (Week 4)
- Write comprehensive documentation ✅ (Complete)
- Create migration guide ✅ (Complete)
- Add unit and integration tests
- Perform end-to-end testing

## Key Features

### 1. Automatic Source Detection
The webhook router automatically identifies the source repository from incoming webhooks and routes to the appropriate configuration.

### 2. Installation Management
Seamlessly switches between GitHub App installations for different organizations, with automatic token caching and refresh.

### 3. Per-Source Metrics
Track webhooks, files, and operations separately for each source repository:

```json
{
  "by_source": {
    "mongodb/docs-code-examples": {
      "webhooks": {"received": 100, "processed": 98},
      "files": {"matched": 200, "uploaded": 195}
    },
    "mongodb/atlas-examples": {
      "webhooks": {"received": 50, "processed": 47},
      "files": {"matched": 120, "uploaded": 115}
    }
  }
}
```

### 4. Flexible Configuration
Support for:
- Centralized configuration (all sources in one file)
- Distributed configuration (config per source repo)
- Global defaults with per-source overrides
- Cross-organization copying

### 5. Enhanced Monitoring
- Health endpoint shows status per source
- Metrics endpoint provides per-source breakdown
- Audit logs include source repository context
- Slack notifications with source information

## Testing Strategy

### Unit Tests
- Configuration loading and validation
- Webhook routing logic
- Installation token management
- Metrics collection per source

### Integration Tests
- Multi-source webhook processing
- Installation switching
- Config format conversion
- Error handling scenarios

### End-to-End Tests
- Complete workflow with 3+ sources
- Cross-organization copying
- Failure recovery
- Performance under load

## Deployment Strategy

### Rollout Approach
1. Deploy with backward compatibility enabled
2. Test in staging with multi-source config
3. Gradual production rollout (canary deployment)
4. Monitor metrics and logs closely
5. Full production deployment
6. Decommission old single-source deployments

### Monitoring During Rollout
- Track webhook success rates per source
- Monitor GitHub API rate limits
- Watch for authentication errors
- Verify file copying success rates
- Check audit logs for anomalies

## Success Criteria

- ✅ Support 3+ source repositories in single deployment
- ✅ 100% backward compatibility with existing configs
- ✅ No performance degradation for single-source use cases
- ✅ Clear documentation and migration path
- ✅ Comprehensive test coverage (target: >80%)
- ✅ Successful production deployment

## Risk Mitigation

### Risk 1: Breaking Changes
**Mitigation**: Full backward compatibility with automatic format detection

### Risk 2: GitHub Rate Limits
**Mitigation**: Per-source rate limiting and monitoring

### Risk 3: Configuration Complexity
**Mitigation**: Clear examples, templates, and validation tools

### Risk 4: Installation Token Management
**Mitigation**: Robust caching and refresh logic with error handling

## Next Steps

### For Implementation Team
1. Review all documentation
2. Set up development environment
3. Begin Phase 1 implementation
4. Create feature branch
5. Implement core infrastructure
6. Write unit tests
7. Submit PR for review

### For Stakeholders
1. Review implementation plan
2. Approve timeline and resources
3. Identify test repositories
4. Plan staging environment
5. Schedule deployment windows

### For Operations Team
1. Review deployment strategy
2. Set up monitoring alerts
3. Prepare rollback procedures
4. Plan capacity for multi-source load

## Resources

### Documentation
- [Implementation Plan](MULTI-SOURCE-IMPLEMENTATION-PLAN.md) - Detailed implementation guide
- [Technical Spec](MULTI-SOURCE-TECHNICAL-SPEC.md) - Technical specifications
- [Migration Guide](MULTI-SOURCE-MIGRATION-GUIDE.md) - Migration instructions
- [Quick Reference](MULTI-SOURCE-QUICK-REFERENCE.md) - Daily operations guide

### Configuration Examples
- [Multi-Source Example](../configs/copier-config.multi-source.example.yaml) - Complete example config

### Diagrams
- Architecture diagram (Mermaid)
- Sequence diagram (Mermaid)
- Component interaction diagram

## Questions & Answers

### Q: When should we migrate?
**A**: Migrate when you need to monitor multiple source repositories or want to consolidate deployments. No rush - legacy format is fully supported.

### Q: What's the effort estimate?
**A**: 4 weeks for full implementation, testing, and deployment. Documentation is complete.

### Q: Will this affect existing deployments?
**A**: No. Existing single-source deployments continue to work without changes.

### Q: Can we test without affecting production?
**A**: Yes. Use dry-run mode and staging environment for thorough testing.

### Q: What if we need to rollback?
**A**: Simple rollback to previous version. Legacy format is always supported.

## Conclusion

The multi-source repository support is a significant enhancement that will:
- Simplify operations and reduce costs
- Improve scalability and flexibility
- Enhance monitoring and observability
- Maintain full backward compatibility

All documentation is complete and ready for implementation. The plan provides a clear path forward with minimal risk and maximum benefit.

---

**Status**: Documentation Complete ✅  
**Next Phase**: Implementation (Phase 1)  
**Timeline**: 4 weeks  
**Risk Level**: Low (backward compatible)

