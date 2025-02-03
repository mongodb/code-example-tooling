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
            if (totals.apiMethodSignature) {
                categoryCounts.apiMethodSignature += totals.apiMethodSignature;
            }
            if (totals.atlasCliCommand) {
                categoryCounts.atlasCliCommand += totals.atlasCliCommand;
            }
            if (totals.exampleConfigObject) {
                categoryCounts.exampleConfigObject += totals.exampleConfigObject;
            }
            if (totals.exampleReturnObject) {
                categoryCounts.exampleReturnObject += totals.exampleReturnObject;
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

// import * as fs from 'fs';
// import * as path from 'path';
// import {sumCategoryTotals} from "./sumCategoryTotals";
// import {CategorySummary, RepoSummary } from "./models/category-report";
// // Function to recursively read files
// export function readFilesRecursively(dir: string, categoryCounts: CategorySummary, repoReports: RepoSummary[] ): [categoryCounts: CategorySummary, repoReports: RepoSummary[]] {
//     // Read the contents of the directory
//     fs.readdir(dir, { withFileTypes: true }, (err, entries) => {
//         if (err) {
//             console.error('Could not read directory:', err);
//             return;
//         }
//         // Iterate over each entry in the directory
//         entries.forEach((entry) => {
//             const fullPath = path.join(dir, entry.name); // Construct the full path
//             if (entry.isDirectory()) {
//                 // If the entry is a directory, recursively read its contents
//                 readFilesRecursively(fullPath, categoryCounts, repoReports);
//             } else if (entry.isFile()) {
//                 // If the entry is a file, read its contents
//                 if (entry.name.includes("language_category_counts")) {
//                     const filePathParts = path.normalize(dir).split(path.sep);
//                     const projectName = filePathParts[filePathParts.length - 1];
//                     fs.readFile(fullPath, 'utf8', (err, data) => {
//                         if (err) {
//                             console.error(`Could not read file ${fullPath}:`, err);
//                             return;
//                         }
//                         const totals = sumCategoryTotals(projectName, data);
//                         if (totals.apiMethodSignature) {
//                             categoryCounts.apiMethodSignature += totals.apiMethodSignature;
//                         }
//                         if (totals.atlasCliCommand) {
//                             categoryCounts.atlasCliCommand += totals.atlasCliCommand;
//                         }
//                         if (totals.exampleConfigObject) {
//                             categoryCounts.exampleConfigObject += totals.exampleConfigObject;
//                         }
//                         if (totals.exampleReturnObject) {
//                             categoryCounts.exampleReturnObject += totals.exampleReturnObject;
//                         }
//                         if (totals.uncategorized) {
//                             categoryCounts.uncategorized += totals.uncategorized
//                         }
//                         if (totals.usageExample) {
//                             categoryCounts.usageExample += totals.usageExample;
//                         }
//                         if (totals.totalCodeBlocks) {
//                             categoryCounts.totalCodeBlocks += totals.totalCodeBlocks;
//                         }
//                         repoReports.push(totals);
//                         //console.log(categoryCounts);
//                     });
//                 }
//             }
//         });
//     });
//     return [categoryCounts, repoReports];
// }
