// src/converters/snooty.ts
import * as fs from 'fs';
import * as path from 'path';

interface ParseOptions {
  filePath: string;
  basePath: string;
  resolveIncludes: boolean;
  resolveSubstitutions: boolean;
  resolveRefs: boolean;
}

interface ParsedContent {
  content: string;
  includes: Record<string, string>;
  substitutions: Record<string, string>;
  refs: Record<string, string>;
}

async function parseSnootyContent(content: string, options: ParseOptions): Promise<ParsedContent> {
  let processedContent = content;
  const includes: Record<string, string> = {};
  const substitutions: Record<string, string> = {};
  const refs: Record<string, string> = {};

  // Handle includes
  if (options.resolveIncludes) {
    processedContent = await processIncludes(processedContent, options, includes);
  }

  // Handle substitutions
  if (options.resolveSubstitutions) {
    processedContent = processSubstitutions(processedContent, substitutions);
  }

  // Handle refs
  if (options.resolveRefs) {
    processedContent = processRefs(processedContent, refs);
  }

  return {
    content: processedContent,
    includes,
    substitutions,
    refs
  };
}

async function processIncludes(content: string, options: ParseOptions, includes: Record<string, string>): Promise<string> {
  // Match pattern like .. include:: /path/to/file (Snooty paths beginning with '/' are project-root relative)
  const includeRegex = /\.\. include:: (.*?)$/gm;
  let match;
  let result = content;

  while ((match = includeRegex.exec(content)) !== null) {
    const includePath = match[1].trim();
    const fullPath = includePath.startsWith('/')
      ? path.join(options.basePath, includePath)
      : path.join(path.dirname(options.filePath), includePath);

    try {
      const includeContent = fs.readFileSync(fullPath, 'utf8');
      // Store for reference
      includes[includePath] = includeContent;

      // Replace include directive with actual content
      result = result.replace(match[0], includeContent);
    } catch (error) {
      console.warn(`Could not process include: ${includePath} (resolved to ${fullPath})`, error);
    }
  }

  return result;
}

function processSubstitutions(content: string, substitutions: Record<string, string>): string {
  // Match pattern like |substitution|
  const substitutionDefRegex = /\.\. \|([^|]+)\| replace:: (.*?)$/gm;
  const substitutionUseRegex = /\|([^|]+)\|/g;

  // Extract substitution definitions
  let match;
  while ((match = substitutionDefRegex.exec(content)) !== null) {
    substitutions[match[1]] = match[2];
  }

  // Apply substitutions
  let result = content;
  Object.entries(substitutions).forEach(([key, value]) => {
    result = result.replace(new RegExp(`\\|${key}\\|`, 'g'), value);
  });

  return result;
}

function processRefs(content: string, refs: Record<string, string>): string {
  // Handle :ref:`text <label>` and :ref:`label` by inlining the visible text
  const refWithTextRegex = /:ref:`([^<`]*)<([^>`]*)>`/g;
  const refSimpleRegex = /:ref:`([^<`][^`>]*)`/g;

  let result = content;

  // Replace refs with explicit text
  result = result.replace(refWithTextRegex, (_m, text, label) => {
    const t = String(text).trim();
    const l = String(label).trim();
    refs[l] = t;
    return t;
  });

  // Replace simple refs by their label text (best effort)
  result = result.replace(refSimpleRegex, (_m, label) => {
    const l = String(label).trim();
    refs[l] = l;
    return l;
  });

  return result;
}

function convertToMarkdown(parsedContent: ParsedContent): string {
  let markdown = parsedContent.content;

  // Convert RST headers to Markdown headers
  markdown = convertHeaders(markdown);

  // Convert RST links to Markdown links
  markdown = convertLinks(markdown);

  // Convert RST code blocks to Markdown code blocks
  markdown = convertCodeBlocks(markdown);

  // Convert RST lists to Markdown lists
  markdown = convertLists(markdown);

  return markdown;
}

function convertHeaders(content: string): string {
  // Convert underlined headers (=== and ---) to # and ##
  let lines = content.split('\n');
  const result = [];

  for (let i = 0; i < lines.length; i++) {
    const line = lines[i];
    const nextLine = i < lines.length - 1 ? lines[i + 1] : '';

    if (nextLine && nextLine.match(/^=+$/)) {
      result.push(`# ${line}`);
      i++; // Skip the === line
    } else if (nextLine && nextLine.match(/^-+$/)) {
      result.push(`## ${line}`);
      i++; // Skip the --- line
    } else {
      result.push(line);
    }
  }

  return result.join('\n');
}

function convertLinks(content: string): string {
  // Convert RST links `Link text <http://example.com>`_ to Markdown [Link text](http://example.com)
  return content.replace(/`([^<]*)<([^>]*)>`_/g, '[$1]($2)');
}

function convertCodeBlocks(content: string): string {
  // Convert RST code blocks to Markdown code blocks
  let result = content;

  // Handle code blocks with :: notation
  result = result.replace(/::(?:\s*\n+)(\s+[\s\S]+?)(?:\n\n|\n$)/g, (match, code) => {
    const indentedCode = code.split('\n')
      .map((line: string) => line.replace(/^\s{4}/, ''))
      .join('\n');
    return '\n```\n' + indentedCode + '\n```\n\n';
  });

  return result;
}

function convertLists(content: string): string {
  // Convert RST lists to Markdown lists
  let result = content;

  // Handle bulleted lists (- and *)
  result = result.replace(/^\* (.*)/gm, '- $1');

  // Handle numbered lists (1., 2., etc)
  result = result.replace(/^(\d+)\. (.*)/gm, '$1. $2');

  return result;
}

export { parseSnootyContent, convertToMarkdown };