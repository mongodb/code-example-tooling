# snooty-to-md-docs-converter

A converter for Snooty-based docs projects to Markdown. This is intended to help migrate the Realm docs, specifically, 
from Snooty to Markdown.
This tool ingests the Snooty Data API for a single project (default: `realm`) and outputs Markdown files, 
handling links, refs, substitutions, and includes (best-effort), logging any issues with pointers to the source page.

## Prerequisites
- Node.js >= 18 (for built-in fetch)
- npm >= 8
- To convert images referenced with absolute paths in API mode, either:
     - create a `.env` with `SHARED_IMAGES_DIR` set to a local folder containing images.
     - have a local `./images` directory where you run the CLI.

## Install dependencies (from repo root)
```
npm ci
```

## Build the converter (from repo root)
```
npm run build
```
This compiles TypeScript to `packages/realm-docs-converter/dist` and wires up the CLI bin (`dist/cli.js`).

## Run the converter (Snooty Data API)
Default mode fetches pages from the Snooty Data API and converts them to Markdown.
```
node packages/realm-docs-converter/dist/cli.js --project realm --out ./output --branch master --base-url https://snooty-data-api.mongodb.com
```
- `--project`: Snooty project slug (default: `realm`).
- `--out`: Output directory (required).
- `--branch`: Branch to fetch (default: `master`).
- `--base-url`: Snooty Data API base URL (defaults to https://snooty-data-api.mongodb.com).

The converter writes one `.md` file per page, mirroring the page path. A `conversion-warnings.log` file is written 
in the output directory if any unresolved includes/substitutions/refs occur.

### Local directory fallback (optional)
For local conversion of a checked-out Snooty project directory containing `.txt/.rst` files, you can use:
```
node packages/realm-docs-converter/dist/cli.js --local <input-dir> --out <output-dir>
```

## Shared images handling
- Many pages reference shared images with absolute paths like `/images/foo.png` (via `:figure:` directives in the source).
- The converter will copy shared images and rewrite references so the Markdown works offline:
  - Local mode: if `<input>/images` exists, it is copied to `<output>/images`.
  - API mode: if you set `REALM_DOCS_SHARED_IMAGES_DIR` to a local folder containing images, it is copied to `<output>/images`. As a fallback, if a local `./images` directory exists where you run the CLI, that will be copied.
  - When `<output>/images` exists, the converter rewrites both Markdown image links like `![alt](/images/path.png)` and HTML `<img src="/images/...">` to use relative paths from each page.
  - If `<output>/images` does not exist, absolute `/images/...` links are left as-is.

## Project overview: Snooty AST → Markdown specifics

This package prioritizes conversion from the Snooty Data API’s JSON AST into Markdown. 
Local RST parsing exists only as a fallback for legacy flows. 

The main data flow is:

1. Fetch pages and ASTs
   - `src/snooty-api.ts` fetches a normalized list of pages for a project/branch via the Snooty Data API, then resolves each page’s AST.
   - Results are normalized to an array of `{ path, ast }`.
2. Convert AST to Markdown
   - `src/ast-to-md.ts` walks the Snooty AST and renders Markdown.
   - A global substitutions map is shared across pages to ensure consistent replacement of substitution_reference nodes.
   - Unhandled or ambiguous nodes trigger warnings with page path and (if available) source position.
3. Write files and warnings
   - Each page becomes a `.md` file under the provided `--out` directory, preserving its relative path (adding `.md` if missing).
   - Any warnings are aggregated to `conversion-warnings.log` in the output directory.

### AST node mapping (what we currently support)

- section/title → Markdown headers (adds HTML anchor if ids/html_id present)
  - Nested sections increase header depth (# to ######). Depth is clamped to 1–6.
- paragraph → a Markdown paragraph (blank line after each)
- inline/literal → backticked inline code: `text`
  - literal, inline_literal nodes are rendered as `code`.
- literal_block → fenced code block
  - Uses ```language if node.language is present; plain ``` otherwise.
- bullet_list → - list items
- enumerated_list → 1. 2. … ordered list items
- reference → links and refs
  - If node.refuri exists, render [label](refuri).
  - If node.refname (internal ref) exists, render the visible label only (best-effort; no cross-file anchor weaving yet).
- image → Markdown images: ![alt](url) when a URL is present
- table → GFM pipe tables (best-effort conversion of header/body rows)
- admonitions (note, tip, warning, important, caution, seealso, or generic admonition with admonition_type) → blockquote with a label
- substitution_definition → populate the global substitutions map
  - Subsequent substitution_reference nodes can resolve using this map.
- substitution_reference → resolve to text
  - If not found in the map, we emit a warning and keep the original |name| syntax as a placeholder.
- include → warn and render children best-effort
  - Snooty Data API often expands includes before emitting AST. If an include node is still present, we log a warning and render any children.

All other nodes
- If a node type is not recognized, the converter walks its children and renders what it can (paragraphs, inlines inside). 
  This allows incremental coverage without failing the entire document.

### Limitations and non-goals

- Tabs, complex directives, advanced roles, and tables are not rendered specially; tables are converted best-effort to GFM.
- Internal anchors/refs are rendered as text; we do not resolve cross-page anchors.
- Frontmatter/metadata and explicit table-of-contents files are not generated.
- We rely on the Data API to resolve most includes; remaining include nodes are rendered best-effort with a warning.

### Programmatic usage

If you already have a Snooty AST and want to get Markdown directly:

```ts
import { astToMarkdown } from 'realm-docs-converter/dist/ast-to-md';

const substitutions: Record<string, string> = {};
const md = astToMarkdown(astJson, {
  substitutions,
  onWarn: (message, ctx) => {
    console.warn(ctx?.path ? `${ctx.path}: ${message}` : message);
  },
  docPath: 'path/to/doc',
});
```

### CLI examples

- Convert via Data API (recommended):
```
node packages/realm-docs-converter/dist/cli.js --project realm --out ./output --branch master
```
- Convert a local directory:
```
node packages/realm-docs-converter/dist/cli.js --local ./path/to/snooty --out ./output
```

## Notes
- The AST-to-Markdown converter implements a pragmatic subset (sections/titles to headers, paragraphs, inline/literal/code blocks,
  lists, links, refs, substitutions). It accumulates substitutions across pages and warns on unresolved items.
- Includes are usually expanded by the Data API; if include nodes remain, they are logged.
