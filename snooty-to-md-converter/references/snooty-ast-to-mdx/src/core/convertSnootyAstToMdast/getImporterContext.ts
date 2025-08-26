import { normalize, dirname } from '../path';
import type { ConversionContext } from './types';

export const getImporterContext = (ctx: ConversionContext) => {
  const importerPosix = normalize(ctx.currentOutfilePath || 'index.mdx');
  const importerDir = dirname(importerPosix);
  return { importerPosix, importerDir };
};
