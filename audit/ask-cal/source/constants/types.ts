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
  bodyContent: {
    queryString: string;
    LanguageFacet: string;
    CategoryFacet: string;
    docsSet: string;
  };
  mock?: boolean;
}

export interface HandleRequestProperties {
  options: RequestInit;
  requestType: RequestType;
}

export interface AiSummaryPayload {
  code: string;
  pageUrl: string;
}

// List of possible request types against the API
export enum RequestType {
  Search = "search",
  ReportFeedback = "reportFeedback",
  RequestExample = "requestExample",
  GetAiSummary = "getAiSummary",
}

export const Requests = {
  Search: {
    type: RequestType.Search,
    method: "POST",
    url: "/.netlify/functions/search",
  },
  ReportFeedback: {
    type: RequestType.ReportFeedback,
    method: "POST",
    url: "/.netlify/functions/TODO",
  },
  RequestExample: {
    type: RequestType.RequestExample,
    method: "POST",
    url: "/.netlify/functions/TODO",
  },
  GetAiSummary: {
    type: RequestType.GetAiSummary,
    method: "POST",
    url: "/.netlify/functions/ai-summary",
  },
};
