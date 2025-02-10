import {CategorySummary, RepoSummary} from "./models/category-report";
import {readCategoryFilesRecursively} from "./readCategoryFilesRecursively";
import {writeJSONToFile} from "./writeToFile";

export async function processCodeExampleCategories() {
    // Path to the top-level directory you want to traverse
    const directoryPath = '/Users/dachary.carey/workspace/code-example-reports/category-reports';
    const categoryCounts: CategorySummary = {
        name: "Totals across all repos",
        exampleConfigObject: 0,
        exampleReturnObject: 0,
        nonMongoCommand: 0,
        syntaxExample: 0,
        usageExample: 0,
        uncategorized: 0,
        totalCodeBlocks: 0,
    };
    const repoReports: RepoSummary[] = [];
    // Start reading files from the top-level directory
    const result = await readCategoryFilesRecursively(directoryPath, categoryCounts, repoReports);
    await writeJSONToFile("./output/aggregate-category-report.json", result[0]);
    await writeJSONToFile("./output/repo-category-summary.json", result[1]);
}