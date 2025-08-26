import type { Node } from 'unist';

type ConversionContext = {
  registerImport?: (componentName: string, importPath: string) => void;
  emitMDXFile?: (outfilePath: string, mdastRoot: MdastNode) => void;
  /** Relative path (POSIX) of the file currently being generated, e.g. 'includes/foo.mdx' */
  currentOutfilePath?: string;
  /** Collected references to emit into a references.ts artifact */
  collectedSubstitutions: Map<string, string>;
  collectedRefs: Map<string, { title: string; url: string }>;
};

export interface MdastNode extends Node {
  [key: string]: any;
}

// Flexible SnootyNode interface that matches what the parser actually produces
// The parser output doesn't strictly follow the types in ast.ts
export interface SnootyNode {
  type: string;
  children?: SnootyNode[];
  value?: string;
  // Snooty specific properties we care about
  refuri?: string;
  language?: string;
  lang?: string;
  start?: number;
  startat?: number;
  depth?: number;
  title?: string;
  name?: string;
  argument?: SnootyNode[] | string;
  options?: Record<string, any>;
  enumtype?: 'ordered' | 'unordered';
  ordered?: boolean;
  label?: string;
  term?: SnootyNode[];
  html_id?: string;
  ids?: string[];
  refname?: string;
  target?: string;
  url?: string;
  domain?: string;
  admonition_type?: string;
  [key: string]: any;
}
