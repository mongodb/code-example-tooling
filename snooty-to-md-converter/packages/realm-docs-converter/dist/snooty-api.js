"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.fetchSnootyProject = fetchSnootyProject;
async function tryFetchAny(url) {
    const res = await fetch(url);
    if (!res.ok) {
        throw new Error(`HTTP ${res.status} for ${url}`);
    }
    const ct = res.headers.get('content-type') || '';
    const text = await res.text();
    // Detect NDJSON or plain text where each line is a JSON object
    const looksNdjson = /ndjson/i.test(ct) || /\n\s*\{/.test(text);
    if (looksNdjson) {
        const lines = text.split(/\r?\n/)
            .map(l => l.trim())
            .filter(l => l && (l.startsWith('{') || l.startsWith('[')));
        const arr = [];
        for (const line of lines) {
            try {
                const parsed = JSON.parse(line);
                if (Array.isArray(parsed))
                    arr.push(...parsed);
                else
                    arr.push(parsed);
            }
            catch {
                // ignore non-JSON line
            }
        }
        if (arr.length) {
            return arr;
        }
        // fall through to single JSON parse if NDJSON heuristic failed to yield items
    }
    try {
        return JSON.parse(text);
    }
    catch (e) {
        // As a last resort, try NDJSON parsing even if content-type lied or heuristics failed
        const lines = text.split(/\r?\n/)
            .map(l => l.trim())
            .filter(l => l && (l.startsWith('{') || l.startsWith('[')));
        const arr = [];
        for (const line of lines) {
            try {
                const parsed = JSON.parse(line);
                if (Array.isArray(parsed))
                    arr.push(...parsed);
                else
                    arr.push(parsed);
            }
            catch {
                // ignore
            }
        }
        if (arr.length)
            return arr;
        const snippet = text.slice(0, 200).replace(/\s+/g, ' ').trim();
        throw new Error(`Unexpected non-JSON/NDJSON response for ${url}: ${e?.message || e}. Snippet: ${snippet}`);
    }
}
// Attempt to fetch the list of pages for a project/branch and each page's AST.
// Returns a normalized array of { path, ast }.
async function fetchSnootyProject({ project, branch = 'master', baseUrl = 'https://snooty-data-api.mongodb.com' }) {
    const errors = [];
    const candidates = [
        `${baseUrl.replace(/\/$/, '')}/prod/projects/${encodeURIComponent(project)}/${encodeURIComponent(branch)}/documents`,
    ];
    let pagesIndex;
    for (const url of candidates) {
        try {
            pagesIndex = await tryFetchAny(url);
            console.log(`[snooty-api] Using pages index endpoint: ${url}`);
            break;
        }
        catch (e) {
            errors.push(e?.message || String(e));
        }
    }
    if (!pagesIndex) {
        throw new Error(`Could not fetch pages index for project=${project} branch=${branch}. Tried: ${candidates.join(', ')}. Errors: ${errors.join(' | ')}`);
    }
    const items = Array.isArray(pagesIndex)
        ? pagesIndex
        : (pagesIndex.documents || pagesIndex.pages || pagesIndex.docs || pagesIndex.items || []);
    if (!Array.isArray(items) || items.length === 0) {
        console.warn(`[snooty-api] Pages index returned no page items for project=${project} branch=${branch}.`);
    }
    const results = [];
    // Stats for observability
    let countFetched = 0;
    let countDeleted = 0;
    let countAssetsSkipped = 0;
    let countNonPageSkipped = 0;
    let countFailed = 0;
    for (const item of items) {
        try {
            if (typeof item === 'string') {
                // If item is a path string, attempt common page AST endpoints
                const pagePath = item.replace(/^\/+/, '');
                const encBase = baseUrl.replace(/\/$/, '');
                const encProject = encodeURIComponent(project);
                const encBranch = encodeURIComponent(branch);
                const encPath = encodeURIComponent(pagePath);
                const pageCandidates = [
                    `${encBase}/prod/projects/${encProject}/${encBranch}/documents/${encPath}`,
                ];
                let ast;
                for (const u of pageCandidates) {
                    try {
                        ast = await tryFetchAny(u);
                        console.log(`[snooty-api] Using page AST endpoint: ${u} (path=${pagePath})`);
                        break;
                    }
                    catch {
                        // try next
                    }
                }
                if (!ast) {
                    throw new Error(`No AST for page ${pagePath}`);
                }
                results.push({ path: pagePath, ast });
                countFetched++;
            }
            else if (typeof item === 'object' && item) {
                const anyItem = item;
                // Skip non-page records, e.g. assets embedded in the index (images, binaries)
                if (anyItem.type && String(anyItem.type).toLowerCase() === 'asset') {
                    countAssetsSkipped++;
                    continue;
                }
                // Handle objects shaped like { type: 'page', data: { filename|page_id, ast } }
                if (anyItem.type && String(anyItem.type).toLowerCase() === 'page' && anyItem.data) {
                    const pd = anyItem.data;
                    const candidatePath = pd.filename || pd.page_id || pd.file || pd.slug || pd.id;
                    if (pd.deleted && candidatePath) {
                        // Mark as deleted; no AST will be processed
                        results.push({ path: String(candidatePath), ast: undefined, deleted: true });
                        countDeleted++;
                        continue;
                    }
                    if (pd.ast && candidatePath) {
                        results.push({ path: String(candidatePath), ast: pd.ast });
                        countFetched++;
                        continue;
                    }
                }
                const possiblePath = anyItem.path || anyItem.doc || anyItem.document || anyItem.file || anyItem.slug || anyItem.id;
                if (anyItem.ast && possiblePath) {
                    results.push({ path: String(possiblePath), ast: anyItem.ast });
                    countFetched++;
                }
                else {
                    const pageUrl = (anyItem).url;
                    // If we don't have a path, a url, or an embedded ast, skip this entry (it's not a page descriptor)
                    if (!possiblePath && !pageUrl) {
                        countNonPageSkipped++;
                        continue;
                    }
                    const pagePath = String(possiblePath || 'index');
                    let ast;
                    if (pageUrl) {
                        ast = await tryFetchAny(pageUrl);
                    }
                    else {
                        const encBase = baseUrl.replace(/\/$/, '');
                        const encProject = encodeURIComponent(project);
                        const encBranch = encodeURIComponent(branch);
                        const encPath = encodeURIComponent(pagePath);
                        const pageCandidates = [
                            `${encBase}/prod/projects/${encProject}/${encBranch}/documents/${encPath}`,
                        ];
                        for (const u of pageCandidates) {
                            try {
                                ast = await tryFetchAny(u);
                                console.log(`[snooty-api] Using page AST endpoint: ${u} (path=${pagePath})`);
                                break;
                            }
                            catch {
                                // try next
                            }
                        }
                    }
                    if (!ast) {
                        // Not a fatal error for non-page items; skip quietly
                        continue;
                    }
                    results.push({ path: pagePath, ast });
                    countFetched++;
                }
            }
        }
        catch (e) {
            console.warn(`[snooty-api] Failed to fetch page: ${JSON.stringify(item)} -> ${e?.message || e}`);
            countFailed++;
        }
    }
    const summary = `Fetched=${countFetched}, Deleted=${countDeleted}, SkippedAssets=${countAssetsSkipped}, SkippedNonPage=${countNonPageSkipped}, Failed=${countFailed}`;
    if (countFailed > 0) {
        console.warn(`[snooty-api] Summary for project=${project} branch=${branch}: ${summary}`);
    }
    else {
        console.log(`[snooty-api] Summary for project=${project} branch=${branch}: ${summary}`);
    }
    return results;
}
