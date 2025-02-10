import { promises as fs } from 'fs';
import * as path from 'path';
import {sumCategoryTotals} from "./sumCategoryTotals";
import {CategorySummary, RepoSummary } from "./models/category-report";

export async function readCategoryFilesRecursively(
    dir: string,
    categoryCounts: CategorySummary,
    repoReports: RepoSummary[]
): Promise<[CategorySummary, RepoSummary[]]> {
    // Read the contents of the directory
    const entries = await fs.readdir(dir, { withFileTypes: true });
    // Iterate over each entry in the directory
    await Promise.all(entries.map(async (entry) => {
        const fullPath = path.join(dir, entry.name); // Construct the full path
        if (entry.isDirectory()) {
            // Recursively process subdirectory
            await readCategoryFilesRecursively(fullPath, categoryCounts, repoReports);
        } else if (entry.isFile() && entry.name.includes("language_category_counts")) {
            const filePathParts = path.normalize(dir).split(path.sep);
            const projectName = filePathParts[filePathParts.length - 1];
            const data = await fs.readFile(fullPath, 'utf8');
            const totals = sumCategoryTotals(projectName, data);
            // Update categoryCounts
            if (totals.exampleConfigObject) {
                categoryCounts.exampleConfigObject += totals.exampleConfigObject;
            }
            if (totals.exampleReturnObject) {
                categoryCounts.exampleReturnObject += totals.exampleReturnObject;
            }
            if (totals.nonMongoCommand) {
                categoryCounts.nonMongoCommand += totals.nonMongoCommand;
            }
            if (totals.syntaxExample) {
                categoryCounts.syntaxExample += totals.syntaxExample;
            }
            if (totals.uncategorized) {
                categoryCounts.uncategorized += totals.uncategorized;
            }
            if (totals.usageExample) {
                categoryCounts.usageExample += totals.usageExample;
            }
            if (totals.totalCodeBlocks) {
                categoryCounts.totalCodeBlocks += totals.totalCodeBlocks;
            }
            repoReports.push(totals);
        }
    }));
    return [categoryCounts, repoReports];
}
