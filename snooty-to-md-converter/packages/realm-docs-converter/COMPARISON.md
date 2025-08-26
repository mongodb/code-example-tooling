Comparison: realm-docs-converter vs references

Scope and purpose
- realm-docs-converter (this package): Focused, minimal Snooty Data API → Markdown converter for Realm docs. Emphasizes robustness across varying API shapes, simple AST→MD mapping, and logging unresolved items. Has an optional local RST fallback.
- references/snooty-ast-to-mdx: A more feature-rich Snooty AST → MDX pipeline built around converting to mdast and then MDX, supporting includes as separate files, references aggregation, image/static asset handling, and MDX/remark ecosystem features.
- references/ingest-mongodb-docs: A docs ingestion toolkit with Snooty parsers/converters (Snooty → MD, RST → Snooty AST), table rendering, directive handling, metadata extraction, and an ingest pipeline.

Inputs and fetching
- realm-docs-converter: Fetches pages via Snooty Data API with multiple endpoint fallbacks; normalizes to [{ path, ast }]. Optional local directory parse fallback for .rst/.txt.
- snooty-ast-to-mdx: Accepts AST JSON directly, or BSON inside a .zip (convertZipFileToMdx). Not focused on live API fetching; more on converting pre-exported artifacts.
- ingest-mongodb-docs: Provides data sources for Snooty JSONL, projects info, and utility to build Snooty AST from RST; geared to ingest pipelines and tests.

Output format
- realm-docs-converter: Plain Markdown (.md). No frontmatter or MDX components.
- snooty-ast-to-mdx: MDX (.mdx) via mdast + remark (remark-frontmatter, remark-gfm, remark-mdx). Supports component imports for advanced nodes.
- ingest-mongodb-docs: Markdown strings used for ingestion; not targeting MDX/remark.

AST conversion strategy
- realm-docs-converter: Direct Snooty AST → Markdown via a small, pragmatic node mapping: sections/titles → headers; paragraphs; inline/inline_literal → backticks; literal_block → fenced code; lists; reference (refuri/refname); substitution_definition/reference; include (warning only); walk-children fallback for unknown nodes.
- snooty-ast-to-mdx: Snooty AST → mdast → MDX. Rich handling in convertSnootyAstToMdast including:
  - Includes: emits separate MDX files for included content through onEmitMDXFile, dedupes repeated includes, and calculates relative import paths.
  - Substitutions/refs: collects across pages and emits/merges a references.ts artifact at the output root for reuse.
  - Images and static assets: parses static_assets mapping, then extracts binary assets from zips and writes them to semantic paths. Handles image directive conversion.
  - Headings, paragraphs, inlines, lists, code, admonitions/directives, ids/anchors. Wraps inline runs and preserves structure via mdast.
  - MDX emission uses remark stack (frontmatter, GFM, MDX) for compatibility with MDX renderers.
- ingest-mongodb-docs: snootyAstToMd adds support around:
  - Tables: renderSnootyTable to Markdown.
  - Directives (subset), references/links, lists, code, headings.
  - Metadata/title extraction functions to pull facets/meta for ingestion.

Includes
- realm-docs-converter: Does not expand includes in AST mode (logs a warning). In local RST fallback, has a simple include replacement based on file paths.
- snooty-ast-to-mdx: Emits include content as separate MDX files and ensures idempotent file emission; also aggregates references from includes.
- ingest-mongodb-docs: Not focused on emitting separate files, but includes handling exists for rendering final Markdown.

Substitutions and refs
- realm-docs-converter: Global substitutions map shared across pages; unresolved substitution_reference triggers a warning and leaves |name| placeholder. Internal ref (refname) rendered as visible text only.
- snooty-ast-to-mdx: Aggregates substitutions and refs across main file and includes; emits/updates a references.ts containing substitutions and refs; more structured cross-file handling. MDX components/imports can be registered.
- ingest-mongodb-docs: Tracks links and constructs link maps for internal/external linking; renders accordingly.

Assets (images, binaries)
- realm-docs-converter: No binary/static asset handling in API mode. Images/directives currently treated as generic/unknown nodes and passed through as text.
- snooty-ast-to-mdx: Extracts static assets from zip exports using checksum→key mapping and writes them to the output tree; supports image directive conversion with importable paths.
- ingest-mongodb-docs: Focused on text ingestion; static asset handling is not a first-class concern.

Tables
- realm-docs-converter: No special table handling yet; table nodes would degrade to plain text via child walking.
- snooty-ast-to-mdx: Has logic to convert complex structures (e.g., directives) to mdast; table coverage may be present within its mdast converters.
- ingest-mongodb-docs: Explicit table rendering via renderSnootyTable, including multi-header tables.

Frontmatter and metadata
- realm-docs-converter: No frontmatter/metadata emitted.
- snooty-ast-to-mdx: Uses remark-frontmatter and can attach references metadata to the mdast tree; emits references.ts as a companion artifact.
- ingest-mongodb-docs: Extracts title and metadata (facets/meta nodes) for ingestion, not necessarily emitted as frontmatter.

Anchors and internal links
- realm-docs-converter: target nodes ignored; refname rendered as text. No anchor generation or cross-page anchor weaving.
- snooty-ast-to-mdx: Preserves ids/html_id for headings and can map to MDX anchors; manages imports for rich components.
- ingest-mongodb-docs: Builds link maps and handles internal/external links more robustly.

CLI UX
- realm-docs-converter: CLI exposes API mode (project, branch, base-url, out) and a local mode fallback (input, out). Writes conversion-warnings.log. Mirrors page path as output.
- snooty-ast-to-mdx: Programmatic APIs for JSON AST and zip conversion; logs counts; handles emitting multiple files per page (includes) and static assets.
- ingest-mongodb-docs: CLI/tests for ingestion, not a standalone converter CLI.

Tests and reliability
- realm-docs-converter: No unit tests in this package yet; practical logging and resilient API fetching.
- snooty-ast-to-mdx: Library-style with structured conversion layers; not much test content visible here, but modular.
- ingest-mongodb-docs: Includes Jest tests covering Snooty sources, tables, and conversion functions.

Summary of gaps for realm-docs-converter (opportunities)
- Broader AST coverage: admonitions, images, tables, directives, tabs, and anchors.
- Include handling in API mode: either expand or emit separate files; aggregate references.
- Asset handling: images and binaries from API or export archives.
- References/substitutions artifact: emit a project-level references file to aid consistent rendering.
- Frontmatter and metadata: optional YAML frontmatter based on Snooty meta/facets.
- Tests: add unit tests around AST mapping and API normalization.

Rationale for current scope
- realm-docs-converter aims to provide a minimal, stable Markdown output from Snooty Data API quickly, with clear logging for unresolved parts, leaving advanced features to future iterations or the richer MDX pipeline in references.
