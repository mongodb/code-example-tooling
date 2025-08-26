/**
 * References artifact helpers
 */

import fs from 'fs';

export interface ReferencesArtifact {
  substitutions: Record<string, string>;
  refs: Record<string, { title: string; url: string }>;
};

export const buildReferencesTs = (artifact: ReferencesArtifact): string => {
  const substitutions = artifact.substitutions || {};
  const refs = artifact.refs || {};
  const esc = (s: string) => {
    const toString = (v: any) => (typeof v === 'string' ? v : String(v ?? ''));
    let str = toString(s);
    const MAX = 1000;
    if (str.length > MAX) str = str.slice(0, MAX) + '…';
    str = str
      .replace(/\\/g, '\\\\')
      .replace(/'/g, "\\'")
      .replace(/\r/g, '\\r')
      .replace(/\n/g, '\\n')
      .replace(/\t/g, '\\t');
    return `'${str}'`;
  };

  const subsLines = Object.entries(substitutions)
    .sort(([a],[b]) => a.localeCompare(b))
    .map(([k, v]) => `    ${JSON.stringify(k)}: ${esc(v)},`)
    .join('\n');

  const refsLines = Object.entries(refs)
    .sort(([a],[b]) => a.localeCompare(b))
    .map(([url, { title }]) => `    ${esc(url)}: { title: ${esc(title)}, url: ${esc(url)} },`)
    .join('\n');

  return `export const substitutions = {\n${subsLines}\n} as const;\n` +
`export const refs = {\n${refsLines}\n} as const;\n` +
`const references = { substitutions, refs } as const;\nexport default references;\n`;
};

export const readExistingReferences = (filePath: string): ReferencesArtifact => {
  try {
    const text = fs.readFileSync(filePath, 'utf8');
    const result: ReferencesArtifact = { substitutions: {}, refs: {} };

    const decodeLiteral = (s: string): string => {
      return s
        .replace(/\\r/g, '\r')
        .replace(/\\n/g, '\n')
        .replace(/\\t/g, '\t')
        .replace(/\\'/g, "'")
        .replace(/\\\"/g, '"')
        .replace(/\\\\/g, '\\');
    };

    const parseSubsBody = (body: string) => {
      const re = /(["'])([^"'\\]*(?:\\.[^"'\\]*)*)\1\s*:\s*(["'])([^"'\\]*(?:\\.[^"'\\]*)*)\3\s*,?/g;
      let m: RegExpExecArray | null;
      while ((m = re.exec(body)) !== null) {
        const key = decodeLiteral(m[2]);
        const val = decodeLiteral(m[4]);
        result.substitutions[key] = val;
      }
    };

    const parseRefsBody = (body: string) => {
      const re = /(["'])([^"'\\]*(?:\\.[^"'\\]*)*)\1\s*:\s*\{\s*title:\s*(["'])([^"'\\]*(?:\\.[^"'\\]*)*)\3\s*,\s*url:\s*(["'])([^"'\\]*(?:\\.[^"'\\]*)*)\5\s*\}\s*,?/g;
      let m: RegExpExecArray | null;
      while ((m = re.exec(body)) !== null) {
        const urlKey = decodeLiteral(m[2]);
        const title = decodeLiteral(m[4]);
        const url = decodeLiteral(m[6]);
        result.refs[urlKey] = { title, url };
      }
    };

    const subsMatchNamed = text.match(/export\s+const\s+substitutions\s*=\s*\{([\s\S]*?)\}\s*as\s+const/);
    const subsMatchDefault = text.match(/substitutions\s*:\s*\{([\s\S]*?)\}\s*,/);
    const subsMatch = subsMatchNamed || subsMatchDefault;
    if (subsMatch) parseSubsBody(subsMatch[1]);

    const refsMatchNamed = text.match(/export\s+const\s+refs\s*=\s*\{([\s\S]*?)\}\s*as\s+const/);
    const refsMatchDefault = text.match(/refs\s*:\s*\{([\s\S]*?)\}\s*\n\s*\}/);
    const refsMatch = refsMatchNamed || refsMatchDefault;
    if (refsMatch) parseRefsBody(refsMatch[1]);

    return result;
  } catch {
    return { substitutions: {}, refs: {} };
  }
};

export const mergeReferences = (base: ReferencesArtifact, add: ReferencesArtifact): ReferencesArtifact => {
  const outSubs: Record<string, string> = { ...base.substitutions };
  for (const [k, v] of Object.entries(add.substitutions || {})) {
    outSubs[k] = v as string;
  }
  const outRefs: Record<string, { title: string; url: string }> = { ...base.refs };
  for (const [url, obj] of Object.entries(add.refs || {})) {
    outRefs[url] = obj as any;
  }
  return { substitutions: outSubs, refs: outRefs };
};
