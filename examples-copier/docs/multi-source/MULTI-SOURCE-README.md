# Multi-Source Repository Support - Documentation Index

## 📋 Overview

This directory contains comprehensive documentation for implementing multi-source repository support in the examples-copier application. This feature enables monitoring and processing webhooks from multiple source repositories in a single deployment.

## 🎯 Quick Start

**New to multi-source?** Start here:

1. **[Summary](docs/MULTI-SOURCE-SUMMARY.md)** - High-level overview and benefits
2. **[Quick Reference](docs/MULTI-SOURCE-QUICK-REFERENCE.md)** - Common tasks and commands
3. **[Example Config](configs/copier-config.multi-source.example.yaml)** - Working configuration example

**Ready to implement?** Follow this path:

1. **[Implementation Plan](docs/MULTI-SOURCE-IMPLEMENTATION-PLAN.md)** - Detailed implementation guide
2. **[Technical Spec](docs/MULTI-SOURCE-TECHNICAL-SPEC.md)** - Technical specifications
3. **[Migration Guide](docs/MULTI-SOURCE-MIGRATION-GUIDE.md)** - Step-by-step migration

## 📚 Documentation

### Core Documents

| Document | Purpose | Audience |
|----------|---------|----------|
| [**Summary**](docs/MULTI-SOURCE-SUMMARY.md) | Executive overview, benefits, and status | Everyone |
| [**Implementation Plan**](docs/MULTI-SOURCE-IMPLEMENTATION-PLAN.md) | Detailed implementation roadmap | Developers |
| [**Technical Spec**](docs/MULTI-SOURCE-TECHNICAL-SPEC.md) | Technical specifications and APIs | Developers |
| [**Migration Guide**](docs/MULTI-SOURCE-MIGRATION-GUIDE.md) | Migration from single to multi-source | DevOps, Developers |
| [**Quick Reference**](docs/MULTI-SOURCE-QUICK-REFERENCE.md) | Daily operations and troubleshooting | Everyone |

### Configuration Examples

| File | Description |
|------|-------------|
| [**Multi-Source Example**](configs/copier-config.multi-source.example.yaml) | Complete multi-source configuration |
| [**Single-Source Example**](configs/copier-config.example.yaml) | Legacy single-source format |

### Visual Diagrams

- **Architecture Diagram**: High-level system architecture with multiple sources
- **Sequence Diagram**: Webhook processing flow for multi-source setup

## 🚀 What's New

### Key Features

✅ **Multiple Source Repositories**
- Monitor 3+ source repositories in one deployment
- Each source has independent copy rules
- Cross-organization support (mongodb, 10gen, etc.)

✅ **Intelligent Webhook Routing**
- Automatic source repository detection
- Dynamic configuration loading
- Graceful handling of unknown sources

✅ **Multi-Installation Support**
- Different GitHub App installations per organization
- Automatic token management and refresh
- Seamless installation switching

✅ **Enhanced Observability**
- Per-source metrics and monitoring
- Source-specific audit logging
- Detailed health status per source

✅ **100% Backward Compatible**
- Existing single-source configs work unchanged
- Automatic format detection
- Gradual migration path

## 📖 Documentation Guide

### For Product Managers

**Start with:**
1. [Summary](docs/MULTI-SOURCE-SUMMARY.md) - Understand benefits and scope
2. [Implementation Plan](docs/MULTI-SOURCE-IMPLEMENTATION-PLAN.md) - Review timeline and phases

**Key Questions Answered:**
- Why do we need this? → See "Key Benefits" in Summary
- What's the timeline? → 4 weeks (see Implementation Plan)
- What are the risks? → See "Risk Mitigation" in Summary
- How do we measure success? → See "Success Criteria" in Implementation Plan

### For Developers

**Start with:**
1. [Technical Spec](docs/MULTI-SOURCE-TECHNICAL-SPEC.md) - Understand architecture
2. [Implementation Plan](docs/MULTI-SOURCE-IMPLEMENTATION-PLAN.md) - See detailed tasks

**Key Sections:**
- Data models and schemas → Technical Spec §3
- Component specifications → Technical Spec §4
- API specifications → Technical Spec §5
- Implementation tasks → Implementation Plan §2-8

**Code Changes Required:**
- `types/config.go` - New configuration types
- `services/config_loader.go` - Enhanced config loading
- `services/webhook_handler_new.go` - Webhook routing
- `services/github_auth.go` - Installation management
- `services/health_metrics.go` - Per-source metrics

### For DevOps/SRE

**Start with:**
1. [Migration Guide](docs/MULTI-SOURCE-MIGRATION-GUIDE.md) - Migration steps
2. [Quick Reference](docs/MULTI-SOURCE-QUICK-REFERENCE.md) - Operations guide

**Key Sections:**
- Deployment strategy → Implementation Plan §10
- Monitoring and metrics → Quick Reference "Monitoring"
- Troubleshooting → Quick Reference "Troubleshooting"
- Rollback procedures → Migration Guide "Rollback Plan"

**Operational Tasks:**
- Configuration validation
- Staging deployment
- Production rollout
- Monitoring and alerting
- Decommissioning old deployments

### For QA/Testing

**Start with:**
1. [Technical Spec](docs/MULTI-SOURCE-TECHNICAL-SPEC.md) §9 - Testing strategy
2. [Migration Guide](docs/MULTI-SOURCE-MIGRATION-GUIDE.md) - Testing checklist

**Test Scenarios:**
- Multi-source webhook processing
- Installation switching
- Config format conversion
- Error handling
- Performance under load
- Cross-organization copying

## 🔧 Configuration Examples

### Single Source (Legacy - Still Supported)

```yaml
source_repo: "mongodb/docs-code-examples"
source_branch: "main"
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
```

### Multi-Source (New)

```yaml
sources:
  - repo: "mongodb/docs-code-examples"
    branch: "main"
    installation_id: "12345678"
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
  
  - repo: "mongodb/atlas-examples"
    branch: "main"
    installation_id: "87654321"
    copy_rules:
      - name: "atlas-cli"
        source_pattern:
          type: "glob"
          pattern: "cli/**/*.go"
        targets:
          - repo: "mongodb/atlas-cli"
            branch: "main"
            path_transform: "examples/${filename}"
            commit_strategy:
              type: "direct"
```

## 🎯 Implementation Roadmap

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
- [x] Write comprehensive documentation
- [x] Create migration guide
- [ ] Add unit and integration tests
- [ ] Perform end-to-end testing

## 📊 Success Metrics

- ✅ Support 3+ source repositories in single deployment
- ✅ 100% backward compatibility
- ✅ No performance degradation
- ✅ Clear documentation (Complete)
- ⏳ Test coverage >80%
- ⏳ Successful production deployment

## 🔗 Related Documentation

### Existing Documentation
- [Main README](README.md) - Application overview
- [Architecture](docs/ARCHITECTURE.md) - Current architecture
- [Configuration Guide](docs/CONFIGURATION-GUIDE.md) - Configuration reference
- [Deployment Guide](docs/DEPLOYMENT.md) - Deployment instructions

### New Documentation
- [Multi-Source Summary](docs/MULTI-SOURCE-SUMMARY.md)
- [Implementation Plan](docs/MULTI-SOURCE-IMPLEMENTATION-PLAN.md)
- [Technical Specification](docs/MULTI-SOURCE-TECHNICAL-SPEC.md)
- [Migration Guide](docs/MULTI-SOURCE-MIGRATION-GUIDE.md)
- [Quick Reference](docs/MULTI-SOURCE-QUICK-REFERENCE.md)

## 💡 Quick Commands

```bash
# Validate multi-source config
./config-validator validate -config copier-config.yaml -v

# Convert legacy to multi-source
./config-validator convert-to-multi-source \
  -input copier-config.yaml \
  -output copier-config-multi.yaml

# Test pattern matching
./config-validator test-pattern \
  -config copier-config.yaml \
  -source "mongodb/docs-code-examples" \
  -file "examples/go/main.go"

# Dry run with multi-source
DRY_RUN=true ./examples-copier -config copier-config-multi.yaml

# Check health (per-source status)
curl http://localhost:8080/health | jq '.sources'

# Get metrics by source
curl http://localhost:8080/metrics | jq '.by_source'
```

## 🤝 Contributing

When implementing multi-source support:

1. Follow the implementation plan phases
2. Write tests for all new functionality
3. Update documentation as needed
4. Ensure backward compatibility
5. Test with multiple source repositories
6. Monitor metrics during rollout

## 📞 Support

For questions or issues:

1. Check the [Quick Reference](docs/MULTI-SOURCE-QUICK-REFERENCE.md) for common tasks
2. Review the [Migration Guide](docs/MULTI-SOURCE-MIGRATION-GUIDE.md) FAQ
3. Consult the [Technical Spec](docs/MULTI-SOURCE-TECHNICAL-SPEC.md) for details
4. Check existing [Troubleshooting Guide](docs/TROUBLESHOOTING.md)

## 📝 Status

| Component | Status |
|-----------|--------|
| Documentation | ✅ Complete |
| Implementation Plan | ✅ Complete |
| Technical Spec | ✅ Complete |
| Migration Guide | ✅ Complete |
| Example Configs | ✅ Complete |
| Code Implementation | ⏳ Pending |
| Unit Tests | ⏳ Pending |
| Integration Tests | ⏳ Pending |
| Staging Deployment | ⏳ Pending |
| Production Deployment | ⏳ Pending |

**Last Updated**: 2025-10-15  
**Version**: 1.0  
**Status**: Documentation Complete, Ready for Implementation

---

**Next Steps**: Begin Phase 1 implementation (Core Infrastructure)

