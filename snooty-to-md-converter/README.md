# snooty-to-md-converter

This project contains a converter for Snooty-based docs projects to Markdown. 
This is intended to help migrate the Realm docs, specifically, from Snooty to Markdown. 
This tool ingests the Snooty Data API for a single project (default: `realm`), 
and outputs Markdown files, handling links, refs, substitutions, and includes (best-effort), 
logging any issues with pointers to the source page.

## Realm Docs Converter

### Prerequisites
- Node.js >= 18 (for built-in fetch)
- npm >= 8
- This project checked out locally

### Install dependencies (from project root)
```
npm ci
```

## Build the converter (from project root)
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

The converter writes one `.md` file per page, mirroring the page path. A `conversion-warnings.log` file is written in the output directory if any unresolved includes/substitutions/refs occur.

### Local directory fallback (optional)
For legacy/local conversion of a checked-out Snooty project directory containing `.txt/.rst` files, you can use:
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

## License

This project is licensed under the [Apache 2.0 License](LICENSE).
