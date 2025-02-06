export interface LangCategoryCounts {
    bash?: number;
    c?: number;
    cpp?: number;
    csharp?: number;
    go?: number;
    java?: number;
    javascript?: number;
    json?: number;
    kotlin?: number;
    php?: number;
    python?: number;
    ruby?: number;
    rust?: number;
    scala?: number;
    shell?: number;
    swift?: number;
    text?: number;
    totals: number;
    typescript?: number,
    xml?: number,
    yaml?: number;
}

export interface RepoCategoryReport {
    total_code_blocks: number;
    categorization_details: {
        llm_categorized_count: number;
        string_matched_count: number;
        accuracy_estimate: number;
    };
    category_language_counts: {
        [category: string]: LangCategoryCounts;
    }
}

export type CategorizationDetails = {
    llm_categorized_count: number;
    string_matched_count: number;
    accuracy_estimate: number;
};

export type CategoryLanguageCounts = {
    [category: string]: {
        [language: string]: number;
        totals: number;
    };
};

export type GeneratedCategoryReport = {
    total_code_blocks: number;
    categorization_details: CategorizationDetails;
    category_language_counts: CategoryLanguageCounts;
};

export type RepoSummary = {
    name: string,
    atlasCliCommand?: number,
    apiMethodSignature?: number,
    exampleReturnObject?: number,
    exampleConfigObject?: number,
    mongoshCommand?: number,
    nonMongoCommand?: number,
    usageExample?: number,
    uncategorized?: number,
    totalCodeBlocks: number,
    categorizationDetails: {
        llmCategorizedCount: number,
        stringMatchedCount: number,
        accuracyEstimate: number
    }
}

export type CategorySummary = {
    name: string,
    atlasCliCommand: number,
    apiMethodSignature: number,
    exampleReturnObject: number,
    exampleConfigObject: number,
    mongoshCommand: number,
    nonMongoCommand: number,
    usageExample: number,
    uncategorized: number,
    totalCodeBlocks: number,
}
