import fs from 'fs';
import path from 'node:path';
import chalk from 'chalk';
import { normalize } from '../path';
import { convertMdastToMdx } from '../convertMdastToMdx';
import { convertSnootyAstToMdast } from '../convertSnootyAstToMdast';
import { buildReferencesTs, readExistingReferences, mergeReferences } from './buildReferencesTs';

type ConvertJsonAstToMdx = (args: {
  astTree: any;
  outputPath: string;
  outputRootDir?: string;
}) => {
  fileCount: number;
  emittedReferencesFile?: string;
}

export const convertJsonAstToMdx: ConvertJsonAstToMdx = ({ astTree, outputPath, outputRootDir }) => {
  // Track unique output file paths we've written during this process to avoid
  // duplicate writes and over-counting when includes repeat across pages.
  const emittedFilePaths = new Set<string>();
  let emittedReferencesFile: string | undefined;
  
  // handle wrapper objects that store AST under `ast` field
  const snootyRoot = astTree.ast ?? astTree;

  let fileCount = 0;
  const rootDir = outputRootDir ?? path.dirname(outputPath);
  const aggregated: { substitutions: Record<string, string>; refs: Record<string, { title: string; url: string }> } = { substitutions: {}, refs: {} };
  const mdast = convertSnootyAstToMdast(snootyRoot, {
    onEmitMDXFile: (emitFilePath, mdastRoot) => {
      let didCreateFile = false;
      try {
        const outPath = path.join(rootDir, emitFilePath);
        const resolvedOutPath = path.resolve(outPath);
        if (!emittedFilePaths.has(resolvedOutPath)) {
          fs.mkdirSync(path.dirname(outPath), { recursive: true });
          const mdxContent = convertMdastToMdx(mdastRoot);
          fs.writeFileSync(outPath, mdxContent);
          emittedFilePaths.add(resolvedOutPath);
          didCreateFile = true;
        }
      } catch (err) {
        console.error(chalk.red('Failed to emit include file:'), emitFilePath, err);
      }

      if (didCreateFile) {
        fileCount++;

        const refs = mdastRoot.__references;
        if (refs) {
          Object.assign(aggregated.substitutions, refs.substitutions || {});
          Object.assign(aggregated.refs, refs.refs || {});
        }
      }
    },
    // Make the current output file path relative to the provided output root directory
    currentOutfilePath: normalize(path.relative(rootDir, outputPath)),
  });

  // If references were collected, emit or update a references.ts file at the output root
  const refsArtifact = mdast.__references;
  if (refsArtifact || Object.keys(aggregated.substitutions).length || Object.keys(aggregated.refs).length) {
    if (refsArtifact) {
      Object.assign(aggregated.substitutions, refsArtifact.substitutions || {});
      Object.assign(aggregated.refs, refsArtifact.refs || {});
    }
    const refsPath = path.join(rootDir, 'references.ts');
    fs.mkdirSync(path.dirname(refsPath), { recursive: true });
    const existing = fs.existsSync(refsPath) ? readExistingReferences(refsPath) : { substitutions: {}, refs: {} };
    const merged = mergeReferences(existing, aggregated);
    const file = buildReferencesTs(merged);
    fs.writeFileSync(refsPath, file);
    emittedReferencesFile = refsPath;
  }
  const mdx = convertMdastToMdx(mdast);
  const resolvedMainOutPath = path.resolve(outputPath);
  if (!emittedFilePaths.has(resolvedMainOutPath)) {
    fs.writeFileSync(outputPath, mdx);
    emittedFilePaths.add(resolvedMainOutPath);
    fileCount++;
  }

  return { fileCount, emittedReferencesFile };
}
