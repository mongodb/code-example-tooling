#!/usr/bin/env node
"use strict";
var __createBinding = (this && this.__createBinding) || (Object.create ? (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    var desc = Object.getOwnPropertyDescriptor(m, k);
    if (!desc || ("get" in desc ? !m.__esModule : desc.writable || desc.configurable)) {
      desc = { enumerable: true, get: function() { return m[k]; } };
    }
    Object.defineProperty(o, k2, desc);
}) : (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    o[k2] = m[k];
}));
var __setModuleDefault = (this && this.__setModuleDefault) || (Object.create ? (function(o, v) {
    Object.defineProperty(o, "default", { enumerable: true, value: v });
}) : function(o, v) {
    o["default"] = v;
});
var __importStar = (this && this.__importStar) || (function () {
    var ownKeys = function(o) {
        ownKeys = Object.getOwnPropertyNames || function (o) {
            var ar = [];
            for (var k in o) if (Object.prototype.hasOwnProperty.call(o, k)) ar[ar.length] = k;
            return ar;
        };
        return ownKeys(o);
    };
    return function (mod) {
        if (mod && mod.__esModule) return mod;
        var result = {};
        if (mod != null) for (var k = ownKeys(mod), i = 0; i < k.length; i++) if (k[i] !== "default") __createBinding(result, mod, k[i]);
        __setModuleDefault(result, mod);
        return result;
    };
})();
Object.defineProperty(exports, "__esModule", { value: true });
require("dotenv/config");
const realm_docs_converter_1 = require("./realm-docs-converter");
const path = __importStar(require("path"));
function parseArgs(argv) {
    const args = { mode: 'api', project: 'realm', out: '', branch: 'master', baseUrl: undefined, input: undefined };
    const rest = [];
    for (let i = 0; i < argv.length; i++) {
        const a = argv[i];
        if (a === '--local')
            args.mode = 'local';
        else if (a === '--project' && argv[i + 1]) {
            args.project = argv[++i];
        }
        else if (a === '--out' && argv[i + 1]) {
            args.out = argv[++i];
        }
        else if (a === '--branch' && argv[i + 1]) {
            args.branch = argv[++i];
        }
        else if (a === '--base-url' && argv[i + 1]) {
            args.baseUrl = argv[++i];
        }
        else
            rest.push(a);
    }
    // local fallback: realm-docs-converter --local <input-dir> --out <output-dir>
    if (args.mode === 'local') {
        if (rest.length >= 1)
            args.input = rest[0];
    }
    return args;
}
async function main() {
    const args = parseArgs(process.argv.slice(2));
    let totalConverted = 0;
    if (args.mode === 'api') {
        if (!args.out) {
            console.log('Usage (API): realm-docs-converter --project realm --out <output-dir> [--branch <branch>] [--base-url <url>]');
            process.exit(1);
        }
        const outDir = path.resolve(args.out);
        console.log(`Fetching Snooty project ${args.project} (branch=${args.branch}) and converting to ${outDir}`);
        totalConverted = await (0, realm_docs_converter_1.convertRealmDocsFromApi)({ project: args.project, outputDir: outDir, branch: args.branch, baseUrl: args.baseUrl });
    }
    else {
        // local mode for backward compatibility
        if (!args.input || !args.out) {
            console.log('Usage (Local): realm-docs-converter --local <input-dir> --out <output-dir>');
            process.exit(1);
        }
        const inputDir = path.resolve(args.input);
        const outputDir = path.resolve(args.out);
        console.log(`Converting local Snooty dir from ${inputDir} to ${outputDir}`);
        totalConverted = await (0, realm_docs_converter_1.convertRealmDocs)({
            inputDir,
            outputDir,
            handleIncludes: true,
            handleSubstitutions: true,
            handleRefs: true,
        });
    }
    console.log('Conversion complete!');
    console.log(`Total pages converted: ${totalConverted}`);
}
main().catch((err) => { console.error(err); process.exit(1); });
