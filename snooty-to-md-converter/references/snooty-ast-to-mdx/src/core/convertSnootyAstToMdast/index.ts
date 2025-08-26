import yaml from 'yaml';
import { toComponentName } from './toComponentName';
import {
  normalize,
  dirname,
  relativePath,
  stripTsExtension,
} from '../path';
import type { ConversionContext, SnootyNode, MdastNode } from './types';
import { convertDirectiveImage } from './convertDirectiveImage';
import { getImporterContext } from './getImporterContext';

// Table node -> JSX element name map
const TABLE_ELEMENT_MAP: Record<string, string> = {
  table: 'Table',
  table_head: 'TableHead',
  table_body: 'TableBody',
  table_row: 'TableRow',
  table_cell: 'TableCell',
};

const convertDirectiveLiteralInclude = (node: SnootyNode): MdastNode => {
  const pathText = Array.isArray(node.argument)
    ? node.argument.map((a: any) => a.value ?? '').join('')
    : String(node.argument || '');
  const codeValue = `// Source: ${pathText.trim()}\n// TODO: Content from external file not available during conversion`;
  return { type: 'code', lang: node.options?.language ?? null, value: codeValue } as MdastNode;
};

const convertDirectiveInclude = (node: SnootyNode, ctx: ConversionContext, sectionDepth: number): MdastNode => {
  const pathText = Array.isArray(node.argument)
    ? node.argument.map((a: any) => a.value ?? '').join('')
    : String(node.argument || '');

  const toMdxIncludePath = (p: string): string => {
    const trimmed = p.trim();
    if (/\.(rst|txt)$/i.test(trimmed)) return trimmed.replace(/\.(rst|txt)$/i, '.mdx');
    if (!/\.mdx$/i.test(trimmed)) return `${trimmed}.mdx`;
    return trimmed;
  };

  const emittedPath = toMdxIncludePath(pathText);
  const emittedPathNormalized = emittedPath.replace(/^\/+/, '');

  const originalChildren: SnootyNode[] = Array.isArray(node.children) ? node.children : [];
  let contentChildren: SnootyNode[] = originalChildren;
  if (
    originalChildren.length === 1 &&
    originalChildren[0] &&
    originalChildren[0].type === 'directive' &&
    String(originalChildren[0].name ?? '').toLowerCase() === 'extract'
  ) {
    contentChildren = Array.isArray(originalChildren[0].children) ? (originalChildren[0].children as SnootyNode[]) : [];
  }

  const nestedRoot: SnootyNode = { type: 'root', children: contentChildren };
  const emittedMdast = convertSnootyAstToMdast(nestedRoot, {
    onEmitMDXFile: ctx.emitMDXFile,
    currentOutfilePath: normalize(emittedPathNormalized),
  });
  ctx.emitMDXFile?.(emittedPathNormalized, emittedMdast);

  const baseName = normalize(emittedPathNormalized).split('/').pop() || '';
  const withoutExt = baseName.replace(/\.mdx$/i, '');
  let componentName = toComponentName(withoutExt).replace(/\./g, '_');
  if (/^\d/.test(componentName)) componentName = `_${componentName}`;

  const { importerDir } = getImporterContext(ctx);
  const targetPosix = emittedPathNormalized.replace(/^\/*/, '').replace(/\\+/g, '/');
  let importPath = relativePath(importerDir, targetPosix);
  if (!importPath.startsWith('.')) importPath = `./${importPath}`;
  ctx.registerImport?.(componentName, importPath);

  return { type: 'mdxJsxFlowElement', name: componentName, attributes: [], children: [] } as MdastNode;
};

/** Convert a list of Snooty nodes to a list of mdast nodes */
const convertChildren = (nodes: SnootyNode[] | undefined, depth: number, ctx: ConversionContext): MdastNode[] => {
  if (!nodes || !Array.isArray(nodes)) return [];
  return nodes
    .map((n) => convertNode(n, depth, ctx))
    .flat()
    .filter(Boolean) as MdastNode[];
}

/** Convert a single Snooty node to mdast. Certain nodes (e.g. `section`) expand
    into multiple mdast siblings, so the return type can be an array. */
const convertNode = (node: SnootyNode, sectionDepth = 1, ctx: ConversionContext): MdastNode | MdastNode[] | null => {
  switch (node.type) {
    case 'text':
      return { type: 'text', value: node.value ?? '' };

    case 'paragraph':
      return {
        type: 'paragraph',
        children: convertChildren(node.children ?? [], sectionDepth, ctx),
      };

    case 'emphasis':
      return {
        type: 'emphasis',
        children: convertChildren(node.children ?? [], sectionDepth, ctx),
      };

    case 'strong':
      return {
        type: 'strong',
        children: convertChildren(node.children ?? [], sectionDepth, ctx),
      };

    case 'literal': { // inline code in Snooty AST
      // Snooty's "literal" inline code nodes sometimes store their text in
      // child "text" nodes rather than the `value` property. Fall back to
      // concatenating child text nodes when `value` is missing so that we
      // don't emit empty inline code (``) in the resulting MDX.
      let value: string = node.value ?? '';
      if (!value && Array.isArray(node.children)) {
        value = node.children
          .filter((c): c is SnootyNode => !!c)
          .filter((c) => c.type === 'text' || 'value' in c)
          .map((c: any) => c.value ?? '')
          .join('');
      }
      return { type: 'inlineCode', value };
    }

    case 'code': // literal_block is mapped to `code` in frontend AST
    case 'literal_block': {
      let value = node.value ?? '';
      if (!value && Array.isArray(node.children)) {
        value = node.children.map((c: any) => c.value ?? '').join('');
      }
      return { type: 'code', lang: node.lang ?? node.language ?? null, value };
    }

    case 'bullet_list':
      return {
        type: 'list',
        ordered: false,
        children: convertChildren(node.children ?? [], sectionDepth, ctx),
      };

    case 'enumerated_list':
    case 'ordered_list':
      return {
        type: 'list',
        ordered: true,
        start: node.start ?? 1,
        children: convertChildren(node.children ?? [], sectionDepth, ctx),
      };

    case 'list_item':
    case 'listItem':
      return {
        type: 'listItem',
        children: convertChildren(node.children ?? [], sectionDepth, ctx),
      };

    // Parser-emitted generic list node (covers both ordered & unordered)
    case 'list': {
      const ordered = (typeof node.enumtype === 'string' ? node.enumtype === 'ordered' : !!node.ordered);
      const start = ordered ? (node.startat ?? node.start ?? 1) : undefined;
      const mdastList: MdastNode = {
        type: 'list',
        ordered,
        children: convertChildren(node.children ?? [], sectionDepth, ctx),
      };
      if (ordered && typeof start === 'number') {
        (mdastList as any).start = start;
      }
      return mdastList;
    }

    // Field list (definition list–like) support
    case 'field_list':
      return {
        type: 'mdxJsxFlowElement',
        name: 'FieldList',
        attributes: [],
        children: convertChildren(node.children ?? [], sectionDepth, ctx),
      } as MdastNode;

    case 'field': {
      const attributes: MdastNode[] = [];
      if (node.name) attributes.push({ type: 'mdxJsxAttribute', name: 'name', value: String(node.name) });
      if (node.label) attributes.push({ type: 'mdxJsxAttribute', name: 'label', value: String(node.label) });
      return {
        type: 'mdxJsxFlowElement',
        name: 'Field',
        attributes,
        children: convertChildren(node.children ?? [], sectionDepth, ctx),
      } as MdastNode;
    }

    // Basic table support - many table types from parser
    case 'table':
    case 'table_head':
    case 'table_body':
    case 'table_row':
    case 'table_cell': {
      return {
        type: 'mdxJsxFlowElement',
        name: TABLE_ELEMENT_MAP[node.type] || 'Table',
        attributes: [],
        children: convertChildren(node.children ?? [], sectionDepth, ctx),
      } as MdastNode;
    }

    case 'reference':
      if (node.refuri) {
        return {
          type: 'link',
          url: node.refuri,
          children: convertChildren(node.children ?? [], sectionDepth, ctx),
        };
      }
      // fallthrough: treat as plain children if no URI
      return convertChildren(node.children ?? [], sectionDepth, ctx);

    case 'section': {
      // Snooty frontend AST uses a `heading` child, parser AST may use `title`.
      const titleNode = (node.children ?? []).find((c) => c.type === 'title' || c.type === 'heading');
      const rest = (node.children ?? []).filter((c) => c !== titleNode);
      const mdast: MdastNode[] = [];

      if (titleNode) {
        mdast.push({
          type: 'heading',
          depth: Math.min(sectionDepth, 6),
          children: convertChildren(titleNode.children ?? [], sectionDepth, ctx),
        });
      }

      rest.forEach((child) => {
        const converted = convertNode(child, sectionDepth + 1, ctx);
        if (Array.isArray(converted)) mdast.push(...converted);
        else if (converted) mdast.push(converted);
      });

      return mdast;
    }

    case 'title':
    case 'heading':
      return {
        type: 'heading',
        depth: node.depth ?? Math.min(sectionDepth, 6),
        children: convertChildren(node.children ?? [], sectionDepth, ctx),
      };

    case 'directive': {
      const directiveName = String(node.name ?? '').toLowerCase();
      // Special-case <Meta> directives here: we collect them at root level.
      if (directiveName === 'meta') {
        // This node will be handled separately – skip here.
        return null;
      }
      // Render figure/image directive as an <Image /> with imported src
      if (directiveName === 'figure' || directiveName === 'image') {
        return convertDirectiveImage(node, ctx);
      }
      // Handle literalinclude specially
      if (directiveName === 'literalinclude') {
        return convertDirectiveLiteralInclude(node);
      }
      // Handle include/sharedinclude by emitting a standalone MDX file and importing/using it
      if (directiveName === 'include' || directiveName === 'sharedinclude') {
        return convertDirectiveInclude(node, ctx, sectionDepth);
      }
      
      // Generic fallback for any Snooty directive (block-level).
      const componentName = toComponentName(node.name ?? 'Directive');
      // Map directive options to JSX attributes.
      const attributes: MdastNode[] = [];
      if (node.options && typeof node.options === 'object') {
        for (const [key, value] of Object.entries(node.options)) {
          if (value === undefined) continue;
          // Strings can be written as-is, everything else becomes an
          // expression so that complex types survive serialisation.
          if (typeof value === 'string') {
            attributes.push({ type: 'mdxJsxAttribute', name: key, value });
          } else {
            attributes.push({
              type: 'mdxJsxAttribute',
              name: key,
              value: { type: 'mdxJsxAttributeValueExpression', value: JSON.stringify(value) },
            });
          }
        }
      }

      // Directive argument: for some directives we want it as an attribute (e.g. "only", "cond").
      let includeArgumentAsChild = true;
      if (node.argument && (directiveName === 'only' || directiveName === 'cond')) {
        // Convert the condition expression into an attribute instead of child text
        const exprText = Array.isArray(node.argument)
          ? node.argument.map((a: any) => a.value ?? '').join('')
          : String(node.argument);
        attributes.push({ type: 'mdxJsxAttribute', name: 'expr', value: exprText.trim() });
        includeArgumentAsChild = false;
      }

      // Collect children coming from the directive's argument and body.
      const children: MdastNode[] = [];
      if (includeArgumentAsChild) {
        if (Array.isArray(node.argument)) {
          children.push(...convertChildren(node.argument, sectionDepth, ctx));
        } else if (typeof node.argument === 'string') {
          children.push({ type: 'text', value: node.argument });
        }
      }
      children.push(...convertChildren(node.children ?? [], sectionDepth, ctx));

      // Filter out empty directive elements that don't contribute to the output
      const emptyDirectives = ['toctree', 'index', 'seealso'];
      if (emptyDirectives.includes(directiveName) && children.length === 0 && attributes.length === 0) {
        return null;
      }

      return {
        type: 'mdxJsxFlowElement',
        name: componentName,
        attributes,
        children,
      } as MdastNode;
    }

    case 'ref_role':
    case 'doc': {  // doc role is like ref_role
      // Cross-document / internal reference emitted as a link
      const url = node.url ?? node.refuri ?? node.target ?? '';
      if (!url) {
        return convertChildren(node.children ?? [], sectionDepth, ctx);
      }

      // Collect the display text for this Ref to centralize it
      const childText = extractInlineDisplayText(node.children ?? []);
      if (childText) {
        ctx.collectedRefs.set(url, { title: childText, url });
        return {
          type: 'mdxJsxTextElement',
          name: 'Ref',
          attributes: [{ type: 'mdxJsxAttribute', name: 'url', value: url }],
          children: [
            {
              type: 'mdxTextExpression',
              value: `refs[${JSON.stringify(url)}].title`,
            } as MdastNode,
          ],
        } as MdastNode;
      }

      // Fallback to original conversion if no child text found
      return {
        type: 'mdxJsxTextElement',
        name: 'Ref',
        attributes: [{ type: 'mdxJsxAttribute', name: 'url', value: url }],
        children: convertChildren(node.children ?? [], sectionDepth, ctx),
      } as MdastNode;
    }

    case 'role': {
      // Inline roles convert to inline JSX elements.
      const componentName = toComponentName(node.name ?? 'Role');
      const attributes: MdastNode[] = [];
      if (node.target) {
        attributes.push({ type: 'mdxJsxAttribute', name: 'target', value: node.target });
      }
      const children = convertChildren(node.children ?? [], sectionDepth, ctx);
      // If the role had a literal value but no children (e.g. :abbr:`abbr`)
      if (!children.length && node.value) {
        children.push({ type: 'text', value: node.value });
      }
      return {
        type: 'mdxJsxTextElement',
        name: componentName,
        attributes,
        children,
      } as MdastNode;
    }

    case 'superscript':
      return {
        type: 'mdxJsxTextElement',
        name: 'sup',
        attributes: [],
        children: convertChildren(node.children ?? [], sectionDepth, ctx),
      } as MdastNode;

    case 'subscript':
      return {
        type: 'mdxJsxTextElement',
        name: 'sub',
        attributes: [],
        children: convertChildren(node.children ?? [], sectionDepth, ctx),
      } as MdastNode;

    case 'definitionList': {
      const children = convertChildren(node.children ?? [], sectionDepth, ctx);
      return {
        type: 'mdxJsxFlowElement',
        name: 'DefinitionList',
        attributes: [],
        children,
      } as MdastNode;
    }

    case 'definitionListItem': {
      const termChildren = convertChildren(node.term ?? [], sectionDepth, ctx);
      const descChildren = convertChildren(node.children ?? [], sectionDepth, ctx);
      return {
        type: 'mdxJsxFlowElement',
        name: 'DefinitionListItem',
        attributes: [],
        children: [...termChildren, ...descChildren],
      } as MdastNode;
    }

    case 'line_block': {
      // Convert each line into a separate text line with <br/> between them
      const lines = (node.children ?? []).flatMap((ln, idx, arr) => {
        const converted = convertChildren([ln], sectionDepth, ctx);
        if (idx < arr.length - 1) {
          // add a hard line break
          converted.push({ type: 'break' });
        }
        return converted;
      });
      return { type: 'paragraph', children: lines } as MdastNode;
    }

    case 'line':
      return { type: 'text', value: node.value ?? '' } as MdastNode;

    case 'title_reference':
      return {
        type: 'emphasis',
        children: convertChildren(node.children ?? [], sectionDepth, ctx),
      } as MdastNode;

    case 'footnote': {
      const identifier = String(node.id ?? node.name ?? '');
      if (!identifier) {
        // Fallback to emitting content inline if id missing
        return convertChildren(node.children ?? [], sectionDepth, ctx);
      }
      return {
        type: 'footnoteDefinition',
        identifier,
        label: node.name ?? undefined,
        children: convertChildren(node.children ?? [], sectionDepth, ctx),
      } as MdastNode;
    }

    case 'footnote_reference': {
      const identifier = String(node.id ?? '');
      if (!identifier) return null;
      return {
        type: 'footnoteReference',
        identifier,
        label: node.refname ?? undefined,
      } as MdastNode;
    }

    case 'named_reference':
      // Named references are link reference definitions that we've already resolved elsewhere; omit.
      return null;

    case 'substitution_definition':
      // Substitution definitions are processed elsewhere, skip them here
      return null;

    case 'substitution_reference':
    case 'substitution': {  // parser sometimes uses 'substitution' instead
      const refname = node.refname || node.name || '';
      const text = extractInlineDisplayText(node.children ?? []);
      if (refname && text) {
        ctx.collectedSubstitutions.set(refname, text);
        // Replace with inline expression reference
        return {
          type: 'mdxTextExpression',
          value: `substitutions[${JSON.stringify(refname)}]`,
        } as MdastNode;
      }
      // Fallback to rendering the original component if missing data
      const subChildren = convertChildren(node.children ?? [], sectionDepth, ctx);
      const attributes: MdastNode[] = [];
      if (refname) {
        attributes.push({ type: 'mdxJsxAttribute', name: 'name', value: refname });
      }
      return {
        type: 'mdxJsxTextElement',
        name: 'SubstitutionReference',
        attributes,
        children: subChildren,
      } as MdastNode;
    }

    case 'directive_argument':
      // Simply collapse and process its children.
      return convertChildren(node.children ?? [], sectionDepth, ctx);

    case 'transition':
      return { type: 'thematicBreak' };

    case 'card-group': {
      // Convert card-group to a JSX component
      const attributes: MdastNode[] = [];
      if (node.options && typeof node.options === 'object') {
        for (const [key, value] of Object.entries(node.options)) {
          if (value === undefined) continue;
          if (typeof value === 'string') {
            attributes.push({ type: 'mdxJsxAttribute', name: key, value });
          } else {
            attributes.push({
              type: 'mdxJsxAttribute',
              name: key,
              value: { type: 'mdxJsxAttributeValueExpression', value: JSON.stringify(value) },
            });
          }
        }
      }
      return {
        type: 'mdxJsxFlowElement',
        name: 'CardGroup',
        attributes,
        children: convertChildren(node.children ?? [], sectionDepth, ctx),
      } as MdastNode;
    }

    case 'cta-banner': {
      // Convert CTA banner to a JSX component
      const attributes: MdastNode[] = [];
      if (node.options && typeof node.options === 'object') {
        for (const [key, value] of Object.entries(node.options)) {
          if (value === undefined) continue;
          if (typeof value === 'string') {
            attributes.push({ type: 'mdxJsxAttribute', name: key, value });
          } else {
            attributes.push({
              type: 'mdxJsxAttribute',
              name: key,
              value: { type: 'mdxJsxAttributeValueExpression', value: JSON.stringify(value) },
            });
          }
        }
      }
      return {
        type: 'mdxJsxFlowElement',
        name: 'CTABanner',
        attributes,
        children: convertChildren(node.children ?? [], sectionDepth, ctx),
      } as MdastNode;
    }

    case 'tabs': {
      // Convert tabs container to a JSX component
      return {
        type: 'mdxJsxFlowElement',
        name: 'Tabs',
        attributes: [],
        children: convertChildren(node.children ?? [], sectionDepth, ctx),
      } as MdastNode;
    }

    case 'only': {
      // Convert only directive to a JSX component with condition
      const condition = Array.isArray(node.argument)
        ? node.argument.map((a: any) => a.value ?? '').join('')
        : String(node.argument || '');
      return {
        type: 'mdxJsxFlowElement',
        name: 'Only',
        attributes: [{ type: 'mdxJsxAttribute', name: 'condition', value: condition.trim() }],
        children: convertChildren(node.children ?? [], sectionDepth, ctx),
      } as MdastNode;
    }

    case 'method-selector': {
      // Convert method selector to a JSX component
      return {
        type: 'mdxJsxFlowElement',
        name: 'MethodSelector',
        attributes: [],
        children: convertChildren(node.children ?? [], sectionDepth, ctx),
      } as MdastNode;
    }

    case 'target': {
      // Convert to one or more invisible anchor <span> elements
      const ids: string[] = [];
      if (typeof node.html_id === 'string') ids.push(node.html_id);
      if (Array.isArray(node.ids)) ids.push(...node.ids);
      if (ids.length === 0 && typeof node.name === 'string') ids.push(node.name);
      if (ids.length === 0) return null;
      return ids.map((id) => ({
        type: 'mdxJsxFlowElement',
        name: 'span',
        attributes: [{ type: 'mdxJsxAttribute', name: 'id', value: id }],
        children: [],
      })) as MdastNode[];
    }

    case 'inline_target':
    case 'target_identifier': {
      const ids: string[] = [];
      if (Array.isArray(node.ids)) ids.push(...node.ids);
      if (typeof node.html_id === 'string') ids.push(node.html_id);
      if (ids.length === 0) return null;
      return ids.map((id) => ({
        type: 'mdxJsxFlowElement',
        name: 'span',
        attributes: [{ type: 'mdxJsxAttribute', name: 'id', value: id }],
        children: [],
      })) as MdastNode[];
    }

    // Additional parser node types not in standard AST types
    case 'block_quote':
      return {
        type: 'blockquote',
        children: convertChildren(node.children ?? [], sectionDepth, ctx),
      };

    case 'admonition': {
      // Admonitions are a type of directive
      const admonitionName = String(node.name ?? node.admonition_type ?? 'note');
      const componentName = toComponentName(admonitionName);
      return {
        type: 'mdxJsxFlowElement',
        name: componentName,
        attributes: [],
        children: convertChildren(node.children ?? [], sectionDepth, ctx),
      } as MdastNode;
    }

    // Parser-specific node types that we skip
    case 'comment':
    case 'comment_block':
      return null;

    default:
      // Unknown node → keep children if any, else emit comment.
      if (node.children && node.children.length) {
        return convertChildren(node.children, sectionDepth, ctx);
      }
      return { type: 'html', value: `<!-- unsupported: ${node.type} -->` };
  }
}

interface ConvertSnootyAstToMdastOptions {
  onEmitMDXFile?: ConversionContext['emitMDXFile'];
  currentOutfilePath?: string
}

export const convertSnootyAstToMdast = (root: SnootyNode, options?: ConvertSnootyAstToMdastOptions): MdastNode => {
  const metaFromDirectives: Record<string, any> = {};
  const contentChildren: MdastNode[] = [];
  const includedImports = new Map<string, string>();
  const collectedSubstitutions = new Map<string, string>();
  const collectedRefs = new Map<string, { title: string; url: string }>();

  const ctx: ConversionContext = {
    registerImport: (componentName: string, importPath: string) => {
      if (!componentName || !importPath) return;
      includedImports.set(componentName, importPath);
    },
    emitMDXFile: options?.onEmitMDXFile,
    currentOutfilePath: options?.currentOutfilePath,
    collectedSubstitutions,
    collectedRefs,
  };

  (root.children ?? []).forEach((child: SnootyNode) => {
    // Collect <meta> directives: they appear as directive nodes with name 'meta'.
    if (child.type === 'directive' && String(child.name).toLowerCase() === 'meta' && child.options) {
      Object.assign(metaFromDirectives, child.options);
      return; // do not include this node in output
    }
    const converted = convertNode(child, 1, ctx);
    if (Array.isArray(converted)) contentChildren.push(...converted);
    else if (converted) contentChildren.push(converted);
  });

  // Merge page-level options that sit on the root node itself.
  const pageOptions = (root as any).options ?? {};
  const frontmatterObj = { ...pageOptions, ...metaFromDirectives };

  // Compose final children array with optional frontmatter
  const children: MdastNode[] = [];
  if (Object.keys(frontmatterObj).length) {
    children.push({ type: 'yaml', value: yaml.stringify(frontmatterObj) } as MdastNode);
  }
  // Inject collected imports as ESM blocks right after frontmatter (or at top if no frontmatter)
  const wantRefs = ctx.collectedRefs.size > 0;
  const wantSubs = ctx.collectedSubstitutions.size > 0;
  if (includedImports.size > 0 || wantRefs || wantSubs) {
    const entries = Array.from(includedImports.entries());
    const nonImage: Array<[string, string]> = [];
    const image: Array<[string, string]> = [];

    const isImagePath = (p: string): boolean => /\.(png|jpe?g|gif|svg|webp|avif)$/i.test(p);
    for (const e of entries) {
      (isImagePath(e[1]) ? image : nonImage).push(e);
    }
    // ensure images are imported last (nice formatting)
    const ordered = [...nonImage, ...image];
    const importLines: string[] = ordered.map(([componentName, importPath]) => `import ${componentName} from '${importPath}';`);

    // Add structured imports for references if needed
    if (wantRefs || wantSubs) {
      const importerPosix = normalize(options?.currentOutfilePath || 'index.mdx');
      const importerDir = dirname(importerPosix);
      let importPath = relativePath(importerDir, 'references.ts');
      if (!importPath.startsWith('.')) importPath = `./${importPath}`;
      importPath = stripTsExtension(importPath);
      const named: string[] = [];
      if (wantRefs) named.push('refs');
      if (wantSubs) named.push('substitutions');
      importLines.push(`import { ${named.join(', ')} } from '${importPath}';`);
    }

    children.push({
      type: 'mdxjsEsm',
      value: importLines.join('\n'),
    } as MdastNode);
  }
  children.push(...contentChildren);

  const rootNode = {
    type: 'root',
    children: wrapInlineRuns(children),
  } as MdastNode;

  // Attach collected references so the caller can emit a references.ts artifact
  if (collectedSubstitutions.size > 0 || collectedRefs.size > 0) {
    const substitutions: Record<string, string> = {};
    for (const [k, v] of collectedSubstitutions.entries()) substitutions[k] = v;
    const refs: Record<string, { title: string; url: string }> = {};
    for (const [k, v] of collectedRefs.entries()) refs[k] = v;
    (rootNode as any).__references = { substitutions, refs };
  }

  return rootNode as MdastNode;
}

/** Ensure that any stray inline nodes at the root (or other flow-level
    parents) are wrapped in paragraphs so that the final mdast is valid and
    spaced correctly when stringified. */
const wrapInlineRuns = (nodes: MdastNode[]): MdastNode[] => {
  const result: MdastNode[] = [];
  let inlineRun: MdastNode[] = [];
  const isInline = (n: MdastNode) => {
    return (
      n.type === 'text' ||
      n.type === 'emphasis' ||
      n.type === 'strong' ||
      n.type === 'inlineCode' ||
      n.type === 'break' ||
      n.type === 'mdxJsxTextElement' ||
      n.type === 'sub' ||
      n.type === 'sup' ||
      n.type === 'link' ||
      n.type === 'footnoteReference'
    );
  };
  const flushInlineRun = () => {
    if (inlineRun.length) {
      result.push({ type: 'paragraph', children: inlineRun } as MdastNode);
      inlineRun = [];
    }
  };
  for (const node of nodes) {
    if (isInline(node)) {
      inlineRun.push(node);
    } else {
      flushInlineRun();
      // Recursively process children that are arrays (e.g., list, listItem, etc.)
      if (Array.isArray((node as any).children)) {
        (node as any).children = wrapInlineRuns((node as any).children as MdastNode[]);
      }
      result.push(node);
    }
  }
  flushInlineRun();
  return result;
};

/** Extract display text from inline children as a single string. */
const extractInlineDisplayText = (children: SnootyNode[]): string => {
  const parts: string[] = [];
  const walk = (n: SnootyNode | undefined) => {
    if (!n) return;
    if (n.type === 'text' && typeof n.value === 'string') {
      parts.push(n.value);
      return;
    }
    if (n.type === 'literal' && typeof n.value === 'string') {
      parts.push(n.value);
      return;
    }
    if (Array.isArray(n.children)) n.children.forEach(walk);
  };
  children.forEach(walk);
  const raw = parts.join('');
  // Unescape common Markdown/MDX backslash escapes (e.g., \_id -> _id)
  const unescaped = raw.replace(/\\([\\`*_{}\[\]()#+\-.!])/g, '$1');
  // Collapse excessive whitespace
  return unescaped.replace(/\s+/g, ' ').trim();
};
