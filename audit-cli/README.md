# audit-cli

A Go CLI tool for extracting and analyzing code examples from MongoDB documentation written in reStructuredText (RST).

## Table of Contents

- [Overview](#overview)
- [Installation](#installation)
- [Usage](#usage)
  - [Extract Commands](#extract-commands)
  - [Search Commands](#search-commands)
  - [Analyze Commands](#analyze-commands)
  - [Compare Commands](#compare-commands)
  - [Count Commands](#count-commands)
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
3. **Analyzing include relationships** to understand file dependencies
4. **Comparing file contents** across documentation versions to identify differences
5. **Following include directives** to process entire documentation trees
6. **Handling MongoDB-specific conventions** like steps files, extracts, and template variables

## Installation

### Build from Source

```bash
cd audit-cli/bin
go build ../
```

This creates an `audit-cli` executable in the `bin` directory.

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
├── search           # Search through extracted content or source files
│   └── find-string
├── analyze          # Analyze RST file structures
│   ├── includes
│   └── usage
├── compare          # Compare files across versions
│   └── file-contents
└── count            # Count code examples and documentation pages
    ├── tested-examples
    └── pages
```

### Extract Commands

#### `extract code-examples`

Extract code examples from reStructuredText files into individual files. For details about what code example directives
are supported and how, refer to the [Supported rST Directives - Code Example Extraction](#code-example-extraction)
section below.

**Use Cases:**

This command helps writers:
- Examine all the code examples that make up a specific page or section
- Split out code examples into individual files for migration to test infrastructure
- Report on the number of code examples by language
- Report on the number of code examples by directive type
- Use additional commands, such as search, to find strings within specific code examples

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
- `-r, --recursive` - Recursively scan directories for RST files. If you do not provide this flag, the tool will only
  extract code examples from the top-level RST file. If you do provide this flag, the tool will recursively scan all
  subdirectories for RST files and extract code examples from all files.
- `-f, --follow-includes` - Follow `.. include::` directives in RST files. If you do not provide this flag, the tool
  will only extract code examples from the top-level RST file. If you do provide this flag, the tool will follow any
  `.. include::` directives in the RST file and extract code examples from all included files. When combined with `-r`,
  the tool will recursively scan all subdirectories for RST files and follow `.. include::` directives in all files. If
  an include filepath is *outside* the input directory, the `-r` flag would not parse it, but the `-f` flag would
  follow the include directive and parse the included file. This effectively lets you parse all the files that make up
  a single page, if you start from the page's root `.txt` file.
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

After extraction, the code extraction report shows:
- Number of files traversed
- Number of output files written
- Code examples by language
- Code examples by directive type

### Search Commands

#### `search find-string`

Search through files for a specific substring. Can search through extracted code example files or RST source files.

**Default Behavior:**
- **Case-insensitive** search (matches "curl", "CURL", "Curl", etc.)
- **Exact word matching** (excludes partial matches like "curl" in "libcurl")

Use `--case-sensitive` to make the search case-sensitive, or `--partial-match` to allow matching the substring as part
of larger words.

**Use Cases:**

This command helps writers:
- Find specific strings across documentation files or pages
  - Search for product names, command names, API methods, or other strings that may need to be updated
- Understand the number of references and impact of changes across documentation files or pages
- Identify files that need to be updated when a string needs to be changed
- Scope work related to specific changes

**Basic Usage:**

```bash
# Search in a single file (case-insensitive, exact word match)
./audit-cli search find-string path/to/file.js "curl"

# Search in a directory (non-recursive)
./audit-cli search find-string path/to/output "substring"

# Search recursively
./audit-cli search find-string path/to/output "substring" -r

# Search an RST file and all files it includes
./audit-cli search find-string path/to/source.rst "substring" -f

# Search a directory recursively and follow includes in RST files
./audit-cli search find-string path/to/source "substring" -r -f

# Verbose output (show file paths and language breakdown)
./audit-cli search find-string path/to/output "substring" -r -v

# Case-sensitive search (only matches exact case)
./audit-cli search find-string path/to/output "CURL" --case-sensitive

# Partial match (includes "curl" in "libcurl")
./audit-cli search find-string path/to/output "curl" --partial-match

# Combine flags for case-sensitive partial matching
./audit-cli search find-string path/to/output "curl" --case-sensitive --partial-match
```

**Flags:**

- `-r, --recursive` - Recursively scan directories for RST files. If you do not provide this flag, the tool will only
  search within the top-level RST file or directory. If you do provide this flag, the tool will recursively scan all
  subdirectories for RST files and search across all files.
- `-f, --follow-includes` - Follow `.. include::` directives in RST files. If you do not provide this flag, the tool
  will search only the top-level RST file or directory. If you do provide this flag, the tool will follow any
  `.. include::` directives in any RST file in the input path and search across all included files. When
  combined with `-r`, the tool will recursively scan all subdirectories for RST files and follow `.. include::` directives
  in all files. If an include filepath is *outside* the input directory, the `-r` flag would not parse it, but the `-f`
  flag would follow the include directive and search the included file. This effectively lets you parse all the files
  that make up a single page, if you start from the page's root `.txt` file.
- `-v, --verbose` - Show file paths and language breakdown
- `--case-sensitive` - Make search case-sensitive (default: case-insensitive)
- `--partial-match` - Allow partial matches within words (default: exact word matching)

**Report:**

The search report shows:
- Number of files scanned
- Number of files containing the substring (each file counted once)

With `-v` flag, also shows:
- List of file paths where substring appears
- Count broken down by language (file extension)

### Analyze Commands

#### `analyze includes`

Analyze `include` directive relationships in RST files to understand file dependencies.

This command recursively follows `.. include::` directives to show all files that are referenced from a starting file. This helps you understand which content is transcluded into a page.

**Use Cases:**

This command helps writers:
- Understand the impact of changes to widely-included files
- Identify circular include dependencies (files included multiple times)
- Document file relationships for maintenance
- Plan refactoring of complex include structures
- See what content is actually pulled into a page

**Basic Usage:**

```bash
# Analyze a single file (shows summary)
./audit-cli analyze includes path/to/file.rst

# Show hierarchical tree structure
./audit-cli analyze includes path/to/file.rst --tree

# Show flat list of all included files
./audit-cli analyze includes path/to/file.rst --list

# Show both tree and list
./audit-cli analyze includes path/to/file.rst --tree --list

# Verbose output (show processing details)
./audit-cli analyze includes path/to/file.rst --tree -v
```

**Flags:**

- `--tree` - Display results as a hierarchical tree structure
- `--list` - Display results as a flat list of all files
- `-v, --verbose` - Show detailed processing information

**Output Formats:**

**Summary** (default - no flags):
- Root file path
- Total number of files
- Maximum depth of include nesting
- Hints to use --tree or --list for more details

**Tree** (--tree flag):
- Hierarchical tree structure showing include relationships
- Uses box-drawing characters for visual clarity
- Shows which files include which other files

**List** (--list flag):
- Flat numbered list of all files
- Files listed in depth-first traversal order
- Shows absolute paths to all files

**Note on File Counting:**

The total file count represents **unique files** discovered through include directives. If a file is included multiple
times (e.g., file A includes file C, and file B also includes file C), the file is counted only once in the total.
However, the tree view will show it in all locations where it appears, with subsequent occurrences marked as circular
includes in verbose mode.

**Note on Toctree:**

This command does **not** follow `.. toctree::` entries. Toctree entries are navigation links to other pages, not content
that's transcluded into the page. If you need to find which files reference a target file through toctree entries, use
the `analyze usage` command with the `--include-toctree` flag.

#### `analyze usage`

Find all files that use a target file through RST directives. This performs reverse dependency analysis, showing which files reference the target file through `include`, `literalinclude`, `io-code-block`, or `toctree` directives.

The command searches all RST files (`.rst` and `.txt` extensions) and YAML files (`.yaml` and `.yml` extensions) in the source directory tree. YAML files are included because extract and release files contain RST directives within their content blocks.

**Use Cases:**

By default, this command searches for content inclusion directives (include, literalinclude,
io-code-block) that transclude content into pages. Use `--include-toctree` to also search
for toctree entries, which are navigation links rather than content transclusion.

This command helps writers:
- Understand the impact of changes to a file (what pages will be affected)
- Find all usages of an include file across the documentation
- Track where code examples are referenced
- Plan refactoring by understanding file dependencies

**Basic Usage:**

```bash
# Find what uses an include file (content inclusion only)
./audit-cli analyze usage path/to/includes/fact.rst

# Find what uses a code example
./audit-cli analyze usage path/to/code-examples/example.js

# Include toctree references (navigation links)
./audit-cli analyze usage path/to/file.rst --include-toctree

# Get JSON output for automation
./audit-cli analyze usage path/to/file.rst --format json

# Show detailed information with line numbers
./audit-cli analyze usage path/to/file.rst --verbose
```

**Flags:**

- `--format <format>` - Output format: `text` (default) or `json`
- `-v, --verbose` - Show detailed information including line numbers and reference paths
- `-c, --count-only` - Only show the count of usages (useful for quick checks and scripting)
- `--paths-only` - Only show the file paths, one per line (useful for piping to other commands)
- `--summary` - Only show summary statistics (total files and usages by type, without file list)
- `-t, --directive-type <type>` - Filter by directive type: `include`, `literalinclude`, `io-code-block`, or `toctree`
- `--include-toctree` - Include toctree entries (navigation links) in addition to content inclusion directives
- `--exclude <pattern>` - Exclude paths matching this glob pattern (e.g., `*/archive/*` or `*/deprecated/*`)

**Understanding the Counts:**

The command shows two metrics:
- **Total Files**: Number of unique files that use the target (deduplicated)
- **Total Usages**: Total number of directive occurrences (includes duplicates)

When a file includes the target multiple times, it counts as:
- 1 file (in Total Files)
- Multiple usages (in Total Usages)

This helps identify both the impact scope (how many files) and duplicate includes (when usages > files).

**Supported Directive Types:**

By default, the command tracks content inclusion directives:

1. **`.. include::`** - RST content includes (transcluded)
   ```rst
   .. include:: /includes/intro.rst
   ```

2. **`.. literalinclude::`** - Code file references (transcluded)
   ```rst
   .. literalinclude:: /code-examples/example.py
      :language: python
   ```

3. **`.. io-code-block::`** - Input/output examples with file arguments (transcluded)
   ```rst
   .. io-code-block::

      .. input:: /code-examples/query.js
         :language: javascript

      .. output:: /code-examples/result.json
         :language: json
   ```

With `--include-toctree`, also tracks:

4. **`.. toctree::`** - Table of contents entries (navigation links, not transcluded)
   ```rst
   .. toctree::
      :maxdepth: 2

      intro
      getting-started
   ```

**Note:** Only file-based references are tracked. Inline content (e.g., `.. input::` with `:language:` but no file path) is not tracked since it doesn't reference external files.

**Output Formats:**

**Text** (default):
```
============================================================
USAGE ANALYSIS
============================================================
Target File: /path/to/includes/intro.rst
Total Files: 3
Total Usages: 4
============================================================

include             : 3 files, 4 usages

  1. [include] duplicate-include-test.rst (2 usages)
  2. [include] include-test.rst
  3. [include] page.rst

```

**Text with --verbose:**
```
============================================================
USAGE ANALYSIS
============================================================
Target File: /path/to/includes/intro.rst
Total Files: 3
Total Usages: 4
============================================================

include             : 3 files, 4 usages

  1. [include] duplicate-include-test.rst (2 usages)
     Line 6: /includes/intro.rst
     Line 13: /includes/intro.rst
  2. [include] include-test.rst
     Line 6: /includes/intro.rst
  3. [include] page.rst
     Line 12: /includes/intro.rst

```

**JSON** (--format json):
```json
{
  "target_file": "/path/to/includes/intro.rst",
  "source_dir": "/path/to/source",
  "total_files": 3,
  "total_usages": 4,
  "using_files": [
    {
      "file_path": "/path/to/duplicate-include-test.rst",
      "directive_type": "include",
      "usage_path": "/includes/intro.rst",
      "line_number": 6
    },
    {
      "file_path": "/path/to/duplicate-include-test.rst",
      "directive_type": "include",
      "usage_path": "/includes/intro.rst",
      "line_number": 13
    },
    {
      "file_path": "/path/to/include-test.rst",
      "directive_type": "include",
      "usage_path": "/includes/intro.rst",
      "line_number": 6
    }
  ]
}
```

**Examples:**

```bash
# Check if an include file is being used
./audit-cli analyze usage ~/docs/source/includes/fact-atlas.rst

# Find all pages that use a specific code example
./audit-cli analyze usage ~/docs/source/code-examples/connect.py

# Get machine-readable output for scripting
./audit-cli analyze usage ~/docs/source/includes/fact.rst --format json | jq '.total_usages'

# See exactly where a file is referenced (with line numbers)
./audit-cli analyze usage ~/docs/source/includes/intro.rst --verbose

# Quick check: just show the count
./audit-cli analyze usage ~/docs/source/includes/fact.rst --count-only
# Output: 5

# Show summary statistics only
./audit-cli analyze usage ~/docs/source/includes/fact.rst --summary
# Output:
# Total Files: 3
# Total Usages: 5
#
# By Type:
#   include             : 3 files, 5 usages

# Get list of files for piping to other commands
./audit-cli analyze usage ~/docs/source/includes/fact.rst --paths-only
# Output:
# page1.rst
# page2.rst
# page3.rst

# Filter to only show include directives (not literalinclude or io-code-block)
./audit-cli analyze usage ~/docs/source/includes/fact.rst --directive-type include

# Filter to only show literalinclude usages
./audit-cli analyze usage ~/docs/source/code-examples/example.py --directive-type literalinclude

# Combine filters: count only literalinclude usages
./audit-cli analyze usage ~/docs/source/code-examples/example.py -t literalinclude -c

# Combine filters: list files that use this as an io-code-block
./audit-cli analyze usage ~/docs/source/code-examples/query.js -t io-code-block --paths-only

# Exclude archived or deprecated files from search
./audit-cli analyze usage ~/docs/source/includes/fact.rst --exclude "*/archive/*"
./audit-cli analyze usage ~/docs/source/includes/fact.rst --exclude "*/deprecated/*"
```

### Compare Commands

#### `compare file-contents`

Compare file contents to identify differences between files. Supports two modes:
1. **Direct comparison** - Compare two specific files
2. **Version comparison** - Compare the same file across multiple documentation versions

**Use Cases:**

This command helps writers:
- Identify content drift across documentation versions
- Verify that updates have been applied consistently
- Scope maintenance work when updating shared content
- Understand how files have diverged over time

**Basic Usage:**

```bash
# Direct comparison of two files
./audit-cli compare file-contents file1.rst file2.rst

# Compare with diff output
./audit-cli compare file-contents file1.rst file2.rst --show-diff

# Version comparison across MongoDB documentation versions
./audit-cli compare file-contents \
  /path/to/manual/manual/source/includes/example.rst \
  --product-dir /path/to/manual \
  --versions manual,upcoming,v8.0,v7.0

# Show which files differ
./audit-cli compare file-contents \
  /path/to/manual/manual/source/includes/example.rst \
  --product-dir /path/to/manual \
  --versions manual,upcoming,v8.0,v7.0 \
  --show-paths

# Show detailed diffs
./audit-cli compare file-contents \
  /path/to/manual/manual/source/includes/example.rst \
  --product-dir /path/to/manual \
  --versions manual,upcoming,v8.0,v7.0 \
  --show-diff

# Verbose output (show processing details)
./audit-cli compare file-contents file1.rst file2.rst -v
```

**Flags:**

- `-p, --product-dir <dir>` - Product directory path (required for version comparison)
- `-V, --versions <list>` - Comma-separated list of versions (e.g., `manual,upcoming,v8.0`)
- `--show-paths` - Display file paths grouped by status (matching, differing, not found)
- `-d, --show-diff` - Display unified diff output (implies `--show-paths`)
- `-v, --verbose` - Show detailed processing information

**Comparison Modes:**

**1. Direct Comparison (Two Files)**

Provide two file paths as arguments:

```bash
./audit-cli compare file-contents path/to/file1.rst path/to/file2.rst
```

This mode:
- Compares exactly two files
- Reports whether they are identical or different
- Can show unified diff with `--show-diff`

**2. Version Comparison (Product Directory)**

Provide one file path plus `--product-dir` and `--versions`:

```bash
./audit-cli compare file-contents \
  /path/to/manual/manual/source/includes/example.rst \
  --product-dir /path/to/manual \
  --versions manual,upcoming,v8.0
```

This mode:
- Extracts the relative path from the reference file
- Resolves the same relative path in each version directory
- Compares all versions against the reference file
- Reports matching, differing, and missing files

**Version Directory Structure:**

The tool expects MongoDB documentation to be organized as:
```
product-dir/
├── manual/
│   └── source/
│       └── includes/
│           └── example.rst
├── upcoming/
│   └── source/
│       └── includes/
│           └── example.rst
└── v8.0/
    └── source/
        └── includes/
            └── example.rst
```

**Output Formats:**

**Summary** (default - no flags):
- Total number of versions compared
- Count of matching, differing, and missing files
- Hints to use `--show-paths` or `--show-diff` for more details

**With --show-paths:**
- Summary (as above)
- List of files that match (with ✓)
- List of files that differ (with ✗)
- List of files not found (with -)

**With --show-diff:**
- Summary and paths (as above)
- Unified diff output for each differing file
- Shows added lines (prefixed with +)
- Shows removed lines (prefixed with -)
- Shows context lines around changes

**Examples:**

```bash
# Check if a file is consistent across all versions
./audit-cli compare file-contents \
  ~/workspace/docs-mongodb-internal/content/manual/manual/source/includes/fact-atlas-search.rst \
  --product-dir ~/workspace/docs-mongodb-internal/content/manual \
  --versions manual,upcoming,v8.0,v7.0,v6.0

# Find differences and see what changed
./audit-cli compare file-contents \
  ~/workspace/docs-mongodb-internal/content/manual/manual/source/includes/fact-atlas-search.rst \
  --product-dir ~/workspace/docs-mongodb-internal/content/manual \
  --versions manual,upcoming,v8.0,v7.0,v6.0 \
  --show-diff

# Compare two specific versions of a file
./audit-cli compare file-contents \
  ~/workspace/docs-mongodb-internal/content/manual/manual/source/includes/example.rst \
  ~/workspace/docs-mongodb-internal/content/manual/v8.0/source/includes/example.rst \
  --show-diff
```

**Exit Codes:**

- `0` - Success (files compared successfully, regardless of whether they match)
- `1` - Error (invalid arguments, file not found, read error, etc.)

**Note on Missing Files:**

Files that don't exist in certain versions are reported separately and do not cause errors. This is expected behavior
since features may be added or removed across versions.

### Count Commands

#### `count tested-examples`

Count tested code examples in the MongoDB documentation monorepo.

This command navigates to the `content/code-examples/tested` directory from the monorepo root and counts all files recursively. The tested directory has a two-level structure: L1 (language directories) and L2 (product directories).

**Use Cases:**

This command helps writers and maintainers:
- Track the total number of tested code examples
- Monitor code example coverage by product
- Identify products with few or many examples
- Count only source files (excluding output files)

**Basic Usage:**

```bash
# Get total count of all tested code examples
./audit-cli count tested-examples /path/to/docs-monorepo

# Count examples for a specific product
./audit-cli count tested-examples /path/to/docs-monorepo --for-product pymongo

# Show counts broken down by product
./audit-cli count tested-examples /path/to/docs-monorepo --count-by-product

# Count only source files (exclude .txt and .sh output files)
./audit-cli count tested-examples /path/to/docs-monorepo --exclude-output
```

**Flags:**

- `--for-product <product>` - Only count code examples for a specific product
- `--count-by-product` - Display counts for each product
- `--exclude-output` - Only count source files (exclude .txt and .sh files)

**Current Valid Products:**

- `mongosh` - MongoDB Shell
- `csharp/driver` - C#/.NET Driver
- `go/driver` - Go Driver
- `go/atlas-sdk` - Atlas Go SDK
- `java/driver-sync` - Java Sync Driver
- `javascript/driver` - Node.js Driver
- `pymongo` - PyMongo Driver

**Output:**

By default, prints a single integer (total count) for use in CI or scripting. With `--count-by-product`, displays a formatted table with product names and counts.

#### `count pages`

Count documentation pages (.txt files) in the MongoDB documentation monorepo.

This command navigates to the `content` directory and recursively counts all `.txt` files, which represent documentation pages that resolve to unique URLs. The command automatically excludes certain directories and file types that don't represent actual documentation pages.

**Use Cases:**

This command helps writers and maintainers:
- Track the total number of documentation pages across the monorepo
- Monitor documentation coverage by product/project
- Identify projects with extensive or minimal documentation
- Exclude auto-generated or deprecated content from counts
- Count only current versions of versioned documentation
- Compare page counts across different documentation versions

**Automatic Exclusions:**

The command automatically excludes:
- Files in `code-examples` directories at the root of `content` or `source` (these contain plain text examples, not pages)
- Files in the following directories at the root of `content`:
  - `404` - Error pages
  - `docs-platform` - Documentation for the MongoDB website and meta content
  - `meta` - MongoDB Meta Documentation - style guide, tools, etc.
  - `table-of-contents` - Navigation files
- All non-`.txt` files (configuration files, YAML, etc.)

**Basic Usage:**

```bash
# Get total count of all documentation pages
./audit-cli count pages /path/to/docs-monorepo

# Count pages for a specific project
./audit-cli count pages /path/to/docs-monorepo --for-project manual

# Show counts broken down by project
./audit-cli count pages /path/to/docs-monorepo --count-by-project

# Exclude specific directories from counting
./audit-cli count pages /path/to/docs-monorepo --exclude-dirs api-reference,generated

# Count only current versions (for versioned projects)
./audit-cli count pages /path/to/docs-monorepo --current-only

# Show counts by project and version
./audit-cli count pages /path/to/docs-monorepo --by-version

# Combine flags: count pages for a specific project, excluding certain directories
./audit-cli count pages /path/to/docs-monorepo --for-project atlas --exclude-dirs deprecated
```

**Flags:**

- `--for-project <project>` - Only count pages for a specific project (directory name under `content/`)
- `--count-by-project` - Display counts for each project in a formatted table
- `--exclude-dirs <dirs>` - Comma-separated list of directory names to exclude from counting (e.g., `deprecated,archive`)
- `--current-only` - Only count pages in the current version (for versioned projects, counts only `current` or `manual` version directories; for non-versioned projects, counts all pages)
- `--by-version` - Display counts grouped by project and version (shows version breakdown for versioned projects; non-versioned projects show as "(no version)")

**Output:**

By default, prints a single integer (total count) for use in CI or scripting. With `--count-by-project`, displays a formatted table with project names and counts. With `--by-version`, displays a hierarchical breakdown by project and version.

**Versioned Documentation:**

Some MongoDB documentation projects contain multiple versions, represented as distinct directories between the project directory and the `source` directory:
- **Versioned project structure**: `content/{project}/{version}/source/...`
- **Non-versioned project structure**: `content/{project}/source/...`

Version directory names follow these patterns:
- `current` or `manual` - The current/latest version
- `upcoming` - Pre-release version
- `v{number}` - Specific version (e.g., `v8.0`, `v7.0`)

The `--current-only` flag counts only files in the current version directory (`current` or `manual`) for versioned projects, while counting all files for non-versioned projects.

The `--by-version` flag shows a breakdown of page counts for each version within each project.

**Note:** The `--current-only` and `--by-version` flags are mutually exclusive.

**Examples:**

```bash
# Quick count for CI/CD
TOTAL_PAGES=$(./audit-cli count pages ~/docs-monorepo)
echo "Total documentation pages: $TOTAL_PAGES"

# Detailed breakdown by project
./audit-cli count pages ~/docs-monorepo --count-by-project
# Output:
# Page Counts by Project:
#
#   app-services                       245
#   atlas                              512
#   manual                            1024
#   ...
#
# Total: 2891

# Count only Atlas pages
./audit-cli count pages ~/docs-monorepo --for-project atlas
# Output: 512

# Exclude deprecated content
./audit-cli count pages ~/docs-monorepo --exclude-dirs deprecated,archive --count-by-project

# Count only current versions
./audit-cli count pages ~/docs-monorepo --current-only
# Output: 1245 (only counts current/manual versions)

# Show breakdown by version
./audit-cli count pages ~/docs-monorepo --by-version
# Output:
# Project: drivers
#   manual                           150
#   upcoming                         145
#   v8.0                             140
#   v7.0                             135
#
# Project: atlas
#   (no version)                     200
#
# Total: 770

# Count current version for a specific project
./audit-cli count pages ~/docs-monorepo --for-project drivers --current-only
# Output: 150
```

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
│   ├── search/                      # Search parent command
│   │   ├── search.go               # Parent command definition
│   │   └── find-string/            # Find string subcommand
│   │       ├── find_string.go      # Command logic
│   │       ├── types.go            # Type definitions
│   │       └── report.go           # Report generation
│   ├── analyze/                     # Analyze parent command
│   │   ├── analyze.go              # Parent command definition
│   │   ├── includes/               # Includes analysis subcommand
│   │   │   ├── includes.go         # Command logic
│   │   │   ├── analyzer.go         # Include tree building
│   │   │   ├── output.go           # Output formatting
│   │   │   └── types.go            # Type definitions
│   │   └── usage/                  # Usage analysis subcommand
│   │       ├── usage.go            # Command logic
│   │       ├── usage_test.go       # Tests
│   │       ├── analyzer.go         # Reference finding logic
│   │       ├── output.go           # Output formatting
│   │       └── types.go            # Type definitions
│   ├── compare/                     # Compare parent command
│   │   ├── compare.go              # Parent command definition
│   │   └── file-contents/          # File contents comparison subcommand
│   │       ├── file_contents.go    # Command logic
│   │       ├── file_contents_test.go # Tests
│   │       ├── comparer.go         # Comparison logic
│   │       ├── differ.go           # Diff generation
│   │       ├── output.go           # Output formatting
│   │       ├── types.go            # Type definitions
│   │       └── version_resolver.go # Version path resolution
│   └── count/                       # Count parent command
│       ├── count.go                # Parent command definition
│       ├── tested-examples/        # Tested examples counting subcommand
│       │   ├── tested_examples.go  # Command logic
│       │   ├── tested_examples_test.go # Tests
│       │   ├── counter.go          # Counting logic
│       │   ├── output.go           # Output formatting
│       │   └── types.go            # Type definitions
│       └── pages/                  # Pages counting subcommand
│           ├── pages.go            # Command logic
│           ├── pages_test.go       # Tests
│           ├── counter.go          # Counting logic
│           ├── output.go           # Output formatting
│           └── types.go            # Type definitions
├── internal/                        # Internal packages
│   ├── pathresolver/               # Path resolution utilities
│   │   ├── pathresolver.go         # Core path resolution
│   │   ├── pathresolver_test.go    # Tests
│   │   ├── source_finder.go        # Source directory detection
│   │   ├── version_resolver.go     # Version path resolution
│   │   └── types.go                # Type definitions
│   └── rst/                        # RST parsing utilities
│       ├── parser.go               # Generic parsing with includes
│       ├── include_resolver.go     # Include directive resolution
│       ├── directive_parser.go     # Directive parsing
│       └── file_utils.go           # File utilities
└── testdata/                        # Test fixtures
    ├── input-files/                # Test RST files
    │   └── source/                 # Source directory (required)
    │       ├── *.rst               # Test files
    │       ├── includes/           # Included RST files
    │       └── code-examples/      # Code files for literalinclude
    ├── expected-output/            # Expected extraction results
    ├── compare/                    # Compare command test data
    │   ├── product/                # Version structure tests
    │   │   ├── manual/             # Manual version
    │   │   ├── upcoming/           # Upcoming version
    │   │   └── v8.0/               # v8.0 version
    │   └── *.txt                   # Direct comparison tests
    └── count-test-monorepo/        # Count command test data
        └── content/code-examples/tested/  # Tested examples structure
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

**Note**: The `testdata` directory name is special in Go - it's automatically ignored during builds, which is important
since it contains non-Go files (`.cpp`, `.rst`, etc.).

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

- **Relative paths**: Tests use `filepath.Join("..", "..", "..", "testdata")` to reference test data (three levels up
  from `commands/extract/code-examples/`)
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

### Code Example Extraction

The tool extracts code examples from the following reStructuredText directives:

#### 1. `literalinclude`

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

#### 2. `code-block`

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

The code has 6 spaces of indentation (3 from `note`, 3 from `code-block`). The tool automatically removes these 6 spaces,
resulting in:

```python
def hello():
    print("Hello")
```

#### 3. `io-code-block`

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

### Include handling

#### 4. `include`

Follows include directives to process entire documentation trees (when `-f` flag is used).

**Syntax:**
```rst
.. include:: /includes/intro.rst
```

**Special MongoDB Conventions:**

The tool handles several MongoDB-specific include patterns:

##### Steps Files
Converts directory-based paths to filename-based paths:
- Input: `/includes/steps/run-mongodb-on-linux.rst`
- Resolves to: `/includes/steps-run-mongodb-on-linux.yaml`

##### Extracts and Release Files
Resolves ref-based includes by searching YAML files:
- Input: `/includes/extracts/install-mongodb.rst`
- Searches: `/includes/extracts-*.yaml` for `ref: install-mongodb`
- Resolves to: The YAML file containing that ref

##### Template Variables
Resolves template variables from YAML replacement sections:
```yaml
replacement:
  release_specification_default: "/includes/release/install-windows-default.rst"
```
- Input: `{{release_specification_default}}`
- Resolves to: `/includes/release/install-windows-default.rst`

**Source Directory Resolution:**

The tool walks up the directory tree to find a directory named "source" or containing a "source" subdirectory. This is
used as the base for resolving relative include paths.

## Internal Packages

### `internal/pathresolver`

Provides centralized path resolution utilities for working with MongoDB documentation structure:

- **Source directory detection** - Finds the documentation root by walking up the directory tree
- **Project info detection** - Identifies product directory, version, and whether a project is versioned
- **Version path resolution** - Resolves file paths across multiple documentation versions
- **Relative path resolution** - Resolves paths relative to the source directory

**Key Functions:**
- `FindSourceDirectory(filePath string)` - Finds the source directory for a given file
- `DetectProjectInfo(filePath string)` - Detects project structure information
- `ResolveVersionPaths(referenceFile, productDir string, versions []string)` - Resolves paths across versions
- `ResolveRelativeToSource(sourceDir, relativePath string)` - Resolves relative paths

See the code in `internal/pathresolver/` for implementation details.

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

| Input          | Normalized   | Extension |
|----------------|--------------|-----------|
| `bash`         | `bash`       | `.sh`     |
| `c`            | `c`          | `.c`      |
| `c++`          | `cpp`        | `.cpp`    |
| `c#`           | `csharp`     | `.cs`     |
| `console`      | `console`    | `.sh`     |
| `cpp`          | `cpp`        | `.cpp`    |
| `cs`           | `csharp`     | `.cs`     |
| `csharp`       | `csharp`     | `.cs`     |
| `go`           | `go`         | `.go`     |
| `golang`       | `go`         | `.go`     |
| `java`         | `java`       | `.java`   |
| `javascript`   | `javascript` | `.js`     |
| `js`           | `javascript` | `.js`     |
| `kotlin`       | `kotlin`     | `.kt`     |
| `kt`           | `kotlin`     | `.kt`     |
| `php`          | `php`        | `.php`    |
| `powershell`   | `powershell` | `.ps1`    |
| `ps1`          | `powershell` | `.ps1`    |
| `ps5`          | `ps5`        | `.ps1`    |
| `py`           | `python`     | `.py`     |
| `python`       | `python`     | `.py`     |
| `rb`           | `ruby`       | `.rb`     |
| `rs`           | `rust`       | `.rs`     |
| `ruby`         | `ruby`       | `.rb`     |
| `rust`         | `rust`       | `.rs`     |
| `scala`        | `scala`      | `.scala`  |
| `sh`           | `shell`      | `.sh`     |
| `shell`        | `shell`      | `.sh`     |
| `swift`        | `swift`      | `.swift`  |
| `text`         | `text`       | `.txt`    |
| `ts`           | `typescript` | `.ts`     |
| `txt`          | `text`       | `.txt`    |
| `typescript`   | `typescript` | `.ts`     |
| (empty string) | `undefined`  | `.txt`    |
| `none`         | `undefined`  | `.txt`    |
| (unknown)      | (unchanged)  | `.txt`    |

**Notes:**
- Language identifiers are case-insensitive
- Unknown languages are returned unchanged by `NormalizeLanguage()` but map to `.txt` extension
- The normalization handles common aliases (e.g., `ts` → `typescript`, `golang` → `go`, `c++` → `cpp`)

## Contributing

When contributing to this project:

1. **Follow the established patterns** - Use the command structure, error handling, and testing patterns described above
2. **Write tests** - All new functionality should have corresponding tests
3. **Update documentation** - Keep this README up to date with new features
4. **Run tests before committing** - Ensure `go test ./...` passes
5. **Use meaningful commit messages** - Describe what changed and why
