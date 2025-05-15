import { CodeExampleCategory } from "./categories";
import { DocsSet } from "./docsSets";
import { ProgrammingLanguage } from "./programmingLanguages";

export interface FacetGroup {
  programmingLanguage: ProgrammingLanguage | "";
  category: CodeExampleCategory | "";
  docsSet: DocsSet | "";
}

export interface Facet {
  facet: keyof FacetGroup;
  value: string | "";
}

export interface CodeExample {
  code: string;
  language: ProgrammingLanguage;
  category: CodeExampleCategory;
  pageUrl: string;
  projectName: string;
  pageTitle: string;
  pageDescription: string;
}

export interface SearchResponse {
  queryId: string;
  codeExamples: CodeExample[];
}

export interface RequestProperties {
  bodyContent: unknown;
  mock: boolean;
}

export interface HandleRequestProperties {
  url: string;
  options: RequestInit;
  requestType: RequestType;
}

// List of possible requests against the API
export enum RequestType {
  Search = "search",
  ReportFeedback = "reportFeedback",
  RequestExample = "requestExample",
}
