// src/ast-to-md.ts
// Very lightweight Snooty AST -> Markdown converter with pragmatic support for
// - sections/titles -> #/## headers (with optional HTML anchors when ids/html_id present)
// - paragraphs and text
// - code/literal blocks
// - links and references
// - substitutions (requires a map)
// - basic admonitions -> blockquotes with a label
// - images -> Markdown image syntax
// - simple tables -> GFM pipe tables (best-effort)
// - includes -> warn but render children if present

export interface AstToMdOptions {
  // Global substitutions map for the project
  substitutions?: Record<string, string>;
  // Called when we cannot resolve a node (include, ref, substitution, etc.)
  onWarn?: (message: string, context?: { path?: string; position?: string }) => void;
  // Current document path for logging
  docPath?: string;
}

interface AstNode {
  type?: string;
  value?: string;
  name?: string; // for substitutions
  refuri?: string; // for links/images
  refname?: string; // for :ref:
  uri?: string; // image/links
  url?: string; // image/links
  alt?: string; // image alt
  children?: AstNode[];
  position?: any; // arbitrary
  language?: string; // for code blocks
  lang?: string; // alternative property for code language
  ids?: string[]; // targets
  html_id?: string; // optional html id
  admonition_type?: string; // for admonitions
  depth?: number; // for heading nodes
}

const EOL = "\n";

export function astToMarkdown(root: any, options: AstToMdOptions = {}): string {
  const ctx: Required<AstToMdOptions> = {
    substitutions: options.substitutions || {},
    onWarn: options.onWarn || (() => {}),
    docPath: options.docPath || '',
  };

  const lines: string[] = [];

  function posToStr(pos: any): string | undefined {
    if (!pos) return undefined;
    try {
      // common position shapes: { start: { line, column }, end: { line, column } }
      const s = pos.start ? `${pos.start.line}:${pos.start.column}` : '';
      return s || undefined;
    } catch {
      return undefined;
    }
  }

  function textOf(node: any): string {
    if (!node) return '';
    if (typeof node === 'string') return node;
    if (node.value) return String(node.value);
    if (Array.isArray(node.children)) return node.children.map(textOf).join('');
    return '';
  }

  function inline(node: AstNode): string {
    switch (node.type) {
      case 'text':
        return node.value || '';
      case 'literal':
      case 'inline_literal':
        return '`' + textOf(node) + '`';
      case 'emphasis':
        return '*' + (node.children ? node.children.map(inline).join('') : textOf(node)) + '*';
      case 'strong':
      case 'strong_emphasis':
        return '**' + (node.children ? node.children.map(inline).join('') : textOf(node)) + '**';
      case 'substitution_reference': {
        const key = node.name || node.value || textOf(node);
        const val = ctx.substitutions[key!];
        if (val == null) {
          ctx.onWarn(`Unresolved substitution |${key}|`, { path: ctx.docPath, position: posToStr(node.position) });
          return `|${key}|`;
        }
        return String(val);
      }
      case 'reference': {
        // Links can be refuri (URL) or refname (internal)
        const label = textOf(node);
        if (node.refuri) return `[${label}](${node.refuri})`;
        if (node.refname) return label; // best-effort inline text for :ref:
        return label;
      }
      case 'image': {
        const url = node.refuri || node.uri || node.url || '';
        const alt = node.alt || textOf(node) || '';
        if (!url) return alt;
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

  function pushAnchor(id?: string) {
    if (id) {
      lines.push(`<a id="${id}"></a>`);
      lines.push('');
    }
  }

  function renderTable(t: AstNode) {
    // Best-effort parse of Snooty table -> GFM table
    // Look for thead/tbody rows; otherwise collect any row/entry structure in order
    const headers: string[][] = [];
    const rows: string[][] = [];

    function cellsOf(node: any): string[] {
      if (!node) return [];
      if (Array.isArray(node.children)) {
        const entryNodes = node.children.filter((c: any) => c && (c.type === 'entry' || c.type === 'cell'));
        if (entryNodes.length) {
          return entryNodes.map((e: any) => (Array.isArray(e.children) ? e.children.map((ch: any) => (ch.type === 'paragraph' ? (ch.children || []).map(inline).join('') : inline(ch as any))).join(' ') : textOf(e)));
        }
      }
      return [];
    }

    function walk(node: any, inHead = false) {
      if (!node) return;
      const type = node.type;
      if (type === 'thead' || type === 'tgrouphead') {
        for (const r of node.children || []) {
          if (r.type === 'row') headers.push(cellsOf(r));
        }
        return;
      }
      if (type === 'tbody' || type === 'tgroupbody') {
        for (const r of node.children || []) {
          if (r.type === 'row') rows.push(cellsOf(r));
        }
        return;
      }
      if (type === 'row') {
        const cells = cellsOf(node);
        if (inHead) headers.push(cells); else rows.push(cells);
        return;
      }
      if (Array.isArray(node.children)) {
        for (const c of node.children) walk(c, inHead || type === 'thead');
      }
    }

    walk(t);
    const header = headers[0] || rows.shift() || [];
    if (!header.length) return; // nothing to render

    const sep = '|' + header.map(() => ' --- ').join('|') + '|';
    const fmt = (r: string[]) => '|' + header.map((_, i) => (r[i] ?? '').replace(/\n/g, ' ')).join('|') + '|';

    lines.push(fmt(header));
    lines.push(sep);
    for (const r of rows) lines.push(fmt(r));
    lines.push('');
  }

  function renderAdmonition(node: AstNode, kind: string) {
    const label = kind.charAt(0).toUpperCase() + kind.slice(1).toLowerCase();
    lines.push(`> ${label}:`);
    for (const c of node.children || []) {
      if (c.type === 'paragraph') {
        const txt = (c.children || []).map(inline).join('');
        lines.push('> ' + txt);
      } else if (c.type === 'literal_block') {
        const language = c.language ? String(c.language) : '';
        const code = textOf(c);
        lines.push('>');
        lines.push('> ' + '```' + language);
        for (const l of code.split(/\r?\n/)) lines.push('> ' + l);
        lines.push('> ' + '```');
      }
    }
    lines.push('');
  }

  function block(node: AstNode, depth: number) {
    switch (node.type) {
      case 'section': {
        // Find first title child for header
        const titleNode = (node.children || []).find((c) => c.type === 'title');
        if (titleNode) {
          // Bump headings up one level (min 1)
          const level = Math.max(1, Math.min(6, depth - 1));
          const hashes = '#'.repeat(level);
          lines.push(`${hashes} ${textOf(titleNode)}`);
          // Anchor from ids/html_id if present
          const anchorId = (titleNode as any).html_id || (titleNode as any).ids?.[0] || (node as any).html_id || (node as any).ids?.[0];
          pushAnchor(anchorId);
        }
        for (const c of node.children || []) {
          if (c === titleNode) continue;
          block(c, depth + 1);
        }
        break;
      }
      case 'title': {
        // Standalone page title (not nested under section)
        const level = Math.max(1, Math.min(6, depth - 1));
        const hashes = '#'.repeat(level);
        lines.push(`${hashes} ${textOf(node)}`);
        const anchorId = (node as any).html_id || (node as any).ids?.[0];
        pushAnchor(anchorId);
        break;
      }
      case 'heading': {
        const level0 = Math.max(1, Math.min(6, (node.depth as number) || depth || 1));
        const level = Math.max(1, level0 - 1);
        const hashes = '#'.repeat(level);
        lines.push(`${hashes} ${textOf(node)}`);
        const anchorId = (node as any).html_id || (node as any).ids?.[0];
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
        const etRaw = (node as any).enumtype;
        const et = typeof etRaw === 'string' ? etRaw.toLowerCase() : undefined;
        const likelyOrderedTypes = new Set(['ordered','arabic','loweralpha','upperalpha','lowerroman','upperroman','alphabetical','roman']);
        let ordered = false;
        if (et && likelyOrderedTypes.has(et)) ordered = true;
        if (!ordered && (node as any).ordered === true) ordered = true;
        if (!ordered && (typeof (node as any).start === 'number' || typeof (node as any).startat === 'number')) ordered = true;

        const startAt = typeof (node as any).startat === 'number' ? (node as any).startat
          : (typeof (node as any).start === 'number' ? (node as any).start : 1);
        if (ordered) {
          let i = Math.max(1, startAt || 1);
          for (const item of node.children || []) {
            const txt = (item.children || []).map((n) => (n.type === 'paragraph' ? (n.children || []).map(inline).join('') : inline(n))).join(' ');
            lines.push(`${i}. ${txt}`);
            i++;
          }
        } else {
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
        lines.push(code);
        lines.push('```');
        lines.push('');
        break;
      }
      case 'table': {
        renderTable(node);
        break;
      }
      case 'admonition': {
        const kind = (node.admonition_type as string) || 'note';
        renderAdmonition(node, kind);
        break;
      }
      case 'note':
      case 'warning':
      case 'tip':
      case 'important':
      case 'caution':
      case 'seealso': {
        renderAdmonition(node, String(node.type));
        break;
      }
      case 'versionadded': {
        const added = textOf(node) || (node.children || []).map(inline).join('');
        if (added) {
          lines.push(`> Added in ${added}`);
          lines.push('');
        }
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
        const key = node.name || (node as any).substname || undefined;
        const value = textOf(node);
        if (key) ctx.substitutions[key] = value;
        break;
      }
      case 'include': {
        // Data API often expands includes; if present, warn and render children best-effort
        ctx.onWarn(`Include directive encountered (rendering children best-effort)`, { path: ctx.docPath, position: posToStr(node.position) });
        for (const c of node.children || []) block(c, depth);
        break;
      }
      case 'directive': {
        const name = String((node as any).name || '').toLowerCase();

        // Handle figure/image directives → render as image with optional caption
        if (name === 'figure' || name === 'image') {
          const n: any = node as any;
          const opts = (n.options || {}) as any;
          // Derive src from argument, standard link fields, or nested image child
          let src = String(n.refuri || n.uri || n.url || '').trim();
          const argToString = (arg: any): string => {
            if (!arg) return '';
            if (typeof arg === 'string') return arg;
            if (Array.isArray(arg)) return arg.map((a: any) => a?.value ?? '').join('').trim();
            return String(arg ?? '');
          };
          if (!src) src = argToString(n.argument);

          // Some ASTs nest an image node inside the figure directive
          let imageChild = (n.children || []).find((c: any) => c && c.type === 'image');
          if (!src && imageChild) {
            src = String((imageChild as any).refuri || (imageChild as any).uri || (imageChild as any).url || '').trim();
          }

          // Caption is often the first paragraph child (exclude image child if present)
          const firstPara = (n.children || []).find((c: any) => c && c.type === 'paragraph');
          const captionText = firstPara ? (firstPara.children || []).map(inline).join('') : '';
          let alt = String(opts.alt || '').trim();
          if (!alt && imageChild) {
            alt = String((imageChild as any).alt || textOf(imageChild) || '').trim();
          }
          if (!alt) alt = String(captionText || '').trim();

          const width = opts.width != null ? String(opts.width).trim() : '';
          const height = opts.height != null ? String(opts.height).trim() : '';

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
          // If no src found, fall through to children rendering
        }

        // Handle RST list-table directive by emitting a GFM table
        if (name === 'list-table' || name === 'list_table') {
          const opts = ((node as any).options || {}) as any;
          const headerRowsOpt = opts?.['header-rows'] ?? opts?.['header_rows'] ?? opts?.['headerrows'];
          const headerRows = Number.isFinite(headerRowsOpt) ? Number(headerRowsOpt) : parseInt(String(headerRowsOpt ?? '1'), 10);

          // If a concrete table child exists, reuse the generic table renderer
          const tableChild = (node.children || []).find((c: any) => c && c.type === 'table');
          if (tableChild) {
            renderTable(tableChild as any);
            break;
          }

          // Otherwise, attempt to parse the list-based structure of list-table
          const escapePipes = (s: string) => s.replace(/\|/g, '\\|');
          const cellText = (n: any): string => {
            if (!n) return '';
            if (n.type === 'paragraph') return (n.children || []).map(inline).join('');
            if (n.type === 'literal_block' || n.type === 'code') {
              const code = textOf(n).trim();
              // Represent multiline code inside a table cell as inline code per line, joined with <br>
              return code.split(/\r?\n/).map((l) => '`' + escapePipes(l) + '`').join('<br>');
            }
            if (Array.isArray(n.children) && n.children.length) return n.children.map(cellText).join(' ').trim();
            return escapePipes(textOf(n));
          };

          const listNode = (node.children || []).find((c: any) => c && (c.type === 'list' || c.type === 'bullet_list')) as any;
          const rows: string[][] = [];
          if (listNode && Array.isArray(listNode.children)) {
            for (const rowItem of listNode.children) {
              if (!rowItem) continue;
              const cellList = (rowItem.children || []).find((c: any) => c && (c.type === 'list' || c.type === 'bullet_list')) as any;
              if (cellList && Array.isArray(cellList.children) && cellList.children.length) {
                const cells: string[] = [];
                for (const cellItem of cellList.children) {
                  const parts: string[] = [];
                  for (const cc of (cellItem.children || [])) {
                    if (!cc) continue;
                    if (cc.type === 'paragraph') parts.push((cc.children || []).map(inline).join(''));
                    else if (cc.type === 'literal_block' || cc.type === 'code') parts.push(cellText(cc));
                    else parts.push(cellText(cc));
                  }
                  const text = escapePipes(parts.join(' ').replace(/\s+/g, ' ').trim());
                  cells.push(text);
                }
                rows.push(cells);
              }
            }
          }

          if (rows.length) {
            const hasHeader = Number.isFinite(headerRows) ? (headerRows as number) > 0 : true;
            let header = rows[0] || [];
            if (hasHeader) {
              rows.shift();
            } else {
              // Synthesize an empty header with the same number of columns
              header = header.map(() => '');
            }
            const sep = '|' + header.map(() => ' --- ').join('|') + '|';
            const fmt = (r: string[]) => '|' + header.map((_, i) => (r[i] ?? '')).join('|') + '|';

            lines.push('|' + header.join('|') + '|');
            lines.push(sep);
            for (const r of rows) lines.push(fmt(r));
            lines.push('');
            break;
          }
          // Fall through to default children rendering if structure wasn't recognized
        }

        // Handle tabbed content containers like "tabs" or "tabs-*"
        if (name === 'tabs' || name.startsWith('tabs-')) {
          const tabs = (node.children || []).filter((c: any) => c && (c.type === 'directive') && String((c as any).name || '').toLowerCase() === 'tab');
          const toTitle = (s: string) => s
            .split(/[-_\s]+/)
            .map(part => part ? part.charAt(0).toUpperCase() + part.slice(1) : '')
            .join(' ')
            .replace(/C\b/,'C'); // keep Objective-C capitalized properly when split
          for (const tab of tabs as any[]) {
            const opts = (tab.options || {}) as any;
            let label = String(opts.tabid || opts.label || '').trim();
            if (!label && Array.isArray((tab as any).argument)) {
              label = (tab as any).argument.map((a: any) => a.value || '').join('').trim();
            }
            if (!label && typeof (tab as any).argument === 'string') {
              label = String((tab as any).argument).trim();
            }
            label = label || 'Tab';
            const heading = toTitle(label).replace(/Objective C/i, 'Objective-C');
            // Emit a sub-heading for the tab, then the tab content
            const hashes = '####';
            lines.push(`${hashes} ${heading}`);
            lines.push('');
            for (const c of (tab.children || [])) block(c as any, depth + 1);
            lines.push('');
          }
          break;
        }

        // Handle a standalone tab directive defensively
        if (name === 'tab') {
          const opts = ((node as any).options || {}) as any;
          let label = String(opts.tabid || opts.label || '').trim();
          if (!label && Array.isArray((node as any).argument)) {
            label = (node as any).argument.map((a: any) => a.value || '').join('').trim();
          }
          if (!label && typeof (node as any).argument === 'string') {
            label = String((node as any).argument).trim();
          }
          const toTitle = (s: string) => s
            .split(/[-_\s]+/)
            .map(part => part ? part.charAt(0).toUpperCase() + part.slice(1) : '')
            .join(' ');
          const heading = toTitle(label || 'Tab').replace(/Objective C/i, 'Objective-C');
          const hashes = '####';
          lines.push(`${hashes} ${heading}`);
          lines.push('');
          for (const c of (node.children || [])) block(c as any, depth + 1);
          lines.push('');
          break;
        }

        // Treat common admonition directives (e.g., .. note::, .. important::)
        const admonitions = new Set(['note','warning','tip','important','caution','danger','seealso']);
        if (admonitions.has(name)) {
          renderAdmonition(node, name);
          break;
        }
        // Fallback: render children
        for (const c of node.children || []) block(c, depth);
        break;
      }
      default: {
        if (node.children && node.children.length) {
          for (const c of node.children) block(c, depth);
        }
      }
    }
  }

  // Some ASTs wrap into a top-level document node
  if (root && root.type === 'root' && Array.isArray(root.children)) {
    for (const c of root.children) block(c as AstNode, 1);
  } else if (root && Array.isArray(root.children)) {
    for (const c of root.children) block(c as AstNode, 1);
  } else if (root) {
    block(root as AstNode, 1);
  }

  // Post-process (disabled): previously converted certain paragraph sequences into bullet lists.
  // This heuristic caused numbered lists to become unordered. We now return lines as-is.
  return lines.join(EOL).replace(/\n{3,}/g, '\n\n');
}
