/**
 * Path helpers using path.posix for OS-agnostic path manipulation
 */

import path from 'node:path';

export const normalize = (p: string): string => path.posix.normalize(p);

export const dirname = (p: string): string => path.posix.dirname(p);

export const relativePath = (from: string, to: string): string => {
  const rel = path.relative(from, to);
  return normalize(rel);
};

export const stripTsExtension = (p: string): string => p.replace(/\.ts$/, '');
