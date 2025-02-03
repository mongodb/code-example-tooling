import {CodeNodeTypesByLang, LangData, RepoLangReport, LangCounts } from "./models/lang-counts";

export const sumLangTotals = (
    projectName: string,
    data: string
): RepoLangReport => {
    const repoCounts: CodeNodeTypesByLang = JSON.parse(data);
    const codeNodeLangCounts: LangCounts = {
        bash: repoCounts.codeNodesByLang.bash,
        c: repoCounts.codeNodesByLang.c,
        cpp: repoCounts.codeNodesByLang.cpp,
        csharp: repoCounts.codeNodesByLang.csharp,
        go: repoCounts.codeNodesByLang.go,
        java: repoCounts.codeNodesByLang.java,
        javascript: repoCounts.codeNodesByLang.javascript,
        json: repoCounts.codeNodesByLang.json,
        kotlin: repoCounts.codeNodesByLang.kotlin,
        php: repoCounts.codeNodesByLang.php,
        python: repoCounts.codeNodesByLang.python,
        ruby: repoCounts.codeNodesByLang.ruby,
        rust: repoCounts.codeNodesByLang.rust,
        scala: repoCounts.codeNodesByLang.scala,
        shell: repoCounts.codeNodesByLang.shell,
        swift: repoCounts.codeNodesByLang.swift,
        text: repoCounts.codeNodesByLang.text,
        typescript: repoCounts.codeNodesByLang.typescript,
        undefined: repoCounts.codeNodesByLang.undefined,
        xml: repoCounts.codeNodesByLang.xml,
        yaml: repoCounts.codeNodesByLang.yaml,
    };
    const literalIncludeNodeLangCounts: LangCounts = {
        bash: repoCounts.literalIncludesByLang.bash,
        c: repoCounts.literalIncludesByLang.c,
        cpp: repoCounts.literalIncludesByLang.cpp,
        csharp: repoCounts.literalIncludesByLang.csharp,
        go: repoCounts.literalIncludesByLang.go,
        java: repoCounts.literalIncludesByLang.java,
        javascript: repoCounts.literalIncludesByLang.javascript,
        json: repoCounts.literalIncludesByLang.json,
        kotlin: repoCounts.literalIncludesByLang.kotlin,
        php: repoCounts.literalIncludesByLang.php,
        python: repoCounts.literalIncludesByLang.python,
        ruby: repoCounts.literalIncludesByLang.ruby,
        rust: repoCounts.literalIncludesByLang.rust,
        scala: repoCounts.literalIncludesByLang.scala,
        shell: repoCounts.literalIncludesByLang.shell,
        swift: repoCounts.literalIncludesByLang.swift,
        text: repoCounts.literalIncludesByLang.text,
        typescript: repoCounts.literalIncludesByLang.typescript,
        undefined: repoCounts.literalIncludesByLang.undefined,
        xml: repoCounts.literalIncludesByLang.xml,
        yaml: repoCounts.literalIncludesByLang.yaml,
    }
    const langData: LangData = {
        codeNodes: repoCounts.totalCodeNodesByDirective,
        literalIncludes: repoCounts.totalLiteralIncludesByDirective,
        issueCount: repoCounts.pagesWithIssues.length || 0,
        codeNodesByLang: codeNodeLangCounts,
        literalIncludesByLang: literalIncludeNodeLangCounts
    };
    return {
        repo: projectName,
        data: langData,
    };
}