# Log Parser Scripts

This directory contains a script to parse GDCD log files and analyze page changes, specifically identifying moved pages vs truly new/removed pages and tracking applied usage examples.

## Files

- `parse-log.go` - Main Go script that performs the log parsing and analysis
- `README.md` - This documentation file

## Purpose

The script analyzes log files to distinguish between:

1. **Moved Pages**: Pages that appear to be removed and created but are actually the same page moved to a new location within the same project
2. **Maybe New Pages**: Pages that may be genuinely new additions
3. **Maybe Removed Pages**: Pages that may be genuinely removed (not moved)
4. **Applied Usage Examples**: New applied usage examples on maybe new pages only

All results are reported with **project context** to clearly show which project each page belongs to.

## Dependencies

- Go

## How It Works

### Page Movement Detection

A page is considered "moved" if **all three conditions** are met:

1. **Same Project**: The removed page and created page are in the same project
2. **Same Code Example Count**: The removed page and created page have the same number of code examples
3. **Shared Segment**: At least one segment of the page ID (separated by `|`) is the same between the removed and created pages

For example:
- In project `ruby-driver`: `connect|tls` (removed, 6 code examples) → `security|tls` (created, 6 code examples)
- Same project AND same code examples AND shared segment `tls` → **MOVED**

### Applied Usage Examples Filtering

Applied usage examples are only counted for truly new pages, not for moved pages. This prevents double-counting when pages are reorganized.

### Maybe New and Maybe Removed Pages

Some conditions may cause moved pages to not meet our criteria for "moved" pages:

- Different number of code examples
  - Example: `connect|tls` is a "maybe removed" page and `security|tls` is a "maybe new" page but the removed page has
    6 code examples and the created page has 7 code examples
- No shared segments in page IDs
  - Example: `crud|update` is a "maybe removed" page and `write|upsert` is a "maybe new" page. Even if they have the same
    number of code examples, they share no segments in their page IDs so we can't programmatically detect that they're
    the same

Because of these conditions, we can only say that a page is "maybe new" or "maybe removed" and not "moved". A human must
manually review the "maybe new" and "maybe removed" results to determine if the page is truly new or removed. If it's
moved, we must manually adjust the count of new applied usage examples to omit the applied usage examples from the
"maybe new" but actually moved page.

## Usage

**Important**: You must be in the scripts directory to run the Go script directly:

```bash
# Navigate to the scripts directory first
cd /Your/Local/Filepath/code-example-tooling/audit/gdcd/scripts

# Then run the Go script
go run parse-log.go ../logs/2025-09-24-18-01-30-app.log
go run parse-log.go /absolute/path/to/your/log/file.log
```

## Output Format

The script produces four sections:

### 1. MOVED PAGES
```
=== MOVED PAGES ===
MOVED [ruby-driver]: connect|tls -> security|tls (6 code examples)
MOVED [ruby-driver]: write|bulk-write -> crud|bulk-write (9 code examples)
MOVED [database-tools]: installation|verify -> verify (0 code examples)
```

### 2. MAYBE NEW PAGES
```
=== MAYBE NEW PAGES ===
NEW [ruby-driver]: atlas-search (2 code examples)
NEW [node]: integrations|prisma (4 code examples)
NEW [atlas-architecture]: solutions-library|rag-technical-documents (6 code examples)
```

### 3. MAYBE REMOVED PAGES
```
=== MAYBE REMOVED PAGES ===
REMOVED [ruby-driver]: common-errors (4 code examples)
REMOVED [cpp-driver]: indexes|work-with-indexes (4 code examples)
REMOVED [docs]: tutorial|install-mongodb-on-windows-unattended (11 code examples)
```

### 4. NEW APPLIED USAGE EXAMPLES
```
=== NEW APPLIED USAGE EXAMPLES ===
APPLIED USAGE [ruby-driver]: atlas-search (1 applied usage examples)
APPLIED USAGE [node]: integrations|prisma (1 applied usage examples)
APPLIED USAGE [pymongo]: data-formats|custom-types|type-codecs (1 applied usage examples)

Total new applied usage examples: 17
```

## Log Format Requirements

The scripts expect log lines in the following formats:

- Project context: `Project changes for <project-name>`
- Page events: `Page removed: Page ID: <page-id>` or `Page created: Page ID: <page-id>`
- Code examples: `Code example removed: Page ID: <page-id>, <count> code examples removed`
- Applied usage: `Applied usage example added: Page ID: <page-id>, <count> new applied usage examples added`

**Important**: The script tracks the current project context from "Project changes for" lines and associates all subsequent page events with that project until a new project context is encountered.
