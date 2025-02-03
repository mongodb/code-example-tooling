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
    [category: string]: LangCategoryCounts;
}

export type RepoSummary = {
    name: string,
    atlasCliCommand?: number,
    apiMethodSignature?: number,
    exampleReturnObject?: number,
    exampleConfigObject?: number,
    usageExample?: number,
    uncategorized?: number,
    totalCodeBlocks: number,
}

export type CategorySummary = {
    name: string,
    atlasCliCommand: number,
    apiMethodSignature: number,
    exampleReturnObject: number,
    exampleConfigObject: number,
    usageExample: number,
    uncategorized: number,
    totalCodeBlocks: number,
}
