#!/usr/bin/env node
import 'dotenv/config';
import { convertRealmDocs, convertRealmDocsFromApi } from './realm-docs-converter';
import * as path from 'path';

function parseArgs(argv: string[]) {
  const args = { mode: 'api' as 'api' | 'local', project: 'realm', out: '', branch: 'master', baseUrl: undefined as string | undefined, input: undefined as string | undefined };
  const rest: string[] = [];
  for (let i = 0; i < argv.length; i++) {
    const a = argv[i];
    if (a === '--local') args.mode = 'local';
    else if (a === '--project' && argv[i+1]) { args.project = argv[++i]; }
    else if (a === '--out' && argv[i+1]) { args.out = argv[++i]; }
    else if (a === '--branch' && argv[i+1]) { args.branch = argv[++i]; }
    else if (a === '--base-url' && argv[i+1]) { args.baseUrl = argv[++i]; }
    else rest.push(a);
  }
  // local fallback: realm-docs-converter --local <input-dir> --out <output-dir>
  if (args.mode === 'local') {
    if (rest.length >= 1) args.input = rest[0];
  }
  return args;
}

async function main() {
  const args = parseArgs(process.argv.slice(2));

  let totalConverted = 0;

  if (args.mode === 'api') {
    if (!args.out) {
      console.log('Usage (API): realm-docs-converter --project realm --out <output-dir> [--branch <branch>] [--base-url <url>]');
      process.exit(1);
    }
    const outDir = path.resolve(args.out);
    console.log(`Fetching Snooty project ${args.project} (branch=${args.branch}) and converting to ${outDir}`);
    totalConverted = await convertRealmDocsFromApi({ project: args.project, outputDir: outDir, branch: args.branch, baseUrl: args.baseUrl });
  } else {
    // local mode for backward compatibility
    if (!args.input || !args.out) {
      console.log('Usage (Local): realm-docs-converter --local <input-dir> --out <output-dir>');
      process.exit(1);
    }
    const inputDir = path.resolve(args.input);
    const outputDir = path.resolve(args.out);
    console.log(`Converting local Snooty dir from ${inputDir} to ${outputDir}`);
    totalConverted = await convertRealmDocs({
      inputDir,
      outputDir,
      handleIncludes: true,
      handleSubstitutions: true,
      handleRefs: true,
    });
  }

  console.log('Conversion complete!');
  console.log(`Total pages converted: ${totalConverted}`);
}

main().catch((err) => { console.error(err); process.exit(1); });