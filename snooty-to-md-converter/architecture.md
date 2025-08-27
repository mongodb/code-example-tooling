# Architecture Overview

This repository provides a toolchain to convert Snooty-based documentation projects to Markdown, with an emphasis on migrating the Realm docs. It is organized as a small npm workspace with a single package that exposes a CLI and a set of conversion utilities.

- Workspace root: scripts, top-level README, licensing, and npm workspace wiring
- Package: packages/realm-docs-converter – the actual converter (TypeScript)
  - CLI entrypoint: src/cli.ts
  - Core conversion orchestrator: src/realm-docs-converter.ts
  - Snooty Data API client: src/snooty-api.ts
  - Snooty AST → Markdown renderer: src/ast-to-md.ts
  - Fallback local RST utilities: src/converters/snooty.ts

## High-level Flow

There are two primary modes of operation: API mode (recommended) and Local mode (fallback). Both end with Markdown files written to an output directory, preserving the original page path structure.

1. API Mode (default)
   - CLI parses args and calls convertRealmDocsFromApi().
   - Snooty Data API is queried to obtain a list of pages and each page’s JSON AST.
   - The AST is converted to Markdown via astToMarkdown().
   - Shared images may be copied to <out>/images if provided locally.
   - Image references beginning with /images/... are rewritten to be relative to each output file.
   - Per-page warnings (e.g., unresolved substitutions/refs) are collected and written to conversion-warnings.log.

2. Local Mode (legacy/fallback)
   - CLI parses args --local <input-dir> --out <output-dir> and calls convertRealmDocs().
   - Local RST-like files (.txt/.rst) are recursively discovered under the input directory.
   - Each file is read as text and processed by parseSnootyContent() to (best-effort) resolve includes, substitutions, and refs.
   - The processed text is converted to Markdown by convertToMarkdown().
   - Shared images are copied from <input>/images to <out>/images when present, and image references are rewritten similarly to API mode.

## Repository Layout

- package.json (root)
  - Declares an npm workspace with packages/realm-docs-converter
  - Scripts delegate to the package (build/start/dev)
- packages/realm-docs-converter/package.json
  - name: realm-docs-converter
  - bin: dist/cli.js (CLI entry)
  - scripts: build via tsc
  - engines: Node >= 18
  - deps: dotenv (for env), fs-extra (not heavily used; core logic mostly uses fs)
- packages/realm-docs-converter/tsconfig.json
  - OutDir: dist, RootDir: src; CommonJS target ES2020

## Key Components

1. CLI (src/cli.ts)
   - Parses arguments:
     - Default mode: API
     - --project <slug> (default: realm)
     - --out <output-dir> (required in API mode)
     - --branch <branch> (default: master)
     - --base-url <url> (optional override for Snooty Data API)
     - --local (switch to Local mode) with positional <input-dir> and --out <dir>
   - Calls:
     - convertRealmDocsFromApi({ project, outputDir, branch, baseUrl }) for API
     - convertRealmDocs({ inputDir, outputDir, handleIncludes: true, handleSubstitutions: true, handleRefs: true }) for Local
   - Logs total pages converted and basic usage help if args are missing.

2. Conversion Orchestrator (src/realm-docs-converter.ts)
   - convertRealmDocs(options: { inputDir, outputDir, handleIncludes, handleSubstitutions, handleRefs })
     - Removes and recreates the output directory on each run.
     - Copies shared images from <input>/images to <out>/images when present.
     - Recursively lists .txt/.rst files.
     - For each file:
       - Reads content, parses with parseSnootyContent() (includes/substitutions/refs), converts with convertToMarkdown().
       - Ensures the output directory exists, rewrites image paths, and writes a .md file mirroring the folder structure.
   - convertRealmDocsFromApi(options: { project, outputDir, branch?, baseUrl? })
     - Removes and recreates the output directory on each run.
     - Attempts to copy shared images into <out>/images from:
       - SHARED_IMAGES_DIR (if set and exists), or
       - ./images next to the current working directory (fallback)
     - Fetches pages and their ASTs via fetchSnootyProject().
     - Maintains a global substitutions map shared across pages.
     - Converts each page’s AST to Markdown via astToMarkdown(), collects warnings with doc paths and writes them to conversion-warnings.log.
     - Writes output .md files, normalizing paths and handling deletions indicated by the API.
   - rewriteImagePaths(markdown, outPath, outputRoot)
     - If <out>/images exists, rewrites:
       - Markdown image URLs: ![alt](/images/path.png)
       - HTML <img src="/images/...">
       to be relative to the output file location.
     - If images directory does not exist, leaves absolute /images/... references as-is.

3. Snooty Data API Client (src/snooty-api.ts)
   - fetchSnootyProject({ project, branch = 'master', baseUrl = 'https://snooty-data-api.mongodb.com' })
     - Tries multiple endpoints to obtain a pages index, then per-page ASTs.
     - Normalizes results to an array of { path, ast } (or { path, deleted: true }).
     - Handles a variety of index shapes (strings, objects, embedded asts) and skips non-page asset entries.
     - Uses a liberal tryFetchAny() that supports JSON and NDJSON-like responses.

4. AST → Markdown (src/ast-to-md.ts)
   - astToMarkdown(root, { substitutions, onWarn, docPath })
     - Walks the Snooty AST and emits Markdown lines.
     - Supported constructs (best-effort):
       - Sections/titles → #..###### (adds anchors when ids/html_id are present internally)
       - Paragraphs, inline formatting (emphasis, strong, inline code)
       - Links: external via refuri; internal refs rendered as text labels
       - Images → Markdown syntax
       - Code/literal blocks (with language if provided)
       - Lists (bulleted/ordered)
       - Tables → GFM pipe tables
       - Admonitions (note, tip, warning, etc.) → blockquotes with a label
       - Substitution references via substitutions map; unresolved ones trigger onWarn
     - Emits warnings for unresolved substitutions/refs/includes with page path and optional position.

5. Local RST Fallback (src/converters/snooty.ts)
   - parseSnootyContent(text, { filePath, basePath, resolveIncludes, resolveSubstitutions, resolveRefs })
     - Best-effort handling of:
       - .. include:: directives (project-root relative when starting with /)
       - Substitution definitions and usages (.. |name| replace:: value, and |name|)
       - :ref:`text <label>` and :ref:`label` inlining
     - Produces { content, includes, substitutions, refs }.
   - convertToMarkdown(parsed)
     - Pragmatic RST-to-Markdown transforms: headers, links, code blocks, and lists.

## Data and Control Flow (API Mode)

CLI → convertRealmDocsFromApi → fetchSnootyProject → pages[]
for each page:
- Normalize output path (ensure .md extension; handle deletions)
- astToMarkdown(page.ast, { substitutions, onWarn, docPath }) → md
- rewriteImagePaths(md, outPath, outputDir) → md'
- Write md' to file
- Collect warnings → conversion-warnings.log

## Configuration and Environment

- CLI flags:
  - --project, --out, --branch, --base-url, --local
- Environment variables:
  - SHARED_IMAGES_DIR: Directory to copy into <out>/images for API mode. If unset, ./images (cwd) is used as a fallback when present.
- Node version: >= 18 (for global fetch and modern fs APIs)

## Error Handling and Logging

- Non-fatal issues (unresolved substitutions, unrecognized nodes, missing images dir) generate warnings collected per document or printed during local conversion.
- Fatal issues (e.g., HTTP failures, no index endpoint) surface as errors and abort the run.
- API client tries multiple endpoint shapes and NDJSON parsing to be resilient to deployment differences.

## Build and Distribution

- TypeScript compilation via tsc produces dist/ with CommonJS modules.
- CLI bin is wired to dist/cli.js via package.json.
- Root workspace scripts proxy to the package scripts:
  - npm run build → build the package
  - npm start → run the CLI (after build)

## Extensibility Notes

- Adding support for more Snooty nodes:
  - Extend ast-to-md.ts with new cases in inline() and block() helpers.
  - Use onWarn to detect new node types before implementing them.
- Enhancing link/ref resolution:
  - Introduce a cross-page index of anchors and rewrite internal links accordingly.
- Improving local mode:
  - Replace the simplistic RST parser with a real parser, or drop local mode if API coverage is sufficient.
