import {GeneratedCategoryReport, RepoSummary} from "./models/category-report";

export const sumCategoryTotals = (
    projectName: string,
    data: string
): RepoSummary => {
    const repoCounts: GeneratedCategoryReport = JSON.parse(data);
    // console.log("I have parsed the data to a GeneratedCategoryReport which looks like this:")
    // console.log(repoCounts);
    const summary: RepoSummary = {
        name: projectName,
        totalCodeBlocks: 0,
        categorizationDetails: {
            llmCategorizedCount: repoCounts.categorization_details.llm_categorized_count || 0,
            stringMatchedCount: repoCounts.categorization_details.string_matched_count || 0,
            accuracyEstimate: repoCounts.categorization_details.accuracy_estimate || 0,
        }
    };
    for (const category in repoCounts.category_language_counts) {
        const categoryData = repoCounts.category_language_counts[category];
        if (category == "Example configuration object") {
            summary.exampleConfigObject = categoryData.totals
        } else if (category == "Example return object") {
            summary.exampleReturnObject = categoryData.totals
        } else if (category == "Non-MongoDB command") {
            summary.nonMongoCommand = categoryData.totals
        } else if (category == "Syntax example") {
            summary.syntaxExample = categoryData.totals
        } else if (category == "Task-based usage") {
            summary.usageExample = categoryData.totals
        } else if (category == "Uncategorized") {
            summary.uncategorized = categoryData.totals
        }
        summary.totalCodeBlocks += categoryData.totals;
    }
    if (summary.totalCodeBlocks != repoCounts.total_code_blocks) {
        console.error("Report says there should be %s code blocks, but mathing the category counts equals %s", repoCounts.total_code_blocks, summary.totalCodeBlocks);
    }
    return summary;
}