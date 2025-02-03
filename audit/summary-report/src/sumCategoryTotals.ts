import {RepoCategoryReport, RepoSummary} from "./models/category-report";

export const sumCategoryTotals = (
    projectName: string,
    data: string
): RepoSummary => {
    const repoCounts: RepoCategoryReport = JSON.parse(data);
    const summary: RepoSummary = {
        name: projectName,
        totalCodeBlocks: 0,
    };
    for (const category in repoCounts) {
        const categoryData = repoCounts[category];
        if (category == "API Method Signature") {
            summary.apiMethodSignature = categoryData.totals
        } else if (category == "Atlas CLI Command") {
            summary.atlasCliCommand = categoryData.totals
        } else if (category == "Example configuration object") {
            summary.exampleConfigObject = categoryData.totals
        } else if (category == "Example return object") {
            summary.exampleReturnObject = categoryData.totals
        } else if (category == "Task-based usage") {
            summary.usageExample = categoryData.totals
        } else if (category == "Uncategorized") {
            summary.uncategorized = categoryData.totals
        }
        // @ts-ignore
        summary.totalCodeBlocks += categoryData.totals;
    }
    return summary;
}