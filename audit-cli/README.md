# audit-cli

A Go CLI tool for extracting and analyzing code examples from MongoDB documentation written in reStructuredText (RST).

## Table of Contents

- [Overview](#overview)
- [Installation](#installation)
- [Usage](#usage)
  - [Extract Commands](#extract-commands)
  - [Search Commands](#search-commands)
- [Development](#development)
  - [Project Structure](#project-structure)
  - [Adding New Commands](#adding-new-commands)
  - [Testing](#testing)
  - [Code Patterns](#code-patterns)
- [Supported RST Directives](#supported-rst-directives)

## Overview

This CLI tool helps maintain code quality across MongoDB's documentation by:

1. **Extracting code examples** from RST files into individual, testable files
2. **Searching extracted code** for specific patterns or substrings
3. **Following include directives** to process entire documentation trees
4. **Handling MongoDB-specific conventions** like steps files, extracts, and template variables

## Installation

### Build from Source

```bash
cd audit-cli
go build
```

This creates an `audit-cli` executable in the current directory.

### Run Without Building

```bash
cd audit-cli
go run main.go [command] [flags]
```

## Usage

The CLI is organized into parent commands with subcommands:

```
audit-cli
├── extract          # Extract content from RST files
│   └── code-examples
└── search           # Search through extracted content
    └── find-string
```

### Extract Commands

#### `extract code-examples`

Extract code examples from reStructuredText files into individual files.

**Basic Usage:**

```bash
# Extract from a single file
./audit-cli extract code-examples path/to/file.rst -o ./output

# Extract from a directory (non-recursive)
./audit-cli extract code-examples path/to/docs -o ./output

# Extract recursively from all subdirectories
./audit-cli extract code-examples path/to/docs -o ./output -r

# Follow include directives
./audit-cli extract code-examples path/to/file.rst -o ./output -f

# Combine recursive scanning and include following
./audit-cli extract code-examples path/to/docs -o ./output -r -f

# Dry run (show what would be extracted without writing files)
./audit-cli extract code-examples path/to/file.rst -o ./output --dry-run

# Verbose output
./audit-cli extract code-examples path/to/file.rst -o ./output -v
```

**Flags:**

- `-o, --output <dir>` - Output directory for extracted files (default: `./output`)
- `-r, --recursive` - Recursively scan directories for RST files
- `-f, --follow-includes` - Follow `.. include::` directives in RST files
- `--dry-run` - Show what would be extracted without writing files
- `-v, --verbose` - Show detailed processing information

**Output Format:**

Extracted files are named: `{source-base}.{directive-type}.{index}.{ext}`

Examples:
- `my-doc.code-block.1.js` - First code-block from my-doc.rst
- `my-doc.literalinclude.2.py` - Second literalinclude from my-doc.rst
- `my-doc.io-code-block.1.input.js` - Input from first io-code-block
- `my-doc.io-code-block.1.output.json` - Output from first io-code-block

**Report:**

After extraction, a report is displayed showing:
- Number of files traversed
- Number of output files written
- Code examples by language
- Code examples by directive type

### Search Commands

#### `search find-string`

Search through extracted code example files for a specific substring.

**Basic Usage:**

```bash
# Search in a single file
./audit-cli search find-string path/to/file.js "substring"

# Search in a directory (non-recursive)
./audit-cli search find-string path/to/output "substring"

# Search recursively
./audit-cli search find-string path/to/output "substring" -r

# Verbose output (show file paths and language breakdown)
./audit-cli search find-string path/to/output "substring" -r -v
```

**Flags:**

- `-r, --recursive` - Recursively search all files in subdirectories
- `-v, --verbose` - Show file paths and language breakdown

**Report:**

The search report shows:
- Number of files scanned
- Number of files containing the substring (each file counted once)

With `-v` flag, also shows:
- List of file paths where substring appears
- Count broken down by language (file extension)

## Development

### Project Structure

```
audit-cli/
├── main.go                          # CLI entry point
├── commands/                        # Command implementations
│   ├── extract/                     # Extract parent command
│   │   ├── extract.go              # Parent command definition
│   │   └── code-examples/          # Code examples subcommand
│   │       ├── code_examples.go    # Command logic
│   │       ├── code_examples_test.go # Tests
│   │       ├── parser.go           # RST directive parsing
│   │       ├── writer.go           # File writing logic
│   │       ├── report.go           # Report generation
│   │       ├── types.go            # Type definitions
│   │       └── language.go         # Language normalization
│   └── search/                      # Search parent command
│       ├── search.go               # Parent command definition
│       └── find-string/            # Find string subcommand
│           ├── find_string.go      # Command logic
│           ├── types.go            # Type definitions
│           └── report.go           # Report generation
├── internal/                        # Internal packages
│   └── rst/                        # RST parsing utilities
│       ├── include.go              # Include directive resolution
│       ├── traverse.go             # Directory traversal
│       └── directive.go            # Directive parsing
└── testdata/                        # Test fixtures
    ├── input-files/                # Test RST files
    │   └── source/                 # Source directory (required)
    │       ├── *.rst               # Test files
    │       ├── includes/           # Included RST files
    │       └── code-examples/      # Code files for literalinclude
    └── expected-output/            # Expected extraction results
```

### Adding New Commands

#### 1. Adding a New Subcommand to an Existing Parent

Example: Adding `extract tables` subcommand

1. **Create the subcommand directory:**
   ```bash
   mkdir -p commands/extract/tables
   ```

2. **Create the command file** (`commands/extract/tables/tables.go`):
   ```go
   package tables

   import (
       "github.com/spf13/cobra"
   )

   func NewTablesCommand() *cobra.Command {
       cmd := &cobra.Command{
           Use:   "tables [filepath]",
           Short: "Extract tables from RST files",
           Args:  cobra.ExactArgs(1),
           RunE: func(cmd *cobra.Command, args []string) error {
               // Implementation here
               return nil
           },
       }

       // Add flags
       cmd.Flags().StringP("output", "o", "./output", "Output directory")

       return cmd
   }
   ```

3. **Register the subcommand** in `commands/extract/extract.go`:
   ```go
   import (
       "github.com/mongodb/code-example-tooling/audit-cli/commands/extract/tables"
   )

   func NewExtractCommand() *cobra.Command {
       cmd := &cobra.Command{...}

       cmd.AddCommand(codeexamples.NewCodeExamplesCommand())
       cmd.AddCommand(tables.NewTablesCommand())  // Add this line

       return cmd
   }
   ```

#### 2. Adding a New Parent Command

Example: Adding `analyze` parent command

1. **Create the parent directory:**
   ```bash
   mkdir -p commands/analyze
   ```

2. **Create the parent command** (`commands/analyze/analyze.go`):
   ```go
   package analyze

   import (
       "github.com/spf13/cobra"
   )

   func NewAnalyzeCommand() *cobra.Command {
       cmd := &cobra.Command{
           Use:   "analyze",
           Short: "Analyze extracted content",
       }

       // Add subcommands here

       return cmd
   }
   ```

3. **Register in main.go:**
   ```go
   import (
       "github.com/mongodb/code-example-tooling/audit-cli/commands/analyze"
   )

   func main() {
       rootCmd.AddCommand(extract.NewExtractCommand())
       rootCmd.AddCommand(search.NewSearchCommand())
       rootCmd.AddCommand(analyze.NewAnalyzeCommand())  // Add this line
   }
   ```



### Testing

#### Running Tests

```bash
# Run all tests
cd audit-cli
go test ./...

# Run tests for a specific package
go test ./commands/extract/code-examples -v

# Run a specific test
go test ./commands/extract/code-examples -run TestRecursiveDirectoryScanning -v

# Run tests with coverage
go test ./... -cover
```

#### Test Structure

Tests use a table-driven approach with test fixtures in the `testdata/` directory:

- **Input files**: `testdata/input-files/source/` - RST files and referenced code
- **Expected output**: `testdata/expected-output/` - Expected extracted files
- **Test pattern**: Compare actual extraction output against expected files

**Note**: The `testdata` directory name is special in Go - it's automatically ignored during builds, which is important since it contains non-Go files (`.cpp`, `.rst`, etc.).

#### Adding New Tests

1. **Create test input files** in `testdata/input-files/source/`:
   ```bash
   # Create a new test RST file
   cat > testdata/input-files/source/my-test.rst << 'EOF'
   .. code-block:: javascript

      console.log("Hello, World!");
   EOF
   ```

2. **Generate expected output**:
   ```bash
   ./audit-cli extract code-examples testdata/input-files/source/my-test.rst \
     -o testdata/expected-output
   ```

3. **Verify the output** is correct before committing

4. **Add test case** in the appropriate `*_test.go` file:
   ```go
   func TestMyNewFeature(t *testing.T) {
       testDataDir := filepath.Join("..", "..", "..", "testdata")
       inputFile := filepath.Join(testDataDir, "input-files", "source", "my-test.rst")
       expectedDir := filepath.Join(testDataDir, "expected-output")

       tempDir, err := os.MkdirTemp("", "test-*")
       if err != nil {
           t.Fatalf("Failed to create temp directory: %v", err)
       }
       defer os.RemoveAll(tempDir)

       report, err := RunExtract(inputFile, tempDir, false, false, false, false)
       if err != nil {
           t.Fatalf("RunExtract failed: %v", err)
       }

       // Add assertions here
   }
   ```

#### Test Conventions

- **Relative paths**: Tests use `filepath.Join("..", "..", "..", "testdata")` to reference test data (three levels up from `commands/extract/code-examples/`)
- **Temporary directories**: Use `os.MkdirTemp()` for test output, clean up with `defer os.RemoveAll()`
- **Exact content matching**: Tests compare byte-for-byte content
- **No trailing newlines**: Expected output files should not have trailing blank lines

#### Updating Expected Output

If you've changed the parsing logic and need to regenerate expected output:

```bash
cd audit-cli

# Update all expected outputs
./audit-cli extract code-examples testdata/input-files/source/literalinclude-test.rst \
  -o testdata/expected-output

./audit-cli extract code-examples testdata/input-files/source/code-block-test.rst \
  -o testdata/expected-output

./audit-cli extract code-examples testdata/input-files/source/nested-code-block-test.rst \
  -o testdata/expected-output

./audit-cli extract code-examples testdata/input-files/source/io-code-block-test.rst \
  -o testdata/expected-output

./audit-cli extract code-examples testdata/input-files/source/include-test.rst \
  -o testdata/expected-output -f
```

**Important**: Always verify the new output is correct before committing!

### Code Patterns

#### 1. Command Structure Pattern

All commands follow this pattern:

```go
package mycommand

import "github.com/spf13/cobra"

func NewMyCommand() *cobra.Command {
    var flagVar string

    cmd := &cobra.Command{
        Use:   "my-command [args]",
        Short: "Brief description",
        Long:  "Detailed description",
        Args:  cobra.ExactArgs(1),  // Or MinimumNArgs, etc.
        RunE: func(cmd *cobra.Command, args []string) error {
            // Get flag values
            flagValue, _ := cmd.Flags().GetString("flag-name")

            // Call the main logic function
            return RunMyCommand(args[0], flagValue)
        },
    }

    // Define flags
    cmd.Flags().StringVarP(&flagVar, "flag-name", "f", "default", "Description")

    return cmd
}

// Separate logic function for testability
func RunMyCommand(arg string, flagValue string) error {
    // Implementation here
    return nil
}
```

**Why this pattern?**
- Separates command definition from logic
- Makes logic testable without Cobra
- Consistent across all commands

#### 2. Error Handling Pattern

Use descriptive error wrapping:

```go
import "fmt"

// Wrap errors with context
file, err := os.Open(filePath)
if err != nil {
    return fmt.Errorf("failed to open file %s: %w", filePath, err)
}

// Check for specific conditions
if !fileInfo.IsDir() {
    return fmt.Errorf("path %s is not a directory", path)
}
```

#### 3. File Processing Pattern

Use the scanner pattern for line-by-line processing:

```go
import (
    "bufio"
    "os"
)

func processFile(filePath string) error {
    file, err := os.Open(filePath)
    if err != nil {
        return fmt.Errorf("failed to open file: %w", err)
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)
    lineNum := 0

    for scanner.Scan() {
        lineNum++
        line := scanner.Text()

        // Process line
    }

    if err := scanner.Err(); err != nil {
        return fmt.Errorf("error reading file: %w", err)
    }

    return nil
}
```

#### 4. Directory Traversal Pattern

Use `filepath.Walk` for recursive traversal:

```go
import (
    "os"
    "path/filepath"
)

func traverseDirectory(rootPath string, recursive bool) ([]string, error) {
    var files []string

    err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }

        // Skip subdirectories if not recursive
        if !recursive && info.IsDir() && path != rootPath {
            return filepath.SkipDir
        }

        // Collect files
        if !info.IsDir() {
            files = append(files, path)
        }

        return nil
    })

    return files, err
}
```

#### 5. Testing Pattern

Use table-driven tests where appropriate:

```go
func TestLanguageNormalization(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
    }{
        {"TypeScript", "ts", "typescript"},
        {"C++", "c++", "cpp"},
        {"Golang", "golang", "go"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := NormalizeLanguage(tt.input)
            if result != tt.expected {
                t.Errorf("NormalizeLanguage(%q) = %q, want %q",
                    tt.input, result, tt.expected)
            }
        })
    }
}
```

#### 6. Verbose Output Pattern

Use a consistent pattern for verbose logging:

```go
func processWithVerbose(filePath string, verbose bool) error {
    if verbose {
        fmt.Printf("Processing: %s\n", filePath)
    }

    // Do work

    if verbose {
        fmt.Printf("Completed: %s\n", filePath)
    }

    return nil
}
```



## Supported RST Directives

The tool extracts code examples from the following reStructuredText directives:

### 1. `literalinclude`

Extracts code from external files with support for partial extraction and dedenting.

**Syntax:**
```rst
.. literalinclude:: /path/to/file.py
   :language: python
   :start-after: start-tag
   :end-before: end-tag
   :dedent:
```

**Supported Options:**
- `:language:` - Specifies the programming language (normalized: `ts` → `typescript`, `c++` → `cpp`, `golang` → `go`)
- `:start-after:` - Extract content after this tag (skips the entire line containing the tag)
- `:end-before:` - Extract content before this tag (cuts before the entire line containing the tag)
- `:dedent:` - Remove common leading whitespace from the extracted content

**Example:**

Given `code-examples/example.py`:
```python
def main():
    # start-example
    result = calculate(42)
    print(result)
    # end-example
```

And RST:
```rst
.. literalinclude:: /code-examples/example.py
   :language: python
   :start-after: start-example
   :end-before: end-example
   :dedent:
```

Extracts:
```python
result = calculate(42)
print(result)
```

### 2. `code-block`

Inline code blocks with automatic dedenting based on the first line's indentation.

**Syntax:**
```rst
.. code-block:: javascript
   :copyable: false
   :emphasize-lines: 2,3

   const greeting = "Hello, World!";
   console.log(greeting);
```

**Supported Options:**
- Language argument - `.. code-block:: javascript` (optional, defaults to `txt`)
- `:language:` - Alternative way to specify language
- `:copyable:` - Parsed but not used for extraction
- `:emphasize-lines:` - Parsed but not used for extraction

**Automatic Dedenting:**

The content is automatically dedented based on the indentation of the first content line. For example:

```rst
.. note::

   .. code-block:: python

      def hello():
          print("Hello")
```

The code has 6 spaces of indentation (3 from `note`, 3 from `code-block`). The tool automatically removes these 6 spaces, resulting in:

```python
def hello():
    print("Hello")
```

### 3. `io-code-block`

Input/output code blocks for interactive examples with nested sub-directives.

**Syntax:**
```rst
.. io-code-block::
   :copyable: true

   .. input::
      :language: javascript

      db.restaurants.aggregate([
         { $match: { category: "cafe" } }
      ])

   .. output::
      :language: json

      [
         { _id: 1, category: 'café', status: 'Open' }
      ]
```

**Supported Options:**
- `:copyable:` - Parsed but not used for extraction
- Nested `.. input::` sub-directive (required)
  - Can have filepath argument: `.. input:: /path/to/file.js`
  - Or inline content with `:language:` option
- Nested `.. output::` sub-directive (optional)
  - Can have filepath argument: `.. output:: /path/to/output.txt`
  - Or inline content with `:language:` option

**File-based Content:**
```rst
.. io-code-block::

   .. input:: /code-examples/query.js
      :language: javascript

   .. output:: /code-examples/result.json
      :language: json
```

**Output Files:**

Generates two files:
- `{source}.io-code-block.{index}.input.{ext}` - The input code
- `{source}.io-code-block.{index}.output.{ext}` - The output (if present)

Example: `my-doc.io-code-block.1.input.js` and `my-doc.io-code-block.1.output.json`

### 4. `include`

Follows include directives to process entire documentation trees (when `-f` flag is used).

**Syntax:**
```rst
.. include:: /includes/intro.rst
```

**Special MongoDB Conventions:**

The tool handles several MongoDB-specific include patterns:

#### Steps Files
Converts directory-based paths to filename-based paths:
- Input: `/includes/steps/run-mongodb-on-linux.rst`
- Resolves to: `/includes/steps-run-mongodb-on-linux.yaml`

#### Extracts and Release Files
Resolves ref-based includes by searching YAML files:
- Input: `/includes/extracts/install-mongodb.rst`
- Searches: `/includes/extracts-*.yaml` for `ref: install-mongodb`
- Resolves to: The YAML file containing that ref

#### Template Variables
Resolves template variables from YAML replacement sections:
```yaml
replacement:
  release_specification_default: "/includes/release/install-windows-default.rst"
```
- Input: `{{release_specification_default}}`
- Resolves to: `/includes/release/install-windows-default.rst`

**Source Directory Resolution:**

The tool walks up the directory tree to find a directory named "source" or containing a "source" subdirectory. This is used as the base for resolving relative include paths.

## Internal Packages

### `internal/rst`

Provides reusable utilities for parsing and processing RST files:

- **Include resolution** - Handles all include directive patterns
- **Directory traversal** - Recursive file scanning
- **Directive parsing** - Extracts structured data from RST directives
- **Template variable resolution** - Resolves YAML-based template variables
- **Source directory detection** - Finds the documentation root

See the code in `internal/rst/` for implementation details.

## Language Normalization

The tool normalizes language identifiers to standard file extensions:

| Input | Normalized | Extension |
|-------|-----------|-----------|
| `ts` | `typescript` | `.ts` |
| `c++` | `cpp` | `.cpp` |
| `golang` | `go` | `.go` |
| `javascript` | `javascript` | `.js` |
| `python` | `python` | `.py` |
| `shell` / `sh` | `sh` | `.sh` |
| `json` | `json` | `.json` |
| `yaml` | `yaml` | `.yaml` |
| (none) | `txt` | `.txt` |

## Contributing

When contributing to this project:

1. **Follow the established patterns** - Use the command structure, error handling, and testing patterns described above
2. **Write tests** - All new functionality should have corresponding tests
3. **Update documentation** - Keep this README up to date with new features
4. **Run tests before committing** - Ensure `go test ./...` passes
5. **Use meaningful commit messages** - Describe what changed and why

## License

[Add license information here]
