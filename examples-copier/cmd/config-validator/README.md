# config-validator

Command-line tool for validating and testing examples-copier configurations.

## Overview

The `config-validator` tool helps you:
- Validate configuration files
- Test pattern matching
- Test path transformations
- Convert legacy JSON configs to YAML
- Debug configuration issues

## Installation

```bash
cd examples-copier
go build -o config-validator ./cmd/config-validator
```

## Commands

### validate

Validate a configuration file.

**Usage:**
```bash
./config-validator validate -config <file> [-v]
```

**Options:**
- `-config` - Path to configuration file (required)
- `-v` - Verbose output (optional)

**Examples:**

```bash
# Validate YAML config
./config-validator validate -config copier-config.yaml

# Validate with verbose output
./config-validator validate -config copier-config.yaml -v

# Validate JSON config
./config-validator validate -config config.json
```

**Output:**
```
✅ Configuration is valid!

Summary:
  Source: mongodb/docs-code-examples
  Branch: main
  Rules: 2
  
Rule 1: Copy generated examples
  Pattern: regex - ^generated-examples/(?P<project>[^/]+)/(?P<rest>.+)$
  Targets: 1
  
Rule 2: Copy all generated examples
  Pattern: prefix - generated-examples/
  Targets: 1
```

### test-pattern

Test if a pattern matches a file path and see what variables are extracted.

**Usage:**
```bash
./config-validator test-pattern -type <type> -pattern <pattern> -file <path>
```

**Options:**
- `-type` - Pattern type: `prefix`, `glob`, or `regex` (required)
- `-pattern` - Pattern to test (required)
- `-file` - File path to test against (required)

**Examples:**

```bash
# Test regex pattern
./config-validator test-pattern \
  -type regex \
  -pattern "^examples/(?P<lang>[^/]+)/(?P<file>.+)$" \
  -file "examples/go/main.go"

# Test prefix pattern
./config-validator test-pattern \
  -type prefix \
  -pattern "examples/" \
  -file "examples/go/main.go"

# Test glob pattern
./config-validator test-pattern \
  -type glob \
  -pattern "**/*.go" \
  -file "examples/go/main.go"
```

**Output (regex):**
```
✅ Pattern matched!

Extracted variables:
  lang = go
  file = main.go
```

**Output (prefix):**
```
✅ Pattern matched!

Extracted variables:
  matched_prefix = examples
  relative_path = go/main.go
```

**Output (no match):**
```
❌ Pattern did not match
```

### test-transform

Test path transformation with variables.

**Usage:**
```bash
./config-validator test-transform -source <path> -template <template> -vars <vars>
```

**Options:**
- `-source` - Source file path (required)
- `-template` - Path transformation template (required)
- `-vars` - Variables as comma-separated key=value pairs (required)

**Examples:**

```bash
# Test transformation with custom variables
./config-validator test-transform \
  -source "examples/go/main.go" \
  -template "docs/${lang}/${file}" \
  -vars "lang=go,file=main.go"

# Test with built-in variables only
./config-validator test-transform \
  -source "examples/go/database/connect.go" \
  -template "docs/${dir}/${filename}" \
  -vars ""

# Test complex transformation
./config-validator test-transform \
  -source "examples/go/database/connect.go" \
  -template "docs/${lang}/${category}/${name}_example${ext}" \
  -vars "lang=go,category=database"
```

**Output:**
```
✅ Transform successful!

Source: examples/go/main.go
Result: docs/go/main.go

Variables used:
  lang = go
  file = main.go
  path = examples/go/main.go
  filename = main.go
  dir = examples/go
  ext = .go
  name = main
```

### convert

Convert legacy JSON configuration to YAML format.

**Usage:**
```bash
./config-validator convert -input <file> -output <file>
```

**Options:**
- `-input` - Input JSON file (required)
- `-output` - Output YAML file (required)

**Example:**

```bash
./config-validator convert -input config.json -output copier-config.yaml
```

**Output:**
```
✅ Conversion successful!

Converted 2 legacy rules to YAML format.
Output written to: copier-config.yaml

Next steps:
1. Review the generated copier-config.yaml
2. Enhance with new features (regex patterns, path transforms)
3. Validate: ./config-validator validate -config copier-config.yaml
```

## Common Use Cases

### Debugging Pattern Matching

When files aren't matching your pattern:

1. **Get actual file paths from logs:**
   ```bash
   grep "sample file path" logs/app.log
   ```

2. **Test your pattern:**
   ```bash
   ./config-validator test-pattern \
     -type regex \
     -pattern "YOUR_PATTERN" \
     -file "ACTUAL_FILE_PATH"
   ```

3. **Adjust pattern and test again**

### Debugging Path Transformation

When files are copied to wrong locations:

1. **Test the transformation:**
   ```bash
   ./config-validator test-transform \
     -source "SOURCE_PATH" \
     -template "YOUR_TEMPLATE" \
     -vars "key1=value1,key2=value2"
   ```

2. **Verify variable names match**

3. **Check built-in variables are available**

### Validating Before Deployment

Before deploying a new configuration:

```bash
# Validate the config
./config-validator validate -config copier-config.yaml -v

# Test with sample file paths
./config-validator test-pattern \
  -type regex \
  -pattern "^examples/(?P<lang>[^/]+)/(?P<file>.+)$" \
  -file "examples/go/main.go"

# Test path transformation
./config-validator test-transform \
  -source "examples/go/main.go" \
  -template "docs/${lang}/${file}" \
  -vars "lang=go,file=main.go"
```

### Migrating from JSON to YAML

```bash
# Convert
./config-validator convert -input config.json -output copier-config.yaml

# Validate
./config-validator validate -config copier-config.yaml -v

# Test patterns
./config-validator test-pattern \
  -type prefix \
  -pattern "examples/" \
  -file "examples/go/main.go"
```

## Exit Codes

- `0` - Success
- `1` - Validation failed or error occurred

## Tips

1. **Use verbose mode** (`-v`) to see detailed validation results

2. **Test patterns before deploying** to avoid issues in production

3. **Use actual file paths** from your repository when testing

4. **Validate after every config change** to catch errors early

5. **Keep test commands** in a script for regression testing

## Examples

### Complete Workflow

```bash
# 1. Create config
cat > copier-config.yaml << EOF
source_repo: "myorg/source-repo"
source_branch: "main"

copy_rules:
  - name: "Copy Go examples"
    source_pattern:
      type: "regex"
      pattern: "^examples/(?P<lang>[^/]+)/(?P<file>.+)$"
    targets:
      - repo: "myorg/target-repo"
        branch: "main"
        path_transform: "docs/code-examples/\${lang}/\${file}"
EOF

# 2. Validate
./config-validator validate -config copier-config.yaml -v

# 3. Test pattern
./config-validator test-pattern \
  -type regex \
  -pattern "^examples/(?P<lang>[^/]+)/(?P<file>.+)$" \
  -file "examples/go/database/connect.go"

# 4. Test transformation
./config-validator test-transform \
  -source "examples/go/database/connect.go" \
  -template "docs/code-examples/\${lang}/\${file}" \
  -vars "lang=go,file=database/connect.go"
```

## See Also

- [Configuration Guide](../../docs/CONFIGURATION-GUIDE.md) - Complete configuration reference
- [Pattern Matching Guide](../../docs/PATTERN-MATCHING-GUIDE.md) - Pattern matching help
- [FAQ](../../docs/FAQ.md) - Frequently asked questions (includes JSON to YAML conversion)
- [Quick Reference](../../QUICK-REFERENCE.md) - All commands

