import {readLangFilesRecursively} from "./readLangFilesRecursively";
import {writeJSONToFile} from "./writeToFile";
import {LangCounts, LangData, RepoLangReport} from "./models/lang-counts";

export async function processCodeLangCounts() {
    // Path to the top-level directory you want to traverse
    const directoryPath = '/Users/dachary.carey/workspace/code-example-reports/code-counts-reports';
    const langCounts: LangCounts = {
        bash: 0,
        c: 0,
        cpp: 0,
        csharp: 0,
        go: 0,
        java: 0,
        javascript: 0,
        json: 0,
        kotlin: 0,
        php: 0,
        python: 0,
        ruby: 0,
        rust: 0,
        scala: 0,
        shell: 0,
        swift: 0,
        text: 0,
        typescript: 0,
        undefined: 0,
        xml: 0,
        yaml: 0
    };
    const langData: LangData = {
        codeNodes: 0,
        literalIncludes: 0,
        issueCount: 0,
        codeNodesByLang: langCounts,
        literalIncludesByLang: langCounts,
    }
    const repoLangReports: RepoLangReport[] = [];
    // Start reading files from the top-level directory
    const result = await readLangFilesRecursively(directoryPath, langData, repoLangReports);
    console.log(result[0]);
    result[1].forEach((entry) => {
        console.log(entry);
    })
    await writeJSONToFile("./output/aggregate-lang-report.json", result[0]);
    await writeJSONToFile("./output/repo-lang-summary.json", result[1]);
}