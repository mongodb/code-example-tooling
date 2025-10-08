# Pattern Matching Cheat Sheet

Quick reference for pattern matching in examples-copier.

## Pattern Types at a Glance

| Type       | Use When                              | Example                         | Extracts Variables?           |
|------------|---------------------------------------|---------------------------------|-------------------------------|
| **Prefix** | Simple directory matching             | `examples/`                     | ✅ Yes (prefix, relative_path) |
| **Glob**   | Wildcard matching                     | `**/*.go`                       | ❌ No                          |
| **Regex**  | Complex patterns, variable extraction | `^examples/(?P<lang>[^/]+)/.*$` | ✅ Yes (custom)                |

## Prefix Patterns

### Syntax
```yaml
source_pattern:
  type: "prefix"
  pattern: "examples/"
```

### Examples
| Pattern     | Matches               | Doesn't Match          |
|-------------|-----------------------|------------------------|
| `examples/` | `examples/go/main.go` | `src/examples/test.go` |
| `src/`      | `src/main.go`         | `examples/src/test.go` |
| `docs/api/` | `docs/api/readme.md`  | `docs/guide/api.md`    |

### Variables
- `${matched_prefix}` - The matched prefix
- `${relative_path}` - Path after the prefix

## Glob Patterns

### Wildcards
| Symbol | Matches                 | Example                     |
|--------|-------------------------|-----------------------------|
| `*`    | Any characters (no `/`) | `*.go` → `main.go`          |
| `**`   | Any directories         | `**/*.go` → `a/b/c/main.go` |
| `?`    | Single character        | `test?.go` → `test1.go`     |

### Examples
| Pattern            | Matches                | Doesn't Match |
|--------------------|------------------------|---------------|
| `*.go`             | `main.go`              | `src/main.go` |
| `**/*.go`          | `a/b/c/main.go`        | `main.py`     |
| `examples/**/*.js` | `examples/node/app.js` | `src/app.js`  |
| `test?.go`         | `test1.go`, `testA.go` | `test12.go`   |

## Regex Patterns

### Common Building Blocks

| Pattern      | Matches                     | Example                |
|--------------|-----------------------------|------------------------|
| `[^/]+`      | One or more non-slash chars | Directory or file name |
| `.+`         | One or more any chars       | Rest of path           |
| `.*`         | Zero or more any chars      | Optional content       |
| `[0-9]+`     | One or more digits          | Version numbers        |
| `(foo\|bar)` | Either foo or bar           | Specific values        |
| `\.go$`      | Ends with .go               | File extension         |
| `^examples/` | Starts with examples/       | Path prefix            |

### Named Capture Groups

```regex
(?P<name>pattern)
```

**Example:**
```regex
^examples/(?P<lang>[^/]+)/(?P<file>.+)$
```

Extracts:
- `lang` from first directory
- `file` from rest of path

### Common Patterns

#### Language + File
```regex
^examples/(?P<lang>[^/]+)/(?P<file>.+)$
```
- `examples/go/main.go` → `lang=go, file=main.go`

#### Language + Category + File
```regex
^examples/(?P<lang>[^/]+)/(?P<category>[^/]+)/(?P<file>.+)$
```
- `examples/go/database/connect.go` → `lang=go, category=database, file=connect.go`

#### Project + Rest
```regex
^generated-examples/(?P<project>[^/]+)/(?P<rest>.+)$
```
- `generated-examples/app/cmd/main.go` → `project=app, rest=cmd/main.go`

#### Version Support
```regex
^examples/(?P<lang>[^/]+)/(?P<version>v[0-9]+\\.x)/(?P<file>.+)$
```
- `examples/node/v6.x/app.js` → `lang=node, version=v6.x, file=app.js`

#### Type + Language + File
```regex
^source/examples/(?P<type>generated|manual)/(?P<lang>[^/]+)/(?P<file>.+)$
```
- `source/examples/generated/node/app.js` → `type=generated, lang=node, file=app.js`

## Path Transformation

### Syntax
```yaml
path_transform: "docs/${lang}/${file}"
```

### Built-in Variables

| Variable      | Value for `examples/go/database/connect.go` |
|---------------|---------------------------------------------|
| `${path}`     | `examples/go/database/connect.go`           |
| `${filename}` | `connect.go`                                |
| `${dir}`      | `examples/go/database`                      |
| `${ext}`      | `.go`                                       |
| `${name}`     | `connect`                                   |

### Common Transformations

| Transform                          | Input                    | Output                     |
|------------------------------------|--------------------------|----------------------------|
| `${path}`                          | `examples/go/main.go`    | `examples/go/main.go`      |
| `docs/${path}`                     | `examples/go/main.go`    | `docs/examples/go/main.go` |
| `docs/${relative_path}`            | `examples/go/main.go`    | `docs/go/main.go`          |
| `${lang}/${file}`                  | `examples/go/main.go`    | `go/main.go`               |
| `docs/${lang}/${category}/${file}` | `examples/go/db/conn.go` | `docs/go/db/conn.go`       |

## Complete Examples

### Example 1: Simple Copy
```yaml
source_pattern:
  type: "prefix"
  pattern: "examples/"
targets:
  - path_transform: "docs/${path}"
```
**Result:** `examples/go/main.go` → `docs/examples/go/main.go`

### Example 2: Language-Based
```yaml
source_pattern:
  type: "regex"
  pattern: "^examples/(?P<lang>[^/]+)/(?P<file>.+)$"
targets:
  - path_transform: "docs/code-examples/${lang}/${file}"
```
**Result:** `examples/go/main.go` → `docs/code-examples/go/main.go`

### Example 3: Categorized
```yaml
source_pattern:
  type: "regex"
  pattern: "^examples/(?P<lang>[^/]+)/(?P<category>[^/]+)/(?P<file>.+)$"
targets:
  - path_transform: "docs/${lang}/${category}/${file}"
```
**Result:** `examples/go/database/connect.go` → `docs/go/database/connect.go`

### Example 4: Glob for Extensions
```yaml
source_pattern:
  type: "glob"
  pattern: "examples/**/*.go"
targets:
  - path_transform: "docs/${path}"
```
**Result:** `examples/go/auth/login.go` → `docs/examples/go/auth/login.go`

### Example 5: Project-Based
```yaml
source_pattern:
  type: "regex"
  pattern: "^generated-examples/(?P<project>[^/]+)/(?P<rest>.+)$"
targets:
  - path_transform: "examples/${project}/${rest}"
```
**Result:** `generated-examples/app/cmd/main.go` → `examples/app/cmd/main.go`

## Testing Commands

### Test Pattern
```bash
./config-validator test-pattern \
  -type regex \
  -pattern "^examples/(?P<lang>[^/]+)/(?P<file>.+)$" \
  -file "examples/go/main.go"
```

### Test Transform
```bash
./config-validator test-transform \
  -source "examples/go/main.go" \
  -template "docs/${lang}/${file}" \
  -vars "lang=go,file=main.go"
```

### Validate Config
```bash
./config-validator validate -config copier-config.yaml -v
```

## Decision Tree

```
What do you need?
│
├─ Copy entire directory tree
│  └─ Use PREFIX pattern
│     pattern: "examples/"
│     transform: "docs/${path}"
│
├─ Match by file extension
│  └─ Use GLOB pattern
│     pattern: "**/*.go"
│     transform: "docs/${path}"
│
├─ Extract language from path
│  └─ Use REGEX pattern
│     pattern: "^examples/(?P<lang>[^/]+)/(?P<file>.+)$"
│     transform: "docs/${lang}/${file}"
│
└─ Complex matching with multiple variables
   └─ Use REGEX pattern
      pattern: "^examples/(?P<lang>[^/]+)/(?P<category>[^/]+)/(?P<file>.+)$"
      transform: "docs/${lang}/${category}/${file}"
```

## Common Mistakes

### ❌ Missing Anchors
```yaml
# Wrong - matches partial paths
pattern: "examples/(?P<lang>[^/]+)/(?P<file>.+)"

# Right - matches full path
pattern: "^examples/(?P<lang>[^/]+)/(?P<file>.+)$"
```

### ❌ Wrong Character Class
```yaml
# Wrong - .+ matches slashes too
pattern: "^examples/(?P<lang>.+)/(?P<file>.+)$"

# Right - [^/]+ doesn't match slashes
pattern: "^examples/(?P<lang>[^/]+)/(?P<file>.+)$"
```

### ❌ Unnamed Groups
```yaml
# Wrong - doesn't extract variables
pattern: "^examples/([^/]+)/(.+)$"

# Right - named groups extract variables
pattern: "^examples/(?P<lang>[^/]+)/(?P<file>.+)$"
```

### ❌ Variable Name Mismatch
```yaml
# Pattern extracts "lang"
pattern: "^examples/(?P<lang>[^/]+)/(?P<file>.+)$"

# Wrong - uses "language"
path_transform: "docs/${language}/${file}"

# Right - uses "lang"
path_transform: "docs/${lang}/${file}"
```

## Tips

1. **Start simple** - Use prefix, then add regex when needed
2. **Test first** - Use `config-validator` before deploying
3. **Use anchors** - Always use `^` and `$` in regex
4. **Be specific** - Use `[^/]+` instead of `.+` for directories
5. **Name clearly** - Use descriptive variable names like `lang`, not `a`
6. **Check logs** - Look for "sample file path" to see actual paths

## See Also

- [Full Pattern Matching Guide](PATTERN-MATCHING-GUIDE.md)
- [Local Testing](LOCAL-TESTING.md)

