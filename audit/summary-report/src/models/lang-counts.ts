export interface LangCounts {
    bash: number;
    c: number;
    cpp: number;
    csharp: number;
    go: number;
    java: number;
    javascript: number;
    json: number;
    kotlin: number;
    php: number;
    python: number;
    ruby: number;
    rust: number;
    scala: number;
    shell: number;
    swift: number;
    text: number;
    typescript: number,
    undefined: number;
    xml: number,
    yaml: number;
}

export interface RepoLangReport {
    repo: string,
    data: LangData,
}

export type LangData = {
    codeNodes: number,
    codeNodesByLang: LangCounts;
    literalIncludes: number,
    literalIncludesByLang: LangCounts;
    ioCodeBlocks: number,
    ioCodeBlockByLang: LangCounts;
    issueCount: number;
}

export interface CodeNodeTypesByLang {
    repo: string,
    totalCodeNodesByDirective: number,
    totalCodeNodesByLangSum: number,
    codeNodesByLang: LangCounts,
    totalLiteralIncludesByDirective: number,
    totalLiteralIncludesByLangSum: number,
    literalIncludesByLang: LangCounts,
    ioCodeBlockCountByDirective: number,
    ioCodeBlockCountByLangSum: number,
    ioCodeBlockByLang: LangCounts,
    pagesWithIssues: string[];
}
