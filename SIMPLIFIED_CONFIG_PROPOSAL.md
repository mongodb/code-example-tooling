# Simplified Configuration Proposal: Copybara-Inspired Workflows

## The Problem

Your current config has **12 rules with 300+ lines** for just 3 target repos. As you add more apps:
- 4 apps = 16 rules, 400+ lines
- 5 apps = 20 rules, 500+ lines
- 10 apps = 40 rules, 1000+ lines

**This doesn't scale.**

---

## What Copybara Does Better

Copybara uses **workflows** instead of individual rules. Each workflow defines:
1. **One source** (origin)
2. **One destination**
3. **Transformations** (what to copy and how to transform it)

### Copybara Example

```python
core.workflow(
    name = "mflix-java",
    origin = git.github_origin(
        url = "https://github.com/mongodb/docs-sample-apps.git",
        ref = "main",
    ),
    destination = git.destination(
        url = "https://github.com/mongodb/sample-app-java-mflix.git",
    ),
    destination_files = glob(["**"], exclude = ["README.md"]),
    authoring = authoring.pass_thru("Bot <bot@mongodb.com>"),
    transformations = [
        core.move("mflix/client", "client"),
        core.move("mflix/server/java-spring", "server"),
        core.replace(
            before = "mflix/README-JAVA-SPRING.md",
            after = "README.md",
        ),
    ],
)
```

**Key insight:** One workflow = one source-to-destination mapping with multiple transformations.

---

## Proposed Simplified Config

### New Format: Workflows

```yaml
source_repo: "mongodb/docs-sample-apps"
source_branch: "main"

# Global defaults (optional)
defaults:
  commit_strategy:
    type: "pull_request"
    auto_merge: false
  deprecation_check:
    enabled: true

# Workflows: one per target repo
workflows:
  # Java MFlix
  - name: "mflix-java"
    destination:
      repo: "mongodb/sample-app-java-mflix"
      branch: "main"
    
    # Simple transformations - just move directories
    transformations:
      - move:
          from: "mflix/client"
          to: "client"
      
      - move:
          from: "mflix/server/java-spring"
          to: "server"
      
      - copy:
          from: "mflix/README-JAVA-SPRING.md"
          to: "README.md"
      
      - copy:
          from: "mflix/.gitignore-java"
          to: ".gitignore"
    
    # Optional: exclude patterns
    exclude:
      - "**/.env"
      - "**/node_modules/**"

  # Node.js MFlix
  - name: "mflix-nodejs"
    destination:
      repo: "mongodb/sample-app-nodejs-mflix"
      branch: "main"
    
    transformations:
      - move:
          from: "mflix/client"
          to: "client"
      
      - move:
          from: "mflix/server/js-express"
          to: "server"
      
      - copy:
          from: "mflix/README-JAVASCRIPT-EXPRESS.md"
          to: "README.md"
      
      - copy:
          from: "mflix/.gitignore-js"
          to: ".gitignore"

  # Python MFlix
  - name: "mflix-python"
    destination:
      repo: "mongodb/sample-app-python-mflix"
      branch: "main"
    
    transformations:
      - move:
          from: "mflix/client"
          to: "client"
      
      - move:
          from: "mflix/server/python-fastapi"
          to: "server"
      
      - copy:
          from: "mflix/README-PYTHON-FASTAPI.md"
          to: "README.md"
      
      - copy:
          from: "mflix/.gitignore-python"
          to: ".gitignore"
```

**Result:** 90 lines instead of 300+ lines. **70% reduction!**

---

## Comparison: Current vs Proposed

### Current Config (12 rules, 300+ lines)

```yaml
copy_rules:
  # Rule 1: Client to Java
  - name: "mflix-client-to-java"
    source_pattern:
      type: "prefix"
      pattern: "mflix/client/"
      exclude_patterns:
        - "\\.gitignore$"
        - "README.md$"
        - "\\.env$"
    targets:
      - repo: "mongodb/sample-app-java-mflix"
        branch: "main"
        path_transform: "client/${relative_path}"
        commit_strategy:
          type: "pull_request"
          pr_title: "Update MFlix client from docs-sample-apps"
          pr_body: |
            Automated update...
          auto_merge: false
        deprecation_check:
          enabled: true

  # Rule 2: Java server
  - name: "java-server"
    source_pattern:
      type: "regex"
      pattern: "^mflix/server/java-spring/(?P<file>.+)$"
      exclude_patterns:
        - "\\.gitignore$"
        - "README.md$"
        - "\\.env$"
    targets:
      - repo: "mongodb/sample-app-java-mflix"
        branch: "main"
        path_transform: "server/${file}"
        commit_strategy:
          type: "pull_request"
          pr_title: "Update MFlix Java server from docs-sample-apps"
          pr_body: |
            Automated update...
          auto_merge: false
        deprecation_check:
          enabled: true

  # ... 10 more rules ...
```

**Problems:**
- ❌ Repetitive (same target repo repeated 4 times)
- ❌ Verbose (commit_strategy repeated 12 times)
- ❌ Complex (need to understand prefix/regex/glob patterns)
- ❌ Error-prone (easy to forget exclude patterns)
- ❌ Hard to visualize (what files go where?)

### Proposed Config (3 workflows, 90 lines)

```yaml
workflows:
  - name: "mflix-java"
    destination:
      repo: "mongodb/sample-app-java-mflix"
    transformations:
      - move: { from: "mflix/client", to: "client" }
      - move: { from: "mflix/server/java-spring", to: "server" }
      - copy: { from: "mflix/README-JAVA-SPRING.md", to: "README.md" }
      - copy: { from: "mflix/.gitignore-java", to: ".gitignore" }
```

**Benefits:**
- ✅ Concise (one workflow per target repo)
- ✅ Clear (easy to see what goes where)
- ✅ Simple (no pattern types to learn)
- ✅ DRY (defaults apply to all workflows)
- ✅ Scalable (adding apps is easy)

---

## Transformation Types

### 1. `move` - Move directory with path transformation

```yaml
- move:
    from: "mflix/server/java-spring"
    to: "server"
```

**What it does:**
- Matches all files under `mflix/server/java-spring/`
- Strips the `from` prefix
- Adds the `to` prefix
- Example: `mflix/server/java-spring/src/Main.java` → `server/src/Main.java`

### 2. `copy` - Copy single file or directory

```yaml
- copy:
    from: "mflix/README-JAVA-SPRING.md"
    to: "README.md"
```

**What it does:**
- Copies specific file(s)
- Can rename during copy
- Example: `mflix/README-JAVA-SPRING.md` → `README.md`

### 3. `glob` - Match files with glob patterns (advanced)

```yaml
- glob:
    pattern: "mflix/**/*.java"
    to: "src/${relative_path}"
```

**What it does:**
- Matches files using glob pattern
- Transforms paths using variables
- For advanced use cases

### 4. `regex` - Match files with regex (advanced)

```yaml
- regex:
    pattern: "^mflix/(?P<lang>[^/]+)/(?P<file>.+)$"
    to: "examples/${lang}/${file}"
```

**What it does:**
- Matches files using regex with named groups
- Extracts variables for path transformation
- For complex transformations

---

## How It Works Internally

### Current System (Pattern Matching)

```
For each changed file:
  For each rule:
    If file matches rule.source_pattern:
      For each target in rule.targets:
        Transform path using target.path_transform
        Queue file for upload
```

**Problem:** O(files × rules × targets) complexity

### Proposed System (Workflows)

```
For each workflow:
  For each transformation in workflow:
    For each changed file:
      If file matches transformation.from:
        Transform path using transformation.to
        Queue file for workflow.destination
```

**Benefit:** O(workflows × transformations × files) - same complexity but clearer structure

---

## Migration Path

### Phase 1: Add Workflow Support (Backward Compatible)

```go
// types/config.go
type YAMLConfig struct {
    // Legacy format (still supported)
    SourceRepo   string     `yaml:"source_repo,omitempty"`
    SourceBranch string     `yaml:"source_branch,omitempty"`
    CopyRules    []CopyRule `yaml:"copy_rules,omitempty"`
    
    // New format
    Workflows []Workflow `yaml:"workflows,omitempty"`
    Defaults  *Defaults  `yaml:"defaults,omitempty"`
}

type Workflow struct {
    Name            string          `yaml:"name"`
    Destination     Destination     `yaml:"destination"`
    Transformations []Transformation `yaml:"transformations"`
    Exclude         []string        `yaml:"exclude,omitempty"`
}

type Destination struct {
    Repo   string `yaml:"repo"`
    Branch string `yaml:"branch,omitempty"`
}

type Transformation struct {
    Move  *MoveTransform  `yaml:"move,omitempty"`
    Copy  *CopyTransform  `yaml:"copy,omitempty"`
    Glob  *GlobTransform  `yaml:"glob,omitempty"`
    Regex *RegexTransform `yaml:"regex,omitempty"`
}

type MoveTransform struct {
    From string `yaml:"from"`
    To   string `yaml:"to"`
}

type CopyTransform struct {
    From string `yaml:"from"`
    To   string `yaml:"to"`
}
```

### Phase 2: Convert Existing Config

```bash
# Tool to convert old config to new format
go run cmd/config-converter/main.go \
  --input copier-config.yaml \
  --output copier-config-v2.yaml
```

### Phase 3: Deprecate Old Format (6 months later)

Add warning when loading old format:
```
WARNING: Legacy copy_rules format is deprecated. 
Please migrate to workflows format. 
See: https://docs.example.com/migration-guide
```

---

## Real-World Example: Adding a New App

### Current Format (Need to add 4 rules)

```yaml
copy_rules:
  # Rule 13: Client to Rust
  - name: "mflix-client-to-rust"
    source_pattern:
      type: "prefix"
      pattern: "mflix/client/"
      exclude_patterns: [...]
    targets:
      - repo: "mongodb/sample-app-rust-mflix"
        branch: "main"
        path_transform: "client/${relative_path}"
        commit_strategy: {...}
        deprecation_check: {...}

  # Rule 14: Rust server
  - name: "rust-server"
    source_pattern:
      type: "regex"
      pattern: "^mflix/server/rust-actix/(?P<file>.+)$"
      exclude_patterns: [...]
    targets:
      - repo: "mongodb/sample-app-rust-mflix"
        branch: "main"
        path_transform: "server/${file}"
        commit_strategy: {...}
        deprecation_check: {...}

  # Rule 15: Rust README
  - name: "mflix-rust-readme"
    source_pattern:
      type: "glob"
      pattern: "mflix/README-RUST-ACTIX.md"
    targets:
      - repo: "mongodb/sample-app-rust-mflix"
        branch: "main"
        path_transform: "README.md"
        commit_strategy: {...}
        deprecation_check: {...}

  # Rule 16: Rust .gitignore
  - name: "mflix-rust-gitignore"
    source_pattern:
      type: "glob"
      pattern: "mflix/.gitignore-rust"
    targets:
      - repo: "mongodb/sample-app-rust-mflix"
        branch: "main"
        path_transform: ".gitignore"
        commit_strategy: {...}
        deprecation_check: {...}
```

**~80 lines of config**

### Proposed Format (Add 1 workflow)

```yaml
workflows:
  # ... existing workflows ...

  # Rust MFlix
  - name: "mflix-rust"
    destination:
      repo: "mongodb/sample-app-rust-mflix"
    transformations:
      - move: { from: "mflix/client", to: "client" }
      - move: { from: "mflix/server/rust-actix", to: "server" }
      - copy: { from: "mflix/README-RUST-ACTIX.md", to: "README.md" }
      - copy: { from: "mflix/.gitignore-rust", to: ".gitignore" }
```

**~10 lines of config** (8x reduction!)

---

## Benefits Summary

| Aspect | Current | Proposed | Improvement |
|--------|---------|----------|-------------|
| **Lines of config** | 300+ | 90 | 70% reduction |
| **Rules per app** | 4 | 1 | 75% reduction |
| **Repetition** | High | Low | DRY principles |
| **Readability** | Complex | Simple | Easy to understand |
| **Maintainability** | Hard | Easy | Clear structure |
| **Scalability** | Poor | Good | Linear growth |
| **Error-prone** | Yes | No | Fewer mistakes |

---

## Implementation Effort

### Estimated Time: 1-2 weeks

**Week 1: Core Implementation**
- Day 1-2: Add workflow types to `types/config.go`
- Day 3-4: Implement workflow processor in `services/workflow_processor.go`
- Day 5: Add backward compatibility layer

**Week 2: Testing & Migration**
- Day 1-2: Write comprehensive tests
- Day 3: Create config converter tool
- Day 4: Convert production config
- Day 5: Deploy and monitor

---

## Recommendation

**✅ Implement the simplified workflow format.**

**Why:**
1. **Immediate benefit:** 70% reduction in config size
2. **Future-proof:** Scales linearly as you add apps
3. **Easier to maintain:** Clear, simple structure
4. **Less error-prone:** Fewer places to make mistakes
5. **Better DX:** Easier for team to understand and modify

**This is a better investment than:**
- Switching to Copybara (5-6 weeks, lose features)
- Continuing with current format (technical debt grows)
- Building from scratch (reinvent the wheel)

---

## Next Steps

1. **Review this proposal** - Does the workflow format meet your needs?
2. **Prototype implementation** - I can build this in 1-2 weeks
3. **Convert your config** - Migrate from 12 rules to 3 workflows
4. **Deploy and test** - Verify it works with real PRs
5. **Document** - Update guides with new format

**Want me to start implementing this?** 🚀

