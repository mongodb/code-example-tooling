# Pattern Matching Guide

Complete guide to pattern matching and path transformation in the examples-copier.

## Table of Contents

- [Overview](#overview)
- [Pattern Types](#pattern-types)
  - [Prefix Patterns](#prefix-patterns)
  - [Glob Patterns](#glob-patterns)
  - [Regex Patterns](#regex-patterns)
- [Path Transformation](#path-transformation)
- [Built-in Variables](#built-in-variables)
- [Common Patterns](#common-patterns)
- [Ignore Patterns](#ignore-patterns)
- [Testing Patterns](#testing-patterns)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)

## Overview

The examples-copier uses a powerful pattern matching system to:
1. **Match files** from merged PRs based on their paths
2. **Extract variables** from file paths (e.g., language, category)
3. **Transform paths** to determine where files should be copied

### How It Works

```
Source File Path → Pattern Matching → Variable Extraction → Path Transformation → Target File Path
```

**Example:**
```
examples/go/database/connect.go
    ↓ (regex pattern)
lang=go, category=database, file=connect.go
    ↓ (path transform)
docs/code-examples/go/database/connect.go
```

## Pattern Types

The copier supports three pattern types: **prefix**, **glob**, and **regex**.

### Prefix Patterns

Simple string prefix matching. Fast and straightforward.

#### Syntax

```yaml
source_pattern:
  type: "prefix"
  pattern: "examples/"
```

#### How It Works

Matches any file path that **starts with** the specified prefix.

#### Examples

**Pattern:** `examples/`

| File Path                 | Matches? |
|---------------------------|----------|
| `examples/go/main.go`     | ✅ Yes    |
| `examples/python/auth.py` | ✅ Yes    |
| `src/examples/test.js`    | ❌ No     |
| `docs/readme.md`          | ❌ No     |

**Pattern:** `source/examples/generated/`

| File Path                               | Matches? |
|-----------------------------------------|----------|
| `source/examples/generated/node/app.js` | ✅ Yes    |
| `source/examples/manual/test.py`        | ❌ No     |
| `examples/generated/go/main.go`         | ❌ No     |

#### Variables Extracted

Prefix patterns automatically extract:
- `matched_prefix` - The prefix that was matched
- `relative_path` - The path after the prefix

**Example:**
```yaml
source_pattern:
  type: "prefix"
  pattern: "examples/"
```

For file `examples/go/database/connect.go`:
- `matched_prefix` = `"examples"`
- `relative_path` = `"go/database/connect.go"`

#### When to Use

- ✅ Simple directory-based matching
- ✅ Copy entire directory trees
- ✅ When you don't need to extract specific variables
- ✅ Maximum performance

### Glob Patterns

Pattern matching with wildcards. More flexible than prefix, simpler than regex.

#### Syntax

```yaml
source_pattern:
  type: "glob"
  pattern: "examples/**/*.go"
```

#### Wildcards

| Wildcard | Matches                     | Example                           |
|----------|-----------------------------|-----------------------------------|
| `*`      | Any characters except `/`   | `*.go` matches `main.go`          |
| `**`     | Any number of directories   | `**/*.go` matches `a/b/c/main.go` |
| `?`      | Single character except `/` | `test?.go` matches `test1.go`     |

#### Examples

**Pattern:** `examples/**/*.go`

| File Path                   | Matches?                            |
|-----------------------------|-------------------------------------|
| `examples/go/main.go`       | ✅ Yes                               |
| `examples/go/auth/login.go` | ✅ Yes                               |
| `examples/python/main.py`   | ❌ No (not .go)                      |
| `src/examples/test.go`      | ❌ No (doesn't start with examples/) |

**Pattern:** `source/*/generated/*.js`

| File Path                              | Matches?               |
|----------------------------------------|------------------------|
| `source/examples/generated/app.js`     | ✅ Yes                  |
| `source/tests/generated/test.js`       | ✅ Yes                  |
| `source/examples/generated/sub/app.js` | ❌ No (extra directory) |
| `source/examples/manual/app.js`        | ❌ No (not generated/)  |

**Pattern:** `examples/go/test?.go`

| File Path               | Matches?              |
|-------------------------|-----------------------|
| `examples/go/test1.go`  | ✅ Yes                 |
| `examples/go/testA.go`  | ✅ Yes                 |
| `examples/go/test12.go` | ❌ No (two characters) |
| `examples/go/test.go`   | ❌ No (no character)   |

#### Variables Extracted

Glob patterns extract:
- `matched_pattern` - The pattern that was matched

**Note:** Glob patterns don't extract specific variables like language or category. Use regex for that.

#### When to Use

- ✅ Match files by extension (e.g., `*.go`, `*.py`)
- ✅ Match files in nested directories (`**/*.js`)
- ✅ Simple wildcard matching
- ❌ Don't use when you need to extract variables

### Regex Patterns

Full regular expression support with named capture groups for variable extraction.

#### Syntax

```yaml
source_pattern:
  type: "regex"
  pattern: "^examples/(?P<lang>[^/]+)/(?P<category>[^/]+)/(?P<file>.+)$"
```

#### Named Capture Groups

Use `(?P<name>...)` to extract variables:

```regex
^examples/(?P<lang>[^/]+)/(?P<category>[^/]+)/(?P<file>.+)$
```

This extracts:
- `lang` - The language (e.g., `go`, `python`)
- `category` - The category (e.g., `database`, `auth`)
- `file` - The filename (e.g., `connect.go`)

#### Common Regex Patterns

**Match any character except `/`:**
```regex
[^/]+
```

**Match everything to end of string:**
```regex
.+
```

**Match optional group:**
```regex
(?P<optional>[^/]+)?
```

**Match specific values:**
```regex
(?P<lang>go|python|node)
```

**Match with anchors:**
```regex
^examples/  # Start of string
\.go$       # End of string
```

#### Examples

**Pattern:** `^examples/(?P<lang>[^/]+)/(?P<file>.+)$`

| File Path                       | Matches? | Variables                         |
|---------------------------------|----------|-----------------------------------|
| `examples/go/main.go`           | ✅ Yes    | `lang=go, file=main.go`           |
| `examples/python/auth/login.py` | ✅ Yes    | `lang=python, file=auth/login.py` |
| `src/examples/go/main.go`       | ❌ No     | -                                 |

**Pattern:** `^generated-examples/(?P<project>[^/]+)/(?P<rest>.+)$`

| File Path                                     | Matches? | Variables                                |
|-----------------------------------------------|----------|------------------------------------------|
| `generated-examples/test-project/cmd/main.go` | ✅ Yes    | `project=test-project, rest=cmd/main.go` |
| `generated-examples/app/internal/auth.go`     | ✅ Yes    | `project=app, rest=internal/auth.go`     |
| `examples/test-project/main.go`               | ❌ No     | -                                        |

**Pattern:** `^source/examples/(?P<type>generated|manual)/(?P<lang>[^/]+)/(?P<file>.+)$`

| File Path                               | Matches? | Variables                                |
|-----------------------------------------|----------|------------------------------------------|
| `source/examples/generated/node/app.js` | ✅ Yes    | `type=generated, lang=node, file=app.js` |
| `source/examples/manual/python/test.py` | ✅ Yes    | `type=manual, lang=python, file=test.py` |
| `source/examples/other/go/main.go`      | ❌ No     | -                                        |

#### When to Use

- ✅ Extract variables from file paths
- ✅ Complex matching logic
- ✅ Match specific patterns
- ✅ Maximum flexibility

## Path Transformation

After matching a file, the copier transforms the source path to determine the target path.

### Syntax

```yaml
path_transform: "docs/code-examples/${lang}/${category}/${file}"
```

### Variable Substitution

Use `${variable}` to insert extracted variables:

```yaml
# Pattern extracts: lang=go, category=database, file=connect.go
path_transform: "docs/${lang}/${category}/${file}"
# Result: docs/go/database/connect.go
```

### Examples

**Keep same structure:**
```yaml
path_transform: "${path}"
# examples/go/main.go → examples/go/main.go
```

**Add prefix:**
```yaml
path_transform: "docs/${path}"
# examples/go/main.go → docs/examples/go/main.go
```

**Reorganize with variables:**
```yaml
# Pattern: ^examples/(?P<lang>[^/]+)/(?P<file>.+)$
path_transform: "code-examples/${lang}/${file}"
# examples/go/database/connect.go → code-examples/go/database/connect.go
```

**Use relative path:**
```yaml
# Prefix pattern: examples/
path_transform: "docs/${relative_path}"
# examples/go/main.go → docs/go/main.go
```

**Flatten structure:**
```yaml
# Pattern: ^examples/(?P<lang>[^/]+)/(?P<category>[^/]+)/(?P<file>[^/]+)$
path_transform: "all-examples/${file}"
# examples/go/database/connect.go → all-examples/connect.go
```

## Built-in Variables

In addition to variables extracted from patterns, these built-in variables are always available:

| Variable      | Description                | Example               |
|---------------|----------------------------|-----------------------|
| `${path}`     | Full source file path      | `examples/go/main.go` |
| `${filename}` | Just the filename          | `main.go`             |
| `${dir}`      | Directory path             | `examples/go`         |
| `${ext}`      | File extension (with dot)  | `.go`                 |
| `${name}`     | Filename without extension | `main`                |

### Example

For file `examples/go/database/connect.go`:

```yaml
path_transform: "${dir}/${name}_copy${ext}"
# Result: examples/go/database/connect_copy.go
```

## Common Patterns

### Pattern 1: Language-Based Examples

**Use Case:** Copy examples organized by programming language

**Structure:**
```
examples/
  go/
    main.go
    auth.go
  python/
    main.py
    auth.py
```

**Configuration:**
```yaml
source_pattern:
  type: "regex"
  pattern: "^examples/(?P<lang>[^/]+)/(?P<file>.+)$"
targets:
  - path_transform: "docs/code-examples/${lang}/${file}"
```

**Result:**
```
examples/go/main.go → docs/code-examples/go/main.go
examples/python/main.py → docs/code-examples/python/main.py
```

### Pattern 2: Categorized Examples

**Use Case:** Examples organized by language and category

**Structure:**
```
examples/
  go/
    database/
      connect.go
    auth/
      login.go
```

**Configuration:**
```yaml
source_pattern:
  type: "regex"
  pattern: "^examples/(?P<lang>[^/]+)/(?P<category>[^/]+)/(?P<file>.+)$"
targets:
  - path_transform: "docs/${lang}/${category}/${file}"
```

**Result:**
```
examples/go/database/connect.go → docs/go/database/connect.go
examples/go/auth/login.go → docs/go/auth/login.go
```

### Pattern 3: Generated vs Manual Examples

**Use Case:** Separate generated and manual examples

**Structure:**
```
source/examples/
  generated/
    node/app.js
  manual/
    python/test.py
```

**Configuration:**
```yaml
source_pattern:
  type: "regex"
  pattern: "^source/examples/(?P<type>generated|manual)/(?P<lang>[^/]+)/(?P<file>.+)$"
targets:
  - path_transform: "docs/${type}-examples/${lang}/${file}"
```

**Result:**
```
source/examples/generated/node/app.js → docs/generated-examples/node/app.js
source/examples/manual/python/test.py → docs/manual-examples/python/test.py
```

### Pattern 4: Project-Based Examples

**Use Case:** Multiple projects with examples

**Structure:**
```
generated-examples/
  project-a/
    cmd/main.go
  project-b/
    internal/auth.go
```

**Configuration:**
```yaml
source_pattern:
  type: "regex"
  pattern: "^generated-examples/(?P<project>[^/]+)/(?P<rest>.+)$"
targets:
  - path_transform: "examples/${project}/${rest}"
```

**Result:**
```
generated-examples/project-a/cmd/main.go → examples/project-a/cmd/main.go
generated-examples/project-b/internal/auth.go → examples/project-b/internal/auth.go
```

### Pattern 5: Copy All Files in Directory

**Use Case:** Copy entire directory tree without transformation

**Configuration:**
```yaml
source_pattern:
  type: "prefix"
  pattern: "examples/"
targets:
  - path_transform: "docs/${path}"
```

**Result:**
```
examples/go/main.go → docs/examples/go/main.go
examples/python/test.py → docs/examples/python/test.py
```

## Ignore Patterns

### Basic Template

```yaml
source_pattern:
  type: "regex"
  pattern: "^base/path/(?P<file>(?!IGNORE_PATTERN).+)$"
```

###  Common Ignore Patterns

#### Ignore Single Directory

```yaml
# Ignore node_modules/
pattern: "^mflix/client/(?P<file>(?!node_modules/).+)$"
```

#### Ignore Multiple Directories

```yaml
# Ignore node_modules/, .next/, dist/, build/
pattern: "^mflix/client/(?P<file>(?!(?:node_modules|.next|dist|build)/).+)$"
```

#### Ignore Files by Extension

```yaml
# Ignore .test.js files
pattern: "^mflix/server/(?P<file>(?!.*\\.test\\.js$).+)$"

# Ignore .pyc files
pattern: "^mflix/server/(?P<file>(?!.*\\.pyc$).+)$"

# Ignore multiple extensions
pattern: "^mflix/server/(?P<file>(?!.*\\.(?:test\\.js|spec\\.js|pyc)$).+)$"
```

#### Ignore Files by Name Pattern

```yaml
# Ignore test files (containing "test" or "Test")
pattern: "^mflix/server/(?P<file>(?!.*[Tt]est.*).+)$"

# Ignore files starting with "test_"
pattern: "^mflix/server/(?P<file>(?!test_.*).+)$"

# Ignore files ending with "_test.py"
pattern: "^mflix/server/(?P<file>(?!.*_test\\.py$).+)$"
```

#### Ignore Hidden Files

```yaml
# Ignore files starting with .
pattern: "^mflix/client/(?P<file>(?!\\.).+)$"
```

#### Ignore Specific File

```yaml
# Ignore config.json
pattern: "^mflix/client/(?P<file>(?!config\\.json$).+)$"

# Ignore .gitignore in server directory
pattern: "^mflix/server/java-spring/(?P<file>(?!\\.gitignore$).+)$"
```

### Language-Specific Patterns

#### JavaScript/Node.js

```yaml
# Ignore common Node.js artifacts
pattern: "^path/(?P<file>(?!(?:node_modules|dist|build|coverage|\\.next)/).+)$"

# Ignore test files
pattern: "^path/(?P<file>(?!.*\\.(?:test|spec)\\.(?:js|jsx|ts|tsx)$).+)$"

# Ignore config files
pattern: "^path/(?P<file>(?!.*\\.config\\.js$).+)$"
```

#### Python

```yaml
# Ignore Python artifacts
pattern: "^path/(?P<file>(?!(?:__pycache__|.*\\.pyc$|\\.pytest_cache)).+)$"

# Ignore test files
pattern: "^path/(?P<file>(?!(?:test_.*\\.py$|.*_test\\.py$)).+)$"

# Ignore virtual environments
pattern: "^path/(?P<file>(?!(?:venv|env|\\.venv)/).+)$"
```

#### Java

```yaml
# Ignore build artifacts
pattern: "^path/(?P<file>(?!(?:target|build|out)/).+)$"

# Ignore test files
pattern: "^path/(?P<file>(?!.*[Tt]est\\.java$).+)$"

# Ignore compiled classes
pattern: "^path/(?P<file>(?!.*\\.class$).+)$"
```

#### Go

```yaml
# Ignore test files
pattern: "^path/(?P<file>(?!.*_test\\.go$).+)$"

# Ignore vendor directory
pattern: "^path/(?P<file>(?!vendor/).+)$"
```

### Complex Ignore Patterns

#### Ignore Multiple Patterns (OR)

```yaml
# Ignore: test files OR config files OR markdown files
pattern: "^path/(?P<file>(?!.*(?:_test\\.go|\\.config\\.js|\\.md)$).+)$"
```

#### Ignore at Any Depth

```yaml
# Ignore node_modules at any level
pattern: "^path/(?P<file>(?!.*node_modules.*).+)$"
```

#### Ignore Multiple Directories with Different Patterns

```yaml
# Ignore: build directories AND test files
pattern: "^path/(?P<file>(?!(?:build|dist|out)/|.*_test\\.py$).+)$"
```

### Real-World Examples

#### Next.js Project

```yaml
source_pattern:
  type: "regex"
  pattern: "^mflix/client/(?P<file>(?!(?:node_modules|.next|out|build|\\.cache|coverage)/).+)$"
```

**Ignores**: node_modules/, .next/, out/, build/, .cache/, coverage/

#### Express Server

```yaml
source_pattern:
  type: "regex"
  pattern: "^mflix/server/express/(?P<file>(?!(?:node_modules|.*\\.test\\.js$|.*\\.spec\\.js$)).+)$"
```

**Ignores**: node_modules/, *.test.js, *.spec.js

#### Python Flask Server

```yaml
source_pattern:
  type: "regex"
  pattern: "^mflix/server/python/(?P<file>(?!(?:__pycache__|.*\\.pyc$|test_.*\\.py$|.*_test\\.py$|\\.pytest_cache|venv)/).+)$"
```

**Ignores**: __pycache__/, *.pyc, test_*.py, *_test.py, .pytest_cache/, venv/

#### Java Spring Boot

```yaml
source_pattern:
  type: "regex"
  pattern: "^mflix/server/java-spring/(?P<file>(?!(?:target|.*[Tt]est\\.java$|.*\\.class$)).+)$"
```

**Ignores**: target/, *Test.java, *test.java, *.class

## Testing Your Pattern

```bash
cd examples-copier

# Test if a file matches
./config-validator test-pattern \
  -type regex \
  -pattern "^mflix/client/(?P<file>(?!node_modules/).+)$" \
  -file "mflix/client/src/App.js"
# Should match: ✅

./config-validator test-pattern \
  -type regex \
  -pattern "^mflix/client/(?P<file>(?!node_modules/).+)$" \
  -file "mflix/client/node_modules/react/index.js"
# Should NOT match: ❌
```

## Common Mistakes

### ❌ Wrong: Dot not escaped

```yaml
pattern: "^path/(?P<file>(?!.test.js$).+)$"
```

### ✅ Right: Dot escaped

```yaml
pattern: "^path/(?P<file>(?!.*\\.test\\.js$).+)$"
```

---

### ❌ Wrong: Negative lookahead at end

```yaml
pattern: "^path/.+(?!node_modules/)$"
```

### ✅ Right: Negative lookahead at start

```yaml
pattern: "^path/(?!node_modules/).+$"
```

---

### ❌ Wrong: Missing capture group

```yaml
pattern: "^path/(?!node_modules/).+$"
```

### ✅ Right: With named capture group

```yaml
pattern: "^path/(?P<file>(?!node_modules/).+)$"
```

## Cheat Sheet

| Goal | Pattern |
|------|---------|
| Ignore 1 dir | `(?!dirname/)` |
| Ignore 2+ dirs | `(?!(?:dir1\|dir2\|dir3)/)` |
| Ignore extension | `(?!.*\\.ext$)` |
| Ignore 2+ extensions | `(?!.*\\.(?:ext1\|ext2)$)` |
| Ignore prefix | `(?!prefix.*)` |
| Ignore suffix | `(?!.*suffix$)` |
| Ignore hidden files | `(?!\\.)` |
| Ignore test files | `(?!.*[Tt]est.*)` |
| Ignore at any depth | `(?!.*pattern.*)` |

## See Also

- [Ignoring Files Guide](IGNORING-FILES-GUIDE.md) - Detailed guide with explanations
- [Pattern Matching Guide](PATTERN-MATCHING-GUIDE.md) - Complete pattern reference
- [Example Config](../configs/mflix-with-ignores.yaml) - Real-world examples



## Testing Patterns

### Using config-validator

Test patterns before deploying:

```bash
# Test if a pattern matches a file
./config-validator test-pattern \
  -type regex \
  -pattern "^examples/(?P<lang>[^/]+)/(?P<file>.+)$" \
  -file "examples/go/main.go"
```

**Output:**
```
✅ Pattern matched!

Extracted variables:
  lang = go
  file = main.go
```

### Test Path Transformation

```bash
./config-validator test-transform \
  -source "examples/go/main.go" \
  -template "docs/${lang}/${file}" \
  -vars "lang=go,file=main.go"
```

**Output:**
```
✅ Transform successful!
Source: examples/go/main.go
Result: docs/go/main.go
```

### Validate Full Configuration

```bash
./config-validator validate -config copier-config.yaml -v
```

## Best Practices

### 1. Start Simple, Then Refine

```yaml
# Start with prefix
source_pattern:
  type: "prefix"
  pattern: "examples/"

# Then add regex when you need variables
source_pattern:
  type: "regex"
  pattern: "^examples/(?P<lang>[^/]+)/(?P<file>.+)$"
```

### 2. Use Anchors in Regex

Always use `^` and `$` to match the entire path:

```yaml
# ✅ Good - matches exact pattern
pattern: "^examples/(?P<lang>[^/]+)/(?P<file>.+)$"

# ❌ Bad - might match partial paths
pattern: "examples/(?P<lang>[^/]+)/(?P<file>.+)"
```

### 3. Be Specific with Character Classes

```yaml
# ✅ Good - matches directory name (no slashes)
pattern: "^examples/(?P<lang>[^/]+)/(?P<file>.+)$"

# ❌ Bad - .+ matches everything including slashes
pattern: "^examples/(?P<lang>.+)/(?P<file>.+)$"
```

### 4. Test with Real File Paths

Use actual file paths from your repository:

```bash
# Get real file paths from a PR
gh pr view 42 --json files --jq '.files[].path'

# Test each one
./config-validator test-pattern -type regex \
  -pattern "^examples/(?P<lang>[^/]+)/(?P<file>.+)$" \
  -file "examples/go/database/connect.go"
```

### 5. Order Rules from Specific to General

```yaml
copy_rules:
  # More specific rule first
  - name: "Copy Go examples"
    source_pattern:
      type: "regex"
      pattern: "^examples/go/(?P<file>.+)$"
  
  # General fallback rule last
  - name: "Copy all examples"
    source_pattern:
      type: "prefix"
      pattern: "examples/"
```

### 6. Use Descriptive Variable Names

```yaml
# ✅ Good - clear what each variable represents
pattern: "^examples/(?P<language>[^/]+)/(?P<category>[^/]+)/(?P<filename>.+)$"

# ❌ Bad - unclear variable names
pattern: "^examples/(?P<a>[^/]+)/(?P<b>[^/]+)/(?P<c>.+)$"
```

## Troubleshooting

### Pattern Not Matching

**Problem:** Files aren't being matched

**Solutions:**

1. **Check the actual file paths:**
   ```bash
   # See what paths the app receives
   # Look for "sample file path" in logs
   ```

2. **Test the pattern:**
   ```bash
   ./config-validator test-pattern \
     -type regex \
     -pattern "YOUR_PATTERN" \
     -file "ACTUAL_FILE_PATH"
   ```

3. **Check for common issues:**
   - Missing `^` or `$` anchors
   - Wrong pattern type (prefix vs glob vs regex)
   - Typos in the pattern
   - Case sensitivity

### Variables Not Extracted

**Problem:** Variables are empty or missing

**Solutions:**

1. **Use named capture groups:**
   ```yaml
   # ✅ Correct - named group
   pattern: "^examples/(?P<lang>[^/]+)/(?P<file>.+)$"
   
   # ❌ Wrong - unnamed group
   pattern: "^examples/([^/]+)/(.+)$"
   ```

2. **Check variable names match:**
   ```yaml
   # Pattern extracts "lang"
   pattern: "^examples/(?P<lang>[^/]+)/(?P<file>.+)$"
   
   # Transform must use "lang" (not "language")
   path_transform: "docs/${lang}/${file}"
   ```

### Path Transformation Fails

**Problem:** Target path is incorrect

**Solutions:**

1. **Check variable names:**
   ```bash
   # See what variables were extracted
   ./config-validator test-pattern ... 
   ```

2. **Test transformation:**
   ```bash
   ./config-validator test-transform \
     -source "examples/go/main.go" \
     -template "docs/${lang}/${file}" \
     -vars "lang=go,file=main.go"
   ```

3. **Use built-in variables:**
   ```yaml
   # If custom variables don't work, try built-ins
   path_transform: "${dir}/${filename}"
   ```

### Files Matched Multiple Times

**Problem:** Same file matches multiple rules

**Solution:** This is expected! Files can match multiple rules and be copied to multiple targets. If you want only one rule to match, make patterns mutually exclusive:

```yaml
copy_rules:
  # Only Go files
  - name: "Go examples"
    source_pattern:
      type: "regex"
      pattern: "^examples/go/(?P<file>.+)$"
  
  # Only Python files (won't match Go)
  - name: "Python examples"
    source_pattern:
      type: "regex"
      pattern: "^examples/python/(?P<file>.+)$"
```

## Advanced Examples

### Multi-Version Support

**Use Case:** Support multiple SDK versions

**Structure:**
```
examples/
  node/
    v5.x/
      connect.js
    v6.x/
      connect.js
```

**Configuration:**
```yaml
source_pattern:
  type: "regex"
  pattern: "^examples/(?P<lang>[^/]+)/(?P<version>v[0-9]+\\.x)/(?P<file>.+)$"
targets:
  - path_transform: "docs/${lang}/${version}/${file}"
```

### Optional Path Segments

**Use Case:** Files may or may not have a category

**Structure:**
```
examples/
  go/
    main.go              # No category
    database/
      connect.go         # Has category
```

**Configuration:**
```yaml
source_pattern:
  type: "regex"
  pattern: "^examples/(?P<lang>[^/]+)/(?:(?P<category>[^/]+)/)?(?P<file>[^/]+)$"
targets:
  - path_transform: "docs/${lang}/${category}${file}"
    # Note: If category is empty, path will be docs/go/main.go
    #       If category exists, path will be docs/go/database/connect.go
```

### File Extension Filtering

**Use Case:** Only copy specific file types

**Configuration:**
```yaml
# Only .go files
source_pattern:
  type: "regex"
  pattern: "^examples/(?P<lang>[^/]+)/(?P<file>.+\\.go)$"

# Only .js and .ts files
source_pattern:
  type: "regex"
  pattern: "^examples/(?P<lang>[^/]+)/(?P<file>.+\\.(js|ts))$"

# Everything except test files
source_pattern:
  type: "regex"
  pattern: "^examples/(?P<lang>[^/]+)/(?P<file>(?!.*_test\\.go$).+)$"
```

### Nested Project Structure

**Use Case:** Complex nested project structure

**Structure:**
```
projects/
  backend/
    services/
      api/
        examples/
          auth.go
```

**Configuration:**
```yaml
source_pattern:
  type: "regex"
  pattern: "^projects/(?P<project>[^/]+)/services/(?P<service>[^/]+)/examples/(?P<file>.+)$"
targets:
  - path_transform: "docs/${project}/${service}/${file}"
```

## Quick Reference Card

### Pattern Type Decision Tree

```
Do you need to extract variables from the path?
├─ No → Use PREFIX pattern
│   └─ Fast, simple, copies directory trees
│
└─ Yes → Do you need complex matching?
    ├─ No → Use GLOB pattern
    │   └─ Wildcards: *, **, ?
    │
    └─ Yes → Use REGEX pattern
        └─ Full control, named capture groups
```

### Common Regex Patterns

```regex
[^/]+           # Match one or more non-slash characters (directory/file name)
.+              # Match one or more of any character
.*              # Match zero or more of any character
[0-9]+          # Match one or more digits
[a-z]+          # Match one or more lowercase letters
(foo|bar)       # Match either "foo" or "bar"
\.go$           # Match files ending with .go
^examples/      # Match paths starting with examples/
(?P<name>...)   # Named capture group (extracts variable)
```

### Built-in Variables Quick Reference

```yaml
${path}      # Full path: examples/go/main.go
${filename}  # Filename: main.go
${dir}       # Directory: examples/go
${ext}       # Extension: .go
${name}      # Name without ext: main
```

### Testing Commands

```bash
# Test pattern matching
./config-validator test-pattern \
  -type regex \
  -pattern "^examples/(?P<lang>[^/]+)/(?P<file>.+)$" \
  -file "examples/go/main.go"

# Test path transformation
./config-validator test-transform \
  -source "examples/go/main.go" \
  -template "docs/${lang}/${file}" \
  -vars "lang=go,file=main.go"

# Validate configuration
./config-validator validate -config copier-config.yaml -v
```

## See Also

- [Local Testing](LOCAL-TESTING.md) - How to test locally
- [Quick Reference](QUICK-REFERENCE.md) - Quick command reference
- [Webhook Testing](WEBHOOK-TESTING.md) - Testing with webhooks

