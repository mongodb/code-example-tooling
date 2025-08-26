import { remark } from 'remark';
import remarkMdx from 'remark-mdx';
import remarkFrontmatter from 'remark-frontmatter';
import remarkGfm from 'remark-gfm';

/** Convert an mdast tree to an MDX string. */
export const convertMdastToMdx = (tree: any): string => {
  const processor = remark()
    .use(remarkFrontmatter, ['yaml'])
    .use(remarkGfm)
    .use(remarkMdx);

  const output = processor.stringify(tree);
  return typeof output === 'string' ? output : String(output);
}
