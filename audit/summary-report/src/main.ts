import {processCodeExampleCategories} from "./processCodeExampleCategories";
import {processCodeLangCounts} from "./processCodeLangCounts";

async function main() {
    await processCodeExampleCategories();
    await processCodeLangCounts();
}

main();