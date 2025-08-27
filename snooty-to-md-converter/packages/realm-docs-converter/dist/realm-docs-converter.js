"use strict";
var __createBinding = (this && this.__createBinding) || (Object.create ? (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    var desc = Object.getOwnPropertyDescriptor(m, k);
    if (!desc || ("get" in desc ? !m.__esModule : desc.writable || desc.configurable)) {
      desc = { enumerable: true, get: function() { return m[k]; } };
    }
    Object.defineProperty(o, k2, desc);
}) : (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    o[k2] = m[k];
}));
var __setModuleDefault = (this && this.__setModuleDefault) || (Object.create ? (function(o, v) {
    Object.defineProperty(o, "default", { enumerable: true, value: v });
}) : function(o, v) {
    o["default"] = v;
});
var __importStar = (this && this.__importStar) || (function () {
    var ownKeys = function(o) {
        ownKeys = Object.getOwnPropertyNames || function (o) {
            var ar = [];
            for (var k in o) if (Object.prototype.hasOwnProperty.call(o, k)) ar[ar.length] = k;
            return ar;
        };
        return ownKeys(o);
    };
    return function (mod) {
        if (mod && mod.__esModule) return mod;
        var result = {};
        if (mod != null) for (var k = ownKeys(mod), i = 0; i < k.length; i++) if (k[i] !== "default") __createBinding(result, mod, k[i]);
        __setModuleDefault(result, mod);
        return result;
    };
})();
Object.defineProperty(exports, "__esModule", { value: true });
exports.convertRealmDocs = convertRealmDocs;
exports.convertRealmDocsFromApi = convertRealmDocsFromApi;
// src/realm-docs-converter.ts
const fs = __importStar(require("fs"));
const path = __importStar(require("path"));
const snooty_1 = require("./converters/snooty");
const snooty_api_1 = require("./snooty-api");
const ast_to_md_1 = require("./ast-to-md");
async function convertRealmDocs(options) {
    const { inputDir, outputDir } = options;
    // Overwrite output directory on each run
    try {
        if (fs.existsSync(outputDir)) {
            fs.rmSync(outputDir, { recursive: true, force: true });
        }
    }
    catch (e) {
        console.warn(`Warning: failed to remove output directory ${outputDir}. Proceeding to recreate.`, e);
    }
    fs.mkdirSync(outputDir, { recursive: true });
    // Copy shared images (if present) from input/images -> output/images
    try {
        const srcImages = path.join(inputDir, 'images');
        const destImages = path.join(outputDir, 'images');
        if (fs.existsSync(srcImages)) {
            copyDirRecursive(srcImages, destImages);
            console.log(`Copied shared images: ${srcImages} -> ${destImages}`);
        }
    }
    catch (e) {
        console.warn(`Warning: failed to copy shared images from local input`, e);
    }
    // Get all .txt files (assuming Snooty uses .txt for RST files)
    const files = getAllFiles(inputDir, ['.txt', '.rst']);
    let convertedCount = 0;
    for (const file of files) {
        try {
            // Read the source file
            const content = fs.readFileSync(file, 'utf8');
            // Parse snooty content
            const parsedContent = await (0, snooty_1.parseSnootyContent)(content, {
                filePath: file,
                basePath: inputDir,
                resolveIncludes: options.handleIncludes,
                resolveSubstitutions: options.handleSubstitutions,
                resolveRefs: options.handleRefs,
            });
            // Convert to markdown
            const markdown = (0, snooty_1.convertRSTToMarkdown)(parsedContent);
            // Write to output file
            const relativePath = path.relative(inputDir, file);
            const outputPath = path.join(outputDir, relativePath.replace(/\.(txt|rst)$/, '.md'));
            // Ensure directory exists
            const outputFileDir = path.dirname(outputPath);
            if (!fs.existsSync(outputFileDir)) {
                fs.mkdirSync(outputFileDir, { recursive: true });
            }
            const rewritten = rewriteImagePaths(markdown, outputPath, outputDir);
            fs.writeFileSync(outputPath, rewritten, 'utf8');
            console.log(`Converted: ${relativePath}`);
            convertedCount++;
        }
        catch (error) {
            console.error(`Error converting ${file}:`, error);
        }
    }
    return convertedCount;
}
async function convertRealmDocsFromApi(options) {
    const { project, outputDir, branch, baseUrl } = options;
    // Overwrite output directory on each run
    try {
        if (fs.existsSync(outputDir)) {
            fs.rmSync(outputDir, { recursive: true, force: true });
        }
    }
    catch (e) {
        console.warn(`Warning: failed to remove output directory ${outputDir}. Proceeding to recreate.`, e);
    }
    fs.mkdirSync(outputDir, { recursive: true });
    // Attempt to copy shared images from a configured directory (optional for API mode)
    try {
        const configuredImagesDir = process.env.SHARED_IMAGES_DIR;
        if (configuredImagesDir) {
            const resolved = path.resolve(configuredImagesDir);
            if (fs.existsSync(resolved) && fs.statSync(resolved).isDirectory()) {
                const dest = path.join(outputDir, 'images');
                copyDirRecursive(resolved, dest);
                console.log(`Copied shared images (API mode): ${resolved} -> ${dest}`);
            }
            else {
                console.warn(`Configured SHARED_IMAGES_DIR does not exist or is not a directory: ${resolved}`);
            }
        }
        else {
            // Fall back to a sibling 'images' directory next to the current working directory, if present
            const fallback = path.resolve('images');
            if (fs.existsSync(fallback) && fs.statSync(fallback).isDirectory()) {
                const dest = path.join(outputDir, 'images');
                copyDirRecursive(fallback, dest);
                console.log(`Copied shared images from fallback: ${fallback} -> ${dest}`);
            }
        }
    }
    catch (e) {
        console.warn(`Warning: failed to copy shared images for API mode. You can set SHARED_IMAGES_DIR to point to a local images folder.`, e);
    }
    // Fetch pages and ASTs
    const pages = await (0, snooty_api_1.fetchSnootyProject)({ project, branch, baseUrl });
    // Global substitutions map accumulated across pages
    const substitutions = {};
    // Simple log collector per doc
    const warnings = [];
    let convertedCount = 0;
    for (const page of pages) {
        try {
            // Normalize output file path for both conversion and deletion cases
            let relPath = page.path.replace(/\\/g, '/').replace(/\/+$/, '');
            if (/\.(txt|rst|mdx?)$/i.test(relPath)) {
                relPath = relPath.replace(/\.(txt|rst|mdx?)$/i, '.md');
            }
            else if (!/\.md$/i.test(relPath)) {
                relPath = `${relPath}.md`;
            }
            const outPath = path.join(outputDir, relPath);
            if (page.deleted) {
                // Handle deleted pages by removing previously generated output file, if present
                if (fs.existsSync(outPath)) {
                    fs.unlinkSync(outPath);
                    console.log(`Deleted (API): ${page.path} -> ${relPath}`);
                }
                else {
                    console.log(`Skipped deleted (no file): ${page.path}`);
                }
                continue;
            }
            const md = (0, ast_to_md_1.astToMarkdown)(page.ast, {
                substitutions,
                onWarn: (message, ctx) => {
                    warnings.push({ path: ctx?.path || page.path, message });
                },
                docPath: page.path,
            });
            const rewritten = rewriteImagePaths(md, outPath, outputDir);
            const dir = path.dirname(outPath);
            if (!fs.existsSync(dir))
                fs.mkdirSync(dir, { recursive: true });
            fs.writeFileSync(outPath, rewritten, 'utf8');
            console.log(`Converted (API): ${page.path} -> ${relPath}`);
            convertedCount++;
        }
        catch (e) {
            console.error(`Error converting page ${page.path}:`, e);
        }
    }
    // Write warnings to a log file with pointers (with summary and de-duplication)
    if (warnings.length) {
        const logPath = path.join(outputDir, 'conversion-warnings.log');
        // De-duplicate identical (path, message) pairs
        const uniqueMap = new Map();
        for (const w of warnings) {
            const key = `${w.path}:::${w.message}`;
            if (!uniqueMap.has(key))
                uniqueMap.set(key, w);
        }
        // Count by message type among unique warnings
        const countsByType = {};
        for (const w of uniqueMap.values()) {
            countsByType[w.message] = (countsByType[w.message] || 0) + 1;
        }
        const totalRaw = warnings.length;
        const totalUnique = uniqueMap.size;
        const docsWithWarnings = new Set(Array.from(uniqueMap.values()).map(w => w.path));
        const byTypeLines = Object.entries(countsByType)
            .sort((a, b) => (b[1] - a[1]) || a[0].localeCompare(b[0]))
            .map(([msg, count]) => `- ${msg}: ${count}`);
        const detailed = Array.from(uniqueMap.values())
            .sort((a, b) => (a.path === b.path ? a.message.localeCompare(b.message) : a.path.localeCompare(b.path)))
            .map((w) => `${w.path}: ${w.message}`);
        const outLines = [
            'Conversion Warnings',
            '',
            `Total warnings (raw): ${totalRaw}`,
            `Total warnings (unique): ${totalUnique}`,
            `Documents with warnings: ${docsWithWarnings.size}`,
            '',
            'By type (unique counts):',
            ...byTypeLines,
            '',
            'Detailed warnings (unique, sorted):',
            ...detailed,
            '',
        ];
        fs.writeFileSync(logPath, outLines.join('\n'), 'utf8');
        console.warn(`Conversion completed with ${totalRaw} warnings (${totalUnique} unique across ${docsWithWarnings.size} docs). See ${logPath}`);
    }
    return convertedCount;
}
function getAllFiles(dir, extensions) {
    let results = [];
    const items = fs.readdirSync(dir);
    for (const item of items) {
        const fullPath = path.join(dir, item);
        const stat = fs.statSync(fullPath);
        if (stat.isDirectory()) {
            results = results.concat(getAllFiles(fullPath, extensions));
        }
        else {
            const ext = path.extname(fullPath);
            if (extensions.includes(ext)) {
                results.push(fullPath);
            }
        }
    }
    return results;
}
// Simple recursive directory copy (preserves hierarchy). Overwrites destination.
function copyDirRecursive(srcDir, destDir) {
    fs.mkdirSync(destDir, { recursive: true });
    // Node 16+ has fs.cpSync
    if (fs.cpSync) {
        fs.cpSync(srcDir, destDir, { recursive: true });
        return;
    }
    // Fallback
    const entries = fs.readdirSync(srcDir, { withFileTypes: true });
    for (const entry of entries) {
        const srcPath = path.join(srcDir, entry.name);
        const dstPath = path.join(destDir, entry.name);
        if (entry.isDirectory()) {
            copyDirRecursive(srcPath, dstPath);
        }
        else if (entry.isFile()) {
            fs.copyFileSync(srcPath, dstPath);
        }
    }
}
// Rewrites Markdown image URLs that start with "/images/" to a relative path from the output file
function rewriteImagePaths(markdown, outPath, outputRoot) {
    try {
        const fromDir = path.dirname(outPath);
        const imagesRoot = path.join(outputRoot, 'images');
        // If the markdown contains references to /images/, ensure the output images directory exists
        const mdImageRe = /!\[([^\]]*)\]\(\s*\/images\/([^\)\s]+)\s*\)/g;
        const htmlImgRe = /(<img\b[^>]*\bsrc=["'])\s*\/images\/([^"'>\s]+)(["'][^>]*>)/gi;
        const hasImages = mdImageRe.test(markdown) || htmlImgRe.test(markdown);
        if (hasImages && !fs.existsSync(imagesRoot)) {
            try {
                fs.mkdirSync(imagesRoot, { recursive: true });
            }
            catch {
                // ignore mkdir failure; we'll fall back to leaving paths as-is if dir still missing
            }
        }
        // If there is still no local images directory, do not rewrite absolute /images paths
        if (!fs.existsSync(imagesRoot)) {
            return markdown;
        }
        // POSIX-style path for Markdown
        const relToImagesRoot = path.relative(fromDir, imagesRoot).replace(/\\/g, '/');
        // Replace inline Markdown images: ![alt](/images/..)
        let out = markdown.replace(mdImageRe, (_m, alt, imgPath) => `![${alt}](${relToImagesRoot}/${imgPath})`);
        // Replace HTML <img> tags with /images/ src
        out = out.replace(htmlImgRe, (_m, pre, imgPath, post) => `${pre}${relToImagesRoot}/${imgPath}${post}`);
        return out;
    }
    catch {
        return markdown;
    }
}
