"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.astToMarkdown = astToMarkdown;
const EOL = "\n";
function astToMarkdown(root, options = {}) {
    const ctx = {
        substitutions: options.substitutions || {},
        onWarn: options.onWarn || (() => { }),
        docPath: options.docPath || '',
    };
    const lines = [];
    // Track the last emitted heading level (1-6) so certain directives (e.g., cards)
    // can render a heading relative to the previous one.
    let lastHeadingLevel = 1;
    function posToStr(pos) {
        if (!pos)
            return undefined;
        try {
            // common position shapes: { start: { line, column }, end: { line, column } }
            const s = pos.start ? `${pos.start.line}:${pos.start.column}` : '';
            return s || undefined;
        }
        catch {
            return undefined;
        }
    }
    function textOf(node) {
        if (!node)
            return '';
        if (typeof node === 'string')
            return node;
        if (node.value)
            return String(node.value);
        if (Array.isArray(node.children))
            return node.children.map(textOf).join('');
        return '';
    }
    // Best-effort stringify of a directive or node argument (string or token array)
    function argToString(arg) {
        if (!arg)
            return '';
        if (typeof arg === 'string')
            return arg.trim();
        if (Array.isArray(arg))
            return arg.map((a) => (a?.value ?? '')).join('').trim();
        return String(arg ?? '').trim();
    }
    function inline(node) {
        switch (node.type) {
            case 'text':
                return node.value || '';
            case 'literal':
            case 'inline_literal':
            case 'code':
            case 'literal_strong':
            case 'literal_emphasis':
                return '`' + textOf(node) + '`';
            case 'emphasis':
                return '*' + (node.children ? node.children.map(inline).join('') : textOf(node)) + '*';
            case 'strong':
            case 'strong_emphasis':
                return '**' + (node.children ? node.children.map(inline).join('') : textOf(node)) + '**';
            case 'substitution_reference': {
                const key = node.name || node.value || textOf(node);
                const val = ctx.substitutions[key];
                if (val == null) {
                    ctx.onWarn(`Unresolved substitution |${key}|`, { path: ctx.docPath, position: posToStr(node.position) });
                    return `|${key}|`;
                }
                return String(val);
            }
            case 'reference': {
                // Links can be refuri (URL) or refname (internal)
                const label = textOf(node);
                if (node.refuri)
                    return `[${label}](${node.refuri})`;
                if (node.refname)
                    return label; // best-effort inline text for :ref:
                return label;
            }
            case 'ref_role': {
                // Snooty inline ref role; render visible label text (children) best-effort
                // Children may include literals/emphasis; recurse to preserve formatting
                return (node.children || []).map(inline).join('') || textOf(node);
            }
            case 'image': {
                const url = node.refuri || node.uri || node.url || '';
                const alt = node.alt || textOf(node) || '';
                if (!url) {
                    ctx.onWarn(`Image node missing URL; emitting alt text only`, { path: ctx.docPath, position: posToStr(node.position) });
                    return alt;
                }
                return `![${alt}](${url})`;
            }
            default:
                // try children concat
                if (node.children && node.children.length) {
                    return node.children.map(inline).join('');
                }
                return textOf(node);
        }
    }
    function pushAnchor(id) {
        if (id) {
            lines.push(`<a id="${id}"></a>`);
            lines.push('');
        }
    }
    function renderTable(t) {
        // Best-effort parse of Snooty table -> GFM table
        // Look for thead/tbody rows; otherwise collect any row/entry structure in order
        const headers = [];
        const rows = [];
        function cellsOf(node) {
            if (!node)
                return [];
            if (Array.isArray(node.children)) {
                const entryNodes = node.children.filter((c) => c && (c.type === 'entry' || c.type === 'cell'));
                if (entryNodes.length) {
                    return entryNodes.map((e) => (Array.isArray(e.children) ? e.children.map((ch) => (ch.type === 'paragraph' ? (ch.children || []).map(inline).join('') : inline(ch))).join(' ') : textOf(e)));
                }
            }
            return [];
        }
        function walk(node, inHead = false) {
            if (!node)
                return;
            const type = node.type;
            if (type === 'thead' || type === 'tgrouphead') {
                for (const r of node.children || []) {
                    if (r.type === 'row')
                        headers.push(cellsOf(r));
                }
                return;
            }
            if (type === 'tbody' || type === 'tgroupbody') {
                for (const r of node.children || []) {
                    if (r.type === 'row')
                        rows.push(cellsOf(r));
                }
                return;
            }
            if (type === 'row') {
                const cells = cellsOf(node);
                if (inHead)
                    headers.push(cells);
                else
                    rows.push(cells);
                return;
            }
            if (Array.isArray(node.children)) {
                for (const c of node.children)
                    walk(c, inHead || type === 'thead');
            }
        }
        walk(t);
        const header = headers[0] || rows.shift() || [];
        if (!header.length) {
            ctx.onWarn(`Table could not be rendered (no header/cells found)`, { path: ctx.docPath, position: posToStr(t.position) });
            return;
        }
        const sep = '|' + header.map(() => ' --- ').join('|') + '|';
        const fmt = (r) => '|' + header.map((_, i) => (r[i] ?? '').replace(/\n/g, ' ')).join('|') + '|';
        lines.push(fmt(header));
        lines.push(sep);
        for (const r of rows)
            lines.push(fmt(r));
        lines.push('');
    }
    /**
       * Render an admonition node into GitHub‑flavored Markdown blockquotes.
       *
       * Inputs
       * - node: Snooty AST node representing either a named admonition directive (.. note::, .. warning::, .. versionadded::, etc.)
       *         or a node with type equal to an admonition (e.g., { type: 'note', children: [...] }).
       * - kind: Lower/upper‑case label of the admonition type. Examples: "note", "warning", "versionadded", "versionchanged".
       * - parentDepth: The current structural depth; used only when delegating to block() so nested content renders correctly.
       *
       * Behavior
       * - Generic admonitions (note, warning, tip, important, caution, danger, seealso, example, see, deprecated):
       *   • Emits a single‑line header "> {CapitalizedKind}:".
       *   • Renders all child nodes normally via block(), then prefixes each emitted line with "> ", including blank lines as plain ">".
       *   • Adds a trailing blank line after the admonition block.
       *
       * - Special handling for versionadded/versionchanged:
       *   • Computes a header label of either "Version added" or "Version changed".
       *   • Attempts to extract a version string from several possible sources commonly found in Snooty ASTs:
       *       argument (string or array tokens), node.value, or options.version.
       *   • Emits header as "> {Label}: {version}" if a version was found; otherwise "> {Label}:".
       *   • Renders all children via block() and prefixes each resulting line with "> " so lists, code blocks, tables, etc. stay inside the admonition.
       *   • Adds a trailing blank line after the block.
       */
    function renderAdmonition(node, kind, parentDepth) {
        const k = String(kind || '').toLowerCase();
        // Special-case version directives to render with proper spacing and optional version argument
        if (k === 'versionadded' || k === 'versionchanged') {
            const label = k === 'versionadded' ? 'Version added' : 'Version changed';
            const n = node;
            // Try to extract version from argument/value/options and separate any trailing body text
            let version = '';
            let remainder = '';
            const extractVersion = (s) => {
                const m = String(s || '').trim().match(/^([0-9]+(?:\.[0-9]+)*(?:-[0-9A-Za-z.]+)?)(.*)$/);
                if (m) {
                    version = m[1] || '';
                    remainder = (m[2] || '').trim();
                }
                else if (!version) {
                    version = String(s || '').trim();
                }
            };
            const arg = n.argument;
            let remainderNodes;
            if (typeof arg === 'string') {
                extractVersion(arg);
            }
            else if (Array.isArray(arg)) {
                // Build a string to extract version, but preserve remainder as node tokens to keep formatting
                const tokens = arg;
                extractVersion(tokens.map((a) => (a?.value ?? '')).join(''));
                // Attempt to remove the leading version text from the token list to form remainder nodes
                if (version) {
                    let consumed = 0;
                    const out = [];
                    for (let i = 0; i < tokens.length; i++) {
                        const t = tokens[i];
                        const val = String(t?.value ?? '');
                        if (consumed < version.length) {
                            const need = version.length - consumed;
                            if (val.length <= need) {
                                consumed += val.length;
                                continue; // skip this token entirely as part of the version
                            }
                            else {
                                // Split this token: drop the first 'need' chars, keep the rest as remainder
                                const restVal = val.slice(need);
                                const copy = { ...t, value: restVal };
                                out.push(copy);
                                consumed = version.length;
                                continue;
                            }
                        }
                        else {
                            out.push(t);
                        }
                    }
                    // Trim leading whitespace in the first texty node
                    if (out.length && typeof out[0]?.value === 'string') {
                        out[0].value = String(out[0].value).replace(/^\s+/, '');
                    }
                    remainderNodes = out;
                }
            }
            if (!version && n.value)
                extractVersion(String(n.value));
            if (!version && n.options && n.options.version)
                extractVersion(String(n.options.version));
            const header = version ? `> ${label}: ${version}` : `> ${label}:`;
            lines.push(header);
            // Build the list of children for the admonition body, injecting any remainder text as the first paragraph
            const bodyChildren = Array.isArray(node.children) ? [...node.children] : [];
            if (Array.isArray(remainderNodes) && remainderNodes.length) {
                bodyChildren.unshift({ type: 'paragraph', children: remainderNodes });
            }
            else if (remainder) {
                // Best-effort inline RST → Markdown for plain-string remainders
                const rstToMd = (s) => {
                    let out = String(s);
                    // ``code`` → `code`
                    out = out.replace(/``([^`]+)``/g, '`$1`');
                    // :ref:`Text <label>` → Text
                    out = out.replace(/:ref:`([^<`]*)<([^>`]*)>`/g, (_m, text) => String(text).trim());
                    // :ref:`label` → label
                    out = out.replace(/:ref:`([^<`][^`>]*)`/g, (_m, label) => String(label).trim());
                    return out;
                };
                bodyChildren.unshift({ type: 'paragraph', children: [{ type: 'text', value: rstToMd(remainder) }] });
            }
            // Render children into the admonition block
            for (const c of bodyChildren) {
                const start = lines.length;
                block(c, parentDepth + 1);
                const end = lines.length;
                for (let i = start; i < end; i++) {
                    lines[i] = lines[i] === '' ? '>' : `> ${lines[i]}`;
                }
            }
            lines.push('');
            return;
        }
        // Default admonition rendering
        const label = kind.charAt(0).toUpperCase() + kind.slice(1).toLowerCase();
        lines.push(`> ${label}:`);
        // Render all children normally, then prefix the emitted lines so all content stays within the admonition
        for (const c of node.children || []) {
            const start = lines.length;
            block(c, parentDepth + 1);
            const end = lines.length;
            for (let i = start; i < end; i++) {
                lines[i] = lines[i] === '' ? '>' : `> ${lines[i]}`;
            }
        }
        // Trailing blank line after the admonition block
        lines.push('');
    }
    function block(node, depth) {
        switch (node.type) {
            case 'section': {
                // Find first title child for header
                const titleNode = (node.children || []).find((c) => c.type === 'title');
                let sectionTitleText;
                if (titleNode) {
                    // Bump headings up one level (min 1)
                    const level = Math.max(1, Math.min(6, depth - 1));
                    const hashes = '#'.repeat(level);
                    sectionTitleText = textOf(titleNode).trim();
                    lines.push(`${hashes} ${sectionTitleText}`);
                    lastHeadingLevel = hashes.length;
                    // Anchor from ids/html_id if present
                    const anchorId = titleNode.html_id || titleNode.ids?.[0] || node.html_id || node.ids?.[0];
                    pushAnchor(anchorId);
                }
                for (const c of node.children || []) {
                    if (c === titleNode)
                        continue;
                    // If the child is a heading/title with the same text as the section title, skip to avoid duplicate headings
                    if (sectionTitleText && c?.type && (c.type === 'heading' || c.type === 'title')) {
                        const childText = textOf(c).trim();
                        if (childText === sectionTitleText)
                            continue;
                    }
                    block(c, depth + 1);
                }
                break;
            }
            case 'title': {
                // Standalone page title (not nested under section)
                const level = Math.max(1, Math.min(6, depth - 1));
                const hashes = '#'.repeat(level);
                lines.push(`${hashes} ${textOf(node)}`);
                lastHeadingLevel = hashes.length;
                const anchorId = node.html_id || node.ids?.[0];
                pushAnchor(anchorId);
                break;
            }
            case 'heading': {
                const level0 = Math.max(1, Math.min(6, node.depth || depth || 1));
                const level = Math.max(1, level0 - 1);
                const hashes = '#'.repeat(level);
                lines.push(`${hashes} ${textOf(node)}`);
                lastHeadingLevel = hashes.length;
                const anchorId = node.html_id || node.ids?.[0];
                pushAnchor(anchorId);
                break;
            }
            case 'paragraph': {
                const content = (node.children || []).map(inline).join('');
                lines.push(content);
                lines.push('');
                break;
            }
            case 'bullet_list': {
                for (const item of node.children || []) {
                    // list_item
                    const txt = (item.children || []).map((n) => (n.type === 'paragraph' ? (n.children || []).map(inline).join('') : inline(n))).join(' ');
                    lines.push(`- ${txt}`);
                }
                lines.push('');
                break;
            }
            case 'enumerated_list':
            case 'ordered_list': {
                let i = 1;
                for (const item of node.children || []) {
                    const txt = (item.children || []).map((n) => (n.type === 'paragraph' ? (n.children || []).map(inline).join('') : inline(n))).join(' ');
                    lines.push(`${i}. ${txt}`);
                    i++;
                }
                lines.push('');
                break;
            }
            case 'list': {
                // Generic list node that may be ordered or unordered
                const etRaw = node.enumtype;
                const et = typeof etRaw === 'string' ? etRaw.toLowerCase() : undefined;
                const likelyOrderedTypes = new Set(['ordered', 'arabic', 'loweralpha', 'upperalpha', 'lowerroman', 'upperroman', 'alphabetical', 'roman']);
                let ordered = false;
                if (et && likelyOrderedTypes.has(et))
                    ordered = true;
                if (!ordered && node.ordered === true)
                    ordered = true;
                if (!ordered && (typeof node.start === 'number' || typeof node.startat === 'number'))
                    ordered = true;
                const startAt = typeof node.startat === 'number' ? node.startat
                    : (typeof node.start === 'number' ? node.start : 1);
                if (ordered) {
                    let i = Math.max(1, startAt || 1);
                    for (const item of node.children || []) {
                        const txt = (item.children || []).map((n) => (n.type === 'paragraph' ? (n.children || []).map(inline).join('') : inline(n))).join(' ');
                        lines.push(`${i}. ${txt}`);
                        i++;
                    }
                }
                else {
                    for (const item of node.children || []) {
                        const txt = (item.children || []).map((n) => (n.type === 'paragraph' ? (n.children || []).map(inline).join('') : inline(n))).join(' ');
                        lines.push(`- ${txt}`);
                    }
                }
                lines.push('');
                break;
            }
            case 'literal_block':
            case 'code':
            case 'code_block': {
                const language = (node.language ? String(node.language) : (node.lang ? String(node.lang) : ''));
                const code = textOf(node);
                lines.push('```' + language);
                for (const l of String(code).split(/\r?\n/)) {
                    lines.push(l);
                }
                lines.push('```');
                lines.push('');
                break;
            }
            case 'table': {
                renderTable(node);
                break;
            }
            case 'admonition': {
                const kind = node.admonition_type || 'note';
                renderAdmonition(node, kind, depth);
                break;
            }
            case 'note':
            case 'warning':
            case 'tip':
            case 'important':
            case 'caution':
            case 'seealso': {
                renderAdmonition(node, String(node.type), depth);
                break;
            }
            case 'versionadded': {
                // Treat versionadded as a regular admonition (ensures all children are blockquoted)
                renderAdmonition(node, 'versionadded', depth);
                break;
            }
            case 'versionchanged': {
                // Treat versionchanged as a regular admonition
                renderAdmonition(node, 'versionchanged', depth);
                break;
            }
            case 'contents': {
                // Table of contents directive - skip rendering
                break;
            }
            case 'image': {
                const url = node.refuri || node.uri || node.url || '';
                const alt = node.alt || textOf(node) || '';
                if (url) {
                    lines.push(`![${alt}](${url})`);
                    lines.push('');
                }
                break;
            }
            case 'target': {
                // anchor; ignore, section handler will emit anchors for titles
                break;
            }
            case 'substitution_definition': {
                // capture: name + replacement content
                const key = node.name || node.substname || undefined;
                const value = textOf(node);
                if (key)
                    ctx.substitutions[key] = value;
                break;
            }
            case 'include': {
                // Data API often expands includes; if present, warn and render children best-effort
                const incArg = node.argument;
                const incPath = argToString(incArg) || String(node.refuri || node.uri || '').trim();
                const msg = `Include directive encountered${incPath ? `: ${incPath}` : ''} (rendering children best-effort)`;
                ctx.onWarn(msg, { path: ctx.docPath, position: posToStr(node.position) });
                for (const c of node.children || [])
                    block(c, depth);
                break;
            }
            case 'directive': {
                const name = String(node.name || '').toLowerCase();
                // Ignore certain non-content directives silently
                const ignoreDirectives = new Set([
                    'meta', 'facet', 'contents', 'toctree', 'kicker',
                    // Additional non-content/structural directives to ignore
                    'button', 'default-domain', 'cssclass', 'introduction', 'banner'
                ]);
                if (ignoreDirectives.has(name)) {
                    break;
                }
                // Treat include-like directives similarly (warn and render children best-effort)
                if (name === 'include' || name === 'literalinclude') {
                    const incArg = node.argument;
                    const incPath = argToString(incArg) || String(node.refuri || node.uri || '').trim();
                    const msg = `${name} directive encountered${incPath ? `: ${incPath}` : ''} (rendering children best-effort)`;
                    ctx.onWarn(msg, { path: ctx.docPath, position: posToStr(node.position) });
                    for (const c of node.children || [])
                        block(c, depth);
                    break;
                }
                // Container: render children as normal (no warning)
                if (name === 'container') {
                    for (const c of node.children || [])
                        block(c, depth);
                    break;
                }
                // Card groups and cards: keep the headline and text; ignore other options
                if (name === 'card-group') {
                    const cards = (node.children || []).filter((c) => c && c.type === 'directive' && String(c.name || '').toLowerCase() === 'card');
                    for (const card of cards) {
                        const opts = (card.options || {});
                        let title = String(opts.headline || '').trim();
                        if (!title) {
                            const arg = card.argument;
                            if (typeof arg === 'string')
                                title = arg.trim();
                            else if (Array.isArray(arg))
                                title = arg.map((a) => a?.value ?? '').join('').trim();
                        }
                        title = title || 'Card';
                        const level = Math.min(6, (lastHeadingLevel || 1) + 1);
                        const hashes = '#'.repeat(level);
                        lines.push(`${hashes} ${title}`);
                        lastHeadingLevel = level;
                        lines.push('');
                        for (const ch of (card.children || [])) {
                            if (ch && ch.type === 'paragraph')
                                block(ch, depth + 1);
                        }
                        lines.push('');
                    }
                    break;
                }
                if (name === 'card') {
                    const opts = (node.options || {});
                    let title = String(opts.headline || '').trim();
                    if (!title) {
                        const arg = node.argument;
                        if (typeof arg === 'string')
                            title = arg.trim();
                        else if (Array.isArray(arg))
                            title = arg.map((a) => a?.value ?? '').join('').trim();
                    }
                    title = title || 'Card';
                    const level = Math.min(6, (lastHeadingLevel || 1) + 1);
                    const hashes = '#'.repeat(level);
                    lines.push(`${hashes} ${title}`);
                    lastHeadingLevel = level;
                    lines.push('');
                    for (const ch of (node.children || [])) {
                        if (ch && ch.type === 'paragraph')
                            block(ch, depth + 1);
                    }
                    lines.push('');
                    break;
                }
                // Handle procedure/step directives
                if (name === 'procedure') {
                    // Prefer headings, not lists: render any non-step children first (intro), then for each step emit a heading
                    const steps = (node.children || []).filter((c) => c && c.type === 'directive' && String(c.name || '').toLowerCase() === 'step');
                    // Render any non-step children before the steps (intro text, etc.)
                    for (const c of (node.children || [])) {
                        if (!(c && c.type === 'directive' && String(c.name || '').toLowerCase() === 'step')) {
                            block(c, depth);
                        }
                    }
                    const norm = (s) => String(s || '').replace(/\s+/g, ' ').trim().toLowerCase();
                    if (steps.length) {
                        for (const st of steps) {
                            // Try to find an existing heading/title inside the step; if present, prefer it and do not emit our own
                            let headingChild = (st.children || []).find((c) => c && (c.type === 'heading' || c.type === 'title'));
                            // Compute a fallback title if no heading child exists
                            const argToString = (arg) => {
                                if (!arg)
                                    return '';
                                if (typeof arg === 'string')
                                    return arg.trim();
                                if (Array.isArray(arg))
                                    return arg.map((a) => a?.value ?? '').join('').trim();
                                return String(arg ?? '').trim();
                            };
                            const opts = (st.options || {});
                            let title = String(opts.title || '').trim();
                            if (!title)
                                title = argToString(st.argument);
                            if (!title) {
                                const firstPara = (st.children || []).find((c) => c && c.type === 'paragraph');
                                if (firstPara)
                                    title = (firstPara.children || []).map(inline).join('');
                            }
                            title = (title || 'Step').trim();
                            // If no direct heading/title, look for a nested section with a title/heading equal to the computed title
                            let sectionWithMatchingTitle;
                            if (!headingChild) {
                                sectionWithMatchingTitle = (st.children || []).find((c) => {
                                    if (!c || c.type !== 'section' || !Array.isArray(c.children))
                                        return false;
                                    const tOrH = c.children.find((cc) => cc && (cc.type === 'title' || cc.type === 'heading'));
                                    return tOrH ? norm(textOf(tOrH).trim()) === norm(title) : false;
                                });
                                if (sectionWithMatchingTitle) {
                                    headingChild = sectionWithMatchingTitle; // treat as existing heading container
                                }
                            }
                            if (!headingChild) {
                                // Emit a heading one level deeper than the last heading
                                const level = Math.min(6, (lastHeadingLevel || 1) + 1);
                                const hashes = '#'.repeat(level);
                                lines.push(`${hashes} ${title}`);
                                lastHeadingLevel = level;
                                lines.push('');
                            }
                            // Render step children. If we synthesized a heading, skip duplicates:
                            for (const ch of (st.children || [])) {
                                if (!headingChild) {
                                    if (ch.type === 'paragraph') {
                                        const paraText = (ch.children || []).map(inline).join('');
                                        if (norm(paraText) === norm(title))
                                            continue;
                                    }
                                    if (ch.type === 'title' || ch.type === 'heading') {
                                        const headingText = textOf(ch);
                                        if (norm(headingText) === norm(title))
                                            continue;
                                    }
                                    if (ch.type === 'section' && Array.isArray(ch.children)) {
                                        // Find section title/heading
                                        const tnode = ch.children.find((cc) => cc && (cc.type === 'title' || cc.type === 'heading'));
                                        const ttext = tnode ? textOf(tnode).trim() : '';
                                        if (norm(ttext) === norm(title)) {
                                            // Render the section without its title/heading to avoid duplicate heading
                                            const children = ch.children.filter((cc) => cc !== tnode);
                                            for (const cc of children)
                                                block(cc, depth + 2);
                                            continue;
                                        }
                                    }
                                }
                                block(ch, depth + 1);
                            }
                            lines.push('');
                        }
                        break;
                    }
                    // If no explicit steps, render children best-effort (already rendered above for non-step children)
                    break;
                }
                if (name === 'step') {
                    // Standalone step: prefer a heading over a list
                    let headingChild = (node.children || []).find((c) => c && (c.type === 'heading' || c.type === 'title'));
                    const argToString = (arg) => {
                        if (!arg)
                            return '';
                        if (typeof arg === 'string')
                            return arg.trim();
                        if (Array.isArray(arg))
                            return arg.map((a) => a?.value ?? '').join('').trim();
                        return String(arg ?? '').trim();
                    };
                    const opts = (node.options || {});
                    let title = String(opts.title || '').trim();
                    if (!title)
                        title = argToString(node.argument);
                    if (!title) {
                        const firstPara = (node.children || []).find((c) => c && c.type === 'paragraph');
                        if (firstPara)
                            title = (firstPara.children || []).map(inline).join('');
                    }
                    title = (title || 'Step').trim();
                    const norm = (s) => String(s || '').replace(/\s+/g, ' ').trim().toLowerCase();
                    // If no direct heading/title, see if a nested section contains a matching title/heading
                    let sectionWithMatchingTitle;
                    if (!headingChild) {
                        sectionWithMatchingTitle = (node.children || []).find((c) => {
                            if (!c || c.type !== 'section' || !Array.isArray(c.children))
                                return false;
                            const tOrH = c.children.find((cc) => cc && (cc.type === 'title' || cc.type === 'heading'));
                            return tOrH ? norm(textOf(tOrH).trim()) === norm(title) : false;
                        });
                        if (sectionWithMatchingTitle) {
                            headingChild = sectionWithMatchingTitle;
                        }
                    }
                    if (!headingChild) {
                        const level = Math.min(6, (lastHeadingLevel || 1) + 1);
                        const hashes = '#'.repeat(level);
                        lines.push(`${hashes} ${title}`);
                        lastHeadingLevel = level;
                        lines.push('');
                    }
                    for (const ch of (node.children || [])) {
                        if (!headingChild) {
                            if (ch.type === 'paragraph') {
                                const paraText = (ch.children || []).map(inline).join('');
                                if (norm(paraText) === norm(title))
                                    continue;
                            }
                            if (ch.type === 'title' || ch.type === 'heading') {
                                const headingText = textOf(ch);
                                if (norm(headingText) === norm(title))
                                    continue;
                            }
                            if (ch.type === 'section' && Array.isArray(ch.children)) {
                                const tnode = ch.children.find((cc) => cc && (cc.type === 'title' || cc.type === 'heading'));
                                const ttext = tnode ? textOf(tnode).trim() : '';
                                if (norm(ttext) === norm(title)) {
                                    const children = ch.children.filter((cc) => cc !== tnode);
                                    for (const cc of children)
                                        block(cc, depth + 2);
                                    continue;
                                }
                            }
                        }
                        block(ch, depth + 1);
                    }
                    lines.push('');
                    break;
                }
                // IO code block: treat nested input/output directives as fenced code blocks
                if (name === 'io-code-block') {
                    const parts = (node.children || []).filter((c) => c && c.type === 'directive' && ['input', 'output'].includes(String(c.name || '').toLowerCase()));
                    const renderCodeFrom = (d) => {
                        // Try to find nested code/literal_block child to preserve language
                        let lang = '';
                        let codeText = '';
                        const codeChild = (d.children || []).find((ch) => ch && (ch.type === 'literal_block' || ch.type === 'code' || ch.type === 'code_block'));
                        if (codeChild) {
                            lang = String(codeChild.language || codeChild.lang || '').trim();
                            codeText = textOf(codeChild);
                        }
                        else {
                            // fallback to concatenated text
                            lang = String((d.options || {}).language || '').trim();
                            codeText = (d.children || []).map((ch) => (ch.type === 'paragraph' ? (ch.children || []).map(inline).join('') : textOf(ch))).join('\n');
                        }
                        lines.push('```' + lang);
                        for (const l of String(codeText).split(/\r?\n/)) {
                            lines.push(l);
                        }
                        lines.push('```');
                        lines.push('');
                    };
                    for (const p of parts)
                        renderCodeFrom(p);
                    break;
                }
                if (name === 'input' || name === 'output') {
                    // Standalone input/output directive → fenced code block
                    let lang = String((node.options || {}).language || '').trim();
                    const codeChild = (node.children || []).find((ch) => ch && (ch.type === 'literal_block' || ch.type === 'code' || ch.type === 'code_block'));
                    const codeText = codeChild ? textOf(codeChild) : (node.children || []).map((ch) => (ch.type === 'paragraph' ? (ch.children || []).map(inline).join('') : textOf(ch))).join('\n');
                    lines.push('```' + lang);
                    for (const l of String(codeText).split(/\r?\n/)) {
                        lines.push(l);
                    }
                    lines.push('```');
                    lines.push('');
                    break;
                }
                // Handle figure/image directives → render as image with optional caption
                if (name === 'figure' || name === 'image') {
                    const n = node;
                    const opts = (n.options || {});
                    // Derive src from argument, standard link fields, or nested image child
                    let src = String(n.refuri || n.uri || n.url || '').trim();
                    const argToString = (arg) => {
                        if (!arg)
                            return '';
                        if (typeof arg === 'string')
                            return arg;
                        if (Array.isArray(arg))
                            return arg.map((a) => a?.value ?? '').join('').trim();
                        return String(arg ?? '');
                    };
                    if (!src)
                        src = argToString(n.argument);
                    // Some ASTs nest an image node inside the figure directive
                    let imageChild = (n.children || []).find((c) => c && c.type === 'image');
                    if (!src && imageChild) {
                        src = String(imageChild.refuri || imageChild.uri || imageChild.url || '').trim();
                    }
                    // Caption is often the first paragraph child (exclude image child if present)
                    const firstPara = (n.children || []).find((c) => c && c.type === 'paragraph');
                    const captionText = firstPara ? (firstPara.children || []).map(inline).join('') : '';
                    let alt = String(opts.alt || '').trim();
                    if (!alt && imageChild) {
                        alt = String(imageChild.alt || textOf(imageChild) || '').trim();
                    }
                    if (!alt)
                        alt = String(captionText || '').trim();
                    if (src) {
                        // Force Markdown image output (ignore width/height attributes)
                        lines.push(`![${alt}](${src})`);
                        lines.push('');
                        if (captionText && captionText !== alt) {
                            lines.push(`_${captionText}_`);
                            lines.push('');
                        }
                        break;
                    }
                    ctx.onWarn(`Figure/Image directive missing src; rendering children best-effort`, { path: ctx.docPath, position: posToStr(node.position) });
                    // If no src found, fall through to children rendering
                }
                // Handle RST list-table directive by emitting a GFM table
                if (name === 'list-table' || name === 'list_table') {
                    const opts = (node.options || {});
                    const headerRowsOpt = opts?.['header-rows'] ?? opts?.['header_rows'] ?? opts?.['headerrows'];
                    const headerRows = Number.isFinite(headerRowsOpt) ? Number(headerRowsOpt) : parseInt(String(headerRowsOpt ?? '1'), 10);
                    // If a concrete table child exists, reuse the generic table renderer
                    const tableChild = (node.children || []).find((c) => c && c.type === 'table');
                    if (tableChild) {
                        renderTable(tableChild);
                        break;
                    }
                    // Otherwise, attempt to parse the list-based structure of list-table
                    const escapePipes = (s) => s.replace(/\|/g, '\\|');
                    const cellText = (n) => {
                        if (!n)
                            return '';
                        if (n.type === 'paragraph')
                            return (n.children || []).map(inline).join('');
                        if (n.type === 'literal_block' || n.type === 'code') {
                            const code = textOf(n).trim();
                            // Represent multiline code inside a table cell as inline code per line, joined with <br>
                            return code.split(/\r?\n/).map((l) => '`' + escapePipes(l) + '`').join('<br>');
                        }
                        if (Array.isArray(n.children) && n.children.length)
                            return n.children.map(cellText).join(' ').trim();
                        return escapePipes(textOf(n));
                    };
                    const listNode = (node.children || []).find((c) => c && (c.type === 'list' || c.type === 'bullet_list'));
                    const rows = [];
                    if (listNode && Array.isArray(listNode.children)) {
                        for (const rowItem of listNode.children) {
                            if (!rowItem)
                                continue;
                            const cellList = (rowItem.children || []).find((c) => c && (c.type === 'list' || c.type === 'bullet_list'));
                            if (cellList && Array.isArray(cellList.children) && cellList.children.length) {
                                const cells = [];
                                for (const cellItem of cellList.children) {
                                    const parts = [];
                                    for (const cc of (cellItem.children || [])) {
                                        if (!cc)
                                            continue;
                                        if (cc.type === 'paragraph')
                                            parts.push((cc.children || []).map(inline).join(''));
                                        else if (cc.type === 'literal_block' || cc.type === 'code')
                                            parts.push(cellText(cc));
                                        else
                                            parts.push(cellText(cc));
                                    }
                                    const text = escapePipes(parts.join(' ').replace(/\s+/g, ' ').trim());
                                    cells.push(text);
                                }
                                rows.push(cells);
                            }
                        }
                    }
                    if (rows.length) {
                        const hasHeader = Number.isFinite(headerRows) ? headerRows > 0 : true;
                        let header = rows[0] || [];
                        if (hasHeader) {
                            rows.shift();
                        }
                        else {
                            // Synthesize an empty header with the same number of columns
                            header = header.map(() => '');
                        }
                        const sep = '|' + header.map(() => ' --- ').join('|') + '|';
                        const fmt = (r) => '|' + header.map((_, i) => (r[i] ?? '')).join('|') + '|';
                        lines.push('|' + header.join('|') + '|');
                        lines.push(sep);
                        for (const r of rows)
                            lines.push(fmt(r));
                        lines.push('');
                        break;
                    }
                    ctx.onWarn(`list-table directive could not be parsed; rendering children best-effort`, { path: ctx.docPath, position: posToStr(node.position) });
                    // Fall through to default children rendering if structure wasn't recognized
                }
                // Handle tabbed content containers like "tabs" or "tabs-*"
                if (name === 'tabs' || name.startsWith('tabs-')) {
                    const tabs = (node.children || []).filter((c) => c && (c.type === 'directive') && String(c.name || '').toLowerCase() === 'tab');
                    const toTitle = (s) => s
                        .split(/[-_\s]+/)
                        .map(part => part ? part.charAt(0).toUpperCase() + part.slice(1) : '')
                        .join(' ')
                        .replace(/C\b/, 'C'); // keep Objective-C capitalized properly when split
                    for (const tab of tabs) {
                        const opts = (tab.options || {});
                        let label = String(opts.tabid || opts.label || '').trim();
                        if (!label && Array.isArray(tab.argument)) {
                            label = tab.argument.map((a) => a.value || '').join('').trim();
                        }
                        if (!label && typeof tab.argument === 'string') {
                            label = String(tab.argument).trim();
                        }
                        label = label || 'Tab';
                        const heading = toTitle(label).replace(/Objective C/i, 'Objective-C');
                        // Emit a sub-heading for the tab, then the tab content
                        const hashes = '####';
                        lines.push(`${hashes} ${heading}`);
                        lastHeadingLevel = 4;
                        lines.push('');
                        for (const c of (tab.children || []))
                            block(c, depth + 1);
                        lines.push('');
                    }
                    break;
                }
                // Handle a standalone tab directive defensively
                if (name === 'tab') {
                    const opts = (node.options || {});
                    let label = String(opts.tabid || opts.label || '').trim();
                    if (!label && Array.isArray(node.argument)) {
                        label = node.argument.map((a) => a.value || '').join('').trim();
                    }
                    if (!label && typeof node.argument === 'string') {
                        label = String(node.argument).trim();
                    }
                    const toTitle = (s) => s
                        .split(/[-_\s]+/)
                        .map(part => part ? part.charAt(0).toUpperCase() + part.slice(1) : '')
                        .join(' ');
                    const heading = toTitle(label || 'Tab').replace(/Objective C/i, 'Objective-C');
                    const hashes = '####';
                    lines.push(`${hashes} ${heading}`);
                    lastHeadingLevel = 4;
                    lines.push('');
                    for (const c of (node.children || []))
                        block(c, depth + 1);
                    lines.push('');
                    break;
                }
                // Treat common admonition directives (e.g., .. note::, .. important::)
                const admonitions = new Set(['note', 'warning', 'tip', 'important', 'caution', 'danger', 'seealso', 'example', 'see', 'versionadded', 'versionchanged', 'deprecated']);
                if (admonitions.has(name)) {
                    renderAdmonition(node, name, depth);
                    break;
                }
                // Warn for unhandled directive types before falling back
                if (!(new Set(['figure', 'image', 'list-table', 'list_table', 'tabs', 'tab', 'note', 'warning', 'tip', 'important', 'caution', 'danger', 'seealso', 'banner', 'example', 'include', 'literalinclude', 'procedure', 'step', 'io-code-block', 'input', 'output', 'see', 'versionadded', 'versionchanged', 'deprecated', 'button', 'default-domain', 'cssclass', 'introduction', 'container', 'card', 'card-group']).has(name))) {
                    ctx.onWarn(`Unhandled directive .. ${name}:: (rendering children best-effort)`, { path: ctx.docPath, position: posToStr(node.position) });
                }
                // Fallback: render children
                for (const c of node.children || [])
                    block(c, depth);
                break;
            }
            default: {
                if (node.children && node.children.length) {
                    for (const c of node.children)
                        block(c, depth);
                }
            }
        }
    }
    // Some ASTs wrap into a top-level document node
    if (root && root.type === 'root' && Array.isArray(root.children)) {
        for (const c of root.children)
            block(c, 1);
    }
    else if (root && Array.isArray(root.children)) {
        for (const c of root.children)
            block(c, 1);
    }
    else if (root) {
        block(root, 1);
    }
    // Post-process (disabled): previously converted certain paragraph sequences into bullet lists.
    // This heuristic caused numbered lists to become unordered. We now return lines as-is.
    return lines.join(EOL).replace(/\n{3,}/g, '\n\n');
}
