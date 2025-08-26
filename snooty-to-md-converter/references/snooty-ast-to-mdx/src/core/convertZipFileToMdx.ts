import fs from 'fs';
import path from 'node:path';
import chalk from 'chalk';
import unzipper from 'unzipper';
import { BSON } from 'bson';
import { convertJsonAstToMdx } from './convertJsonAstToMdx';

/** some BSON files are not AST JSON, but rather raw text or RST */
const IGNORED_FILE_SUFFIXES = ['.txt.bson', '.rst.bson'] as const;

type ConvertZipFileToMdx = (args: {
  zipPath: string;
  outputPrefix?: string;
}) => Promise<{
  outputDirectory: string;
  fileCount: number;
}>

/** Convert a zip file to a folder of MDX files, preserving the zip's directory structure */
export const convertZipFileToMdx: ConvertZipFileToMdx = async ({ zipPath, outputPrefix }) => {
  const zipDir = await unzipper.Open.file(zipPath);

  const zipBaseNameRaw = path.basename(zipPath, '.zip');
  const zipBaseName = outputPrefix ? path.join(outputPrefix, zipBaseNameRaw) : zipBaseNameRaw;
  fs.mkdirSync(zipBaseName, { recursive: true });

  // Map asset checksum (compressed filename) -> semantic key (e.g., /images/foo.png)
  const checksumToKey = new Map<string, string>();
  const seenAssetChecksums = new Set<string>();

  let totalCount = 0;
  for (const file of zipDir.files) {
    // skip files that are not BSON files or have ignored suffixes
    if (file.type !== 'File' || !file.path.endsWith('.bson') || IGNORED_FILE_SUFFIXES.some(suffix => file.path.endsWith(suffix))) {
      (file as any).autodrain?.();
      continue;
    }

    // Read the BSON file as a buffer
    const buf = await file.buffer();
    // parse the buffer into BSON documents
    const docs: any[] = [];
    let offset = 0;
    while (offset < buf.length) {
      const size = buf.readInt32LE(offset);
      const slice = buf.subarray(offset, offset + size);
      docs.push(BSON.deserialize(slice));
      offset += size;
    }

    if (!docs.length) {
      continue;
    }
    if (docs.length > 1) {
      console.log(chalk.yellow(
        `\nWarning: ${chalk.cyan(file.path)} contains ${chalk.cyan(docs.length)} BSON documents - only the first one will be converted to MDX.\n`
      ));
    }

    const astTree = docs[0];
    // Collect static asset mappings for this page, if present
    if (astTree && Array.isArray(astTree.static_assets)) {
      for (const asset of astTree.static_assets) {
        const checksum = asset?.checksum;
        const key = asset?.key;
        if (typeof checksum === 'string' && typeof key === 'string' && checksum && key) {
          checksumToKey.set(checksum, key);
        }
      }
    }
    const relativePath = file.path.replace('.bson', '.mdx');
    const outputPath = path.join(zipBaseName, relativePath);
    // ensure the (potentially nested) output directory exists
    fs.mkdirSync(path.dirname(outputPath), { recursive: true });

    const { fileCount } = convertJsonAstToMdx({ astTree, outputPath, outputRootDir: zipBaseName });
    
    totalCount += fileCount;
    process.stdout.write(`\r${chalk.green(`✓ Wrote ${chalk.yellow(totalCount)} MDX files`)}`);
  }

  // ensure new line to print static asset logs, don't overwrite file count logs
  console.log('\n');

  // Second pass: extract non-BSON files that correspond to collected checksums
  for (const file of zipDir.files) {
    if (file.type !== 'File' || file.path.endsWith('.bson')) {
      (file as any).autodrain?.();
      continue;
    }
    const base = path.basename(file.path);
    const semanticKey = checksumToKey.get(base);
    if (!semanticKey) {
      (file as any).autodrain?.();
      continue;
    }
    if (seenAssetChecksums.has(base)) {
      (file as any).autodrain?.();
      continue;
    }
    // Read the "static asset" file as a buffer
    const buf = await file.buffer();
    const assetPath = semanticKey.replace(/^\/+/, '').replace(/\\+/g, '/');
    const outPath = path.join(zipBaseName, assetPath);
    fs.mkdirSync(path.dirname(outPath), { recursive: true });
    fs.writeFileSync(outPath, buf);

    seenAssetChecksums.add(base);
    process.stdout.write(`\r${chalk.green(`✓ Wrote ${chalk.yellow(seenAssetChecksums.size)} static assets`)}`);
  }

  totalCount += seenAssetChecksums.size;

  return { outputDirectory: zipBaseName, fileCount: totalCount };
}
