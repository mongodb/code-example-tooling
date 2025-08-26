import { relativePath } from '../path';
import type { ConversionContext, SnootyNode, MdastNode } from './types';
import { toComponentName } from './toComponentName';
import { getImporterContext } from './getImporterContext';

export const convertDirectiveImage = (node: SnootyNode, ctx: ConversionContext): MdastNode => {
  const argText = Array.isArray(node.argument)
    ? node.argument.map((a: any) => a.value ?? '').join('')
    : (typeof node.argument === 'string' ? node.argument : '');

  let pathText = extractPathFromNodes(node.children) || String(argText || '');
  let assetPosix = pathText
    .replace(/\\+/g, '/')
    .replace(/^\/+/, '')
    .replace(/^\/+/, '')
    .replace(/^\.\//, '')
    .replace(/\\\./g, '.');
  if (!assetPosix) {
    return { type: 'html', value: '<!-- figure missing src -->' } as MdastNode;
  }

  const { importerPosix, importerDir } = getImporterContext(ctx);

  const topLevel = importerPosix.includes('/') ? importerPosix.split('/')[0] : '';
  let targetPosix = assetPosix.replace(/^\/+/, '');
  const imagesIdx = targetPosix.indexOf('images/');
  if (imagesIdx >= 0) {
    const after = targetPosix.slice(imagesIdx + 'images/'.length);
    targetPosix = topLevel ? `${topLevel}/images/${after}` : `images/${after}`;
  } else if (assetPosix.startsWith('images/')) {
    const after = assetPosix.slice('images/'.length);
    targetPosix = topLevel ? `${topLevel}/images/${after}` : `images/${after}`;
  } else if (!targetPosix.includes('/')) {
    targetPosix = topLevel ? `${topLevel}/images/${targetPosix}` : `images/${targetPosix}`;
  }

  let importPath = relativePath(importerDir, targetPosix);
  if (!importPath.startsWith('.')) importPath = `./${importPath}`;
  if (importPath.startsWith('./')) {
    importPath = importPath.replace(/^\.\/+/, '../');
  } else {
    importPath = `../${importPath}`;
  }

  const baseName = targetPosix.split('/').pop() || 'image';
  const withoutExt = baseName.replace(/\.[^.]+$/, '') || 'image';
  let imageIdent = toComponentName(withoutExt).replace(/[^A-Za-z0-9_]/g, '_');
  if (/^\d/.test(imageIdent)) imageIdent = `_${imageIdent}`;
  imageIdent = `${imageIdent}Img`;

  ctx.registerImport?.(imageIdent, importPath);

  const attrs: MdastNode[] = [];
  attrs.push({
    type: 'mdxJsxAttribute',
    name: 'src',
    value: { type: 'mdxJsxAttributeValueExpression', value: imageIdent },
  } as MdastNode);
  const altText = typeof node.options?.alt === 'string' ? node.options.alt : '';
  if (altText) attrs.push({ type: 'mdxJsxAttribute', name: 'alt', value: altText } as MdastNode);
  const widthAttr = toNumericAttr('width', node.options?.width);
  const heightAttr = toNumericAttr('height', node.options?.height);
  if (widthAttr) attrs.push(widthAttr);
  if (heightAttr) attrs.push(heightAttr);

  return { type: 'mdxJsxFlowElement', name: 'Image', attributes: attrs, children: [] } as MdastNode;
};

const extractPathFromNodes = (nodes: SnootyNode[] | undefined): string => {
  if (!Array.isArray(nodes)) return '';
  const parts: string[] = [];
  const walk = (n: SnootyNode) => {
    if (!n) return;
    if (typeof n.value === 'string') parts.push(n.value);
    if (Array.isArray(n.children)) n.children.forEach(walk);
  };
  nodes.forEach(walk);
  return parts.join('').trim();
};

const toNumericAttr = (name: string, v: any): MdastNode | null => {
  if (v === undefined || v === null || v === '') return null;
  const num = typeof v === 'number' ? v : parseFloat(String(v));
  if (!Number.isNaN(num)) {
    return {
      type: 'mdxJsxAttribute',
      name,
      value: { type: 'mdxJsxAttributeValueExpression', value: String(num) },
    } as MdastNode;
  }
  return { type: 'mdxJsxAttribute', name, value: String(v) } as MdastNode;
};
