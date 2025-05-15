import { CodeExampleCategory } from "./categories";
import { DocsSet } from "./docsSets";
import { ProgrammingLanguage } from "./programmingLanguages";

export interface FacetGroup {
  programmingLanguage: ProgrammingLanguage | null;
  category: CodeExampleCategory | null;
  docsSet: DocsSet | null;
}

export interface Facet {
  facet: keyof FacetGroup;
  value: string | null;
}
