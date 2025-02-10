import { promises as fs } from 'fs';
import * as path from 'path';
import {sumLangTotals} from "./sumLangTotals";
import {LangData, RepoLangReport} from "./models/lang-counts";

function removeFileExtension(filename: string): string {
    // Get the extension of the file
    const extension = path.extname(filename);
    // Return the filename without the extension
    return path.basename(filename, extension);
}

export async function readLangFilesRecursively(
    dir: string,
    langCounts: LangData,
    repoReports: RepoLangReport[]
): Promise<[LangData, RepoLangReport[]]> {
    // Read the contents of the directory
    const entries = await fs.readdir(dir, { withFileTypes: true });
    // Iterate over each entry in the directory
    await Promise.all(entries.map(async (entry) => {
        const fullPath = path.join(dir, entry.name); // Construct the full path
        if (entry.isDirectory()) {
            // Recursively process subdirectory
            await readLangFilesRecursively(fullPath, langCounts, repoReports);
        } else if (entry.isFile() && entry.name.includes("pages")) {
            // Do nothing because we aren't parsing the `-pages` report yet
        } else if (entry.isFile() && !entry.name.includes("pages")) {
            const prefix = "report-";
            let projectName = entry.name.slice(prefix.length);
            projectName = removeFileExtension(projectName);
            const data = await fs.readFile(fullPath, 'utf8');
            const totals = sumLangTotals(projectName, data);
            repoReports.push(totals);
            langCounts.codeNodes += totals.data.codeNodes;
            langCounts.literalIncludes += totals.data.literalIncludes;
            langCounts.ioCodeBlocks += totals.data.ioCodeBlocks;
            langCounts.issueCount += totals.data.issueCount;
            langCounts.codeNodesByLang.bash += totals.data.codeNodesByLang.bash;
            langCounts.codeNodesByLang.c += totals.data.codeNodesByLang.c;
            langCounts.codeNodesByLang.cpp += totals.data.codeNodesByLang.cpp;
            langCounts.codeNodesByLang.csharp += totals.data.codeNodesByLang.csharp;
            langCounts.codeNodesByLang.go += totals.data.codeNodesByLang.go;
            langCounts.codeNodesByLang.java += totals.data.codeNodesByLang.java;
            langCounts.codeNodesByLang.javascript += totals.data.codeNodesByLang.javascript;
            langCounts.codeNodesByLang.json += totals.data.codeNodesByLang.json;
            langCounts.codeNodesByLang.kotlin += totals.data.codeNodesByLang.kotlin;
            langCounts.codeNodesByLang.php += totals.data.codeNodesByLang.php;
            langCounts.codeNodesByLang.python += totals.data.codeNodesByLang.python;
            langCounts.codeNodesByLang.ruby += totals.data.codeNodesByLang.ruby;
            langCounts.codeNodesByLang.rust += totals.data.codeNodesByLang.rust;
            langCounts.codeNodesByLang.scala += totals.data.codeNodesByLang.scala;
            langCounts.codeNodesByLang.shell += totals.data.codeNodesByLang.shell;
            langCounts.codeNodesByLang.swift += totals.data.codeNodesByLang.swift;
            langCounts.codeNodesByLang.text += totals.data.codeNodesByLang.text;
            langCounts.codeNodesByLang.typescript += totals.data.codeNodesByLang.typescript;
            langCounts.codeNodesByLang.undefined += totals.data.codeNodesByLang.undefined;
            langCounts.codeNodesByLang.xml += totals.data.codeNodesByLang.xml;
            langCounts.codeNodesByLang.yaml += totals.data.codeNodesByLang.yaml;
            langCounts.literalIncludesByLang.bash += totals.data.literalIncludesByLang.bash;
            langCounts.literalIncludesByLang.c += totals.data.literalIncludesByLang.c;
            langCounts.literalIncludesByLang.cpp += totals.data.literalIncludesByLang.cpp;
            langCounts.literalIncludesByLang.csharp += totals.data.literalIncludesByLang.csharp;
            langCounts.literalIncludesByLang.go += totals.data.literalIncludesByLang.go;
            langCounts.literalIncludesByLang.java += totals.data.literalIncludesByLang.java;
            langCounts.literalIncludesByLang.javascript += totals.data.literalIncludesByLang.javascript;
            langCounts.literalIncludesByLang.json += totals.data.literalIncludesByLang.json;
            langCounts.literalIncludesByLang.kotlin += totals.data.literalIncludesByLang.kotlin;
            langCounts.literalIncludesByLang.php += totals.data.literalIncludesByLang.php;
            langCounts.literalIncludesByLang.python += totals.data.literalIncludesByLang.python;
            langCounts.literalIncludesByLang.ruby += totals.data.literalIncludesByLang.ruby;
            langCounts.literalIncludesByLang.rust += totals.data.literalIncludesByLang.rust;
            langCounts.literalIncludesByLang.scala += totals.data.literalIncludesByLang.scala;
            langCounts.literalIncludesByLang.shell += totals.data.literalIncludesByLang.shell;
            langCounts.literalIncludesByLang.swift += totals.data.literalIncludesByLang.swift;
            langCounts.literalIncludesByLang.text += totals.data.literalIncludesByLang.text;
            langCounts.literalIncludesByLang.typescript += totals.data.literalIncludesByLang.typescript;
            langCounts.literalIncludesByLang.undefined += totals.data.literalIncludesByLang.undefined;
            langCounts.literalIncludesByLang.xml += totals.data.literalIncludesByLang.xml;
            langCounts.literalIncludesByLang.yaml += totals.data.literalIncludesByLang.yaml;
            langCounts.ioCodeBlockByLang.bash += totals.data.ioCodeBlockByLang.bash;
            langCounts.ioCodeBlockByLang.c += totals.data.ioCodeBlockByLang.c;
            langCounts.ioCodeBlockByLang.cpp += totals.data.ioCodeBlockByLang.cpp;
            langCounts.ioCodeBlockByLang.csharp += totals.data.ioCodeBlockByLang.csharp;
            langCounts.ioCodeBlockByLang.go += totals.data.ioCodeBlockByLang.go;
            langCounts.ioCodeBlockByLang.java += totals.data.ioCodeBlockByLang.java;
            langCounts.ioCodeBlockByLang.javascript += totals.data.ioCodeBlockByLang.javascript;
            langCounts.ioCodeBlockByLang.json += totals.data.ioCodeBlockByLang.json;
            langCounts.ioCodeBlockByLang.kotlin += totals.data.ioCodeBlockByLang.kotlin;
            langCounts.ioCodeBlockByLang.php += totals.data.ioCodeBlockByLang.php;
            langCounts.ioCodeBlockByLang.python += totals.data.ioCodeBlockByLang.python;
            langCounts.ioCodeBlockByLang.ruby += totals.data.ioCodeBlockByLang.ruby;
            langCounts.ioCodeBlockByLang.rust += totals.data.ioCodeBlockByLang.rust;
            langCounts.ioCodeBlockByLang.scala += totals.data.ioCodeBlockByLang.scala;
            langCounts.ioCodeBlockByLang.shell += totals.data.ioCodeBlockByLang.shell;
            langCounts.ioCodeBlockByLang.swift += totals.data.ioCodeBlockByLang.swift;
            langCounts.ioCodeBlockByLang.text += totals.data.ioCodeBlockByLang.text;
            langCounts.ioCodeBlockByLang.typescript += totals.data.ioCodeBlockByLang.typescript;
            langCounts.ioCodeBlockByLang.undefined += totals.data.ioCodeBlockByLang.undefined;
            langCounts.ioCodeBlockByLang.xml += totals.data.ioCodeBlockByLang.xml;
            langCounts.ioCodeBlockByLang.yaml += totals.data.ioCodeBlockByLang.yaml;
        }
    }));
    return [langCounts, repoReports];
}
