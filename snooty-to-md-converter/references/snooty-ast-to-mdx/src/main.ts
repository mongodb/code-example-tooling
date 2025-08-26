import fs from 'fs';
import chalk from 'chalk';
import { convertJsonAstToMdx } from './core/convertJsonAstToMdx';
import { convertZipFileToMdx } from './core/convertZipFileToMdx';

function printUsage() {
  console.log(chalk.magenta('\nUsage:'));
  console.log(chalk.cyan('    pnpm start'), chalk.yellow('/path/to/ast-input.json'));
  console.log(chalk.cyan('    pnpm start'), chalk.yellow('/path/to/doc-site.zip'), chalk.gray('/optional/output/folder'), '\n');
}

const main = async () => {
  const [_, __, input, outputPrefix] = process.argv;

  if (!input) {
    console.log(chalk.red('Error: No input file provided'));
    printUsage();
    process.exit(1);
  }

  const isJson = input.endsWith('.json');
  const isZip = input.endsWith('.zip');

  if (!isJson && !isZip) {
    console.log(chalk.red('Error: Input file must end in .json or .zip'));
    printUsage();
    process.exit(1);
  }

  if (isJson) {
    console.log(chalk.magenta(`Converting ${chalk.yellow(input)} to MDX...`), '\n');

    const astTree = JSON.parse(fs.readFileSync(input, 'utf8'));
    const outputPath = input.replace('.json', '_output.mdx');

    const { fileCount, emittedReferencesFile } = convertJsonAstToMdx({ astTree, outputPath });

    console.log(chalk.green(`✓ Wrote ${chalk.yellow(fileCount)} file(s)`), '\n');
    if (emittedReferencesFile) {
      console.log(chalk.green(`✓ Wrote ${chalk.yellow('./' + emittedReferencesFile)}`), '\n');
    }
    console.log(chalk.green(`✓ Wrote ${chalk.yellow(outputPath)}`), '\n');
  } else {
    console.log(chalk.magenta(`Converting ${chalk.yellow(input)} to MDX...`), '\n');

    const { outputDirectory, fileCount } = await convertZipFileToMdx({ zipPath: input, outputPrefix });

    console.log(chalk.green(`\n\n✓ Wrote folder ${chalk.yellow(outputDirectory + '/')} -- ${chalk.yellow(fileCount)} total files\n`));
  }
}

main();
