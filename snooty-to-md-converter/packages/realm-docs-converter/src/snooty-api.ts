// src/snooty-api.ts
// Minimal Snooty Data API client for fetching a single project's pages
// NOTE: The exact endpoints can vary by deployment. We allow configuring a baseUrl and branch
// and try a couple of common patterns. This is intentionally lightweight.

export interface SnootyPage {
  path: string; // e.g. "docs/intro" or "index"
  ast: any;     // Snooty AST JSON for the page (undefined/null if deleted)
  deleted?: boolean; // indicates this page was marked as deleted by the API
}

export interface FetchOptions {
  project: string; // e.g. "realm"
  branch?: string; // e.g. "master", "main", "stable", "current"
  baseUrl?: string; // e.g. "https://snooty-data-api.mongodb.workers.dev"
}

async function tryFetchAny(url: string): Promise<any> {
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
    const arr: any[] = [];
    for (const line of lines) {
      try {
        const parsed = JSON.parse(line);
        if (Array.isArray(parsed)) arr.push(...parsed); else arr.push(parsed);
      } catch {
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
  } catch (e: any) {
    // As a last resort, try NDJSON parsing even if content-type lied or heuristics failed
    const lines = text.split(/\r?\n/)
      .map(l => l.trim())
      .filter(l => l && (l.startsWith('{') || l.startsWith('[')));
    const arr: any[] = [];
    for (const line of lines) {
      try {
        const parsed = JSON.parse(line);
        if (Array.isArray(parsed)) arr.push(...parsed); else arr.push(parsed);
      } catch {
        // ignore
      }
    }
    if (arr.length) return arr;

    const snippet = text.slice(0, 200).replace(/\s+/g, ' ').trim();
    throw new Error(`Unexpected non-JSON/NDJSON response for ${url}: ${e?.message || e}. Snippet: ${snippet}`);
  }
}

// Attempt to fetch the list of pages for a project/branch and each page's AST.
// Returns a normalized array of { path, ast }.
export async function fetchSnootyProject({ project, branch = 'master', baseUrl = 'https://snooty-data-api.mongodb.com' }: FetchOptions): Promise<SnootyPage[]> {
  const errors: string[] = [];

  // Try MongoDB Snooty Data API known endpoint first, then fallbacks.
  const candidates = [
    `${baseUrl.replace(/\/$/, '')}/prod/projects/${encodeURIComponent(project)}/${encodeURIComponent(branch)}/documents`,
    // fallbacks (older or alternative deployments)
    `${baseUrl.replace(/\/$/, '')}/project/${encodeURIComponent(project)}/branch/${encodeURIComponent(branch)}/pages`,
    `${baseUrl.replace(/\/$/, '')}/projects/${encodeURIComponent(project)}/${encodeURIComponent(branch)}/pages`,
  ];

  let pagesIndex: any;
  for (const url of candidates) {
    try {
      pagesIndex = await tryFetchAny(url);
      console.log(`[snooty-api] Using pages index endpoint: ${url}`);
      break;
    } catch (e: any) {
      errors.push(e?.message || String(e));
    }
  }

  if (!pagesIndex) {
    throw new Error(`Could not fetch pages index for project=${project} branch=${branch}. Tried: ${candidates.join(', ')}. Errors: ${errors.join(' | ')}`);
  }

  // Support multiple shapes of the pages index; normalize to an array of page descriptors with a path and a URL to fetch the AST.
  type PageIndexItem = { path?: string; url?: string; ast?: any } | string;
  const items: PageIndexItem[] = Array.isArray(pagesIndex)
    ? pagesIndex
    : (pagesIndex.documents || pagesIndex.pages || pagesIndex.docs || pagesIndex.items || []);

  const results: SnootyPage[] = [];

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
          `${encBase}/project/${encProject}/branch/${encBranch}/page/${encPath}`,
          `${encBase}/projects/${encProject}/${encBranch}/page/${encPath}`,
        ];
        let ast: any | undefined;
        for (const u of pageCandidates) {
          try {
            ast = await tryFetchAny(u);
            console.log(`[snooty-api] Using page AST endpoint: ${u} (path=${pagePath})`);
            break;
          } catch {
            // try next
          }
        }
        if (!ast) {
          throw new Error(`No AST for page ${pagePath}`);
        }
        results.push({ path: pagePath, ast });
      } else if (typeof item === 'object' && item) {
        const anyItem = item as any;
        // Skip non-page records, e.g., assets embedded in the index (images, binaries)
        if (anyItem.type && String(anyItem.type).toLowerCase() === 'asset') {
          continue;
        }
        // Handle objects shaped like { type: 'page', data: { filename|page_id, ast } }
        if (anyItem.type && String(anyItem.type).toLowerCase() === 'page' && anyItem.data) {
          const pd = anyItem.data as any;
          const candidatePath = pd.filename || pd.page_id || pd.file || pd.slug || pd.id;
          if (pd.deleted && candidatePath) {
            // Mark as deleted; no AST will be processed
            results.push({ path: String(candidatePath), ast: undefined, deleted: true });
            continue;
          }
          if (pd.ast && candidatePath) {
            results.push({ path: String(candidatePath), ast: pd.ast });
            continue;
          }
        }
        const possiblePath = anyItem.path || anyItem.doc || anyItem.document || anyItem.file || anyItem.slug || anyItem.id;
        if (anyItem.ast && possiblePath) {
          results.push({ path: String(possiblePath), ast: anyItem.ast });
        } else {
          const pageUrl = (anyItem).url as string | undefined;
          // If we don't have a path, a url, or an embedded ast, skip this entry (it's not a page descriptor)
          if (!possiblePath && !pageUrl) {
            continue;
          }
          const pagePath = String(possiblePath || 'index');
          let ast: any | undefined;
          if (pageUrl) {
            ast = await tryFetchAny(pageUrl);
          } else {
            const encBase = baseUrl.replace(/\/$/, '');
            const encProject = encodeURIComponent(project);
            const encBranch = encodeURIComponent(branch);
            const encPath = encodeURIComponent(pagePath);
            const pageCandidates = [
              `${encBase}/prod/projects/${encProject}/${encBranch}/documents/${encPath}`,
              `${encBase}/project/${encProject}/branch/${encBranch}/page/${encPath}`,
              `${encBase}/projects/${encProject}/${encBranch}/page/${encPath}`,
            ];
            for (const u of pageCandidates) {
              try {
                ast = await tryFetchAny(u);
                console.log(`[snooty-api] Using page AST endpoint: ${u} (path=${pagePath})`);
                break;
              } catch {
                // try next
              }
            }
          }
          if (!ast) {
            // Not a fatal error for non-page items; skip quietly
            continue;
          }
          results.push({ path: pagePath, ast });
        }
      }
    } catch (e: any) {
      console.warn(`[snooty-api] Failed to fetch page: ${JSON.stringify(item)} -> ${e?.message || e}`);
    }
  }

  return results;
}
